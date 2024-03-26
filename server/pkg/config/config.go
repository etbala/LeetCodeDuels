package config

import "os"

type Config struct {
	DB_URL  string
	DB_USER string
	DB_PASS string
}

// LoadConfig reads configuration from environment variables or configuration files.
func LoadConfig() (*Config, error) {
	return &Config{
		DB_URL:  os.Getenv("DATABASE_URL"),
		DB_USER: os.Getenv("DATABASE_USERNAME"),
		DB_PASS: os.Getenv("DATABASE_PASSWORD"),
	}, nil
}
