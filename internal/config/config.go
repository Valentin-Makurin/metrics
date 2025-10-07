package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type Config struct {
	logger        *zap.SugaredLogger
	RunAddr       string
	StoreInterval int
	FilePath      string
	Restore       bool
	DBConnStr     string
	KeyH          string
}

type ConfigAgent struct {
	HTTPAddr       string
	KeyH           string
	PollInterval   int
	ReportInterval int
	RateLimit      int
}

func ParseFlagsServer(logger *zap.SugaredLogger) Config {
	cfg := Config{}
	cfg.logger = logger

	cfg.parseCommandLineServer()
	cfg.parseEnvironmentServer()

	return cfg
}

func (cfg *Config) parseCommandLineServer() {

	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	storeIntervalTmp := flag.Int("i", 2, "interval to save metrics")
	filePathTmp := flag.String("f", "/tmp/metrics.json", "path to storage file")
	restoreTmp := flag.Bool("r", true, "restore metrics from file on startup")
	connStrTmp := flag.String("d", "", "postgress connection string")
	KeyHTmp := flag.String("k", "", "hash key")

	flag.Parse()

	if addrTmp != nil {
		cfg.RunAddr = *addrTmp
	}

	if storeIntervalTmp != nil {
		cfg.StoreInterval = *storeIntervalTmp
	}
	if filePathTmp != nil {
		cfg.FilePath = *filePathTmp
	}
	if restoreTmp != nil {
		cfg.Restore = *restoreTmp
	}
	if connStrTmp != nil {
		cfg.DBConnStr = *connStrTmp
	}
	if KeyHTmp != nil {
		cfg.KeyH = *KeyHTmp
	}
}

func (cfg *Config) parseEnvironmentServer() {
	varAdrHost, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.RunAddr = varAdrHost
	}

	varStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		intStoreInterval, err := strconv.Atoi(varStoreInterval)
		if err != nil {
			cfg.logger.Error("Filed to convert string to int, envInterval", err)
		}
		cfg.StoreInterval = intStoreInterval
	}

	varFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		cfg.FilePath = varFileStoragePath
	}

	varRestore, ok := os.LookupEnv("RESTORE")
	if ok {
		boolVarRestore, err := strconv.ParseBool(varRestore)
		if err != nil {
			cfg.logger.Error("Filed to convert string to bool, envRestore", err)
		}
		cfg.Restore = boolVarRestore
	}

	varDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		cfg.DBConnStr = varDatabaseDSN
	}

	varKeyH, ok := os.LookupEnv("KEY")
	if ok {
		cfg.KeyH = varKeyH
	}
}

func ParseFlagsAgent() ConfigAgent {
	cfg := ConfigAgent{}

	cfg.parseCommandLineAgent()
	cfg.parseEnvironmentAgent()

	return cfg
}

func (cfg *ConfigAgent) parseCommandLineAgent() {
	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	pollInterval := flag.Int("p", 2, "pollInterval")
	reportInterval := flag.Int("r", 10, "reportInterval")
	keyH := flag.String("k", "", "hash key")
	rateLim := flag.Int("l", 5, "rateLim")

	flag.Parse()

	if addrTmp != nil {
		cfg.HTTPAddr = *addrTmp
	}

	if pollInterval != nil {
		cfg.PollInterval = *pollInterval
	}

	if reportInterval != nil {
		cfg.ReportInterval = *reportInterval
	}

	if keyH != nil {
		cfg.KeyH = *keyH
	}

	if rateLim != nil {
		cfg.RateLimit = *rateLim
	}
}

func (cfg *ConfigAgent) parseEnvironmentAgent() {
	varAdrHost, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.HTTPAddr = varAdrHost
	}

	varPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		intPollInterval, err := strconv.Atoi(varPollInterval)
		if err != nil {
			log.Println("Filed to convert string to int, varPollInterval", err)
		}
		cfg.PollInterval = intPollInterval
	}

	varReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		intReportInterval, err := strconv.Atoi(varReportInterval)
		if err != nil {
			log.Println("Filed to convert string to int, varPollInterval", err)
		}
		cfg.ReportInterval = intReportInterval
	}

	varKeyH, ok := os.LookupEnv("KEY")
	if ok {
		cfg.KeyH = varKeyH
	}

	varRateLim, ok := os.LookupEnv("RATE_LIMIT")
	if ok {
		intRateLim, err := strconv.Atoi(varRateLim)
		if err != nil {
			log.Println("Filed to convert string to int, varRateLim", err)
		}
		cfg.RateLimit = intRateLim
	}
}
