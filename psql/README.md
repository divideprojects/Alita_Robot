# PostgreSQL Migration System

This directory contains cleaned PostgreSQL migrations adapted from Supabase for use with any standard PostgreSQL database.

## Overview

The migration system allows you to deploy the Alita Robot database schema to any PostgreSQL instance without Supabase-specific dependencies. All Supabase-specific elements (roles, extensions with schema) have been automatically cleaned during the preparation process.

## Quick Start

### 1. Prepare Migrations

First, generate cleaned migrations from the Supabase files:

```bash
make psql-prepare
```

This will:
- Create the `psql/migrations/` directory
- Clean all SQL files (remove Supabase-specific GRANT statements and fix extensions)
- Prepare the migration script

### 2. Configure Database Connection

#### Option A: Using .env file (Recommended)

```bash
# Copy the sample environment file
cp psql/sample.env psql/.env

# Edit with your database credentials
nano psql/.env
```

#### Option B: Using environment variables

```bash
export PSQL_DB_HOST="your-postgres-host.com"
export PSQL_DB_PORT="5432"              # Optional, defaults to 5432
export PSQL_DB_NAME="alita_robot"
export PSQL_DB_USER="your_username"
export PSQL_DB_PASSWORD="your_password"
export PSQL_DB_SSLMODE="require"        # Optional, defaults to require
```

### 3. Apply Migrations

Run the migrations:

```bash
make psql-migrate
```

Or run directly:

```bash
cd psql && ./migrate.sh
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make psql-prepare` | Clean and prepare migrations from Supabase files |
| `make psql-migrate` | Apply migrations to the configured database |
| `make psql-status` | Check migration status and show applied migrations |
| `make psql-rollback` | Show information about the last applied migration |
| `make psql-reset` | **DANGER**: Drop all tables and reset the database |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PSQL_DB_HOST` | Yes | - | PostgreSQL server hostname or IP |
| `PSQL_DB_PORT` | No | 5432 | PostgreSQL server port |
| `PSQL_DB_NAME` | Yes | - | Database name |
| `PSQL_DB_USER` | Yes | - | Database username |
| `PSQL_DB_PASSWORD` | Yes | - | Database password |
| `PSQL_DB_SSLMODE` | No | require | SSL mode (disable/require/verify-ca/verify-full) |

## Migration Files

The cleaned migrations are stored in `psql/migrations/` and include:

1. **Initial Migration** - Creates all base tables, sequences, and indexes
2. **Foreign Keys** - Adds referential integrity constraints
3. **Index Optimizations** - Performance-tuned indexes based on production usage
4. **Extensions** - PostgreSQL extensions (without Supabase-specific schemas)

## What Gets Cleaned

The preparation process automatically removes:

- All `GRANT` statements for Supabase-specific roles (anon, authenticated, service_role)
- Schema specifications from extension creation (`with schema "extensions"`)
- Converts `create extension` to `CREATE EXTENSION IF NOT EXISTS`

## Migration Tracking

The system uses a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Each migration is applied only once and recorded in this table.

## Using with Docker

You can use these migrations with a Docker PostgreSQL container:

```bash
# Start PostgreSQL
docker run -d \
  --name alita-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=alita_robot \
  -p 5432:5432 \
  postgres:15

# Configure connection
export PSQL_DB_HOST="localhost"
export PSQL_DB_NAME="alita_robot"
export PSQL_DB_USER="postgres"
export PSQL_DB_PASSWORD="password"

# Apply migrations
make psql-migrate
```

## Using with Cloud Providers

### AWS RDS

```bash
export PSQL_DB_HOST="mydb.abc123.us-east-1.rds.amazonaws.com"
export PSQL_DB_PORT="5432"
export PSQL_DB_NAME="alita_robot"
export PSQL_DB_USER="postgres"
export PSQL_DB_PASSWORD="your-password"
export PSQL_DB_SSLMODE="require"

make psql-migrate
```

### Google Cloud SQL

```bash
export PSQL_DB_HOST="34.XXX.XXX.XXX"  # Or use Cloud SQL Proxy
export PSQL_DB_NAME="alita_robot"
export PSQL_DB_USER="postgres"
export PSQL_DB_PASSWORD="your-password"

make psql-migrate
```

### Azure Database for PostgreSQL

```bash
export PSQL_DB_HOST="myserver.postgres.database.azure.com"
export PSQL_DB_NAME="alita_robot"
export PSQL_DB_USER="username@myserver"
export PSQL_DB_PASSWORD="your-password"
export PSQL_DB_SSLMODE="require"

make psql-migrate
```

## Manual Migration

If you prefer to apply migrations manually:

```bash
# Apply all migrations in order
for file in psql/migrations/*.sql; do
  psql -h host -U user -d database -f "$file"
done
```

Or concatenate all migrations:

```bash
cat psql/migrations/*.sql | psql -h host -U user -d database
```

## Troubleshooting

### Connection Issues

If you can't connect to the database:

1. Check network connectivity: `ping your-host`
2. Verify PostgreSQL is running: `telnet your-host 5432`
3. Check credentials: `psql -h host -U user -d database -c "SELECT 1;"`
4. Verify SSL settings if required by your provider

### Migration Failures

If a migration fails:

1. Check the error message for specific issues
2. Verify the database user has sufficient permissions:
   ```sql
   GRANT CREATE ON DATABASE alita_robot TO your_user;
   GRANT ALL PRIVILEGES ON SCHEMA public TO your_user;
   ```
3. Check for existing tables that might conflict
4. Review the specific migration file that failed

### Permission Issues

Ensure your database user has the following permissions:

- `CREATE` - To create tables and indexes
- `ALTER` - To modify tables
- `DROP` - For reset functionality
- `INSERT`, `UPDATE`, `DELETE`, `SELECT` - For data operations

### Extension Issues

If extension creation fails:

1. Ensure your PostgreSQL user has superuser privileges or:
2. Have the extensions pre-installed by your database administrator
3. Required extensions:
   - `hypopg` (optional, for index analysis)
   - `index_advisor` (optional, for index recommendations)

## Directory Structure

```
psql/
├── migrations/       # Cleaned SQL migration files
├── migrate.sh       # Migration application script
├── sample.env       # Sample environment configuration
├── .env            # Your database configuration (gitignored)
├── .gitignore      # Git ignore rules
├── rollback/        # Rollback scripts (future feature)
└── README.md        # This file
```

## Notes

- Migrations are applied in alphabetical order (sorted by timestamp in filename)
- Each migration runs in a transaction (when possible)
- The system is idempotent - safe to run multiple times
- Failed migrations stop the process to prevent data corruption

## Support

For issues specific to the PostgreSQL migration system, check:

1. The main project README
2. PostgreSQL logs for detailed error messages
3. The `schema_migrations` table for applied migrations

## License

These migrations are part of the Alita Robot project and follow the same license terms.