package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config bevat alle configuratie-instellingen voor de applicatie
type Config struct {
	// Server configuratie
	ServerAddress string
	Environment   string // "development", "production", "test"

	// Database configuratie
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DropTables bool // Alleen true in development!

	// JWT configuratie
	JWTSecret          string
	JWTExpirationHours int

	// Logging configuratie
	LogLevel string // "debug", "info", "warn", "error"
}

// LoadConfig laadt configuratie uit environment variables
func LoadConfig() *Config {
	environment := getEnv("APP_ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// Waarschuwing voor default JWT secret in productie
	if environment == "production" && jwtSecret == "your-secret-key" {
		log.Println("WAARSCHUWING: Default JWT secret wordt gebruikt in productie! Dit is onveilig.")
		log.Println("Stel een sterke JWT_SECRET in via environment variables.")
	}

	// Parse JWT expiration hours
	jwtExpirationHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExpirationHours = 24 // Default als parsing mislukt
	}

	// Parse drop tables boolean
	dropTables := getEnvBool("DB_DROP_TABLES", false)
	if environment == "production" && dropTables {
		log.Println("WAARSCHUWING: DB_DROP_TABLES=true in productie! Dit zal alle data wissen.")
		log.Println("Overweeg om DB_DROP_TABLES=false te zetten of controleer je environment.")
	}

	return &Config{
		// Server configuratie
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		Environment:   environment,

		// Database configuratie
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "Bootje@12"),
		DBName:     getEnv("DB_NAME", "odomosml"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		DropTables: dropTables,

		// JWT configuratie
		JWTSecret:          jwtSecret,
		JWTExpirationHours: jwtExpirationHours,

		// Logging configuratie
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

// GetDSN geeft de database connection string terug
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// IsProduction controleert of de applicatie in productie draait
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment controleert of de applicatie in development draait
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Helper functies voor environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return defaultValue
}
