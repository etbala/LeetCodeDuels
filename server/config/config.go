package config

import "os"

type Config struct {
	GITHUB_CLIENT_ID     string
	GITHUB_CLIENT_SECRET string
	DB_URL               string
	JWT_SECRET           string
}

// LoadConfig reads configuration from environment variables or configuration files.
func LoadConfig() (*Config, error) {
	return &Config{
		GITHUB_CLIENT_ID:     os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_CLIENT_SECRET: os.Getenv("GITHUB_CLIENT_SECRET"),
		DB_URL:               os.Getenv("DB_URL"),
		JWT_SECRET:           os.Getenv("JWT_SECRET"),
	}, nil
}