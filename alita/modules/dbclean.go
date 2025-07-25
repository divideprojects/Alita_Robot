package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/db"

	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
	log "github.com/sirupsen/logrus"
)

// dbClean presents a UI for developers to initiate database cleanup actions.
//
// Only accessible by the owner or devs. Presents options via inline keyboard.
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

// dbCleanButtonHandler handles callback queries for database cleanup actions.
//
// Performs cleanup of redundant or inactive chats, marks them as inactive in the database, and updates the user with progress.
// Only accessible by the owner or devs.
func (moduleStruct) dbCleanButtonHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.CallbackQuery
	user := ctx.EffectiveSender.User
	msg := query.Message
	memStatus := db.GetTeamMemInfo(user.Id)

	// permissions check
	// only dev can access this
	if user.Id != config.OwnerId && !memStatus.Dev {
		if _, err := query.Answer(
			b,
			&gotgbot.AnswerCallbackQueryOpts{
				Text: "This button can only be used by an admin!",
			},
		); err != nil {
			log.Error("Failed to answer callback query:", err)
		}
		return ext.ContinueGroups
	}

	action := strings.Split(query.Data, ".")[1]
	var finalText, progressString string
	var progress int

	switch action {
	case "chats":
		log.Infof("Chats cleanup requested by %s", user.Username)
		allChats := db.GetAllChats()
		finalText = "No redundant chats found!"
		var chatIds, inactiveChats []int64
		for k := range allChats {
			chatIds = append(chatIds, k)
		}

		for _, chatId := range chatIds {
			// updates the message when percentage gets above the progress gap we have defined
			if (string_handling.FindIndexInt64(chatIds, chatId)) > progress {
				progressString = fmt.Sprintf("%d completed in getting invalid chats.", progress)
				if _, _, err := msg.EditText(b, progressString, nil); err != nil {
					log.Error("Failed to edit progress message:", err)
				}
				progress += 5
			}

			// skip chats who are marked as not IsInactive
			if !db.GetChatSettings(chatId).IsInactive {
				time.Sleep(250 * time.Millisecond)
				_, err := b.GetChat(chatId, nil)
				// only mark chat as failed if it's giving bad request or unauthorized
				if err != nil {
					inactiveChats = append(inactiveChats, chatId)
				}
			}
		}

		if len(inactiveChats) > 0 {
			finalText = fmt.Sprintf("%d chats marked as inactive!", len(inactiveChats))
			log.Infof("These chats have been marked as inactive: %v", inactiveChats)
			for _, chatId := range inactiveChats {
				time.Sleep(250 * time.Millisecond)
				db.ToggleInactiveChat(chatId, true)
			}
		}
	}

	_, _, err := msg.EditText(b, finalText, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	if _, err := query.Answer(b, nil); err != nil {
		log.Error("Failed to answer final callback query:", err)
	}

	return ext.EndGroups
}
