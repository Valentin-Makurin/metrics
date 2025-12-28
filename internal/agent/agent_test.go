// Package agent предоставляет функциональность для сбора метрик системы.
package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/config"
)

// TestAgent_Start тестирует запуск и остановку агента сбора метрик.
func TestAgent_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender("localhost:8080", "", nil)

	agent := NewAgent(ctx, config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, collector, storage, sender)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		agent.Start()
	}()

	time.Sleep(time.Second)
	cancel()
	wg.Wait()
}

// TestAgent_writeMtr тестирует метод writeMtr агента, который записывает метрики в хранилище.
func TestAgent_writeMtr(t *testing.T) {
	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender("localhost:8080", "", nil)

	agent := NewAgent(context.Background(), config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, collector, storage, sender)

	agent.writeMtr()

	if len(agent.storage.GetAllGauges()) == 0 {
		t.Error("Expected metrics to be written to guideStorage")
	}

	if agent.storage.GetCounter("PollCount") != 1 {
		t.Errorf("Expected PollCount 1")
	}
}
