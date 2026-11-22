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
	store := newTestStorage(t)
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
	setTestBudget(t, store, ledger.Budget{Category: "travel", Limit: 5000})
	setTestBudget(t, store, ledger.Budget{Category: "food", Limit: 2500})
	check(`[{"category":"food","limit":2500},{"category":"travel","limit":5000}]`)
}

func TestCreateBudget(t *testing.T) {
	post := func(mux http.Handler, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}

	for _, tc := range []struct {
		name      string
		body      string
		wantError string
	}{
		{"invalid_json", `{"category":"food","limit":}`, "invalid JSON:"},
		{"zero_limit", `{"category":"food","limit":0}`, "limit must be a positive number"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStorage(t)
			mux := http.NewServeMux()
			RegisterRoutes(mux, store)

			w := post(mux, tc.body)
			var body map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &body)
			if w.Code != http.StatusBadRequest || w.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
				err != nil || len(body) != 1 || !strings.HasPrefix(body["error"], tc.wantError) {
				t.Fatalf("status = %d, headers = %v, body = %s, decode error = %v", w.Code, w.Header(), w.Body, err)
			}
			if budgets, err := store.ListBudgets(); err != nil || len(budgets) != 0 {
				t.Fatalf("budgets after invalid request = %v, error = %v", budgets, err)
			}
		})
	}

	t.Run("created_and_listed", func(t *testing.T) {
		mux := http.NewServeMux()
		RegisterRoutes(mux, newTestStorage(t))

		const budget = `{"category":"food","limit":2500}`
		w := post(mux, budget)
		if w.Code != http.StatusCreated || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != budget {
			t.Fatalf("status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
		}

		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/budgets", nil))
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != "["+budget+"]" {
			t.Fatalf("list budgets: status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
		}

		w = postTransaction(mux, validTransaction)
		if w.Code != http.StatusConflict {
			t.Fatalf("transaction above the saved limit: status = %d, body = %s", w.Code, w.Body)
		}
	})
}
