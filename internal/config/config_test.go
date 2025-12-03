package config

import (
	"flag"
	"os"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestParseFlagsServer(t *testing.T) {
	// Сохраняем оригинальные значения
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("DefaultValues", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()

		// Устанавливаем тестовые аргументы (только имя программы)
		os.Args = []string{"test-program"}

		// Сбрасываем флаги для нового теста
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)

		cfg := ParseFlagsServer(logger)

		if cfg.RunAddr != "localhost:8080" {
			t.Errorf("Expected RunAddr localhost:8080, got %s", cfg.RunAddr)
		}
		if cfg.StoreInterval != 2 {
			t.Errorf("Expected StoreInterval 2, got %d", cfg.StoreInterval)
		}
		if cfg.FilePath != "/tmp/metrics.json" {
			t.Errorf("Expected FilePath /tmp/metrics.json, got %s", cfg.FilePath)
		}
		if cfg.Restore != true {
			t.Errorf("Expected Restore true, got %v", cfg.Restore)
		}
		if cfg.DBConnStr != "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable" {
			t.Errorf("Expected default DBConnStr, got %s", cfg.DBConnStr)
		}
		if cfg.KeyH != "" {
			t.Errorf("Expected empty KeyH, got %s", cfg.KeyH)
		}
		if cfg.AuditFilePath != "" {
			t.Errorf("Expected empty AuditFilePath, got %s", cfg.AuditFilePath)
		}
		if cfg.AuditURL != "" {
			t.Errorf("Expected empty AuditURL, got %s", cfg.AuditURL)
		}
		if cfg.logger != logger {
			t.Error("Logger should be set")
		}
	})

	t.Run("CommandLineFlags", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()

		// Устанавливаем тестовые аргументы командной строки
		os.Args = []string{
			"test-program",
			"-a", "127.0.0.1:9090",
			"-i", "30",
			"-f", "/var/log/metrics.json",
			"-r=false",
			"-d", "postgres://user:pass@localhost:5432/test",
			"-k", "secret-key",
			"-audit-file", "/tmp/audit.log",
			"-audit-url", "http://audit.server/log",
		}

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)

		cfg := ParseFlagsServer(logger)

		if cfg.RunAddr != "127.0.0.1:9090" {
			t.Errorf("Expected RunAddr 127.0.0.1:9090, got %s", cfg.RunAddr)
		}
		if cfg.StoreInterval != 30 {
			t.Errorf("Expected StoreInterval 30, got %d", cfg.StoreInterval)
		}
		if cfg.FilePath != "/var/log/metrics.json" {
			t.Errorf("Expected FilePath /var/log/metrics.json, got %s", cfg.FilePath)
		}
		if cfg.Restore != false {
			t.Errorf("Expected Restore false, got %v", cfg.Restore)
		}
		if cfg.DBConnStr != "postgres://user:pass@localhost:5432/test" {
			t.Errorf("Expected DBConnStr postgres://user:pass@localhost:5432/test, got %s", cfg.DBConnStr)
		}
		if cfg.KeyH != "secret-key" {
			t.Errorf("Expected KeyH secret-key, got %s", cfg.KeyH)
		}
		if cfg.AuditFilePath != "/tmp/audit.log" {
			t.Errorf("Expected AuditFilePath /tmp/audit.log, got %s", cfg.AuditFilePath)
		}
		if cfg.AuditURL != "http://audit.server/log" {
			t.Errorf("Expected AuditURL http://audit.server/log, got %s", cfg.AuditURL)
		}
	})
}

func TestParseFlagsServer_Environment(t *testing.T) {
	// Тесты с переменными окружения
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("EnvironmentVariables", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()

		// Устанавливаем переменные окружения
		os.Setenv("ADDRESS", "0.0.0.0:8080")
		os.Setenv("STORE_INTERVAL", "60")
		os.Setenv("FILE_STORAGE_PATH", "/data/metrics.json")
		os.Setenv("RESTORE", "false")
		os.Setenv("DATABASE_DSN", "postgres://env:env@localhost:5432/envdb")
		os.Setenv("KEY", "env-secret")
		os.Setenv("AUDIT_FILE", "/var/audit.log")
		os.Setenv("AUDIT_URL", "http://env.audit/log")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("STORE_INTERVAL")
			os.Unsetenv("FILE_STORAGE_PATH")
			os.Unsetenv("RESTORE")
			os.Unsetenv("DATABASE_DSN")
			os.Unsetenv("KEY")
			os.Unsetenv("AUDIT_FILE")
			os.Unsetenv("AUDIT_URL")
		}()

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg := ParseFlagsServer(logger)

		if cfg.RunAddr != "0.0.0.0:8080" {
			t.Errorf("Expected RunAddr 0.0.0.0:8080, got %s", cfg.RunAddr)
		}
		if cfg.StoreInterval != 60 {
			t.Errorf("Expected StoreInterval 60, got %d", cfg.StoreInterval)
		}
		if cfg.FilePath != "/data/metrics.json" {
			t.Errorf("Expected FilePath /data/metrics.json, got %s", cfg.FilePath)
		}
		if cfg.Restore != false {
			t.Errorf("Expected Restore false, got %v", cfg.Restore)
		}
		if cfg.DBConnStr != "postgres://env:env@localhost:5432/envdb" {
			t.Errorf("Expected DBConnStr postgres://env:env@localhost:5432/envdb, got %s", cfg.DBConnStr)
		}
		if cfg.KeyH != "env-secret" {
			t.Errorf("Expected KeyH env-secret, got %s", cfg.KeyH)
		}
		if cfg.AuditFilePath != "/var/audit.log" {
			t.Errorf("Expected AuditFilePath /var/audit.log, got %s", cfg.AuditFilePath)
		}
		if cfg.AuditURL != "http://env.audit/log" {
			t.Errorf("Expected AuditURL http://env.audit/log, got %s", cfg.AuditURL)
		}
	})

	t.Run("EnvironmentOverridesCommandLine", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()

		// Командная строка
		os.Args = []string{
			"test-program",
			"-a", "cmdline:8080",
			"-i", "10",
			"-k", "cmd-secret",
		}

		// Окружение (должно переопределить командную строку)
		os.Setenv("ADDRESS", "env:9090")
		os.Setenv("STORE_INTERVAL", "20")
		os.Setenv("KEY", "env-secret")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("STORE_INTERVAL")
			os.Unsetenv("KEY")
		}()

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)

		cfg := ParseFlagsServer(logger)

		// Окружение должно победить
		if cfg.RunAddr != "env:9090" {
			t.Errorf("Expected RunAddr env:9090 (from env), got %s", cfg.RunAddr)
		}
		if cfg.StoreInterval != 20 {
			t.Errorf("Expected StoreInterval 20 (from env), got %d", cfg.StoreInterval)
		}
		if cfg.KeyH != "env-secret" {
			t.Errorf("Expected KeyH env-secret (from env), got %s", cfg.KeyH)
		}
	})
}

func TestConfig_parseCommandLineServer(t *testing.T) {
	t.Run("NilFlagValues", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()
		cfg := &Config{logger: logger}

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg.parseCommandLineServer()

		// Проверяем значения по умолчанию
		if cfg.RunAddr != "localhost:8080" {
			t.Errorf("Expected default RunAddr, got %s", cfg.RunAddr)
		}
	})

	t.Run("EmptyCommandLine", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()
		cfg := &Config{logger: logger}

		// Пустая командная строка
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg.parseCommandLineServer()

		// Должны быть значения по умолчанию
		if cfg.RunAddr != "localhost:8080" {
			t.Errorf("Expected default RunAddr, got %s", cfg.RunAddr)
		}
	})
}

func TestConfig_parseEnvironmentServer(t *testing.T) {
	t.Run("AllEnvironmentVariables", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()
		cfg := &Config{logger: logger}

		// Устанавливаем все переменные
		os.Setenv("ADDRESS", "test:8080")
		os.Setenv("STORE_INTERVAL", "99")
		os.Setenv("FILE_STORAGE_PATH", "/test/path.json")
		os.Setenv("RESTORE", "false")
		os.Setenv("DATABASE_DSN", "test://dsn")
		os.Setenv("KEY", "test-key")
		os.Setenv("AUDIT_FILE", "/test/audit.log")
		os.Setenv("AUDIT_URL", "http://test/audit")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("STORE_INTERVAL")
			os.Unsetenv("FILE_STORAGE_PATH")
			os.Unsetenv("RESTORE")
			os.Unsetenv("DATABASE_DSN")
			os.Unsetenv("KEY")
			os.Unsetenv("AUDIT_FILE")
			os.Unsetenv("AUDIT_URL")
		}()

		// Устанавливаем начальные значения
		cfg.RunAddr = "initial:8080"
		cfg.StoreInterval = 1
		cfg.FilePath = "/initial/path.json"
		cfg.Restore = true
		cfg.DBConnStr = "initial://dsn"
		cfg.KeyH = "initial-key"
		cfg.AuditFilePath = "/initial/audit.log"
		cfg.AuditURL = "http://initial/audit"

		cfg.parseEnvironmentServer()

		// Проверяем, что значения перезаписаны
		if cfg.RunAddr != "test:8080" {
			t.Errorf("Expected RunAddr test:8080, got %s", cfg.RunAddr)
		}
		if cfg.StoreInterval != 99 {
			t.Errorf("Expected StoreInterval 99, got %d", cfg.StoreInterval)
		}
		if cfg.FilePath != "/test/path.json" {
			t.Errorf("Expected FilePath /test/path.json, got %s", cfg.FilePath)
		}
		if cfg.Restore != false {
			t.Errorf("Expected Restore false, got %v", cfg.Restore)
		}
		if cfg.DBConnStr != "test://dsn" {
			t.Errorf("Expected DBConnStr test://dsn, got %s", cfg.DBConnStr)
		}
		if cfg.KeyH != "test-key" {
			t.Errorf("Expected KeyH test-key, got %s", cfg.KeyH)
		}
		if cfg.AuditFilePath != "/test/audit.log" {
			t.Errorf("Expected AuditFilePath /test/audit.log, got %s", cfg.AuditFilePath)
		}
		if cfg.AuditURL != "http://test/audit" {
			t.Errorf("Expected AuditURL http://test/audit, got %s", cfg.AuditURL)
		}
	})

	t.Run("InvalidInteger", func(t *testing.T) {
		logger := zaptest.NewLogger(t).Sugar()
		cfg := &Config{logger: logger}

		os.Setenv("STORE_INTERVAL", "not-a-number")
		defer os.Unsetenv("STORE_INTERVAL")

		cfg.StoreInterval = 50 // Начальное значение
		cfg.parseEnvironmentServer()

		if cfg.StoreInterval != 50 {
			t.Logf("StoreInterval changed from 50 to %d on parse error", cfg.StoreInterval)
		}
	})
}

func TestParseFlagsAgent(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("DefaultValues", func(t *testing.T) {
		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg := ParseFlagsAgent()

		if cfg.HTTPAddr != "localhost:8080" {
			t.Errorf("Expected HTTPAddr localhost:8080, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 2 {
			t.Errorf("Expected PollInterval 2, got %d", cfg.PollInterval)
		}
		if cfg.ReportInterval != 10 {
			t.Errorf("Expected ReportInterval 10, got %d", cfg.ReportInterval)
		}
		if cfg.KeyH != "" {
			t.Errorf("Expected empty KeyH, got %s", cfg.KeyH)
		}
		if cfg.RateLimit != 5 {
			t.Errorf("Expected RateLimit 5, got %d", cfg.RateLimit)
		}
	})

	t.Run("CommandLineFlags", func(t *testing.T) {
		os.Args = []string{
			"test-program",
			"-a", "server:9090",
			"-p", "5",
			"-r", "30",
			"-k", "agent-secret",
			"-l", "10",
		}

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)

		cfg := ParseFlagsAgent()

		if cfg.HTTPAddr != "server:9090" {
			t.Errorf("Expected HTTPAddr server:9090, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 5 {
			t.Errorf("Expected PollInterval 5, got %d", cfg.PollInterval)
		}
		if cfg.ReportInterval != 30 {
			t.Errorf("Expected ReportInterval 30, got %d", cfg.ReportInterval)
		}
		if cfg.KeyH != "agent-secret" {
			t.Errorf("Expected KeyH agent-secret, got %s", cfg.KeyH)
		}
		if cfg.RateLimit != 10 {
			t.Errorf("Expected RateLimit 10, got %d", cfg.RateLimit)
		}
	})
}

func TestParseFlagsAgent_Environment(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("EnvironmentVariables", func(t *testing.T) {
		os.Setenv("ADDRESS", "env-server:8080")
		os.Setenv("POLL_INTERVAL", "3")
		os.Setenv("REPORT_INTERVAL", "15")
		os.Setenv("KEY", "env-agent-key")
		os.Setenv("RATE_LIMIT", "20")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("REPORT_INTERVAL")
			os.Unsetenv("KEY")
			os.Unsetenv("RATE_LIMIT")
		}()

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg := ParseFlagsAgent()

		if cfg.HTTPAddr != "env-server:8080" {
			t.Errorf("Expected HTTPAddr env-server:8080, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 3 {
			t.Errorf("Expected PollInterval 3, got %d", cfg.PollInterval)
		}
		if cfg.ReportInterval != 15 {
			t.Errorf("Expected ReportInterval 15, got %d", cfg.ReportInterval)
		}
		if cfg.KeyH != "env-agent-key" {
			t.Errorf("Expected KeyH env-agent-key, got %s", cfg.KeyH)
		}
		if cfg.RateLimit != 20 {
			t.Errorf("Expected RateLimit 20, got %d", cfg.RateLimit)
		}
	})

	t.Run("EnvironmentOverridesCommandLine", func(t *testing.T) {
		os.Args = []string{
			"test-program",
			"-a", "cmd-server:8080",
			"-p", "1",
			"-k", "cmd-key",
		}

		os.Setenv("ADDRESS", "env-override:9090")
		os.Setenv("POLL_INTERVAL", "2")
		os.Setenv("KEY", "env-override-key")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("KEY")
		}()

		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)

		cfg := ParseFlagsAgent()

		// Окружение должно победить
		if cfg.HTTPAddr != "env-override:9090" {
			t.Errorf("Expected HTTPAddr env-override:9090, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 2 {
			t.Errorf("Expected PollInterval 2, got %d", cfg.PollInterval)
		}
		if cfg.KeyH != "env-override-key" {
			t.Errorf("Expected KeyH env-override-key, got %s", cfg.KeyH)
		}
	})
}

func TestConfigAgent_parseCommandLineAgent(t *testing.T) {
	t.Run("EmptyCommandLine", func(t *testing.T) {
		cfg := &ConfigAgent{}

		// Пустая командная строка
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		os.Args = []string{"test-program"}

		cfg.parseCommandLineAgent()

		// Проверяем значения по умолчанию
		if cfg.HTTPAddr != "localhost:8080" {
			t.Errorf("Expected default HTTPAddr, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 2 {
			t.Errorf("Expected default PollInterval, got %d", cfg.PollInterval)
		}
	})
}

func TestConfigAgent_parseEnvironmentAgent(t *testing.T) {
	t.Run("AllEnvironmentVariables", func(t *testing.T) {
		cfg := &ConfigAgent{}

		// Устанавливаем все переменные
		os.Setenv("ADDRESS", "env-test:8080")
		os.Setenv("POLL_INTERVAL", "7")
		os.Setenv("REPORT_INTERVAL", "25")
		os.Setenv("KEY", "env-test-key")
		os.Setenv("RATE_LIMIT", "15")
		defer func() {
			os.Unsetenv("ADDRESS")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("REPORT_INTERVAL")
			os.Unsetenv("KEY")
			os.Unsetenv("RATE_LIMIT")
		}()

		// Начальные значения
		cfg.HTTPAddr = "initial:8080"
		cfg.PollInterval = 1
		cfg.ReportInterval = 10
		cfg.KeyH = "initial-key"
		cfg.RateLimit = 5

		cfg.parseEnvironmentAgent()

		// Проверяем перезапись
		if cfg.HTTPAddr != "env-test:8080" {
			t.Errorf("Expected HTTPAddr env-test:8080, got %s", cfg.HTTPAddr)
		}
		if cfg.PollInterval != 7 {
			t.Errorf("Expected PollInterval 7, got %d", cfg.PollInterval)
		}
		if cfg.ReportInterval != 25 {
			t.Errorf("Expected ReportInterval 25, got %d", cfg.ReportInterval)
		}
		if cfg.KeyH != "env-test-key" {
			t.Errorf("Expected KeyH env-test-key, got %s", cfg.KeyH)
		}
		if cfg.RateLimit != 15 {
			t.Errorf("Expected RateLimit 15, got %d", cfg.RateLimit)
		}
	})
}
