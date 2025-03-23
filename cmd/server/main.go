package main

import (
	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/repo/storage/mem"
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
	storage := mem.NewStorage()
	serverUsecase := usecase.NewSeverUsecase(storage)
	backup := backup.NewBackupUsecase(storage)
	handlers := handlers.NewServerHandler(serverUsecase)

	serverFlags := flags.NewServerFlags()

	r := server.NewServer(log, handlers, serverFlags, backup)

	/*fmt.Println("Server running on:", s.Address)
	if err := http.ListenAndServe(s.Address, r); err != nil {
		log.Error("failed to start server")
	}*/
}
