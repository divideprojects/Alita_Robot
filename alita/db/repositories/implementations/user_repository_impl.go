package implementations

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/db/repositories"
	"github.com/divideprojects/Alita_Robot/alita/db/repositories/interfaces"
)

// userRepositoryImpl implements the UserRepository interface
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository implementation.
// It takes a GORM database instance and returns a UserRepository interface.
func NewUserRepository(database *gorm.DB) interfaces.UserRepository {
	return &userRepositoryImpl{
		db: database,
	}
}

// CreateOrUpdate creates a new user or updates an existing user in the database.
// It uses GORM's FirstOrCreate with Assign to perform an upsert operation based on user_id.
func (r *userRepositoryImpl) CreateOrUpdate(ctx context.Context, user *db.User) error {
	if user == nil {
		return repositories.ErrUserNil
	}

	if user.UserId == 0 {
		return repositories.ErrUserIDZero
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", user.UserId).Assign(user).FirstOrCreate(&db.User{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] CreateOrUpdate failed: %v", result.Error)
		return fmt.Errorf("failed to create or update user: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %d created/updated successfully", user.UserId)
	return nil
}

// GetByID retrieves a user from the database by their unique user ID.
// Returns nil without error if the user is not found, actual errors are returned as-is.
func (r *userRepositoryImpl) GetByID(ctx context.Context, userID int64) (*db.User, error) {
	if userID == 0 {
		return nil, repositories.ErrUserIDZero
	}

	var user db.User
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // User not found, but not an error
	}

	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] GetByID failed for user %d: %v", userID, result.Error)
		return nil, fmt.Errorf("failed to get user by ID: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %d retrieved successfully", userID)
	return &user, nil
}

// GetByUsername retrieves a user from the database by their username.
// Returns nil without error if the user is not found, actual errors are returned as-is.
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	if username == "" {
		return nil, repositories.ErrUsernameEmpty
	}

	var user db.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // User not found, but not an error
	}

	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] GetByUsername failed for username %s: %v", username, result.Error)
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %s retrieved successfully", username)
	return &user, nil
}

// Update modifies the username and name of an existing user identified by userID.
// Returns an error if the user is not found or if the database operation fails.
func (r *userRepositoryImpl) Update(ctx context.Context, userID int64, username, name string) error {
	if userID == 0 {
		return repositories.ErrUserIDZero
	}

	updates := map[string]interface{}{
		"username": username,
		"name":     name,
	}

	result := r.db.WithContext(ctx).Model(&db.User{}).Where("user_id = ?", userID).Updates(updates)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] Update failed for user %d: %v", userID, result.Error)
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repositories.ErrUserNotFound
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %d updated successfully", userID)
	return nil
}

// Exists checks whether a user with the given userID exists in the database.
// Returns true if found, false if not found, and an error if the database query fails.
func (r *userRepositoryImpl) Exists(ctx context.Context, userID int64) (bool, error) {
	if userID == 0 {
		return false, repositories.ErrUserIDZero
	}

	var count int64
	result := r.db.WithContext(ctx).Model(&db.User{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] Exists check failed for user %d: %v", userID, result.Error)
		return false, fmt.Errorf("failed to check user existence: %w", result.Error)
	}

	return count > 0, nil
}

// GetTotalCount returns the total number of users registered in the system.
// This count is useful for statistics, pagination, and monitoring purposes.
func (r *userRepositoryImpl) GetTotalCount(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&db.User{}).Count(&count)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] GetTotalCount failed: %v", result.Error)
		return 0, fmt.Errorf("failed to get total user count: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[UserRepository] Total user count: %d", count)
	return count, nil
}

// CreateOrUpdateBatch performs bulk create or update operations for multiple users.
// All operations are wrapped in a single database transaction for atomicity and performance.
func (r *userRepositoryImpl) CreateOrUpdateBatch(ctx context.Context, users []*db.User) error {
	if len(users) == 0 {
		return nil // Nothing to do
	}

	// Validate all users first
	for i, user := range users {
		if user == nil {
			return fmt.Errorf("%w: index %d", repositories.ErrUserAtIndexNil, i)
		}
		if user.UserId == 0 {
			return fmt.Errorf("%w: index %d", repositories.ErrUserAtIndexZeroID, i)
		}
	}

	// Use transaction for batch operation
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			result := tx.Where("user_id = ?", user.UserId).Assign(user).FirstOrCreate(&db.User{})
			if result.Error != nil {
				log.WithContext(ctx).Errorf("[UserRepository] Batch operation failed for user %d: %v", user.UserId, result.Error)
				return fmt.Errorf("failed to create or update user %d: %w", user.UserId, result.Error)
			}
		}
		return nil
	})
}

// Delete removes a user from the database permanently.
// This operation is primarily used for GDPR compliance when users request data deletion.
func (r *userRepositoryImpl) Delete(ctx context.Context, userID int64) error {
	if userID == 0 {
		return repositories.ErrUserIDZero
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&db.User{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] Delete failed for user %d: %v", userID, result.Error)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repositories.ErrUserNotFound
	}

	log.WithContext(ctx).Infof("[UserRepository] User %d deleted successfully", userID)
	return nil
}
