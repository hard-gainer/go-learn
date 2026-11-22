package api

import (
	"net/http"

	ledger "ledger/core"
)

func createBudget(setBudget func(ledger.Budget)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateBudgetRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		budget := req.ToDomain()
		if err := budget.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		setBudget(budget)
		writeJSON(w, http.StatusCreated, NewBudgetResponse(budget))
	}
}
