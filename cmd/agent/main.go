package main

import (
	"time"

	"github.com/zetcan333/metrics-collector/internal/agent"
)

func main() {
	//Настройки агента
	serverURL := "http://localhost:8080"
	contentType := "text/plain"
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	//Инициализация и запуск
	agent := agent.NewAgent(serverURL, contentType, pollInterval, reportInterval)
	agent.Start()

	select {}
}
