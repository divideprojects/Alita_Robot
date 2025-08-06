package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OptimizedLockQueries provides optimized queries for lock operations
type OptimizedLockQueries struct {
	db *gorm.DB
}

// NewOptimizedLockQueries creates a new instance of optimized lock queries
func NewOptimizedLockQueries() *OptimizedLockQueries {
	if DB == nil {
		log.Error("[OptimizedLockQueries] Database not initialized")
		return &OptimizedLockQueries{db: nil}
	}
	return &OptimizedLockQueries{db: DB}
}

// GetLockStatus retrieves only the lock status for a specific lock type (optimized for 319K calls)
func (o *OptimizedLockQueries) GetLockStatus(chatID int64, lockType string) (bool, error) {
	if o.db == nil {
		return false, errors.New("database not initialized")
	}

	var locked bool
	err := o.db.Model(&LockSettings{}).
		Select("locked").
		Where("chat_id = ? AND lock_type = ?", chatID, lockType).
		Scan(&locked).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // Default to unlocked
	}
	if err != nil {
		log.Errorf("[OptimizedLockQueries] GetLockStatus: %v", err)
		return false, err
	}
	
	return locked, nil
}

// GetChatLocksOptimized retrieves all locks for a chat with minimal columns
func (o *OptimizedLockQueries) GetChatLocksOptimized(chatID int64) (map[string]bool, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	type LockResult struct {
		LockType string
		Locked   bool
	}
	
	var locks []LockResult
	err := o.db.Model(&LockSettings{}).
		Select("lock_type, locked").
		Where("chat_id = ?", chatID).
		Find(&locks).Error
	
	if err != nil {
		log.Errorf("[OptimizedLockQueries] GetChatLocksOptimized: %v", err)
		return nil, err
	}
	
	result := make(map[string]bool)
	for _, lock := range locks {
		result[lock.LockType] = lock.Locked
	}
	
	return result, nil
}

// OptimizedUserQueries provides optimized queries for user operations
type OptimizedUserQueries struct {
	db *gorm.DB
}

// NewOptimizedUserQueries creates a new instance of optimized user queries
func NewOptimizedUserQueries() *OptimizedUserQueries {
	if DB == nil {
		log.Error("[OptimizedUserQueries] Database not initialized")
		return &OptimizedUserQueries{db: nil}
	}
	return &OptimizedUserQueries{db: DB}
}

// GetUserBasicInfo retrieves only essential user information (optimized for 61K calls)
func (o *OptimizedUserQueries) GetUserBasicInfo(userID int64) (*User, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var user User
	err := o.db.Model(&User{}).
		Select("id, user_id, username, name, language").
		Where("user_id = ?", userID).
		First(&user).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[OptimizedUserQueries] GetUserBasicInfo: %v", err)
	}
	
	return &user, err
}

// GetUserLanguage retrieves only the user's language preference
func (o *OptimizedUserQueries) GetUserLanguage(userID int64) (string, error) {
	if o.db == nil {
		return "en", errors.New("database not initialized")
	}

	var language string
	err := o.db.Model(&User{}).
		Select("language").
		Where("user_id = ?", userID).
		Scan(&language).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "en", nil // Default language
	}
	if err != nil {
		log.Errorf("[OptimizedUserQueries] GetUserLanguage: %v", err)
		return "en", err
	}
	
	return language, nil
}

// OptimizedChatQueries provides optimized queries for chat operations
type OptimizedChatQueries struct {
	db *gorm.DB
}

// NewOptimizedChatQueries creates a new instance of optimized chat queries
func NewOptimizedChatQueries() *OptimizedChatQueries {
	if DB == nil {
		log.Error("[OptimizedChatQueries] Database not initialized")
		return &OptimizedChatQueries{db: nil}
	}
	return &OptimizedChatQueries{db: DB}
}

// GetChatBasicInfo retrieves only essential chat information (optimized for 123K calls)
func (o *OptimizedChatQueries) GetChatBasicInfo(chatID int64) (*Chat, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var chat Chat
	err := o.db.Model(&Chat{}).
		Select("id, chat_id, chat_name, language, users, is_inactive").
		Where("chat_id = ?", chatID).
		First(&chat).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[OptimizedChatQueries] GetChatBasicInfo: %v", err)
	}
	
	return &chat, err
}

// GetChatLanguage retrieves only the chat's language preference
func (o *OptimizedChatQueries) GetChatLanguage(chatID int64) (string, error) {
	if o.db == nil {
		return "en", errors.New("database not initialized")
	}

	var language string
	err := o.db.Model(&Chat{}).
		Select("language").
		Where("chat_id = ?", chatID).
		Scan(&language).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "en", nil // Default language
	}
	if err != nil {
		log.Errorf("[OptimizedChatQueries] GetChatLanguage: %v", err)
		return "en", err
	}
	
	return language, nil
}

// IsChatActive checks if a chat is active
func (o *OptimizedChatQueries) IsChatActive(chatID int64) (bool, error) {
	if o.db == nil {
		return false, errors.New("database not initialized")
	}

	var isInactive bool
	err := o.db.Model(&Chat{}).
		Select("is_inactive").
		Where("chat_id = ?", chatID).
		Scan(&isInactive).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if err != nil {
		log.Errorf("[OptimizedChatQueries] IsChatActive: %v", err)
		return false, err
	}
	
	return !isInactive, nil
}

// OptimizedAntifloodQueries provides optimized queries for antiflood operations
type OptimizedAntifloodQueries struct {
	db *gorm.DB
}

// NewOptimizedAntifloodQueries creates a new instance of optimized antiflood queries
func NewOptimizedAntifloodQueries() *OptimizedAntifloodQueries {
	if DB == nil {
		log.Error("[OptimizedAntifloodQueries] Database not initialized")
		return &OptimizedAntifloodQueries{db: nil}
	}
	return &OptimizedAntifloodQueries{db: DB}
}

// GetAntifloodSettings retrieves antiflood settings with minimal columns (optimized for 58K calls)
func (o *OptimizedAntifloodQueries) GetAntifloodSettings(chatID int64) (*AntifloodSettings, error) {
	if o.db == nil {
		return &AntifloodSettings{
			ChatId: chatID,
			Limit:  5,
			Action: "mute",
		}, errors.New("database not initialized")
	}

	var settings AntifloodSettings
	err := o.db.Model(&AntifloodSettings{}).
		Select("id, chat_id, limit, action, action_duration").
		Where("chat_id = ?", chatID).
		First(&settings).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return default settings
		return &AntifloodSettings{
			ChatId: chatID,
			Limit:  5,
			Action: "mute",
		}, nil
	}
	if err != nil {
		log.Errorf("[OptimizedAntifloodQueries] GetAntifloodSettings: %v", err)
		return nil, err
	}
	
	return &settings, nil
}

// IsAntifloodEnabled checks if antiflood is enabled for a chat
func (o *OptimizedAntifloodQueries) IsAntifloodEnabled(chatID int64) (bool, error) {
	if o.db == nil {
		return false, errors.New("database not initialized")
	}

	var limit int
	err := o.db.Model(&AntifloodSettings{}).
		Select("limit").
		Where("chat_id = ?", chatID).
		Scan(&limit).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		log.Errorf("[OptimizedAntifloodQueries] IsAntifloodEnabled: %v", err)
		return false, err
	}
	
	return limit > 0, nil
}

// OptimizedFilterQueries provides optimized queries for filter operations
type OptimizedFilterQueries struct {
	db *gorm.DB
}

// NewOptimizedFilterQueries creates a new instance of optimized filter queries
func NewOptimizedFilterQueries() *OptimizedFilterQueries {
	if DB == nil {
		log.Error("[OptimizedFilterQueries] Database not initialized")
		return &OptimizedFilterQueries{db: nil}
	}
	return &OptimizedFilterQueries{db: DB}
}

// GetChatFiltersOptimized retrieves filters with minimal columns (optimized for 34K calls)
func (o *OptimizedFilterQueries) GetChatFiltersOptimized(chatID int64) ([]*ChatFilters, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var filters []*ChatFilters
	err := o.db.Model(&ChatFilters{}).
		Select("id, keyword, filter_reply, msg_type").
		Where("chat_id = ?", chatID).
		Find(&filters).Error
	
	if err != nil {
		log.Errorf("[OptimizedFilterQueries] GetChatFiltersOptimized: %v", err)
		return nil, err
	}
	
	return filters, nil
}

// GetFilterByKeyword retrieves a specific filter by keyword
func (o *OptimizedFilterQueries) GetFilterByKeyword(chatID int64, keyword string) (*ChatFilters, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var filter ChatFilters
	err := o.db.Model(&ChatFilters{}).
		Select("id, keyword, filter_reply, msg_type").
		Where("chat_id = ? AND keyword = ?", chatID, keyword).
		First(&filter).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[OptimizedFilterQueries] GetFilterByKeyword: %v", err)
	}
	
	return &filter, err
}

// OptimizedBlacklistQueries provides optimized queries for blacklist operations
type OptimizedBlacklistQueries struct {
	db *gorm.DB
}

// NewOptimizedBlacklistQueries creates a new instance of optimized blacklist queries
func NewOptimizedBlacklistQueries() *OptimizedBlacklistQueries {
	if DB == nil {
		log.Error("[OptimizedBlacklistQueries] Database not initialized")
		return &OptimizedBlacklistQueries{db: nil}
	}
	return &OptimizedBlacklistQueries{db: DB}
}

// GetChatBlacklistOptimized retrieves blacklist with minimal columns (optimized for 33K calls)
func (o *OptimizedBlacklistQueries) GetChatBlacklistOptimized(chatID int64) ([]*BlacklistSettings, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var blacklist []*BlacklistSettings
	err := o.db.Model(&BlacklistSettings{}).
		Select("id, word, action").
		Where("chat_id = ?", chatID).
		Find(&blacklist).Error
	
	if err != nil {
		log.Errorf("[OptimizedBlacklistQueries] GetChatBlacklistOptimized: %v", err)
		return nil, err
	}
	
	return blacklist, nil
}

// GetBlacklistWords retrieves only the blacklisted words for a chat
func (o *OptimizedBlacklistQueries) GetBlacklistWords(chatID int64) ([]string, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var words []string
	err := o.db.Model(&BlacklistSettings{}).
		Select("word").
		Where("chat_id = ?", chatID).
		Pluck("word", &words).Error
	
	if err != nil {
		log.Errorf("[OptimizedBlacklistQueries] GetBlacklistWords: %v", err)
		return nil, err
	}
	
	return words, nil
}

// OptimizedChannelQueries provides optimized queries for channel operations
type OptimizedChannelQueries struct {
	db *gorm.DB
}

// NewOptimizedChannelQueries creates a new instance of optimized channel queries
func NewOptimizedChannelQueries() *OptimizedChannelQueries {
	if DB == nil {
		log.Error("[OptimizedChannelQueries] Database not initialized")
		return &OptimizedChannelQueries{db: nil}
	}
	return &OptimizedChannelQueries{db: DB}
}

// GetChannelSettings retrieves channel settings with minimal columns
func (o *OptimizedChannelQueries) GetChannelSettings(chatID int64) (*ChannelSettings, error) {
	if o.db == nil {
		return nil, errors.New("database not initialized")
	}

	var settings ChannelSettings
	err := o.db.Model(&ChannelSettings{}).
		Select("id, chat_id, channel_id").
		Where("chat_id = ?", chatID).
		First(&settings).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("[OptimizedChannelQueries] GetChannelSettings: %v", err)
	}
	
	return &settings, err
}

// CachedOptimizedQueries provides caching layer for optimized queries
type CachedOptimizedQueries struct {
	lockQueries      *OptimizedLockQueries
	userQueries      *OptimizedUserQueries
	chatQueries      *OptimizedChatQueries
	antifloodQueries *OptimizedAntifloodQueries
	filterQueries    *OptimizedFilterQueries
	blacklistQueries *OptimizedBlacklistQueries
	channelQueries   *OptimizedChannelQueries
}

// NewCachedOptimizedQueries creates a new instance with all optimized queries
func NewCachedOptimizedQueries() *CachedOptimizedQueries {
	return &CachedOptimizedQueries{
		lockQueries:      NewOptimizedLockQueries(),
		userQueries:      NewOptimizedUserQueries(),
		chatQueries:      NewOptimizedChatQueries(),
		antifloodQueries: NewOptimizedAntifloodQueries(),
		filterQueries:    NewOptimizedFilterQueries(),
		blacklistQueries: NewOptimizedBlacklistQueries(),
		channelQueries:   NewOptimizedChannelQueries(),
	}
}

// GetLockStatusCached retrieves lock status with caching
func (c *CachedOptimizedQueries) GetLockStatusCached(chatID int64, lockType string) (bool, error) {
	if c == nil || c.lockQueries == nil {
		return false, errors.New("lock queries not initialized")
	}
	
	cacheKey := lockCacheKey(chatID, lockType)
	
	// Try to get from cache first
	cached, err := getFromCacheOrLoad(cacheKey, 1*time.Hour, func() (bool, error) {
		return c.lockQueries.GetLockStatus(chatID, lockType)
	})
	
	if err != nil {
		// Fallback to direct query on cache error
		return c.lockQueries.GetLockStatus(chatID, lockType)
	}
	
	return cached, nil
}

// GetUserBasicInfoCached retrieves user info with caching
func (c *CachedOptimizedQueries) GetUserBasicInfoCached(userID int64) (*User, error) {
	if c == nil || c.userQueries == nil {
		return nil, errors.New("user queries not initialized")
	}
	
	cacheKey := userCacheKey(userID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 1*time.Hour, func() (*User, error) {
		return c.userQueries.GetUserBasicInfo(userID)
	})
	
	if err != nil {
		return c.userQueries.GetUserBasicInfo(userID)
	}
	
	return cached, nil
}

// GetChatBasicInfoCached retrieves chat info with caching
func (c *CachedOptimizedQueries) GetChatBasicInfoCached(chatID int64) (*Chat, error) {
	if c == nil || c.chatQueries == nil {
		return nil, errors.New("chat queries not initialized")
	}
	
	cacheKey := chatCacheKey(chatID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 30*time.Minute, func() (*Chat, error) {
		return c.chatQueries.GetChatBasicInfo(chatID)
	})
	
	if err != nil {
		return c.chatQueries.GetChatBasicInfo(chatID)
	}
	
	return cached, nil
}

// GetAntifloodSettingsCached retrieves antiflood settings with caching
func (c *CachedOptimizedQueries) GetAntifloodSettingsCached(chatID int64) (*AntifloodSettings, error) {
	if c == nil || c.antifloodQueries == nil {
		return nil, errors.New("antiflood queries not initialized")
	}
	
	cacheKey := optimizedAntifloodCacheKey(chatID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 1*time.Hour, func() (*AntifloodSettings, error) {
		return c.antifloodQueries.GetAntifloodSettings(chatID)
	})
	
	if err != nil {
		return c.antifloodQueries.GetAntifloodSettings(chatID)
	}
	
	return cached, nil
}

// GetChatFiltersCached retrieves filters with caching
func (c *CachedOptimizedQueries) GetChatFiltersCached(chatID int64) ([]*ChatFilters, error) {
	if c == nil || c.filterQueries == nil {
		return nil, errors.New("filter queries not initialized")
	}
	
	cacheKey := filterListCacheKey(chatID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 15*time.Minute, func() ([]*ChatFilters, error) {
		return c.filterQueries.GetChatFiltersOptimized(chatID)
	})
	
	if err != nil {
		return c.filterQueries.GetChatFiltersOptimized(chatID)
	}
	
	return cached, nil
}

// GetChatBlacklistCached retrieves blacklist with caching
func (c *CachedOptimizedQueries) GetChatBlacklistCached(chatID int64) ([]*BlacklistSettings, error) {
	if c == nil || c.blacklistQueries == nil {
		return nil, errors.New("blacklist queries not initialized")
	}
	
	cacheKey := blacklistCacheKey(chatID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 15*time.Minute, func() ([]*BlacklistSettings, error) {
		return c.blacklistQueries.GetChatBlacklistOptimized(chatID)
	})
	
	if err != nil {
		return c.blacklistQueries.GetChatBlacklistOptimized(chatID)
	}
	
	return cached, nil
}

// GetChannelSettingsCached retrieves channel settings with caching
func (c *CachedOptimizedQueries) GetChannelSettingsCached(chatID int64) (*ChannelSettings, error) {
	if c == nil || c.channelQueries == nil {
		return nil, errors.New("channel queries not initialized")
	}
	
	cacheKey := channelCacheKey(chatID)
	
	cached, err := getFromCacheOrLoad(cacheKey, 30*time.Minute, func() (*ChannelSettings, error) {
		return c.channelQueries.GetChannelSettings(chatID)
	})
	
	if err != nil {
		return c.channelQueries.GetChannelSettings(chatID)
	}
	
	return cached, nil
}

// Helper functions for cache keys
func lockCacheKey(chatID int64, lockType string) string {
	return fmt.Sprintf("alita:lock:%d:%s", chatID, lockType)
}

func userCacheKey(userID int64) string {
	return fmt.Sprintf("alita:user:%d", userID)
}

func chatCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:chat:%d", chatID)
}

func optimizedAntifloodCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:antiflood:%d", chatID)
}

func channelCacheKey(chatID int64) string {
	return fmt.Sprintf("alita:channel:%d", chatID)
}

// Global instance for optimized queries (singleton pattern with lazy initialization)
var (
	optimizedQueries     *CachedOptimizedQueries
	optimizedQueriesMu   sync.RWMutex
)

// BatchPrefetchContext provides context-aware prefetching
type BatchPrefetchContext struct {
	ctx context.Context
	db  *gorm.DB
}

// NewBatchPrefetchContext creates a new batch prefetch context
func NewBatchPrefetchContext(ctx context.Context) *BatchPrefetchContext {
	if DB == nil {
		log.Error("[BatchPrefetchContext] Database not initialized")
		return &BatchPrefetchContext{
			ctx: ctx,
			db:  nil,
		}
	}
	return &BatchPrefetchContext{
		ctx: ctx,
		db:  DB.WithContext(ctx),
	}
}

// PrefetchUserData prefetches user data for multiple users efficiently
func (b *BatchPrefetchContext) PrefetchUserData(userIDs []int64) (map[int64]*User, error) {
	if b.db == nil {
		return nil, errors.New("database not initialized")
	}

	if len(userIDs) == 0 {
		return make(map[int64]*User), nil
	}
	
	var users []*User
	err := b.db.Model(&User{}).
		Select("id, user_id, username, name, language").
		Where("user_id IN ?", userIDs).
		Find(&users).Error
	
	if err != nil {
		log.Errorf("[BatchPrefetch] PrefetchUserData: %v", err)
		return nil, err
	}
	
	userMap := make(map[int64]*User)
	for _, user := range users {
		userMap[user.UserId] = user
	}
	
	return userMap, nil
}

// PrefetchChatData prefetches chat data for multiple chats efficiently
func (b *BatchPrefetchContext) PrefetchChatData(chatIDs []int64) (map[int64]*Chat, error) {
	if b.db == nil {
		return nil, errors.New("database not initialized")
	}

	if len(chatIDs) == 0 {
		return make(map[int64]*Chat), nil
	}
	
	var chats []*Chat
	err := b.db.Model(&Chat{}).
		Select("id, chat_id, chat_name, language, users, is_inactive").
		Where("chat_id IN ?", chatIDs).
		Find(&chats).Error
	
	if err != nil {
		log.Errorf("[BatchPrefetch] PrefetchChatData: %v", err)
		return nil, err
	}
	
	chatMap := make(map[int64]*Chat)
	for _, chat := range chats {
		chatMap[chat.ChatId] = chat
	}
	
	return chatMap, nil
}

// GetOptimizedQueries returns the singleton instance of optimized queries with thread-safe lazy initialization
func GetOptimizedQueries() *CachedOptimizedQueries {
	// Fast path: Check if already initialized with a valid DB
	optimizedQueriesMu.RLock()
	if optimizedQueries != nil && DB != nil {
		// Check if the instance has valid internal queries
		if optimizedQueries.userQueries != nil && optimizedQueries.userQueries.db != nil {
			optimizedQueriesMu.RUnlock()
			return optimizedQueries
		}
	}
	optimizedQueriesMu.RUnlock()
	
	// Slow path: Need to initialize or re-initialize
	optimizedQueriesMu.Lock()
	defer optimizedQueriesMu.Unlock()
	
	// Double-check after acquiring write lock
	if optimizedQueries != nil && DB != nil && 
		optimizedQueries.userQueries != nil && optimizedQueries.userQueries.db != nil {
		return optimizedQueries
	}
	
	// Initialize or re-initialize
	if DB == nil {
		log.Warn("[GetOptimizedQueries] Database not initialized yet, queries will fail")
		// Return a properly initialized empty instance that will return errors
		return &CachedOptimizedQueries{
			lockQueries:      &OptimizedLockQueries{db: nil},
			userQueries:      &OptimizedUserQueries{db: nil},
			chatQueries:      &OptimizedChatQueries{db: nil},
			antifloodQueries: &OptimizedAntifloodQueries{db: nil},
			filterQueries:    &OptimizedFilterQueries{db: nil},
			blacklistQueries: &OptimizedBlacklistQueries{db: nil},
			channelQueries:   &OptimizedChannelQueries{db: nil},
		}
	}
	
	log.Debug("[GetOptimizedQueries] Initializing optimized queries with valid DB")
	optimizedQueries = NewCachedOptimizedQueries()
	return optimizedQueries
}

// InitOptimizedQueries forces reinitialization of optimized queries (useful for testing or reconnection)
func InitOptimizedQueries() {
	optimizedQueriesMu.Lock()
	defer optimizedQueriesMu.Unlock()
	optimizedQueries = nil
	// The next call to GetOptimizedQueries() will reinitialize
}