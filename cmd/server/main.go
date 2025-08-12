package main

import (
	"log"
	"net/http"

	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	storage := db.NewStorage()
	mtrHandler := handler.NewMtrHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Println("Filed to start server", err)
	}
}
