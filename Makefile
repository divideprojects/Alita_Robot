.PHONY: run tidy vendor build lint-i18n check-i18n

GO_CMD = go
GORELEASER_CMD = goreleaser

run:
	$(GO_CMD) run main.go

tidy:
	$(GO_CMD) mod tidy

vendor:
	$(GO_CMD) mod vendor

build:
	$(GORELEASER_CMD) release --snapshot --skip=publish --clean --skip=sign

# I18n validation and linting targets
lint-i18n:
	@echo "🔍 Linting i18n key format compliance..."
	@python3 scripts/lint_i18n.py

check-i18n:
	@echo "📊 Checking i18n keys and production readiness..."
	@python3 scripts/check_code_keys.py

# Combined validation target
validate: lint-i18n check-i18n
	@echo "✅ All i18n validation checks completed"

# Development workflow target
dev-check: tidy lint-i18n check-i18n
	@echo "🚀 Development checks completed"
