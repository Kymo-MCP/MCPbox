package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var McpEnvironmentRepo *McpEnvironmentRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewMcpEnvironmentRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_environment table: %v", err))
		}
	})
}

// McpEnvironmentRepository MCP环境仓库
type McpEnvironmentRepository struct{}

// NewMcpEnvironmentRepository 创建MCP环境仓库实例
func NewMcpEnvironmentRepository() *McpEnvironmentRepository {
	McpEnvironmentRepo = &McpEnvironmentRepository{}
	return McpEnvironmentRepo
}

func (r *McpEnvironmentRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpEnvironment{})
}

// Create 创建MCP环境
func (r *McpEnvironmentRepository) Create(ctx context.Context, environment *model.McpEnvironment) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	logger.Info("Creating MCP environment",
		zap.String("name", environment.Name),
		zap.String("environment", string(environment.Environment)),
		zap.String("namespace", environment.Namespace))

	err := r.getDB().WithContext(ctx).Create(environment).Error
	if err != nil {
		logger.Error("Failed to create MCP environment",
			zap.String("name", environment.Name),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created MCP environment",
		zap.Uint("id", environment.ID),
		zap.String("name", environment.Name))

	return nil
}

// Update 更新MCP环境
func (r *McpEnvironmentRepository) Update(ctx context.Context, environment *model.McpEnvironment) error {
	environment.PrepareForUpdate()
	return r.getDB().WithContext(ctx).Where("id = ?", environment.ID).Save(environment).Error
}

// Delete 删除MCP环境（软删除）
func (r *McpEnvironmentRepository) Delete(ctx context.Context, id uint) error {
	updateMod := &model.McpEnvironment{
		ID:        id,
		IsDeleted: true,
		UpdatedAt: time.Now(),
	}
	return r.getDB().WithContext(ctx).Model(&model.McpEnvironment{}).
		Where("id = ?", id).
		Updates(updateMod).Error
}

// FindByID 根据ID查找MCP环境（排除已删除）
func (r *McpEnvironmentRepository) FindByID(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var environment model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&environment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("environment not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to find environment: %v", err)
	}
	return &environment, nil
}

// FindNamesByIDs 根据环境ID列表查找环境名称
func (r *McpEnvironmentRepository) FindNamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var environments []*model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("id IN ?", ids).
		Find(&environments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find environments by ids: %v", err)
	}

	names := make(map[string]string)
	for _, env := range environments {
		names[fmt.Sprintf("%d", env.ID)] = env.Name
	}
	return names, nil
}

// FindByName 根据名称查找MCP环境（排除已删除）
func (r *McpEnvironmentRepository) FindByName(ctx context.Context, name string) (*model.McpEnvironment, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// 参数验证
	if name == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	var environment model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("name = ? AND is_deleted = ?", name, false).
		First(&environment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("environment not found: %s", name)
		}
		// 记录详细的数据库错误信息
		return nil, fmt.Errorf("failed to find environment by name '%s': %v", name, err)
	}
	return &environment, nil
}

// FindAll 查找所有MCP环境（排除已删除）
func (r *McpEnvironmentRepository) FindAll(ctx context.Context) ([]*model.McpEnvironment, error) {
	var environments []*model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("is_deleted = ?", false).
		Find(&environments).Error
	if err != nil {
		return nil, err
	}
	return environments, nil
}

// FindByEnvironment 根据环境类型查找MCP环境（排除已删除）
func (r *McpEnvironmentRepository) FindByEnvironment(ctx context.Context, environmentType model.McpEnvironmentType) ([]*model.McpEnvironment, error) {
	var environments []*model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("environment = ? AND is_deleted = ?", environmentType, false).
		Find(&environments).Error
	if err != nil {
		return nil, err
	}
	return environments, nil
}

// FindDeletedByID 根据ID查找已删除的MCP环境
func (r *McpEnvironmentRepository) FindDeletedByID(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	var environment model.McpEnvironment
	err := r.getDB().WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, true).
		First(&environment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("deleted environment not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to find deleted environment: %v", err)
	}
	return &environment, nil
}

// FindAllWithDeleted 查找所有MCP环境（包括已删除）
func (r *McpEnvironmentRepository) FindAllWithDeleted(ctx context.Context) ([]*model.McpEnvironment, error) {
	var environments []*model.McpEnvironment
	err := r.getDB().WithContext(ctx).Find(&environments).Error
	if err != nil {
		return nil, err
	}
	return environments, nil
}

// RestoreEnvironment 恢复已删除的环境
func (r *McpEnvironmentRepository) RestoreEnvironment(ctx context.Context, id uint) error {
	return r.getDB().WithContext(ctx).Model(&model.McpEnvironment{}).
		Where("id = ? AND is_deleted = ?", id, true).
		Updates(map[string]interface{}{
			"is_deleted": false,
			"updated_at": time.Now(),
		}).Error
}

// HealthCheck 检查数据库连接健康状态
func (r *McpEnvironmentRepository) HealthCheck(ctx context.Context) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := r.getDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	return sqlDB.PingContext(ctx)
}

// InitTable 初始化表结构
func (r *McpEnvironmentRepository) InitTable() error {
	// 创建表
	mod := &model.McpEnvironment{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查索引是否存在
	var count int64
	r.getDB().Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		mod.TableName(), "idx_name").Scan(&count)

	// 如果索引不存在，则创建
	if count == 0 {
		if err := r.getDB().Exec(fmt.Sprintf("CREATE INDEX idx_name ON %s (name)", mod.TableName())).Error; err != nil {
			return fmt.Errorf("failed to create index idx_name: %v", err)
		}
	}

	// 检查环境类型索引是否存在
	r.getDB().Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		mod.TableName(), "idx_environment").Scan(&count)

	if count == 0 {
		if err := r.getDB().Exec(fmt.Sprintf("CREATE INDEX idx_environment ON %s (environment)", mod.TableName())).Error; err != nil {
			return fmt.Errorf("failed to create index idx_environment: %v", err)
		}
	}

	// 检查删除状态索引是否存在
	r.getDB().Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		mod.TableName(), "idx_is_deleted").Scan(&count)

	if count == 0 {
		if err := r.getDB().Exec(fmt.Sprintf("CREATE INDEX idx_is_deleted ON %s (is_deleted)", mod.TableName())).Error; err != nil {
			return fmt.Errorf("failed to create index idx_is_deleted: %v", err)
		}
	}

	return nil
}
