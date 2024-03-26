package config

import "os"

type Config struct {
	DatabaseURL string
}

// LoadConfig reads configuration from environment variables or configuration files.
func LoadConfig() (*Config, error) {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}, nil
}
