package main

import (
	"fmt"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/gziprespose"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/mygzip"
	mwLogger "github.com/zetcan333/metrics-collector/internal/handlers/middleware/logger"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
	"github.com/zetcan333/metrics-collector/internal/setup"
	"github.com/zetcan333/metrics-collector/internal/usercase"
	"go.uber.org/zap"
)

func main() {
	mylog, err := zap.NewProduction()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer mylog.Sync()
	storage := mem.NewStorage()
	serverUsecase := usercase.NewSeverUsecase(storage)
	h := handlers.NewServerHandler(serverUsecase)

	s := flags.NewServerFlags()

	r := chi.NewRouter()
	r.Use(mwLogger.New(mylog))
	r.Use(mygzip.GzipMiddleware)
	r.Use(gziprespose.GzipResponseMiddleware)

	setup := setup.NewSetup(h)
	setup.SetRoutes(r)

	fmt.Println("Server running on:", s.Address)
	if err := http.ListenAndServe(s.Address, r); err != nil {
		mylog.Error("failed to start server")
	}
}
