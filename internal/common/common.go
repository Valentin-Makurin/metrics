// Package common предоставляет общие утилиты и вспомогательные функции,
// используемые в различных частях приложения.
package common

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
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
	pool *sync.Pool
}

// New конструктор для Pool
func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
	}
}

// Get возвращает объект из пула или создаёт новый, если пул пуст
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put - помещает объект в пул
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}

func FirstPrint(w io.Writer, buildVersion, buildDate, buildCommit string) {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Fprintf(w, "Build version: %s\n", buildVersion)
	fmt.Fprintf(w, "Build date: %s\n", buildDate)
	fmt.Fprintf(w, "Build commit: %s\n", buildCommit)
}

// ReadPrivateKey читает приватный ключ по указанному адресу
func ReadPrivateKey(priv string) (*rsa.PrivateKey, error) {
	var privKey *rsa.PrivateKey
	if priv != "" {
		keyBytes, err := os.ReadFile(priv)
		if err != nil {
			return nil, err
		}

		// Декодируем PEM блок
		block, _ := pem.Decode(keyBytes)
		if block == nil {
			return nil, err
		}

		// Парсим публичный ключ
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return privKey, nil
}

func ReadPubKey(pub string) (*rsa.PublicKey, error) {
	var rsaPubKey *rsa.PublicKey
	var ok bool
	if pub != "" {
		keyBytes, err := os.ReadFile(pub)
		if err != nil {
			return nil, err
		}

		// Декодируем PEM блок
		block, _ := pem.Decode(keyBytes)
		if block == nil {
			return nil, err
		}

		// Парсим публичный ключ
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		// Приводим к типу *rsa.PublicKey
		rsaPubKey, ok = pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, err
		}
	}
	return rsaPubKey, nil

}
