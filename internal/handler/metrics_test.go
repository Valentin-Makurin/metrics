// Package handler предоставляет HTTP-обработчики для работы с метриками.
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Valentin-Makurin/metrics/internal/db"
	models "github.com/Valentin-Makurin/metrics/internal/model"
	"go.uber.org/zap"
)

// TestHandlePost_MethodNotAllowed тестирует обработку недопустимых HTTP-методов.
func TestHandlePost_MethodNotAllowed(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	storage := db.NewStorage("test", 0, false, sugar)
	handler := NewMtrHandler(storage, sugar)

	tests := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range tests {
		req := httptest.NewRequest(method, "/update/gauge/test/123", nil)
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("For method %s expected status %d, got %d", method, http.StatusMethodNotAllowed, w.Code)
		}
	}
}

// TestHandlePost_InvalidPath тестирует обработку некорректных URL путей.
func TestHandlePost_InvalidPath(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	storage := db.NewStorage("test", 0, false, sugar)
	handler := NewMtrHandler(storage, sugar)

	testCases := []struct {
		path       string
		statusCode int
	}{
		{"/update", http.StatusNotFound},
		{"/update/gauge", http.StatusNotFound},
		{"/update/gauge/", http.StatusNotFound},
		{"/update/gauge/test", http.StatusNotFound},
		{"/update/gauge/test/123/extra", http.StatusNotFound},
		{"/wrong/gauge/test/123", http.StatusNotFound},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodPost, tc.path, nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		if w.Code != tc.statusCode {
			t.Errorf("For path %s expected status %d, got %d", tc.path, tc.statusCode, w.Code)
		}
	}
}

// TestHandlePost_InvalidMetricType тестирует обработку некорректных типов метрик.
func TestHandlePost_InvalidMetricType(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	storage := db.NewStorage("test", 0, false, sugar)
	handler := NewMtrHandler(storage, sugar)

	req := httptest.NewRequest(http.MethodPost, "/update/invalid/test/123", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestHandleRoot тестирует корневой обработчик возвращающий HTML страницу.
func TestHandleRoot(t *testing.T) {
	storage := &db.MtrStorage{}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	handler := NewMtrHandler(storage, sugar)

	gaugeValue := 123.45
	counterValue := int64(42)
	storage.SetVal("test_gauge", models.Metrics{MType: models.Gauge, Value: &gaugeValue})
	storage.SetVal("test_counter", models.Metrics{MType: models.Counter, Delta: &counterValue})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.HandleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}

	body := w.Body.String()

	if !strings.Contains(body, html.EscapeString("test_gauge")) {
		t.Error("Gauge metric name not found in HTML")
	}

	if !strings.Contains(body, html.EscapeString("123.45")) {
		t.Error("Gauge metric value not found in HTML")
	}

	if !strings.Contains(body, html.EscapeString("test_counter")) {
		t.Error("Counter metric name not found in HTML")
	}

	if !strings.Contains(body, html.EscapeString("42")) {
		t.Error("Counter metric value not found in HTML")
	}
}

// TestHandlePostUpdate тестирует JSON обработчик для обновления одиночной метрики.
func TestHandlePostUpdate(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("MethodNotAllowed", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/update", nil)
			w := httptest.NewRecorder()
			handler.HandlePostUpdate(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("For method %s expected %d, got %d", method, http.StatusMethodNotAllowed, w.Code)
			}
		}
	})

	t.Run("InvalidContentType", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected %d, got %d", http.StatusUnsupportedMediaType, w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("EmptyMetricID", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "",
			MType: models.Gauge,
			Value: float64Ptr(123.45),
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("InvalidMetricType", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test",
			MType: "invalid",
			Value: float64Ptr(123.45),
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("GaugeWithoutValue", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_gauge",
			MType: models.Gauge,
			Value: nil,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CounterWithoutDelta", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_counter",
			MType: models.Counter,
			Delta: nil,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("SuccessGauge", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_gauge",
			MType: models.Gauge,
			Value: float64Ptr(123.45),
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		if !storage.SetValCalled {
			t.Error("SetVal was not called")
		}

		if len(storage.SetValArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(storage.SetValArgs))
		}

		key, ok := storage.SetValArgs[0].(string)
		if !ok || key != "test_gauge" {
			t.Errorf("Expected key 'test_gauge', got %v", storage.SetValArgs[0])
		}
	})

	t.Run("SuccessCounter", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_counter",
			MType: models.Counter,
			Delta: int64Ptr(42),
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		if !storage.AddValCalled {
			t.Error("AddVal was not called")
		}

		if len(storage.AddValArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(storage.AddValArgs))
		}

		key, ok := storage.AddValArgs[0].(string)
		if !ok || key != "test_counter" {
			t.Errorf("Expected key 'test_counter', got %v", storage.AddValArgs[0])
		}
	})
}

// TestHandlePostUpdates тестирует JSON обработчик для пакетного обновления метрик.
func TestHandlePostUpdates(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("MethodNotAllowed", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/updates", nil)
			w := httptest.NewRecorder()
			handler.HandlePostUpdates(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("For method %s expected %d, got %d", method, http.StatusMethodNotAllowed, w.Code)
			}
		}
	})

	t.Run("InvalidContentType", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/updates", nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected %d, got %d", http.StatusUnsupportedMediaType, w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("EmptyMetricIDInArray", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "valid",
				MType: models.Gauge,
				Value: float64Ptr(123.45),
			},
			{
				ID:    "",
				MType: models.Gauge,
				Value: float64Ptr(456.78),
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("InvalidMetricTypeInArray", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "test",
				MType: "invalid",
				Value: float64Ptr(123.45),
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("GaugeWithoutValueInArray", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "test_gauge",
				MType: models.Gauge,
				Value: nil,
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CounterWithoutDeltaInArray", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "test_counter",
				MType: models.Counter,
				Delta: nil,
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CounterAggregation", func(t *testing.T) {
		storage := &MockStorage{}
		var capturedGauge []models.Metrics
		var capturedCounter map[string]models.Metrics

		storage.UpsertBatchCustom = func(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error {
			capturedGauge = GaugeMtr
			capturedCounter = CntMtr
			return nil
		}

		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "counter1",
				MType: models.Counter,
				Delta: int64Ptr(10),
			},
			{
				ID:    "counter1",
				MType: models.Counter,
				Delta: int64Ptr(20),
			},
			{
				ID:    "counter2",
				MType: models.Counter,
				Delta: int64Ptr(30),
			},
			{
				ID:    "gauge1",
				MType: models.Gauge,
				Value: float64Ptr(123.45),
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		// Проверяем агрегацию counter
		if len(capturedCounter) != 2 {
			t.Errorf("Expected 2 counters, got %d", len(capturedCounter))
		}

		if *capturedCounter["counter1"].Delta != 30 {
			t.Errorf("Expected counter1 delta 30, got %d", *capturedCounter["counter1"].Delta)
		}

		if *capturedCounter["counter2"].Delta != 30 {
			t.Errorf("Expected counter2 delta 30, got %d", *capturedCounter["counter2"].Delta)
		}

		// Проверяем gauge
		if len(capturedGauge) != 1 {
			t.Errorf("Expected 1 gauge, got %d", len(capturedGauge))
		}

		if capturedGauge[0].ID != "gauge1" {
			t.Errorf("Expected gauge1, got %s", capturedGauge[0].ID)
		}
	})

	t.Run("UpsertBatchError", func(t *testing.T) {
		storage := &MockStorage{
			UpsertBatchError: errors.New("storage error"),
		}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "test_gauge",
				MType: models.Gauge,
				Value: float64Ptr(123.45),
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metrics := []models.Metrics{
			{
				ID:    "gauge1",
				MType: models.Gauge,
				Value: float64Ptr(123.45),
			},
			{
				ID:    "counter1",
				MType: models.Counter,
				Delta: int64Ptr(42),
			},
		}
		body, _ := json.Marshal(metrics)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePostUpdates(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		if !storage.UpsertBatchCalled {
			t.Error("UpsertBatch was not called")
		}
	})
}

// TestHandleGetValue тестирует JSON обработчик для получения значений метрик.
func TestHandleGetValue(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("InvalidContentType", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/value", nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected %d, got %d", http.StatusUnsupportedMediaType, w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("EmptyMetricID", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "",
			MType: models.Gauge,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("InvalidMetricType", func(t *testing.T) {
		storage := &MockStorage{}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test",
			MType: "invalid",
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("MetricNotFound", func(t *testing.T) {
		storage := &MockStorage{
			GetValResult: models.Metrics{MType: ""},
		}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "not_found",
			MType: models.Gauge,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("MetricTypeMismatch", func(t *testing.T) {
		storage := &MockStorage{
			GetValResult: models.Metrics{
				MType: models.Counter,
				Delta: int64Ptr(42),
			},
		}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test",
			MType: models.Gauge, // Запрашиваем gauge, но хранится counter
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("SuccessGauge", func(t *testing.T) {
		expectedValue := 123.456
		storage := &MockStorage{
			GetValResult: models.Metrics{
				MType: models.Gauge,
				Value: &expectedValue,
			},
		}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_gauge",
			MType: models.Gauge,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		var response models.Metrics
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != "test_gauge" {
			t.Errorf("Expected ID test_gauge, got %s", response.ID)
		}

		if response.MType != models.Gauge {
			t.Errorf("Expected type gauge, got %s", response.MType)
		}

		if response.Value == nil || *response.Value != expectedValue {
			t.Errorf("Expected value %f, got %v", expectedValue, response.Value)
		}
	})

	t.Run("SuccessCounter", func(t *testing.T) {
		expectedDelta := int64(42)
		storage := &MockStorage{
			GetValResult: models.Metrics{
				MType: models.Counter,
				Delta: &expectedDelta,
			},
		}
		handler := NewMtrHandler(storage, logger)

		metric := models.Metrics{
			ID:    "test_counter",
			MType: models.Counter,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleGetValue(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Metrics
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Delta == nil || *response.Delta != expectedDelta {
			t.Errorf("Expected delta %d, got %v", expectedDelta, response.Delta)
		}
	})
}

// TestHandlePing тестирует обработчик проверки состояния хранилища.
func TestHandlePing(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("PingSuccess", func(t *testing.T) {
		storage := &MockStorage{
			PingError: nil,
		}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.HandlePing(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
		}

		if !storage.PingCalled {
			t.Error("Ping was not called")
		}
	})

	t.Run("PingError", func(t *testing.T) {
		storage := &MockStorage{
			PingError: errors.New("connection failed"),
		}
		handler := NewMtrHandler(storage, logger)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		handler.HandlePing(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if !storage.PingCalled {
			t.Error("Ping was not called")
		}
	})
}
