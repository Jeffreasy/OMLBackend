package main

import (
	"log"
	"odomosml/config"
	"odomosml/docs"
	"odomosml/internal/app"
	"odomosml/pkg/database"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           OdomosMintLogic API
// @version         1.0
// @description     API voor het OdomosMintLogic systeem
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.odomosml.com/support
// @contact.email  support@odomosml.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Laad configuratie uit environment variables
	cfg := config.LoadConfig()

	// Log startup informatie
	log.Printf("Starting OdomosML API in %s mode", cfg.Environment)
	log.Printf("Server will listen on %s", cfg.ServerAddress)

	// Initialiseer database connectie
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialiseer en start de applicatie
	application := app.NewApp(db, cfg)

	// Voeg Swagger route toe
	docs.SwaggerInfo.BasePath = "/api/v1"
	application.GetRouter().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start de applicatie
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
