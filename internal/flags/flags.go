package flags

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

type AgentFlags struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type ServerFlags struct {
	Address string
}

func NewAgentFlags() *AgentFlags {
	// Значения по умолчанию
	var (
		pollSec, reportSec int
		addr               string
	)

	// Читаем значения из переменных окружения, если они заданы
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	} else {
		addr = "localhost:8080"
	}

	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := strconv.Atoi(envPoll); err == nil {
			pollSec = val
		}
	} else {
		pollSec = 2
	}

	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := strconv.Atoi(envReport); err == nil {
			reportSec = val
		}
	} else {
		reportSec = 10
	}

	// Регистрируем флаги (они могут переопределить env)
	pflag.StringVarP(&addr, "a", "a", addr, "Address and port for connection")
	pflag.IntVarP(&pollSec, "p", "p", pollSec, "Set poll interval")
	pflag.IntVarP(&reportSec, "r", "r", reportSec, "Set report interval")

	// Парсим флаги
	pflag.Parse()

	// Преобразуем флаги в финальные значения
	serverURL := "http://" + addr
	pollInterval := time.Duration(pollSec) * time.Second
	reportInterval := time.Duration(reportSec) * time.Second

	return &AgentFlags{
		ServerURL:      serverURL,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
}

func NewServerFlags() *ServerFlags {
	// Значение по умолчанию
	addr := "localhost:8080"

	// Проверяем переменную окружения
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	// Добавляем поддержку флага (он имеет меньший приоритет, чем env)
	pflag.StringVarP(&addr, "a", "a", addr, "Address and port for the server")
	pflag.Parse()

	return &ServerFlags{
		Address: addr,
	}
}
