// Package main provides a test script for SQL statement splitting functionality
// Run this as: go run scripts/test_sql_splitter.go
package main

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

	// Test SQL splitter with various edge cases
	testSQLSplitter()

	// Test with actual migration file
	testMigrationFile()
}

func testSQLSplitter() {
	log.Info("Testing SQL splitter with edge cases...")

	// Create a dummy database connection (we only need it for the MigrationRunner)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Warn("DATABASE_URL not set, using dummy connection for testing")
		databaseURL = "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
	}

	database, _ := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	runner := db.NewMigrationRunner(database)

	// Test cases
	testCases := []struct {
		name     string
		sql      string
		expected int
	}{
		{
			name: "Simple statements",
			sql: `CREATE TABLE test1 (id INT);
CREATE TABLE test2 (id INT);`,
			expected: 2,
		},
		{
			name: "Statement with single quotes",
			sql: `INSERT INTO test VALUES ('value; with semicolon');
CREATE TABLE test2 (id INT);`,
			expected: 2,
		},
		{
			name: "Statement with double quotes",
			sql: `CREATE TABLE "test;table" (id INT);
INSERT INTO test VALUES (1);`,
			expected: 2,
		},
		{
			name: "Statement with line comment",
			sql: `-- This is a comment; with semicolon
CREATE TABLE test (id INT);
INSERT INTO test VALUES (1);`,
			expected: 2,
		},
		{
			name: "Statement with block comment",
			sql: `/* This is a block comment;
with semicolon */
CREATE TABLE test (id INT);
INSERT INTO test VALUES (1);`,
			expected: 2,
		},
		{
			name: "Dollar-quoted function",
			sql: `CREATE FUNCTION test() RETURNS void AS $$
BEGIN
  RAISE NOTICE 'test; with semicolon';
END;
$$ LANGUAGE plpgsql;
CREATE TABLE test (id INT);`,
			expected: 2,
		},
		{
			name: "Empty statements",
			sql: `CREATE TABLE test1 (id INT);
;
;
CREATE TABLE test2 (id INT);`,
			expected: 2,
		},
		{
			name: "Complex nested quotes",
			sql: `INSERT INTO test VALUES ('It''s a test');
INSERT INTO test VALUES ("double;quotes");
CREATE TABLE test (id INT);`,
			expected: 3,
		},
	}

	// Use reflection to access the private method for testing
	// In production, this would be tested through the public interface
	for _, tc := range testCases {
		// Since splitSQLStatements is now a method, we need to call it properly
		statements := runner.SplitSQLStatementsForTesting(tc.sql)
		if len(statements) != tc.expected {
			log.Errorf("Test '%s' failed: expected %d statements, got %d", tc.name, tc.expected, len(statements))
			for i, stmt := range statements {
				log.Debugf("  Statement %d: %s", i+1, truncate(stmt, 50))
			}
		} else {
			log.Infof("Test '%s' passed: %d statements", tc.name, len(statements))
		}
	}
}

func testMigrationFile() {
	log.Info("\nTesting with actual migration file...")

	// Read the first migration file
	migrationPath := "supabase/migrations/20250805200527_initial_migration.sql"
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Errorf("Failed to read migration file: %v", err)
		return
	}

	// Create a dummy runner
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
	}
	database, _ := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	runner := db.NewMigrationRunner(database)

	// Clean the SQL
	cleaned := runner.CleanSupabaseSQLForTesting(string(content))

	// Split into statements
	statements := runner.SplitSQLStatementsForTesting(cleaned)

	log.Infof("Original file size: %d bytes", len(content))
	log.Infof("Cleaned SQL size: %d bytes", len(cleaned))
	log.Infof("Number of statements: %d", len(statements))

	// Analyze statement types
	statementTypes := make(map[string]int)
	for _, stmt := range statements {
		stmtType := getStatementType(stmt)
		statementTypes[stmtType]++
	}

	log.Info("Statement types:")
	for stmtType, count := range statementTypes {
		log.Infof("  %s: %d", stmtType, count)
	}

	// Show first few statements as examples
	log.Info("\nFirst 5 statements (truncated):")
	for i := 0; i < 5 && i < len(statements); i++ {
		log.Infof("  %d: %s", i+1, truncate(statements[i], 80))
	}
}

func getStatementType(stmt string) string {
	stmt = strings.ToUpper(strings.TrimSpace(stmt))
	switch {
	case strings.HasPrefix(stmt, "CREATE SEQUENCE"):
		return "CREATE SEQUENCE"
	case strings.HasPrefix(stmt, "CREATE TABLE"):
		return "CREATE TABLE"
	case strings.HasPrefix(stmt, "CREATE INDEX"):
		return "CREATE INDEX"
	case strings.HasPrefix(stmt, "CREATE UNIQUE INDEX"):
		return "CREATE UNIQUE INDEX"
	case strings.HasPrefix(stmt, "ALTER SEQUENCE"):
		return "ALTER SEQUENCE"
	case strings.HasPrefix(stmt, "ALTER TABLE"):
		return "ALTER TABLE"
	case strings.HasPrefix(stmt, "GRANT"):
		return "GRANT"
	case strings.HasPrefix(stmt, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(stmt, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(stmt, "DELETE"):
		return "DELETE"
	default:
		return "OTHER"
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
