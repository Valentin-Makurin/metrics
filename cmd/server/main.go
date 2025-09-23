package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/Valentin-Makurin/metrics/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	flagRunAddr       string
	flagStoreInterval int
	flagFilePath      string
	flagRestore       bool
	flagDBConnStr     string
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	initFlags(sugar)

	mtrHandler := &handler.MtrHandler{}

	if flagDBConnStr != "" {
		dbConn, err := db.NewDatabase(flagDBConnStr, sugar)
		if err != nil {
			log.Fatalf("Ошибка подключения к БД: %v", err)
		}
		defer dbConn.Close()
		err = dbConn.Ping()
		if err != nil {
			log.Fatalf("Ошибка пинга к БД: %v", err)
		}
		dbConn.RunMigrations()
		mtrHandler = handler.NewMtrHandler(dbConn, sugar)
	} else {
		storage := db.NewStorage(flagFilePath, flagStoreInterval, flagRestore, sugar)
		storage.PrepareFile()
		storage.StartTicker()
		mtrHandler = handler.NewMtrHandler(storage, sugar)
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(sugar))
	r.Use(middleware.GzipMiddleware)
	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)

	r.Get("/ping", mtrHandler.HandlePing)
	r.Post("/update/", mtrHandler.HandlePostUpdate)
	r.Post("/value/", mtrHandler.HandleGetValue)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Running server on", flagRunAddr)
	err = http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		log.Println("Filed to start server", err)
	}
}

func initFlags(logger *zap.SugaredLogger) {
	defaultFilePath := "/tmp/metrics.json"

	envAddr := os.Getenv("ADDRESS")
	envInterval := os.Getenv("STORE_INTERVAL")
	envFilePath := os.Getenv("FILE_STORAGE_PATH")
	envRestore := os.Getenv("RESTORE")
	envConnStr := os.Getenv("DATABASE_DSN")
	var err error

	if envAddr == "" {
		flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	} else {
		flagRunAddr = envAddr
	}

	if envInterval == "" {
		flag.IntVar(&flagStoreInterval, "i", 300, "interval to save metrics")
	} else {
		flagStoreInterval, err = strconv.Atoi(envInterval)
		if err != nil {
			logger.Error("Filed to convert string to int, envInterval", err)
		}
	}

	if envFilePath == "" {
		flag.StringVar(&flagFilePath, "f", defaultFilePath, "path to storage file")
	} else {
		flagFilePath = envFilePath
	}

	if envRestore == "" {
		flag.BoolVar(&flagRestore, "r", true, "restore metrics from file on startup")
	} else {
		flagRestore, err = strconv.ParseBool(envRestore)
		if err != nil {
			logger.Error("Filed to convert string to bool, envRestore", err)
		}
	}

	if envConnStr == "" {
		flag.StringVar(&flagDBConnStr, "d", "", "postgress connection string")
	} else {
		flagDBConnStr = envConnStr
	}
	flag.Parse()
}
