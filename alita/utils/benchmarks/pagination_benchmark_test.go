package benchmarks

import (
	"context"
	"testing"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	benchmarkCollectionSize = 10000 // Reduced size for testing
	benchmarkPageSize       = 100
)

var (
	testCollection *mongo.Collection
)

func BenchmarkGetAllFilters(b *testing.B) {
	setupBenchmarkCollection(b)
	defer cleanupBenchmarkCollection(b)

	b.Run("LegacyImplementation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = db.GetAllFilters(0) // 0 gets all records
		}
	})

	b.Run("PaginatedImplementation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, _ := db.GetAllFiltersPaginated(0, db.PaginationOptions{
				Limit: benchmarkPageSize,
			})
			_ = result.Data
		}
	})
}

func BenchmarkCursorPagination(b *testing.B) {
	setupBenchmarkCollection(b)
	defer cleanupBenchmarkCollection(b)

	paginator := db.NewMongoPagination[bson.M](testCollection)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cursor interface{}
		for {
			result, _ := paginator.GetNextPage(context.Background(), db.PaginationOptions{
				Cursor: cursor,
				Limit:  benchmarkPageSize,
			})
			if len(result.Data) == 0 {
				break
			}
			cursor = result.NextCursor
		}
	}
}

func BenchmarkOffsetPagination(b *testing.B) {
	setupBenchmarkCollection(b)
	defer cleanupBenchmarkCollection(b)

	paginator := db.NewMongoPagination[bson.M](testCollection)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := 0
		for {
			result, _ := paginator.GetPageByOffset(context.Background(), db.PaginationOptions{
				Offset: offset,
				Limit:  benchmarkPageSize,
			})
			if len(result.Data) == 0 {
				break
			}
			offset += benchmarkPageSize
		}
	}
}

func setupBenchmarkCollection(b *testing.B) {
	// Get test collection from db package
	testCollection = db.GetTestCollection()

	// Insert benchmark data
	docs := make([]interface{}, benchmarkCollectionSize)
	for i := 0; i < benchmarkCollectionSize; i++ {
		docs[i] = bson.M{"value": i}
	}
	_, err := testCollection.InsertMany(context.Background(), docs)
	if err != nil {
		b.Fatalf("Failed to setup benchmark collection: %v", err)
	}
}

func cleanupBenchmarkCollection(b *testing.B) {
	_, err := testCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		b.Logf("Failed to clean up benchmark collection: %v", err)
	}
}
