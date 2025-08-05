package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func EnsureBotInDb(b *gotgbot.Bot) {
	usersUpdate := &User{UserId: b.Id, UserName: b.Username, Name: b.FirstName}
	err := DB.Where("user_id = ?", b.Id).Assign(usersUpdate).FirstOrCreate(&User{})
	if err.Error != nil {
		log.Errorf("[Database] EnsureBotInDb: %v", err.Error)
	}
	log.Infof("[Database] Bot Updated in Database!")
}

func checkUserInfo(userId int64) (userc *User) {
	userc = &User{}
	err := DB.Where("user_id = ?", userId).First(userc)
	if errors.Is(err.Error, gorm.ErrRecordNotFound) {
		userc = nil
	} else if err.Error != nil {
		log.Errorf("[Database] checkUserInfo: %v - %d", err.Error, userId)
		userc = &User{UserId: userId}
	}
	return userc
}

func UpdateUser(userId int64, username, name string) {
	userc := checkUserInfo(userId)

	if userc != nil {
		if userc.Name == name && userc.UserName == username {
			return
		}
		userc.Name = name
		userc.UserName = username
	} else {
		userc = &User{
			UserId:   userId,
			UserName: username,
			Name:     name,
		}
	}

	err := DB.Where("user_id = ?", userId).Assign(userc).FirstOrCreate(&User{})
	if err.Error != nil {
		log.Errorf("[Database] UpdateUser: %v - %d", err.Error, userId)
		return
	}
	log.Infof("[Database] UpdateUser: %d", userId)
}

func GetUserIdByUserName(username string) int64 {
	var guids User
	err := DB.Where("username = ?", username).First(&guids)
	if errors.Is(err.Error, gorm.ErrRecordNotFound) {
		return 0
	} else if err.Error != nil {
		log.Errorf("[Database] GetUserIdByUserName: %v - %d", err.Error, guids.UserId)
		return 0
	}
	log.Infof("[Database] GetUserIdByUserName: %d", guids.UserId)
	return guids.UserId
}

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

func LoadUsersStats() (count int64) {
	err := DB.Model(&User{}).Count(&count)
	if err.Error != nil {
		log.Errorf("[Database] loadStats: %v", err.Error)
		return
	}
	return
}
