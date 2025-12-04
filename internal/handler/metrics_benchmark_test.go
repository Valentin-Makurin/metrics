// Package handler предоставляет HTTP-обработчики для работы с метриками.
package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"go.uber.org/zap"
)

// mockStorage реализует мок-хранилище для тестирования обработчиков без реальной БД.
type mockStorage struct{}

func (m *mockStorage) SetVal(key string, mtr models.Metrics) {}
func (m *mockStorage) AddVal(key string, mtr models.Metrics) {}
func (m *mockStorage) GetVal(key string) models.Metrics {
	if key == "existing_metric" {
		val := 123.45
		return models.Metrics{
			ID:    key,
			MType: models.Gauge,
			Value: &val,
		}
	}
	return models.Metrics{}
}
func (m *mockStorage) GetAllVal() map[string]string {
	return map[string]string{
		"metric1": "value1",
		"metric2": "value2",
	}
}
func (m *mockStorage) UpsertBatch(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error {
	return nil
}
func (m *mockStorage) Ping() error {
	return nil
}

// BenchmarkHandlePost_Gauge измеряет производительность обработчика POST запросов для gauge метрик через URL.
func BenchmarkHandlePost_Gauge(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/update/gauge/test_metric/123.45", nil)
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)
	}
}

// BenchmarkHandlePost_Counter измеряет производительность обработчика POST запросов для counter метрик через URL.
func BenchmarkHandlePost_Counter(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/update/counter/test_metric/100", nil)
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)
	}
}

// BenchmarkHandlePostUpdate_Gauge измеряет производительность обработчика POST запросов для gauge метрик через JSON API.
func BenchmarkHandlePostUpdate_Gauge(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(metric)
		req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandlePostUpdate(w, req)
	}
}

// BenchmarkHandlePostUpdate_Counter измеряет производительность обработчика POST запросов для counter метрик через JSON API.
func BenchmarkHandlePostUpdate_Counter(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Counter,
		Delta: func() *int64 { v := int64(100); return &v }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(metric)
		req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandlePostUpdate(w, req)
	}
}

// BenchmarkHandleGet измеряет производительность обработчика GET запросов для получения метрик через URL.
func BenchmarkHandleGet(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/value/gauge/existing_metric", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)
	}
}

// BenchmarkHandleGetValue измеряет производительность обработчика POST запросов для получения метрик через JSON API.
func BenchmarkHandleGetValue(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	metric := models.Metrics{
		ID:    "existing_metric",
		MType: models.Gauge,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(metric)
		req := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandleGetValue(w, req)
	}
}

// BenchmarkHandleRoot измеряет производительность обработчика корневого пути, возвращающего все метрики.
func BenchmarkHandleRoot(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.HandleRoot(w, req)
	}
}

// BenchmarkHandlePing измеряет производительность обработчика проверки доступности хранилища.
func BenchmarkHandlePing(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()
		handler.HandlePing(w, req)
	}
}

// BenchmarkHandlePostUpdates измеряет производительность обработчика пакетного обновления метрик.
func BenchmarkHandlePostUpdates(b *testing.B) {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	handler := NewMtrHandler(&mockStorage{}, sugar)

	metrics := []models.Metrics{
		{
			ID:    "gauge1",
			MType: models.Gauge,
			Value: func() *float64 { v := 100.0; return &v }(),
		},
		{
			ID:    "counter1",
			MType: models.Counter,
			Delta: func() *int64 { v := int64(10); return &v }(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(metrics)
		req := httptest.NewRequest("POST", "/updates/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandlePostUpdates(w, req)
	}
}
