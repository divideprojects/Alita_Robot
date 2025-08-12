package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Standalone version of splitSQLStatements for testing
func splitSQLStatements(sql string) []string {
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

// Standalone version of cleanSupabaseSQL for testing
func cleanSupabaseSQL(sql string) string {
	// Pattern to match GRANT statements for Supabase roles
	grantPattern := regexp.MustCompile(`(?i)(grant|GRANT)\s+.*\s+(to|TO)\s+(anon|authenticated|service_role).*?;`)

	// Pattern to match policy creation for Supabase roles
	policyPattern := regexp.MustCompile(`(?i)(create policy|CREATE POLICY)\s+.*\s+(for|FOR)\s+.*\s+(to|TO)\s+(anon|authenticated|service_role).*?;`)

	// Clean the SQL
	cleaned := sql

	// Remove GRANT statements
	cleaned = grantPattern.ReplaceAllString(cleaned, "")

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

func main() {
	// Setup logging
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	fmt.Println("=== Testing SQL Statement Splitter ===\n")

	// Test basic cases
	testBasicCases()

	// Test with actual migration file
	testMigrationFile()
}

func testBasicCases() {
	fmt.Println("Testing basic SQL splitting cases:")
	fmt.Println("=" + strings.Repeat("=", 50))

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
			name: "Statement with escaped quotes",
			sql: `INSERT INTO test VALUES ('It''s a test');
CREATE TABLE test2 (id INT);`,
			expected: 2,
		},
		{
			name: "Multiple semicolons",
			sql: `CREATE TABLE test1 (id INT);;;
CREATE TABLE test2 (id INT);`,
			expected: 2,
		},
	}

	allPassed := true
	for _, tc := range testCases {
		statements := splitSQLStatements(tc.sql)
		if len(statements) != tc.expected {
			fmt.Printf("❌ Test '%s' FAILED: expected %d statements, got %d\n",
				tc.name, tc.expected, len(statements))
			for i, stmt := range statements {
				fmt.Printf("   Statement %d: %s\n", i+1, truncate(stmt, 60))
			}
			allPassed = false
		} else {
			fmt.Printf("✅ Test '%s' PASSED: %d statements\n", tc.name, len(statements))
		}
	}

	if allPassed {
		fmt.Println("\n✅ All basic tests passed!")
	} else {
		fmt.Println("\n⚠️  Some tests failed")
	}
	fmt.Println()
}

func testMigrationFile() {
	fmt.Println("Testing with actual migration file:")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Read the first migration file
	migrationPath := "supabase/migrations/20250805200527_initial_migration.sql"
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		fmt.Printf("❌ Failed to read migration file: %v\n", err)
		return
	}

	fmt.Printf("📁 Original file size: %d bytes\n", len(content))

	// Clean the SQL
	cleaned := cleanSupabaseSQL(string(content))
	fmt.Printf("🧹 Cleaned SQL size: %d bytes (removed %d bytes)\n",
		len(cleaned), len(content)-len(cleaned))

	// Split into statements
	statements := splitSQLStatements(cleaned)
	fmt.Printf("📊 Number of statements: %d\n", len(statements))

	// Analyze statement types
	statementTypes := make(map[string]int)
	for _, stmt := range statements {
		stmtType := getStatementType(stmt)
		statementTypes[stmtType]++
	}

	fmt.Println("\n📈 Statement type breakdown:")
	totalStmts := 0
	for stmtType, count := range statementTypes {
		if stmtType != "GRANT" { // Skip GRANT statements as they should be removed
			fmt.Printf("   %s: %d\n", stmtType, count)
			totalStmts += count
		}
	}

	// Verify no GRANT statements remain
	if grantCount, exists := statementTypes["GRANT"]; exists && grantCount > 0 {
		fmt.Printf("\n⚠️  WARNING: Found %d GRANT statements that should have been removed!\n", grantCount)
	} else {
		fmt.Println("\n✅ No GRANT statements found (correctly cleaned)")
	}

	// Show sample statements
	fmt.Println("\n📝 Sample statements (first 5):")
	for i := 0; i < 5 && i < len(statements); i++ {
		fmt.Printf("   %d. %s\n", i+1, truncate(statements[i], 70))
	}

	fmt.Printf("\n✅ Successfully processed migration file with %d statements\n", len(statements))
}

func getStatementType(stmt string) string {
	stmt = strings.ToUpper(strings.TrimSpace(stmt))
	switch {
	case strings.HasPrefix(stmt, "CREATE SEQUENCE"):
		return "CREATE SEQUENCE"
	case strings.HasPrefix(stmt, "CREATE TABLE"):
		return "CREATE TABLE"
	case strings.HasPrefix(stmt, "CREATE UNIQUE INDEX"):
		return "CREATE UNIQUE INDEX"
	case strings.HasPrefix(stmt, "CREATE INDEX"):
		return "CREATE INDEX"
	case strings.HasPrefix(stmt, "ALTER SEQUENCE"):
		return "ALTER SEQUENCE"
	case strings.HasPrefix(stmt, "ALTER TABLE"):
		return "ALTER TABLE"
	case strings.HasPrefix(stmt, "GRANT"):
		return "GRANT"
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
