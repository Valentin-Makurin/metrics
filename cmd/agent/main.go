package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
)

func main() {
	var cfg agent.ConfigAgent

	flag.StringVar(&cfg.HTTPAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reportInterval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "pollInterval")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(60 * time.Second)
		cancel()
	}()

	storage := agent.NewMemStorage()
	collector := agent.NewRuntimeCollector()
	sender := agent.NewHTTPSender(cfg.HTTPAddr)

	agent := agent.NewAgent(ctx, cfg, collector, storage, sender)
	agent.Start()

	log.Println("job is done")
}
