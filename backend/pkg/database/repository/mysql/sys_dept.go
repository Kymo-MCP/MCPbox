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

var SysDeptRepo *SysDeptRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysDeptRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_dept table: %v", err))
		}
	})
}

// SysDeptRepository 部门仓库
type SysDeptRepository struct{}

// NewSysDeptRepository 创建部门仓库实例
func NewSysDeptRepository() *SysDeptRepository {
	SysDeptRepo = &SysDeptRepository{}
	return SysDeptRepo
}

// getDB 获取数据库连接
func (r *SysDeptRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysDept{})
}

// Create 创建部门
func (r *SysDeptRepository) Create(ctx context.Context, dept *model.SysDept) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 创建前的准备工作
	if err := dept.PrepareForCreate(); err != nil {
		return fmt.Errorf("prepare for create failed: %v", err)
	}

	// 验证数据
	if err := dept.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Creating department",
		zap.String("name", dept.Name),
		zap.String("source", string(dept.Source)),
		zap.Int("sort", dept.DeptSort))

	err := r.getDB().WithContext(ctx).Create(dept).Error
	if err != nil {
		logger.Error("Failed to create department",
			zap.String("name", dept.Name),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created department",
		zap.Uint("id", dept.DeptID),
		zap.String("name", dept.Name))

	return nil
}

// Update 更新部门
func (r *SysDeptRepository) Update(ctx context.Context, dept *model.SysDept) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 更新前的准备工作
	if err := dept.PrepareForUpdate(); err != nil {
		return fmt.Errorf("prepare for update failed: %v", err)
	}

	// 验证数据
	if err := dept.ValidateForUpdate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Updating department",
		zap.Uint("id", dept.DeptID),
		zap.String("name", dept.Name))

	err := r.getDB().WithContext(ctx).Save(dept).Error
	if err != nil {
		logger.Error("Failed to update department",
			zap.Uint("id", dept.DeptID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully updated department",
		zap.Uint("id", dept.DeptID),
		zap.String("name", dept.Name))

	return nil
}

// Delete 删除部门（物理删除）
func (r *SysDeptRepository) Delete(ctx context.Context, id uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	logger.Info("Deleting department", zap.Uint("id", id))

	err := r.getDB().WithContext(ctx).Delete(&model.SysDept{}, id).Error
	if err != nil {
		logger.Error("Failed to delete department",
			zap.Uint("id", id),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted department", zap.Uint("id", id))
	return nil
}

// FindByID 根据ID查找部门
func (r *SysDeptRepository) FindByID(ctx context.Context, id uint) (*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var dept model.SysDept
	err := r.getDB().WithContext(ctx).Where("dept_id = ?", id).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to find department: %v", err)
	}
	return &dept, nil
}

// FindByName 根据名称查找部门
func (r *SysDeptRepository) FindByName(ctx context.Context, name string) (*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if name == "" {
		return nil, fmt.Errorf("department name cannot be empty")
	}

	var dept model.SysDept
	err := r.getDB().WithContext(ctx).Where("name = ?", name).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found: %s", name)
		}
		return nil, fmt.Errorf("failed to find department by name '%s': %v", name, err)
	}
	return &dept, nil
}

// FindByParentID 根据上级部门ID查找子部门
func (r *SysDeptRepository) FindByParentID(ctx context.Context, parentID uint) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("pid = ?", parentID).
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find departments by parent id %d: %v", parentID, err)
	}
	return depts, nil
}

// FindRootDepts 查找根部门（没有上级部门的部门）
func (r *SysDeptRepository) FindRootDepts(ctx context.Context) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("pid IS NULL OR pid = 0").
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find root departments: %v", err)
	}
	return depts, nil
}

// FindAll 查找所有部门
func (r *SysDeptRepository) FindAll(ctx context.Context) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all departments: %v", err)
	}
	return depts, nil
}

// FindByEnabled 根据启用状态查找部门
func (r *SysDeptRepository) FindByEnabled(ctx context.Context, enabled bool) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("enabled = ?", enabled).
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find departments by enabled status %v: %v", enabled, err)
	}
	return depts, nil
}

// FindBySource 根据部门来源查找部门
func (r *SysDeptRepository) FindBySource(ctx context.Context, source model.DeptSource) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("source = ?", source).
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find departments by source %s: %v", source, err)
	}
	return depts, nil
}

// FindByCorpID 根据企业ID查找部门（用于第三方集成）
func (r *SysDeptRepository) FindByCorpID(ctx context.Context, corpID string) ([]*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if corpID == "" {
		return nil, fmt.Errorf("corp id cannot be empty")
	}

	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("corp_id = ?", corpID).
		Order("dept_sort ASC, dept_id ASC").
		Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find departments by corp id %s: %v", corpID, err)
	}
	return depts, nil
}

// FindByOpenDepartmentID 根据第三方部门ID查找部门
func (r *SysDeptRepository) FindByOpenDepartmentID(ctx context.Context, openDeptID string) (*model.SysDept, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if openDeptID == "" {
		return nil, fmt.Errorf("open department id cannot be empty")
	}

	var dept model.SysDept
	err := r.getDB().WithContext(ctx).
		Where("open_department_id = ?", openDeptID).
		First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found with open department id: %s", openDeptID)
		}
		return nil, fmt.Errorf("failed to find department by open department id %s: %v", openDeptID, err)
	}
	return &dept, nil
}

// UpdateSubCount 更新子部门数量
func (r *SysDeptRepository) UpdateSubCount(ctx context.Context, parentID uint, count int) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	err := r.getDB().WithContext(ctx).
		Model(&model.SysDept{}).
		Where("dept_id = ?", parentID).
		Update("sub_count", count).Error
	if err != nil {
		return fmt.Errorf("failed to update sub count for department %d: %v", parentID, err)
	}
	return nil
}

// UpdateEnabled 更新部门启用状态
func (r *SysDeptRepository) UpdateEnabled(ctx context.Context, id uint, enabled bool) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysDept{}).
		Where("dept_id = ?", id).
		Updates(map[string]interface{}{
			"enabled":     enabled,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update enabled status for department %d: %v", id, err)
	}
	return nil
}

// CountByParentID 统计指定上级部门的子部门数量
func (r *SysDeptRepository) CountByParentID(ctx context.Context, parentID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysDept{}).
		Where("pid = ?", parentID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count departments by parent id %d: %v", parentID, err)
	}
	return count, nil
}

// ExistsByName 检查指定名称的部门是否存在
func (r *SysDeptRepository) ExistsByName(ctx context.Context, name string, excludeID ...uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if name == "" {
		return false, fmt.Errorf("department name cannot be empty")
	}

	query := r.getDB().WithContext(ctx).Model(&model.SysDept{}).Where("name = ?", name)

	// 排除指定ID（用于更新时检查重名）
	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("dept_id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check department name existence: %v", err)
	}
	return count > 0, nil
}

// HealthCheck 检查数据库连接健康状态
func (r *SysDeptRepository) HealthCheck(ctx context.Context) error {
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
func (r *SysDeptRepository) InitTable() error {
	// 创建表
	mod := &model.SysDept{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
	}{
		{"idx_pid", "pid"},
		{"idx_enabled", "enabled"},
		{"idx_name", "name"},
		{"idx_source", "source"},
		{"idx_corp_id", "corp_id"},
		{"idx_open_department_id", "open_department_id"},
		{"idx_dept_sort", "dept_sort"},
	}

	for _, idx := range indexes {
		var count int64
		r.getDB().Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
			mod.TableName(), idx.name).Scan(&count)

		if count == 0 {
			if err := r.getDB().Exec(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.name, mod.TableName(), idx.column)).Error; err != nil {
				return fmt.Errorf("failed to create index %s: %v", idx.name, err)
			}
		}
	}

	return nil
}
