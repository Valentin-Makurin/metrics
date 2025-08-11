package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Valentin-Makurin/metrics/internal/db"
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

func TestHandlePost_InvalidContentType(t *testing.T) {
	storage := &db.MtrStorage{}
	handler := NewMtrHandler(storage)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlePost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
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
