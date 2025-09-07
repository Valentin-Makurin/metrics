package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/compress"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// LoggerMiddleware - middleware для логирования запросов и ответов
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
