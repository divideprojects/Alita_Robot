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

// splitSQLStatements splits a SQL string into individual statements
// It handles various edge cases including:
// - Quoted strings (single quotes, double quotes)
// - Dollar-quoted strings (PostgreSQL specific)
// - Comments (single-line and multi-line)
// - Semicolons inside strings
func (m *MigrationRunner) splitSQLStatements(sql string) []string {
	var statements []string
	var currentStmt strings.Builder

	runes := []rune(sql)
	length := len(runes)

	inSingleQuote := false
	inDoubleQuote := false
	inDollarQuote := false
	inLineComment := false
	inBlockComment := false
	dollarQuoteTag := ""

	for i := 0; i < length; i++ {
		char := runes[i]
		nextChar := rune(0)
		if i+1 < length {
			nextChar = runes[i+1]
		}

		// Handle line comments
		if !inSingleQuote && !inDoubleQuote && !inDollarQuote && !inBlockComment {
			if char == '-' && nextChar == '-' {
				inLineComment = true
				currentStmt.WriteRune(char)
				continue
			}
		}

		if inLineComment {
			currentStmt.WriteRune(char)
			if char == '\n' {
				inLineComment = false
			}
			continue
		}

		// Handle block comments
		if !inSingleQuote && !inDoubleQuote && !inDollarQuote && !inLineComment {
			if char == '/' && nextChar == '*' {
				inBlockComment = true
				currentStmt.WriteRune(char)
				continue
			}
		}

		if inBlockComment {
			currentStmt.WriteRune(char)
			if char == '*' && nextChar == '/' {
				currentStmt.WriteRune(nextChar)
				i++
				inBlockComment = false
			}
			continue
		}

		// Handle dollar quotes (PostgreSQL)
		if !inSingleQuote && !inDoubleQuote && !inLineComment && !inBlockComment {
			if char == '$' {
				// Check if this is the start or end of a dollar quote
				tagEnd := i + 1
				for tagEnd < length && (runes[tagEnd] != '$' && runes[tagEnd] != ' ' && runes[tagEnd] != '\n' && runes[tagEnd] != ';') {
					tagEnd++
				}

				if tagEnd < length && runes[tagEnd] == '$' {
					tag := string(runes[i : tagEnd+1])
					if inDollarQuote {
						// Check if this closes the current dollar quote
						if tag == dollarQuoteTag {
							inDollarQuote = false
							dollarQuoteTag = ""
						}
					} else {
						// Start a new dollar quote
						inDollarQuote = true
						dollarQuoteTag = tag
					}

					// Add the entire tag to the current statement
					for j := i; j <= tagEnd; j++ {
						currentStmt.WriteRune(runes[j])
					}
					i = tagEnd
					continue
				}
			}
		}

		// Handle single quotes
		if !inDoubleQuote && !inDollarQuote && !inLineComment && !inBlockComment {
			if char == '\'' {
				// Check for escaped single quote
				if i+1 < length && runes[i+1] == '\'' {
					currentStmt.WriteRune(char)
					currentStmt.WriteRune(runes[i+1])
					i++
					continue
				}
				inSingleQuote = !inSingleQuote
			}
		}

		// Handle double quotes
		if !inSingleQuote && !inDollarQuote && !inLineComment && !inBlockComment {
			if char == '"' {
				// Check for escaped double quote
				if i+1 < length && runes[i+1] == '"' {
					currentStmt.WriteRune(char)
					currentStmt.WriteRune(runes[i+1])
					i++
					continue
				}
				inDoubleQuote = !inDoubleQuote
			}
		}

		// Handle semicolons (statement separator)
		if char == ';' && !inSingleQuote && !inDoubleQuote && !inDollarQuote && !inLineComment && !inBlockComment {
			// End of statement
			stmt := strings.TrimSpace(currentStmt.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStmt.Reset()
		} else {
			currentStmt.WriteRune(char)
		}
	}

	// Add any remaining statement
	if currentStmt.Len() > 0 {
		stmt := strings.TrimSpace(currentStmt.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
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

	// Split SQL into individual statements
	statements := m.splitSQLStatements(sql)
	if len(statements) == 0 {
		log.Warnf("[Migrations] No statements found in migration %s", version)
		return nil
	}

	log.Debugf("[Migrations] Migration %s contains %d statements", version, len(statements))

	// Begin transaction for atomicity
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Apply each statement individually
	for i, stmt := range statements {
		// Skip empty statements
		if strings.TrimSpace(stmt) == "" {
			continue
		}

		// Log progress for large migrations
		if len(statements) > 50 && i%50 == 0 {
			log.Debugf("[Migrations] Progress: %d/%d statements executed", i, len(statements))
		}

		// Execute the statement
		if err := tx.Exec(stmt).Error; err != nil {
			tx.Rollback()
			// Include statement preview in error for debugging
			preview := stmt
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			return fmt.Errorf("failed to execute statement %d/%d: %w\nStatement preview: %s",
				i+1, len(statements), err, preview)
		}
	}

	log.Debugf("[Migrations] All %d statements executed successfully", len(statements))

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
	// Pattern to match GRANT statements for Supabase roles (now handles quotes properly)
	grantPattern := regexp.MustCompile(`(?i)grant\s+[^;]+\s+to\s+["']?(anon|authenticated|service_role)["']?\s*;`)

	// Pattern to match RLS (Row Level Security) commands - commented out for now as we may want to keep these
	// rlsPattern := regexp.MustCompile(`(?i)(alter table|ALTER TABLE)\s+.*\s+(enable|ENABLE)\s+(row level security|ROW LEVEL SECURITY).*?;`)

	// Pattern to match policy creation for Supabase roles
	policyPattern := regexp.MustCompile(`(?i)create\s+policy\s+[^;]+\s+to\s+["']?(anon|authenticated|service_role)["']?\s*;`)

	// Clean the SQL
	cleaned := sql

	// Remove GRANT statements (handles both quoted and unquoted role names)
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

// SplitSQLStatementsForTesting exposes the SQL splitter for testing purposes
func (m *MigrationRunner) SplitSQLStatementsForTesting(sql string) []string {
	return m.splitSQLStatements(sql)
}

// CleanSupabaseSQLForTesting exposes the SQL cleaner for testing purposes
func (m *MigrationRunner) CleanSupabaseSQLForTesting(sql string) string {
	return m.cleanSupabaseSQL(sql)
}
