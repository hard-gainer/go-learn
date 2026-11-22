package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ledger "ledger/core"
)

func TestServerJSONResponses(t *testing.T) {
	handler := newHandler(ledger.NewStorage())
	for _, tc := range []struct {
		method string
		path   string
		status int
		body   string
	}{
		{http.MethodGet, "/ping", http.StatusOK, `"pong"`},
		{http.MethodGet, "/missing", http.StatusNotFound, `{"error":"Not Found"}`},
		{http.MethodPost, "/ping", http.StatusMethodNotAllowed, `{"error":"Method Not Allowed"}`},
		{http.MethodPut, "/api/budgets/food", http.StatusNotImplemented, `{"error":"Not Implemented"}`},
		{http.MethodGet, "//ping", http.StatusTemporaryRedirect, `{"location":"/ping"}`},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || strings.TrimSpace(w.Body.String()) != tc.body {
				t.Fatalf("status = %d, headers = %v, body = %s", w.Code, w.Header(), w.Body)
			}
			if tc.status == http.StatusMethodNotAllowed && w.Header().Get("Allow") != "GET, HEAD" {
				t.Fatalf("Allow = %q", w.Header().Get("Allow"))
			}
			if tc.status == http.StatusTemporaryRedirect && w.Header().Get("Location") != "/ping" {
				t.Fatalf("Location = %q", w.Header().Get("Location"))
			}
		})
	}
}
