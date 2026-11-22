package main_test

import (
	"errors"
	"testing"

	ledger "ledger/core"
)

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
	store := ledger.NewStorage()
	t.Cleanup(store.Reset)
	store.SetBudget(ledger.Budget{Category: "еда", Limit: 5000})

	created, err := store.AddTransaction(ledger.Transaction{
		Amount: 3000, Category: "еда", Description: "продукты",
	})
	if err != nil {
		t.Fatalf("transaction within the budget: %v", err)
	}
	before := store.ListTransactions()
	if len(before) != 1 || before[0] != created {
		t.Fatalf("stored transactions = %+v, want [%+v]", before, created)
	}

	_, err = store.AddTransaction(ledger.Transaction{
		Amount: 2500, Category: "еда", Description: "ужин",
	})
	if !errors.Is(err, ledger.ErrBudgetExceeded) {
		t.Fatalf("adding above the limit: error = %v, want ErrBudgetExceeded", err)
	}
	after := store.ListTransactions()
	if len(after) != len(before) || after[0] != created {
		t.Fatalf("stored transactions after rejection = %+v, want %+v", after, before)
	}
}

func TestStorageReset(t *testing.T) {
	store := ledger.NewStorage()
	t.Cleanup(store.Reset)
	store.SetBudget(ledger.Budget{Category: "food", Limit: 50})

	tx := ledger.Transaction{Amount: 50, Category: "food", Description: "lunch"}
	if _, err := store.AddTransaction(tx); err != nil {
		t.Fatal(err)
	}

	store.Reset()
	if budgets, transactions := store.ListBudgets(), store.ListTransactions(); len(budgets) != 0 || len(transactions) != 0 {
		t.Fatalf("after reset: budgets = %v, transactions = %v", budgets, transactions)
	}

	store.SetBudget(ledger.Budget{Category: "food", Limit: 75})
	tx.Amount = 75
	created, err := store.AddTransaction(tx)
	if err != nil || created.ID != 1 {
		t.Fatalf("first transaction after reset = %+v, error = %v", created, err)
	}
}
