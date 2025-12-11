// Package common предоставляет общие утилиты и вспомогательные функции,
// используемые в различных частях приложения.
package common

import (
	"fmt"
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
