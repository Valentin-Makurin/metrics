package agent

import (
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
		url := prepareURL(models.Gauge, key, s.baseURL, val)
		resp, err := s.client.Post(url, "text/plain", nil)
		if err != nil {
			log.Println("post Gauge error", err, "key", key, "val", val)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	for key, val := range counters {
		url := prepareURL(models.Counter, key, s.baseURL, val)
		resp, err := s.client.Post(url, "text/plain", nil)
		if err != nil {
			log.Println("post Counter error", err, "key", key, "val", val)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return nil
}

func prepareURL(ty, key, addr string, val interface{}) string {
	strVal := fmt.Sprintf("%v", val)
	return "http://" + addr + "/update/" + ty + "/" + key + "/" + strVal
}
