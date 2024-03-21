package util

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Hash(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func GenRandString(length uint64) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randBytes := make([]byte, length)
	for i := range randBytes {
		randBytes[i] = charset[r.Intn(len(charset))]
	}
	return string(randBytes)
}
