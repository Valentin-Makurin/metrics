package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"sync"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/config"
	models "github.com/Valentin-Makurin/metrics/internal/model"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type agent struct {
	ctx       context.Context
	wg        sync.WaitGroup
	storage   MetricsStorage
	collector MetricsCollector
	sender    MetricsSender
	config    config.ConfigAgent
}

func NewAgent(
	ctx context.Context,
	cfg config.ConfigAgent,
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
	ch := make(chan []models.Metrics, 40)
	a.wg.Add(3)
	go a.collect()
	go a.router(ch)
	// go a.send()
	go a.send2(ch)
	a.wg.Wait()
}

func (a *agent) collect() {
	ticker := time.NewTicker(time.Second * time.Duration(a.config.PollInterval))
	defer ticker.Stop()
	defer a.wg.Done()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			wg.Add(2)
			go func() {
				defer wg.Done()
				a.writeMtr()
			}()
			go func() {
				defer wg.Done()
				a.writeMtrExtra()
			}()
			wg.Wait()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *agent) router(ch chan<- []models.Metrics) {
	ticker := time.NewTicker(time.Second * time.Duration(a.config.ReportInterval))
	defer ticker.Stop()
	defer a.wg.Done()

	for {
		select {
		case <-ticker.C:

			gaugesData := a.storage.GetAllGauges()
			couterData := a.storage.GetAllCounters()

			res := []models.Metrics{}
			for key, val := range gaugesData {
				metric := models.Metrics{
					ID:    key,
					MType: models.Gauge,
				}

				switch v := val.(type) {
				case float64:
					metric.Value = &v
				case uint64:
					floatVal := float64(v)
					metric.Value = &floatVal
				case uint32:
					floatVal := float64(v)
					metric.Value = &floatVal
				case int64:
					floatVal := float64(v)
					metric.Value = &floatVal
				case int:
					floatVal := float64(v)
					metric.Value = &floatVal
				default:
					fmt.Printf("unsupported gauge value type: %T", val)
				}
				res = append(res, metric)
				// ch <- metric

			}

			for key, val := range couterData {
				intVal := int64(val)
				metric := models.Metrics{
					ID:    key,
					MType: models.Counter,
					Delta: &intVal,
				}
				// ch <- metric
				res = append(res, metric)
			}
			ch <- res
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
func (a *agent) send2(ch <-chan []models.Metrics) {
	for i := range a.config.RateLimit {
		go func() {
			fmt.Println("job started", i)
			for val := range ch {
				a.postMtr2(val)
			}
		}()
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

func (a *agent) writeMtrExtra() {
	memVal, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Failed to Collect memVal Extra")
		return
	}

	cpuVal, err := cpu.Counts(true)
	if err != nil {
		log.Printf("Failed to Collect cpuVal Extra")
		return
	}

	a.storage.SetGauge("TotalMemory", memVal.Total)
	a.storage.SetGauge("FreeMemory", memVal.Free)
	a.storage.SetGauge("CPUutilization1", cpuVal)

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

func (a *agent) postMtr2(mtrs []models.Metrics) {
	for _, val := range mtrs {
		err := a.sender.SendJSONRequest(val)
		if err != nil {
			log.Printf("Failed to SendJSONRequest metrics")
		}
	}

	err := a.sender.SendJSONRequestBatch(mtrs)
	if err != nil {
		log.Printf("Failed to SendJSONRequestBatch metrics")
	}

	// err := a.sender.Send(a.storage.GetAllGauges(), a.storage.GetAllCounters())
	// if err != nil {
	// 	log.Printf("Failed to Send metrics")
	// }
	// err = a.sender.SendBatch(a.storage.GetAllGauges(), a.storage.GetAllCounters())
	// if err != nil {
	// 	log.Printf("Failed to Send metrics")
	// }
}
