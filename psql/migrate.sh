#!/bin/bash

# PostgreSQL Migration Script for Alita Robot
# This script applies cleaned migrations to any PostgreSQL database

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Load .env file if it exists
if [[ -f "$SCRIPT_DIR/.env" ]]; then
    echo -e "${BLUE}Loading configuration from .env file...${NC}"
    set -a  # Export all variables
    source "$SCRIPT_DIR/.env"
    set +a  # Stop exporting
elif [[ -f "$SCRIPT_DIR/sample.env" ]] && [[ ! -f "$SCRIPT_DIR/.env" ]]; then
    echo -e "${YELLOW}No .env file found. Copy sample.env to .env and update with your credentials:${NC}"
    echo -e "${YELLOW}  cp $SCRIPT_DIR/sample.env $SCRIPT_DIR/.env${NC}"
    echo
fi

# Database configuration from environment variables
DB_HOST="${PSQL_DB_HOST}"
DB_PORT="${PSQL_DB_PORT:-5432}"
DB_NAME="${PSQL_DB_NAME}"
DB_USER="${PSQL_DB_USER}"
DB_PASSWORD="${PSQL_DB_PASSWORD}"
DB_SSLMODE="${PSQL_DB_SSLMODE:-require}"

# Migration directory
MIGRATIONS_DIR="$SCRIPT_DIR/migrations"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Function to execute SQL
execute_sql() {
    PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        -v ON_ERROR_STOP=1 \
        "$@" 2>&1
}

# Function to check if a migration has been applied
is_migration_applied() {
    local version=$1
    local result=$(execute_sql -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '${version}';" 2>/dev/null | tr -d ' ')
    [[ "$result" == "1" ]]
}

# Function to apply a migration
apply_migration() {
    local migration_file=$1
    local version=$(basename "$migration_file")
    
    print_color "$BLUE" "  → Applying ${version}..."
    
    # Apply the migration
    if execute_sql -f "$migration_file"; then
        # Record successful migration
        execute_sql -c "INSERT INTO schema_migrations (version) VALUES ('${version}');"
        print_color "$GREEN" "    ✓ Applied successfully"
        return 0
    else
        print_color "$RED" "    ✗ Failed to apply migration"
        return 1
    fi
}

# Main execution
main() {
    print_color "$BLUE" "=========================================="
    print_color "$BLUE" "PostgreSQL Migration Tool for Alita Robot"
    print_color "$BLUE" "=========================================="
    echo
    
    # Check required environment variables
    if [[ -z "$DB_HOST" || -z "$DB_NAME" || -z "$DB_USER" ]]; then
        print_color "$RED" "Error: Required environment variables not set"
        print_color "$YELLOW" "Please configure your database connection:"
        echo
        echo "  Option 1: Create a .env file"
        echo "    cp $SCRIPT_DIR/sample.env $SCRIPT_DIR/.env"
        echo "    # Edit .env with your database credentials"
        echo
        echo "  Option 2: Set environment variables"
        echo "    export PSQL_DB_HOST=your-host"
        echo "    export PSQL_DB_NAME=your-database"
        echo "    export PSQL_DB_USER=your-username"
        echo "    export PSQL_DB_PASSWORD=your-password"
        echo "    export PSQL_DB_PORT=5432         # optional, default: 5432"
        echo "    export PSQL_DB_SSLMODE=require   # optional, default: require"
        exit 1
    fi
    
    # Check if migrations directory exists
    if [[ ! -d "$MIGRATIONS_DIR" ]]; then
        print_color "$RED" "Error: Migrations directory not found: $MIGRATIONS_DIR"
        print_color "$YELLOW" "Run 'make psql-prepare' first to prepare migrations"
        exit 1
    fi
    
    # Test database connection
    print_color "$BLUE" "Testing database connection..."
    if ! execute_sql -c "SELECT 1;" > /dev/null; then
        print_color "$RED" "Error: Cannot connect to database"
        print_color "$YELLOW" "Connection details:"
        echo "  Host: $DB_HOST:$DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  User: $DB_USER"
        exit 1
    fi
    print_color "$GREEN" "✓ Connected to database"
    echo
    
    # Create migrations tracking table if it doesn't exist
    print_color "$BLUE" "Ensuring migrations table exists..."
    execute_sql <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
EOF
    print_color "$GREEN" "✓ Migrations table ready"
    echo
    
    # Get list of migration files
    print_color "$BLUE" "Scanning for migrations..."
    migration_files=($(ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort))
    
    if [[ ${#migration_files[@]} -eq 0 ]]; then
        print_color "$YELLOW" "No migration files found in $MIGRATIONS_DIR"
        exit 0
    fi
    
    print_color "$GREEN" "Found ${#migration_files[@]} migration files"
    echo
    
    # Apply migrations
    print_color "$BLUE" "Applying migrations..."
    applied_count=0
    skipped_count=0
    failed_count=0
    
    for migration_file in "${migration_files[@]}"; do
        version=$(basename "$migration_file")
        
        # Check if already applied
        if is_migration_applied "$version"; then
            print_color "$YELLOW" "  ○ Skipping ${version} (already applied)"
            ((skipped_count++))
        else
            if apply_migration "$migration_file"; then
                ((applied_count++))
            else
                ((failed_count++))
                print_color "$RED" "Migration failed. Stopping execution."
                exit 1
            fi
        fi
    done
    
    echo
    print_color "$BLUE" "=========================================="
    print_color "$GREEN" "Migration Summary:"
    echo "  • Applied: $applied_count"
    echo "  • Skipped: $skipped_count"
    echo "  • Failed: $failed_count"
    print_color "$BLUE" "=========================================="
    
    # Show current migration status
    echo
    print_color "$BLUE" "Current migration status:"
    execute_sql -c "SELECT version, executed_at FROM schema_migrations ORDER BY executed_at DESC LIMIT 5;"
    
    # Show table count
    echo
    table_count=$(execute_sql -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" | tr -d ' ')
    print_color "$GREEN" "✓ Database has $table_count tables"
    
    exit 0
}

# Run main function
main "$@"