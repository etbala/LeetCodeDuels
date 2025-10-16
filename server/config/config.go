package config

// Package used to load configuration from environment variables

import "os"

type Config struct {
	GITHUB_CLIENT_ID      string
	GITHUB_CLIENT_SECRET  string
	PORT                  string
	DB_URL                string
	RDB_URL               string
	JWT_SECRET            string
	LOG_LEVEL             string // "debug", "info", "warn", "error", "fatal", "panic", "trace"
	SUBMISSION_VALIDATION bool
}

var appConfig *Config = nil

func InitConfig() (*Config, error) {
	if appConfig != nil {
		// Prevent re-loading config
		return appConfig, nil
	}
	var err error
	appConfig, err = loadConfig()
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}

func GetConfig() *Config {
	if appConfig == nil {
		panic("Config not initialized. Call InitConfig() first.")
	}
	return appConfig
}

// LoadConfig reads configuration from environment variables
func loadConfig() (*Config, error) {
	return &Config{
		GITHUB_CLIENT_ID:      os.Getenv("GH_CLIENT_ID"),
		GITHUB_CLIENT_SECRET:  os.Getenv("GH_CLIENT_SECRET"),
		PORT:                  getEnv("PORT", "8080"),
		DB_URL:                os.Getenv("DB_URL"),
		RDB_URL:               os.Getenv("RDB_URL"),
		JWT_SECRET:            os.Getenv("JWT_SECRET"),
		LOG_LEVEL:             getEnv("LOG_LEVEL", "debug"),
		SUBMISSION_VALIDATION: getEnv("SUBMISSION_VALIDATION", "enable") != "disable", // only disable if "disable"
	}, nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
