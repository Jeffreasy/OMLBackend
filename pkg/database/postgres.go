package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	auditModel "odomosml/internal/audit/model"
	customerModel "odomosml/internal/customer/model"
	userModel "odomosml/internal/user/model"
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

// NewPostgresDB maakt een nieuwe database connectie
func NewPostgresDB(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	// Enable detailed logging
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	log.Println("Connected to database, dropping existing tables...")

	// Drop existing tables
	err = db.Migrator().DropTable(
		&customerModel.Customer{},
		&userModel.User{},
		&auditModel.AuditLog{},
	)
	if err != nil {
		log.Printf("failed to drop tables: %v", err)
	}

	log.Println("Creating new tables...")

	// Auto Migrate the schema
	err = db.AutoMigrate(
		&customerModel.Customer{},
		&userModel.User{},
		&auditModel.AuditLog{},
	)
	if err != nil {
		log.Printf("failed to migrate database: %v", err)
		return nil, err
	}

	log.Println("Creating admin user...")

	// Create admin user
	adminUser := &userModel.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "admin123",
		Role:     userModel.RoleAdmin,
		Active:   true,
	}

	if err := db.Create(adminUser).Error; err != nil {
		log.Printf("failed to create admin user: %v", err)
	} else {
		log.Printf("Admin user created successfully with ID: %d", adminUser.ID)
	}

	return db, nil
}
