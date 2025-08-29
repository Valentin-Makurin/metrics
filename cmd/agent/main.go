package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
)

func main() {
	var cfg agent.ConfigAgent
	var err error

	cfg.HTTPAddr = os.Getenv("ADDRESS")

	cfg.ReportInterval, err = strconv.Atoi(os.Getenv("REPORT_INTERVAL"))
	if err != nil {
		log.Println("filed to pars string to int")
	}

	cfg.PollInterval, err = strconv.Atoi(os.Getenv("POLL_INTERVAL"))
	if err != nil {
		log.Println("filed to pars string to int")
	}

	if cfg.HTTPAddr == "" {
		flag.StringVar(&cfg.HTTPAddr, "a", "localhost:8080", "address and port to run server")
	}

	if cfg.ReportInterval == 0 {
		flag.IntVar(&cfg.ReportInterval, "r", 10, "reportInterval")
	}

	if cfg.PollInterval == 0 {
		flag.IntVar(&cfg.PollInterval, "p", 2, "pollInterval")
	}

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
