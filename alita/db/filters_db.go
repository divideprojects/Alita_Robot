package db

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

func GetFilter(chatID int64, keyword string) (filtSrc *ChatFilters) {
	filtSrc = &ChatFilters{}
	err := GetRecord(filtSrc, map[string]interface{}{"chat_id": chatID, "keyword": keyword})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		filtSrc = &ChatFilters{}
	} else if err != nil {
		log.Errorf("[Database] GetFilter: %v - %d", err, chatID)
		filtSrc = &ChatFilters{}
	}
	return
}

//goland:noinspection GoUnusedExportedFunction
func GetAllFilters(chatID int64) (allFilters []*ChatFilters) {
	err := GetRecords(&allFilters, map[string]interface{}{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database] GetAllFilters: %v - %d", err, chatID)
		return []*ChatFilters{}
	}
	return
}

func GetFiltersList(chatID int64) (allFilterWords []string) {
	var results []*ChatFilters
	err := GetRecords(&results, map[string]interface{}{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database] GetFiltersList: %v - %d", err, chatID)
		return []string{}
	}

	for _, j := range results {
		allFilterWords = append(allFilterWords, j.KeyWord)
	}
	return
}

func DoesFilterExists(chatId int64, keyword string) bool {
	return string_handling.FindInStringSlice(GetFiltersList(chatId), strings.ToLower(keyword))
}

func AddFilter(chatID int64, keyWord, replyText, fileID string, buttons []Button, filtType int) {
	if string_handling.FindInStringSlice(GetFiltersList(chatID), keyWord) {
		return
	}

	// add the filter
	newFilter := ChatFilters{
		ChatId:      chatID,
		KeyWord:     keyWord,
		FilterReply: replyText,
		MsgType:     filtType,
		FileID:      fileID,
		Buttons:     ButtonArray(buttons),
	}

	err := CreateRecord(&newFilter)
	if err != nil {
		log.Errorf("[Database][AddFilter]: %d - %v", chatID, err)
		return
	}
}

func RemoveFilter(chatID int64, keyWord string) {
	if !string_handling.FindInStringSlice(GetFiltersList(chatID), keyWord) {
		return
	}

	err := DB.Where("chat_id = ? AND keyword = ?", chatID, keyWord).Delete(&ChatFilters{}).Error
	if err != nil {
		log.Errorf("[Database][RemoveFilter]: %d - %v", chatID, err)
		return
	}
}

func RemoveAllFilters(chatID int64) {
	err := DB.Where("chat_id = ?", chatID).Delete(&ChatFilters{}).Error
	if err != nil {
		log.Errorf("[Database][RemoveAllFilters]: %d - %v", chatID, err)
	}
}

func CountFilters(chatID int64) (filtersNum int64) {
	err := DB.Model(&ChatFilters{}).Where("chat_id = ?", chatID).Count(&filtersNum).Error
	if err != nil {
		log.Errorf("[Database][CountFilters]: %d - %v", chatID, err)
	}
	return
}

func LoadFilterStats() (filtersNum, filtersUsingChats int64) {
	var filterStruct []*ChatFilters
	filtersMap := make(map[int64][]ChatFilters)

	err := GetRecords(&filterStruct, map[string]interface{}{})
	if err != nil {
		log.Errorf("[Database][LoadFilterStats]: %v", err)
		return
	}

	for _, filterC := range filterStruct {
		filtersNum++ // count number of filters
		filtersMap[filterC.ChatId] = append(filtersMap[filterC.ChatId], *filterC)
	}

	filtersUsingChats = int64(len(filtersMap))

	return
}
