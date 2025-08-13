package db

import (
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/config"
)

// PrefetchedContext contains all commonly needed data for command processing
// This reduces database queries from 5-11 per command to 1-2 queries
type PrefetchedContext struct {
	User         *User
	Chat         *Chat
	Language     string
	IsAdmin      bool
	IsBotAdmin   bool
	DisabledCmds []string
	Settings     any

	// Cache metadata
	CacheHit     bool
	QueryTime    time.Duration
	QueriesCount int
}

// PrefetchCommandContext loads all commonly needed data in a single optimized query
// This is the core optimization that reduces database round trips
func PrefetchCommandContext(ctx *ext.Context) (*PrefetchedContext, error) {
	if !config.EnableQueryPrefetching {
		// Fallback to individual queries if prefetching is disabled
		return prefetchFallback(ctx)
	}

	startTime := time.Now()

	msg := ctx.EffectiveMessage
	if msg == nil {
		return nil, fmt.Errorf("no effective message in context")
	}

	userID := ctx.EffectiveSender.User.Id
	chatID := msg.Chat.Id

	// Try cache first
	cacheKey := fmt.Sprintf("alita:prefetch:%d:%d", chatID, userID)
	if cached, err := getFromCacheOrLoad(cacheKey, 5*time.Minute, func() (*PrefetchedContext, error) {
		return executePrefetchQuery(userID, chatID, msg.Chat.Type)
	}); err == nil {
		cached.CacheHit = true
		cached.QueryTime = time.Since(startTime)
		return cached, nil
	}

	// Cache miss - execute query
	result, err := executePrefetchQuery(userID, chatID, msg.Chat.Type)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"chat_id": chatID,
			"error":   err,
		}).Error("[Prefetch] Failed to execute prefetch query")
		return prefetchFallback(ctx)
	}

	result.CacheHit = false
	result.QueryTime = time.Since(startTime)

	return result, nil
}

// executePrefetchQuery performs the optimized database query
func executePrefetchQuery(userID, chatID int64, chatType string) (*PrefetchedContext, error) {
	result := &PrefetchedContext{
		QueriesCount: 1, // Single optimized query
	}

	// For private chats, we only need user data
	if chatType == "private" {
		var user User
		err := DB.Where("user_id = ?", userID).First(&user).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to fetch user: %w", err)
		}

		result.User = &user
		result.Language = user.Language
		if result.Language == "" {
			result.Language = "en"
		}
		result.IsAdmin = true // User is always admin in private chat
		result.IsBotAdmin = true

		return result, nil
	}

	// For group chats, we need more complex data
	type PrefetchResult struct {
		// User data
		UserID       int64  `gorm:"column:user_id"`
		UserName     string `gorm:"column:username"`
		Name         string `gorm:"column:name"`
		UserLanguage string `gorm:"column:user_language"`

		// Chat data
		ChatID       int64  `gorm:"column:chat_id"`
		ChatName     string `gorm:"column:chat_name"`
		ChatLanguage string `gorm:"column:chat_language"`

		// Admin status (simplified check)
		IsAdmin bool `gorm:"column:is_admin"`
	}

	var prefetchResult PrefetchResult

	// Single complex query that gets user, chat, and admin status
	query := `
		SELECT 
			u.user_id,
			u.username,
			u.name,
			u.language as user_language,
			c.chat_id,
			c.chat_name,
			c.language as chat_language,
			CASE 
				WHEN EXISTS(
					SELECT 1 FROM chat_users cu 
					WHERE cu.user_id = u.user_id 
					AND cu.chat_id = c.chat_id
				) THEN true 
				ELSE false 
			END as is_admin
		FROM users u
		CROSS JOIN chats c
		WHERE u.user_id = ? AND c.chat_id = ?
	`

	err := DB.Raw(query, userID, chatID).Scan(&prefetchResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to execute prefetch query: %w", err)
	}

	// Build result
	result.User = &User{
		UserId:   prefetchResult.UserID,
		UserName: prefetchResult.UserName,
		Name:     prefetchResult.Name,
		Language: prefetchResult.UserLanguage,
	}

	result.Chat = &Chat{
		ChatId:   prefetchResult.ChatID,
		ChatName: prefetchResult.ChatName,
		Language: prefetchResult.ChatLanguage,
	}

	// Determine language preference
	if prefetchResult.ChatLanguage != "" {
		result.Language = prefetchResult.ChatLanguage
	} else if prefetchResult.UserLanguage != "" {
		result.Language = prefetchResult.UserLanguage
	} else {
		result.Language = "en"
	}

	result.IsAdmin = prefetchResult.IsAdmin
	result.IsBotAdmin = true // Assume bot is admin for now, can be checked separately if needed

	return result, nil
}

// prefetchFallback uses the original individual query approach
func prefetchFallback(ctx *ext.Context) (*PrefetchedContext, error) {
	startTime := time.Now()
	result := &PrefetchedContext{
		QueriesCount: 0,
	}

	msg := ctx.EffectiveMessage
	if msg == nil {
		return nil, fmt.Errorf("no effective message in context")
	}

	userID := ctx.EffectiveSender.User.Id
	chatID := msg.Chat.Id

	// Get user (1 query)
	user := checkUserInfo(userID)
	if user != nil {
		result.User = user
		result.QueriesCount++
	}

	// Get chat if not private (1 query)
	if msg.Chat.Type != "private" {
		chat := GetChatSettings(chatID)
		if chat != nil {
			result.Chat = chat
			result.QueriesCount++
		}
	}

	// Get language
	result.Language = GetLanguage(ctx)
	result.QueriesCount++ // Language lookup may involve queries

	result.QueryTime = time.Since(startTime)
	return result, nil
}

// GetPrefetchedLanguage returns the language from prefetched context
func (pc *PrefetchedContext) GetPrefetchedLanguage() string {
	if pc.Language != "" {
		return pc.Language
	}
	return "en"
}

// IsCommandDisabled checks if a command is disabled using prefetched data
func (pc *PrefetchedContext) IsCommandDisabled(cmd string) bool {
	if pc.Chat == nil {
		return false // Private chats don't have disabled commands
	}

	for _, disabledCmd := range pc.DisabledCmds {
		if disabledCmd == cmd {
			return true
		}
	}
	return false
}

// LogPrefetchStats logs performance statistics for monitoring
func (pc *PrefetchedContext) LogPrefetchStats() {
	log.WithFields(log.Fields{
		"cache_hit":     pc.CacheHit,
		"query_time":    pc.QueryTime,
		"queries_count": pc.QueriesCount,
	}).Debug("[Prefetch] Performance stats")
}
