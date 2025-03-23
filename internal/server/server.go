package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/gziprespose"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/mygzip"
	mwLogger "github.com/zetcan333/metrics-collector/internal/handlers/middleware/logger"
	"github.com/zetcan333/metrics-collector/internal/usecase/backup"
	"go.uber.org/zap"
)

type A interface {
	Some()
}

type Server struct {
	log    *zap.Logger
	router *chi.Mux
	flags  *flags.ServerFlags
	backup *backup.BackupUsecase
}

func NewServer(log *zap.Logger, handlers *handlers.ServerHandler, flags *flags.ServerFlags, backup *backup.BackupUsecase) *Server {
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

	return &Server{log: log, router: router, flags: flags, backup: backup}
}

func (s *Server) Start(ctx context.Context) {
	if s.flags.Restore {
		if err := s.backup.LoadBackup(s.flags.FileStoragePath); err != nil {
			s.log.Sugar().Errorln("falied to load backup", zap.Error(err))
		}

		go func() {
			fmt.Println("Starting server...")
			err := http.ListenAndServe(s.flags.Address, s.router)
			if err != nil {
				s.log.Sugar().Fatalln("failed to start server", zap.Error(err))
			}
		}()

	}
}
