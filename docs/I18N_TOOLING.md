# i18n Tooling for Alita Robot

This document describes the comprehensive i18n (internationalization) tooling available for the Alita Robot project.

## Overview

The project includes three powerful command-line tools to manage translations:

1. **i18ngen** - Code generator for discovering and registering i18n messages
2. **i18nflatten** - YAML flattening tool for converting nested translations to flat structure
3. **i18nlint** - Validation and linting tool for translation consistency

## Quick Start

### Run all i18n tools
```bash
make i18n-check
```

### Individual tools
```bash
make i18n-gen        # Generate code and analyze usage
make i18n-flatten    # Flatten YAML structure
make i18n-lint       # Run validation and linting
```

## Tool Details

### 1. i18ngen - Code Generator

**Purpose**: Scans Go source code to discover i18n usage patterns and generate registration code.

**Features**:
- Scans all Go files in `alita/modules/` directory
- Detects existing `tr.GetString()` calls
- Identifies hardcoded strings that should be internationalized
- Generates message registration code for the catalog system
- Creates detailed analysis reports

**Usage**:
```bash
make i18n-gen
# or
go run cmd/i18ngen/main.go
```

**Output**:
- `generated/i18n/analysis_report.md` - Detailed analysis report
- `generated/i18n/{module}_messages.go` - Registration code for each module
- Console output with colored summary and statistics

**Example Output**:
```
📊 Analysis Report
================

📁 Files processed: 26/26
🔧 i18n usages found: 9
📝 Hardcoded strings: 662
⚠️ Missing translation keys: 0

📈 Module i18n Usage Coverage:
  help: 5 usages
  formatting: 3 usages
  admin: 1 usages
```

### 2. i18nflatten - YAML Flattening Tool

**Purpose**: Converts nested YAML translation structure to flat key structure for better performance and consistency.

**Features**:
- Processes all YAML files in `locales/` directory
- Converts nested keys like `strings.Admin.promote.success` to `admin_promote_success`
- Maintains all translation content while restructuring
- Creates timestamped backups before modification
- Generates transformation reports

**Usage**:
```bash
make i18n-flatten
# or
go run cmd/i18nflatten/main.go
```

**Key Transformations**:
- Removes `strings.` and `main.` prefixes
- Converts dots to underscores
- Normalizes to lowercase
- Handles camelCase to snake_case conversion

**Example Transformations**:
```
strings.Admin.promote.success → admin_promote_success
strings.Antiflood.setflood.success → antiflood_set_flood_success
main.language_name → language_name
```

**Safety Features**:
- Creates backup files with timestamp suffix
- Validates YAML structure before and after transformation
- Provides preview mode (future enhancement)

### 3. i18nlint - Validation and Linting Tool

**Purpose**: Comprehensive validation of i18n usage, translations, and consistency across languages.

**Features**:
- Validates all message keys are used in code
- Finds unused translation keys
- Detects hardcoded strings that should use i18n
- Checks parameter consistency between code and translations
- Generates coverage reports per language
- Provides detailed issue reporting with suggestions

**Usage**:
```bash
make i18n-lint
# or
go run cmd/i18nlint/main.go
```

**Checks Performed**:

1. **Missing Keys**: Keys used in code but not in translation files
2. **Unused Keys**: Keys in translation files but not used in code
3. **Hardcoded Strings**: String literals that should be internationalized
4. **Parameter Consistency**: Mismatched parameters between languages
5. **Naming Conventions**: Key naming best practices
6. **Duplicate Values**: Identical translations that might be copy-paste errors

**Output**:
- `generated/i18n/lint_report.md` - Detailed report with all issues
- `generated/i18n/coverage_report.md` - Language coverage metrics
- `generated/i18n/lint_report.json` - Machine-readable report for CI/CD
- Console output with colored issue breakdown

**Example Output**:
```
📋 Lint Report
=============

📊 Summary:
  Files processed: 26 Go files
  Translation files: 1
  Total translations: 172
  Issues found: 820

🌍 Language Coverage:
  en: 100.0%
  es: 85.2%
  fr: 78.9%
```

## Integration

### CI/CD Pipeline

Add to your CI/CD pipeline:

```yaml
# .github/workflows/i18n-check.yml
name: i18n Validation

on: [push, pull_request]

jobs:
  i18n-check:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Run i18n validation
      run: make i18n-check
    
    - name: Upload reports
      uses: actions/upload-artifact@v4
      with:
        name: i18n-reports
        path: generated/i18n/
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running i18n validation..."
make i18n-lint
if [ $? -ne 0 ]; then
  echo "i18n validation failed. Run 'make i18n-check' to see details."
  exit 1
fi
```

## Best Practices

### Key Naming Conventions

1. Use flat, descriptive keys: `admin_promote_success`
2. Follow module.action.result pattern: `bans_ban_success`
3. Use snake_case for consistency
4. Avoid abbreviations unless common
5. Keep keys under 50 characters when possible

### Translation Management

1. Always use the i18n system for user-facing text
2. Keep English as the reference language
3. Ensure parameter consistency across all languages
4. Use meaningful parameter names: `{username}` not `{param1}`
5. Test translations in context

### Development Workflow

1. Write code with hardcoded English strings initially
2. Run `make i18n-gen` to identify strings needing i18n
3. Move strings to translation files
4. Update code to use `tr.GetString()`
5. Run `make i18n-lint` to validate
6. Add translations for other languages
7. Run full check with `make i18n-check`

## Configuration

### Environment Variables

- `I18N_DEBUG=true` - Enable debug output
- `I18N_NO_COLOR=true` - Disable colored output
- `I18N_OUTPUT_DIR=path` - Custom output directory

### Makefile Targets

- `make i18n-gen` - Run code generator
- `make i18n-flatten` - Flatten YAML files (⚠️ modifies files)
- `make i18n-lint` - Run validation
- `make i18n-check` - Run all i18n tools

## Troubleshooting

### Common Issues

**Issue**: "Failed to parse YAML"
- **Solution**: Check YAML syntax with `yamllint locales/`

**Issue**: "Missing translation keys"
- **Solution**: Add missing keys to translation files or update code

**Issue**: "Parameter mismatch"
- **Solution**: Ensure all languages use the same parameter names

**Issue**: "Too many hardcoded strings"
- **Solution**: Gradually migrate strings to i18n system

### Debugging

Enable debug mode:
```bash
I18N_DEBUG=true make i18n-lint
```

Check individual files:
```bash
go run cmd/i18nlint/main.go --file=alita/modules/bans.go
```

## Advanced Features

### Custom Validation Rules

The linter supports custom validation rules through configuration:

```yaml
# .i18nlint.yml
rules:
  max_key_length: 50
  require_descriptions: true
  allowed_prefixes: ["admin", "user", "error"]
  parameter_style: "snake_case"
```

### Batch Operations

Process specific modules only:
```bash
go run cmd/i18ngen/main.go --modules=bans,admin,help
```

Generate reports in different formats:
```bash
go run cmd/i18nlint/main.go --format=json,markdown,junit
```

### Integration with Translation Services

The tools can be extended to work with translation services:

```bash
# Export for translation service
go run cmd/i18ngen/main.go --export-keys > keys_for_translation.json

# Import translations
go run cmd/i18nlint/main.go --import-translations translations.csv
```

## Migration Guide

### From Nested to Flat YAML

1. **Backup**: Automatic backups are created
2. **Preview**: Review transformation examples
3. **Flatten**: Run `make i18n-flatten`
4. **Update Code**: Update Go code to use new flat keys
5. **Validate**: Run `make i18n-lint` to check for issues
6. **Test**: Test the application thoroughly

### Updating Existing Code

Before:
```go
text, _ := tr.GetString("strings.Admin.promote.success")
```

After:
```go
text, _ := tr.GetString("admin_promote_success")
```

## Performance Considerations

- Flat keys improve lookup performance
- Reduced memory usage with flattened structure
- Faster startup time with optimized key registration
- Better caching efficiency with consistent key patterns

## Contributing

To contribute improvements to the i18n tooling:

1. Fork the repository
2. Create feature branch: `git checkout -b feature/i18n-enhancement`
3. Add tests for new functionality
4. Run the full test suite: `make test`
5. Submit pull request with detailed description

## Support

For questions or issues with the i18n tooling:

- Create an issue in the GitHub repository
- Join the project's Discord server
- Check existing documentation in `/docs`

---

*This tooling is designed to make internationalization easy and maintainable for the Alita Robot project. Regular use of these tools ensures high-quality, consistent translations across all supported languages.*