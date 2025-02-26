package main

import (
	"net/http"

	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/storage/mem"
)

func main() {
	// Инициализация хранилища MemStorage
	storage := mem.NewStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", handlers.UpdateGaugeHandler(storage))
	mux.HandleFunc("/update/counter/", handlers.UpdateCounterHandler(storage))

	//Запускаем сервер
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
