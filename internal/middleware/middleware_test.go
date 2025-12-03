package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Valentin-Makurin/metrics/internal/common"
	"go.uber.org/zap/zaptest"
)

func TestLoggerMiddleware(t *testing.T) {
	// Создаем тестовый логгер
	logger := zaptest.NewLogger(t).Sugar()
	defer logger.Sync()

	// Хендлер, который будет обернут мидлварой
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	loggerMiddleware := LoggerMiddleware(logger)
	wrappedHandler := loggerMiddleware(testHandler)

	t.Run("LogsRequestInfo", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("LogsDifferentStatus", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		wrapped := loggerMiddleware(handler)
		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("LogsPostRequest", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Created"))
		})

		wrapped := loggerMiddleware(handler)
		req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader("data"))
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestGzipMiddleware(t *testing.T) {
	// Тестовый хендлер
	responseBody := "Hello, Gzipped World!"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(responseBody))
	})

	// Создаем мидлвару
	gzipMiddleware := GzipMiddleware(testHandler)

	t.Run("NoCompressionWhenNotAccepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		gzipMiddleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Encoding") != "" {
			t.Errorf("Expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
		}

		if w.Body.String() != responseBody {
			t.Errorf("Expected body '%s', got '%s'", responseBody, w.Body.String())
		}
	})

	t.Run("CompressesResponseWhenAccepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		gzipMiddleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding gzip, got %s", w.Header().Get("Content-Encoding"))
		}

		// Проверяем, что тело сжато (не равно оригиналу)
		body := w.Body.Bytes()
		if string(body) == responseBody {
			t.Error("Body should be compressed")
		}
	})

	t.Run("ErrorOnInvalidGzipRequestBody", func(t *testing.T) {
		// Некорректные gzip данные
		invalidGzipData := []byte("not a valid gzip")

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(invalidGzipData))
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()

		gzipMiddleware.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestHashMiddleware(t *testing.T) {
	// Тестовый хендлер
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})

	// Тестовый ключ
	testKey := "test-secret-key"

	t.Run("NoKeyNoCheck", func(t *testing.T) {
		// Без ключа мидлвара должна просто пропустить запрос
		hashMiddleware := HashMiddleware("")
		wrappedHandler := hashMiddleware(testHandler)

		body := `{"test": "data"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("HashSHA256") != "" {
			t.Errorf("Expected no HashSHA256 header, got %s", w.Header().Get("HashSHA256"))
		}
	})

	t.Run("ValidHash", func(t *testing.T) {
		hashMiddleware := HashMiddleware(testKey)
		wrappedHandler := hashMiddleware(testHandler)

		body := `{"id": "test", "type": "gauge", "value": 123.45}`
		bodyBytes := []byte(body)

		// Вычисляем правильный хеш
		validHash := common.HashVal(testKey, bodyBytes)

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("HashSHA256", validHash)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Проверяем, что в ответе тоже есть хеш
		if w.Header().Get("HashSHA256") == "" {
			t.Error("Expected HashSHA256 header in response")
		}
	})

	t.Run("InvalidHash", func(t *testing.T) {
		hashMiddleware := HashMiddleware(testKey)
		wrappedHandler := hashMiddleware(testHandler)

		body := `{"id": "test", "type": "gauge", "value": 123.45}`

		// Неправильный хеш
		invalidHash := "wrong-hash-value"

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("HashSHA256", invalidHash)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("HashPresentButNoKey", func(t *testing.T) {
		// Если ключ пустой, но хеш передан - игнорируем
		hashMiddleware := HashMiddleware("")
		wrappedHandler := hashMiddleware(testHandler)

		body := `{"test": "data"}`

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("HashSHA256", "some-hash")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("ErrorReadingBody", func(t *testing.T) {
		hashMiddleware := HashMiddleware(testKey)
		wrappedHandler := hashMiddleware(testHandler)

		// Создаем request с body, которое вернет ошибку при чтении
		req := httptest.NewRequest(http.MethodPost, "/test", errorReader{})
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("ResponseHashAdded", func(t *testing.T) {
		hashMiddleware := HashMiddleware(testKey)
		wrappedHandler := hashMiddleware(testHandler)

		body := `{"test": "response hash"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		responseHash := w.Header().Get("HashSHA256")
		if responseHash == "" {
			t.Error("Expected HashSHA256 header in response")
		}

		// Проверяем, что хеш корректный (для тела "hello")
		expectedHash := common.HashVal(testKey, []byte("hello"))
		if responseHash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, responseHash)
		}
	})
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (errorReader) Close() error {
	return nil
}
