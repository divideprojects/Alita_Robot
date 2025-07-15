# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
- `make run` - Run the bot locally using Go
- `make build` - Build production binaries using GoReleaser
- `make tidy` - Clean and tidy Go modules
- `make vendor` - Vendor dependencies

### Testing
- `go test ./...` - Run all tests
- `go test ./alita/modules/...` - Run module tests
- `go test -bench=.` - Run benchmarks
- `go test -v ./alita/modules/antiflood_test.go` - Run specific test files

### Docker Development
- `docker-compose up` - Run bot with MongoDB and Redis (production mode)
- `docker-compose -f local.docker-compose.yml up` - Local development setup
- `docker-compose -f debug.docker-compose.yml up` - Debug mode with detailed logging

## Architecture Overview

Alita is a Telegram bot written in Go using the gotgbot library. The architecture follows a modular design:

### Core Components

**main.go** - Entry point that:
- Loads configuration from environment variables
- Initializes i18n locales from embedded files
- Sets up bot polling and dispatcher
- Loads all modules and starts the CAPTCHA scheduler

**alita/main.go** - Core bot initialization:
- Performs initial checks and cache setup
- Manages module loading with dependency order
- Implements resource monitoring for memory and goroutines
- Handles graceful startup and shutdown

**alita/config/** - Configuration management:
- Environment variable parsing with defaults
- Structured logging configuration (JSON format)
- Database connection settings (MongoDB + Redis)
- Bot token and API server configuration

**alita/db/** - Database layer:
- MongoDB collections for all bot data
- Optimized indexes for performance
- Retry logic and slow query monitoring
- Connection pooling and timeout handling

### Module System

The bot uses a modular architecture where each feature is a separate module:

- **Admin module** - Admin commands and permissions
- **Antiflood module** - Rate limiting and flood protection
- **Captcha module** - User verification system
- **Filters module** - Custom message filters
- **Notes module** - Saved notes and responses
- **Warnings module** - Warning system
- **Greetings module** - Welcome/goodbye messages
- **Locks module** - Message type restrictions
- **Blacklists module** - Banned word filtering

Modules are loaded in alita/main.go:LoadModules() with help module loaded last.

### Key Patterns

**Database Operations**: All database operations use helper functions with retry logic, timing, and slow query logging (alita/db/db.go:313-499)

**Caching**: Redis-based caching for frequently accessed data (alita/utils/cache/)

**Error Handling**: Structured logging with context and error wrapping

**Testing**: Unit tests for critical modules (antiflood_test.go, captcha_test.go, string_handling_test.go)

## Environment Configuration

Required environment variables (see app.json and sample.env):
- `BOT_TOKEN` - Telegram bot token
- `DB_URI` - MongoDB connection string  
- `DB_NAME` - MongoDB database name (default: Alita_Robot)
- `OWNER_ID` - Bot owner Telegram user ID
- `MESSAGE_DUMP` - Chat ID for bot logs
- `REDIS_ADDRESS` - Redis server address
- `REDIS_PASSWORD` - Redis password

## Database Schema

The bot uses MongoDB with these main collections:
- `admin` - Admin settings per chat
- `filters` - Custom message filters
- `notes` - Saved notes and responses
- `warns_users` - User warning records
- `captchas` - CAPTCHA challenges
- `chats` - Chat configuration
- `users` - User data

All collections have optimized indexes for performance (see alita/db/db.go:72-260).

## Localization

Internationalization files are stored in `locales/` directory:
- `config.yml` - Locale configuration
- `en.yml` - English translations
- Embedded using `//go:embed` directive

## Performance Considerations

- Connection pooling for MongoDB with configurable limits
- Redis caching for frequently accessed data
- Resource monitoring with goroutine and memory tracking
- Slow query logging for database operations >100ms
- Rate limiting with token bucket algorithm for antiflood

## Deployment

The bot supports multiple deployment methods:
- **Heroku**: Uses `heroku.yml` and `Procfile`
- **Docker**: Production, debug, and local configurations
- **Binary**: Direct execution of compiled binaries
- **Development**: Local Go execution with hot reload