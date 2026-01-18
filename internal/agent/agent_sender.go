// Package agent предоставляет функциональность для сбора метрик системы.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"

	"github.com/Valentin-Makurin/metrics/internal/common"
	models "github.com/Valentin-Makurin/metrics/internal/model"
	pb "github.com/Valentin-Makurin/metrics/internal/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetricsSender определяет интерфейс для отправки метрик на сервер.
type MetricsSender interface {
	// SendJSONRequestBatch отправляет пакет метрик на сервер.
	SendJSONRequestBatch(metric any) error
	// SendJSONRequestBatchgRPC отправляет пакет метрик на сервер по gRPC.
	SendJSONRequestBatchgRPC(metrics []models.Metrics) error
	// SendJSONRequest отправляет одиночную метрику на сервер.
	SendJSONRequest(metric models.Metrics) error
}

// HTTPSender реализует интерфейс MetricsSender для отправки метрик по HTTP.
// generate:reset
type HTTPSender struct {
	ctx        context.Context
	client     *http.Client
	protoAgent pb.MetricsClient
	PublicKey  *rsa.PublicKey
	baseURL    string
	KeyH       string
	localIP    string
}

// NewHTTPSender создает и возвращает новый экземпляр HTTPSender.
func NewHTTPSender(ctx context.Context, baseURL, ketH string, pub *rsa.PublicKey, protoAgent pb.MetricsClient) *HTTPSender {
	res := HTTPSender{
		ctx:        ctx,
		client:     &http.Client{},
		PublicKey:  pub,
		baseURL:    baseURL,
		KeyH:       ketH,
		localIP:    getLocalIP(),
		protoAgent: protoAgent,
	}
	return &res
}

// SendJSONRequest отправляет одиночную метрику на сервер.
// Метод выполняет сериализацию метрики в JSON, сжатие gzip,
// добавление HMAC подписи если задан ключ, и отправку POST-запросом
// на эндпоинт /update/.
func (s *HTTPSender) SendJSONRequest(metric models.Metrics) error {
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = gz.Write(jsonData)
	if err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	var encryptedData bytes.Buffer

	if s.PublicKey != nil {
		encryptedBytes, err := rsa.EncryptPKCS1v15(
			rand.Reader,
			s.PublicKey,
			compressedData.Bytes(),
		)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
		encryptedData.Write(encryptedBytes)
	} else {
		encryptedData = compressedData
	}

	url := "http://" + s.baseURL + "/update/"
	req, err := http.NewRequest("POST", url, &encryptedData) //compressedData
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	if s.KeyH != "" {
		hashData := common.HashVal(s.KeyH, jsonData)
		req.Header.Set("HashSHA256", hashData)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	if s.localIP != "" {
		req.Header.Set("X-Real-IP", s.localIP)
	}

	if s.PublicKey != nil {
		req.Header.Set("X-Encrypted", "RSA-OAEP") // Флаг шифрования
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		fmt.Print("gzip ok ")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SendJSONRequestBatchgRPC отправляет пакет метрик на сервер по gRPC.
func (s *HTTPSender) SendJSONRequestBatchgRPC(metrics []models.Metrics) error {
	metricsPb := []*pb.Metric{}

	for _, val := range metrics {

		tmpMtr := pb.Metric_builder{
			Id: val.ID,
		}

		if val.MType == "gauge" {
			tmpMtr.Type = pb.Metric_GAUGE
			tmpMtr.Value = *val.Value
		} else {
			tmpMtr.Type = pb.Metric_COUNTER
			tmpMtr.Delta = *val.Delta
		}

		metricsPb = append(metricsPb, tmpMtr.Build())
	}

	request := pb.UpdateMetricsRequest_builder{
		Metrics: metricsPb,
	}.Build()

	ctx := context.Background()
	if s.localIP != "" {
		md := metadata.New(map[string]string{"X-Real-IP": s.localIP})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	_, err := s.protoAgent.UpdateMetrics(ctx, request)
	if err != nil {
		// Пробуем получить статус gRPC ошибки
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("gRPC ошибка: код=%v, сообщение=%v", st.Code(), st.Message())
		}
		return fmt.Errorf("ошибка при вызове UpdateMetrics: %w", err)
	}

	return nil
}

// SendJSONRequestBatch отправляет пакет метрик на сервер.
// Метод выполняет сериализацию данных в JSON, сжатие gzip
// и отправку POST-запросом на эндпоинт /updates/.
func (s *HTTPSender) SendJSONRequestBatch(metric any) error {
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = gz.Write(jsonData)
	if err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	var encryptedData bytes.Buffer

	if s.PublicKey != nil {
		encryptedBytes, err := rsa.EncryptPKCS1v15(
			rand.Reader,
			s.PublicKey,
			compressedData.Bytes(),
		)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
		encryptedData.Write(encryptedBytes)
	} else {
		encryptedData = compressedData
	}

	url := "http://" + s.baseURL + "/updates/"
	req, err := http.NewRequest("POST", url, &encryptedData) //compressedData
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	if s.localIP != "" {
		req.Header.Set("X-Real-IP", s.localIP)
	}

	if s.PublicKey != nil {
		req.Header.Set("X-Encrypted", "RSA-OAEP") // Флаг шифрования
	}

	resp, err := s.runReq(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		fmt.Print("gzip ok ")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *HTTPSender) runReq(req *http.Request) (*http.Response, error) {

	return common.RetryOperation(func() (*http.Response, error) {
		return s.client.Do(req)
	},
		isTemporaryError)
}

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		switch syscallErr {
		case syscall.ECONNREFUSED, // Connection refused
			syscall.ECONNRESET,   // Connection reset by peer
			syscall.ETIMEDOUT,    // Connection timed out
			syscall.EHOSTDOWN,    // Host is down
			syscall.EHOSTUNREACH, // Host is unreachable
			syscall.ENETDOWN,     // Network is down
			syscall.ENETUNREACH,  // Network is unreachable
			syscall.EWOULDBLOCK:  // Operation would block
			return true
		}
	}
	return false
}

// getLocalIP получает локальный IP адрес
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
