package db

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/divideprojects/Alita_Robot/alita/config"
)

// MigrationRunner handles automatic database migrations
type MigrationRunner struct {
	db             *gorm.DB
	migrationsPath string
	cleanSQL       bool
}

// SchemaMigration represents a migration record in the database
type SchemaMigration struct {
	Version    string    `gorm:"primaryKey;column:version"`
	ExecutedAt time.Time `gorm:"column:executed_at"`
}

// TableName returns the database table name for schema migrations
func (SchemaMigration) TableName() string {
	return "schema_migrations"
}

// NewMigrationRunner creates a new migration runner instance
func NewMigrationRunner(db *gorm.DB) *MigrationRunner {
	return &MigrationRunner{
		db:             db,
		migrationsPath: config.MigrationsPath,
		cleanSQL:       true, // Always clean Supabase-specific SQL
	}
}

// RunMigrations executes all pending database migrations
func (m *MigrationRunner) RunMigrations() error {
	log.Info("[Migrations] Starting automatic database migration...")

	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	if len(files) == 0 {
		log.Info("[Migrations] No migration files found")
		return nil
	}

	log.Infof("[Migrations] Found %d migration files", len(files))

	// Track statistics
	applied := 0
	skipped := 0

	// Apply each migration
	for _, file := range files {
		version := filepath.Base(file)

		// Check if already applied
		if m.isMigrationApplied(version) {
			log.Debugf("[Migrations] Skipping %s (already applied)", version)
			skipped++
			continue
		}

		// Apply migration
		log.Infof("[Migrations] Applying %s...", version)
		if err := m.applyMigration(file, version); err != nil {
			// Note: We return immediately on failure, so failed count would always be 1
			// Keeping for potential future use where we might continue on certain errors
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}
		applied++
		log.Infof("[Migrations] Successfully applied %s", version)
	}

	// Log summary
	log.Infof("[Migrations] Migration complete - Applied: %d, Skipped: %d",
		applied, skipped)

	// Log current migration status
	m.logMigrationStatus()

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *MigrationRunner) ensureMigrationsTable() error {
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	return m.db.Exec(sql).Error
}

// getMigrationFiles returns a sorted list of migration SQL files
func (m *MigrationRunner) getMigrationFiles() ([]string, error) {
	// Check if migrations path exists
	if _, err := os.Stat(m.migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations path does not exist: %s", m.migrationsPath)
	}

	// Find all SQL files
	pattern := filepath.Join(m.migrationsPath, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Sort files to ensure consistent order
	sort.Strings(files)
	return files, nil
}

// isMigrationApplied checks if a migration version has already been applied
func (m *MigrationRunner) isMigrationApplied(version string) bool {
	var count int64
	m.db.Model(&SchemaMigration{}).Where("version = ?", version).Count(&count)
	return count > 0
}

// applyMigration reads, cleans, and applies a single migration file
func (m *MigrationRunner) applyMigration(filepath, version string) error {
	// Read migration file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Clean SQL if needed
	sql := string(content)
	if m.cleanSQL {
		sql = m.cleanSupabaseSQL(sql)
	}

	// Begin transaction for atomicity
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Apply migration SQL
	if err := tx.Exec(sql).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	migration := SchemaMigration{
		Version:    version,
		ExecutedAt: time.Now().UTC(),
	}
	if err := tx.Create(&migration).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// cleanSupabaseSQL removes Supabase-specific SQL commands
func (m *MigrationRunner) cleanSupabaseSQL(sql string) string {
	// Pattern to match GRANT statements for Supabase roles
	grantPattern := regexp.MustCompile(`(?i)(grant|GRANT)\s+.*\s+(to|TO)\s+(anon|authenticated|service_role).*?;`)

	// Pattern to match RLS (Row Level Security) commands - commented out for now as we may want to keep these
	// rlsPattern := regexp.MustCompile(`(?i)(alter table|ALTER TABLE)\s+.*\s+(enable|ENABLE)\s+(row level security|ROW LEVEL SECURITY).*?;`)

	// Pattern to match policy creation for Supabase roles
	policyPattern := regexp.MustCompile(`(?i)(create policy|CREATE POLICY)\s+.*\s+(for|FOR)\s+.*\s+(to|TO)\s+(anon|authenticated|service_role).*?;`)

	// Clean the SQL
	cleaned := sql

	// Remove GRANT statements
	cleaned = grantPattern.ReplaceAllString(cleaned, "")

	// Remove RLS statements (optional - you may want to keep these)
	// cleaned = rlsPattern.ReplaceAllString(cleaned, "")

	// Remove policy statements
	cleaned = policyPattern.ReplaceAllString(cleaned, "")

	// Remove "with schema extensions" clauses
	cleaned = strings.ReplaceAll(cleaned, ` with schema "extensions"`, "")
	cleaned = strings.ReplaceAll(cleaned, ` WITH SCHEMA "extensions"`, "")

	// Make CREATE EXTENSION idempotent
	cleaned = regexp.MustCompile(`(?i)create extension\s+`).ReplaceAllString(cleaned, "CREATE EXTENSION IF NOT EXISTS ")

	// Remove empty lines created by cleaning
	lines := strings.Split(cleaned, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" || strings.Contains(line, "--") { // Keep comments
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}

// logMigrationStatus logs the current migration status
func (m *MigrationRunner) logMigrationStatus() {
	var migrations []SchemaMigration
	m.db.Order("executed_at DESC").Limit(5).Find(&migrations)

	if len(migrations) > 0 {
		log.Info("[Migrations] Recent migrations:")
		for _, migration := range migrations {
			log.Infof("  - %s (applied at %s)", migration.Version, migration.ExecutedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Count total migrations
	var count int64
	m.db.Model(&SchemaMigration{}).Count(&count)
	log.Infof("[Migrations] Total migrations applied: %d", count)
}

// GetAppliedMigrations returns a list of all applied migrations
func (m *MigrationRunner) GetAppliedMigrations() ([]SchemaMigration, error) {
	var migrations []SchemaMigration
	err := m.db.Order("executed_at ASC").Find(&migrations).Error
	return migrations, err
}

// GetPendingMigrations returns a list of migration files that haven't been applied yet
func (m *MigrationRunner) GetPendingMigrations() ([]string, error) {
	files, err := m.getMigrationFiles()
	if err != nil {
		return nil, err
	}

	var pending []string
	for _, file := range files {
		version := filepath.Base(file)
		if !m.isMigrationApplied(version) {
			pending = append(pending, version)
		}
	}

	return pending, nil
}

// RollbackMigration attempts to rollback a specific migration (requires down migrations)
// Note: This is a placeholder as the current SQL migrations don't include rollback scripts
func (m *MigrationRunner) RollbackMigration(version string) error {
	// Check if migration exists
	if !m.isMigrationApplied(version) {
		return fmt.Errorf("migration %s has not been applied", version)
	}

	// Note: Actual rollback would require down migration scripts
	// For now, just remove the migration record (manual intervention needed for schema changes)
	if err := m.db.Delete(&SchemaMigration{Version: version}).Error; err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Warnf("[Migrations] Removed migration record for %s - manual schema rollback may be required", version)
	return nil
}
