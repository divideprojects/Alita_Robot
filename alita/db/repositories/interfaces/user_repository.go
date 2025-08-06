package interfaces

import (
	"context"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

// UserRepository defines the interface for user database operations
type UserRepository interface {
	// CreateOrUpdate creates a new user or updates an existing user in the database.
	// If a user with the same UserId already exists, it will be updated with the new information.
	CreateOrUpdate(ctx context.Context, user *db.User) error

	// GetByID retrieves a user from the database by their unique user ID.
	// Returns nil if the user is not found, without treating it as an error.
	GetByID(ctx context.Context, userID int64) (*db.User, error)

	// GetByUsername retrieves a user from the database by their username.
	// Returns nil if the user is not found, without treating it as an error.
	GetByUsername(ctx context.Context, username string) (*db.User, error)

	// Update modifies the username and name of an existing user identified by userID.
	// Returns an error if the user is not found or if the update operation fails.
	Update(ctx context.Context, userID int64, username, name string) error

	// Exists checks whether a user with the given userID exists in the database.
	// Returns true if the user exists, false if not found, and an error if the check fails.
	Exists(ctx context.Context, userID int64) (bool, error)

	// GetTotalCount returns the total number of users registered in the system.
	// This is useful for statistics and pagination calculations.
	GetTotalCount(ctx context.Context) (int64, error)

	// CreateOrUpdateBatch performs bulk create or update operations for multiple users.
	// This operation is wrapped in a transaction for atomicity and better performance.
	CreateOrUpdateBatch(ctx context.Context, users []*db.User) error

	// Delete removes a user from the database permanently.
	// This operation is primarily used for GDPR compliance when users request data deletion.
	Delete(ctx context.Context, userID int64) error
}
