package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"odomosml/config"
	"odomosml/internal/middleware"
	"odomosml/pkg/database"

	auditHttp "odomosml/internal/audit/delivery/http"
	auditRepo "odomosml/internal/audit/repository"
	auditService "odomosml/internal/audit/service"

	authHttp "odomosml/internal/auth/delivery/http"
	authService "odomosml/internal/auth/service"

	customerHttp "odomosml/internal/customer/delivery/http"
	customerRepo "odomosml/internal/customer/repository"
	customerService "odomosml/internal/customer/service"

	userHttp "odomosml/internal/user/delivery/http"
	userModel "odomosml/internal/user/model"
	userRepo "odomosml/internal/user/repository"
	userService "odomosml/internal/user/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Create a new Gin router with default middleware
	router := gin.Default()

	// Configure CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Initialize repositories
	customerRepository := customerRepo.NewCustomerRepository(db)
	userRepository := userRepo.NewUserRepository(db)
	auditRepository := auditRepo.NewAuditRepository(db)

	// Initialize services
	custService := customerService.NewCustomerService(customerRepository)
	userService := userService.NewUserService(userRepository)
	auditService := auditService.NewAuditService(auditRepository)
	authService := authService.NewAuthService(userRepository, cfg)

	// Initialize handlers
	customerHandler := customerHttp.NewCustomerHandler(custService)
	userHandler := userHttp.NewUserHandler(userService)
	auditHandler := auditHttp.NewAuditHandler(auditService)
	authHandler := authHttp.NewAuthHandler(authService)

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(authService)
	adminOnly := middleware.RoleMiddleware(userModel.RoleAdmin)
	adminOrUser := middleware.RoleMiddleware(userModel.RoleAdmin, userModel.RoleUser)

	// Public routes (no authentication required)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/register", authHandler.Register)

	// Add audit middleware for authenticated routes
	router.Use(authMiddleware)

	// Configure audit middleware
	auditMiddlewareConfig := middleware.AuditMiddlewareConfig{
		AuditService: auditService,
		CustomerRepo: customerRepository,
		UserRepo:     userRepository,
	}
	router.Use(middleware.NewAuditMiddleware(auditMiddlewareConfig))

	// Protected routes
	// Customer routes (requires admin or user role)
	customerGroup := router.Group("/api/klanten", adminOrUser)
	{
		customerGroup.GET("", customerHandler.GetAll)
		customerGroup.GET("/:id", customerHandler.GetByID)
		customerGroup.POST("", customerHandler.Create)
		customerGroup.PUT("/:id", customerHandler.Update)
		customerGroup.PATCH("/:id", customerHandler.PartialUpdate)
		customerGroup.DELETE("/:id", customerHandler.Delete)
	}

	// User management routes (requires admin role)
	userGroup := router.Group("/api/users", adminOnly)
	{
		userGroup.GET("", userHandler.GetAll)
		userGroup.GET("/:id", userHandler.GetByID)
		userGroup.POST("", userHandler.Create)
		userGroup.PUT("/:id", userHandler.Update)
		userGroup.DELETE("/:id", userHandler.Delete)
	}

	// Auth routes
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// Audit log routes (requires admin role)
	auditGroup := router.Group("/api/logs", adminOnly)
	{
		auditGroup.GET("", auditHandler.GetLogs)
	}

	// Start the server
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
