# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `make run` - Run the bot using `go run main.go`
- `make build` - Build release binaries using goreleaser
- `make tidy` - Clean up Go modules with `go mod tidy`
- `make vendor` - Vendor dependencies with `go mod vendor`

### Testing
- `go test ./...` - Run all tests
- `go test ./alita/modules/antiflood_test.go` - Run specific module tests
- `go test ./alita/utils/benchmarks/...` - Run benchmark tests
- `go test -bench=.` - Run benchmarks

### Environment Setup
- Copy `sample.env` to `.env` and configure required variables
- Required environment variables: `BOT_TOKEN`, `OWNER_ID`, `DB_URI`, `MESSAGE_DUMP`
- Optional: `DEBUG=true` for debug mode, `REDIS_ADDRESS`, `REDIS_PASSWORD`

## Architecture Overview

### Core Structure
- `main.go` - Entry point that initializes bot, loads modules, and starts polling
- `alita/` - Main bot package containing all functionality
- `alita/main.go` - Bot initialization, module loading, and resource monitoring
- `alita/config/` - Configuration management and environment variable parsing
- `alita/modules/` - Bot command handlers and feature modules
- `alita/db/` - Database layer with MongoDB operations and caching
- `alita/utils/` - Utility functions, helpers, and shared components

### Key Components

#### Module System
- Modules are loaded in `alita/main.go:LoadModules()` in specific order
- Help module loads last to register all commands
- Each module in `alita/modules/` handles specific bot functionality (admin, bans, filters, etc.)
- Modules use the `ext.Dispatcher` pattern from gotgbot for handling updates

#### Database Layer
- MongoDB as primary database with connection pooling
- Redis for caching with configurable settings
- Database files in `alita/db/` handle specific data types (users, chats, filters, etc.)
- Pagination utilities in `alita/db/pagination.go`

#### Configuration
- Environment-based configuration in `alita/config/config.go`
- Supports both debug and release modes
- Configurable resource monitoring thresholds
- Internationalization support with locale loading

#### Caching System
- Multi-level caching using Redis and in-memory (Ristretto)
- Cache initialization in `alita/utils/cache/`
- Admin cache for permission checks

### Bot Features
- Multi-language support with embedded locales
- CAPTCHA system with scheduler
- Admin management and permissions
- Flood protection and anti-spam
- Filters, notes, and blacklists
- Connection system for managing multiple chats
- Warns and mute system
- Greeting and rules management

### Development Patterns
- Use structured logging with logrus
- Follow the existing module pattern when adding new features
- Database operations should include proper error handling
- Use the cache system for frequently accessed data
- Resource monitoring is built-in with configurable thresholds

### Docker Support
- `docker-compose.yml` for production deployment
- `debug.docker-compose.yml` for development
- `local.docker-compose.yml` for local testing
- Alpine-based Docker images for minimal footprint