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

// NewChatRepository creates a new chat repository implementation
func NewChatRepository(database *gorm.DB) interfaces.ChatRepository {
	return &chatRepositoryImpl{
		db: database,
	}
}

// CreateOrUpdate creates or updates a chat in the database
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

// GetByID retrieves a chat by its ID
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

// Update updates chat information
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

// Exists checks if a chat exists in the database
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

// GetTotalCount returns the total number of chats
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

// List returns a paginated list of chats
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

// Delete removes a chat from the database
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