# Configuration Package

This package provides robust configuration management for the Alita Robot with proper validation, error handling, and a modern API.

## Features

- ✅ **Validation**: Validates required fields and data types with detailed error messages
- ✅ **Defaults**: Centralized default values with clear constants
- ✅ **Type Safety**: Proper parsing with error handling for numbers, booleans, and arrays
- ✅ **Environment Variables**: Loads from `.env` files and environment variables
- ✅ **Testing**: Comprehensive unit tests included

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/divideprojects/Alita_Robot/alita/config"
    "github.com/divideprojects/Alita_Robot/alita/utils/logger"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Config error:", err)
    }
    
    // Setup logger
    logger.Setup(cfg.Debug)
    
    // Use configuration - pass cfg to functions that need it
    fmt.Printf("Bot token: %s\n", cfg.BotToken)
    fmt.Printf("Owner ID: %d\n", cfg.OwnerId)
    
    // Pass config to services
    // db.Initialize(cfg)
    // myService.Start(cfg)
}
```

## Configuration Variables

### Required Variables

- `BOT_TOKEN`: Telegram bot token from @BotFather
- `DB_URI`: MongoDB connection URI
- `OWNER_ID`: Bot owner's Telegram user ID
- `MESSAGE_DUMP`: Chat ID for bot logs and startup messages

### Optional Variables

- `BOT_VERSION`: Bot version (default: "2.1.3")
- `DB_NAME`: Database name (default: "Alita_Robot")
- `DEBUG`: Enable debug logging (default: false)
- `DROP_PENDING_UPDATES`: Drop pending updates on startup (default: true)
- `API_SERVER`: Telegram API server (default: "https://api.telegram.org")
- `ENABLED_LOCALES`: Comma-separated locale codes (default: "en")
- `ALLOWED_UPDATES`: Comma-separated Telegram update types (default: all types)
- `REDIS_ADDRESS`: Redis server address (default: "localhost:6379")
- `REDIS_PASSWORD`: Redis password (default: empty)
- `REDIS_DB`: Redis database number (default: 0)

## Environment Setup

1. Copy `sample.env` to `.env`
2. Fill in the required values
3. Optionally set any optional values

The configuration loader will:
1. Try to load from `.env` file
2. Fall back to environment variables
3. Apply defaults for unset optional values
4. Validate all required fields
5. Return detailed errors for any issues

## Error Handling

The configuration system provides detailed validation errors:

```go
cfg, err := config.Load()
if err != nil {
    if configErr, ok := err.(*config.ConfigValidationError); ok {
        // Handle multiple validation errors
        for _, validationErr := range configErr.Errors {
            fmt.Printf("Field %s: %s\n", validationErr.Field, validationErr.Message)
        }
    }
    return err
}
```

## Testing

Run the config tests:

```bash
go test ./alita/config -v
```

The tests cover:
- Valid configuration loading
- Missing required fields
- Invalid data types
- Helper function behavior

## Usage Pattern

The configuration system follows a simple pattern:

1. **Load configuration**: Use `config.Load()` in `main()` to get a validated Config struct
2. **Setup services**: Pass the config to services that need it (logger, database, etc.)
3. **Use throughout app**: Pass the config struct to functions and modules that need it

### Dependency Injection Pattern

```go
// Good: Explicit dependency injection
func MyService(cfg *config.Config) {
    if cfg.Debug {
        log.Debug("Debug mode enabled")
    }
}

// Load and pass config
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

MyService(cfg)
```

This approach makes dependencies explicit, improves testability, and follows Go best practices. 