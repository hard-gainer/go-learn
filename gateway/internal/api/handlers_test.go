package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ledger "ledger/core"
)

func newTestStorage(t *testing.T) *ledger.Storage {
	t.Helper()
	store := ledger.NewStorage()
	t.Cleanup(store.Reset)
	return store
}

func TestBudgetAndTransactionFlow(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, newTestStorage(t))

	const (
		budget      = `{"category":"food","limit":3000}`
		transaction = `{"id":1,"amount":3000,"category":"food","description":"lunch","date":"2026-11-22T10:00:00Z"}`
	)
	for _, step := range []struct {
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{http.MethodGet, "/api/budgets", "", http.StatusOK, "[]"},
		{http.MethodGet, "/api/transactions", "", http.StatusOK, "[]"},
		{http.MethodPost, "/api/budgets", budget, http.StatusCreated, budget},
		{http.MethodPost, "/api/transactions", validTransaction, http.StatusCreated, transaction},
		{http.MethodPost, "/api/transactions", validTransaction, http.StatusConflict, `{"error":"budget exceeded"}`},
		{http.MethodGet, "/api/budgets", "", http.StatusOK, "[" + budget + "]"},
		{http.MethodGet, "/api/transactions", "", http.StatusOK, "[" + transaction + "]"},
	} {
		req := httptest.NewRequest(step.method, step.path, strings.NewReader(step.body))
		if step.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != step.wantStatus || strings.TrimSpace(w.Body.String()) != step.wantBody {
			t.Fatalf("%s %s: status = %d, body = %s, want %d and %s", step.method, step.path, w.Code, w.Body, step.wantStatus, step.wantBody)
		}
	}
}
