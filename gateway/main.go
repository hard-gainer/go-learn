package main

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"gateway/internal/api"
	ledger "ledger/core"
)

// В сервисе Gateway напишите функцию main, которая запускает простой HTTP-сервер на порту
// (например, 8080). Реализуйте обработчик пути /ping, который при GET-запросе возвращает
// строку “pong” со статусом 200 OK. Это поможет проверить, что сервер запущен и отвечает.

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {
	budgets, err := os.ReadFile("../ledger/budgets.json")
	if err != nil {
		log.Fatal(err)
	}
	store := ledger.NewStorage()
	if err := store.LoadBudgets(bytes.NewReader(budgets)); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)
	api.RegisterRoutes(mux, store)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
