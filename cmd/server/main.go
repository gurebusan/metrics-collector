package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
	"github.com/zetcan333/metrics-collector/internal/usercase"
)

func main() {

	storage := mem.NewStorage()
	serverUsecase := usercase.NewSeverUsecase(storage)
	h := handlers.NewServerHandler(serverUsecase)

	s := flags.NewServerFlags()

	r := h.NewRouter()
	fmt.Println("Server running on:", s.Address)
	log.Fatal(http.ListenAndServe(s.Address, r))
}
