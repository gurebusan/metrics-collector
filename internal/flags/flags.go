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
	Key            string
}

type ServerFlags struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DataBaseDSN     string
	Key             string
}

const (
	defaultAddress         = "localhost:8080"
	defaultPollSec         = 2
	defaultReportSec       = 10
	defaultStoreSec        = 300
	defaultFileStoragePath = "/backup/metrcs_db"
	defaultRestore         = false
	defaultDataBaseDSN     = ""
	defaultKey             = ""
)

func NewAgentFlags() *AgentFlags {

	addrPtr := pflag.StringP("a", "a", getEnvOrDefaultString("ADDRESS", defaultAddress), "Address and port for connection")
	pollSecPtr := pflag.IntP("p", "p", getEnvOrDefaultInt("POLL_INTERVAL", defaultPollSec), "Set poll interval")
	reportSecPtr := pflag.IntP("r", "r", getEnvOrDefaultInt("REPORT_INTERVAL", defaultReportSec), "Set report interval")
	keyPtr := pflag.StringP("k", "k", getEnvOrDefaultString("KEY", defaultKey), "Set key")

	pflag.Parse() // Парсим все флаги разом

	// Преобразуем в финальные значения
	return &AgentFlags{
		ServerURL:      "http://" + *addrPtr,
		PollInterval:   time.Duration(*pollSecPtr) * time.Second,
		ReportInterval: time.Duration(*reportSecPtr) * time.Second,
		Key:            *keyPtr,
	}
}

func NewServerFlags() *ServerFlags {
	// Значение по умолчанию
	addrPtr := pflag.StringP("a", "a", getEnvOrDefaultString("ADDRESS", defaultAddress), "Address and port for the server")
	storeSecPtr := pflag.IntP("i", "i", getEnvOrDefaultInt("STORE_INTERVAL", defaultStoreSec), "Store interval for backup")
	filePathPtr := pflag.StringP("f", "f", getEnvOrDefaultString("FILE_STORAGE_PATH", defaultFileStoragePath), "File storage path for backup")
	restorePtr := pflag.BoolP("r", "r", getEnvOrDefaultBool("RESTORE", defaultRestore), "Use for load db from file")
	dbDSNPtr := pflag.StringP("d", "d", getEnvOrDefaultString("DATABASE_DSN", defaultDataBaseDSN), "Connect postgres via DSN")
	keyPtr := pflag.StringP("k", "k", getEnvOrDefaultString("KEY", defaultKey), "Set key")

	pflag.Parse()

	return &ServerFlags{
		Address:         *addrPtr,
		StoreInterval:   time.Duration(*storeSecPtr) * time.Second,
		FileStoragePath: *filePathPtr,
		Restore:         *restorePtr,
		DataBaseDSN:     *dbDSNPtr,
		Key:             *keyPtr,
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
