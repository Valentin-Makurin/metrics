package config

import (
	"flag"
	"go.uber.org/zap"
	"os"
	"strconv"
)

type Config struct {
	logger        *zap.SugaredLogger
	RunAddr       string
	StoreInterval int
	FilePath      string
	Restore       bool
	DBConnStr     string
}

func ParseFlags(logger *zap.SugaredLogger) Config {
	cfg := Config{}
	cfg.logger = logger

	cfg.parseCommandLine()
	cfg.parseEnvironment()

	return cfg
}

func (cfg *Config) parseCommandLine() {

	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	storeIntervalTmp := flag.Int("i", 2, "interval to save metrics")
	filePathTmp := flag.String("f", "/tmp/metrics.json", "path to storage file")
	restoreTmp := flag.Bool("r", true, "restore metrics from file on startup")
	connStrTmp := flag.String("d", "", "postgress connection string")

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
}

func (cfg *Config) parseEnvironment() {
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
}
