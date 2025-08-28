package db

import (
	"fmt"
	"strconv"
	"sync"

	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type MtrStorage struct {
	mtr map[string]models.Metrics
	mu  sync.RWMutex
}

func NewStorage() *MtrStorage {
	return &MtrStorage{
		mtr: make(map[string]models.Metrics),
	}
}

func (s *MtrStorage) SetVal(key string, mtr models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mtr == nil {
		s.mtr = make(map[string]models.Metrics)
	}
	s.mtr[key] = mtr
}

func (s *MtrStorage) AddVal(key string, mtr models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.mtr[key]
	if val.MType != "" && val.Delta != nil && mtr.Delta != nil {
		*mtr.Delta += *val.Delta
	}
	s.mtr[key] = mtr
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
