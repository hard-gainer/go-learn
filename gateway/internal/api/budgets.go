package api

import (
	"maps"
	"net/http"
	"slices"

	ledger "ledger/core"
)

func listBudgets(list func() map[string]ledger.Budget) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		budgets := list()
		response := make([]BudgetResponse, 0, len(budgets))
		for _, category := range slices.Sorted(maps.Keys(budgets)) {
			response = append(response, NewBudgetResponse(budgets[category]))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

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
