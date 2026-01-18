// Package agent предоставляет функциональность для сбора метрик системы.
package agent

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/config"
	pb "github.com/Valentin-Makurin/metrics/internal/proto"
)

// TestAgent_Start тестирует запуск и остановку агента сбора метрик.
func TestAgent_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := NewProtoConn(":3200")
	if err != nil {
		log.Fatalf("Ошибка создания grpc агента: %v", err)
	}
	defer conn.Close()

	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender(ctx, "localhost:8080", "", nil, pb.NewMetricsClient(conn))

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
	conn, err := NewProtoConn(":3200")
	if err != nil {
		log.Fatalf("Ошибка создания grpc агента: %v", err)
	}
	defer conn.Close()

	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender(context.Background(), "localhost:8080", "", nil, pb.NewMetricsClient(conn))

	agent := NewAgent(context.Background(), config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, collector, storage, sender)

	agent.writeMtr()

	if len(agent.storage.GetAllGauges()) == 0 {
		t.Error("Expected metrics to be written to guideStorage")
	}

	if agent.storage.GetCounter("PollCount") != 1 {
		t.Errorf("Expected PollCount 1")
	}
}
