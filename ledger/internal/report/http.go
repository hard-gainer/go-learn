package report

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

func (s *Service) ServeSummary(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	from, fromErr := time.Parse(time.DateOnly, query.Get("from"))
	to, toErr := time.Parse(time.DateOnly, query.Get("to"))
	if fromErr != nil || toErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "from and to must be YYYY-MM-DD"})
		return
	}

	summary, cacheStatus, err := s.getSummary(r.Context(), from, to)
	switch {
	case errors.Is(err, errInvalidRange):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case err != nil:
		log.Printf("get report summary: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	default:
		w.Header().Set("X-Cache", cacheStatus)
		writeJSON(w, http.StatusOK, summary)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write report response: %v", err)
	}
}
