package main

import (
	"fmt"
	"net/http"
)

// В сервисе Gateway напишите функцию main, которая запускает простой HTTP-сервер на порту
// (например, 8080). Реализуйте обработчик пути /ping, который при GET-запросе возвращает
// строку “pong” со статусом 200 OK. Это поможет проверить, что сервер запущен и отвечает.

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {
	http.HandleFunc("GET /ping", handlePing)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
