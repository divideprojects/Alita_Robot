package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	// "sort" // Uncomment when using preview functions
	"strings"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

var (
	// Color functions for output
	successColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	errorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	warningColor = color.New(color.FgYellow, color.Bold).SprintFunc()
	infoColor    = color.New(color.FgCyan).SprintFunc()
	headerColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

type FlattenerStats struct {
	ProcessedFiles int
	BackupFiles    int
	OriginalKeys   int
	FlatKeys       int
	SavedBytes     int
}

type TranslationEntry struct {
	Key         string
	Value       string
	OriginalKey string
	Language    string
}

func main() {
	fmt.Println(headerColor("🔧 Alita Robot YAML Flattening Tool"))
	fmt.Println(headerColor("===================================="))
	
	stats, err := flattenAllYAMLFiles()
	if err != nil {
		fmt.Printf("%s Error flattening YAML files: %v\n", errorColor("❌"), err)
		os.Exit(1)
	}
	
	printStats(stats)
	fmt.Printf("\n%s YAML flattening complete!\n", successColor("✅"))
}

func flattenAllYAMLFiles() (*FlattenerStats, error) {
	stats := &FlattenerStats{}
	localesDir := "locales"
	
	fmt.Printf("%s Scanning %s directory...\n", infoColor("📂"), localesDir)
	
	err := filepath.Walk(localesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Process only YAML files
		if !isYAMLFile(path) {
			return nil
		}
		
		fmt.Printf("%s Processing %s...\n", infoColor("🔄"), path)
		
		if err := flattenYAMLFile(path, stats); err != nil {
			return fmt.Errorf("processing %s: %w", path, err)
		}
		
		return nil
	})
	
	return stats, err
}

func isYAMLFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".yml") || 
		   strings.HasSuffix(strings.ToLower(path), ".yaml")
}

func flattenYAMLFile(filePath string, stats *FlattenerStats) error {
	// Read the original file
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	
	// Parse YAML
	var content map[string]interface{}
	if err := yaml.Unmarshal(originalData, &content); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}
	
	// Create backup
	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("  %s Created backup: %s\n", successColor("💾"), backupPath)
	stats.BackupFiles++
	
	// Count original keys
	originalKeyCount := countNestedKeys(content)
	stats.OriginalKeys += originalKeyCount
	
	// Flatten the structure
	flattened := make(map[string]interface{})
	flattenMap("", content, flattened)
	
	// Transform keys to flat format
	transformed := transformKeys(flattened)
	
	// Count flattened keys
	stats.FlatKeys += len(transformed)
	
	// Generate the new YAML content
	newData, err := yaml.Marshal(transformed)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	
	// Add header comment
	header := fmt.Sprintf("# Flattened YAML structure - Generated: %s\n# Original file backed up as: %s\n\n",
		time.Now().Format("2006-01-02 15:04:05"),
		filepath.Base(backupPath))
	
	finalData := []byte(header + string(newData))
	
	// Write the flattened file
	if err := os.WriteFile(filePath, finalData, 0644); err != nil {
		return fmt.Errorf("writing flattened file: %w", err)
	}
	
	// Calculate size difference
	sizeDiff := len(originalData) - len(finalData)
	stats.SavedBytes += sizeDiff
	
	fmt.Printf("  %s Flattened %d → %d keys (saved %d bytes)\n", 
		successColor("✨"), originalKeyCount, len(transformed), sizeDiff)
	
	stats.ProcessedFiles++
	return nil
}

func countNestedKeys(data map[string]interface{}) int {
	count := 0
	for _, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			count += countNestedKeys(v)
		default:
			count++
		}
	}
	return count
}

func flattenMap(prefix string, data map[string]interface{}, result map[string]interface{}) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(fullKey, v, result)
		default:
			result[fullKey] = value
		}
	}
}

func transformKeys(flattened map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})
	
	for originalKey, value := range flattened {
		flatKey := transformKeyToFlat(originalKey)
		transformed[flatKey] = value
	}
	
	return transformed
}

func transformKeyToFlat(originalKey string) string {
	// Transform nested keys like "strings.Admin.promote.success" to "admin.promote_success"
	
	// Remove "strings." prefix if present
	key := strings.TrimPrefix(originalKey, "strings.")
	key = strings.TrimPrefix(key, "main.")
	
	// Split into parts
	parts := strings.Split(key, ".")
	
	// Transform each part
	var transformedParts []string
	for _, part := range parts {
		transformed := transformKeyPart(part)
		if transformed != "" {
			transformedParts = append(transformedParts, transformed)
		}
	}
	
	// Join with underscores for better flat structure
	result := strings.Join(transformedParts, "_")
	
	// Convert to lowercase for consistency
	result = strings.ToLower(result)
	
	// Clean up multiple underscores
	result = regexp.MustCompile(`_+`).ReplaceAllString(result, "_")
	result = strings.Trim(result, "_")
	
	// If result is empty or too generic, use original
	if result == "" || result == "strings" || result == "main" {
		result = strings.ToLower(strings.ReplaceAll(originalKey, ".", "_"))
	}
	
	return result
}

func transformKeyPart(part string) string {
	// Handle special cases and convert camelCase to snake_case
	
	// Skip empty parts
	if part == "" {
		return ""
	}
	
	// Convert camelCase to snake_case
	result := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(part, "${1}_${2}")
	
	// Convert to lowercase
	result = strings.ToLower(result)
	
	// Handle special mappings for common patterns
	specialMappings := map[string]string{
		"admin":           "admin",
		"adminlist":       "admin_list", 
		"antiflood":       "antiflood",
		"checkflood":      "check_flood",
		"setflood":        "set_flood",
		"setfloodmode":    "set_flood_mode",
		"anon_admin":      "anon_admin",
		"is_admin":        "is_admin",
		"is_bot_itself":   "is_bot_itself",
		"is_owner":        "is_owner",
		"success_promote": "success_promote",
		"success_demote":  "success_demote",
		"user_approved":   "user_approved",
		"user_unapproved": "user_unapproved",
		"bl_watcher":      "blacklist_watcher",
		"common_strings":  "common",
	}
	
	if mapped, exists := specialMappings[result]; exists {
		return mapped
	}
	
	return result
}

func printStats(stats *FlattenerStats) {
	fmt.Printf("\n%s Flattening Statistics\n", headerColor("📊"))
	fmt.Printf("=======================\n\n")
	
	fmt.Printf("%s Processed files: %d\n", infoColor("📁"), stats.ProcessedFiles)
	fmt.Printf("%s Backup files created: %d\n", infoColor("💾"), stats.BackupFiles)
	fmt.Printf("%s Original nested keys: %d\n", infoColor("🔑"), stats.OriginalKeys)
	fmt.Printf("%s Flattened keys: %d\n", infoColor("🎯"), stats.FlatKeys)
	
	if stats.SavedBytes > 0 {
		fmt.Printf("%s Space saved: %d bytes (%.1f KB)\n", 
			successColor("💾"), stats.SavedBytes, float64(stats.SavedBytes)/1024)
	} else {
		fmt.Printf("%s Space difference: %d bytes\n", 
			infoColor("📏"), -stats.SavedBytes)
	}
	
	if stats.ProcessedFiles > 0 {
		avgKeysPerFile := float64(stats.FlatKeys) / float64(stats.ProcessedFiles)
		fmt.Printf("%s Average keys per file: %.1f\n", infoColor("📈"), avgKeysPerFile)
	}
	
	// Show transformation examples
	fmt.Printf("\n%s Key Transformation Examples:\n", headerColor("🔄"))
	examples := map[string]string{
		"strings.Admin.promote.success":            "admin_promote_success",
		"strings.Antiflood.setflood.success":       "antiflood_set_flood_success", 
		"strings.Bans.ban.normal_ban":              "bans_ban_normal_ban",
		"strings.CommonStrings.admin_cache.loaded": "common_admin_cache_loaded",
		"main.language_name":                       "language_name",
		"main.language_flag":                       "language_flag",
	}
	
	for original, flattened := range examples {
		fmt.Printf("  %s → %s\n", 
			warningColor(original), successColor(flattened))
	}
	
	// Show backup information
	if stats.BackupFiles > 0 {
		fmt.Printf("\n%s Backup Information:\n", warningColor("⚠️"))
		fmt.Printf("  Original files are backed up with timestamp suffix\n")
		fmt.Printf("  To restore: mv <file>.backup.<timestamp> <file>\n")
		fmt.Printf("  To clean up: rm locales/*.backup.*\n")
	}
}

// Additional utility functions for advanced flattening

// generateKeyMappingReport creates a detailed report of key mappings
// Commented out as it's not currently used but may be useful in future
/*
func generateKeyMappingReport(stats *FlattenerStats) error {
	fmt.Printf("\n%s Generating key mapping report...\n", infoColor("📝"))
	
	mappingFile := "generated/i18n/key_mapping.md"
	if err := os.MkdirAll("generated/i18n", 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	
	var content strings.Builder
	content.WriteString("# YAML Key Mapping Report\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	
	content.WriteString("## Summary\n\n")
	content.WriteString(fmt.Sprintf("- Files processed: %d\n", stats.ProcessedFiles))
	content.WriteString(fmt.Sprintf("- Backup files: %d\n", stats.BackupFiles))
	content.WriteString(fmt.Sprintf("- Original keys: %d\n", stats.OriginalKeys))
	content.WriteString(fmt.Sprintf("- Flattened keys: %d\n", stats.FlatKeys))
	
	content.WriteString("\n## Transformation Rules\n\n")
	content.WriteString("1. Remove `strings.` and `main.` prefixes\n")
	content.WriteString("2. Convert nested dots to underscores\n") 
	content.WriteString("3. Convert camelCase to snake_case\n")
	content.WriteString("4. Normalize to lowercase\n")
	content.WriteString("5. Clean up multiple underscores\n")
	
	content.WriteString("\n## Example Transformations\n\n")
	content.WriteString("| Original Key | Flattened Key |\n")
	content.WriteString("|--------------|---------------|\n")
	
	examples := [][]string{
		{"strings.Admin.promote.success", "admin_promote_success"},
		{"strings.Antiflood.setflood.disabled", "antiflood_set_flood_disabled"},
		{"strings.Bans.ban.is_admin", "bans_ban_is_admin"},
		{"strings.CommonStrings.admin_cache.loaded", "common_admin_cache_loaded"},
		{"main.language_name", "language_name"},
	}
	
	for _, example := range examples {
		content.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", example[0], example[1]))
	}
	
	content.WriteString("\n## Notes\n\n")
	content.WriteString("- All original files are backed up with timestamp suffix\n")
	content.WriteString("- Use `git diff` to review changes before committing\n")
	content.WriteString("- Update Go code to use new flat key structure\n")
	
	if err := os.WriteFile(mappingFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("writing mapping report: %w", err)
	}
	
	fmt.Printf("%s Generated mapping report: %s\n", successColor("✅"), mappingFile)
	return nil
}
*/

// Validation functions
// validateYAMLStructure checks if a YAML file is valid
// Commented out as it's not currently used
/*
func validateYAMLStructure(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	var content map[string]interface{}
	return yaml.Unmarshal(data, &content)
}
*/

// findDuplicateKeys detects duplicate keys in flattened map
// Commented out as it's not currently used
/*
func findDuplicateKeys(flattened map[string]interface{}) []string {
	// This would be useful to detect key conflicts during flattening
	var duplicates []string
	keyCount := make(map[string]int)
	
	for key := range flattened {
		keyCount[key]++
		if keyCount[key] > 1 {
			duplicates = append(duplicates, key)
		}
	}
	
	return duplicates
}
*/

// previewFlattening shows what would be changed without modifying files
// Commented out as it's not currently used but may be useful for debugging
/*
func previewFlattening(filePath string) error {
	fmt.Printf("%s Preview mode for %s:\n", warningColor("👀"), filePath)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return err
	}
	
	flattened := make(map[string]interface{})
	flattenMap("", content, flattened)
	
	transformed := transformKeys(flattened)
	
	// Show first few transformations as preview
	var keys []string
	for key := range transformed {
		keys = append(keys, key)
	}
	// sort.Strings(keys) // Commented out - uncomment when sort is imported
	
	fmt.Printf("  Would create %d flattened keys:\n", len(keys))
	for i, key := range keys {
		if i >= 10 { // Show first 10
			fmt.Printf("  ... and %d more\n", len(keys)-i)
			break
		}
		fmt.Printf("    %s\n", key)
	}
	
	return nil
}
*/