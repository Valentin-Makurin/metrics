package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	mtrs, err := s.prepareMtrData(gauges, counters)
	if err != nil {
		return fmt.Errorf("failed prepare mtr data: %w", err)
	}

	s.sendJSONRequestBatch(mtrs)

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
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = gz.Write(jsonData)
	if err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	url := "http://" + s.baseURL + "/update/"
	req, err := http.NewRequest("POST", url, &compressedData)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		fmt.Print("gzip ok ")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *HTTPSender) sendJSONRequestBatch(metric any) error {
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = gz.Write(jsonData)
	if err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	url := "http://" + s.baseURL + "/updates/"
	req, err := http.NewRequest("POST", url, &compressedData)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		fmt.Print("gzip ok ")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *HTTPSender) prepareMtrData(gauges map[string]any, counters map[string]uint) ([]models.Metrics, error) {
	res := []models.Metrics{}
	for key, val := range gauges {
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
		default:
			return nil, fmt.Errorf("unsupported gauge value type: %T", val)
		}
		res = append(res, metric)
	}

	for key, val := range counters {
		intVal := int64(val)
		metric := models.Metrics{
			ID:    key,
			MType: models.Counter,
			Delta: &intVal,
		}
		res = append(res, metric)
	}

	return res, nil
}
