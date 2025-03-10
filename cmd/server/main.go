package main

import (
	"net/http"

	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/storage/mem"
)

func main() {
	// Инициализация хранилища MemStorage
	storage := mem.NewStorage()

	//Регистрация хэндлеров
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateHandler(storage))

	//Запускаем сервер
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
