package main

import (
	"github.com/zetcan333/metrics-collector/internal/agent"
	"github.com/zetcan333/metrics-collector/internal/flags"
)

func main() {
	//Настройки агента
	contentType := "text/plain"
	a := flags.NewAgentFlags()

	//Инициализация и запуск
	agent := agent.NewAgent(a.ServerURL, contentType, a.PollInterval, a.PollInterval)
	agent.Start()

	select {}
}
