package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cfg "github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/internal/market/service"
	"github.com/kymo-mcp/mcpcan/internal/market/task"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/middleware"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"github.com/kymo-mcp/mcpcan/pkg/scheduler"
	"github.com/kymo-mcp/mcpcan/pkg/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// App application structure
type App struct {
	// config configuration
	config *cfg.Config

	// logger logger
	logger *zap.Logger

	// scheduler scheduler
	scheduler scheduler.Scheduler

	// taskManager task manager
	taskManager task.TaskManager

	// httpServer HTTP server
	httpServer *http.Server

	// ginEngine Gin engine
	ginEngine *gin.Engine

	// shutdownCtx shutdown context
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New creates new application instance
func New() (*App, error) {
	// Load global configuration
	cfg, err := cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Print configuration information
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", cfg.VersionInfo)))

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config:         cfg,
		logger:         logger.L().Logger,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}, nil
}

// Initialize initialize all application components
func (a *App) Initialize() error {
	// Initialize database
	if err := database.Init(&a.config.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis
	if err := redis.Init(&a.config.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Use global database repository instance (already initialized in init)
	if mysql.McpInstanceRepo == nil {
		return fmt.Errorf("McpInstanceRepo not properly initialized, please check database initialization process")
	}

	// Load service configuration
	if err := services.LoadServices(&a.config.Services); err != nil {
		return fmt.Errorf("failed to load service configuration: %w", err)
	}

	// Start scheduler
	if err := a.initializeScheduler(); err != nil {
		return fmt.Errorf("failed to initialize scheduler: %w", err)
	}

	// Initialize task manager, no longer depends on global container runtime
	a.taskManager = task.NewTaskManager(
		mysql.McpInstanceRepo,
		a.scheduler,
		a.logger,
	)

	// Set up global tasks
	if err := a.taskManager.SetupGlobalTasks(a.shutdownCtx); err != nil {
		return fmt.Errorf("failed to set up global tasks: %w", err)
	}

	// Initialize HTTP server
	if err := a.initializeHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	a.logger.Info("Application initialization completed")
	return nil
}

// initializeScheduler initialize scheduler
func (a *App) initializeScheduler() error {
	globalScheduler := scheduler.GetGlobalScheduler()
	if globalScheduler == nil {
		return fmt.Errorf("global scheduler not initialized")
	}

	a.scheduler = globalScheduler.GetTaskManager().GetScheduler()
	a.logger.Info("Scheduler initialized successfully")

	return nil
}

// initializeHTTPServer initialize HTTP server
func (a *App) initializeHTTPServer() error {

	a.ginEngine = gin.Default()

	// Set up middleware
	a.setupMiddleware()

	// Initialize Gin engine
	a.setupHttpServer()

	// Create HTTP server
	serverAddr := fmt.Sprintf(":%d", a.config.Server.HttpPort)
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: a.ginEngine,
	}

	return nil
}

// Run run application
func (a *App) Run() error {
	// Start task manager
	err := a.taskManager.StartMonitoring(a.shutdownCtx)
	if err != nil {
		return fmt.Errorf("failed to start task manager: %w", err)
	}

	// Start HTTP server
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server startup failed", zap.Error(err))
		}
	}()

	a.logger.Info("Application started successfully",
		zap.String("address", a.httpServer.Addr))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down application...")

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shutdown application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Stop task manager
	if a.taskManager != nil {
		err := a.taskManager.StopMonitoring(ctx)
		if err != nil {
			a.logger.Error("Failed to stop task manager", zap.Error(err))
		}
	}

	// Close HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown failed", zap.Error(err))
			return err
		}
	}

	// Cancel application context
	if a.shutdownCancel != nil {
		a.shutdownCancel()
	}

	a.logger.Info("Application has been gracefully shut down")
	return nil
}

// setupHttpServer initialize Gin engine and register all routes
func (a *App) setupHttpServer() {
	// Set file upload size limit, default is 32 MiB, according to configuration file set to 100 MiB
	a.ginEngine.MaxMultipartMemory = int64(a.config.Code.Upload.MaxFileSize) << 20

	// Get route prefix
	routerPrefix := common.GetMarketRoutePrefix()
	routerPrefix = strings.Trim(routerPrefix, "/")

	// Register instance management interface
	instanceService := service.NewInstanceService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/create", routerPrefix), instanceService.CreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/instance/:instanceId", routerPrefix), instanceService.DetailHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/edit", routerPrefix), instanceService.EditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/list", routerPrefix), instanceService.ListHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/disabled", routerPrefix), instanceService.DisabledHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/restart", routerPrefix), instanceService.RestartHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/instance/:instanceId", routerPrefix), instanceService.DeleteHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/instance/status/:instanceId", routerPrefix), instanceService.StatusHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/logs", routerPrefix), instanceService.LogsHandler)

	// Create resource management service instance
	resourceService := service.NewResourceService(context.Background())
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.ListPVCsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.CreatePVCHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/nodes", routerPrefix), resourceService.ListNodesHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/storage-classes", routerPrefix), resourceService.ListStorageClassesHandler)

	// Create environment management service instance
	environmentService := service.NewEnvironmentService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/environments", routerPrefix), environmentService.CreateEnvironmentHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/environments/:id", routerPrefix), environmentService.UpdateEnvironmentHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/environments/:id", routerPrefix), environmentService.DeleteEnvironmentHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/environments", routerPrefix), environmentService.ListEnvironmentsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/namespaces", routerPrefix), environmentService.ListNamespacesHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/:id/test", routerPrefix), environmentService.TestConnectivityHandler)

	// Register code management interface
	codeService := service.NewCodeService()
	a.ginEngine.POST(fmt.Sprintf("/%s/code/upload", routerPrefix), codeService.UploadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/tree", routerPrefix), codeService.GetCodeTree)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/get", routerPrefix), codeService.GetCodeFile)
	a.ginEngine.POST(fmt.Sprintf("/%s/code/edit", routerPrefix), codeService.EditCodeFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/download/:packageId", routerPrefix), codeService.DownloadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/packages", routerPrefix), codeService.GetCodePackageList)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/code/packages/:packageId", routerPrefix), codeService.DeleteCodePackage)

	// Register template management interface
	templateService := service.NewTemplateService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/template/create", routerPrefix), templateService.TemplateCreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDetailHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/template/edit", routerPrefix), templateService.TemplateEditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/template/list", routerPrefix), templateService.TemplateListHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/list/pagination", routerPrefix), templateService.TemplateListWithPaginationHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDeleteHandler)

	// Register market management interface
	marketService := service.NewMarketService()
	if marketService != nil {
		a.ginEngine.POST(fmt.Sprintf("/%s/market/list", routerPrefix), marketService.ListMarketServices)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/detail", routerPrefix), marketService.GetMarketServiceDetail)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/category", routerPrefix), marketService.GetMarketCategories)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/config", routerPrefix), marketService.GetMarketConfig)
	}

	// Register storage management interface
	storageService := service.NewStorageService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/storage/image", routerPrefix), storageService.UploadImageHandler)

	// Register dashboard management interface
	dashboardService := service.NewDashboardService(context.Background())
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/statistical", routerPrefix), dashboardService.StatisticalHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/available-cases", routerPrefix), dashboardService.AvailableCasesHandler)

	// Health check
	a.ginEngine.GET("/health", func(c *gin.Context) {
		i18n.SuccessResponse(c, gin.H{"status": "ok"})
	})
}

// setupMiddleware set up middleware
func (a *App) setupMiddleware() {
	// Add panic recovery middleware
	a.ginEngine.Use(middleware.PanicRecovery())

	// Add request response logging middleware
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// Add cross domain handling
	domains := []string{"*"}
	a.ginEngine.Use(middleware.CORSMiddleware(domains))

	// Add internationalization middleware
	a.ginEngine.Use(middleware.I18nMiddleware())

	// Add security middleware
	a.ginEngine.Use(middleware.SecurityMiddleware(a.config.Secret))

	// Add authentication middleware
	a.ginEngine.Use(middleware.AuthTokenMiddleware(a.config.Secret))

	// Add error handling middleware (must be last)
	a.ginEngine.Use(middleware.ErrorHandler())

	// Set custom error handler
	a.ginEngine.NoRoute(middleware.NotFoundHandler)
	a.ginEngine.NoMethod(middleware.MethodNotAllowedHandler)
}
