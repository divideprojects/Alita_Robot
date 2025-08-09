.PHONY: run tidy vendor build lint psql-prepare psql-migrate psql-status psql-rollback psql-reset psql-verify i18n-gen i18n-flatten i18n-lint i18n-check i18n-clean

GO_CMD = go
GORELEASER_CMD = goreleaser
GOLANGCI_LINT_CMD = golangci-lint

# PostgreSQL Migration Variables
PSQL_SCRIPT = scripts/migrate_psql.sh
PSQL_MIGRATIONS_DIR ?= tmp/migrations_cleaned
SUPABASE_MIGRATIONS_DIR ?= supabase/migrations

run:
	$(GO_CMD) run main.go

tidy:
	$(GO_CMD) mod tidy

vendor:
	$(GO_CMD) mod vendor

build:
	$(GORELEASER_CMD) release --snapshot --skip=publish --clean --skip=sign

lint:
	@which $(GOLANGCI_LINT_CMD) > /dev/null || (echo "golangci-lint not found, install it from https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT_CMD) run
	@echo "🔍 Checking for i18n issues..."
	@$(GO_CMD) run cmd/i18nlint/main.go -quiet 2>/dev/null || echo "⚠️  Some i18n issues found (non-blocking)"

# PostgreSQL Migration Targets
psql-prepare:
	@echo "🔧 Preparing PostgreSQL migrations (cleaning Supabase SQL)..."
	@mkdir -p $(PSQL_MIGRATIONS_DIR)
	@for file in $(SUPABASE_MIGRATIONS_DIR)/*.sql; do \
		filename=$$(basename "$$file"); \
		echo "  Processing $$filename..."; \
		sed -E '/(grant|GRANT).*(anon|authenticated|service_role)/d' "$$file" | \
		sed 's/ with schema "extensions"//g' | \
		sed 's/create extension if not exists/CREATE EXTENSION IF NOT EXISTS/g' | \
		sed 's/create extension/CREATE EXTENSION IF NOT EXISTS/g' > "$(PSQL_MIGRATIONS_DIR)/$$filename"; \
	done
	@echo "✅ PostgreSQL migrations prepared in $(PSQL_MIGRATIONS_DIR)"
	@echo "📋 Found $$(ls -1 $(PSQL_MIGRATIONS_DIR)/*.sql 2>/dev/null | wc -l) migration files"

psql-migrate:
	@echo "🚀 Applying PostgreSQL migrations..."
	@if [ -z "$(PSQL_DB_HOST)" ] || [ -z "$(PSQL_DB_NAME)" ] || [ -z "$(PSQL_DB_USER)" ]; then \
		echo "❌ Error: Required environment variables not set"; \
		echo "   Please set: PSQL_DB_HOST, PSQL_DB_NAME, PSQL_DB_USER, PSQL_DB_PASSWORD"; \
		exit 1; \
	fi
	@chmod +x $(PSQL_SCRIPT) 2>/dev/null || true
	@bash $(PSQL_SCRIPT)

psql-status:
	@echo "📊 PostgreSQL Migration Status"
	@if [ -z "$(PSQL_DB_HOST)" ] || [ -z "$(PSQL_DB_NAME)" ] || [ -z "$(PSQL_DB_USER)" ]; then \
		echo "❌ Error: Required environment variables not set"; \
		echo "   Please set: PSQL_DB_HOST, PSQL_DB_NAME, PSQL_DB_USER, PSQL_DB_PASSWORD"; \
		exit 1; \
	fi
	@echo "🔍 Checking migration status..."
	@PGPASSWORD=$(PSQL_DB_PASSWORD) psql -h $(PSQL_DB_HOST) -p $${PSQL_DB_PORT:-5432} -U $(PSQL_DB_USER) -d $(PSQL_DB_NAME) -c \
		"SELECT version, executed_at FROM schema_migrations ORDER BY executed_at DESC;" 2>/dev/null || \
		echo "⚠️  No migrations table found. Run 'make psql-migrate' to initialize."

psql-rollback:
	@echo "⏪ Rolling back last PostgreSQL migration..."
	@if [ -z "$(PSQL_DB_HOST)" ] || [ -z "$(PSQL_DB_NAME)" ] || [ -z "$(PSQL_DB_USER)" ]; then \
		echo "❌ Error: Required environment variables not set"; \
		exit 1; \
	fi
	@echo "⚠️  Rollback functionality requires manual intervention"
	@echo "   Last applied migration:"
	@PGPASSWORD=$(PSQL_DB_PASSWORD) psql -h $(PSQL_DB_HOST) -p $${PSQL_DB_PORT:-5432} -U $(PSQL_DB_USER) -d $(PSQL_DB_NAME) -t -c \
		"SELECT version FROM schema_migrations ORDER BY executed_at DESC LIMIT 1;" 2>/dev/null

psql-reset:
	@echo "🔥 WARNING: This will DROP ALL TABLES in the database!"
	@echo "   Database: $(PSQL_DB_NAME) on $(PSQL_DB_HOST)"
	@echo "   Type 'yes' to confirm: " && read confirm && [ "$$confirm" = "yes" ] || (echo "Cancelled" && exit 1)
	@echo "💣 Resetting database..."
	@PGPASSWORD=$(PSQL_DB_PASSWORD) psql -h $(PSQL_DB_HOST) -p $${PSQL_DB_PORT:-5432} -U $(PSQL_DB_USER) -d $(PSQL_DB_NAME) -c \
		"DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "✅ Database reset complete"

psql-verify:
	@echo "🔎 Verifying cleaned migrations are in sync"
	@TMP=$$(mktemp -d); \
	echo "Using temp dir: $$TMP"; \
	$(MAKE) --no-print-directory psql-prepare PSQL_MIGRATIONS_DIR="$$TMP"; \
	git diff --no-index --exit-code $(PSQL_MIGRATIONS_DIR) "$$TMP" || (echo "❌ Drift detected between supabase/migrations and $(PSQL_MIGRATIONS_DIR)" && exit 1)

# i18n Tooling Targets
i18n-gen:
	@echo "🔧 Running i18n code generator..."
	@mkdir -p generated/i18n
	$(GO_CMD) run cmd/i18ngen/main.go
	@echo "✅ i18n code generation complete"

i18n-flatten:
	@echo "🔧 Flattening YAML translation files..."
	@echo "⚠️  This will modify files in locales/ directory"
	@echo "Original files will be backed up with timestamp suffix"
	$(GO_CMD) run cmd/i18nflatten/main.go
	@echo "✅ YAML flattening complete"

i18n-lint:
	@echo "🔍 Running i18n linter and validation..."
	@mkdir -p generated/i18n
	$(GO_CMD) run cmd/i18nlint/main.go

i18n-check: i18n-gen i18n-lint
	@echo "🎯 Running all i18n tools..."
	@echo "✅ All i18n checks complete"

i18n-clean:
	@echo "🧹 Cleaning i18n generated files and backups..."
	@chmod +x scripts/clean_i18n.sh
	@./scripts/clean_i18n.sh