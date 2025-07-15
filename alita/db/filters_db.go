package db

import (
	"strings"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/cache"
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
	err := findOne(getCollection("filters"), bson.M{"chat_id": chatID, "keyword": keyword}).Decode(&filtSrc)
	if err == mongo.ErrNoDocuments {
		filtSrc = &ChatFilters{}
	} else if err != nil {
		log.Errorf("[Database] GetFilter: %v - %d", err, chatID)
	}
	return
}

// GetAllFiltersPaginated returns paginated filters for a chat.
func GetAllFiltersPaginated(_ int64, opts PaginationOptions) (PaginatedResult[*ChatFilters], error) {
	paginator := NewMongoPagination[*ChatFilters](getCollection("filters"))

	if opts.Cursor == nil && opts.Offset == 0 {
		// Default to cursor-based pagination
		return paginator.GetNextPage(bgCtx, bson.M{}, PaginationOptions{
			Limit:         opts.Limit,
			SortDirection: 1,
		})
	}

	if opts.Offset > 0 {
		return paginator.GetPageByOffset(bgCtx, bson.M{}, PaginationOptions{
			Offset:        opts.Offset,
			Limit:         opts.Limit,
			SortDirection: 1,
		})
	}

	return paginator.GetNextPage(bgCtx, bson.M{}, opts)
}

// GetFiltersList returns a list of all filter keywords for a chat.
func GetFiltersList(chatID int64) (allFilterWords []string) {
	var results []*ChatFilters
	cursor := findAll(getCollection("filters"), bson.M{"chat_id": chatID})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &results)
	for _, j := range results {
		allFilterWords = append(allFilterWords, j.KeyWord)
	}
	return
}

// DoesFilterExists returns true if a filter with the given keyword exists in the chat.
func DoesFilterExists(chatId int64, keyword string) bool {
	filtersList := GetFiltersList(chatId)
	filtersMap := string_handling.StringSliceToMap(filtersList)
	return string_handling.FindInStringMap(filtersMap, strings.ToLower(keyword))
}

// AddFilter adds a new filter to the chat with the specified properties.
// If a filter with the same keyword already exists, no action is taken.
// Returns true if a new filter was added, false if it already existed.
func AddFilter(chatID int64, keyWord, replyText, fileID string, buttons []Button, filtType int) bool {
	filter := bson.M{"chat_id": chatID, "keyword": keyWord}
	update := bson.M{
		"$setOnInsert": bson.M{
			"chat_id":        chatID,
			"keyword":        keyWord,
			"filter_reply":   replyText,
			"msgtype":        filtType,
			"fileid":         fileID,
			"filter_buttons": buttons,
		},
	}

	result := &ChatFilters{}
	err := withTransaction(bgCtx, func(sessCtx mongo.SessionContext) error {
		// Perform the upsert operation within the transaction
		err := findOneAndUpsert(getCollection("filters"), filter, update, result)
		if err != nil {
			return err
		}

		// Update cache only if this was a new insert
		if result.ChatId == chatID && result.KeyWord == keyWord {
			if err := cache.Marshal.Set(cache.Context, chatID, result, store.WithExpiration(10*time.Minute)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Errorf("[Database][AddFilter]: %d - %v", chatID, err)
		return false
	}

	// Return true if this was a new insert
	return result.ChatId == chatID && result.KeyWord == keyWord
}

// RemoveFilter deletes a filter by keyword from the chat.
func RemoveFilter(chatID int64, keyWord string) {
	err := withTransaction(bgCtx, func(sessCtx mongo.SessionContext) error {
		// Perform the delete operation within the transaction
		err := deleteOne(getCollection("filters"), bson.M{"chat_id": chatID, "keyword": keyWord})
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}

		// Invalidate cache after successful deletion
		if err := cache.Marshal.Delete(cache.Context, chatID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Errorf("[Database][RemoveFilter]: %d - %v", chatID, err)
	}
}

// RemoveAllFilters deletes all filters from the specified chat.
func RemoveAllFilters(chatID int64) {
	err := deleteMany(getCollection("filters"), bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][RemoveAllFilters]: %d - %v", chatID, err)
	}
}

// CountFilters returns the number of filters for a chat.
func CountFilters(chatID int64) (filtersNum int64) {
	filtersNum, err := countDocs(getCollection("filters"), bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][CountFilters]: %d - %v", chatID, err)
	}
	return
}

// LoadFilterStats returns the total number of filters and the number of chats using filters.
// Uses MongoDB aggregation pipeline for optimal performance.
func LoadFilterStats() (filtersNum, filtersUsingChats int64) {
	// Use MongoDB aggregation for optimal performance
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":         "$chat_id",
				"filterCount": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id":          nil,
				"totalFilters": bson.M{"$sum": "$filterCount"},
				"totalChats":   bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := getCollection("filters").Aggregate(bgCtx, pipeline)
	if err != nil {
		log.Error("Failed to aggregate filter stats:", err)
		// Fallback to manual method if aggregation fails
		return loadFilterStatsManual()
	}
	defer cursor.Close(bgCtx)

	var result struct {
		TotalFilters int64 `bson:"totalFilters"`
		TotalChats   int64 `bson:"totalChats"`
	}

	if cursor.Next(bgCtx) {
		if err := cursor.Decode(&result); err != nil {
			log.Error("Failed to decode filter stats:", err)
			// Fallback to manual method if decode fails
			return loadFilterStatsManual()
		}
		return result.TotalFilters, result.TotalChats
	}

	// No results found, return zeros
	return 0, 0
}

/*
loadFilterStatsManual is the fallback manual implementation with pagination.
Used when MongoDB aggregation fails for any reason.
*/
func loadFilterStatsManual() (filtersNum, filtersUsingChats int64) {
	paginator := NewMongoPagination[*ChatFilters](getCollection("filters"))
	chatsMap := make(map[int64]struct{})

	// Process in paginated batches
	for {
		result, err := paginator.GetNextPage(bgCtx, bson.M{}, PaginationOptions{
			Limit:         1000, // Process 1000 docs at a time
			SortDirection: 1,
		})
		if err != nil || len(result.Data) == 0 {
			break
		}

		for _, filter := range result.Data {
			filtersNum++
			chatsMap[filter.ChatId] = struct{}{}
		}
	}

	return filtersNum, int64(len(chatsMap))
}
