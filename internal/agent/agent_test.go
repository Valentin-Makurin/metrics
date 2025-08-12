package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestAgent_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	agent := NewAgent(ctx, ConfigAgent{HTTPAddr: "localhost:8080", PollInterval: 2, ReportInterval: 10})

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
	agent := NewAgent(context.Background(), ConfigAgent{HTTPAddr: "localhost:8080", PollInterval: 2, ReportInterval: 10})
	var memStats runtime.MemStats

	agent.writeMtr(&memStats)

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	if len(agent.guideStorage) == 0 {
		t.Error("Expected metrics to be written to guideStorage")
	}

	if agent.counterStorage["PollCount"] != 1 {
		t.Errorf("Expected PollCount 1")
	}
}

func TestAgent_postMtr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	agent := NewAgent(context.Background(), ConfigAgent{HTTPAddr: "localhost:8080", PollInterval: 2, ReportInterval: 10})
	agent.client = ts.Client()

	agent.mu.Lock()
	agent.guideStorage["TestGauge"] = 123.45
	agent.counterStorage["TestCounter"] = 42
	agent.mu.Unlock()

	agent.postMtr()

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	if val, ok := agent.guideStorage["TestGauge"]; !ok || val != 123.45 {
		t.Error("Gauge data changed unexpectedly")
	}

	if val, ok := agent.counterStorage["TestCounter"]; !ok || val != 42 {
		t.Error("Counter data changed unexpectedly")
	}
}

func TestAgent_ConcurrentAccess(t *testing.T) {
	agent := NewAgent(context.Background(), ConfigAgent{HTTPAddr: "localhost:8080", PollInterval: 2, ReportInterval: 10})
	var memStats runtime.MemStats

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.writeMtr(&memStats)
			agent.postMtr()
		}()
	}
	wg.Wait()

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	if agent.counterStorage["PollCount"] != 10 {
		t.Errorf("Expected PollCount 10")
	}
}
