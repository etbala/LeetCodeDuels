package services

import (
	"leetcodeduels/config"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MainTest(m *testing.M) {
	os.Setenv("JWT_SECRET", "0")

	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %s", err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestJWT(t *testing.T) {
	jwt, err := GenerateJWT(999)
	assert.NoError(t, err)

	claims, err := ValidateJWT(jwt)
	assert.NoError(t, err)

	assert.Equal(t, int64(999), claims.UserID)
}
