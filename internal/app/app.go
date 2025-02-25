package app

import (
	"fmt"
	auditHttp "odomosml/internal/audit/delivery/http"
	auditRepo "odomosml/internal/audit/repository"
	auditService "odomosml/internal/audit/service"
	authHttp "odomosml/internal/auth/delivery/http"
	authService "odomosml/internal/auth/service"
	"odomosml/internal/config"
	customerHttp "odomosml/internal/customer/delivery/http"
	customerRepo "odomosml/internal/customer/repository"
	customerService "odomosml/internal/customer/service"
	"odomosml/internal/middleware"
	userHttp "odomosml/internal/user/delivery/http"
	userRepo "odomosml/internal/user/repository"
	userService "odomosml/internal/user/service"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	db     *gorm.DB
	router *gin.Engine
	config *config.Config
}

func NewApp(db *gorm.DB) *App {
	// Laad configuratie
	cfg := &config.Config{
		JWTSecret:          getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		JWTExpirationHours: 24,
	}

	return &App{
		db:     db,
		router: gin.Default(),
		config: cfg,
	}
}

func (a *App) Run() error {
	// Repositories
	userRepository := userRepo.NewUserRepository(a.db)
	customerRepository := customerRepo.NewCustomerRepository(a.db)
	auditRepository := auditRepo.NewAuditRepository(a.db)

	// Services
	userSvc := userService.NewUserService(userRepository)
	customerSvc := customerService.NewCustomerService(customerRepository)
	authSvc := authService.NewAuthService(userRepository, a.config)
	auditSvc := auditService.NewAuditService(auditRepository)

	// Middleware config
	auditConfig := middleware.AuditMiddlewareConfig{
		AuditService: auditSvc,
		CustomerRepo: customerRepository,
		UserRepo:     userRepository,
	}

	// Middleware
	authMiddleware := middleware.NewJWTAuthMiddleware(a.config.JWTSecret)
	auditMiddleware := middleware.NewAuditMiddleware(auditConfig)
	adminMiddleware := middleware.NewRoleMiddleware("ADMIN")

	// Handlers
	authHandler := authHttp.NewAuthHandler(authSvc)
	userHandler := userHttp.NewUserHandler(userSvc)
	customerHandler := customerHttp.NewCustomerHandler(customerSvc)
	auditHandler := auditHttp.NewAuditHandler(auditSvc)

	// Public routes
	auth := a.router.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authMiddleware, authHandler.Refresh)
	}

	// Protected routes
	api := a.router.Group("/api")
	api.Use(authMiddleware)
	{
		// Klanten routes
		customers := api.Group("/klanten")
		customers.Use(auditMiddleware)
		{
			customers.GET("", customerHandler.GetAll)
			customers.GET("/:id", customerHandler.GetByID)
			customers.POST("", customerHandler.Create)
			customers.PUT("/:id", customerHandler.Update)
			customers.PATCH("/:id", customerHandler.PartialUpdate)
			customers.DELETE("/:id", customerHandler.Delete)
		}

		// Gebruikers routes (alleen voor admins)
		users := api.Group("/users")
		users.Use(adminMiddleware, auditMiddleware)
		{
			users.GET("", userHandler.GetAll)
			users.GET("/:id", userHandler.GetByID)
			users.POST("", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// Audit logs routes (alleen voor admins)
		logs := api.Group("/logs")
		logs.Use(adminMiddleware)
		{
			logs.GET("", auditHandler.GetLogs)
		}
	}

	// Start de server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return a.router.Run(fmt.Sprintf(":%s", port))
}

// getEnvOrDefault haalt een environment variable op of geeft een default waarde terug
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
