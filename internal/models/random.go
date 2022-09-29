package models

import (
	"crypto/rand"
	"encoding/hex"
)

// NewRandomString generates a random string.
// It panics if the source of randomness fails.
func GenerateRandomBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
