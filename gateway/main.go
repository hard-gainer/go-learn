package main

import (
	"fmt"
	"net/http"

	"gateway/internal/api"
)

// В сервисе Gateway напишите функцию main, которая запускает простой HTTP-сервер на порту
// (например, 8080). Реализуйте обработчик пути /ping, который при GET-запросе возвращает
// строку “pong” со статусом 200 OK. Это поможет проверить, что сервер запущен и отвечает.

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)
	api.RegisterRoutes(mux)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}
}
