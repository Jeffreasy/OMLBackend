package app

import (
	"log"
	"odomosml/config"
	auditHandler "odomosml/internal/audit/delivery/http"
	auditRepo "odomosml/internal/audit/repository"
	auditService "odomosml/internal/audit/service"
	authHandler "odomosml/internal/auth/delivery/http"
	authService "odomosml/internal/auth/service"
	customerHandler "odomosml/internal/customer/delivery/http"
	customerRepo "odomosml/internal/customer/repository"
	customerService "odomosml/internal/customer/service"
	"odomosml/internal/middleware"
	userHandler "odomosml/internal/user/delivery/http"
	userModel "odomosml/internal/user/model"
	userRepo "odomosml/internal/user/repository"
	userService "odomosml/internal/user/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App struct bevat de applicatie configuratie
type App struct {
	router *gin.Engine
	db     *gorm.DB
	config *config.Config
}

// GetRouter retourneert de gin router instance
func (a *App) GetRouter() *gin.Engine {
	return a.router
}

// NewApp maakt een nieuwe applicatie instantie
func NewApp(db *gorm.DB, cfg *config.Config) *App {
	// Stel Gin mode in op basis van environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Maak een nieuwe app instantie
	app := &App{
		router: router,
		db:     db,
		config: cfg,
	}

	// Initialiseer routes
	app.setupRoutes()

	return app
}

// setupRoutes initialiseert alle routes en middlewares
func (a *App) setupRoutes() {
	// Initialiseer repositories
	userRepository := userRepo.NewUserRepository(a.db)
	customerRepository := customerRepo.NewCustomerRepository(a.db)
	auditRepository := auditRepo.NewAuditRepository(a.db)

	// Initialiseer services
	userSvc := userService.NewUserService(userRepository)
	customerSvc := customerService.NewCustomerService(customerRepository)
	auditSvc := auditService.NewAuditService(auditRepository)
	authSvc := authService.NewAuthService(userRepository, a.config)

	// Initialiseer middlewares
	authMiddleware := middleware.AuthMiddleware(authSvc)
	auditMiddleware := middleware.NewAuditMiddleware(auditSvc)

	// Initialiseer handlers
	userHandler := userHandler.NewUserHandler(userSvc)
	customerHandler := customerHandler.NewCustomerHandler(customerSvc)
	auditHandler := auditHandler.NewAuditHandler(auditSvc)
	authHandler := authHandler.NewAuthHandler(authSvc)

	// API routes
	api := a.router.Group("/api")

	// Auth routes (publiek)
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authMiddleware, authHandler.Refresh)
	}

	// User routes (alleen admin)
	users := api.Group("/users")
	users.Use(authMiddleware, middleware.RoleMiddleware(userModel.RoleAdmin), auditMiddleware)
	{
		users.GET("", userHandler.GetAll)
		users.GET("/:id", userHandler.GetByID)
		users.POST("", userHandler.Create)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
	}

	// Customer routes (admin en user)
	customers := api.Group("/klanten")
	customers.Use(authMiddleware, middleware.RoleMiddleware(userModel.RoleAdmin, userModel.RoleUser), auditMiddleware)
	{
		customers.GET("", customerHandler.GetAll)
		customers.GET("/:id", customerHandler.GetByID)
		customers.POST("", customerHandler.Create)
		customers.PUT("/:id", customerHandler.Update)
		customers.PATCH("/:id", customerHandler.PartialUpdate)
		customers.DELETE("/:id", customerHandler.Delete)
	}

	// Audit log routes (alleen admin)
	logs := api.Group("/logs")
	logs.Use(authMiddleware, middleware.RoleMiddleware(userModel.RoleAdmin))
	{
		logs.GET("", auditHandler.GetLogs)
	}
}

// Run start de applicatie
func (a *App) Run() error {
	log.Printf("Starting server on %s", a.config.ServerAddress)
	return a.router.Run(a.config.ServerAddress)
}
