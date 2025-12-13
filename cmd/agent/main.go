package main

import (
	"context"
	"log"

	"time"

	"github.com/Valentin-Makurin/metrics/internal/agent"
	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/config"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	common.FirstPrint(buildVersion, buildDate, buildCommit)

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
