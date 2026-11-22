package main_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	ledger "ledger/core"
	"ledger/testutil"
)

const testMigration = "migrations/00001_create_budgets_and_expenses.up.sql"

func TestTransactionValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		tx      ledger.Transaction
		wantErr bool
	}{
		{"valid", ledger.Transaction{Amount: 50, Category: "food", Description: "lunch"}, false},
		{"zero amount", ledger.Transaction{Category: "food", Description: "lunch"}, true},
		{"negative amount", ledger.Transaction{Amount: -1, Category: "food", Description: "lunch"}, true},
		{"missing category", ledger.Transaction{Amount: 50, Description: "lunch"}, true},
		{"missing description", ledger.Transaction{Amount: 50, Category: "food"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.tx.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, want an error: %t", err, tc.wantErr)
			}
		})
	}
}

func TestBudgetValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		limit    float64
		category string
		wantErr  bool
	}{
		{"positive limit", 50, "food", false},
		{"zero limit", 0, "food", true},
		{"negative limit", -1, "food", true},
		{"empty category", 50, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			budget := ledger.Budget{Limit: tc.limit, Category: tc.category}
			if err := budget.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, want an error: %t", err, tc.wantErr)
			}
		})
	}
}

func TestStorageBudgetLimit(t *testing.T) {
	store := testutil.NewStorage(t, testMigration)
	if err := store.SetBudget(ledger.Budget{Category: "еда", Limit: 5000}); err != nil {
		t.Fatal(err)
	}

	created, err := store.AddTransaction(ledger.Transaction{
		Amount: 3000, Category: "еда", Description: "продукты", Date: time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("transaction within the budget: %v", err)
	}
	before, err := store.ListTransactions()
	if err != nil || len(before) != 1 || before[0] != created || created.ID == 0 {
		t.Fatalf("stored transactions = %+v, want [%+v], error = %v", before, created, err)
	}

	_, err = store.AddTransaction(ledger.Transaction{
		Amount: 2500, Category: "еда", Description: "ужин",
	})
	if !errors.Is(err, ledger.ErrBudgetExceeded) {
		t.Fatalf("adding above the limit: error = %v, want ErrBudgetExceeded", err)
	}
	after, err := store.ListTransactions()
	if err != nil || !slices.Equal(before, after) {
		t.Fatalf("stored transactions after rejection = %+v, want %+v, error = %v", after, before, err)
	}
}

func TestStorageUpsertsBudgets(t *testing.T) {
	store := testutil.NewStorage(t, testMigration)
	for _, budget := range []ledger.Budget{
		{Category: "travel", Limit: 50},
		{Category: "food", Limit: 20},
		{Category: "food", Limit: 75.25},
	} {
		if err := store.SetBudget(budget); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.SetBudget(ledger.Budget{Category: "food", Limit: 0}); err == nil {
		t.Fatal("invalid budget was accepted")
	}
	want := []ledger.Budget{{Category: "food", Limit: 75.25}, {Category: "travel", Limit: 50}}
	if got, err := store.ListBudgets(); err != nil || !slices.Equal(got, want) {
		t.Fatalf("budgets = %v, want %v, error = %v", got, want, err)
	}
}

func TestStorageListsTransactionsNewestFirst(t *testing.T) {
	store := testutil.NewStorage(t, testMigration)
	if _, err := store.AddTransaction(ledger.Transaction{Category: "food", Description: "invalid amount"}); err == nil {
		t.Fatal("invalid transaction was accepted")
	}
	newest := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	var created []ledger.Transaction
	for _, date := range []time.Time{newest, newest.AddDate(0, 0, -1), newest} {
		tx, err := store.AddTransaction(ledger.Transaction{Amount: 0.10, Category: "food", Description: "lunch", Date: date})
		if err != nil {
			t.Fatalf("add transaction without budget: %v", err)
		}
		created = append(created, tx)
	}
	want := []ledger.Transaction{created[2], created[0], created[1]}
	if got, err := store.ListTransactions(); err != nil || !slices.Equal(got, want) {
		t.Fatalf("transactions = %v, want %v, error = %v", got, want, err)
	}
}
