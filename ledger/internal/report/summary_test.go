package report

import (
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"ledger/core"
	"ledger/internal/cache"
	"ledger/testutil"

	"github.com/redis/go-redis/v9"
)

func TestGetReportSummaryRejectsReversedRange(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	if _, err := New(nil, nil).GetReportSummary(t.Context(), from, from.AddDate(0, 0, -1)); err == nil {
		t.Fatal("expected an error for a reversed date range")
	}
}

func TestGetReportSummaryRedis(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set TEST_REDIS_ADDR to run Redis integration tests")
	}
	t.Setenv("REDIS_ADDR", addr)
	store := testutil.NewStorage(t, "../../migrations/00001_create_budgets_and_expenses.up.sql")
	ctx := t.Context()
	client, err := cache.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	inspect := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("REDIS_PASSWORD"), DB: db})
	defer inspect.Close()
	const key = "report:summary:2040-03-04:2040-03-05"
	if err := inspect.Del(ctx, key).Err(); err != nil {
		t.Fatal(err)
	}
	defer inspect.Del(ctx, key)

	from := time.Date(2040, 3, 4, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)
	add := func(category string, amount float64, date time.Time) {
		t.Helper()
		if _, err := store.AddTransaction(core.Transaction{
			Amount: amount, Category: category, Description: "test expense", Date: date,
		}); err != nil {
			t.Fatal(err)
		}
	}
	reports := New(store, client)
	check := func(want []core.CategoryTotal) {
		t.Helper()
		got, err := reports.GetReportSummary(ctx, from, to)
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("summary = %v, want %v, error = %v", got, want, err)
		}
	}

	add("food", 1.25, from)
	add("food", 2, to)
	add("travel", 8, to)
	add("food", 50, to.AddDate(0, 0, 1))
	initial := []core.CategoryTotal{{Category: "food", Total: 3.25}, {Category: "travel", Total: 8}}
	check(initial)
	if ttl, err := inspect.PTTL(ctx, key).Result(); err != nil || ttl <= 0 || ttl > summaryTTL {
		t.Fatalf("cache TTL = %s, error = %v", ttl, err)
	}

	add("food", 4, from)
	check(initial)

	if err := inspect.Expire(ctx, key, 0).Err(); err != nil {
		t.Fatal(err)
	}
	updated := []core.CategoryTotal{{Category: "food", Total: 7.25}, {Category: "travel", Total: 8}}
	check(updated)

	if err := inspect.Set(ctx, key, "invalid JSON", summaryTTL).Err(); err != nil {
		t.Fatal(err)
	}
	check(updated)

	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	add("food", 1, from)
	check([]core.CategoryTotal{{Category: "food", Total: 8.25}, {Category: "travel", Total: 8}})
}
