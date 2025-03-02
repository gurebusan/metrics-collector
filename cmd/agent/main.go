package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/zetcan333/metrics-collector/internal/agent"
)

var (
	serverURL      string
	contentType    string
	pollInterval   time.Duration
	reportInterval time.Duration
)

func main() {
	//Настройки агента
	parseFlags()
	contentType = "text/plain"

	//Инициализация и запуск
	agent := agent.NewAgent(serverURL, contentType, pollInterval, reportInterval)
	agent.Start()

	select {}
}

func parseFlags() {
	//Регистрируем флаги
	pflag.StringVarP(&serverURL, "a", "a", "http://localhost:8080", "Address and port for connection")
	pflag.DurationVarP(&pollInterval, "p", "p", 2*time.Second, "Set poll interval")
	pflag.DurationVarP(&reportInterval, "r", "r", 10*time.Second, "Set report interval")

	// Парсим флаги
	pflag.Parse()

	// Проверяем, есть ли неизвестные флаги
	flagSet := make(map[string]bool)
	pflag.VisitAll(func(f *pflag.Flag) {
		flagSet[f.Name] = true
	})

	pflag.Visit(func(f *pflag.Flag) {
		if !flagSet[f.Name] {
			fmt.Fprintf(os.Stderr, "Unknown flag: -%s\n", f.Name)
			os.Exit(1)
		}
	})
}
