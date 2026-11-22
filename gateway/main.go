package main

import (
	"log"
	"net/http"
	"time"

	"gateway/internal/api"
	ledger "ledger/core"
)

// В сервисе Gateway напишите функцию main, которая запускает простой HTTP-сервер на порту
// (например, 8080). Реализуйте обработчик пути /ping, который при GET-запросе возвращает
// JSON-строку “pong” со статусом 200 OK. Это поможет проверить, что сервер запущен и отвечает.

func handlePing(w http.ResponseWriter, r *http.Request) {
	api.WriteJSON(w, http.StatusOK, "pong")
}

func main() {
	store, err := ledger.OpenStorage()
	if err != nil {
		log.Fatalf("Gateway database initialization failed: %v", err)
	}
	defer store.Close()

	if err := http.ListenAndServe(":8080", newHandler(store)); err != nil {
		log.Fatal(err)
	}
}

func newHandler(store *ledger.Storage) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)
	api.RegisterRoutes(mux, store)
	return logRequests(api.JSONResponses(mux))
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("method=%s path=%q duration=%s", r.Method, r.URL.Path, time.Since(start))
		}()

		next.ServeHTTP(w, r)
	})
}
