package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/config"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	common.FirstPrint(os.Stdout, buildVersion, buildDate, buildCommit)

	cfg := config.ParseFlagsAgent()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	storage := agent.NewMemStorage()
	collector := agent.NewRuntimeCollector()
	sender := agent.NewHTTPSender(cfg.HTTPAddr, cfg.KeyH)

	agent := agent.NewAgent(ctx, cfg, collector, storage, sender)

	agentDone := make(chan struct{})

	go func() {
		defer close(agentDone)
		agent.Start()
	}()

	select {
	case sig := <-shutdown:
		log.Printf("received signal - %v", sig)
		cancel()

		select {
		case <-agentDone:
			log.Println("agent stopped gracefully")
		case <-time.After(5 * time.Second):
			log.Println("agent shutdown timeout")
		}
	case <-time.After(30 * time.Second):
		log.Println("time is over")
		cancel()
	case <-agentDone:
		log.Println("agent finished")
	}

	log.Println("job is done")
}
