package interfaces

import (
	"context"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// UserRepository defines the interface for user database operations
type UserRepository interface {
	// Create or update user information
	CreateOrUpdate(ctx context.Context, user *db.User) error
	
	// Get user by ID
	GetByID(ctx context.Context, userID int64) (*db.User, error)
	
	// Get user by username
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	
	// Update user information
	Update(ctx context.Context, userID int64, username, name string) error
	
	// Check if user exists
	Exists(ctx context.Context, userID int64) (bool, error)
	
	// Get user statistics
	GetTotalCount(ctx context.Context) (int64, error)
	
	// Batch operations
	CreateOrUpdateBatch(ctx context.Context, users []*db.User) error
	
	// Delete user (for GDPR compliance)
	Delete(ctx context.Context, userID int64) error
}