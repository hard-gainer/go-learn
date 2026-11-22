package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
	dsn := dsnFromEnv()
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL connection pool: %w", err)
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		var parseErr *pgconn.ParseConfigError
		if errors.As(err, &parseErr) {
			return nil, errors.New("invalid PostgreSQL connection settings in DATABASE_URL or DB_*")
		}
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return conn, nil
}

func dsnFromEnv() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	password, set := os.LookupEnv("DB_PASS")
	if !set {
		password = "postgres"
	}
	database := envOrDefault("DB_NAME", "cashapp")
	dsn := url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort(envOrDefault("DB_HOST", "localhost"), envOrDefault("DB_PORT", "5432")),
		User:     url.UserPassword(envOrDefault("DB_USER", "postgres"), password),
		Path:     "/" + database,
		RawPath:  "/" + url.PathEscape(database),
		RawQuery: "sslmode=disable",
	}
	return dsn.String()
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
