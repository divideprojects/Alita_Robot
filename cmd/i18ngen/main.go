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
)

type I18nUsage struct {
	File     string
	Line     int
	KeyPath  string
	Usage    string
	Context  string
}

type HardcodedString struct {
	File    string
	Line    int
	Content string
	Context string
}

type AnalysisReport struct {
	I18nUsages         []I18nUsage
	HardcodedStrings   []HardcodedString
	MissingKeys        []string
	ModuleCoverage     map[string]int // module name -> i18n usage count
	TotalFiles         int
	ProcessedFiles     int
	AvailableKeys      map[string]bool
}

func main() {
	fmt.Println(headerColor("🔍 Alita Robot i18n Code Generator"))
	fmt.Println(headerColor("====================================="))
	
	report, err := analyzeProject()
	if err != nil {
		fmt.Printf("%s Error analyzing project: %v\n", errorColor("❌"), err)
		os.Exit(1)
	}
	
	printReport(report)
	
	if err := generateRegistrationCode(report); err != nil {
		fmt.Printf("%s Error generating code: %v\n", errorColor("❌"), err)
		os.Exit(1)
	}
	
	fmt.Printf("\n%s Analysis complete! Generated registration code for %d i18n usages.\n", 
		successColor("✅"), len(report.I18nUsages))
}

func analyzeProject() (*AnalysisReport, error) {
	fmt.Print(infoColor("📂 Scanning project structure..."))
	
	report := &AnalysisReport{
		ModuleCoverage: make(map[string]int),
		AvailableKeys:  make(map[string]bool),
	}
	
	// Load existing translations to check for missing keys
	if err := loadAvailableKeys(report); err != nil {
		return nil, fmt.Errorf("loading available keys: %w", err)
	}
	
	// Analyze Go files in modules directory
	modulesDir := "alita/modules"
	err := filepath.Walk(modulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		report.TotalFiles++
		return analyzeGoFile(path, report)
	})
	
	if err != nil {
		return nil, fmt.Errorf("analyzing modules: %w", err)
	}
	
	fmt.Printf(" %s\n", successColor("Done"))
	
	// Check for missing keys
	for _, usage := range report.I18nUsages {
		if !report.AvailableKeys[usage.KeyPath] {
			report.MissingKeys = append(report.MissingKeys, usage.KeyPath)
		}
	}
	
	return report, nil
}

func loadAvailableKeys(report *AnalysisReport) error {
	localesDir := "locales"
	return filepath.Walk(localesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		var content map[string]interface{}
		if err := yaml.Unmarshal(data, &content); err != nil {
			return err
		}
		
		// Flatten YAML structure to extract keys
		flattenKeys("", content, report.AvailableKeys)
		return nil
	})
}

func flattenKeys(prefix string, data map[string]interface{}, keys map[string]bool) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			flattenKeys(fullKey, v, keys)
		case string:
			keys[fullKey] = true
		}
	}
}

func analyzeGoFile(filePath string, report *AnalysisReport) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	
	report.ProcessedFiles++
	moduleName := getModuleName(filePath)
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			analyzeCallExpression(x, fset, filePath, moduleName, report)
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				analyzeStringLiteral(x, fset, filePath, report)
			}
		}
		return true
	})
	
	return nil
}

func analyzeCallExpression(call *ast.CallExpr, fset *token.FileSet, filePath, moduleName string, report *AnalysisReport) {
	// Check for i18n usage patterns
	if isI18nCall(call) {
		usage := extractI18nUsage(call, fset, filePath, moduleName)
		if usage != nil {
			report.I18nUsages = append(report.I18nUsages, *usage)
			report.ModuleCoverage[moduleName]++
		}
	}
}

func isI18nCall(call *ast.CallExpr) bool {
	// Check for tr.GetString() calls
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "GetString"
	}
	return false
}

func extractI18nUsage(call *ast.CallExpr, fset *token.FileSet, filePath, moduleName string) *I18nUsage {
	if len(call.Args) == 0 {
		return nil
	}
	
	pos := fset.Position(call.Pos())
	
	// Extract key path from first argument
	var keyPath string
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		keyPath = strings.Trim(lit.Value, `"`)
	} else {
		return nil // Dynamic keys are harder to analyze
	}
	
	return &I18nUsage{
		File:     filePath,
		Line:     pos.Line,
		KeyPath:  keyPath,
		Usage:    "tr.GetString()",
		Context:  fmt.Sprintf("Module: %s", moduleName),
	}
}

func analyzeStringLiteral(lit *ast.BasicLit, fset *token.FileSet, filePath string, report *AnalysisReport) {
	content := strings.Trim(lit.Value, `"`)
	
	// Skip short strings, common patterns, and technical strings
	if shouldSkipString(content) {
		return
	}
	
	pos := fset.Position(lit.Pos())
	
	report.HardcodedStrings = append(report.HardcodedStrings, HardcodedString{
		File:    filePath,
		Line:    pos.Line,
		Content: content,
		Context: "Potential user-facing text",
	})
}

func shouldSkipString(content string) bool {
	// Skip empty or very short strings
	if len(content) < 10 {
		return true
	}
	
	// Skip technical patterns
	technicalPatterns := []string{
		"^[A-Z_]+$",                    // CONSTANTS
		"^[a-z_]+$",                    // snake_case identifiers
		"^[a-zA-Z_][a-zA-Z0-9_]*\\.",  // function calls
		"^/[a-zA-Z_]+",                 // command patterns
		"^[0-9]+",                      // numeric strings
		"\\.(jpg|png|gif|pdf|html)$",   // file extensions
		"^http[s]?://",                 // URLs
		"^[a-zA-Z_][a-zA-Z0-9_]*:",     // log fields
		"panic",                        // panic messages
		"error",                        // error handling
		"Failed to",                    // error prefixes
	}
	
	for _, pattern := range technicalPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	
	// Skip strings with only punctuation and common words
	commonWords := []string{
		"true", "false", "nil", "done", "ok", "error", "warning", "info", "debug",
		"admin", "user", "chat", "message", "command", "button", "callback",
	}
	
	lowerContent := strings.ToLower(content)
	for _, word := range commonWords {
		if lowerContent == word {
			return true
		}
	}
	
	return false
}

func getModuleName(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, ".go")
}

func printReport(report *AnalysisReport) {
	fmt.Printf("\n%s Analysis Report\n", headerColor("📊"))
	fmt.Printf("================\n\n")
	
	fmt.Printf("%s Files processed: %d/%d\n", infoColor("📁"), report.ProcessedFiles, report.TotalFiles)
	fmt.Printf("%s i18n usages found: %d\n", infoColor("🔧"), len(report.I18nUsages))
	fmt.Printf("%s Hardcoded strings: %d\n", infoColor("📝"), len(report.HardcodedStrings))
	fmt.Printf("%s Missing translation keys: %d\n", warningColor("⚠️"), len(report.MissingKeys))
	
	// Module coverage
	if len(report.ModuleCoverage) > 0 {
		fmt.Printf("\n%s Module i18n Usage Coverage:\n", headerColor("📈"))
		
		// Sort modules by usage count
		type moduleUsage struct {
			name  string
			count int
		}
		var modules []moduleUsage
		for name, count := range report.ModuleCoverage {
			modules = append(modules, moduleUsage{name, count})
		}
		sort.Slice(modules, func(i, j int) bool {
			return modules[i].count > modules[j].count
		})
		
		for _, mod := range modules {
			fmt.Printf("  %s: %d usages\n", infoColor(mod.name), mod.count)
		}
	}
	
	// Missing keys
	if len(report.MissingKeys) > 0 {
		fmt.Printf("\n%s Missing Translation Keys:\n", warningColor("🔍"))
		uniqueKeys := make(map[string]bool)
		for _, key := range report.MissingKeys {
			if !uniqueKeys[key] {
				fmt.Printf("  %s\n", errorColor(key))
				uniqueKeys[key] = true
			}
		}
	}
	
	// Sample hardcoded strings
	if len(report.HardcodedStrings) > 0 {
		fmt.Printf("\n%s Sample Hardcoded Strings (consider i18n):\n", warningColor("💬"))
		count := 0
		for _, str := range report.HardcodedStrings {
			if count >= 10 { // Show only first 10
				fmt.Printf("  ... and %d more\n", len(report.HardcodedStrings)-count)
				break
			}
			fmt.Printf("  %s:%d: %s\n", infoColor(str.File), str.Line, 
				strings.TrimSpace(str.Content))
			count++
		}
	}
}

func generateRegistrationCode(report *AnalysisReport) error {
	fmt.Printf("\n%s Generating registration code...\n", headerColor("🔧"))
	
	// Group usages by module
	moduleUsages := make(map[string][]I18nUsage)
	for _, usage := range report.I18nUsages {
		moduleName := getModuleName(usage.File)
		moduleUsages[moduleName] = append(moduleUsages[moduleName], usage)
	}
	
	// Create output directory
	outputDir := "generated/i18n"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	
	// Generate consolidated report
	reportFile := filepath.Join(outputDir, "analysis_report.md")
	if err := generateMarkdownReport(reportFile, report); err != nil {
		return fmt.Errorf("generating markdown report: %w", err)
	}
	
	// Generate Go registration code for each module
	for moduleName, usages := range moduleUsages {
		if err := generateModuleRegistration(outputDir, moduleName, usages); err != nil {
			return fmt.Errorf("generating registration for %s: %w", moduleName, err)
		}
	}
	
	fmt.Printf("%s Generated files:\n", successColor("✅"))
	fmt.Printf("  📄 %s\n", reportFile)
	for moduleName := range moduleUsages {
		fmt.Printf("  📄 %s\n", filepath.Join(outputDir, moduleName+"_messages.go"))
	}
	
	return nil
}

func generateMarkdownReport(filename string, report *AnalysisReport) error {
	var content strings.Builder
	
	content.WriteString("# i18n Analysis Report\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	
	content.WriteString("## Summary\n\n")
	content.WriteString(fmt.Sprintf("- Files processed: %d/%d\n", report.ProcessedFiles, report.TotalFiles))
	content.WriteString(fmt.Sprintf("- i18n usages found: %d\n", len(report.I18nUsages)))
	content.WriteString(fmt.Sprintf("- Hardcoded strings: %d\n", len(report.HardcodedStrings)))
	content.WriteString(fmt.Sprintf("- Missing translation keys: %d\n", len(report.MissingKeys)))
	
	// Module coverage table
	if len(report.ModuleCoverage) > 0 {
		content.WriteString("\n## Module Coverage\n\n")
		content.WriteString("| Module | i18n Usages |\n")
		content.WriteString("|--------|-------------|\n")
		
		var modules []string
		for name := range report.ModuleCoverage {
			modules = append(modules, name)
		}
		sort.Strings(modules)
		
		for _, name := range modules {
			content.WriteString(fmt.Sprintf("| %s | %d |\n", name, report.ModuleCoverage[name]))
		}
	}
	
	// Missing keys
	if len(report.MissingKeys) > 0 {
		content.WriteString("\n## Missing Translation Keys\n\n")
		uniqueKeys := make(map[string]bool)
		for _, key := range report.MissingKeys {
			if !uniqueKeys[key] {
				content.WriteString(fmt.Sprintf("- `%s`\n", key))
				uniqueKeys[key] = true
			}
		}
	}
	
	// Hardcoded strings
	if len(report.HardcodedStrings) > 0 {
		content.WriteString("\n## Hardcoded Strings (Sample)\n\n")
		content.WriteString("| File | Line | Content |\n")
		content.WriteString("|------|------|----------|\n")
		
		for i, str := range report.HardcodedStrings {
			if i >= 20 { // Limit to first 20
				content.WriteString(fmt.Sprintf("| ... | ... | *and %d more* |\n", len(report.HardcodedStrings)-i))
				break
			}
			escapedContent := strings.ReplaceAll(str.Content, "|", "\\|")
			content.WriteString(fmt.Sprintf("| %s | %d | %s |\n", str.File, str.Line, escapedContent))
		}
	}
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func generateModuleRegistration(outputDir, moduleName string, usages []I18nUsage) error {
	var content strings.Builder
	
	content.WriteString("// Code generated by i18ngen. DO NOT EDIT.\n")
	content.WriteString(fmt.Sprintf("// Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	
	content.WriteString("package i18n\n\n")
	content.WriteString("import (\n")
	content.WriteString("\t\"github.com/divideprojects/Alita_Robot/alita/i18n/catalog\"\n")
	content.WriteString(")\n\n")
	
	moduleTitle := strings.ToUpper(moduleName[:1]) + moduleName[1:]
	content.WriteString(fmt.Sprintf("// Register%sMessages registers all i18n messages for the %s module\n", 
		moduleTitle, moduleName))
	content.WriteString(fmt.Sprintf("func Register%sMessages() {\n", moduleTitle))
	
	// Group by key to avoid duplicates
	keyMap := make(map[string]bool)
	var uniqueKeys []string
	
	for _, usage := range usages {
		if !keyMap[usage.KeyPath] {
			keyMap[usage.KeyPath] = true
			uniqueKeys = append(uniqueKeys, usage.KeyPath)
		}
	}
	
	sort.Strings(uniqueKeys)
	
	for _, keyPath := range uniqueKeys {
		// Generate message registration
		defaultText := generateDefaultText(keyPath)
		content.WriteString(fmt.Sprintf("\tcatalog.MustRegister(\"%s\", \"%s\")\n", 
			keyPath, defaultText))
	}
	
	content.WriteString("}\n\n")
	
	// Generate init function
	content.WriteString("func init() {\n")
	moduleTitleInit := strings.ToUpper(moduleName[:1]) + moduleName[1:]
	content.WriteString(fmt.Sprintf("\tRegister%sMessages()\n", moduleTitleInit))
	content.WriteString("}\n")
	
	filename := filepath.Join(outputDir, moduleName+"_messages.go")
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func generateDefaultText(keyPath string) string {
	// Generate a reasonable default text based on the key path
	parts := strings.Split(keyPath, ".")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Convert snake_case or camelCase to human readable
		result := regexp.MustCompile(`[_-]+`).ReplaceAllString(lastPart, " ")
		result = regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(result, "$1 $2")
		
		// Capitalize first letter
		if len(result) > 0 {
			result = strings.ToUpper(result[:1]) + result[1:]
		}
		
		return result
	}
	return "Message not found"
}