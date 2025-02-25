package config

import (
	"os"
)

type Config struct {
	ServerAddress string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
}

func LoadConfig() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "Bootje@12"),
		DBName:        getEnv("DB_NAME", "odomosml"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
