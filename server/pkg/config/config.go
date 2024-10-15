package config

import "os"

type Config struct {
	GITHUB_CLIENT_ID     string
	GITHUB_CLIENT_SECRET string
	GITHUB_REDIRECT_URI  string
	DB_URL               string
}

// LoadConfig reads configuration from environment variables or configuration files.
func LoadConfig() (*Config, error) {
	return &Config{
		GITHUB_CLIENT_ID:     os.Getenv("GITHUB_CLIENT_ID"),
		GITHUB_CLIENT_SECRET: os.Getenv("GITHUB_CLIENT_SECRET"),
		GITHUB_REDIRECT_URI:  os.Getenv("GITHUB_REDIRECT_URI"),
		DB_URL:               os.Getenv("DB_URL"),
	}, nil
}
