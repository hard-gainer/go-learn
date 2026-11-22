package api

import (
	"encoding/json"
	"errors"
	"io"
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
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
			writeError(w, http.StatusBadRequest, "request body must contain exactly one JSON object")
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

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
