package util

import (
	"math/rand"
	"time"
)

// Global random source
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// Generates a random integer between min and max, inclusive
func RandInt(min, max int) int {
	return random.Intn(max-min+1) + min
}

// Generates a random alphanumeric string of the specified length
func RandAlphaNumericString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[random.Intn(len(charset))]
	}
	return string(b)
}
