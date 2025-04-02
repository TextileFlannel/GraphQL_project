package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPPort    string
	StorageType string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
}

func LoadConfig() (*Config, error) {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		StorageType: strings.ToLower(getEnv("STORAGE_TYPE", "inmem")),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "links"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
