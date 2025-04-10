package main

import (
	"context"

	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/ping"
	strg "github.com/zetcan333/metrics-collector/internal/repo/storage"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/postgres"
	"github.com/zetcan333/metrics-collector/internal/server"
	"github.com/zetcan333/metrics-collector/internal/usecase"
	"github.com/zetcan333/metrics-collector/internal/usecase/backup"
	"go.uber.org/zap"
)

var (
	storage     strg.Storage
	pingHandler *ping.PingHandler
	bkp         *backup.BackupUsecase
)

func main() {

	log, err := zap.NewProduction()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer log.Sync()

	serverFlags := flags.NewServerFlags()
	ctx := context.WithoutCancel(context.Background())

	if serverFlags.DataBaseDSN != "" {
		storage, err = postgres.NewStorage(ctx, serverFlags.DataBaseDSN)
		if err != nil {
			fallInMememory(log)
		}
		if err := storage.InitTable(ctx); err != nil {
			fallInMememory(log)
		} else {
			log.Sugar().Infoln("postgres storage initialized")
			pingHandler = ping.New(storage)
		}
	} else {
		storage = mem.NewStorage()
		log.Sugar().Infoln("in-memory storage initialized")
		bkp = backup.NewBackupUsecase(storage)
	}

	serverUsecase := usecase.NewSeverUsecase(storage)
	handlers := handlers.NewServerHandler(log, serverUsecase)
	server := server.NewServer(log, handlers, pingHandler, serverFlags, bkp)

	server.Start(ctx)

	if s, ok := storage.(*postgres.PgStorage); ok {
		log.Sugar().Infoln("closing postgres pool")
		s.Close()
	}
}

func fallInMememory(log *zap.Logger) {
	log.Sugar().Infoln("falling back to in-memory storage")
	storage = mem.NewStorage()
	bkp = backup.NewBackupUsecase(storage)
}
