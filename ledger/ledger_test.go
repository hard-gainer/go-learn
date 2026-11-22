package main_test

import (
	"errors"
	"testing"

	ledger "ledger/core"
)

func TestTransactionValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		amount      float64
		category    string
		description string
		wantErr     bool
	}{
		{"valid", 50, "food", "lunch", false},
		{"valid fractional amount", 0.01, "food", "lunch", false},
		{"zero amount", 0, "food", "lunch", true},
		{"negative amount", -1, "food", "lunch", true},
		{"missing category", 50, "", "lunch", true},
		{"missing description", 50, "food", "", true},
		{"empty transaction", 0, "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx := ledger.Transaction{
				Amount:      tc.amount,
				Category:    tc.category,
				Description: tc.description,
			}
			if err := tx.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, want an error: %t", err, tc.wantErr)
			}
		})
	}
}

func TestStorageBudgetLimit(t *testing.T) {
	store := ledger.NewStorage()
	t.Cleanup(store.Reset)
	store.SetBudget(ledger.Budget{Category: "food", Limit: 50})

	tx := ledger.Transaction{Amount: 50, Category: "food", Description: "lunch"}
	created, err := store.AddTransaction(tx)
	if err != nil || created.ID != 1 {
		t.Fatalf("transaction at the budget limit = %+v, error = %v", created, err)
	}
	if _, err := store.AddTransaction(tx); !errors.Is(err, ledger.ErrBudgetExceeded) {
		t.Fatalf("adding above the limit: error = %v, want ErrBudgetExceeded", err)
	}
	if transactions := store.ListTransactions(); len(transactions) != 1 || transactions[0] != created {
		t.Fatalf("stored transactions = %+v, want [%+v]", transactions, created)
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
