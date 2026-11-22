package api

import (
	"net/http"

	ledger "ledger/core"
)

func listBudgets(list func() ([]ledger.Budget, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		budgets, err := list()
		if err != nil {
			writeInternalError(w, "list budgets", err)
			return
		}
		response := make([]BudgetResponse, 0, len(budgets))
		for _, budget := range budgets {
			response = append(response, NewBudgetResponse(budget))
		}
		WriteJSON(w, http.StatusOK, response)
	}
}

func createBudget(setBudget func(ledger.Budget) error) http.HandlerFunc {
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

		if err := setBudget(budget); err != nil {
			writeInternalError(w, "set budget", err)
			return
		}
		WriteJSON(w, http.StatusCreated, NewBudgetResponse(budget))
	}
}
