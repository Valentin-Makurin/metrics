package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
)

var cfg agent.ConfigAgent

func main() {
	flag.StringVar(&cfg.HttpAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reportInterval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "pollInterval")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(60 * time.Second)
		cancel()
	}()
	agent := agent.NewAgent(ctx, cfg)
	agent.Start()

	log.Println("job is done")
}
