package main

import (
	"context"

	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/ping"
	"github.com/zetcan333/metrics-collector/internal/repo/storage"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/postgres"
	"github.com/zetcan333/metrics-collector/internal/server"
	"github.com/zetcan333/metrics-collector/internal/usecase"
	"github.com/zetcan333/metrics-collector/internal/usecase/backup"
	"go.uber.org/zap"
)

func main() {

	log, err := zap.NewProduction()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer log.Sync()

	serverFlags := flags.NewServerFlags()
	ctx := context.WithoutCancel(context.Background())

	var (
		storage     storage.Storage
		pingHandler *ping.PingHandler
		bkp         *backup.BackupUsecase
	)
	if serverFlags.DataBaseDSN != "" {
		storage, err = postgres.NewStorage(ctx, serverFlags.DataBaseDSN)
		if err != nil {
			log.Sugar().Errorln("cannot initialize postgres, falling back to in-memory storage:", zap.Error(err))
			storage = mem.NewStorage()
			bkp = backup.NewBackupUsecase(storage)
		}
		pingHandler = ping.New(storage)
	} else {
		storage = mem.NewStorage()
		bkp = backup.NewBackupUsecase(storage)
	}

	serverUsecase := usecase.NewSeverUsecase(storage)
	handlers := handlers.NewServerHandler(serverUsecase)
	server := server.NewServer(log, handlers, pingHandler, serverFlags, bkp)

	server.Start(ctx)
}
