# Alita Robot - Agent Guidelines

## Build/Lint/Test Commands
```bash
make run          # Run the bot locally
make tidy         # Clean up go.mod and go.sum
make vendor       # Create vendor directory
make build        # Build release artifacts with goreleaser
go vet ./...      # Run Go vet for static analysis
go mod tidy       # Ensure dependencies are correct
```

## Code Style Guidelines
- **Language**: Go 1.23+ (see go.mod)
- **Imports**: Group stdlib, external deps, then internal packages. Use goimports formatting
- **Functions**: Document exported functions with comments starting with function name
- **Error Handling**: Use error_handling.FatalError() for logging with context, HandleErr() for simple cases
- **Logging**: Use sirupsen/logrus with appropriate log levels (Error, Info, Debug)
- **Database**: MongoDB for persistence, Redis for caching (via eko/gocache)
- **Bot Framework**: PaulSonOfLars/gotgbot/v2 for Telegram API
- **Config**: Environment variables loaded via godotenv, parsed with custom typeConvertor
- **Naming**: CamelCase for exports, camelCase for internal. Descriptive names preferred
- **Testing**: No test files found - create *_test.go files when adding new features