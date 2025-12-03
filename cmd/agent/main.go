package main

import (
	"context"
	"log"

	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
	"github.com/Valentin-Makurin/metrics/internal/config"
)

func main() {

	cfg := config.ParseFlagsAgent()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Second)
		cancel()
	}()

	storage := agent.NewMemStorage()
	collector := agent.NewRuntimeCollector()
	sender := agent.NewHTTPSender(cfg.HTTPAddr, cfg.KeyH)

	agent := agent.NewAgent(ctx, cfg, collector, storage, sender)
	agent.Start()

	log.Println("job is done")
}
