// Package agent предоставляет функциональность для сбора метрик системы.
package agent

import "sync"

// MetricsStorage определяет интерфейс для хранилища метрик.
type MetricsStorage interface {
	// SetGauge устанавливает значение метрики типа gauge.
	SetGauge(key string, value any)
	// AddCounter увеличивает значение метрики типа counter.
	AddCounter(key string, value uint)
	// GetGauge возвращает значение метрики типа gauge по ключу.
	GetGauge(key string) any
	// GetCounter возвращает значение метрики типа counter по ключу.
	GetCounter(key string) uint
	// GetAllGauges возвращает все метрики типа gauge.
	GetAllGauges() map[string]any
	// GetAllCounters возвращает все метрики типа counter.
	GetAllCounters() map[string]uint
}

// MemStorage реализует интерфейс MetricsStorage для хранения метрик в памяти.
type MemStorage struct {
	gaugesStorage  map[string]interface{}
	counterStorage map[string]uint
	mu             sync.RWMutex
}

// NewMemStorage создает и возвращает новый экземпляр MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gaugesStorage:  make(map[string]interface{}),
		counterStorage: make(map[string]uint),
	}
}

// SetGauge устанавливает значение метрики типа gauge.
func (s *MemStorage) SetGauge(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gaugesStorage[key] = value
}

// AddCounter увеличивает значение метрики типа counter.
func (s *MemStorage) AddCounter(key string, value uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counterStorage[key] += value
}

// GetGauge возвращает значение метрики типа gauge по ключу.
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

// GetCounter возвращает значение метрики типа counter по ключу.
func (s *MemStorage) GetCounter(key string) uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val := s.counterStorage[key]
	return val
}

// GetAllGauges возвращает копию всех метрик типа gauge.
func (s *MemStorage) GetAllGauges() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any)
	for k, v := range s.gaugesStorage {
		result[k] = v
	}
	return result
}

// GetAllCounters возвращает копию всех метрик типа counter.
func (s *MemStorage) GetAllCounters() map[string]uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]uint)
	for k, v := range s.counterStorage {
		result[k] = v
	}
	return result
}
