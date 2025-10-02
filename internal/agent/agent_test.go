package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/config"
)

func TestAgent_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender("localhost:8080", "")

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

func TestAgent_writeMtr(t *testing.T) {
	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender("localhost:8080", "")

	agent := NewAgent(context.Background(), config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, collector, storage, sender)

	agent.writeMtr()

	if len(agent.storage.GetAllGauges()) == 0 {
		t.Error("Expected metrics to be written to guideStorage")
	}

	if agent.storage.GetCounter("PollCount") != 1 {
		t.Errorf("Expected PollCount 1")
	}
}

func TestAgent_postMtr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	storage := NewMemStorage()
	sender := NewHTTPSender(ts.URL, "")
	agent := NewAgent(context.Background(), config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, nil, storage, sender)

	agent.storage.SetGauge("TestGauge", 123.45)
	agent.storage.AddCounter("TestCounter", 42)

	agent.postMtr()

	if val := agent.storage.GetGauge("TestGauge"); val != 123.45 {
		t.Error("Gauge data changed unexpectedly")
	}

	if val := agent.storage.GetCounter("TestCounter"); val != 42 {
		t.Error("Counter data changed unexpectedly")
	}
}

func TestAgent_ConcurrentAccess(t *testing.T) {
	collector := NewRuntimeCollector()
	storage := NewMemStorage()
	sender := NewHTTPSender("localhost:8080", "")

	agent := NewAgent(context.Background(), config.ConfigAgent{PollInterval: 2, ReportInterval: 10}, collector, storage, sender)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.writeMtr()
			agent.postMtr()
		}()
	}
	wg.Wait()

	if agent.storage.GetCounter("PollCount") != 10 {
		t.Errorf("Expected PollCount 10")
	}
}
