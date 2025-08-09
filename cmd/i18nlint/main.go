package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	faintColor   = color.New(color.Faint).SprintFunc()
)

type LintIssue struct {
	Type        string // "missing_key", "unused_key", "hardcoded_string", "param_mismatch", "invalid_key"
	Severity    string // "error", "warning", "info"
	File        string
	Line        int
	Key         string
	Message     string
	Suggestion  string
	Context     string
}

type TranslationFile struct {
	Path         string
	Language     string
	Keys         map[string]string
	KeysWithParams map[string][]string
}

type LintReport struct {
	Issues              []LintIssue
	TranslationFiles    []TranslationFile
	CatalogKeys         map[string]bool
	CodeUsages          map[string][]Usage
	LanguageCoverage    map[string]float64
	UnusedKeys          []string
	MissingKeys         []string
	DuplicateKeys       []string
	TotalTranslations   int
	ProcessedFiles      int
	StartTime           time.Time
}

type Usage struct {
	File    string
	Line    int
	Context string
	Params  []string
}

// Main entry point
func main() {
	fmt.Println(headerColor("🔍 Alita Robot i18n Linter"))
	fmt.Println(headerColor("==========================="))
	
	report, err := runLinter()
	if err != nil {
		fmt.Printf("%s Error running linter: %v\n", errorColor("❌"), err)
		os.Exit(1)
	}
	
	printReport(report)
	generateReports(report)
	
	// Exit with appropriate code
	exitCode := calculateExitCode(report)
	if exitCode > 0 {
		fmt.Printf("\n%s Linting completed with issues (exit code: %d)\n", 
			warningColor("⚠️"), exitCode)
	} else {
		fmt.Printf("\n%s All i18n checks passed!\n", successColor("✅"))
	}
	
	os.Exit(exitCode)
}

func runLinter() (*LintReport, error) {
	report := &LintReport{
		StartTime:        time.Now(),
		CatalogKeys:      make(map[string]bool),
		CodeUsages:       make(map[string][]Usage),
		LanguageCoverage: make(map[string]float64),
	}
	
	fmt.Printf("%s Initializing linter...\n", infoColor("🚀"))
	
	// Step 1: Load all translation files
	if err := loadTranslationFiles(report); err != nil {
		return nil, fmt.Errorf("loading translations: %w", err)
	}
	
	// Step 2: Load catalog keys (if available)
	if err := loadCatalogKeys(report); err != nil {
		// Non-fatal, catalog might not exist yet
		fmt.Printf("%s Warning: Could not load catalog keys: %v\n", warningColor("⚠️"), err)
	}
	
	// Step 3: Analyze Go source code
	if err := analyzeSourceCode(report); err != nil {
		return nil, fmt.Errorf("analyzing source code: %w", err)
	}
	
	// Step 4: Run validation checks
	runValidationChecks(report)
	
	// Step 5: Calculate coverage metrics
	calculateCoverageMetrics(report)
	
	return report, nil
}

func loadTranslationFiles(report *LintReport) error {
	fmt.Printf("%s Loading translation files...\n", infoColor("📚"))
	
	localesDir := "locales"
	return filepath.Walk(localesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(strings.ToLower(path), ".yml") && 
		   !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}
		
		// Skip config files
		if strings.Contains(strings.ToLower(filepath.Base(path)), "config") {
			return nil
		}
		
		translationFile, err := parseTranslationFile(path)
		if err != nil {
			report.Issues = append(report.Issues, LintIssue{
				Type:     "invalid_yaml",
				Severity: "error",
				File:     path,
				Message:  fmt.Sprintf("Failed to parse YAML: %v", err),
			})
			return nil // Continue with other files
		}
		
		report.TranslationFiles = append(report.TranslationFiles, *translationFile)
		report.TotalTranslations += len(translationFile.Keys)
		
		fmt.Printf("  %s %s: %d keys\n", 
			successColor("✓"), translationFile.Language, len(translationFile.Keys))
		
		return nil
	})
}

func parseTranslationFile(filePath string) (*TranslationFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	
	// Extract language from filename
	language := strings.TrimSuffix(filepath.Base(filePath), ".yml")
	language = strings.TrimSuffix(language, ".yaml")
	
	translationFile := &TranslationFile{
		Path:           filePath,
		Language:       language,
		Keys:           make(map[string]string),
		KeysWithParams: make(map[string][]string),
	}
	
	// Flatten the YAML structure
	flattenYAMLKeys("", content, translationFile)
	
	return translationFile, nil
}

func flattenYAMLKeys(prefix string, data map[string]interface{}, tf *TranslationFile) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			flattenYAMLKeys(fullKey, v, tf)
		case string:
			tf.Keys[fullKey] = v
			// Extract parameters from the translation
			params := extractParameters(v)
			if len(params) > 0 {
				tf.KeysWithParams[fullKey] = params
			}
		}
	}
}

func extractParameters(text string) []string {
	// Match {param} style parameters
	paramRegex := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := paramRegex.FindAllStringSubmatch(text, -1)
	
	var params []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	
	// Also match %s, %d style parameters (legacy)
	printfRegex := regexp.MustCompile(`%[sd]`)
	printfMatches := printfRegex.FindAllString(text, -1)
	for i, match := range printfMatches {
		paramName := fmt.Sprintf("param%d_%s", i+1, strings.TrimPrefix(match, "%"))
		if !seen[paramName] {
			params = append(params, paramName)
			seen[paramName] = true
		}
	}
	
	return params
}

func loadCatalogKeys(report *LintReport) error {
	// Look for catalog registration patterns in Go files
	catalogDir := "alita/i18n/catalog"
	
	return filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Non-fatal
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil // Non-fatal
		}
		
		// Look for Register calls
		ast.Inspect(node, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if isRegisterCall(call) {
					key := extractKeyFromCall(call)
					if key != "" {
						report.CatalogKeys[key] = true
					}
				}
			}
			return true
		})
		
		return nil
	})
}

func isRegisterCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "Register" || sel.Sel.Name == "MustRegister"
	}
	
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "Register" || ident.Name == "MustRegister"
	}
	
	return false
}

func extractKeyFromCall(call *ast.CallExpr) string {
	if len(call.Args) > 0 {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			return strings.Trim(lit.Value, `"`)
		}
	}
	return ""
}

func analyzeSourceCode(report *LintReport) error {
	fmt.Printf("%s Analyzing source code...\n", infoColor("🔍"))
	
	modulesDir := "alita/modules"
	return filepath.Walk(modulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		if err := analyzeGoFile(path, report); err != nil {
			// Log error but continue
			report.Issues = append(report.Issues, LintIssue{
				Type:     "parse_error", 
				Severity: "warning",
				File:     path,
				Message:  fmt.Sprintf("Failed to parse Go file: %v", err),
			})
		}
		
		report.ProcessedFiles++
		return nil
	})
}

func analyzeGoFile(filePath string, report *LintReport) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			analyzeCallForI18n(x, fset, filePath, report)
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				analyzeStringLiteral(x, fset, filePath, report)
			}
		}
		return true
	})
	
	return nil
}

func analyzeCallForI18n(call *ast.CallExpr, fset *token.FileSet, filePath string, report *LintReport) {
	if isI18nGetStringCall(call) {
		pos := fset.Position(call.Pos())
		
		if len(call.Args) > 0 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				key := strings.Trim(lit.Value, `"`)
				
				usage := Usage{
					File:    filePath,
					Line:    pos.Line,
					Context: "tr.GetString()",
				}
				
				// Extract any additional parameters
				if len(call.Args) > 1 {
					// This would be for parameterized calls
					usage.Params = []string{"dynamic"} // Simplified
				}
				
				report.CodeUsages[key] = append(report.CodeUsages[key], usage)
			}
		}
	}
}

func isI18nGetStringCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "GetString"
	}
	return false
}

func analyzeStringLiteral(lit *ast.BasicLit, fset *token.FileSet, filePath string, report *LintReport) {
	content := strings.Trim(lit.Value, `"`)
	
	// Check if this looks like user-facing text that should be internationalized
	if shouldBeI18n(content) {
		pos := fset.Position(lit.Pos())
		
		report.Issues = append(report.Issues, LintIssue{
			Type:       "hardcoded_string",
			Severity:   "warning",
			File:       filePath,
			Line:       pos.Line,
			Message:    "Hardcoded string that should be internationalized",
			Suggestion: fmt.Sprintf("Consider moving to i18n: \"%s\"", content),
			Context:    content,
		})
	}
}

func shouldBeI18n(content string) bool {
	// Skip short strings and technical content
	if len(content) < 15 {
		return false
	}
	
	// Skip obvious technical patterns
	technicalPatterns := []string{
		`^[A-Z_]+$`,                              // CONSTANTS
		`^[a-z_]+$`,                              // identifiers
		`^[0-9\-\s]*$`,                           // numbers/dates
		`\.(jpg|png|gif|pdf|html|json|xml)$`,     // file extensions
		`^http[s]?://`,                           // URLs
		`^/[a-zA-Z]`,                             // commands
		`error|Error|failed|Failed`,              // error handling
		`debug|DEBUG|log|Log`,                    // logging
		`panic|Panic`,                            // panic messages
	}
	
	for _, pattern := range technicalPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return false
		}
	}
	
	// Look for user-facing characteristics
	userFacingPatterns := []string{
		`successfully|Successfully`,
		`welcome|Welcome|hello|Hello`,
		`banned|kicked|muted`,
		`admin|Admin|user|User`,
		`group|Group|chat|Chat`,
		`message|Message`,
		`permission|Permission`,
	}
	
	for _, pattern := range userFacingPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	
	// Check if it contains common sentence patterns
	if strings.Contains(content, " ") && 
	   (strings.Contains(content, "!") || 
	    strings.Contains(content, "?") ||
	    strings.Contains(content, ".")) {
		return true
	}
	
	return false
}

func runValidationChecks(report *LintReport) {
	fmt.Printf("%s Running validation checks...\n", infoColor("🔍"))
	
	// Check 1: Find unused translation keys
	findUnusedKeys(report)
	
	// Check 2: Find missing translation keys
	findMissingKeys(report)
	
	// Check 3: Check parameter consistency
	checkParameterConsistency(report)
	
	// Check 4: Find duplicate keys across languages
	findDuplicateKeys(report)
	
	// Check 5: Validate key naming conventions
	validateKeyNaming(report)
}

func findUnusedKeys(report *LintReport) {
	// Keys that exist in translations but aren't used in code
	for _, tf := range report.TranslationFiles {
		for key := range tf.Keys {
			if _, used := report.CodeUsages[key]; !used {
				// Also check if it's registered in catalog
				if !report.CatalogKeys[key] {
					report.UnusedKeys = append(report.UnusedKeys, key)
				}
			}
		}
	}
	
	// Deduplicate
	seen := make(map[string]bool)
	var uniqueUnused []string
	for _, key := range report.UnusedKeys {
		if !seen[key] {
			uniqueUnused = append(uniqueUnused, key)
			seen[key] = true
		}
	}
	report.UnusedKeys = uniqueUnused
	
	// Add issues
	for _, key := range report.UnusedKeys {
		report.Issues = append(report.Issues, LintIssue{
			Type:       "unused_key",
			Severity:   "info",
			Key:        key,
			Message:    "Translation key is not used in code",
			Suggestion: "Remove unused key or add usage in code",
		})
	}
}

func findMissingKeys(report *LintReport) {
	// Keys used in code but missing from translations
	for key := range report.CodeUsages {
		missing := false
		
		// Check if key exists in at least one translation file
		found := false
		for _, tf := range report.TranslationFiles {
			if _, exists := tf.Keys[key]; exists {
				found = true
				break
			}
		}
		
		if !found {
			missing = true
			report.MissingKeys = append(report.MissingKeys, key)
		}
		
		if missing {
			usages := report.CodeUsages[key]
			for _, usage := range usages {
				report.Issues = append(report.Issues, LintIssue{
					Type:       "missing_key",
					Severity:   "error",
					File:       usage.File,
					Line:       usage.Line,
					Key:        key,
					Message:    "Translation key not found in any language file",
					Suggestion: fmt.Sprintf("Add key '%s' to translation files", key),
					Context:    usage.Context,
				})
			}
		}
	}
}

func checkParameterConsistency(report *LintReport) {
	// Check if parameters are consistent across languages
	if len(report.TranslationFiles) < 2 {
		return // Need at least 2 languages to compare
	}
	
	// Get reference language (usually English)
	var refLang *TranslationFile
	for _, tf := range report.TranslationFiles {
		if tf.Language == "en" || tf.Language == "english" {
			refLang = &tf
			break
		}
	}
	
	if refLang == nil {
		refLang = &report.TranslationFiles[0] // Use first language as reference
	}
	
	// Compare parameters across languages
	for key, refParams := range refLang.KeysWithParams {
		for _, tf := range report.TranslationFiles {
			if tf.Language == refLang.Language {
				continue
			}
			
			if params, exists := tf.KeysWithParams[key]; exists {
				if !equalStringSlices(refParams, params) {
					report.Issues = append(report.Issues, LintIssue{
						Type:     "param_mismatch",
						Severity: "error",
						File:     tf.Path,
						Key:      key,
						Message:  fmt.Sprintf("Parameter mismatch in %s: expected %v, got %v", 
							tf.Language, refParams, params),
						Suggestion: "Ensure all languages have the same parameters",
						Context:    tf.Keys[key],
					})
				}
			}
		}
	}
}

func findDuplicateKeys(report *LintReport) {
	// This would find keys with identical values, potentially indicating copy-paste errors
	for _, tf := range report.TranslationFiles {
		valueToKeys := make(map[string][]string)
		
		for key, value := range tf.Keys {
			// Skip very short values
			if len(value) < 10 {
				continue
			}
			
			valueToKeys[value] = append(valueToKeys[value], key)
		}
		
		for value, keys := range valueToKeys {
			if len(keys) > 1 {
				report.Issues = append(report.Issues, LintIssue{
					Type:       "duplicate_value",
					Severity:   "warning",
					File:       tf.Path,
					Message:    fmt.Sprintf("Duplicate translation value in %d keys: %v", len(keys), keys),
					Suggestion: "Review if these should have different translations",
					Context:    value,
				})
			}
		}
	}
}

func validateKeyNaming(report *LintReport) {
	// Check key naming conventions
	for _, tf := range report.TranslationFiles {
		for key := range tf.Keys {
			issues := validateKeyName(key)
			for _, issue := range issues {
				report.Issues = append(report.Issues, LintIssue{
					Type:       "naming_convention",
					Severity:   "info",
					File:       tf.Path,
					Key:        key,
					Message:    issue,
					Suggestion: "Follow key naming conventions",
				})
			}
		}
	}
}

func validateKeyName(key string) []string {
	var issues []string
	
	// Check for proper snake_case
	if strings.Contains(key, " ") {
		issues = append(issues, "Key contains spaces")
	}
	
	if strings.Contains(key, "--") {
		issues = append(issues, "Key contains double dashes")
	}
	
	// Check for mixed case inconsistencies
	if key != strings.ToLower(key) && !regexp.MustCompile(`^[a-z][a-zA-Z0-9_]*$`).MatchString(key) {
		issues = append(issues, "Key should use consistent casing (prefer snake_case)")
	}
	
	// Check for very long keys
	if len(key) > 50 {
		issues = append(issues, "Key is very long (>50 chars), consider shortening")
	}
	
	// Check for non-descriptive names
	nonDescriptive := []string{"temp", "tmp", "test", "foo", "bar", "baz"}
	keyLower := strings.ToLower(key)
	for _, bad := range nonDescriptive {
		if strings.Contains(keyLower, bad) {
			issues = append(issues, fmt.Sprintf("Key contains non-descriptive term '%s'", bad))
			break
		}
	}
	
	return issues
}

func calculateCoverageMetrics(report *LintReport) {
	if len(report.TranslationFiles) == 0 {
		return
	}
	
	// Use first/reference language to calculate coverage
	var refKeys map[string]string
	refLang := "en"
	
	for _, tf := range report.TranslationFiles {
		if tf.Language == refLang {
			refKeys = tf.Keys
			break
		}
	}
	
	if refKeys == nil && len(report.TranslationFiles) > 0 {
		refKeys = report.TranslationFiles[0].Keys
		refLang = report.TranslationFiles[0].Language
	}
	
	totalKeys := len(refKeys)
	
	for _, tf := range report.TranslationFiles {
		if tf.Language == refLang {
			report.LanguageCoverage[tf.Language] = 100.0
			continue
		}
		
		matchingKeys := 0
		for key := range refKeys {
			if _, exists := tf.Keys[key]; exists {
				matchingKeys++
			}
		}
		
		coverage := 0.0
		if totalKeys > 0 {
			coverage = float64(matchingKeys) / float64(totalKeys) * 100.0
		}
		
		report.LanguageCoverage[tf.Language] = coverage
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	// Sort both slices for comparison
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	sort.Strings(aCopy)
	sort.Strings(bCopy)
	
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	
	return true
}

func printReport(report *LintReport) {
	duration := time.Since(report.StartTime)
	
	fmt.Printf("\n%s Lint Report\n", headerColor("📋"))
	fmt.Printf("=============\n\n")
	
	// Summary statistics
	fmt.Printf("%s Summary:\n", headerColor("📊"))
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Printf("  Files processed: %d Go files\n", report.ProcessedFiles)
	fmt.Printf("  Translation files: %d\n", len(report.TranslationFiles))
	fmt.Printf("  Total translations: %d\n", report.TotalTranslations)
	fmt.Printf("  Issues found: %d\n", len(report.Issues))
	
	// Issue breakdown
	issueCounts := make(map[string]int)
	severityCounts := make(map[string]int)
	
	for _, issue := range report.Issues {
		issueCounts[issue.Type]++
		severityCounts[issue.Severity]++
	}
	
	if len(issueCounts) > 0 {
		fmt.Printf("\n%s Issue Breakdown:\n", headerColor("🔍"))
		for issueType, count := range issueCounts {
			fmt.Printf("  %s: %d\n", issueType, count)
		}
		
		fmt.Printf("\n%s Severity Breakdown:\n", headerColor("⚠️"))
		for severity, count := range severityCounts {
			var colorFunc func(...interface{}) string
			switch severity {
			case "error":
				colorFunc = errorColor
			case "warning":
				colorFunc = warningColor
			default:
				colorFunc = infoColor
			}
			fmt.Printf("  %s: %d\n", colorFunc(severity), count)
		}
	}
	
	// Language coverage
	if len(report.LanguageCoverage) > 0 {
		fmt.Printf("\n%s Language Coverage:\n", headerColor("🌍"))
		
		// Sort languages by coverage
		type langCoverage struct {
			lang     string
			coverage float64
		}
		var langCoverages []langCoverage
		for lang, coverage := range report.LanguageCoverage {
			langCoverages = append(langCoverages, langCoverage{lang, coverage})
		}
		sort.Slice(langCoverages, func(i, j int) bool {
			return langCoverages[i].coverage > langCoverages[j].coverage
		})
		
		for _, lc := range langCoverages {
			var colorFunc func(...interface{}) string
			switch {
			case lc.coverage >= 95:
				colorFunc = successColor
			case lc.coverage >= 80:
				colorFunc = warningColor
			default:
				colorFunc = errorColor
			}
			fmt.Printf("  %s: %s\n", lc.lang, colorFunc(fmt.Sprintf("%.1f%%", lc.coverage)))
		}
	}
	
	// Show most critical issues
	if len(report.Issues) > 0 {
		fmt.Printf("\n%s Critical Issues:\n", headerColor("🚨"))
		
		errorCount := 0
		for _, issue := range report.Issues {
			if issue.Severity == "error" {
				if errorCount >= 10 { // Limit output
					fmt.Printf("  ... and %d more errors\n", severityCounts["error"]-errorCount)
					break
				}
				
				fmt.Printf("  %s %s:%d - %s\n", 
					errorColor("❌"), issue.File, issue.Line, issue.Message)
				if issue.Suggestion != "" {
					fmt.Printf("    %s %s\n", faintColor("💡"), issue.Suggestion)
				}
				errorCount++
			}
		}
	}
	
	// Quick stats
	fmt.Printf("\n%s Quick Stats:\n", headerColor("📈"))
	fmt.Printf("  Unused keys: %d\n", len(report.UnusedKeys))
	fmt.Printf("  Missing keys: %d\n", len(report.MissingKeys))
	fmt.Printf("  Code usages tracked: %d\n", len(report.CodeUsages))
}

func generateReports(report *LintReport) {
	fmt.Printf("\n%s Generating detailed reports...\n", infoColor("📝"))
	
	outputDir := "generated/i18n"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("%s Error creating output directory: %v\n", errorColor("❌"), err)
		return
	}
	
	// Generate detailed report
	if err := generateDetailedReport(filepath.Join(outputDir, "lint_report.md"), report); err != nil {
		fmt.Printf("%s Error generating detailed report: %v\n", errorColor("❌"), err)
	}
	
	// Generate coverage report
	if err := generateCoverageReport(filepath.Join(outputDir, "coverage_report.md"), report); err != nil {
		fmt.Printf("%s Error generating coverage report: %v\n", errorColor("❌"), err)
	}
	
	// Generate JSON report for CI integration
	if err := generateJSONReport(filepath.Join(outputDir, "lint_report.json"), report); err != nil {
		fmt.Printf("%s Error generating JSON report: %v\n", errorColor("❌"), err)
	}
}

func generateDetailedReport(filename string, report *LintReport) error {
	var content strings.Builder
	
	content.WriteString("# i18n Lint Report\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Duration: %v\n\n", time.Since(report.StartTime)))
	
	// Summary
	content.WriteString("## Summary\n\n")
	content.WriteString(fmt.Sprintf("- Files processed: %d\n", report.ProcessedFiles))
	content.WriteString(fmt.Sprintf("- Translation files: %d\n", len(report.TranslationFiles)))
	content.WriteString(fmt.Sprintf("- Total issues: %d\n", len(report.Issues)))
	
	// Issues by type
	if len(report.Issues) > 0 {
		content.WriteString("\n## Issues by Type\n\n")
		
		issuesByType := make(map[string][]LintIssue)
		for _, issue := range report.Issues {
			issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
		}
		
		for issueType, issues := range issuesByType {
			titleCase := strings.ReplaceAll(issueType, "_", " ")
			if len(titleCase) > 0 {
				titleCase = strings.ToUpper(titleCase[:1]) + titleCase[1:]
			}
			content.WriteString(fmt.Sprintf("### %s (%d issues)\n\n", titleCase, len(issues)))
			
			for _, issue := range issues {
				content.WriteString(fmt.Sprintf("- **%s** ", issue.Severity))
				if issue.File != "" {
					content.WriteString(fmt.Sprintf("`%s:%d` - ", issue.File, issue.Line))
				}
				content.WriteString(fmt.Sprintf("%s\n", issue.Message))
				
				if issue.Key != "" {
					content.WriteString(fmt.Sprintf("  - Key: `%s`\n", issue.Key))
				}
				if issue.Suggestion != "" {
					content.WriteString(fmt.Sprintf("  - Suggestion: %s\n", issue.Suggestion))
				}
				if issue.Context != "" && len(issue.Context) < 100 {
					content.WriteString(fmt.Sprintf("  - Context: `%s`\n", issue.Context))
				}
				content.WriteString("\n")
			}
		}
	}
	
	// Missing keys
	if len(report.MissingKeys) > 0 {
		content.WriteString("\n## Missing Translation Keys\n\n")
		for _, key := range report.MissingKeys {
			content.WriteString(fmt.Sprintf("- `%s`\n", key))
		}
	}
	
	// Unused keys
	if len(report.UnusedKeys) > 0 {
		content.WriteString("\n## Unused Translation Keys\n\n")
		for _, key := range report.UnusedKeys {
			content.WriteString(fmt.Sprintf("- `%s`\n", key))
		}
	}
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func generateCoverageReport(filename string, report *LintReport) error {
	var content strings.Builder
	
	content.WriteString("# i18n Coverage Report\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	
	// Language coverage table
	content.WriteString("## Language Coverage\n\n")
	content.WriteString("| Language | Coverage | Keys | Status |\n")
	content.WriteString("|----------|----------|------|---------|\n")
	
	for _, tf := range report.TranslationFiles {
		coverage := report.LanguageCoverage[tf.Language]
		status := "✅ Complete"
		if coverage < 95 {
			status = "⚠️ Incomplete"
		}
		if coverage < 50 {
			status = "❌ Poor"
		}
		
		content.WriteString(fmt.Sprintf("| %s | %.1f%% | %d | %s |\n", 
			tf.Language, coverage, len(tf.Keys), status))
	}
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func generateJSONReport(filename string, report *LintReport) error {
	// This would generate a structured JSON report for CI integration
	// Simplified version - in practice you'd use json.Marshal
	content := fmt.Sprintf(`{
  "timestamp": "%s",
  "duration_ms": %d,
  "files_processed": %d,
  "total_issues": %d,
  "error_count": %d,
  "warning_count": %d,
  "unused_keys": %d,
  "missing_keys": %d
}`,
		time.Now().Format(time.RFC3339),
		time.Since(report.StartTime).Milliseconds(),
		report.ProcessedFiles,
		len(report.Issues),
		countIssuesBySeverity(report.Issues, "error"),
		countIssuesBySeverity(report.Issues, "warning"),
		len(report.UnusedKeys),
		len(report.MissingKeys),
	)
	
	return os.WriteFile(filename, []byte(content), 0644)
}

func countIssuesBySeverity(issues []LintIssue, severity string) int {
	count := 0
	for _, issue := range issues {
		if issue.Severity == severity {
			count++
		}
	}
	return count
}

func calculateExitCode(report *LintReport) int {
	errorCount := countIssuesBySeverity(report.Issues, "error")
	if errorCount > 0 {
		return 1
	}
	return 0
}