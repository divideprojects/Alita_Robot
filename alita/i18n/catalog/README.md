# I18n Message Catalog System

A modern, simplified i18n system for the Alita Robot with embedded English defaults and flat key structure.

## Overview

This new catalog system replaces the complex viper-based i18n implementation with a simpler approach that:

- Embeds English defaults directly in code
- Uses flat key structure (`admin.promote_success` instead of nested `strings.Admin.promote.success`)
- Supports simple `{param}` style parameter interpolation
- Provides compile-time safety through message registration
- Works without external configuration dependencies

## Key Features

- **Embedded Defaults**: English text is embedded in code, eliminating missing translation issues
- **Flat Keys**: Simple dot-notation keys like `admin.promote_success`
- **Simple Interpolation**: Uses `{param}` syntax instead of complex formats
- **Parameter Validation**: Optional strict validation of message parameters
- **Thread-Safe**: Concurrent access safe with read-write mutexes
- **No Viper Dependency**: Pure Go implementation with standard YAML loading
- **Fallback Support**: Automatic fallback to English defaults
- **Statistics**: Built-in analytics for translation coverage

## Basic Usage

### 1. Register Messages

Register messages with embedded English defaults in your init functions:

```go
package admin

import "github.com/divideprojects/alita_robot/alita/i18n/catalog"

func init() {
    // Register admin messages
    catalog.MustRegister("admin.promote_success", "Successfully promoted {user}!", "user")
    catalog.MustRegister("admin.demote_success", "Successfully demoted {user}!", "user") 
    catalog.MustRegister("admin.cannot_promote_self", "I cannot promote myself!")
    catalog.MustRegister("admin.user_already_admin", "{user} is already an admin!", "user")
}
```

### 2. Initialize Translator Manager

Set up the global translator manager in your main function:

```go
func main() {
    config := catalog.DefaultConfig()
    catalog.InitGlobalManager(config, "./locales")
    
    // Your bot code here...
}
```

### 3. Use Messages

Get translators and use messages:

```go
func promoteUser(userID int64, lang string) {
    translator, err := catalog.T(lang)
    if err != nil {
        translator, _ = catalog.T("en") // Fallback to English
    }
    
    params := catalog.Params{
        "user": getUserName(userID),
    }
    
    message := translator.Message("admin.promote_success", params)
    sendMessage(message)
}
```

## YAML Translation Files

Create flat YAML files for translations:

```yaml
# locales/en.yaml
admin.promote_success: "Successfully promoted {user}!"
admin.demote_success: "Successfully demoted {user}!"
admin.cannot_promote_self: "I cannot promote myself!"
admin.user_already_admin: "{user} is already an admin!"

greetings.welcome: "Welcome {user} to {chat}!"
greetings.goodbye: "Goodbye {user}!"

errors.user_not_found: "User not found!"
errors.bot_not_admin: "I need admin rights to do that!"
```

```yaml
# locales/es.yaml  
admin.promote_success: "¡{user} promovido exitosamente!"
admin.demote_success: "¡{user} degradado exitosamente!"
admin.cannot_promote_self: "¡No puedo promocionarme a mí mismo!"
admin.user_already_admin: "¡{user} ya es un administrador!"

greetings.welcome: "¡Bienvenido {user} a {chat}!"
greetings.goodbye: "¡Adiós {user}!"
```

## Advanced Usage

### Parameter Validation

Enable strict parameter validation:

```go
config := catalog.Config{
    DefaultLanguage:          "en",
    StrictValidation:         true,
    AllowMissingTranslations: true,
}
```

### Plural Messages

Handle plural forms:

```go
// Register plural forms
catalog.MustRegister("user.count.zero", "No users")
catalog.MustRegister("user.count.one", "One user")  
catalog.MustRegister("user.count.other", "{count} users", "count")

// Use plural messages
count := 5
params := catalog.Params{"count": count}
message := translator.Plural("user.count", count, params)
// Returns: "5 users"
```

### Statistics and Analytics

Get translation coverage statistics:

```go
// Catalog statistics
stats := catalog.GetStats()
fmt.Printf("Total messages: %d\n", stats.TotalMessages)
fmt.Printf("Top prefixes: %v\n", stats.TopPrefixes)

// Translator statistics  
translator, _ := catalog.T("es")
fmt.Printf("Spanish coverage: %.1f%%\n", translator.Coverage())
fmt.Printf("Missing keys: %v\n", translator.MissingKeys())
```

### Error Handling

Handle translation errors gracefully:

```go
message, err := translator.GetMessage("admin.promote_success", params)
if err != nil {
    log.Printf("Translation error: %v", err)
    // message will contain the default English text or key placeholder
}
```

## Migration from Old System

### Key Structure Changes

```go
// Old nested structure
"strings.Admin.promote.success" 

// New flat structure  
"admin.promote_success"
```

### Parameter Changes

```go
// Old sprintf style
"Successfully promoted %s!"

// New template style
"Successfully promoted {user}!"
```

### Usage Changes

```go
// Old way
translator.GetString("strings.Admin.promote.success", i18n.TranslationParams{"0": userName})

// New way
translator.Message("admin.promote_success", catalog.Params{"user": userName})
```

## API Reference

### Core Types

- `Message`: Represents a translatable message with key, default text, and expected parameters
- `Params`: Map of parameter values for interpolation (`map[string]any`)
- `Translator`: Provides translation methods for a specific language
- `TranslatorManager`: Manages multiple translators for different languages

### Key Functions

- `Register(key, defaultText, params...)`: Register a message with the global catalog
- `T(lang)`: Get translator for a language
- `Message(key, params)`: Get translated message with parameter interpolation
- `Plural(key, count, params)`: Get pluralized message based on count

### Configuration

```go
type Config struct {
    DefaultLanguage          string // Default: "en"
    StrictValidation         bool   // Default: false  
    CacheTranslations        bool   // Default: true
    AllowMissingTranslations bool   // Default: true
}
```

## Best Practices

1. **Register Early**: Register all messages in init functions
2. **Use Meaningful Keys**: Choose descriptive, hierarchical keys like `admin.promote_success`
3. **Embed Defaults**: Always provide English default text
4. **Validate Parameters**: List expected parameters for documentation and validation
5. **Handle Errors**: Always handle translation errors gracefully
6. **Test Coverage**: Use statistics to ensure good translation coverage
7. **Consistent Naming**: Use consistent prefixes for related functionality

## Performance

- Message registration: O(1) with concurrent safety
- Message lookup: O(1) map access
- Parameter interpolation: O(n) where n is parameter count
- Translation loading: Lazy loading with caching
- Memory usage: Minimal overhead with shared catalog

## Thread Safety

All operations are thread-safe:
- Concurrent message registration (init functions)
- Concurrent translator creation and access
- Concurrent message translation and interpolation
- Safe for use in HTTP handlers and goroutines

## Error Types

- `CatalogError`: Errors in catalog operations (registration, lookup)
- `ValidationError`: Parameter validation failures
- `InterpolationError`: Parameter interpolation failures

All errors implement the standard Go error interface and can be unwrapped for root cause analysis.