package db

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

// ChatFilters represents a filter rule for a chat.
//
// Fields:
//   - ChatId: Unique identifier for the chat.
//   - KeyWord: The keyword that triggers the filter.
//   - FilterReply: The reply message for the filter.
//   - MsgType: Type of message (e.g., text, media).
//   - FileID: Optional file ID for media attachments.
//   - NoNotif: Whether to suppress notifications when sending the filter reply.
//   - Buttons: List of buttons to attach to the filter reply.
type ChatFilters struct {
	ChatId      int64    `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	KeyWord     string   `bson:"keyword,omitempty" json:"keyword,omitempty"`
	FilterReply string   `bson:"filter_reply,omitempty" json:"filter_reply,omitempty"`
	MsgType     int      `bson:"msgtype,omitempty" json:"msgtype,omitempty"`
	FileID      string   `bson:"fileid,omitempty" json:"fileid,omitempty"`
	NoNotif     bool     `bson:"nonotif,omitempty" json:"nonotif,omitempty"`
	Buttons     []Button `bson:"filter_buttons,omitempty" json:"filter_buttons,omitempty"`
}

// GetFilter retrieves a filter by keyword for a chat.
// Returns a new ChatFilters struct if the filter does not exist.
func GetFilter(chatID int64, keyword string) (filtSrc *ChatFilters) {
	err := findOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyword}).Decode(&filtSrc)
	if err == mongo.ErrNoDocuments {
		filtSrc = &ChatFilters{}
	} else if err != nil {
		log.Errorf("[Database] GetFilter: %v - %d", err, chatID)
	}
	return
}

// GetAllFilters returns all filters for a chat.
func GetAllFilters(chatID int64) (allFilters []*ChatFilters) {
	cursor := findAll(filterColl, bson.M{"chat_id": chatID})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &allFilters)
	return
}

// GetFiltersList returns a list of all filter keywords for a chat.
func GetFiltersList(chatID int64) (allFilterWords []string) {
	var results []*ChatFilters
	cursor := findAll(filterColl, bson.M{"chat_id": chatID})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &results)
	for _, j := range results {
		allFilterWords = append(allFilterWords, j.KeyWord)
	}
	return
}

// DoesFilterExists returns true if a filter with the given keyword exists in the chat.
func DoesFilterExists(chatId int64, keyword string) bool {
	return string_handling.FindInStringSlice(GetFiltersList(chatId), strings.ToLower(keyword))
}

// AddFilter adds a new filter to the chat with the specified properties.
// If a filter with the same keyword already exists, no action is taken.
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
		Buttons:     buttons,
	}

	err := updateOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyWord}, newFilter)
	if err != nil {
		log.Errorf("[Database][AddFilter]: %d - %v", chatID, err)
		return
	}
}

// RemoveFilter deletes a filter by keyword from the chat.
// If the filter does not exist, no action is taken.
func RemoveFilter(chatID int64, keyWord string) {
	if !string_handling.FindInStringSlice(GetFiltersList(chatID), keyWord) {
		return
	}

	err := deleteOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyWord})
	if err != nil {
		log.Errorf("[Database][RemoveFilter]: %d - %v", chatID, err)
		return
	}
}

// RemoveAllFilters deletes all filters from the specified chat.
func RemoveAllFilters(chatID int64) {
	err := deleteMany(filterColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][RemoveAllFilters]: %d - %v", chatID, err)
	}
}

// CountFilters returns the number of filters for a chat.
func CountFilters(chatID int64) (filtersNum int64) {
	filtersNum, err := countDocs(filterColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][CountFilters]: %d - %v", chatID, err)
	}
	return
}

// LoadFilterStats returns the total number of filters and the number of chats using filters.
func LoadFilterStats() (filtersNum, filtersUsingChats int64) {
	return CountByChat(filterColl, bson.M{}, "chat_id")
}
