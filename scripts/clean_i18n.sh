#!/bin/bash

# Alita Robot i18n Cleanup Script
# Removes generated files and backups created by i18n tools

set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧹 Alita Robot i18n Cleanup Script${NC}"
echo -e "${GREEN}===================================${NC}"

# Function to confirm action
confirm() {
    local prompt="$1"
    local response
    
    while true; do
        echo -e "${YELLOW}$prompt (y/n): ${NC}"
        read -r response
        case $response in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

# Clean generated i18n files
if [ -d "generated/i18n" ]; then
    if confirm "🗂️  Remove generated i18n reports and code?"; then
        rm -rf generated/i18n
        echo -e "${GREEN}✅ Removed generated/i18n directory${NC}"
    fi
fi

# Clean YAML backup files
backup_files=$(find locales/ -name "*.backup.*" 2>/dev/null || true)
if [ -n "$backup_files" ]; then
    echo -e "${YELLOW}📁 Found YAML backup files:${NC}"
    echo "$backup_files"
    
    if confirm "🗂️  Remove YAML backup files?"; then
        find locales/ -name "*.backup.*" -delete
        echo -e "${GREEN}✅ Removed YAML backup files${NC}"
    fi
fi

# Clean build artifacts from i18n tools
if [ -f "main" ]; then
    if confirm "🔧 Remove build artifacts (./main)?"; then
        rm -f main
        echo -e "${GREEN}✅ Removed build artifacts${NC}"
    fi
fi

# Clean Go build cache for i18n tools
if confirm "🔄 Clean Go build cache for i18n tools?"; then
    go clean -cache cmd/i18ngen/...
    go clean -cache cmd/i18nflatten/...
    go clean -cache cmd/i18nlint/...
    echo -e "${GREEN}✅ Cleaned Go build cache${NC}"
fi

# Optional: Reset to original YAML structure
if [ -d "locales" ]; then
    recent_backups=$(find locales/ -name "*.backup.*" -mtime -1 2>/dev/null || true)
    if [ -n "$recent_backups" ]; then
        echo -e "${YELLOW}⚠️  Found recent backups (less than 24h old)${NC}"
        if confirm "🔄 Restore from most recent backups?"; then
            for backup in $recent_backups; do
                original=${backup%.backup.*}
                if [ -f "$backup" ]; then
                    cp "$backup" "$original"
                    echo -e "${GREEN}✅ Restored $original from backup${NC}"
                fi
            done
        fi
    fi
fi

echo -e "${GREEN}🎉 Cleanup complete!${NC}"

# Show remaining files
echo -e "${YELLOW}📋 Current i18n-related files:${NC}"
echo "Generated files:"
ls -la generated/ 2>/dev/null || echo "  (none)"
echo
echo "Locale files:"
ls -la locales/ 2>/dev/null || echo "  (none)"
echo
echo "Backup files:"
find . -name "*.backup.*" 2>/dev/null || echo "  (none)"

echo
echo -e "${GREEN}💡 To run i18n tools again:${NC}"
echo "  make i18n-gen     - Generate code and analyze usage"
echo "  make i18n-flatten - Flatten YAML structure" 
echo "  make i18n-lint    - Run validation and linting"
echo "  make i18n-check   - Run all i18n tools"