package config

import "os"

// Used to access secrets
type Config struct {
	GITHUB_CLIENT_ID     string
	GITHUB_CLIENT_SECRET string
	DB_URL               string
	RDB_URL              string
	JWT_SECRET           string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() (*Config, error) {
	return &Config{
		GITHUB_CLIENT_ID:     os.Getenv("GH_CLIENT_ID"),
		GITHUB_CLIENT_SECRET: os.Getenv("GH_CLIENT_SECRET"),
		DB_URL:               os.Getenv("DB_URL"),
		RDB_URL:              os.Getenv("RDB_URL"),
		JWT_SECRET:           os.Getenv("JWT_SECRET"),
	}, nil
}
