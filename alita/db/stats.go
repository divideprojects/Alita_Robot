package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CountByChat iterates over the provided Mongo collection and counts
// (1) the total number of documents that match the supplied filter and
// (2) the number of unique chats (identified by the field name provided in chatField).
//
// It returns two ints: `items` (total documents) and `chats` (unique chat IDs).
//
// This helper consolidates the repetitive stats code that previously existed as
// LoadXYZStats functions in every *_db.go file.
func CountByChat(collection *mongo.Collection, filter bson.M, chatField string) (items, chats int64) {
    chatSet := make(map[int64]struct{})

    cursor := findAll(collection, filter)
    if cursor == nil {
        return 0, 0 // findAll already logs the error
    }
    defer cursor.Close(bgCtx)

    for cursor.Next(bgCtx) {
        var doc bson.M
        if err := cursor.Decode(&doc); err != nil {
            // skip malformed docs – the caller may handle/log aggregated errors
            continue
        }

        items++

        if chatIDRaw, ok := doc[chatField]; ok {
            switch v := chatIDRaw.(type) {
            case int64:
                chatSet[v] = struct{}{}
            case int32:
                chatSet[int64(v)] = struct{}{}
            case float64:
                chatSet[int64(v)] = struct{}{}
            }
        }
    }

    chats = int64(len(chatSet))
    return
} 