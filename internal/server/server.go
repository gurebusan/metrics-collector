package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/gziprespose"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/mygzip"
	mwLogger "github.com/zetcan333/metrics-collector/internal/handlers/middleware/logger"
	"go.uber.org/zap"
)

type A interface {
	Some()
}

type Server struct {
	logger   *zap.Logger
	router   *chi.Mux
	handlers *handlers.ServerHandler
}

func NewServer(log *zap.Logger, handlers *handlers.ServerHandler) *Server {
	router := chi.NewRouter()

	router.Use(mwLogger.New(log))
	router.Use(mygzip.GzipMiddleware)
	router.Use(gziprespose.GzipResponseMiddleware)

	router.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", handlers.UpdateHandler)
			r.Post("/", handlers.UpdateJSONHandler)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", handlers.GetValueHandler)
			r.Post("/", handlers.GetJSONHandler)
		})
	})

	return &Server{logger: log, router: router, handlers: handlers}
}
