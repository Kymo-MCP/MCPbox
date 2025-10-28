package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/internal/authz/config"
	"qm-mcp-server/internal/authz/service"
	"qm-mcp-server/pkg/common"
	dbpkg "qm-mcp-server/pkg/database"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/middleware"
	"qm-mcp-server/pkg/redis"
)

// App application structure
type App struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	ginEngine  *gin.Engine
}

// New creates application instance
func New() *App {
	// Load configuration
	if err := config.Load(); err != nil {
		return nil
	}

	// Initialize logger
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil
	}

	// Print configuration information
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", config.GlobalConfig.VersionInfo)))

	return &App{
		config: config.GlobalConfig,
		logger: logger.L().Logger,
	}
}

// Initialize initializes the application
func (a *App) Initialize() error {
	// Initialize Redis
	if err := redis.Init(&config.GlobalConfig.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	// Initialize database
	if err := dbpkg.Init(&a.config.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Setup HTTP server
	if err := a.setupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	logger.Info("Authz service initialized successfully")
	return nil
}

// setupHTTPServer sets up HTTP server
func (a *App) setupHTTPServer() error {
	// Set Gin mode
	gin.SetMode(gin.DebugMode)

	// Create Gin engine
	a.ginEngine = gin.New()

	// Setup middleware
	a.setupMiddleware()

	// Setup routes
	a.setupRoutes()

	// Create HTTP server
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Server.HttpPort),
		Handler: a.ginEngine,
	}

	logger.Info("HTTP server setup completed")
	return nil
}

// setupMiddleware sets up middleware
func (a *App) setupMiddleware() {
	// Add middleware
	a.ginEngine.Use(gin.Recovery())
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// Add CORS handling
	a.ginEngine.Use(middleware.CORSMiddleware([]string{"*"}))

	// Add internationalization middleware
	a.ginEngine.Use(middleware.I18nMiddleware())

	// Add security middleware
	a.ginEngine.Use(middleware.SecurityMiddleware(config.GlobalConfig.Secret))

	// Add authentication middleware
	a.ginEngine.Use(middleware.AuthTokenMiddleware(config.GlobalConfig.Secret))
}

// setupRoutes sets up routes
func (a *App) setupRoutes() {
	// Create service instances
	userService := service.NewUserService()

	userAuthService := service.NewUserAuthService()

	// Health check
	a.ginEngine.GET("/health", func(c *gin.Context) {
		common.GinSuccess(c, map[string]string{"status": "ok"})
	})

	// API version prefix
	authzGroup := a.ginEngine.Group(common.GetAuthzRoutePrefix())

	// User related routes
	userGroup := authzGroup.Group("/users")
	{
		userGroup.POST("", userService.CreateUser)
		userGroup.GET("/:id", userService.GetUserById)
		userGroup.PUT("/:id", userService.UpdateUser)
		userGroup.DELETE("/:id", userService.DeleteUser)
		userGroup.GET("", userService.ListUsers)
		// update-password
		userGroup.PUT("/update-password", userService.UpdatePassword)
		// update-avatar
		userGroup.PUT("/update-avatar", userService.UpdateAvatar)
	}

	// Authentication related routes - updated to use UserAuthService

	{
		// User login
		authzGroup.POST("/login", userAuthService.Login)

		// User logout
		authzGroup.POST("/logout", userAuthService.Logout)

		// Get user configuration
		authzGroup.GET("/user-info", userAuthService.GetUserInfo)

		// Refresh Token
		authzGroup.POST("/refresh", userAuthService.RefreshToken)

		// Validate Token
		authzGroup.POST("/validate", userAuthService.ValidateToken)

		// Get encryption key
		authzGroup.POST("/encryption-key", userAuthService.GetEncryptionKey)
	}
}

// Run runs the application
func (a *App) Run() error {
	// Start HTTP server
	logger.Info("Starting HTTP server", zap.Int("port", a.config.Server.HttpPort))
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down authz service...")

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if a.httpServer != nil {
		logger.Info("Shutting down HTTP server...")
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}
	}

	// Close database connection
	if err := mysql.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	logger.Info("Authz service shutdown completed")
	return nil
}
