package api

import (
	"errors"
	"log"
	"net/http"

	ledger "ledger/core"
)

func listTransactions(list func() []ledger.Transaction) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		transactions := list()
		response := make([]TransactionResponse, 0, len(transactions))
		for _, tx := range transactions {
			response = append(response, NewTransactionResponse(tx))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func createTransaction(addTransaction func(ledger.Transaction) (ledger.Transaction, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateTransactionRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		tx := req.ToDomain()
		if err := tx.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		created, err := addTransaction(tx)
		switch {
		case errors.Is(err, ledger.ErrBudgetExceeded):
			writeError(w, http.StatusConflict, ledger.ErrBudgetExceeded.Error())
		case err != nil:
			log.Printf("add transaction: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		default:
			writeJSON(w, http.StatusCreated, NewTransactionResponse(created))
		}
	}
}
