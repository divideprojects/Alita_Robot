package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// EnsureBotInDb ensures that the bot's information is stored in the database.
// Creates or updates the bot's user record with current username and name.
func EnsureBotInDb(b *gotgbot.Bot) {
	usersUpdate := &User{UserId: b.Id, UserName: b.Username, Name: b.FirstName}
	err := DB.Where("user_id = ?", b.Id).Assign(usersUpdate).FirstOrCreate(&User{})
	if err.Error != nil {
		log.Errorf("[Database] EnsureBotInDb: %v", err.Error)
	}
	log.Infof("[Database] Bot Updated in Database!")
}

// checkUserInfo retrieves user information using optimized cached queries.
// Returns nil if the user doesn't exist, or a default User struct on error.
func checkUserInfo(userId int64) (userc *User) {
	// Use optimized cached query instead of SELECT *
	userc, err := GetOptimizedQueries().GetUserBasicInfoCached(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Errorf("[Database] checkUserInfo: %v - %d", err, userId)
		return &User{UserId: userId}
	}
	return userc
}

// UpdateUser creates or updates user information in the database.
// Only updates fields that have actually changed to minimize database operations.
// Invalidates user cache after successful update.
func UpdateUser(userId int64, username, name string) {
	userc := checkUserInfo(userId)

	if userc != nil {
		// Check if update is actually needed
		if userc.Name == name && userc.UserName == username {
			return
		}
		// Only update changed fields
		updates := make(map[string]interface{})
		if userc.Name != name {
			updates["name"] = name
		}
		if userc.UserName != username {
			updates["username"] = username
		}
		if len(updates) > 0 {
			err := DB.Model(&User{}).Where("user_id = ?", userId).Updates(updates).Error
			if err != nil {
				log.Errorf("[Database] UpdateUser: %v - %d", err, userId)
				return
			}
			// Invalidate cache after update
			deleteCache(userCacheKey(userId))
			log.Debugf("[Database] UpdateUser: %d", userId)
		}
	} else {
		// Create new user
		userc = &User{
			UserId:   userId,
			UserName: username,
			Name:     name,
		}
		err := DB.Create(userc).Error
		if err != nil {
			log.Errorf("[Database] UpdateUser: %v - %d", err, userId)
			return
		}
		// Invalidate cache after create
		deleteCache(userCacheKey(userId))
		log.Infof("[Database] UpdateUser: created new user %d", userId)
	}
}

// GetUserIdByUserName retrieves a user ID by their username.
// Returns 0 if the user is not found or an error occurs.
func GetUserIdByUserName(username string) int64 {
	var userId int64
	// Only fetch the user_id column
	err := DB.Model(&User{}).Select("user_id").Where("username = ?", username).Scan(&userId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0
	} else if err != nil {
		log.Errorf("[Database] GetUserIdByUserName: %v - %s", err, username)
		return 0
	}
	log.Debugf("[Database] GetUserIdByUserName: %d", userId)
	return userId
}

// GetUserInfoById retrieves username and name for a given user ID.
// Returns empty strings and false if the user is not found.
func GetUserInfoById(userId int64) (username, name string, found bool) {
	user := checkUserInfo(userId)
	if user != nil {
		username = user.UserName
		name = user.Name
		found = true
		log.Debugf("%+v", user)
	}
	return
}

// LoadUsersStats returns the total count of users in the database.
// Used for generating system statistics and monitoring.
func LoadUsersStats() (count int64) {
	err := DB.Model(&User{}).Count(&count)
	if err.Error != nil {
		log.Errorf("[Database] loadStats: %v", err.Error)
		return
	}
	return
}
