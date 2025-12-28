package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Valentin-Makurin/metrics/internal/common"
	models "github.com/Valentin-Makurin/metrics/internal/model"
)

func BenchmarkRuntimeCollector_Collect(b *testing.B) {
	collector := NewRuntimeCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkHashCalculation(b *testing.B) {
	sender := NewHTTPSender("localhost:8080", "test-key", nil)
	metric := models.Metrics{
		ID:    "testMetric",
		MType: "gauge",
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = common.HashVal(sender.KeyH, jsonData)
	}
}
func BenchmarkHTTPSender_CreateRequest(b *testing.B) {
	sender := NewHTTPSender("localhost:8080", "test-key", nil)
	metric := models.Metrics{
		ID:    "testMetric",
		MType: "gauge",
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		b.Fatal(err)
	}

	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)
	_, err = gz.Write(jsonData)
	if err != nil {
		b.Fatal(err)
	}
	gz.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := "http://" + sender.baseURL + "/update/"
		req, err := http.NewRequest("POST", url, &compressedData)
		if err != nil {
			b.Fatal(err)
		}

		if sender.KeyH != "" {
			hashData := common.HashVal(sender.KeyH, jsonData)
			req.Header.Set("HashSHA256", hashData)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		_ = req
	}
}

func BenchmarkMemStorage_SetGauge(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.SetGauge("test_gauge", float64(i))
	}
}

func BenchmarkMemStorage_AddCounter(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.AddCounter("test_counter", 1)
	}
}

func BenchmarkMemStorage_GetGauge(b *testing.B) {
	storage := NewMemStorage()
	storage.SetGauge("test_gauge", 123.45)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetGauge("test_gauge")
	}
}
func BenchmarkMemStorage_GetCounter(b *testing.B) {
	storage := NewMemStorage()
	storage.AddCounter("test_counter", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetCounter("test_counter")
	}
}

func BenchmarkMemStorage_GetAllGauges(b *testing.B) {
	storage := NewMemStorage()

	for i := 0; i < 100; i++ {
		storage.SetGauge("gauge_"+string(rune('a'+i%26)), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetAllGauges()
	}
}

func BenchmarkMemStorage_GetAllCounters(b *testing.B) {
	storage := NewMemStorage()

	for i := 0; i < 100; i++ {
		storage.AddCounter("counter_"+string(rune('a'+i%26)), uint(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetAllCounters()
	}
}

func BenchmarkMemStorage_SetGauge_Parallel(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			storage.SetGauge("parallel_gauge", float64(i))
			i++
		}
	})
}

func BenchmarkMemStorage_AddCounter_Parallel(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			storage.AddCounter("parallel_counter", 1)
		}
	})
}

func BenchmarkMemStorage_GetGauge_Parallel(b *testing.B) {
	storage := NewMemStorage()
	storage.SetGauge("test_gauge", 123.45)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = storage.GetGauge("test_gauge")
		}
	})
}
func BenchmarkMemStorage_MixedOperations(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.SetGauge("mixed_gauge", float64(i))
		storage.AddCounter("mixed_counter", 1)

		_ = storage.GetGauge("mixed_gauge")
		_ = storage.GetCounter("mixed_counter")
	}
}
