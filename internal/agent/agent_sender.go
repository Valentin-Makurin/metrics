package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/Valentin-Makurin/metrics/internal/common"
	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type MetricsSender interface {
	SendJSONRequestBatch(metric any) error
	SendJSONRequest(metric models.Metrics) error
}

type HTTPSender struct {
	client  *http.Client
	baseURL string
	KeyH    string
}

func NewHTTPSender(baseURL, KetH string) *HTTPSender {
	return &HTTPSender{
		client:  &http.Client{},
		baseURL: baseURL,
		KeyH:    KetH,
	}
}

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

	url := "http://" + s.baseURL + "/update/"
	req, err := http.NewRequest("POST", url, &compressedData)
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

	url := "http://" + s.baseURL + "/updates/"
	req, err := http.NewRequest("POST", url, &compressedData)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
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
