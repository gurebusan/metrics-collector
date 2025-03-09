package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
	"github.com/zetcan333/metrics-collector/internal/usercase"
)

func main() {
	//
	storage := mem.NewStorage()
	serverUsecase := usercase.NewSeverUsecase(storage)
	h := handlers.NewServerHandler(serverUsecase)

	s := flags.NewServerFlags()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetAllMetricsHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", h.UpdateHandle)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}/", h.GetValueHandler)
		})
	})
	fmt.Println("Server running on:", s.Address)
	log.Fatal(http.ListenAndServe(s.Address, r))
}
