// Package models предоставляет основные структуры данных для работы с метриками.
package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics представляет структуру метрики для передачи между клиентом и сервером.
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Event представляет событие аудита для отслеживания операций с метриками.
type Event struct {
	IPAddress string   `json:"ip_address"`
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
}
