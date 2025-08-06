.PHONY: run tidy vendor build lint psql-prepare psql-migrate psql-status psql-rollback psql-reset

GO_CMD = go
GORELEASER_CMD = goreleaser
GOLANGCI_LINT_CMD = golangci-lint

# PostgreSQL Migration Variables
PSQL_DIR = psql
PSQL_MIGRATIONS_DIR = $(PSQL_DIR)/migrations
SUPABASE_MIGRATIONS_DIR = supabase/migrations

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

# PostgreSQL Migration Targets
psql-prepare:
	@echo "🔧 Preparing PostgreSQL migrations..."
	@mkdir -p $(PSQL_MIGRATIONS_DIR)
	@echo "📝 Cleaning Supabase-specific elements from migrations..."
	@for file in $(SUPABASE_MIGRATIONS_DIR)/*.sql; do \
		filename=$$(basename "$$file"); \
		echo "  Processing $$filename..."; \
		sed -E '/(grant|GRANT).*(anon|authenticated|service_role)/d' "$$file" | \
		sed 's/ with schema "extensions"//g' | \
		sed 's/create extension if not exists/CREATE EXTENSION IF NOT EXISTS/g' | \
		sed 's/create extension/CREATE EXTENSION IF NOT EXISTS/g' > "$(PSQL_MIGRATIONS_DIR)/$$filename"; \
	done
	@echo "📜 Creating migration script..."
	@chmod +x $(PSQL_DIR)/migrate.sh 2>/dev/null || true
	@echo "✅ PostgreSQL migrations prepared in $(PSQL_MIGRATIONS_DIR)"
	@echo "📋 Found $$(ls -1 $(PSQL_MIGRATIONS_DIR)/*.sql 2>/dev/null | wc -l) migration files"

psql-migrate:
	@echo "🚀 Applying PostgreSQL migrations..."
	@if [ -z "$(PSQL_DB_HOST)" ] || [ -z "$(PSQL_DB_NAME)" ] || [ -z "$(PSQL_DB_USER)" ]; then \
		echo "❌ Error: Required environment variables not set"; \
		echo "   Please set: PSQL_DB_HOST, PSQL_DB_NAME, PSQL_DB_USER, PSQL_DB_PASSWORD"; \
		exit 1; \
	fi
	@if [ ! -f "$(PSQL_DIR)/migrate.sh" ]; then \
		echo "❌ Error: migrate.sh not found. Run 'make psql-prepare' first"; \
		exit 1; \
	fi
	@bash $(PSQL_DIR)/migrate.sh

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