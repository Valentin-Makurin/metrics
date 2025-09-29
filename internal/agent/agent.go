package agent

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

type ConfigAgent struct {
	HTTPAddr       string
	PollInterval   int
	ReportInterval int
}

type agent struct {
	ctx       context.Context
	wg        sync.WaitGroup
	storage   MetricsStorage
	collector MetricsCollector
	sender    MetricsSender
	config    ConfigAgent
}

func NewAgent(
	ctx context.Context,
	cfg ConfigAgent,
	collector MetricsCollector,
	storage MetricsStorage,
	sender MetricsSender,
) *agent {
	return &agent{
		ctx:       ctx,
		storage:   storage,
		collector: collector,
		sender:    sender,
		config:    cfg,
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

	for {
		select {
		case <-ticker.C:
			a.writeMtr()
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

func (a *agent) writeMtr() {
	memStats, err := a.collector.Collect()
	if err != nil {
		log.Printf("Failed to Collect metrics")
	}

	a.storage.SetGauge("Alloc", memStats.Alloc)
	a.storage.SetGauge("BuckHashSys", memStats.BuckHashSys)
	a.storage.SetGauge("Frees", memStats.Frees)
	a.storage.SetGauge("GCCPUFraction", memStats.GCCPUFraction)
	a.storage.SetGauge("GCSys", memStats.GCSys)
	a.storage.SetGauge("HeapAlloc", memStats.HeapAlloc)
	a.storage.SetGauge("HeapIdle", memStats.HeapIdle)
	a.storage.SetGauge("HeapInuse", memStats.HeapInuse)
	a.storage.SetGauge("HeapObjects", memStats.HeapObjects)
	a.storage.SetGauge("HeapReleased", memStats.HeapReleased)
	a.storage.SetGauge("HeapSys", memStats.HeapSys)
	a.storage.SetGauge("LastGC", memStats.LastGC)
	a.storage.SetGauge("Lookups", memStats.Lookups)
	a.storage.SetGauge("MCacheInuse", memStats.MCacheInuse)
	a.storage.SetGauge("MCacheSys", memStats.MCacheSys)
	a.storage.SetGauge("MSpanInuse", memStats.MSpanInuse)
	a.storage.SetGauge("MSpanSys", memStats.MSpanSys)
	a.storage.SetGauge("Mallocs", memStats.Mallocs)
	a.storage.SetGauge("NextGC", memStats.NextGC)
	a.storage.SetGauge("NumForcedGC", memStats.NumForcedGC)
	a.storage.SetGauge("NumGC", memStats.NumGC)
	a.storage.SetGauge("OtherSys", memStats.OtherSys)
	a.storage.SetGauge("PauseTotalNs", memStats.PauseTotalNs)
	a.storage.SetGauge("StackInuse", memStats.StackInuse)
	a.storage.SetGauge("StackSys", memStats.StackSys)
	a.storage.SetGauge("Sys", memStats.Sys)
	a.storage.SetGauge("TotalAlloc", memStats.TotalAlloc)
	a.storage.SetGauge("RandomValue", rand.Int63n(321))

	a.storage.AddCounter("PollCount", 1)
}

func (a *agent) postMtr() {
	err := a.sender.Send(a.storage.GetAllGauges(), a.storage.GetAllCounters())
	if err != nil {
		log.Printf("Failed to Send metrics")
	}
	err = a.sender.SendBatch(a.storage.GetAllGauges(), a.storage.GetAllCounters())
	if err != nil {
		log.Printf("Failed to Send metrics")
	}
}
