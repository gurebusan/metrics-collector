package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/storage/mem"
)

func main() {
	// Инициализация хранилища MemStorage
	storage := mem.NewStorage()

	//Инициализация роутера, регистрация хэндлеров
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(storage))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(storage))

	//Запускаем сервер
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
