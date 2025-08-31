package middleware

import (
	"net/http"
	"time"

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
