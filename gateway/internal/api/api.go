package api

import (
	"net/http"
	"time"

	ledger "ledger/core"
)

type CreateTransactionRequest struct {
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type TransactionResponse struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (req CreateTransactionRequest) ToDomain() ledger.Transaction {
	return ledger.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}
}

func NewTransactionResponse(tx ledger.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func (req CreateBudgetRequest) ToDomain() ledger.Budget {
	return ledger.Budget{
		Category: req.Category,
		Limit:    req.Limit,
	}
}

func NewBudgetResponse(budget ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: budget.Category,
		Limit:    budget.Limit,
	}
}

func RegisterRoutes(mux *http.ServeMux, store *ledger.Storage) {
	const prefix = "/api"

	mux.HandleFunc("GET "+prefix+"/transactions", listTransactions(store.ListTransactions))
	mux.HandleFunc("POST "+prefix+"/transactions", createTransaction(store.AddTransaction))
	mux.HandleFunc("GET "+prefix+"/budgets", listBudgets(store.ListBudgets))
	mux.HandleFunc("POST "+prefix+"/budgets", createBudget(store.SetBudget))
	mux.HandleFunc("PUT "+prefix+"/budgets/{category}", notImplemented)
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}
