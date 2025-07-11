# Configuration Package

This package provides robust configuration management for the Alita Robot with proper validation, error handling, and both modern and backward-compatible APIs.

## Features

- ✅ **Validation**: Validates required fields and data types with detailed error messages
- ✅ **Defaults**: Centralized default values with clear constants
- ✅ **Type Safety**: Proper parsing with error handling for numbers, booleans, and arrays
- ✅ **Environment Variables**: Loads from `.env` files and environment variables
- ✅ **Backward Compatibility**: Maintains the old global variable API
- ✅ **Testing**: Comprehensive unit tests included

## Quick Start

### Basic Usage (Recommended)

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
    
    // Use configuration
    fmt.Printf("Bot token: %s\n", cfg.BotToken)
    fmt.Printf("Owner ID: %d\n", cfg.OwnerId)
}
```

### Backward Compatible Usage

For existing code that uses global variables:

```go
import "github.com/divideprojects/Alita_Robot/alita/config"

func main() {
    // Load and set global variables
    if err := config.LoadAndSetGlobals(); err != nil {
        log.Fatal("Config error:", err)
    }
    
    // Use the old global variables
    fmt.Printf("Bot token: %s\n", config.BotToken)
    fmt.Printf("Owner ID: %d\n", config.OwnerId)
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

## Migration from Old Config

The old config system used global variables and an `init()` function that had side effects. The new system:

1. **Separates concerns**: Configuration loading is separate from logger setup
2. **Provides validation**: Required fields are validated with clear error messages  
3. **Improves testability**: Pure functions that can be tested in isolation
4. **Maintains compatibility**: Old global variables still work via `LoadAndSetGlobals()`

To migrate existing code:
1. Replace config imports with explicit config loading in `main()`
2. Set up logger explicitly after loading config
3. Use the new Config struct instead of global variables (recommended)
4. Update any code that depended on config's init() side effects 