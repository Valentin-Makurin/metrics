package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type MetricsSender interface {
	Send(gauges map[string]any, counters map[string]uint) error
}

type HTTPSender struct {
	client  *http.Client
	baseURL string
}

func NewHTTPSender(baseURL string) *HTTPSender {
	return &HTTPSender{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (s *HTTPSender) Send(gauges map[string]any, counters map[string]uint) error {
	for key, val := range gauges {
		if err := s.sendGaugeMetric(key, val); err != nil {
			log.Println("post Gauge error", err, "key", key, "val", val)
		}
	}

	for key, val := range counters {
		if err := s.sendCounterMetric(key, val); err != nil {
			log.Println("post Counter error", err, "key", key, "val", val)
		}
	}
	return nil
}

func (s *HTTPSender) sendGaugeMetric(key string, value interface{}) error {
	metric := models.Metrics{
		ID:    key,
		MType: models.Gauge,
	}

	switch v := value.(type) {
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
	default:
		return fmt.Errorf("unsupported gauge value type: %T", value)
	}

	return s.sendJSONRequest(metric)
}

func (s *HTTPSender) sendCounterMetric(key string, value uint) error {
	intVal := int64(value)
	metric := models.Metrics{
		ID:    key,
		MType: models.Counter,
		Delta: &intVal,
	}

	return s.sendJSONRequest(metric)
}

func (s *HTTPSender) sendJSONRequest(metric models.Metrics) error {
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	url := "http://" + s.baseURL + "/update/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
