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
	// cfg.validate()

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
	// cfg.RunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	// cfg.StoreInterval = flag.Int("i", 2, "interval to save metrics")
	// cfg.FilePath = flag.String("f", "/tmp/metrics.json", "path to storage file")
	// cfg.Restore = flag.Bool("r", true, "restore metrics from file on startup")
	// cfg.DBConnStr = flag.String("d", "", "postgress connection string")
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

	// flagRunAddr = *addrTmp
	// flagStoreInterval = *storeIntervalTmp
	// flagFilePath = *filePathTmp
	// flagRestore = *restoreTmp
	// flagDBConnStr = *connStrTmp
}

// func (cfg *Config) validate() {

// }

// addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
// storeIntervalTmp := flag.Int("i", 2, "interval to save metrics")
// filePathTmp := flag.String("f", "/tmp/metrics.json", "path to storage file")
// restoreTmp := flag.Bool("r", true, "restore metrics from file on startup")
// connStrTmp := flag.String("d", "", "postgress connection string")
// flag.Parse()

// varAdrHost, ok := os.LookupEnv("ADDRESS")
// if ok {
// 	addrTmp = &varAdrHost
// }
// varStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
// if ok {
// 	intStoreInterval, err := strconv.Atoi(varStoreInterval)
// 	if err != nil {
// 		logger.Error("Filed to convert string to int, envInterval", err)
// 	}
// 	storeIntervalTmp = &intStoreInterval
// }
// varFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
// if ok {
// 	filePathTmp = &varFileStoragePath
// }
// varRestore, ok := os.LookupEnv("RESTORE")
// if ok {
// 	boolVarRestore, err := strconv.ParseBool(varRestore)
// 	if err != nil {
// 		logger.Error("Filed to convert string to bool, envRestore", err)
// 	}
// 	restoreTmp = &boolVarRestore
// }
// varDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
// if ok {
// 	connStrTmp = &varDatabaseDSN
// }

// flagRunAddr = *addrTmp
// flagStoreInterval = *storeIntervalTmp
// flagFilePath = *filePathTmp
// flagRestore = *restoreTmp
// flagDBConnStr = *connStrTmp

// log.Println("envAddr", flagRunAddr)
// log.Println("envInterval", flagStoreInterval)
// log.Println("envFilePath", flagFilePath)
// log.Println("envRestore", flagRestore)
// log.Println("envConnStr", flagDBConnStr)
