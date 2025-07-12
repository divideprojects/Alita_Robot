package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SettingsHandler[T any] struct {
	Collection *mongo.Collection
	Default    func(chatID int64) *T
}

func (h *SettingsHandler[T]) CheckOrInit(chatID int64) *T {
	defaultVal := h.Default(chatID)
	var result T

	err := h.Collection.FindOne(context.TODO(), bson.M{"_id": chatID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		_, err := h.Collection.InsertOne(context.TODO(), defaultVal)
		if err != nil {
			log.Printf("[SettingsHandler] Insert error: %v", err)
		}
		return defaultVal
	} else if err != nil {
		log.Printf("[SettingsHandler] Find error: %v", err)
		return defaultVal
	}
	return &result
}
