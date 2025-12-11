// Package compress предоставляет обертки для сжатия и распаковки HTTP-трафика с использованием gzip.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

// compressWriter реализует http.ResponseWriter с поддержкой gzip-сжатия.
// Автоматически сжимает данные перед отправкой клиенту, если тот поддерживает gzip.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter создает новый compressWriter для сжатия HTTP-ответов.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает HTTP-заголовки оригинального ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write сжимает данные с использованием gzip перед записью в ResponseWriter.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает код статуса HTTP ответа.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip writer, завершая сжатие и запись данных.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует io.ReadCloser с поддержкой gzip-распаковки.
// Автоматически распаковывает сжатые gzip данные из HTTP-запросов.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает новый compressReader для распаковки gzip-сжатых HTTP-запросов.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и распаковывает данные из gzip потока.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает как оригинальный reader, так и gzip reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
