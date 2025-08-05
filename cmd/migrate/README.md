# MongoDB to PostgreSQL Migration Tool

This tool migrates data from MongoDB to PostgreSQL for the Alita Robot project.

## Features

- Batch processing for efficient migration of large datasets
- Progress tracking and statistics
- Dry-run mode for testing without writing data
- Upsert support to handle existing data
- Comprehensive error handling and logging
- Support for all Alita Robot collections

## Prerequisites

1. MongoDB instance with existing data
2. PostgreSQL database with schema already created (run the migration SQL first)
3. Go 1.21 or higher

## Installation

```bash
cd cmd/migrate
go build -o migrate
```

## Configuration

Create a `.env` file based on `.env.example`:

```env
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=alita

# PostgreSQL Configuration
DATABASE_URL=postgres://user:password@localhost:5432/alita_db?sslmode=disable
```

## Usage

### Basic Migration

```bash
./migrate
```

### With Command-Line Flags

```bash
./migrate \
  -mongo-uri="mongodb://localhost:27017" \
  -mongo-db="alita" \
  -postgres-dsn="postgres://user:pass@localhost/alita_db" \
  -batch-size=500 \
  -verbose
```

### Dry Run (Test Mode)

```bash
./migrate -dry-run
```

This will simulate the migration without writing any data to PostgreSQL.

## Command-Line Options

- `-mongo-uri`: MongoDB connection URI
- `-mongo-db`: MongoDB database name (default: "alita")
- `-postgres-dsn`: PostgreSQL connection DSN
- `-batch-size`: Number of records to process in each batch (default: 1000)
- `-dry-run`: Perform a dry run without writing to PostgreSQL
- `-verbose`: Enable verbose logging

## Migration Process

The tool migrates the following collections:

1. **users** - User information and preferences
2. **chats** - Chat groups and their metadata
3. **admin** - Admin settings per chat
4. **notes_settings** - Note configuration per chat
5. **notes** - Saved notes and messages
6. **filters** - Keyword filters and auto-responses
7. **greetings** - Welcome and goodbye messages
8. **locks** - Permission locks and restrictions
9. **pins** - Pin settings per chat
10. **rules** - Chat rules
11. **warns_settings** - Warning system configuration
12. **warns_users** - User warnings
13. **antiflood_settings** - Anti-flood configuration
14. **blacklists** - Blacklisted words
15. **channels** - Linked channels
16. **connection** - User-chat connections
17. **connection_settings** - Connection configuration
18. **disable** - Disabled commands per chat
19. **report_user_settings** - User report settings
20. **report_chat_settings** - Chat report settings

## Data Transformations

The migration handles several data transformations:

- **Nested documents** are flattened (e.g., greetings.welcome_settings)
- **Arrays** are converted to JSONB (e.g., warns, chat users) or expanded to individual rows (e.g., blacklist triggers)
- **MongoDB Long types** are converted to PostgreSQL bigint
- **Missing fields** are handled with appropriate defaults
- **Permissions/Restrictions** in locks are expanded to individual rows
- **Blacklist triggers** are expanded from a single array to individual word entries

## Special Considerations

1. **Chat Users**: The `chats.users` array is migrated to both a JSONB column and a separate `chat_users` junction table
2. **Locks**: Permissions and restrictions are expanded from nested objects to individual lock_type rows
3. **Disable**: Commands array is expanded to individual command rows
4. **Blacklists**: The `triggers` array is expanded so each blacklisted word becomes a separate row in PostgreSQL

## Error Handling

- Failed records are logged but don't stop the migration
- Each collection is migrated independently
- Statistics show success/failure counts
- Detailed error messages are provided for debugging

## Post-Migration

After migration:

1. Verify row counts match between MongoDB and PostgreSQL
2. Test the application with the migrated data
3. Update PostgreSQL sequences if needed
4. Consider creating additional indexes for performance

## Troubleshooting

### Connection Issues

- Ensure MongoDB is accessible and running
- Verify PostgreSQL credentials and database exists
- Check network connectivity between the tool and databases

### Data Issues

- Run with `-verbose` flag for detailed logging
- Use `-dry-run` to test without writing data
- Check the error logs for specific record failures

### Performance

- Adjust `-batch-size` based on your system resources
- Smaller batches use less memory but take longer
- Larger batches are faster but require more memory

## Development

To modify the migration:

1. Update models in `models.go` for data structures
2. Modify migration functions in `migrations.go` for transformation logic
3. Adjust batch processing in `main.go` for performance tuning