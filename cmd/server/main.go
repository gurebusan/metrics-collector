package main

import (
	"context"

	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/ping"
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

	pgstorage, err := postgres.New(ctx, serverFlags.DataBaseDSN)
	if err != nil {
		log.Sugar().Errorln(err)
	}
	ping := ping.New(pgstorage)

	storage := mem.NewStorage()
	serverUsecase := usecase.NewSeverUsecase(storage)
	backup := backup.NewBackupUsecase(storage)
	handlers := handlers.NewServerHandler(serverUsecase)

	server := server.NewServer(log, handlers, ping, serverFlags, backup)

	server.Start(ctx)
}
