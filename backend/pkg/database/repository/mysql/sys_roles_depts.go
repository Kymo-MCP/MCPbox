package mysql

import (
	"context"
	"fmt"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var SysRolesDeptsRepo *SysRolesDeptsRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysRolesDeptsRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_roles_depts table: %v", err))
		}
	})
}

// SysRolesDeptsRepository 角色部门关联仓库
type SysRolesDeptsRepository struct{}

// NewSysRolesDeptsRepository 创建角色部门关联仓库实例
func NewSysRolesDeptsRepository(db *gorm.DB) *SysRolesDeptsRepository {
	SysRolesDeptsRepo = &SysRolesDeptsRepository{}
	return SysRolesDeptsRepo
}

// getDB 获取数据库连接
func (r *SysRolesDeptsRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysRolesDepts{})
}

// Create 创建角色部门关联
func (r *SysRolesDeptsRepository) Create(ctx context.Context, rolesDepts *model.SysRolesDepts) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 验证数据
	if err := rolesDepts.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Creating role-dept association",
		zap.Uint("roleId", rolesDepts.RoleID),
		zap.Uint("deptId", rolesDepts.DeptID))

	err := r.getDB().WithContext(ctx).Create(rolesDepts).Error
	if err != nil {
		logger.Error("Failed to create role-dept association",
			zap.Uint("roleId", rolesDepts.RoleID),
			zap.Uint("deptId", rolesDepts.DeptID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created role-dept association",
		zap.Uint("roleId", rolesDepts.RoleID),
		zap.Uint("deptId", rolesDepts.DeptID))

	return nil
}

// Delete 删除角色部门关联
func (r *SysRolesDeptsRepository) Delete(ctx context.Context, roleID, deptID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}
	if deptID == 0 {
		return fmt.Errorf("部门ID不能为空")
	}

	logger.Info("Deleting role-dept association",
		zap.Uint("roleId", roleID),
		zap.Uint("deptId", deptID))

	err := r.getDB().WithContext(ctx).
		Where("role_id = ? AND dept_id = ?", roleID, deptID).
		Delete(&model.SysRolesDepts{}).Error
	if err != nil {
		logger.Error("Failed to delete role-dept association",
			zap.Uint("roleId", roleID),
			zap.Uint("deptId", deptID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted role-dept association",
		zap.Uint("roleId", roleID),
		zap.Uint("deptId", deptID))

	return nil
}

// DeleteByRoleID 删除指定角色的所有部门关联
func (r *SysRolesDeptsRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}

	logger.Info("Deleting all dept associations for role", zap.Uint("roleId", roleID))

	err := r.getDB().WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&model.SysRolesDepts{}).Error
	if err != nil {
		logger.Error("Failed to delete role-dept associations by role",
			zap.Uint("roleId", roleID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted all dept associations for role", zap.Uint("roleId", roleID))
	return nil
}

// DeleteByDeptID 删除指定部门的所有角色关联
func (r *SysRolesDeptsRepository) DeleteByDeptID(ctx context.Context, deptID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return fmt.Errorf("部门ID不能为空")
	}

	logger.Info("Deleting all role associations for dept", zap.Uint("deptId", deptID))

	err := r.getDB().WithContext(ctx).
		Where("dept_id = ?", deptID).
		Delete(&model.SysRolesDepts{}).Error
	if err != nil {
		logger.Error("Failed to delete role-dept associations by dept",
			zap.Uint("deptId", deptID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted all role associations for dept", zap.Uint("deptId", deptID))
	return nil
}

// FindByRoleID 根据角色ID查找部门关联
func (r *SysRolesDeptsRepository) FindByRoleID(ctx context.Context, roleID uint) ([]*model.SysRolesDepts, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	var associations []*model.SysRolesDepts
	err := r.getDB().WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&associations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find dept associations for role %d: %v", roleID, err)
	}
	return associations, nil
}

// FindByDeptID 根据部门ID查找角色关联
func (r *SysRolesDeptsRepository) FindByDeptID(ctx context.Context, deptID uint) ([]*model.SysRolesDepts, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return nil, fmt.Errorf("部门ID不能为空")
	}

	var associations []*model.SysRolesDepts
	err := r.getDB().WithContext(ctx).
		Where("dept_id = ?", deptID).
		Find(&associations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role associations for dept %d: %v", deptID, err)
	}
	return associations, nil
}

// FindDeptIDsByRoleID 根据角色ID查找关联的部门ID列表
func (r *SysRolesDeptsRepository) FindDeptIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	var deptIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRolesDepts{}).
		Where("role_id = ?", roleID).
		Pluck("dept_id", &deptIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find dept IDs for role %d: %v", roleID, err)
	}
	return deptIDs, nil
}

// FindRoleIDsByDeptID 根据部门ID查找关联的角色ID列表
func (r *SysRolesDeptsRepository) FindRoleIDsByDeptID(ctx context.Context, deptID uint) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return nil, fmt.Errorf("部门ID不能为空")
	}

	var roleIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRolesDepts{}).
		Where("dept_id = ?", deptID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role IDs for dept %d: %v", deptID, err)
	}
	return roleIDs, nil
}

// Exists 检查角色部门关联是否存在
func (r *SysRolesDeptsRepository) Exists(ctx context.Context, roleID, deptID uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return false, fmt.Errorf("角色ID不能为空")
	}
	if deptID == 0 {
		return false, fmt.Errorf("部门ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRolesDepts{}).
		Where("role_id = ? AND dept_id = ?", roleID, deptID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check role-dept association existence: %v", err)
	}
	return count > 0, nil
}

// BatchCreate 批量创建角色部门关联
func (r *SysRolesDeptsRepository) BatchCreate(ctx context.Context, associations []*model.SysRolesDepts) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if len(associations) == 0 {
		return fmt.Errorf("关联列表不能为空")
	}

	// 验证所有关联
	for i, assoc := range associations {
		if err := assoc.ValidateForCreate(); err != nil {
			return fmt.Errorf("validation failed for association %d: %v", i, err)
		}
	}

	logger.Info("Batch creating role-dept associations", zap.Int("count", len(associations)))

	err := r.getDB().WithContext(ctx).CreateInBatches(associations, 100).Error
	if err != nil {
		logger.Error("Failed to batch create role-dept associations",
			zap.Int("count", len(associations)),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully batch created role-dept associations", zap.Int("count", len(associations)))
	return nil
}

// BatchDeleteByRoleID 批量删除指定角色的部门关联，并创建新的关联
func (r *SysRolesDeptsRepository) BatchUpdateByRoleID(ctx context.Context, roleID uint, deptIDs []uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}

	return r.getDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("role_id = ?", roleID).Delete(&model.SysRolesDepts{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing associations: %v", err)
		}

		// 创建新关联
		if len(deptIDs) > 0 {
			var associations []*model.SysRolesDepts
			for _, deptID := range deptIDs {
				associations = append(associations, &model.SysRolesDepts{
					RoleID: roleID,
					DeptID: deptID,
				})
			}
			if err := tx.CreateInBatches(associations, 100).Error; err != nil {
				return fmt.Errorf("failed to create new associations: %v", err)
			}
		}

		return nil
	})
}

// CountByRoleID 统计指定角色的部门关联数量
func (r *SysRolesDeptsRepository) CountByRoleID(ctx context.Context, roleID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return 0, fmt.Errorf("角色ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRolesDepts{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count dept associations for role %d: %v", roleID, err)
	}
	return count, nil
}

// CountByDeptID 统计指定部门的角色关联数量
func (r *SysRolesDeptsRepository) CountByDeptID(ctx context.Context, deptID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return 0, fmt.Errorf("部门ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysRolesDepts{}).
		Where("dept_id = ?", deptID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count role associations for dept %d: %v", deptID, err)
	}
	return count, nil
}

// FindAll 查找所有角色部门关联
func (r *SysRolesDeptsRepository) FindAll(ctx context.Context) ([]*model.SysRolesDepts, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var associations []*model.SysRolesDepts
	err := r.getDB().WithContext(ctx).
		Order("role_id ASC, dept_id ASC").
		Find(&associations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all role-dept associations: %v", err)
	}
	return associations, nil
}

// HealthCheck 检查数据库连接健康状态
func (r *SysRolesDeptsRepository) HealthCheck(ctx context.Context) error {
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
func (r *SysRolesDeptsRepository) InitTable() error {
	// 创建表
	mod := &model.SysRolesDepts{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
	}{
		{"idx_role_id", "role_id"},
		{"idx_dept_id", "dept_id"},
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
