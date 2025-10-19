package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashVal(key string, val []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(val)
	return hex.EncodeToString(h.Sum(nil))
}

func CheckHash(key, suspect string, value []byte) bool {
	return hmac.Equal([]byte(HashVal(key, value)), []byte(suspect))
}
