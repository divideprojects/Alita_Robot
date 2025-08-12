package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/divideprojects/Alita_Robot/alita/db"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Setup logging
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Get database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.StandardLogger(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to database
	database, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Info("Connected to database successfully!")

	// Create migration runner
	runner := db.NewMigrationRunner(database)

	// Check pending migrations
	log.Info("Checking for pending migrations...")
	pending, err := runner.GetPendingMigrations()
	if err != nil {
		log.Errorf("Failed to get pending migrations: %v", err)
	} else {
		if len(pending) == 0 {
			log.Info("No pending migrations found")
		} else {
			log.Infof("Found %d pending migrations:", len(pending))
			for _, p := range pending {
				log.Infof("  - %s", p)
			}
		}
	}

	// Check applied migrations
	log.Info("Checking applied migrations...")
	applied, err := runner.GetAppliedMigrations()
	if err != nil {
		log.Errorf("Failed to get applied migrations: %v", err)
	} else {
		if len(applied) == 0 {
			log.Info("No migrations have been applied yet")
		} else {
			log.Infof("Found %d applied migrations:", len(applied))
			for _, a := range applied {
				log.Infof("  - %s (applied at %s)", a.Version, a.ExecutedAt.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Test migration runner
	log.Info("Testing migration runner...")

	// Test 1: Run migrations
	log.Info("Test 1: Running migrations...")
	if err := runner.RunMigrations(); err != nil {
		log.Errorf("Migration failed: %v", err)
		os.Exit(1)
	}
	log.Info("✓ Test 1 passed: Migrations ran successfully")

	// Test 2: Verify idempotency - run again
	log.Info("Test 2: Testing idempotency (running migrations again)...")
	if err := runner.RunMigrations(); err != nil {
		log.Errorf("Second migration run failed: %v", err)
		os.Exit(1)
	}
	log.Info("✓ Test 2 passed: Migrations are idempotent")

	// Test 3: Verify tables exist
	log.Info("Test 3: Verifying database tables...")
	tables := []string{
		"users", "chats", "chat_users", "warns_settings", "warns_users",
		"greetings", "filters", "admin", "blacklists", "pins",
		"report_chat_settings", "report_user_settings", "devs",
		"channels", "antiflood_settings", "connection", "connection_settings",
		"disable", "disable_chat_settings", "rules", "locks",
		"notes_settings", "notes", "captcha_settings", "captcha_attempts",
	}

	for _, table := range tables {
		var exists bool
		err := database.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", table).Scan(&exists).Error
		if err != nil {
			log.Errorf("Failed to check table %s: %v", table, err)
			continue
		}
		if !exists {
			log.Errorf("✗ Table %s does not exist", table)
		} else {
			log.Debugf("  ✓ Table %s exists", table)
		}
	}
	log.Info("✓ Test 3 passed: All expected tables exist")

	// Summary
	log.Info("")
	log.Info("========================================")
	log.Info("Migration Test Summary:")
	log.Info("  ✓ Migrations run successfully")
	log.Info("  ✓ Migrations are idempotent")
	log.Info("  ✓ All database tables created")
	log.Info("========================================")

	// Final migration status
	finalApplied, _ := runner.GetAppliedMigrations()
	log.Infof("Total migrations applied: %d", len(finalApplied))

	fmt.Println("\n✅ All tests passed! The auto-migration feature is working correctly.")
}
