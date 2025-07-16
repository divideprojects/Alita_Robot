package db

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MaxPageSize        = 500
	DefaultPageSize    = 100
	MaxOffsetThreshold = 10000
)

type PaginationOptions struct {
	Cursor        interface{}
	Offset        int
	Limit         int
	SortDirection int
}

type PaginatedResult[T any] struct {
	Data       []T
	NextCursor interface{}
	PrevCursor interface{}
	TotalCount int64
}

type MongoPagination[T any] struct {
	collection *mongo.Collection
}

func NewMongoPagination[T any](collection *mongo.Collection) *MongoPagination[T] {
	return &MongoPagination[T]{
		collection: collection,
	}
}

func applySafetyLimits(opts *PaginationOptions) {
	if opts.Limit <= 0 || opts.Limit > MaxPageSize {
		opts.Limit = DefaultPageSize
	}
	if opts.Offset > MaxOffsetThreshold {
		opts.Offset = 0
	}
	if opts.SortDirection != 1 && opts.SortDirection != -1 {
		opts.SortDirection = 1
	}
}

func (mp *MongoPagination[T]) GetNextPage(ctx context.Context, opts PaginationOptions, additionalFilter ...bson.M) (PaginatedResult[T], error) {
	applySafetyLimits(&opts)

	// Create context with timeout if not already provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	filter := bson.M{}
	if opts.Cursor != nil {
		filter["_id"] = bson.M{"$gt": opts.Cursor}
	}

	// Merge additional filter if provided
	if len(additionalFilter) > 0 {
		for key, value := range additionalFilter[0] {
			filter[key] = value
		}
	}

	findOpts := options.Find().
		SetLimit(int64(opts.Limit)).
		SetSort(bson.D{{Key: "_id", Value: opts.SortDirection}})

	cur, err := mp.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return PaginatedResult[T]{}, err
	}
	defer func() {
		if err := cur.Close(ctx); err != nil {
			log.Error("Failed to close pagination cursor:", err)
		}
	}()

	var results []T
	if err := cur.All(ctx, &results); err != nil {
		return PaginatedResult[T]{}, err
	}

	var nextCursor interface{}
	if len(results) > 0 {
		if doc, ok := any(results[len(results)-1]).(bson.M); ok {
			nextCursor = doc["_id"]
		}
	}

	return PaginatedResult[T]{
		Data:       results,
		NextCursor: nextCursor,
	}, nil
}

func (mp *MongoPagination[T]) GetPageByOffset(ctx context.Context, opts PaginationOptions, additionalFilter ...bson.M) (PaginatedResult[T], error) {
	applySafetyLimits(&opts)

	// Create context with timeout if not already provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	filter := bson.M{}
	// Merge additional filter if provided
	if len(additionalFilter) > 0 {
		for key, value := range additionalFilter[0] {
			filter[key] = value
		}
	}

	findOpts := options.Find().
		SetSkip(int64(opts.Offset)).
		SetLimit(int64(opts.Limit)).
		SetSort(bson.D{{Key: "_id", Value: opts.SortDirection}})

	total, err := mp.collection.CountDocuments(ctx, filter)
	if err != nil {
		return PaginatedResult[T]{}, err
	}

	cur, err := mp.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return PaginatedResult[T]{}, err
	}
	defer func() {
		if err := cur.Close(ctx); err != nil {
			log.Error("Failed to close pagination cursor:", err)
		}
	}()

	var results []T
	if err := cur.All(ctx, &results); err != nil {
		return PaginatedResult[T]{}, err
	}

	nextOffset := opts.Offset + opts.Limit
	if nextOffset >= int(total) {
		nextOffset = -1
	}
	prevOffset := opts.Offset - opts.Limit
	if prevOffset < 0 {
		prevOffset = -1
	}

	return PaginatedResult[T]{
		Data:       results,
		TotalCount: total,
		NextCursor: nextOffset,
		PrevCursor: prevOffset,
	}, nil
}
