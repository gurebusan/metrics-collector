package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/gziprespose"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/compressor/mygzip"
	mwLogger "github.com/zetcan333/metrics-collector/internal/handlers/middleware/logger"
	"github.com/zetcan333/metrics-collector/internal/handlers/middleware/signchecker"
	"github.com/zetcan333/metrics-collector/internal/handlers/ping"
	"github.com/zetcan333/metrics-collector/internal/usecase/backup"
	"go.uber.org/zap"
)

type Server struct {
	log    *zap.Logger
	router *chi.Mux
	flags  *flags.ServerFlags
	backup *backup.BackupUsecase
}

func NewServer(log *zap.Logger, handlers *handlers.ServerHandler, ping *ping.PingHandler, flags *flags.ServerFlags, backup *backup.BackupUsecase) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(mygzip.GzipMiddleware)
	router.Use(gziprespose.GzipResponseMiddleware)
	if flags.Key != "" {
		router.Use(signchecker.New(flags.Key))
	}

	router.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetrics)

		if ping != nil {
			r.Get("/ping", ping.Ping)
		}

		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", handlers.UpdateMetric)
			r.Post("/", handlers.UpdateViaModel)
		})

		r.Post("/updates/", handlers.UpdateMetricsWithBatch)

		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", handlers.GetMetric)
			r.Post("/", handlers.GetViaModel)
		})
	})

	return &Server{log: log, router: router, flags: flags, backup: backup}
}

func (s *Server) Start(ctx context.Context) {

	if s.backup != nil {
		if s.flags.Restore {
			if err := s.backup.LoadBackup(s.flags.FileStoragePath); err != nil {
				s.log.Sugar().Errorln("failed to load backup", zap.Error(err))
			} else {
				s.log.Sugar().Infoln("backup loaded")
			}
		}
	}

	server := &http.Server{
		Addr:    s.flags.Address,
		Handler: s.router,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		s.log.Sugar().Infoln("Starting server...")
		s.log.Sugar().Infoln("Server key", s.flags.Key)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.log.Sugar().Fatalln("failed to start server", zap.Error(err))
		}
	}()

	if s.backup != nil {
		ticker := time.NewTicker(s.flags.StoreInterval)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					if err := s.backup.SaveBackup(s.flags.FileStoragePath); err != nil {
						s.log.Sugar().Errorln("Failed to save backup", zap.Error(err))

					} else {
						s.log.Sugar().Infoln("Backup saved")
					}

				case <-ctx.Done():
					return
				}
			}
		}()
	}

	select {
	case <-ctx.Done():
	case <-stop:
	}

	s.log.Sugar().Infoln("Shutting down server...")

	if s.backup != nil {
		if err := s.backup.SaveBackup(s.flags.FileStoragePath); err != nil {
			s.log.Sugar().Errorln("Failed to save final backup", zap.Error(err))
		} else {
			s.log.Sugar().Infoln("Final backup saved")
		}
	}
	if err := server.Shutdown(context.Background()); err != nil {
		s.log.Sugar().Errorln("Failed to shutdown server", zap.Error(err))
	}
}
