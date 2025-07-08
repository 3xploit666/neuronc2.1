package config

import (
	"os"
	"time"
)

type Config struct {
	ServerName     string
	ServerVersion  string
	DatabasePath   string
	ScreenshotDir  string
	Port           string
	CommandTimeout time.Duration
}

func Load() (*Config, error) {
	return &Config{
		ServerName:     getEnv("SERVER_NAME", "NeuronC2"),
		ServerVersion:  getEnv("SERVER_VERSION", "1.0.0"),
		DatabasePath:   getEnv("DATABASE_PATH", "./c2_database.db"),
		ScreenshotDir:  getEnv("SCREENSHOT_DIR", "screenshots"),
		Port:           getEnv("PORT", ":8080"),
		CommandTimeout: 30 * time.Second,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
