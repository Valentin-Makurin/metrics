package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"go.uber.org/zap"
)

// ExampleMtrHandler_HandlePost демонстрирует использование текстового API для обновления метрик.
func ExampleMtrHandler_HandlePost() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)

	server := httptest.NewServer(http.HandlerFunc(handler.HandlePost))
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/memory_usage/75.5", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Gauge metric updated, status: %s", resp.Status)
	// Output: Gauge metric updated, status: 200 OK
}

// ExampleMtrHandler_HandlePost_gauge демонстрирует обновление gauge метрики через текстовый API.
func ExampleMtrHandler_HandlePost_gauge() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePost))
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/cpu_temperature/42.3", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("CPU temperature updated, status: %s", resp.Status)
	// Output: CPU temperature updated, status: 200 OK
}

// ExampleMtrHandler_HandlePost_counter демонстрирует обновление counter метрики через текстовый API.
func ExampleMtrHandler_HandlePost_counter() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePost))
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/counter/request_count/100", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request counter incremented, status: %s", resp.Status)
	// Output: Request counter incremented, status: 200 OK
}

// ExampleMtrHandler_HandlePostUpdate демонстрирует использование JSON API для обновления одиночной метрики.
func ExampleMtrHandler_HandlePostUpdate() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePostUpdate))
	defer server.Close()

	metric := models.Metrics{
		ID:    "disk_usage",
		MType: models.Gauge,
		Value: func() *float64 { v := 87.2; return &v }(),
	}

	jsonData, _ := json.Marshal(metric)

	resp, err := http.Post(server.URL+"/update/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Disk usage metric updated via JSON API, status: %s", resp.Status)
	// Output: Disk usage metric updated via JSON API, status: 200 OK
}

// ExampleMtrHandler_HandlePostUpdate_counter демонстрирует обновление counter метрики через JSON API.
func ExampleMtrHandler_HandlePostUpdate_counter() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePostUpdate))
	defer server.Close()

	metric := models.Metrics{
		ID:    "user_sessions",
		MType: models.Counter,
		Delta: func() *int64 { v := int64(150); return &v }(),
	}

	jsonData, _ := json.Marshal(metric)

	resp, err := http.Post(server.URL+"/update/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("User sessions counter updated, status: %s", resp.Status)
	// Output: User sessions counter updated, status: 200 OK
}

// ExampleMtrHandler_HandlePostUpdates демонстрирует пакетное обновление метрик через JSON API.
func ExampleMtrHandler_HandlePostUpdates() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePostUpdates))
	defer server.Close()

	metrics := []models.Metrics{
		{
			ID:    "memory_used",
			MType: models.Gauge,
			Value: func() *float64 { v := 2048.5; return &v }(),
		},
		{
			ID:    "errors_total",
			MType: models.Counter,
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
		{
			ID:    "errors_total", // Дублирование для агрегации counter
			MType: models.Counter,
			Delta: func() *int64 { v := int64(3); return &v }(),
		},
	}

	jsonData, _ := json.Marshal(metrics)

	resp, err := http.Post(server.URL+"/updates/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Batch metrics updated, status: %s", resp.Status)
	// Output: Batch metrics updated, status: 200 OK
}

// ExampleMtrHandler_HandleGet демонстрирует получение значений метрик через текстовый API.
func ExampleMtrHandler_HandleGet() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	value := 95.7
	storage.GetValResult = models.Metrics{
		ID:    "cpu_load",
		MType: models.Gauge,
		Value: &value,
	}

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandleGet))
	defer server.Close()

	resp, err := http.Get(server.URL + "/value/gauge/cpu_load")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("CPU load value: %s", body)
	// Output: CPU load value: 95.7
}

// ExampleMtrHandler_HandleGetValue демонстрирует получение значений метрик через JSON API.
func ExampleMtrHandler_HandleGetValue() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	delta := int64(42)
	storage.GetValResult = models.Metrics{
		ID:    "api_calls",
		MType: models.Counter,
		Delta: &delta,
	}

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandleGetValue))
	defer server.Close()

	request := models.Metrics{
		ID:    "api_calls",
		MType: models.Counter,
	}

	jsonData, _ := json.Marshal(request)

	resp, err := http.Post(server.URL+"/value/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	var metric models.Metrics
	json.NewDecoder(resp.Body).Decode(&metric)

	fmt.Printf("API calls counter value: %d", *metric.Delta)
	// Output: API calls counter value: 42
}

// ExampleMtrHandler_HandleRoot демонстрирует получение HTML страницы со всеми метриками.
func ExampleMtrHandler_HandleRoot() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	storage.GetAllValResult = map[string]string{
		"cpu_usage":     "75.5",
		"memory_used":   "2048.2",
		"request_count": "1000",
		"error_count":   "5",
	}

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandleRoot))
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("HTML page with metrics received, status: %s", resp.Status)
	// Output: HTML page with metrics received, status: 200 OK
}

// ExampleMtrHandler_HandlePing демонстрирует проверку доступности хранилища.
func ExampleMtrHandler_HandlePing() {
	storage := &MockStorage{}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePing))
	defer server.Close()

	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Storage is available")
	} else {
		fmt.Println("Storage is unavailable")
	}
	// Output: Storage is available
}

// ExampleMtrHandler_HandlePing_error демонстрирует обработку ошибки при проверке хранилища.
func ExampleMtrHandler_HandlePing_error() {
	storage := &MockStorage{
		PingError: fmt.Errorf("connection failed"),
	}
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync()

	handler := NewMtrHandler(storage, sugar)
	server := httptest.NewServer(http.HandlerFunc(handler.HandlePing))
	defer server.Close()

	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Storage check failed with status: %s", resp.Status)
	// Output: Storage check failed with status: 500 Internal Server Error
}
