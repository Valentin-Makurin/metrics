// Package middleware предоставляет HTTP middleware для обработки запросов.
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/compress"
	models "github.com/Valentin-Makurin/metrics/internal/model"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// LoggerMiddleware создает middleware для логирования HTTP запросов и ответов.
func LoggerMiddleware(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			logger.Infow("HTTP request",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", duration.String(),
				"status", ww.Status(),
				"size", ww.BytesWritten(),
			)
		})
	}
}

// GzipMiddleware создает middleware для сжатия gzip HTTP трафика.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := compress.NewCompressWriter(w)
			ow = cw
			w.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {

			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

// HashMiddleware создает middleware для проверки HMAC подписей запросов.
func HashMiddleware(KeyH string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			suspect := r.Header.Get("HashSHA256")

			if KeyH != "" && suspect != "" && !common.CheckHash(KeyH, suspect, bodyBytes) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body.Close()

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if KeyH != "" {
				hash := common.HashVal(KeyH, []byte("hello"))
				w.Header().Set("HashSHA256", hash)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuditMiddleware создает middleware для аудита операций с метриками.
func AuditMiddleware(client *http.Client, logger *zap.SugaredLogger, auditFile, auditURL string) func(next http.Handler) http.Handler {
	var muWrite sync.Mutex
	var muWSend sync.Mutex
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" &&
				(auditFile != "" || auditURL != "") &&
				(r.URL.Path == "/update/" || r.URL.Path == "/updates/") {

				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Infow("audit read body error", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// prepare event
				event := models.Event{}
				err = prepareEvent(&event, bodyBytes, r.URL.Path, r.RemoteAddr)
				if err != nil {
					logger.Infow("audit prepare event error", "err", err)
				}

				eventData, err := json.Marshal(event)
				if err != nil {
					logger.Infow("Error marshaling audit message", "err", err)
				}

				// send event
				if auditURL != "" {
					go sendAuditMessage(auditURL, eventData, &muWSend, client, logger)
				}

				// save event
				if auditFile != "" {
					go writeAuditMessage(auditFile, eventData, &muWrite, logger)
				}

			}
			next.ServeHTTP(w, r)
		})
	}
}

// prepareEvent подготавливает событие аудита на основе тела запроса.
func prepareEvent(msg *models.Event, body []byte, RPath, RAddr string) error {
	switch RPath {
	case "/update/":
		var metric models.Metrics
		if err := json.Unmarshal(body, &metric); err != nil {
			return err
		}
		msg.Metrics = append(msg.Metrics, metric.ID)
	case "/updates/":
		var metrics []models.Metrics
		if err := json.Unmarshal(body, &metrics); err != nil {
			return err
		}
		for _, metric := range metrics {
			msg.Metrics = append(msg.Metrics, metric.ID)
		}
	}

	msg.TS = time.Now().Unix()
	msg.IPAddress = RAddr
	return nil
}

// writeAuditMessage записывает событие аудита в файл.
func writeAuditMessage(filePath string, msg []byte, mu *sync.Mutex, logger *zap.SugaredLogger) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Infow("audit open file error", "err", err)
	}
	defer file.Close()

	if _, err := file.WriteString(string(msg) + "\n"); err != nil {
		logger.Infow("audit save updates to file error", "err", err)
	}
}

// sendAuditMessage отправляет событие аудита на внешний сервер.
func sendAuditMessage(auditURL string, msg []byte, mu *sync.Mutex, client *http.Client, logger *zap.SugaredLogger) {
	mu.Lock()
	defer mu.Unlock()

	bodyReader := bytes.NewBuffer(msg)

	req, err := http.NewRequest("POST", auditURL, bodyReader)
	if err != nil {
		logger.Infow("audit prepare request error", "err", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Infow("audit send request error", "err", err)
		return
	}
	resp.Body.Close()
}
