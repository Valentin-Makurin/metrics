package db

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type MtrStorage struct {
	file          *os.File
	writer        *bufio.Writer
	encoder       *json.Encoder
	mtr           map[string]models.Metrics
	mu            sync.RWMutex
	filePath      string
	storeInterval int
	restore       bool
	saveChan      chan struct{}
	logger        *zap.SugaredLogger
	pool          *pgxpool.Pool
}

func NewStorage(filePath string, storeInterval int, restore bool, logger *zap.SugaredLogger) *MtrStorage {
	storage := &MtrStorage{
		mtr:           make(map[string]models.Metrics),
		filePath:      filePath,
		storeInterval: storeInterval,
		restore:       restore,
		saveChan:      make(chan struct{}, 1),
		logger:        logger,
	}

	return storage
}

func (s *MtrStorage) PrepareFile() {
	if s.restore && s.filePath != "" {
		if err := s.LoadFromFile(); err != nil {
			s.logger.Error("Failed to load metrics from file: %v", err)
		}
	}

	if s.filePath != "" {
		file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			s.file = file
			s.writer = bufio.NewWriter(file)
			s.encoder = json.NewEncoder(s.writer)
		}
	}
}

func (s *MtrStorage) StartTicker() {
	if s.storeInterval > 0 && s.filePath != "" {
		go s.periodicSave()
	}
}

func (s *MtrStorage) SaveToFile() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	tmpFile, err := os.CreateTemp("", "metrics_*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	writer := bufio.NewWriter(tmpFile)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	metrics := make([]models.Metrics, 0, len(s.mtr))
	for key, metric := range s.mtr {
		metricCopy := metric
		metricCopy.ID = key
		metrics = append(metrics, metricCopy)
	}

	if err := encoder.Encode(metrics); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), s.filePath)
}

func (s *MtrStorage) LoadFromFile() error {
	if s.filePath == "" {
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var metrics []models.Metrics
	if err := decoder.Decode(&metrics); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, metric := range metrics {
		s.mtr[metric.ID] = metric
	}

	return nil
}

func (s *MtrStorage) periodicSave() {
	ticker := time.NewTicker(time.Duration(s.storeInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.SaveToFile(); err != nil {
				s.logger.Error("Failed to save metrics: %v\n", err)
			}
		case <-s.saveChan:
			return
		}
	}
}

func (s *MtrStorage) Close() {
	close(s.saveChan)
	if err := s.SaveToFile(); err != nil {
		s.logger.Error("Final save failed: %v\n", err)
	}
}

func (s *MtrStorage) SaveByEvent() {
	if s.storeInterval == 0 && s.filePath != "" {
		if err := s.SaveToFile(); err != nil {
			s.logger.Error("Sync save failed: %v\n", err)
		}
	}
}

func (s *MtrStorage) SetVal(key string, mtr models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mtr == nil {
		s.mtr = make(map[string]models.Metrics)
	}
	s.mtr[key] = mtr

	s.SaveByEvent()
}

func (s *MtrStorage) AddVal(key string, mtr models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.mtr[key]
	if val.MType != "" && val.Delta != nil && mtr.Delta != nil {
		*mtr.Delta += *val.Delta
	}
	s.mtr[key] = mtr

	s.SaveByEvent()
}

func (s *MtrStorage) GetVal(key string) models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mtr[key]
}

func (s *MtrStorage) GetAllVal() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string]string)
	for key, val := range s.mtr {
		str := ""
		if val.Value != nil {
			str = fmt.Sprintf("%.4f", *val.Value)
		} else if val.Delta != nil {
			str = strconv.FormatInt(*val.Delta, 10)
		}
		res[key] = str
	}
	return res
}

func (s *MtrStorage) Ping() error {
	return nil
}
