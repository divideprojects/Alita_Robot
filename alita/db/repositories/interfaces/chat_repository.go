package interfaces

import (
	"context"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// ChatRepository defines the interface for chat database operations
type ChatRepository interface {
	// Create or update chat information
	CreateOrUpdate(ctx context.Context, chat *db.Chat) error

	// Get chat by ID
	GetByID(ctx context.Context, chatID int64) (*db.Chat, error)

	// Update chat information
	Update(ctx context.Context, chat *db.Chat) error

	// Check if chat exists
	Exists(ctx context.Context, chatID int64) (bool, error)

	// Get chat statistics
	GetTotalCount(ctx context.Context) (int64, error)

	// List all chats with pagination
	List(ctx context.Context, limit, offset int) ([]*db.Chat, error)

	// Delete chat
	Delete(ctx context.Context, chatID int64) error
}
