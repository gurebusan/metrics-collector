package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/pflag"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/storage/mem"
)

var addr string

func main() {
	// Инициализация хранилища MemStorage
	storage := mem.NewStorage()

	//Инициализация роутера, регистрация хэндлеров
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(storage))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(storage))
	r.Get("/", handlers.GetAllMetricsHandler(storage))

	//Запуск сервера с флагом
	parseFlags()
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}

func parseFlags() {
	pflag.StringVarP(&addr, "a", "a", "localhost:8080", "Address and port for connection")
	pflag.Parse()

	// Проверяем, есть ли неизвестные флаги
	flagSet := make(map[string]bool)
	pflag.VisitAll(func(f *pflag.Flag) {
		flagSet[f.Name] = true
	})

	pflag.Visit(func(f *pflag.Flag) {
		if !flagSet[f.Name] {
			fmt.Fprintf(os.Stderr, "Unknown flag: -%s\n", f.Name)
			os.Exit(1)
		}
	})
}
