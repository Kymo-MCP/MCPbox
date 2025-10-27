package mysql

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	once   sync.Once
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
)

// Config MySQL配置
type Config struct {
	Host                string        `validate:"required"`
	Port                int           `validate:"required,min=1,max=65535"`
	Username            string        `validate:"required"`
	Password            string        `validate:"required"`
	Database            string        `validate:"required"`
	ConnectTimeout      time.Duration `validate:"required"`
	MaxIdleConns        int           `validate:"min=0"`
	MaxOpenConns        int           `validate:"min=0"`
	HealthCheckInterval time.Duration `validate:"required"`
	MaxRetries          int           `validate:"min=0"`
	RetryInterval       time.Duration `validate:"required"`
}

// InitHook 初始化钩子函数类型
type InitHook func(db *gorm.DB)

// HookManager 钩子管理器
type HookManager struct {
	hooks []InitHook
	mu    sync.RWMutex
}

var hookManager = NewHookManager()

// NewHookManager 创建钩子管理器
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]InitHook, 0),
	}
}

// Register 注册钩子
func (m *HookManager) Register(hook InitHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// CallHooks 调用所有钩子
func (m *HookManager) CallHooks(db *gorm.DB) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.hooks {
		hook(db)
	}
}

// RegisterInit 注册初始化钩子
func RegisterInit(initHook InitHook) {
	hookManager.Register(initHook)
}

// InitDB 初始化数据库连接
func InitDB(config *Config) error {
	if config == nil {
		return errors.New("no mysql config")
	}

	var initErr error
	once.Do(func() {
		// 创建带有取消功能的上下文
		ctx, cancel = context.WithCancel(context.Background())

		// 初始化数据库连接
		if err := initConnection(config); err != nil {
			initErr = err
			return
		}

		// 启动健康检查和自动重连的goroutine
		startHealthChecker(config.HealthCheckInterval, config.MaxRetries, config.RetryInterval)
	})

	return initErr
}

// initConnection 初始化数据库连接
func initConnection(config *Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	var err error
	db, err = gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if healthErr := HealthCheck(); healthErr != nil {
		return fmt.Errorf("failed to health check: %v", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// 调用所有初始化钩子
	hookManager.CallHooks(db)

	return nil
}

// startHealthChecker 启动健康检查和自动重连
func startHealthChecker(interval time.Duration, maxRetries int, retryInterval time.Duration) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := HealthCheck(); err != nil {
					fmt.Printf("Database health check failed: %v\n", err)
					if err := reconnect(maxRetries, retryInterval); err != nil {
						fmt.Printf("Failed to reconnect to database: %v\n", err)
					} else {
						fmt.Println("Successfully reconnected to database")
					}
				}
			}
		}
	}()
}

// reconnect 尝试重新连接数据库
func reconnect(maxRetries int, retryInterval time.Duration) error {
	mu.Lock()
	defer mu.Unlock()

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Retrying to connect to database (attempt %d/%d)...\n", i, maxRetries)
			time.Sleep(retryInterval)
		}

		// 获取当前配置
		sqlDB, err := db.DB()
		if err == nil {
			// 尝试关闭现有连接
			_ = sqlDB.Close()
		}

		// 重新建立连接
		if err := HealthCheck(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %v", maxRetries, lastErr)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return db
}

// HealthCheck 健康检查
func HealthCheck() error {
	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		return errors.New("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	return sqlDB.Ping()
}

// Close 关闭数据库连接
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	// 取消健康检查的goroutine
	if cancel != nil {
		cancel()
	}

	// 等待所有goroutine结束
	wg.Wait()

	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	return sqlDB.Close()
}
