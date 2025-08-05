package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MigrationConfig struct {
	MongoURI      string
	MongoDatabase string
	PostgresDSN   string
	BatchSize     int
	DryRun        bool
	Verbose       bool
}

type Migrator struct {
	config  *MigrationConfig
	mongoDB *mongo.Database
	pgDB    *gorm.DB
	ctx     context.Context
	stats   *MigrationStats
}

type MigrationStats struct {
	StartTime        time.Time
	EndTime          time.Time
	TotalCollections int
	TotalRecords     int64
	SuccessRecords   int64
	FailedRecords    int64
	Errors           []string
}

func main() {
	var config MigrationConfig

	flag.StringVar(&config.MongoURI, "mongo-uri", "", "MongoDB connection URI")
	flag.StringVar(&config.MongoDatabase, "mongo-db", "alita", "MongoDB database name")
	flag.StringVar(&config.PostgresDSN, "postgres-dsn", "", "PostgreSQL connection DSN")
	flag.IntVar(&config.BatchSize, "batch-size", 1000, "Batch size for processing records")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Perform a dry run without writing to PostgreSQL")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Load environment variables if not provided via flags
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using flags or environment variables")
	}

	if config.MongoURI == "" {
		config.MongoURI = os.Getenv("MONGO_URI")
	}
	if config.PostgresDSN == "" {
		config.PostgresDSN = os.Getenv("DATABASE_URL")
	}

	if config.MongoURI == "" || config.PostgresDSN == "" {
		log.Fatal("MongoDB URI and PostgreSQL DSN are required")
	}

	migrator, err := NewMigrator(&config)
	if err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}
	defer migrator.Close()

	if err := migrator.Run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	migrator.PrintStats()
}

func NewMigrator(config *MigrationConfig) (*Migrator, error) {
	ctx := context.Background()

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	mongoDB := mongoClient.Database(config.MongoDatabase)

	// Connect to PostgreSQL
	gormConfig := &gorm.Config{}
	if !config.Verbose {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	pgDB, err := gorm.Open(postgres.Open(config.PostgresDSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &Migrator{
		config:  config,
		mongoDB: mongoDB,
		pgDB:    pgDB,
		ctx:     ctx,
		stats: &MigrationStats{
			StartTime: time.Now(),
			Errors:    []string{},
		},
	}, nil
}

func (m *Migrator) Close() {
	if client := m.mongoDB.Client(); client != nil {
		_ = client.Disconnect(m.ctx)
	}
	if sqlDB, err := m.pgDB.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func (m *Migrator) Run() error {
	log.Println("Starting migration...")

	collections := []string{
		"users",
		"chats",
		"admin",
		"notes_settings",
		"notes",
		"filters",
		"greetings",
		"locks",
		"pins",
		"rules",
		"warns_settings",
		"warns_users",
		"antiflood_settings",
		"blacklists",
		"channels",
		"connection",
		"connection_settings",
		"disable",
		"report_user_settings",
		"report_chat_settings",
	}

	m.stats.TotalCollections = len(collections)

	for _, collection := range collections {
		log.Printf("Migrating collection: %s", collection)
		if err := m.migrateCollection(collection); err != nil {
			errMsg := fmt.Sprintf("Failed to migrate %s: %v", collection, err)
			m.stats.Errors = append(m.stats.Errors, errMsg)
			log.Printf("ERROR: %s", errMsg)
			if !m.config.DryRun {
				// Continue with other collections even if one fails
				continue
			}
		}
	}

	m.stats.EndTime = time.Now()
	return nil
}

func (m *Migrator) migrateCollection(collectionName string) error {
	switch collectionName {
	case "users":
		return m.migrateUsers()
	case "chats":
		return m.migrateChats()
	case "admin":
		return m.migrateAdmin()
	case "notes_settings":
		return m.migrateNotesSettings()
	case "notes":
		return m.migrateNotes()
	case "filters":
		return m.migrateFilters()
	case "greetings":
		return m.migrateGreetings()
	case "locks":
		return m.migrateLocks()
	case "pins":
		return m.migratePins()
	case "rules":
		return m.migrateRules()
	case "warns_settings":
		return m.migrateWarnsSettings()
	case "warns_users":
		return m.migrateWarnsUsers()
	case "antiflood_settings":
		return m.migrateAntifloodSettings()
	case "blacklists":
		return m.migrateBlacklists()
	case "channels":
		return m.migrateChannels()
	case "connection":
		return m.migrateConnections()
	case "connection_settings":
		return m.migrateConnectionSettings()
	case "disable":
		return m.migrateDisable()
	case "report_user_settings":
		return m.migrateReportUserSettings()
	case "report_chat_settings":
		return m.migrateReportChatSettings()
	default:
		return fmt.Errorf("unknown collection: %s", collectionName)
	}
}

func (m *Migrator) processBatch(collection *mongo.Collection, processor func([]bson.M) error) error {
	cursor, err := collection.Find(m.ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(m.ctx)

	batch := make([]bson.M, 0, m.config.BatchSize)

	for cursor.Next(m.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			m.stats.FailedRecords++
			log.Printf("Failed to decode document: %v", err)
			continue
		}

		batch = append(batch, doc)

		if len(batch) >= m.config.BatchSize {
			if err := processor(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// Process remaining documents
	if len(batch) > 0 {
		if err := processor(batch); err != nil {
			return err
		}
	}

	return cursor.Err()
}

func (m *Migrator) PrintStats() {
	duration := m.stats.EndTime.Sub(m.stats.StartTime)

	fmt.Println("\n=== Migration Statistics ===")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Total Collections: %d\n", m.stats.TotalCollections)
	fmt.Printf("Total Records: %d\n", m.stats.TotalRecords)
	fmt.Printf("Successful Records: %d\n", m.stats.SuccessRecords)
	fmt.Printf("Failed Records: %d\n", m.stats.FailedRecords)

	if len(m.stats.Errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range m.stats.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if m.config.DryRun {
		fmt.Println("\n[DRY RUN] No data was actually written to PostgreSQL")
	}
}

// Helper function to convert MongoDB Long/Int to int64
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case map[string]interface{}:
		// Handle MongoDB Long type
		if longVal, ok := val["$numberLong"]; ok {
			if strVal, ok := longVal.(string); ok {
				var result int64
				fmt.Sscanf(strVal, "%d", &result)
				return result
			}
		}
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return i
		}
	}
	return 0
}

// Helper function to convert interface to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Helper function to convert interface to bool
func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	default:
		return false
	}
}
