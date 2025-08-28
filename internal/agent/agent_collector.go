package agent

import "runtime"

type MetricsCollector interface {
	Collect() (runtime.MemStats, error)
}

type RuntimeCollector struct{}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{}
}

func (c *RuntimeCollector) Collect() (runtime.MemStats, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats, nil
}
