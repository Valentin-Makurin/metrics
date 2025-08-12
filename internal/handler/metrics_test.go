package handler

import (
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Valentin-Makurin/metrics/internal/db"
	models "github.com/Valentin-Makurin/metrics/internal/model"
)

func TestHandlePost_MethodNotAllowed(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

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

func TestHandlePost_InvalidPath(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

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

func TestHandlePost_EmptyMetricName(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge//123", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandlePost_InvalidMetricType(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	req := httptest.NewRequest(http.MethodPost, "/update/invalid/test/123", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_InvalidGaugeValue(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/invalid", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_InvalidCounterValue(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/test/invalid", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
func TestHandleGet(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	gaugeValue := 123.456
	counterValue := int64(42)
	storage.SetVal("test_gauge", models.Metrics{MType: models.Gauge, Value: &gaugeValue})
	storage.SetVal("test_counter", models.Metrics{MType: models.Counter, Delta: &counterValue})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid gauge metric",
			path:           "/value/gauge/test_gauge",
			expectedStatus: http.StatusOK,
			expectedBody:   "123.456",
		},
		{
			name:           "valid counter metric",
			path:           "/value/counter/test_counter",
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:           "invalid path format",
			path:           "/value/gauge",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid metric type",
			path:           "/value/invalid/test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent metric",
			path:           "/value/gauge/non_existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty metric name",
			path:           "/value/gauge/",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.HandleGet(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && strings.TrimSpace(w.Body.String()) != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleRoot(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

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
