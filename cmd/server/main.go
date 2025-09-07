package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/Valentin-Makurin/metrics/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var flagRunAddr string

func main() {
	flagRunAddr = os.Getenv("ADDRESS")
	if flagRunAddr == "" {
		flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
		flag.Parse()
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	storage := db.NewStorage()
	mtrHandler := handler.NewMtrHandler(storage)

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(sugar))
	r.Use(middleware.GzipMiddleware)
	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Post("/update/", mtrHandler.HandlePostUpdate)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)
	r.Post("/value/", mtrHandler.HandleGetValue)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Running server on", flagRunAddr)
	err = http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		log.Println("Filed to start server", err)
	}
}
