// Package common предоставляет общие утилиты и вспомогательные функции,
// используемые в различных частях приложения.
package common

import (
	"fmt"
	"sync"
	"time"
)

// RetryOperation выполняет операцию с повторными попытками при временных ошибках.
// Поддерживает экспоненциальную backoff стратегию с фиксированными интервалами.
// Функция является обобщенной (generic) и может работать с любым типом результата.
func RetryOperation[T any](
	operation func() (T, error),
	isTemporary func(error) bool,
) (T, error) {
	var zero T
	var lastErr error

	intervals := map[int]time.Duration{0: 1 * time.Second, 1: 3 * time.Second, 2: 5 * time.Second}

	for i := range 3 {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isTemporary(err) {
			return zero, fmt.Errorf("permanent error: %v", err)
		}

		if i < len(intervals) {
			time.Sleep(intervals[i])
		}
	}

	return zero, fmt.Errorf("max attempts exceeded, last error: %v", lastErr)
}

type Resetter interface {
	Reset()
}

// Pool пул объектов с дженерик-параметром T
type Pool[T Resetter] struct {
	pool    []T
	mu      sync.Mutex
	factory func() T
}

// New конструктор для Pool
func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool:    make([]T, 0),
		factory: factory,
	}
}

// Get возвращает объект из пула или создаёт новый, если пул пуст
func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pool) > 0 {
		obj := p.pool[len(p.pool)-1]
		p.pool = p.pool[:len(p.pool)-1]
		return obj
	}

	return p.factory()
}

// Put - помещает объект в пул
func (p *Pool[T]) Put(obj T) {
	obj.Reset()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.pool = append(p.pool, obj)
}
