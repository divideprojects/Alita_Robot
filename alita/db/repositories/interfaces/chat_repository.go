package interfaces

import (
	"context"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// ChatRepository defines the interface for chat database operations
type ChatRepository interface {
	// CreateOrUpdate creates a new chat or updates an existing chat in the database.
	// If a chat with the same ChatId already exists, it will be updated with the new information.
	CreateOrUpdate(ctx context.Context, chat *db.Chat) error

	// GetByID retrieves a chat from the database by its unique chat ID.
	// Returns nil if the chat is not found, without treating it as an error.
	GetByID(ctx context.Context, chatID int64) (*db.Chat, error)

	// Update modifies an existing chat with the provided chat data.
	// Returns an error if the chat is not found or if the update operation fails.
	Update(ctx context.Context, chat *db.Chat) error

	// Exists checks whether a chat with the given chatID exists in the database.
	// Returns true if the chat exists, false if not found, and an error if the check fails.
	Exists(ctx context.Context, chatID int64) (bool, error)

	// GetTotalCount returns the total number of chats registered in the system.
	// This is useful for statistics and monitoring purposes.
	GetTotalCount(ctx context.Context) (int64, error)

	// List retrieves a paginated list of chats from the database.
	// The limit parameter controls how many chats to return, and offset controls pagination.
	List(ctx context.Context, limit, offset int) ([]*db.Chat, error)

	// Delete removes a chat from the database permanently.
	// Returns an error if the chat is not found or if the deletion fails.
	Delete(ctx context.Context, chatID int64) error
}
