package common

import (
	"fmt"
	"time"
)

func RetryOperation[T any](
	operation func() (T, error),
	isTemporary func(error) bool,
	// maxAttempts int,
	// intervals []time.Duration,
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
