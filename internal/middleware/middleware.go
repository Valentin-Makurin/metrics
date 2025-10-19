package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Valentin-Makurin/metrics/internal/common"
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

// func HashMiddleware2(next http.Handler, KeyH string) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		bodyBytes, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		suspect := r.Header.Get("HashSHA256")

// 		if !common.CheckHash(KeyH, suspect, bodyBytes) && suspect != "" {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }
