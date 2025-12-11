// Package common предоставляет общие утилиты и вспомогательные функции,
// используемые в различных частях приложения.
package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HashVal вычисляет HMAC-SHA256 хеш для переданных данных с использованием ключа.
func HashVal(key string, val []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(val)
	return hex.EncodeToString(h.Sum(nil))
}

// CheckHash проверяет соответствие HMAC-подписи данных.
func CheckHash(key, suspect string, value []byte) bool {
	return hmac.Equal([]byte(HashVal(key, value)), []byte(suspect))
}
