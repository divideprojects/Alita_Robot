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

// userRepositoryImpl implements the UserRepository interface
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(database *gorm.DB) interfaces.UserRepository {
	return &userRepositoryImpl{
		db: database,
	}
}

// CreateOrUpdate creates or updates a user in the database
func (r *userRepositoryImpl) CreateOrUpdate(ctx context.Context, user *db.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.UserId == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", user.UserId).Assign(user).FirstOrCreate(&db.User{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] CreateOrUpdate failed: %v", result.Error)
		return fmt.Errorf("failed to create or update user: %w", result.Error)
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %d created/updated successfully", user.UserId)
	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepositoryImpl) GetByID(ctx context.Context, userID int64) (*db.User, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be zero")
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

// GetByUsername retrieves a user by their username
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
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

// Update updates user information
func (r *userRepositoryImpl) Update(ctx context.Context, userID int64, username, name string) error {
	if userID == 0 {
		return fmt.Errorf("user ID cannot be zero")
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
		return fmt.Errorf("user %d not found", userID)
	}

	log.WithContext(ctx).Debugf("[UserRepository] User %d updated successfully", userID)
	return nil
}

// Exists checks if a user exists in the database
func (r *userRepositoryImpl) Exists(ctx context.Context, userID int64) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("user ID cannot be zero")
	}

	var count int64
	result := r.db.WithContext(ctx).Model(&db.User{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] Exists check failed for user %d: %v", userID, result.Error)
		return false, fmt.Errorf("failed to check user existence: %w", result.Error)
	}

	return count > 0, nil
}

// GetTotalCount returns the total number of users
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

// CreateOrUpdateBatch creates or updates multiple users in a single transaction
func (r *userRepositoryImpl) CreateOrUpdateBatch(ctx context.Context, users []*db.User) error {
	if len(users) == 0 {
		return nil // Nothing to do
	}

	// Validate all users first
	for i, user := range users {
		if user == nil {
			return fmt.Errorf("user at index %d cannot be nil", i)
		}
		if user.UserId == 0 {
			return fmt.Errorf("user at index %d has zero ID", i)
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

// Delete removes a user from the database (GDPR compliance)
func (r *userRepositoryImpl) Delete(ctx context.Context, userID int64) error {
	if userID == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&db.User{})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("[UserRepository] Delete failed for user %d: %v", userID, result.Error)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user %d not found", userID)
	}

	log.WithContext(ctx).Infof("[UserRepository] User %d deleted successfully", userID)
	return nil
}