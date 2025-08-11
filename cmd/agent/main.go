package main

import (
	"context"
	"log"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(60 * time.Second)
		cancel()
	}()
	agent := agent.NewAgent(2, 10, ctx)
	agent.Start()

	log.Println("job done")
}
