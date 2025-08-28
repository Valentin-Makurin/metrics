package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/go-chi/chi/v5"
)

var flagRunAddr string

func main() {

	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	storage := db.NewStorage()
	mtrHandler := handler.NewMtrHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)
	r.Get("/", mtrHandler.HandleRoot)

	log.Println("Running server on", flagRunAddr)
	err := http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		log.Println("Filed to start server", err)
	}
}
