package main

import (
	"log"
	"net/http"

	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/config"
	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/Valentin-Makurin/metrics/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg := config.ParseFlagsServer(sugar)

	mtrHandler := &handler.MtrHandler{}

	if cfg.DBConnStr != "" {
		dbConn, err := db.NewDatabase(cfg.DBConnStr, sugar)
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

		dbConn.RunMigrations()
		mtrHandler = handler.NewMtrHandler(dbConn, sugar)
	} else {
		storage := db.NewStorage(cfg.FilePath, cfg.StoreInterval, cfg.Restore, sugar)
		storage.PrepareFile()
		storage.StartTicker()
		mtrHandler = handler.NewMtrHandler(storage, sugar)
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(sugar))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.HashMiddleware(cfg.KeyH))
	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)

	r.Get("/ping", mtrHandler.HandlePing)
	r.Post("/update/", mtrHandler.HandlePostUpdate)
	r.Post("/updates/", mtrHandler.HandlePostUpdates)
	r.Post("/value/", mtrHandler.HandleGetValue)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Running server on", cfg.RunAddr)
	err = http.ListenAndServe(cfg.RunAddr, r)
	if err != nil {
		log.Println("Filed to start server", err)
	}
}
