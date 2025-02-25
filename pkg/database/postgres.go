package database

import (
	"fmt"
	"log"
	"odomosml/config"
	auditModel "odomosml/internal/audit/model"
	customerModel "odomosml/internal/customer/model"
	userModel "odomosml/internal/user/model"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB initialiseert een nieuwe PostgreSQL database connectie
func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	// Configureer GORM logger op basis van de applicatie log level
	gormLogLevel := logger.Info
	if cfg.IsProduction() {
		gormLogLevel = logger.Error
	}

	// Maak connectie met de database
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configureer connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Drop tabellen indien nodig (alleen in development!)
	if cfg.DropTables {
		if cfg.IsProduction() {
			log.Println("WAARSCHUWING: Tabellen droppen in productie is gevaarlijk!")
		}
		log.Println("Dropping existing tables...")
		if err := dropTables(db); err != nil {
			return nil, fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	// Migreer database schema
	if err := migrateSchema(db); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	// Maak admin gebruiker aan indien nodig
	if err := ensureAdminExists(db); err != nil {
		return nil, fmt.Errorf("failed to ensure admin exists: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// dropTables verwijdert alle tabellen uit de database
func dropTables(db *gorm.DB) error {
	log.Println("Dropping tables: audit_logs, customers, users")
	if err := db.Migrator().DropTable(&auditModel.AuditLog{}, "customers", &userModel.User{}); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	return nil
}

// createIndexes maakt indexen aan voor betere performance
func createIndexes(db *gorm.DB) error {
	// Maak de pg_trgm extensie aan voor trigram indexen (voor ILIKE queries)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;").Error; err != nil {
		log.Printf("Waarschuwing: Kon pg_trgm extensie niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}

	// Indexen voor User model
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);").Error; err != nil {
		return err
	}

	// Trigram indexen voor User model (voor ILIKE zoekopdrachten)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);").Error; err != nil {
		log.Printf("Waarschuwing: Kon trigram index voor users.email niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin (username gin_trgm_ops);").Error; err != nil {
		log.Printf("Waarschuwing: Kon trigram index voor users.username niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}

	// Indexen voor Customer model
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);").Error; err != nil {
		return err
	}

	// Trigram indexen voor Customer model (voor ILIKE zoekopdrachten)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_customers_email_trgm ON customers USING gin (email gin_trgm_ops);").Error; err != nil {
		log.Printf("Waarschuwing: Kon trigram index voor customers.email niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_customers_name_trgm ON customers USING gin (name gin_trgm_ops);").Error; err != nil {
		log.Printf("Waarschuwing: Kon trigram index voor customers.name niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}

	// Indexen voor AuditLog model
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_action_type ON audit_logs(action_type);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_type ON audit_logs(entity_type);").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);").Error; err != nil {
		return err
	}

	return nil
}

// migrateSchema migreert het database schema
func migrateSchema(db *gorm.DB) error {
	log.Println("Migrating database schema...")

	// Migreer modellen
	if err := db.AutoMigrate(
		&userModel.User{},
		&customerModel.Customer{},
		&auditModel.AuditLog{},
	); err != nil {
		return err
	}

	// Maak indexen aan
	if err := createIndexes(db); err != nil {
		log.Printf("Waarschuwing: Kon sommige indexen niet aanmaken: %v", err)
		// Ga door, dit is niet kritiek
	}

	return nil
}

// ensureAdminExists zorgt ervoor dat er minstens één admin gebruiker bestaat
func ensureAdminExists(db *gorm.DB) error {
	var count int64
	db.Model(&userModel.User{}).Where("role = ?", userModel.RoleAdmin).Count(&count)
	if count == 0 {
		log.Println("Creating admin user...")
		adminUser := &userModel.User{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "admin123", // Dit wordt automatisch gehasht door BeforeSave hook
			Role:     userModel.RoleAdmin,
			Active:   true,
		}
		if err := db.Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
		log.Println("Admin user created successfully")
	}
	return nil
}
