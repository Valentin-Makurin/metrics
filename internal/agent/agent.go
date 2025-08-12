package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type ConfigAgent struct {
	HTTPAddr       string
	PollInterval   int
	ReportInterval int
}

type agent struct {
	client         *http.Client
	ctx            context.Context
	wg             sync.WaitGroup
	guideStorage   map[string]interface{}
	counterStorage map[string]uint
	mu             sync.RWMutex
	config         ConfigAgent
}

func NewAgent(ctx context.Context, cfg ConfigAgent) *agent {
	return &agent{
		client:         &http.Client{},
		ctx:            ctx,
		guideStorage:   make(map[string]interface{}),
		counterStorage: make(map[string]uint),
		config:         cfg,
	}
}

func (a *agent) Start() {
	a.wg.Add(2)
	go a.collect()
	go a.send()
	a.wg.Wait()
}

func (a *agent) collect() {
	ticker := time.NewTicker(time.Second * time.Duration(a.config.PollInterval))
	defer ticker.Stop()
	defer a.wg.Done()

	var memStats runtime.MemStats
	for {
		select {
		case <-ticker.C:
			a.writeMtr(&memStats)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *agent) send() {
	ticker := time.NewTicker(time.Second * time.Duration(a.config.ReportInterval))
	defer ticker.Stop()
	defer a.wg.Done()
	for {
		select {
		case <-ticker.C:
			a.postMtr()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *agent) writeMtr(memStats *runtime.MemStats) {

	runtime.ReadMemStats(memStats)
	a.mu.Lock()
	defer a.mu.Unlock()

	a.guideStorage["Alloc"] = memStats.Alloc
	a.guideStorage["BuckHashSys"] = memStats.BuckHashSys
	a.guideStorage["Frees"] = memStats.Frees
	a.guideStorage["GCCPUFraction"] = memStats.GCCPUFraction
	a.guideStorage["GCSys"] = memStats.GCSys
	a.guideStorage["HeapAlloc"] = memStats.HeapAlloc
	a.guideStorage["HeapIdle"] = memStats.HeapIdle
	a.guideStorage["HeapInuse"] = memStats.HeapInuse
	a.guideStorage["HeapObjects"] = memStats.HeapObjects
	a.guideStorage["HeapReleased"] = memStats.HeapReleased
	a.guideStorage["HeapSys"] = memStats.HeapSys
	a.guideStorage["LastGC"] = memStats.LastGC
	a.guideStorage["Lookups"] = memStats.Lookups
	a.guideStorage["MCacheInuse"] = memStats.MCacheInuse
	a.guideStorage["MCacheSys"] = memStats.MCacheSys
	a.guideStorage["MSpanInuse"] = memStats.MSpanInuse
	a.guideStorage["MSpanSys"] = memStats.MSpanSys
	a.guideStorage["Mallocs"] = memStats.Mallocs
	a.guideStorage["NextGC"] = memStats.NextGC
	a.guideStorage["NumForcedGC"] = memStats.NumForcedGC
	a.guideStorage["NumGC"] = memStats.NumGC
	a.guideStorage["OtherSys"] = memStats.OtherSys
	a.guideStorage["PauseTotalNs"] = memStats.PauseTotalNs
	a.guideStorage["StackInuse"] = memStats.StackInuse
	a.guideStorage["StackSys"] = memStats.StackSys
	a.guideStorage["Sys"] = memStats.Sys
	a.guideStorage["TotalAlloc"] = memStats.TotalAlloc
	a.guideStorage["RandomValue"] = rand.Int63n(321)

	val := a.counterStorage["PollCount"]
	a.counterStorage["PollCount"] = (val + 1)
}

func (a *agent) postMtr() {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for key, val := range a.guideStorage {
		url := prepareURL(models.Gauge, key, a.config.HTTPAddr, val)
		resp, err := a.client.Post(url, "text/plain", nil)
		if err != nil {
			log.Println("post Gauge error", err, "key", key, "val", val)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	for key, val := range a.counterStorage {
		url := prepareURL(models.Counter, key, a.config.HTTPAddr, val)
		resp, err := a.client.Post(url, "text/plain", nil)
		if err != nil {
			log.Println("post Counter error", err, "key", key, "val", val)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func prepareURL(ty, key, addr string, val interface{}) string {
	strVal := fmt.Sprintf("%v", val)
	return "http://" + addr + "/update/" + ty + "/" + key + "/" + strVal
}
