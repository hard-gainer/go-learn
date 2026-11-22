package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ledger "ledger/core"
)

func TestListBudgets(t *testing.T) {
	store := ledger.NewStorage()
	mux := http.NewServeMux()
	RegisterRoutes(mux, store)

	check := func(want string) {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/budgets", nil))
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != want {
			t.Fatalf("status = %d, headers = %v, body = %s, want %s", w.Code, w.Header(), w.Body, want)
		}
	}

	check("[]")
	store.SetBudget(ledger.Budget{Category: "travel", Limit: 5000})
	store.SetBudget(ledger.Budget{Category: "food", Limit: 2500})
	check(`[{"category":"food","limit":2500},{"category":"travel","limit":5000}]`)
}

func TestCreateBudget(t *testing.T) {
	store := ledger.NewStorage()
	mux := http.NewServeMux()
	RegisterRoutes(mux, store)

	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(body)))
		return w
	}

	for _, tc := range []struct {
		body      string
		wantError string
	}{
		{`{"category":"food","limit":}`, "invalid JSON:"},
		{`{"category":"food","limit":0}`, "limit must be a positive number"},
	} {
		w := post(tc.body)
		var body map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if w.Code != http.StatusBadRequest || err != nil || !strings.HasPrefix(body["error"], tc.wantError) {
			t.Fatalf("status = %d, body = %s, decode error = %v", w.Code, w.Body, err)
		}
	}
	if len(store.ListBudgets()) != 0 {
		t.Fatal("invalid request saved a budget")
	}

	const budget = `{"category":"food","limit":2500}`
	w := post(budget)
	if w.Code != http.StatusCreated || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != budget {
		t.Fatalf("status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
	}

	w = postTransaction(mux, validTransaction)
	if w.Code != http.StatusConflict {
		t.Fatalf("transaction above the saved limit: status = %d, body = %s", w.Code, w.Body)
	}
}
