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

// App 应用程序结构
type App struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	ginEngine  *gin.Engine
}

// New 创建应用程序实例
func New() *App {
	// 加载配置
	if err := config.Load(); err != nil {
		return nil
	}

	// 初始化日志
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil
	}

	// 打印配置信息
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", config.GlobalConfig.VersionInfo)))

	return &App{
		config: config.GlobalConfig,
		logger: logger.L().Logger,
	}
}

// Initialize 初始化应用程序
func (a *App) Initialize() error {
	// 初始化Redis
	if err := redis.Init(&config.GlobalConfig.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	// 初始化数据库
	if err := dbpkg.Init(&a.config.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// 设置 HTTP 服务器
	if err := a.setupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	logger.Info("Authz service initialized successfully")
	return nil
}

// setupHTTPServer 设置 HTTP 服务器
func (a *App) setupHTTPServer() error {
	// 设置Gin模式
	gin.SetMode(gin.DebugMode)

	// 创建Gin引擎
	a.ginEngine = gin.New()

	// 设置中间件
	a.setupMiddleware()

	// 设置路由
	a.setupRoutes()

	// 创建HTTP服务器
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Server.HttpPort),
		Handler: a.ginEngine,
	}

	logger.Info("HTTP server setup completed")
	return nil
}

// setupMiddleware 设置中间件
func (a *App) setupMiddleware() {
	// 添加中间件
	a.ginEngine.Use(gin.Recovery())
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// 添加跨域处理
	a.ginEngine.Use(middleware.CORSMiddleware([]string{"*"}))

	// 添加国际化中间件
	a.ginEngine.Use(middleware.I18nMiddleware())

	// 添加安全中间件
	a.ginEngine.Use(middleware.SecurityMiddleware(config.GlobalConfig.Secret))

	// 添加认证中间件
	a.ginEngine.Use(middleware.AuthTokenMiddleware(config.GlobalConfig.Secret))
}

// setupRoutes 设置路由
func (a *App) setupRoutes() {
	// 创建服务实例
	userService := service.NewUserService()

	userAuthService := service.NewUserAuthService()

	// 健康检查
	a.ginEngine.GET("/health", func(c *gin.Context) {
		common.GinSuccess(c, map[string]string{"status": "ok"})
	})

	// API版本前缀
	authzGroup := a.ginEngine.Group(common.GetAuthzRoutePrefix())

	// 用户相关路由
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

	// 认证相关路由 - 更新为使用UserAuthService

	{
		// 用户登录
		authzGroup.POST("/login", userAuthService.Login)

		// 用户退出
		authzGroup.POST("/logout", userAuthService.Logout)

		// 获取用户配置
		authzGroup.GET("/user-info", userAuthService.GetUserInfo)

		// 刷新Token
		authzGroup.POST("/refresh", userAuthService.RefreshToken)

		// 校验Token
		authzGroup.POST("/validate", userAuthService.ValidateToken)

		// 获取加密密钥
		authzGroup.POST("/encryption-key", userAuthService.GetEncryptionKey)
	}
}

// Run 运行应用程序
func (a *App) Run() error {
	// 启动 HTTP 服务器
	logger.Info("Starting HTTP server", zap.Int("port", a.config.Server.HttpPort))
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down authz service...")

	// 优雅关闭
	return a.Shutdown()
}

// Shutdown 优雅关闭应用程序
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if a.httpServer != nil {
		logger.Info("Shutting down HTTP server...")
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}
	}

	// 关闭数据库连接
	if err := mysql.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	logger.Info("Authz service shutdown completed")
	return nil
}
