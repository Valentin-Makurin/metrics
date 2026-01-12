// Package config предоставляет функциональность для загрузки и парсинга конфигурации приложения.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Config содержит конфигурационные параметры сервера метрик.
// generate:reset
type Config struct {
	logger        *zap.SugaredLogger
	RunAddr       string
	FilePath      string
	DBConnStr     string
	KeyH          string
	AuditFilePath string
	AuditURL      string
	CryptoKey     string
	TrustedSubnet string
	StoreInterval int
	Restore       bool
}

// ConfigAgent содержит конфигурационные параметры агента сбора метрик.
// generate:reset
type ConfigAgent struct {
	HTTPAddr       string
	KeyH           string
	CryptoKey      string
	PollInterval   int
	ReportInterval int
	RateLimit      int
}

// TmplFileConfigAgent структура для параметров конфигурации из json файла агента
type TmplFileConfigAgent struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// TmplFileConfigServer структура для параметров конфигурации из json файла сервера
type TmplFileConfigServer struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// ParseFlagsServer парсит конфигурацию для сервера из командной строки и переменных окружения.
func ParseFlagsServer(logger *zap.SugaredLogger) Config {
	cfg := Config{}
	cfg.logger = logger

	cfg.parseConfigFileServer()
	cfg.parseCommandLineServer()
	cfg.parseEnvironmentServer()

	return cfg
}

// parseConfigFileServer парсит файл конфигурации строки для сервера.
func (cfg *Config) parseConfigFileServer() {
	configPath := ""
	tmplFileConfig := TmplFileConfigServer{}

	// Флаг для конфигурационного файла
	configPathFl := flag.String("c,config", "", "Path to config file")
	if configPathFl != nil {
		configPath = *configPathFl
	}

	// ENV переменная для конфигурационного файла
	if configPath == "" {
		varConfig, ok := os.LookupEnv("CONFIG")
		if ok {
			configPath = varConfig
		}
	}

	if configPath != "" {
		bytes, err := os.ReadFile(configPath)
		if err != nil {
			cfg.logger.Error("Filed to ReadFile in parseConfigFileAgent ", err)
			return
		}

		if err := json.Unmarshal(bytes, &tmplFileConfig); err != nil {
			cfg.logger.Error("Filed to Unmarshal in parseConfigFileAgent ", err)
			return
		}

		cfg.RunAddr = tmplFileConfig.Address
		cfg.Restore = tmplFileConfig.Restore
		cfg.FilePath = tmplFileConfig.StoreFile
		cfg.DBConnStr = tmplFileConfig.DatabaseDsn
		cfg.CryptoKey = tmplFileConfig.CryptoKey
		cfg.TrustedSubnet = tmplFileConfig.TrustedSubnet

		cfg.StoreInterval, err = prepareSeconds(tmplFileConfig.StoreInterval)
		if err != nil {
			cfg.logger.Error("Filed to prepareSeconds to StoreInterval", err)
		}

	}
}

// parseCommandLineServer парсит флаги командной строки для сервера.
func (cfg *Config) parseCommandLineServer() {
	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	storeIntervalTmp := flag.Int("i", 2, "interval to save metrics")
	filePathTmp := flag.String("f", "/tmp/metrics.json", "path to storage file")
	restoreTmp := flag.Bool("r", true, "restore metrics from file on startup")
	connStrTmp := flag.String("d", "", "postgress connection string")
	keyHTmp := flag.String("k", "", "hash key")
	auditFilePathTmp := flag.String("audit-file", "", "audir file path ")
	auditURLTmp := flag.String("audit-url", "", "audit url")
	crKeyTmp := flag.String("crypto-key", "private.pem", "crypto key")
	truSubTmp := flag.String("t", "", "trusted subnet")

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
	if keyHTmp != nil {
		cfg.KeyH = *keyHTmp
	}
	if auditFilePathTmp != nil {
		cfg.AuditFilePath = *auditFilePathTmp
	}
	if auditURLTmp != nil {
		cfg.AuditURL = *auditURLTmp
	}
	if crKeyTmp != nil {
		cfg.CryptoKey = *crKeyTmp
	}
	if truSubTmp != nil {
		cfg.TrustedSubnet = *truSubTmp
	}
}

// parseEnvironmentServer парсит переменные окружения для сервера.
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

	varAuditFilePath, ok := os.LookupEnv("AUDIT_FILE")
	if ok {
		cfg.AuditFilePath = varAuditFilePath
	}

	varAuditURL, ok := os.LookupEnv("AUDIT_URL")
	if ok {
		cfg.AuditURL = varAuditURL
	}

	varCrKey, ok := os.LookupEnv("CRYPTO_KEY")
	if ok {
		cfg.CryptoKey = varCrKey
	}

	varTruSub, ok := os.LookupEnv("TRUSTED_SUBNET")
	if ok {
		cfg.TrustedSubnet = varTruSub
	}
}

// ParseFlagsAgent парсит конфигурацию для агента из командной строки и переменных окружения.
func ParseFlagsAgent() ConfigAgent {
	cfg := ConfigAgent{}

	cfg.parseConfigFileAgent()
	cfg.parseCommandLineAgent()
	cfg.parseEnvironmentAgent()

	return cfg
}

// parseConfigFileAgent парсит файл конфигурации строки для агента.
func (cfg *ConfigAgent) parseConfigFileAgent() {
	configPath := ""

	tmplFileConfig := TmplFileConfigAgent{}

	// Флаг для конфигурационного файла
	configPathFl := flag.String("c,config", "", "Path to config file")
	if configPathFl != nil {
		configPath = *configPathFl
	}

	// ENV переменная для конфигурационного файла
	if configPath == "" {
		varConfig, ok := os.LookupEnv("CONFIG")
		if ok {
			configPath = varConfig
		}
	}

	if configPath != "" {
		bytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Println("Filed to ReadFile in parseConfigFileAgent ", err)
			return
		}

		if err := json.Unmarshal(bytes, &tmplFileConfig); err != nil {
			log.Println("Filed to Unmarshal in parseConfigFileAgent ", err)
			return
		}

		cfg.HTTPAddr = tmplFileConfig.Address
		cfg.CryptoKey = tmplFileConfig.CryptoKey

		cfg.ReportInterval, err = prepareSeconds(tmplFileConfig.ReportInterval)
		if err != nil {
			log.Println("Filed to prepareSeconds to ReportInterval", err)
		}

		cfg.PollInterval, err = prepareSeconds(tmplFileConfig.PollInterval)
		if err != nil {
			log.Println("Filed to prepareSeconds to PollInterval", err)
		}

	}
}

// parseCommandLineAgent парсит флаги командной строки для агента.
func (cfg *ConfigAgent) parseCommandLineAgent() {
	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
	pollInterval := flag.Int("p", 2, "pollInterval")
	reportInterval := flag.Int("r", 10, "reportInterval")
	keyH := flag.String("k", "", "hash key")
	rateLim := flag.Int("l", 5, "rateLim")
	crKey := flag.String("crypto-key", "public.pem", "crypto key")

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

	if crKey != nil {
		cfg.CryptoKey = *crKey
	}
}

// parseEnvironmentAgent парсит переменные окружения для агента.
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

	varCrKey, ok := os.LookupEnv("CRYPTO_KEY")
	if ok {
		cfg.CryptoKey = varCrKey
	}
}

func prepareSeconds(tm string) (int, error) {
	duration, err := time.ParseDuration(tm)
	if err != nil {
		return 0, err
	}
	return int(duration.Seconds()), nil
}
