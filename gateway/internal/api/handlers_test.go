package api

import (
	"encoding/json"
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
	const (
		budget      = `{"category":"food","limit":5000}`
		input       = `{"amount":3000,"category":"food","description":"lunch","date":"2026-11-22"}`
		transaction = `{"id":1,"amount":3000,"category":"food","description":"lunch","date":"2026-11-22T00:00:00Z"}`
	)
	request := func(t *testing.T, handler http.Handler, method, path, body string, wantStatus int) string {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != wantStatus || w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Fatalf("%s %s: status = %d, want %d; headers = %v, body = %s", method, path, w.Code, wantStatus, w.Header(), w.Body)
		}
		return strings.TrimSpace(w.Body.String())
	}
	setup := func(t *testing.T) http.Handler {
		t.Helper()
		mux := http.NewServeMux()
		RegisterRoutes(mux, newTestStorage(t))
		if got := request(t, mux, http.MethodPost, "/api/budgets", budget, http.StatusCreated); got != budget {
			t.Fatalf("created budget = %s, want %s", got, budget)
		}
		return mux
	}

	t.Run("ok", func(t *testing.T) {
		handler := setup(t)
		if got := request(t, handler, http.MethodPost, "/api/transactions", input, http.StatusCreated); got != transaction {
			t.Fatalf("created transaction = %s, want %s", got, transaction)
		}
		if got, want := request(t, handler, http.MethodGet, "/api/transactions", "", http.StatusOK), "["+transaction+"]"; got != want {
			t.Fatalf("transactions = %s, want %s", got, want)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		handler := setup(t)
		request(t, handler, http.MethodPost, "/api/transactions", input, http.StatusCreated)
		const excess = `{"amount":2500,"category":"food","description":"dinner","date":"2026-11-22"}`
		got := request(t, handler, http.MethodPost, "/api/transactions", excess, http.StatusConflict)
		if want := `{"error":"budget exceeded"}`; got != want {
			t.Fatalf("error response = %s, want %s", got, want)
		}
		if got, want := request(t, handler, http.MethodGet, "/api/transactions", "", http.StatusOK), "["+transaction+"]"; got != want {
			t.Fatalf("transactions after conflict = %s, want %s", got, want)
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		handler := setup(t)
		got := request(t, handler, http.MethodPost, "/api/transactions", `{"amount":}`, http.StatusBadRequest)
		var body map[string]string
		if err := json.Unmarshal([]byte(got), &body); err != nil || len(body) != 1 || !strings.HasPrefix(body["error"], "invalid JSON: ") {
			t.Fatalf("error response = %s, decode error = %v", got, err)
		}
	})
}
