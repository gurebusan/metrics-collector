package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/storage/mem"
)

func main() {
	// Инициализация хранилища MemStorage
	storage := mem.NewStorage()

	//Инициализация флагов
	s := flags.NewServerFlags()

	//Инициализация роутера, регистрация хэндлеров
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(storage))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(storage))
	r.Get("/", handlers.GetAllMetricsHandler(storage))

	//Запуск сервера с флагом
	if err := http.ListenAndServe(s.ServerUrl, r); err != nil {
		panic(err)
	}
}
