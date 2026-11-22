GOOSE ?= ledger/.bin/goose
LEDGER_MIGRATION := 00001_create_budgets_and_expenses

.PHONY: migrate-up migrate-down

migrate-up migrate-down:
	@set -eu; \
	database_url="$${DATABASE_URL:-$${LEDGER_DATABASE_URL:-}}"; \
	: "$${database_url:?Set DATABASE_URL to the PostgreSQL connection string}"; \
	migration_dir=$$(mktemp -d); \
	migration_file="$$migration_dir/$(LEDGER_MIGRATION).sql"; \
	trap 'rm -f -- "$$migration_file"; rmdir -- "$$migration_dir"' EXIT; \
	cat "ledger/migrations/$(LEDGER_MIGRATION).up.sql" \
	    "ledger/migrations/$(LEDGER_MIGRATION).down.sql" > "$$migration_file"; \
	"$(GOOSE)" -dir "$$migration_dir" postgres "$$database_url" $(@:migrate-%=%)
