package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/internal/init/config"
	dbpkg "github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// App application structure
type App struct {
	config          *config.InitConfig
	logger          *zap.Logger
	adminUser       *model.SysUser
	codePackageList []*model.McpCodePackage
	mcpTemplateList []*model.McpTemplate
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
		config:          config.GlobalConfig,
		logger:          logger.L().Logger,
		adminUser:       &model.SysUser{},
		codePackageList: make([]*model.McpCodePackage, 0),
		mcpTemplateList: make([]*model.McpTemplate, 0),
	}
}

// Initialize initializes the application
func (a *App) Initialize() error {
	// 初始化数据库
	if err := dbpkg.Init(&a.config.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("Authz service initialized successfully")
	return nil
}

// Run 运行应用程序
func (a *App) Run() error {
	// 拷贝项目路径下 data 目录所有基础数据到挂载根目录中
	if err := a.copyInitData("./init-data/static", a.config.Storage.StaticPath); err != nil {
		return fmt.Errorf("failed to copy data directory: %w", err)
	}

	// 创建管理员用户
	adminUser, err := a.createAdminUser()
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// init data scope
	if err := a.initDataScope(context.Background(), adminUser); err != nil {
		return fmt.Errorf("failed to init data scope: %w", err)
	}

	logger.Info("Shutting down authz service...")

	// 优雅关闭
	return a.Shutdown()
}

// Shutdown 优雅关闭应用程序
func (a *App) Shutdown() error {
	// 关闭数据库连接
	if err := mysql.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	logger.Info("Authz service shutdown completed")
	return nil
}

// initDataScope creates the default environment
func (a *App) initDataScope(ctx context.Context, adminUser *model.SysUser) error {
	// 初始化默认 Kubernetes 环境
	envMod, err := a.initDefaultKubernetesEnvironment(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("failed to init default kubernetes environment: %w", err)
	}
	// 初始化代码包数据
	if err := a.initCodePackage(ctx); err != nil {
		return fmt.Errorf("failed to init code package data: %w", err)
	}
	// 初始化 MCP 模板数据（使用嵌入式模板 JSON）
	if err := a.initMcpTemplateData(ctx, envMod); err != nil {
		return fmt.Errorf("failed to init mcp template data: %w", err)
	}
	return nil
}
