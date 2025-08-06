package implementations

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/db/repositories/interfaces"
)

// chatRepositoryImpl implements the ChatRepository interface
type chatRepositoryImpl struct {
	db *gorm.DB
}

// NewChatRepository creates a new chat repository implementation.
// It takes a GORM database instance and returns a ChatRepository interface.
func NewChatRepository(database *gorm.DB) interfaces.ChatRepository {
	return &chatRepositoryImpl{
		db: database,
	}
}

// CreateOrUpdate creates a new chat or updates an existing chat in the database.
// It uses GORM's FirstOrCreate with Assign to perform an upsert operation based on chat_id.
func (r *chatRepositoryImpl) CreateOrUpdate(ctx context.Context, chat *db.Chat) error {
	if chat == nil {
		return fmt.Errorf("chat cannot be nil")
	}

	if chat.ChatId == 0 {
		return fmt.Errorf("chat ID cannot be zero")
	}

	result := r.db.WithContext(ctx).Where("chat_id = ?", chat.ChatId).Assign(chat).FirstOrCreate(&db.Chat{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] CreateOrUpdate failed: %v", result.Error)
		return fmt.Errorf("failed to create or update chat: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[ChatRepository] Chat %d created/updated successfully", chat.ChatId)
	return nil
}

// GetByID retrieves a chat from the database by its unique chat ID.
// Returns nil without error if the chat is not found, actual errors are returned as-is.
func (r *chatRepositoryImpl) GetByID(ctx context.Context, chatID int64) (*db.Chat, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat ID cannot be zero")
	}

	var chat db.Chat
	result := r.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&chat)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // Chat not found, but not an error
	}

	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] GetByID failed for chat %d: %v", chatID, result.Error)
		return nil, fmt.Errorf("failed to get chat by ID: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[ChatRepository] Chat %d retrieved successfully", chatID)
	return &chat, nil
}

// Update modifies an existing chat with the provided chat data.
// Returns an error if the chat is not found or if the database operation fails.
func (r *chatRepositoryImpl) Update(ctx context.Context, chat *db.Chat) error {
	if chat == nil {
		return fmt.Errorf("chat cannot be nil")
	}

	if chat.ChatId == 0 {
		return fmt.Errorf("chat ID cannot be zero")
	}

	result := r.db.WithContext(ctx).Model(&db.Chat{}).Where("chat_id = ?", chat.ChatId).Updates(chat)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] Update failed for chat %d: %v", chat.ChatId, result.Error)
		return fmt.Errorf("failed to update chat: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("chat %d not found", chat.ChatId)
	}

	log.WithContext(ctx).Debugf("[ChatRepository] Chat %d updated successfully", chat.ChatId)
	return nil
}

// Exists checks whether a chat with the given chatID exists in the database.
// Returns true if found, false if not found, and an error if the database query fails.
func (r *chatRepositoryImpl) Exists(ctx context.Context, chatID int64) (bool, error) {
	if chatID == 0 {
		return false, fmt.Errorf("chat ID cannot be zero")
	}

	var count int64
	result := r.db.WithContext(ctx).Model(&db.Chat{}).Where("chat_id = ?", chatID).Count(&count)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] Exists check failed for chat %d: %v", chatID, result.Error)
		return false, fmt.Errorf("failed to check chat existence: %w", result.Error)
	}

	return count > 0, nil
}

// GetTotalCount returns the total number of chats registered in the system.
// This count is useful for statistics, monitoring, and administrative purposes.
func (r *chatRepositoryImpl) GetTotalCount(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&db.Chat{}).Count(&count)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] GetTotalCount failed: %v", result.Error)
		return 0, fmt.Errorf("failed to get total chat count: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[ChatRepository] Total chat count: %d", count)
	return count, nil
}

// List retrieves a paginated list of chats from the database.
// It applies sensible defaults and limits to prevent excessive memory usage.
func (r *chatRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*db.Chat, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}
	if offset < 0 {
		offset = 0
	}

	var chats []*db.Chat
	result := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&chats)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] List failed: %v", result.Error)
		return nil, fmt.Errorf("failed to list chats: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[ChatRepository] Listed %d chats", len(chats))
	return chats, nil
}

// Delete removes a chat from the database permanently.
// Returns an error if the chat is not found or if the deletion operation fails.
func (r *chatRepositoryImpl) Delete(ctx context.Context, chatID int64) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be zero")
	}

	result := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Delete(&db.Chat{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[ChatRepository] Delete failed for chat %d: %v", chatID, result.Error)
		return fmt.Errorf("failed to delete chat: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("chat %d not found", chatID)
	}

	log.WithContext(ctx).Infof("[ChatRepository] Chat %d deleted successfully", chatID)
	return nil
}
