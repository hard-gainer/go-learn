package db

import (
	"os"
	"strings"
	"testing"
)

func TestDSNFromEnv(t *testing.T) {
	for _, name := range []string{"DATABASE_URL", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME"} {
		t.Setenv(name, "")
	}
	if err := os.Unsetenv("DB_PASS"); err != nil {
		t.Fatal(err)
	}

	t.Run("defaults", func(t *testing.T) {
		want := "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
		if got := dsnFromEnv(); got != want {
			t.Fatalf("dsn = %q, want %q", got, want)
		}
	})

	t.Run("individual settings", func(t *testing.T) {
		t.Setenv("DB_HOST", "::1")
		t.Setenv("DB_PORT", "5544")
		t.Setenv("DB_USER", "ledger")
		t.Setenv("DB_PASS", "p@ss:word")
		t.Setenv("DB_NAME", "cash app")
		want := "postgres://ledger:p%40ss%3Aword@[::1]:5544/cash%20app?sslmode=disable"
		if got := dsnFromEnv(); got != want {
			t.Fatalf("dsn = %q, want %q", got, want)
		}
	})

	t.Run("database URL takes precedence", func(t *testing.T) {
		const url = "postgres://ledger:example@db:5432/custom?sslmode=require"
		t.Setenv("DATABASE_URL", url)
		t.Setenv("DB_HOST", "ignored")
		t.Setenv("DB_PORT", "also-ignored")
		if got := dsnFromEnv(); got != url {
			t.Fatalf("dsn = %q, want %q", got, url)
		}
	})
}

func TestOpenRejectsInvalidConnectionSettings(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://ledger:example-secret@localhost/cashapp?sslmode=invalid")
	conn, err := Open()
	if conn != nil {
		_ = conn.Close()
		t.Fatal("invalid connection settings returned a pool")
	}
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") || strings.Contains(err.Error(), "example-secret") {
		t.Fatalf("expected a configuration error without credentials, got %v", err)
	}
}
