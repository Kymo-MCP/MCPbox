package mysql

import (
	"context"
	"fmt"
	"time"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var SysRoleRepo *SysRoleRepository

// SysRoleRepository 角色仓库
type SysRoleRepository struct{}

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysRoleRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_role table: %v", err))
		}
	})
}

// NewSysRoleRepository 创建角色仓库实例
func NewSysRoleRepository() *SysRoleRepository {
	SysRoleRepo = &SysRoleRepository{}
	return SysRoleRepo
}

// getDB 获取数据库连接
func (r *SysRoleRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysRole{})
}

// Create 创建角色
func (r *SysRoleRepository) Create(ctx context.Context, role *model.SysRole) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 创建前的准备工作
	if err := role.PrepareForCreate(); err != nil {
		return fmt.Errorf("prepare for create failed: %v", err)
	}

	// 验证数据
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Creating role",
		zap.String("name", role.Name),
		zap.String("dataScope", role.GetDataScope()),
		zap.Int("level", role.GetLevel()))

	err := r.getDB().WithContext(ctx).Create(role).Error
	if err != nil {
		logger.Error("Failed to create role",
			zap.String("name", role.Name),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created role",
		zap.Uint("id", role.RoleID),
		zap.String("name", role.Name))

	return nil
}

// Update 更新角色
func (r *SysRoleRepository) Update(ctx context.Context, role *model.SysRole) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 更新前的准备工作
	if err := role.PrepareForUpdate(); err != nil {
		return fmt.Errorf("prepare for update failed: %v", err)
	}

	// 验证数据
	if err := role.ValidateForUpdate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Updating role",
		zap.Uint("id", role.RoleID),
		zap.String("name", role.Name))

	err := r.getDB().WithContext(ctx).Save(role).Error
	if err != nil {
		logger.Error("Failed to update role",
			zap.Uint("id", role.RoleID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully updated role",
		zap.Uint("id", role.RoleID),
		zap.String("name", role.Name))

	return nil
}

// Delete 删除角色（物理删除）
func (r *SysRoleRepository) Delete(ctx context.Context, id uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	logger.Info("Deleting role", zap.Uint("id", id))

	err := r.getDB().WithContext(ctx).Delete(&model.SysRole{}, id).Error
	if err != nil {
		logger.Error("Failed to delete role",
			zap.Uint("id", id),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted role", zap.Uint("id", id))
	return nil
}

// FindByID 根据ID查找角色
func (r *SysRoleRepository) FindByID(ctx context.Context, id uint) (*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var role model.SysRole
	err := r.getDB().WithContext(ctx).Where("role_id = ?", id).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to find role: %v", err)
	}
	return &role, nil
}

// FindByName 根据名称查找角色
func (r *SysRoleRepository) FindByName(ctx context.Context, name string) (*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	var role model.SysRole
	err := r.getDB().WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found: %s", name)
		}
		return nil, fmt.Errorf("failed to find role by name '%s': %v", name, err)
	}
	return &role, nil
}

// FindAll 查找所有角色
func (r *SysRoleRepository) FindAll(ctx context.Context) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all roles: %v", err)
	}
	return roles, nil
}

// FindByLevel 根据角色级别查找角色
func (r *SysRoleRepository) FindByLevel(ctx context.Context, level int) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("level = ?", level).
		Order("role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles by level %d: %v", level, err)
	}
	return roles, nil
}

// FindByLevelRange 根据角色级别范围查找角色
func (r *SysRoleRepository) FindByLevelRange(ctx context.Context, minLevel, maxLevel int) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("level >= ? AND level <= ?", minLevel, maxLevel).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles by level range [%d, %d]: %v", minLevel, maxLevel, err)
	}
	return roles, nil
}

// FindByDataScope 根据数据权限范围查找角色
func (r *SysRoleRepository) FindByDataScope(ctx context.Context, dataScope string) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if dataScope == "" {
		return nil, fmt.Errorf("data scope cannot be empty")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("data_scope = ?", dataScope).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles by data scope %s: %v", dataScope, err)
	}
	return roles, nil
}

// FindByCreator 根据创建者查找角色
func (r *SysRoleRepository) FindByCreator(ctx context.Context, creator string) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if creator == "" {
		return nil, fmt.Errorf("creator cannot be empty")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("create_by = ?", creator).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles by creator %s: %v", creator, err)
	}
	return roles, nil
}

// FindHigherLevelRoles 查找级别高于指定级别的角色
func (r *SysRoleRepository) FindHigherLevelRoles(ctx context.Context, level int) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("level > ?", level).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find higher level roles than %d: %v", level, err)
	}
	return roles, nil
}

// FindLowerLevelRoles 查找级别低于指定级别的角色
func (r *SysRoleRepository) FindLowerLevelRoles(ctx context.Context, level int) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("level < ?", level).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find lower level roles than %d: %v", level, err)
	}
	return roles, nil
}

// FindRolesWithoutLevel 查找未设置级别的角色
func (r *SysRoleRepository) FindRolesWithoutLevel(ctx context.Context) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).
		Where("level IS NULL").
		Order("role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles without level: %v", err)
	}
	return roles, nil
}

// UpdateLevel 更新角色级别
func (r *SysRoleRepository) UpdateLevel(ctx context.Context, id uint, level int) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Where("role_id = ?", id).
		Updates(map[string]interface{}{
			"level":       level,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update level for role %d: %v", id, err)
	}
	return nil
}

// UpdateDataScope 更新角色数据权限范围
func (r *SysRoleRepository) UpdateDataScope(ctx context.Context, id uint, dataScope string) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Where("role_id = ?", id).
		Updates(map[string]interface{}{
			"data_scope":  dataScope,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update data scope for role %d: %v", id, err)
	}
	return nil
}

// ExistsByName 检查指定名称的角色是否存在
func (r *SysRoleRepository) ExistsByName(ctx context.Context, name string, excludeID ...uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if name == "" {
		return false, fmt.Errorf("role name cannot be empty")
	}

	query := r.getDB().WithContext(ctx).Model(&model.SysRole{}).Where("name = ?", name)

	// 排除指定ID（用于更新时检查重名）
	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("role_id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check role name existence: %v", err)
	}
	return count > 0, nil
}

// CountByLevel 统计指定级别的角色数量
func (r *SysRoleRepository) CountByLevel(ctx context.Context, level int) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Where("level = ?", level).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count roles by level %d: %v", level, err)
	}
	return count, nil
}

// CountByDataScope 统计指定数据权限范围的角色数量
func (r *SysRoleRepository) CountByDataScope(ctx context.Context, dataScope string) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Where("data_scope = ?", dataScope).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count roles by data scope %s: %v", dataScope, err)
	}
	return count, nil
}

// GetMaxLevel 获取最高角色级别
func (r *SysRoleRepository) GetMaxLevel(ctx context.Context) (int, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var maxLevel int
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Select("COALESCE(MAX(level), 0)").
		Scan(&maxLevel).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get max level: %v", err)
	}
	return maxLevel, nil
}

// GetMinLevel 获取最低角色级别
func (r *SysRoleRepository) GetMinLevel(ctx context.Context) (int, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var minLevel int
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRole{}).
		Select("COALESCE(MIN(level), 0)").
		Scan(&minLevel).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get min level: %v", err)
	}
	return minLevel, nil
}

// SearchByKeyword 根据关键词搜索角色（名称或描述）
func (r *SysRoleRepository) SearchByKeyword(ctx context.Context, keyword string) ([]*model.SysRole, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if keyword == "" {
		return r.FindAll(ctx)
	}

	var roles []*model.SysRole
	searchPattern := "%" + keyword + "%"
	err := r.getDB().WithContext(ctx).
		Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern).
		Order("level DESC, role_id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search roles by keyword '%s': %v", keyword, err)
	}
	return roles, nil
}

// HealthCheck 检查数据库连接健康状态
func (r *SysRoleRepository) HealthCheck(ctx context.Context) error {
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
func (r *SysRoleRepository) InitTable() error {
	// 创建表
	mod := &model.SysRole{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
	}{
		{"idx_level", "level"},
		{"idx_data_scope", "data_scope"},
		{"idx_create_by", "create_by"},
		{"idx_create_time", "create_time"},
		{"idx_update_time", "update_time"},
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
