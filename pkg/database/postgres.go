package database

import (
	"fmt"
	"log"
	"odomosml/internal/user/model"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config bevat de database configuratie
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB initialiseert een nieuwe PostgreSQL database connectie
func NewPostgresDB(config Config) (*gorm.DB, error) {
	// Check of tabellen moeten worden gedropped
	shouldDropTables := getEnvOrDefault("DB_DROP_TABLES", "false") == "true"

	// Gebruik config of fallback naar environment variables
	dbHost := getConfigOrEnv(config.Host, "DB_HOST", "localhost")
	dbPort := getConfigOrEnv(config.Port, "DB_PORT", "5432")
	dbUser := getConfigOrEnv(config.User, "DB_USER", "postgres")
	dbPass := getConfigOrEnv(config.Password, "DB_PASSWORD", "postgres")
	dbName := getConfigOrEnv(config.DBName, "DB_NAME", "odomosml")
	sslMode := getConfigOrEnv(config.SSLMode, "DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	// Configureer GORM met debug logging
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Connected to database")

	// Drop tables alleen als DB_DROP_TABLES=true
	if shouldDropTables {
		log.Println("Dropping existing tables...")
		if err := dropTables(db); err != nil {
			return nil, err
		}
	}

	log.Println("Creating or updating tables...")
	if err := createTables(db); err != nil {
		return nil, err
	}

	// Maak admin gebruiker aan als deze nog niet bestaat
	if err := ensureAdminExists(db); err != nil {
		return nil, err
	}

	return db, nil
}

// getConfigOrEnv haalt een waarde op uit config, environment variable of default
func getConfigOrEnv(configValue, envKey, defaultValue string) string {
	if configValue != "" {
		return configValue
	}
	return getEnvOrDefault(envKey, defaultValue)
}

// dropTables verwijdert bestaande tabellen
func dropTables(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS audit_logs CASCADE").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS users CASCADE").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS customers CASCADE").Error; err != nil {
		return err
	}
	return nil
}

// createTables maakt de benodigde tabellen aan
func createTables(db *gorm.DB) error {
	// Auto-migrate zal tabellen aanmaken of updaten zonder data te verwijderen
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("error migrating users table: %v", err)
	}

	// Voeg hier andere modellen toe voor auto-migrate
	// if err := db.AutoMigrate(&model.Customer{}); err != nil {
	//     return fmt.Errorf("error migrating customers table: %v", err)
	// }

	return nil
}

// ensureAdminExists zorgt dat er een admin gebruiker bestaat
func ensureAdminExists(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Where("role = ?", "ADMIN").Count(&count)

	if count == 0 {
		log.Println("Creating admin user...")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminUser := model.User{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			Role:      "ADMIN",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
		log.Printf("Admin user created successfully with ID: %d\n", adminUser.ID)
	}

	return nil
}

// getEnvOrDefault haalt een environment variable op of geeft een default waarde terug
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
