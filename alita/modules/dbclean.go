package modules

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"

	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	log "github.com/sirupsen/logrus"
)

// ChatValidationJob represents a chat validation job
type ChatValidationJob struct {
	ChatID int64
	Index  int
}

// ChatValidationResult represents the result of chat validation
type ChatValidationResult struct {
	ChatID     int64
	Index      int
	IsInactive bool
	Error      error
}

// ChatValidationWorkerPool manages concurrent chat validation
type ChatValidationWorkerPool struct {
	workerCount    int
	jobs           chan ChatValidationJob
	results        chan ChatValidationResult
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	bot            *gotgbot.Bot
	rateLimitDelay time.Duration
}

// NewChatValidationWorkerPool creates a new worker pool for chat validation
func NewChatValidationWorkerPool(workerCount int, bot *gotgbot.Bot) *ChatValidationWorkerPool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	return &ChatValidationWorkerPool{
		workerCount:    workerCount,
		jobs:           make(chan ChatValidationJob, workerCount*2),
		results:        make(chan ChatValidationResult, workerCount*2),
		ctx:            ctx,
		cancel:         cancel,
		bot:            bot,
		rateLimitDelay: 250 * time.Millisecond,
	}
}

// Start begins processing jobs with the configured number of workers
func (pool *ChatValidationWorkerPool) Start() {
	for i := 0; i < pool.workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}
}

// worker processes chat validation jobs
func (pool *ChatValidationWorkerPool) worker(workerID int) {
	defer pool.wg.Done()

	for {
		select {
		case job, ok := <-pool.jobs:
			if !ok {
				return
			}

			result := ChatValidationResult{
				ChatID: job.ChatID,
				Index:  job.Index,
			}

			// Skip chats marked as inactive
			if !db.GetChatSettings(job.ChatID).IsInactive {
				// Rate limiting to avoid hitting Telegram API limits
				time.Sleep(pool.rateLimitDelay)

				_, err := pool.bot.GetChat(job.ChatID, nil)
				if err != nil {
					result.IsInactive = true
					result.Error = err
				}
			}

			pool.results <- result

		case <-pool.ctx.Done():
			log.WithField("worker_id", workerID).Warn("Worker context cancelled")
			return
		}
	}
}

// AddJob adds a job to the worker pool
func (pool *ChatValidationWorkerPool) AddJob(chatID int64, index int) {
	select {
	case pool.jobs <- ChatValidationJob{ChatID: chatID, Index: index}:
	case <-pool.ctx.Done():
		log.Warn("Cannot add job: worker pool context cancelled")
	}
}

// Close shuts down the worker pool gracefully
func (pool *ChatValidationWorkerPool) Close() {
	close(pool.jobs)
	pool.wg.Wait()
	close(pool.results)
	pool.cancel()
}

func (moduleStruct) dbClean(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	memStatus := db.GetTeamMemInfo(user.Id)

	// only dev can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
		return ext.ContinueGroups
	}

	msg := ctx.EffectiveMessage

	_, err := msg.Reply(
		b,
		"What do you want to clean?",
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{{Text: "Chats", CallbackData: "dbclean.chats"}},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

func (moduleStruct) dbCleanButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := ctx.EffectiveSender.User
	msg := query.Message
	memStatus := db.GetTeamMemInfo(user.Id)

	// permissions check
	// only dev can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
		_, err := query.Answer(
			b,
			&gotgbot.AnswerCallbackQueryOpts{
				Text: "This button can only be used by an admin!",
			},
		)
		if err != nil {
			log.Error(err)
		}
		return ext.ContinueGroups
	}

	action := strings.Split(query.Data, ".")[1]
	var finalText, progressString string

	switch action {
	case "chats":
		log.Infof("Chats cleanup requested by %s", user.Username)
		allChats := db.GetAllChats()
		finalText = "No redundant chats found!"
		var chatIds []int64
		for k := range allChats {
			chatIds = append(chatIds, k)
		}

		if len(chatIds) == 0 {
			break
		}

		// Create worker pool with 10 workers for concurrent chat validation
		workerCount := 10
		if len(chatIds) < workerCount {
			workerCount = len(chatIds)
		}

		pool := NewChatValidationWorkerPool(workerCount, b)
		pool.Start()

		// Submit all jobs
		for i, chatId := range chatIds {
			pool.AddJob(chatId, i)
		}

		// Close job channel to signal no more jobs
		close(pool.jobs)

		// Collect results with progress tracking
		var inactiveChats []int64
		processedCount := 0
		totalChats := len(chatIds)

		// Process results in background goroutine
		go func() {
			for result := range pool.results {
				processedCount++

				if result.IsInactive {
					inactiveChats = append(inactiveChats, result.ChatID)
				}

				// Update progress every 10% or every 100 chats, whichever is smaller
				progressGap := totalChats / 10
				if progressGap > 100 {
					progressGap = 100
				}
				if progressGap == 0 {
					progressGap = 1
				}

				if processedCount%progressGap == 0 {
					percentage := (processedCount * 100) / totalChats
					progressString = fmt.Sprintf("%d%% completed (%d/%d chats validated)",
						percentage, processedCount, totalChats)
					_, _, err := msg.EditText(b, progressString, nil)
					if err != nil {
						log.WithError(err).Warn("Failed to update progress message")
					}
				}
			}
		}()

		// Wait for all workers to complete
		pool.wg.Wait()
		close(pool.results)

		// Clean up worker pool
		pool.cancel()

		// Mark inactive chats in database with concurrent processing
		if len(inactiveChats) > 0 {
			finalText = fmt.Sprintf("%d chats marked as inactive!", len(inactiveChats))
			log.Infof("These chats have been marked as inactive: %v", inactiveChats)

			// Process database updates concurrently with limited goroutines
			dbWorkerCount := 5
			if len(inactiveChats) < dbWorkerCount {
				dbWorkerCount = len(inactiveChats)
			}

			dbJobs := make(chan int64, len(inactiveChats))
			var dbWg sync.WaitGroup

			// Start database update workers
			for i := 0; i < dbWorkerCount; i++ {
				dbWg.Add(1)
				go func(workerID int) {
					defer dbWg.Done()
					for chatID := range dbJobs {
						db.ToggleInactiveChat(chatID, true)
						// Small delay to avoid overwhelming database
						time.Sleep(50 * time.Millisecond)
					}
				}(i)
			}

			// Submit database update jobs
			for _, chatID := range inactiveChats {
				dbJobs <- chatID
			}
			close(dbJobs)

			// Wait for all database updates to complete
			dbWg.Wait()

			log.WithField("count", len(inactiveChats)).Info("Completed marking chats as inactive")
		}
	}

	_, _, err := msg.EditText(b, finalText, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = query.Answer(b, nil)
	if err != nil {
		log.Error(err)
	}

	return ext.EndGroups
}
