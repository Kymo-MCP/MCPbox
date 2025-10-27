package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qm-mcp-server/internal/gateway/config"
	"qm-mcp-server/pkg/database"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/logger"

	"go.uber.org/zap"
)

// App 应用程序结构体
type App struct {
	// config 配置
	config *config.Config

	// logger 日志记录器
	logger *zap.Logger

	// httpServer HTTP服务器
	httpServer *http.Server

	// shutdownCtx 关闭上下文
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New 创建新的应用程序实例
func New() (*App, error) {
	// 加载全局配置
	err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	// 打印配置信息
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", config.GlobalConfig.VersionInfo)))

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config:         config.GlobalConfig,
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

	// 使用全局数据库仓库实例（已在init中初始化）
	if mysql.McpInstanceRepo == nil {
		return fmt.Errorf("McpInstanceRepo 未正确初始化，请检查数据库初始化流程")
	}

	// 初始化 HTTP 服务器
	if err := a.initializeHTTPServer(); err != nil {
		return fmt.Errorf("初始化HTTP服务器失败: %w", err)
	}

	a.logger.Info("应用程序初始化完成")
	return nil
}

// initializeHTTPServer 初始化HTTP服务器
func (a *App) initializeHTTPServer() error {
	// 初始化 Gin 引擎
	r := NewServer()

	// 创建 HTTP 服务器
	serverAddr := fmt.Sprintf(":%d", config.GlobalConfig.Server.HttpPort)
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	return nil
}

// Run 运行应用程序
func (a *App) Run() error {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
