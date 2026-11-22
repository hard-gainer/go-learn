package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ledger "ledger/core"
)

const validTransaction = `{"amount":3000,"category":"food","description":"lunch","date":"2026-11-22T10:00:00Z"}`

func TestListTransactions(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, newTestStorage(t))

	get := func() string {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/transactions", nil))
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Fatalf("status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
		}
		return strings.TrimSpace(w.Body.String())
	}

	if got := get(); got != "[]" {
		t.Fatalf("empty list = %s, want []", got)
	}

	created := postTransaction(mux, validTransaction)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body)
	}
	want := "[" + strings.TrimSpace(created.Body.String()) + "]"
	if got := get(); got != want {
		t.Fatalf("list = %s, want %s", got, want)
	}
}

func postTransaction(handler http.Handler, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body)))
	return w
}

func TestCreateTransaction(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{"invalid JSON", `{"amount":}`, http.StatusBadRequest, "invalid JSON:"},
		{"trailing JSON", validTransaction + " {}", http.StatusBadRequest, "request body must contain exactly one JSON object"},
		{"invalid date", strings.Replace(validTransaction, "2026-11-22T10:00:00Z", "2026-02-30", 1), http.StatusBadRequest, "invalid JSON: date must be YYYY-MM-DD or RFC 3339"},
		{"invalid amount", strings.Replace(validTransaction, "3000", "-1", 1), http.StatusBadRequest, "amount must be a positive number"},
		{"created", validTransaction, http.StatusCreated, ""},
		{"budget exceeded", validTransaction, http.StatusConflict, "budget exceeded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestStorage(t)
			setTestBudget(t, store, ledger.Budget{Category: "food", Limit: 5000})
			mux := http.NewServeMux()
			RegisterRoutes(mux, store)

			if tc.wantStatus == http.StatusConflict {
				first := postTransaction(mux, validTransaction)
				if first.Code != http.StatusCreated {
					t.Fatalf("initial transaction: status = %d, body = %s", first.Code, first.Body)
				}
			}
			before := len(storedTransactions(t, store))
			w := postTransaction(mux, tc.body)
			if w.Code != tc.wantStatus || w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
			}
			if tc.wantError != "" {
				var body map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || !strings.HasPrefix(body["error"], tc.wantError) {
					t.Fatalf("body = %s, decode error = %v", w.Body, err)
				}
				if got := len(storedTransactions(t, store)); got != before {
					t.Fatalf("stored transactions = %d, want %d", got, before)
				}
				return
			}
			var tx TransactionResponse
			if err := json.Unmarshal(w.Body.Bytes(), &tx); err != nil || tx.ID != 1 || tx.Amount != 3000 || tx.Category != "food" || tx.Description != "lunch" || tx.Date.IsZero() {
				t.Fatalf("transaction = %+v, decode error = %v", tx, err)
			}
			if got := len(storedTransactions(t, store)); got != before+1 {
				t.Fatalf("stored transactions = %d, want %d", got, before+1)
			}
		})
	}
}

func TestCreateTransactionUnexpectedError(t *testing.T) {
	t.Parallel()

	handler := createTransaction(func(ledger.Transaction) (ledger.Transaction, error) {
		return ledger.Transaction{}, errors.New("unexpected storage error")
	})
	w := postTransaction(handler, validTransaction)
	if w.Code != http.StatusInternalServerError || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
}

func TestConcurrentTransactionsRespectBudget(t *testing.T) {
	store := newTestStorage(t)
	setTestBudget(t, store, ledger.Budget{Category: "food", Limit: 5000})
	mux := http.NewServeMux()
	RegisterRoutes(mux, store)

	statuses := make(chan int, 2)
	for range 2 {
		go func() { statuses <- postTransaction(mux, validTransaction).Code }()
	}
	counts := make(map[int]int)
	for range 2 {
		counts[<-statuses]++
	}
	if counts[http.StatusCreated] != 1 || counts[http.StatusConflict] != 1 || len(storedTransactions(t, store)) != 1 {
		t.Fatalf("response status counts = %v", counts)
	}
}
