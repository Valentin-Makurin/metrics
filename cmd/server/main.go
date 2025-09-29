package main

import (
	// "flag"
	"log"
	"net/http"

	// "os"
	// "strconv"

	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/config"
	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/Valentin-Makurin/metrics/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// var (
// 	flagRunAddr       string
// 	flagStoreInterval int
// 	flagFilePath      string
// 	flagRestore       bool
// 	flagDBConnStr     string
// )

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// initFlags(sugar)
	cfg := config.ParseFlags(sugar)

	mtrHandler := &handler.MtrHandler{}

	if cfg.DBConnStr != "" { //flagDBConnStr
		dbConn, err := db.NewDatabase(cfg.DBConnStr, sugar) //flagDBConnStr
		if err != nil {
			log.Fatalf("Ошибка подключения к БД: %v", err)
		}
		defer dbConn.Close()

		_, err = common.RetryOperation(
			func() (interface{}, error) {
				return nil, dbConn.Ping()
			},
			db.IsTemporaryError)
		if err != nil {
			log.Fatalf("Ошибка пинга к БД: %v", err)
		}

		// flag := false
		// interval := map[int]time.Duration{0: 1 * time.Second, 1: 3 * time.Second, 2: 5 * time.Second}
		// for i := range 3 {
		// 	err = dbConn.Ping()
		// 	if err == nil {
		// 		flag = true
		// 		break
		// 	}
		// 	if db.IsTemporaryError(err) {
		// 		tm := interval[i]
		// 		time.Sleep(tm)
		// 	} else {
		// 		log.Fatalf("failed attempt error: %v", err)
		// 	}
		// }
		// if !flag {
		// 	log.Fatalf("Ошибка пинга к БД: %v", err)
		// }

		dbConn.RunMigrations()
		mtrHandler = handler.NewMtrHandler(dbConn, sugar)
	} else {
		storage := db.NewStorage(cfg.FilePath, cfg.StoreInterval, cfg.Restore, sugar) //flagFilePath, flagStoreInterval, flagRestore
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
	r.Post("/updates/", mtrHandler.HandlePostUpdates)
	r.Post("/value/", mtrHandler.HandleGetValue)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Running server on", cfg.RunAddr) //flagRunAddr
	err = http.ListenAndServe(cfg.RunAddr, r)     //flagRunAddr
	if err != nil {
		log.Println("Filed to start server", err)
	}
}

// func initFlags(logger *zap.SugaredLogger) {

// 	addrTmp := flag.String("a", "localhost:8080", "address and port to run server")
// 	storeIntervalTmp := flag.Int("i", 2, "interval to save metrics")
// 	filePathTmp := flag.String("f", "/tmp/metrics.json", "path to storage file")
// 	restoreTmp := flag.Bool("r", true, "restore metrics from file on startup")
// 	connStrTmp := flag.String("d", "", "postgress connection string")
// 	flag.Parse()

// 	varAdrHost, ok := os.LookupEnv("ADDRESS")
// 	if ok {
// 		addrTmp = &varAdrHost
// 	}
// 	varStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
// 	if ok {
// 		intStoreInterval, err := strconv.Atoi(varStoreInterval)
// 		if err != nil {
// 			logger.Error("Filed to convert string to int, envInterval", err)
// 		}
// 		storeIntervalTmp = &intStoreInterval
// 	}
// 	varFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
// 	if ok {
// 		filePathTmp = &varFileStoragePath
// 	}
// 	varRestore, ok := os.LookupEnv("RESTORE")
// 	if ok {
// 		boolVarRestore, err := strconv.ParseBool(varRestore)
// 		if err != nil {
// 			logger.Error("Filed to convert string to bool, envRestore", err)
// 		}
// 		restoreTmp = &boolVarRestore
// 	}
// 	varDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
// 	if ok {
// 		connStrTmp = &varDatabaseDSN
// 	}

// 	flagRunAddr = *addrTmp
// 	flagStoreInterval = *storeIntervalTmp
// 	flagFilePath = *filePathTmp
// 	flagRestore = *restoreTmp
// 	flagDBConnStr = *connStrTmp

// 	log.Println("envAddr", flagRunAddr)
// 	log.Println("envInterval", flagStoreInterval)
// 	log.Println("envFilePath", flagFilePath)
// 	log.Println("envRestore", flagRestore)
// 	log.Println("envConnStr", flagDBConnStr)

// }
