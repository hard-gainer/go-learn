package testutil

import (
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"ledger/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// NewStorage runs the schema migration in a separate PostgreSQL schema for each test.
func NewStorage(t *testing.T, migrationPath string) *core.Storage {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatal(err)
	}

	admin := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = admin.Close() })
	schema := "ledger_test_" + strings.ToLower(rand.Text())
	if _, err := admin.ExecContext(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	config.RuntimeParams["search_path"] = schema
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	migrator := stdlib.OpenDB(*config)
	_, err = migrator.ExecContext(t.Context(), string(migration))
	_ = migrator.Close()
	if err != nil {
		t.Fatalf("apply migration to test schema: %v", err)
	}

	config.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	store := core.NewStorage(stdlib.OpenDB(*config))
	t.Cleanup(func() { _ = store.Close() })
	return store
}
