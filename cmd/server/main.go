package main

import (
	"log"
	"net/http"

	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/handler"
)

func main() {
	storage := db.NewStorage()
	mtrHandler := handler.NewMtrHandler(storage)

	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", http.HandlerFunc(mtrHandler.HandlePost))
	if err != nil {
		log.Println("Filed to start server", err)
	}
}
