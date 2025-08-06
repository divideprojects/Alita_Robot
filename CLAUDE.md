# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Alita Robot is a modern Telegram group management bot built with Go and the gotgbot library. It provides comprehensive group administration features including user management, filters, greetings, anti-spam, and multi-language support.

## Development Commands

### Basic Commands
```bash
make run          # Run the bot locally with current code
make build        # Build release artifacts using goreleaser (creates binaries for multiple platforms)
make lint         # Run golangci-lint for code quality checks (requires golangci-lint installed)
make tidy         # Clean up and download go.mod dependencies
make vendor       # Vendor all dependencies locally
```

### PostgreSQL Migration Commands
```bash
make psql-prepare  # Prepare migrations from Supabase migration files
make psql-migrate  # Apply all pending PostgreSQL migrations
make psql-status   # Check current migration status
make psql-rollback # Show rollback information (does not execute)
make psql-reset    # Reset database - DANGEROUS: drops and recreates all tables
```

## Architecture Overview

### Core Structure (`alita/`)
- **config/** - Configuration management, reading environment variables
- **db/** - Database layer
  - PostgreSQL with GORM ORM
  - Repository pattern with interfaces in `repositories/interfaces/`
  - Implementations in `repositories/implementations/`
  - Models for each entity (users, chats, filters, etc.)
  - Individual DB files for each module (admin_db.go, filters_db.go, etc.)
- **modules/** - Bot command handlers
  - Each file handles specific functionality (admin.go, filters.go, greetings.go, etc.)
  - Commands are registered via dispatcher in main.go
- **utils/** - Utility packages
  - **cache/** - Redis + Ristretto dual-layer caching system
  - **chat_status/** - User permission checking
  - **decorators/** - Command decorators for handler middleware
  - **error_handling/** - Centralized error handling
  - **extraction/** - Message parsing and entity extraction
  - **string_handling/** - Text manipulation utilities
  - **webhook/** - Webhook server implementation with security validation
- **i18n/** - Internationalization with YAML locale files

### Supporting Components
- **cmd/migrate/** - MongoDB to PostgreSQL migration tool with batch processing
- **locales/** - Language files in YAML format (currently English is primary)
- **supabase/migrations/** - PostgreSQL schema migrations

## Database Schema

The bot uses PostgreSQL with the following key tables:
- users, chats - Core entities
- admin_settings, locks, pins - Permission management
- filters, notes - Content management
- greetings - Welcome/goodbye messages
- warns_settings, warns_users - Warning system
- antiflood_settings - Spam protection
- blacklists - Word filtering
- And more for various features

## Environment Configuration

Required environment variables (see sample.env):
```
# Core Configuration
BOT_TOKEN          # Telegram bot token from @BotFather
DATABASE_URL       # PostgreSQL connection string
REDIS_ADDRESS      # Redis server address
REDIS_PASSWORD     # Redis password (if required)
MESSAGE_DUMP       # Chat ID for bot logs (must start with -100)
OWNER_ID           # Your Telegram user ID
ENABLED_LOCALES    # Comma-separated locale codes (default: en)

# Webhook Configuration (optional - for webhook mode)
USE_WEBHOOKS       # Set to 'true' to enable webhook mode (default: false/polling)
WEBHOOK_DOMAIN     # Your webhook domain (e.g., https://your-bot.example.com)
WEBHOOK_SECRET     # Random secret string for webhook security
WEBHOOK_PORT       # Port for webhook server (default: 8080)
CLOUDFLARE_TUNNEL_TOKEN # Cloudflare tunnel token (if using cloudflared)

# Performance Tuning (optional)
MAX_DB_POOL_SIZE   # Maximum database connection pool size (default: calculated)
MAX_IDLE_CONNS     # Maximum idle database connections (default: calculated)
CONN_MAX_LIFETIME  # Maximum connection lifetime in seconds (default: 3600)
CONN_MAX_IDLE_TIME # Maximum connection idle time in seconds (default: 1800)
DB_BATCH_SIZE      # Batch size for bulk operations (default: 100)
CACHE_TTL          # Cache time-to-live in seconds (default: 300)
CACHE_SIZE         # In-memory cache size in MB (default: 100)
RISTRETTO_CACHE_SIZE # Ristretto cache size (default: 1<<30)
RISTRETTO_NUM_COUNTERS # Ristretto counter estimate (default: 1e7)
WORKER_POOL_SIZE   # Worker pool size for concurrent operations (default: 10)
BULK_UPDATE_WORKERS # Workers for bulk updates (default: 5)
QUERY_TIMEOUT      # Database query timeout in seconds (default: 30)
ENABLE_STATS       # Enable background statistics collection (default: false)
STATS_INTERVAL     # Statistics collection interval in seconds (default: 60)
```

**Note:** The app.json file contains outdated MongoDB references (DB_URI, DB_NAME) that should be ignored. Use sample.env as the authoritative source for environment variables.

## Key Technical Details

1. **Database**: PostgreSQL with GORM ORM, connection pooling, and batch operations
2. **Caching**: Two-layer cache with Ristretto (L1) and Redis (L2)
3. **Concurrency**: Dispatcher limited to 100 max routines to prevent goroutine explosion
4. **Monitoring**: Built-in resource monitor tracks memory and goroutine usage
5. **Localization**: Multi-language support via i18n package and YAML locale files
6. **Message Types**: Supports text, sticker, document, photo, audio, voice, video, video note
7. **Deployment Modes**: Supports both polling and webhook modes with automatic mode detection
8. **Webhook Security**: Built-in secret token validation and graceful shutdown handling

## Module Development

When adding new features:
1. Create database models in `alita/db/` if needed
2. Add database operations in a new `*_db.go` file
3. Implement command handlers in `alita/modules/`
4. Register commands in the module's init function
5. Add translations to locale files if user-facing

## Testing & Quality

- Use `make lint` to check code quality before commits
- The project uses golangci-lint for comprehensive linting
- No dedicated test files currently exist in the project

## CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment:

### Continuous Integration (.github/workflows/ci.yml)
- Triggered on all pushes and pull requests
- Runs GoReleaser in snapshot mode to verify builds
- Tests compilation across multiple platforms and architectures

### Release Pipeline (.github/workflows/release.yml)
- Triggered on version tags (e.g., v1.0.0)
- Builds multi-architecture binaries for Darwin, Linux, and Windows
- Creates and publishes Docker images to GitHub Container Registry (ghcr.io/divideprojects/alita_robot)
- Signs releases with GPG for security verification
- Generates release notes and checksums automatically

## Build and Release Process

The project uses GoReleaser for building and releasing:

### GoReleaser Configuration (.goreleaser.yaml)
- **Architectures:** amd64, arm64 for all platforms
- **Platforms:** Darwin (macOS), Linux, Windows
- **Docker:** Multi-platform images (linux/amd64, linux/arm64)
- **Features:**
  - Automatic changelog generation
  - GPG signing for security
  - Archive creation with checksums
  - Docker image publishing to GHCR
  - Binary stripping for size optimization

### Creating a Release
1. Tag the commit: `git tag -a v1.0.0 -m "Release version 1.0.0"`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions will automatically build and publish the release

## Deployment Options

### Deployment Modes

The bot supports two deployment modes:

#### Polling Mode (Default)
- Uses long polling to receive updates from Telegram
- No external network configuration required
- Suitable for development and simple deployments
- Set `USE_WEBHOOKS=false` or leave unset

#### Webhook Mode
- Telegram sends updates via HTTP POST to your server
- Requires public HTTPS endpoint
- Better for production deployments with high traffic
- Set `USE_WEBHOOKS=true` and configure webhook variables

**Webhook Setup:**
1. Set up Cloudflare tunnel: `cloudflared tunnel create alita-bot`
2. Configure tunnel to point to your webhook port
3. Set environment variables:
   ```bash
   USE_WEBHOOKS=true
   WEBHOOK_DOMAIN=https://your-tunnel-domain.trycloudflare.com
   WEBHOOK_SECRET=your-random-secret-string
   WEBHOOK_PORT=8080
   CLOUDFLARE_TUNNEL_TOKEN=your-tunnel-token
   ```
4. Uncomment cloudflared service in docker-compose.yml
5. Start the bot - it will automatically configure the webhook

**Security Features:**
- Webhook secret token validation
- Request method validation
- Graceful shutdown with webhook cleanup
- Health check endpoint at `/health`

### Docker Deployment

The project provides multiple Docker configurations:

#### Production (docker-compose.yml)
- Full stack with PostgreSQL, Redis, and optional Cloudflare tunnel
- Resource limits configured for all services
- Health checks for database and cache
- Persistent volumes for data

#### Local Development (local.docker-compose.yml)
- Includes MongoDB for legacy compatibility
- Simplified configuration for local testing

#### Debug Mode (debug.docker-compose.yml)
- Enhanced logging and debugging capabilities
- Useful for troubleshooting production issues

#### Dockerfiles Available
- **Dockerfile.alpine**: Production Alpine Linux image (smallest size)
- **Dockerfile.alpine.debug**: Alpine with debugging tools
- **Dockerfile.goreleaser**: Used by GoReleaser for automated builds

## Database Migration Systems

### MongoDB to PostgreSQL Migration

For migrating existing MongoDB data:
1. Build migration tool: `cd cmd/migrate && go build`
2. Configure connections in `.env`:
   ```bash
   # MongoDB source
   DB_URI=mongodb://localhost:27017
   DB_NAME=your_mongo_db
   
   # PostgreSQL target
   DATABASE_URL=postgresql://user:pass@localhost/alita
   ```
3. Run migration with options:
   ```bash
   ./migrate                    # Standard migration
   ./migrate --dry-run          # Preview without changes
   ./migrate --batch-size=500   # Custom batch size
   ./migrate --verbose          # Detailed logging
   ```
4. Verify data integrity post-migration

**Migration Features:**
- Batch processing for large datasets
- Dry-run mode for testing
- Progress tracking and error reporting
- Automatic data type conversion
- Detailed documentation in docs/MIGRATION_MONGO_TO_PSQL.md

### Native PostgreSQL Migrations

For fresh PostgreSQL deployments or schema updates:
1. Place migration files in `supabase/migrations/`
2. Run `make psql-prepare` to prepare migrations
3. Run `make psql-migrate` to apply
4. Check status with `make psql-status`

**Migration System Features:**
- Automatic Supabase dependency cleaning
- Version tracking and rollback support
- Schema validation
- Migration history in database

## Important Patterns

1. **Error Handling**: Use the centralized error_handling package
2. **Permissions**: Check user permissions via chat_status utilities
3. **Caching**: Use the cache package for frequently accessed data
4. **Database**: Follow repository pattern for database operations
5. **Commands**: Use decorators for common command middleware

## Supabase Integration

The project includes Supabase configuration files, though the bot now uses direct PostgreSQL connections:
- **supabase/migrations/**: Contains database schema migrations
- **supabase/config.toml**: Supabase project configuration
- Migration files are used by the native PostgreSQL migration system

## Update Instructions

When discovering new architectural patterns, modules, or significant changes not documented here:
1. Update this CLAUDE.md file to reflect the changes
2. Include any new commands, environment variables, or architectural decisions
3. Document any new dependencies or external service requirements
4. Keep the architecture overview current with actual code structure
5. Update sample.env when adding new configuration options
6. Ensure CI/CD workflows reflect any build process changes