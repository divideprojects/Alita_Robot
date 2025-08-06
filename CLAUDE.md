# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Alita Robot is a modern Telegram group management bot built with Go and the gotgbot library. It provides comprehensive group administration features including user management, filters, greetings, anti-spam, and multi-language support.

## Development Commands

```bash
make run          # Run the bot locally with current code
make build        # Build release artifacts using goreleaser (creates binaries for multiple platforms)
make lint         # Run golangci-lint for code quality checks (requires golangci-lint installed)
make tidy         # Clean up and download go.mod dependencies
make vendor       # Vendor all dependencies locally
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
```

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

## Deployment Modes

The bot supports two deployment modes:

### Polling Mode (Default)
- Uses long polling to receive updates from Telegram
- No external network configuration required
- Suitable for development and simple deployments
- Set `USE_WEBHOOKS=false` or leave unset

### Webhook Mode
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

## Migration from MongoDB

If migrating from MongoDB:
1. Build migration tool: `cd cmd/migrate && go build`
2. Configure MongoDB and PostgreSQL connections in `.env`
3. Run migration: `./migrate`
4. Verify data integrity post-migration

## Important Patterns

1. **Error Handling**: Use the centralized error_handling package
2. **Permissions**: Check user permissions via chat_status utilities
3. **Caching**: Use the cache package for frequently accessed data
4. **Database**: Follow repository pattern for database operations
5. **Commands**: Use decorators for common command middleware

## Update Instructions

When discovering new architectural patterns, modules, or significant changes not documented here:
1. Update this CLAUDE.md file to reflect the changes
2. Include any new commands, environment variables, or architectural decisions
3. Document any new dependencies or external service requirements
4. Keep the architecture overview current with actual code structure