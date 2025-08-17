package agent

import "sync"

type MetricsStorage interface {
	SetGauge(key string, value any)
	AddCounter(key string, value uint)
	GetGauge(key string) any
	GetCounter(key string) uint
	GetAllGauges() map[string]any
	GetAllCounters() map[string]uint
}

type MemStorage struct {
	gaugesStorage  map[string]interface{}
	counterStorage map[string]uint
	mu             sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gaugesStorage:  make(map[string]interface{}),
		counterStorage: make(map[string]uint),
	}
}

func (s *MemStorage) SetGauge(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gaugesStorage[key] = value
}

func (s *MemStorage) AddCounter(key string, value uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counterStorage[key] += value
}

func (s *MemStorage) GetGauge(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.gaugesStorage[key]
	if !exists {
		return 0
	}

	if floatVal, ok := val.(float64); ok {
		return floatVal
	}
	return 0
}

func (s *MemStorage) GetCounter(key string) uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val := s.counterStorage[key]
	return val
}

func (s *MemStorage) GetAllGauges() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any)
	for k, v := range s.gaugesStorage {
		if floatVal, ok := v.(float64); ok {
			result[k] = floatVal
		}
	}
	return result
}

func (s *MemStorage) GetAllCounters() map[string]uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]uint)
	for k, v := range s.counterStorage {
		result[k] = v
	}
	return result
}
