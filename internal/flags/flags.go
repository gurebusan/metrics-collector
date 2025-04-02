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
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool

	DataBaseDSN string
}

const (
	defaultAddress         = "localhost:8080"
	defaultPollSec         = 2
	defaultReportSec       = 10
	defaultStoreSec        = 300
	defaultFileStoragePath = "/backup/metrcs_db"
	defaultRestore         = false
	defaultDataBaseDSN     = ""
)

func NewAgentFlags() *AgentFlags {

	addr := getEnvOrDefaultString("ADDRESS", defaultAddress)
	pollSec := getEnvOrDefaultInt("POLL_INTERVAL", defaultPollSec)
	reportSec := getEnvOrDefaultInt("REPORT_INTERVAL", defaultReportSec)

	pflag.StringVarP(&addr, "a", "a", addr, "Address and port for connection")
	pflag.IntVarP(&pollSec, "p", "p", pollSec, "Set poll interval")
	pflag.IntVarP(&reportSec, "r", "r", reportSec, "Set report interval")
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
	addr := getEnvOrDefaultString("ADDRESS", defaultAddress)
	storeSec := getEnvOrDefaultInt("STORE_INTERVAL", defaultStoreSec)
	fileStoragePath := getEnvOrDefaultString("FILE_STORAGE_PATH", defaultFileStoragePath)
	restore := getEnvOrDefaultBool("RESTORE", defaultRestore)

	dbDSN := getEnvOrDefaultString("DATABASE_DSN", defaultDataBaseDSN)

	pflag.StringVarP(&addr, "a", "a", addr, "Address and port for the server")
	pflag.IntVarP(&storeSec, "i", "i", storeSec, "Store interval for backup")
	pflag.StringVarP(&fileStoragePath, "f", "f", fileStoragePath, "File storage path for backup")
	pflag.BoolVarP(&restore, "r", "r", restore, "Use for load db fron file")
	pflag.StringVarP(&dbDSN, "d", "d", dbDSN, "Connect postgres via DSN")
	pflag.Parse()

	storeInterval := time.Duration(storeSec) * time.Second

	return &ServerFlags{
		Address:         addr,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
		DataBaseDSN:     dbDSN,
	}
}

func getEnvOrDefaultString(envVar string, defaultValue string) string {
	if value, ok := os.LookupEnv(envVar); ok {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultBool(envVar string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(envVar); ok {
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			return parsedValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultInt(envVar string, defaultValue int) int {
	if value, ok := os.LookupEnv(envVar); ok {
		if parsedValue, err := strconv.Atoi(value); err == nil {
			return parsedValue
		}
	}
	return defaultValue
}
