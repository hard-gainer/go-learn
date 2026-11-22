package report

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeSummaryRejectsInvalidDates(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		query string
		want  string
	}{
		{"missing date", "?from=2025-09-01", "from and to must be YYYY-MM-DD"},
		{"invalid date", "?from=2025-09-31&to=2025-09-30", "from and to must be YYYY-MM-DD"},
		{"reversed range", "?from=2025-09-30&to=2025-09-01", "report start date must not be after end date"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			New(nil, nil).ServeSummary(w, httptest.NewRequest(http.MethodGet, "/api/reports/summary"+tc.query, nil))
			var body map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &body)
			if w.Code != http.StatusBadRequest || err != nil || body["error"] != tc.want {
				t.Fatalf("status = %d, body = %s, decode error = %v", w.Code, w.Body, err)
			}
		})
	}
}
