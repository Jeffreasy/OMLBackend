package main

import (
	"log"
	"odomosml/internal/app"
	"odomosml/pkg/database"
)

func main() {
	// Database configuratie
	dbConfig := database.Config{
		Host:     "", // Leeg laten voor environment variable
		Port:     "",
		User:     "",
		Password: "",
		DBName:   "",
		SSLMode:  "",
	}

	// Initialiseer database
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start de applicatie
	app := app.NewApp(db)
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
