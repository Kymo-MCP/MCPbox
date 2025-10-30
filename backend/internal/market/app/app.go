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

	cfg "qm-mcp-server/internal/market/config"
	"qm-mcp-server/internal/market/service"
	"qm-mcp-server/internal/market/task"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/middleware"
	"qm-mcp-server/pkg/redis"
	"qm-mcp-server/pkg/scheduler"
	"qm-mcp-server/pkg/services"

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
	// 加载全局配置
	cfg, err := cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	// 打印配置信息
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", cfg.VersionInfo)))

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config:         cfg,
		logger:         logger.L().Logger,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}, nil
}

// Initialize 初始化应用程序所有组件
func (a *App) Initialize() error {
	// 初始化数据库
	if err := database.Init(&a.config.Database); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化Redis
	if err := redis.Init(&a.config.Database.Redis); err != nil {
		return fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 使用全局数据库仓库实例（已在init中初始化）
	if mysql.McpInstanceRepo == nil {
		return fmt.Errorf("McpInstanceRepo 未正确初始化，请检查数据库初始化流程")
	}

	// 加载服务配置
	if err := services.LoadServices(&a.config.Services); err != nil {
		return fmt.Errorf("加载服务配置失败: %w", err)
	}

	// 启动调度器
	if err := a.initializeScheduler(); err != nil {
		return fmt.Errorf("初始化调度器失败: %w", err)
	}

	// 初始化任务管理器，不再依赖全局容器运行时
	a.taskManager = task.NewTaskManager(
		mysql.McpInstanceRepo,
		a.scheduler,
		a.logger,
	)

	// 设置全局任务
	if err := a.taskManager.SetupGlobalTasks(a.shutdownCtx); err != nil {
		return fmt.Errorf("设置全局任务失败: %w", err)
	}

	// 初始化 HTTP 服务器
	if err := a.initializeHTTPServer(); err != nil {
		return fmt.Errorf("初始化HTTP服务器失败: %w", err)
	}

	a.logger.Info("应用程序初始化完成")
	return nil
}

// initializeScheduler 初始化调度器
func (a *App) initializeScheduler() error {
	globalScheduler := scheduler.GetGlobalScheduler()
	if globalScheduler == nil {
		return fmt.Errorf("全局调度器未初始化")
	}

	a.scheduler = globalScheduler.GetTaskManager().GetScheduler()
	a.logger.Info("调度器初始化成功")

	return nil
}

// initializeHTTPServer 初始化HTTP服务器
func (a *App) initializeHTTPServer() error {

	a.ginEngine = gin.Default()

	// 设置中间件
	a.setupMiddleware()

	// 初始化 Gin 引擎
	a.setupHttpServer()

	// 创建 HTTP 服务器
	serverAddr := fmt.Sprintf(":%d", a.config.Server.HttpPort)
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: a.ginEngine,
	}

	return nil
}

// Run 运行应用程序
func (a *App) Run() error {
	// 启动任务管理器
	err := a.taskManager.StartMonitoring(a.shutdownCtx)
	if err != nil {
		return fmt.Errorf("启动任务管理器失败: %w", err)
	}

	// 启动HTTP服务器
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	a.logger.Info("应用程序启动成功",
		zap.String("address", a.httpServer.Addr))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("正在关闭应用程序...")

	// 优雅关闭
	return a.Shutdown()
}

// Shutdown 优雅关闭应用程序
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 停止任务管理器
	if a.taskManager != nil {
		err := a.taskManager.StopMonitoring(ctx)
		if err != nil {
			a.logger.Error("停止任务管理器失败", zap.Error(err))
		}
	}

	// 关闭HTTP服务器
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP服务器关闭失败", zap.Error(err))
			return err
		}
	}

	// 取消应用程序上下文
	if a.shutdownCancel != nil {
		a.shutdownCancel()
	}

	a.logger.Info("应用程序已优雅关闭")
	return nil
}

// setupHttpServer 初始化 Gin 引擎并注册所有路由
func (a *App) setupHttpServer() {
	// 设置文件上传大小限制，默认为 32 MiB，根据配置文件设置为 100 MiB
	a.ginEngine.MaxMultipartMemory = int64(a.config.Code.Upload.MaxFileSize) << 20

	// 获取路由前缀
	routerPrefix := common.GetMarketRoutePrefix()
	routerPrefix = strings.Trim(routerPrefix, "/")

	// 注册实例管理接口
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

	// 创建资源管理服务实例
	resourceService := service.NewResourceService(context.Background())
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.ListPVCsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.CreatePVCHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/nodes", routerPrefix), resourceService.ListNodesHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/storage-classes", routerPrefix), resourceService.ListStorageClassesHandler)

	// 创建环境管理服务实例
	environmentService := service.NewEnvironmentService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/environments", routerPrefix), environmentService.CreateEnvironmentHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/environments/:id", routerPrefix), environmentService.UpdateEnvironmentHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/environments/:id", routerPrefix), environmentService.DeleteEnvironmentHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/environments", routerPrefix), environmentService.ListEnvironmentsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/namespaces", routerPrefix), environmentService.ListNamespacesHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/:id/test", routerPrefix), environmentService.TestConnectivityHandler)

	// 注册代码管理接口
	codeService := service.NewCodeService()
	a.ginEngine.POST(fmt.Sprintf("/%s/code/upload", routerPrefix), codeService.UploadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/tree", routerPrefix), codeService.GetCodeTree)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/get", routerPrefix), codeService.GetCodeFile)
	a.ginEngine.POST(fmt.Sprintf("/%s/code/edit", routerPrefix), codeService.EditCodeFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/download/:packageId", routerPrefix), codeService.DownloadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/packages", routerPrefix), codeService.GetCodePackageList)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/code/packages/:packageId", routerPrefix), codeService.DeleteCodePackage)

	// 注册模板管理接口
	templateService := service.NewTemplateService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/template/create", routerPrefix), templateService.TemplateCreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDetailHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/template/edit", routerPrefix), templateService.TemplateEditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/template/list", routerPrefix), templateService.TemplateListHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/list/pagination", routerPrefix), templateService.TemplateListWithPaginationHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDeleteHandler)

	// 注册市场管理接口
	marketService := service.NewMarketService()
	if marketService != nil {
		a.ginEngine.POST(fmt.Sprintf("/%s/market/list", routerPrefix), marketService.ListMarketServices)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/detail", routerPrefix), marketService.GetMarketServiceDetail)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/category", routerPrefix), marketService.GetMarketCategories)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/config", routerPrefix), marketService.GetMarketConfig)
	}

	// 注册存储管理接口
	storageService := service.NewStorageService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/storage/image", routerPrefix), storageService.UploadImageHandler)

	// 注册 dashboard 管理接口
	dashboardService := service.NewDashboardService(context.Background())
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/statistical", routerPrefix), dashboardService.StatisticalHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/available-cases", routerPrefix), dashboardService.AvailableCasesHandler)

	// 健康检查
	a.ginEngine.GET("/health", func(c *gin.Context) {
		i18n.SuccessResponse(c, gin.H{"status": "ok"})
	})
}

// setupMiddleware 设置中间件
func (a *App) setupMiddleware() {
	// 添加恐慌恢复中间件
	a.ginEngine.Use(middleware.PanicRecovery())

	// 添加请求响应日志中间件
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// 添加跨域处理
	domains := []string{"*"}
	a.ginEngine.Use(middleware.CORSMiddleware(domains))

	// 添加国际化中间件
	a.ginEngine.Use(middleware.I18nMiddleware())

	// 添加安全中间件
	a.ginEngine.Use(middleware.SecurityMiddleware(a.config.Secret))

	// 添加认证中间件
	a.ginEngine.Use(middleware.AuthTokenMiddleware(a.config.Secret))

	// 添加错误处理中间件（必须在最后）
	a.ginEngine.Use(middleware.ErrorHandler())

	// 设置自定义错误处理器
	a.ginEngine.NoRoute(middleware.NotFoundHandler)
	a.ginEngine.NoMethod(middleware.MethodNotAllowedHandler)
}
