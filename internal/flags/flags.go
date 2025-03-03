package flags

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

type AgentFlags struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type ServerFlags struct {
	ServerURL string
}

func NewAgentFlags() *AgentFlags {
	//Регистрируем флаги
	var (
		pollSec, reportSec int
		addr               string
	)

	pflag.StringVarP(&addr, "a", "a", "localhost:8080", "Address and port for connection")
	pflag.IntVarP(&pollSec, "p", "p", 2, "Set poll interval")
	pflag.IntVarP(&reportSec, "r", "r", 10, "Set report interval")

	// Парсим флаги
	pflag.Parse()

	// Преобразуем флаги  в аргументы для агента
	serverURL := "http://" + addr
	pollInterval := time.Duration(pollSec) * time.Second
	reportInterval := time.Duration(reportSec) * time.Second

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
	return &AgentFlags{
		ServerURL:      serverURL,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
}

func NewServerFlags() *ServerFlags {
	var serverURL string
	pflag.StringVarP(&serverURL, "a", "a", "localhost:8080", "Address and port for connection")
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
	return &ServerFlags{ServerURL: serverURL}
}
