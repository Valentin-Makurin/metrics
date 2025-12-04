// Package agent предоставляет функциональность для сбора метрик системы.
package agent

import "runtime"

// MetricsCollector определяет интерфейс для сборщиков метрик.
type MetricsCollector interface {
	// Collect собирает и возвращает статистику памяти.
	Collect() (runtime.MemStats, error)
}

// RuntimeCollector реализует MetricsCollector для сбора метрик времени выполнения Go.
type RuntimeCollector struct{}

// NewRuntimeCollector создает и возвращает новый экземпляр RuntimeCollector.
func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{}
}

// Collect собирает статистику памяти времени выполнения Go.
// Метод читает текущую статистику памяти через runtime.ReadMemStats
// и возвращает заполненную структуру MemStats.
func (c *RuntimeCollector) Collect() (runtime.MemStats, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats, nil
}
