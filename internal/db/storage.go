package db

import (
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
	s.mtr[key] = mtr
}
