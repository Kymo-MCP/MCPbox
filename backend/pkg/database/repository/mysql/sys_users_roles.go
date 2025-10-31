package mysql

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var SysUsersRolesRepo *SysUsersRolesRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysUsersRolesRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_users_roles table: %v", err))
		}
	})
}

// SysUsersRolesRepository 用户角色关联仓库
type SysUsersRolesRepository struct{}

// NewSysUsersRolesRepository 创建用户角色关联仓库实例
func NewSysUsersRolesRepository(db *gorm.DB) *SysUsersRolesRepository {
	SysUsersRolesRepo = &SysUsersRolesRepository{}
	return SysUsersRolesRepo
}

// getDB 获取数据库连接
func (r *SysUsersRolesRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysUsersRoles{})
}

// Create 创建用户角色关联
func (r *SysUsersRolesRepository) Create(ctx context.Context, userRole *model.SysUsersRoles) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 验证数据
	if err := userRole.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Creating user role association",
		zap.Uint("user_id", userRole.GetUserID()),
		zap.Uint("role_id", userRole.GetRoleID()))

	err := r.getDB().WithContext(ctx).Create(userRole).Error
	if err != nil {
		logger.Error("Failed to create user role association",
			zap.Uint("user_id", userRole.GetUserID()),
			zap.Uint("role_id", userRole.GetRoleID()),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created user role association",
		zap.Uint("user_id", userRole.GetUserID()),
		zap.Uint("role_id", userRole.GetRoleID()))

	return nil
}

// Delete 删除用户角色关联
func (r *SysUsersRolesRepository) Delete(ctx context.Context, userID, roleID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if userID == 0 || roleID == 0 {
		return fmt.Errorf("用户ID和角色ID不能为空")
	}

	logger.Info("Deleting user role association",
		zap.Uint("user_id", userID),
		zap.Uint("role_id", roleID))

	err := r.getDB().WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.SysUsersRoles{}).Error
	if err != nil {
		logger.Error("Failed to delete user role association",
			zap.Uint("user_id", userID),
			zap.Uint("role_id", roleID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted user role association",
		zap.Uint("user_id", userID),
		zap.Uint("role_id", roleID))

	return nil
}

// DeleteByUserID 删除指定用户的所有角色关联
func (r *SysUsersRolesRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	logger.Info("Deleting all role associations for user", zap.Uint("user_id", userID))

	err := r.getDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.SysUsersRoles{}).Error
	if err != nil {
		logger.Error("Failed to delete user role associations",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted all role associations for user", zap.Uint("user_id", userID))
	return nil
}

// DeleteByRoleID 删除指定角色的所有用户关联
func (r *SysUsersRolesRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}

	logger.Info("Deleting all user associations for role", zap.Uint("role_id", roleID))

	err := r.getDB().WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&model.SysUsersRoles{}).Error
	if err != nil {
		logger.Error("Failed to delete role user associations",
			zap.Uint("role_id", roleID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted all user associations for role", zap.Uint("role_id", roleID))
	return nil
}

// FindByUserID 根据用户ID查找角色关联
func (r *SysUsersRolesRepository) FindByUserID(ctx context.Context, userID uint) ([]*model.SysUsersRoles, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	var userRoles []*model.SysUsersRoles
	err := r.getDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("role_id ASC").
		Find(&userRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user roles by user %d: %v", userID, err)
	}
	return userRoles, nil
}

// FindByRoleID 根据角色ID查找用户关联
func (r *SysUsersRolesRepository) FindByRoleID(ctx context.Context, roleID uint) ([]*model.SysUsersRoles, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	var userRoles []*model.SysUsersRoles
	err := r.getDB().WithContext(ctx).
		Where("role_id = ?", roleID).
		Order("user_id ASC").
		Find(&userRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role users by role %d: %v", roleID, err)
	}
	return userRoles, nil
}

// FindRoleIDsByUserID 根据用户ID获取角色ID列表
func (r *SysUsersRolesRepository) FindRoleIDsByUserID(ctx context.Context, userID uint) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	var roleIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role IDs by user %d: %v", userID, err)
	}
	return roleIDs, nil
}

// FindUserIDsByRoleID 根据角色ID获取用户ID列表
func (r *SysUsersRolesRepository) FindUserIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	var userIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user IDs by role %d: %v", roleID, err)
	}
	return userIDs, nil
}

// Exists 检查用户角色关联是否存在
func (r *SysUsersRolesRepository) Exists(ctx context.Context, userID, roleID uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if userID == 0 || roleID == 0 {
		return false, fmt.Errorf("用户ID和角色ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user role association existence: %v", err)
	}
	return count > 0, nil
}

// BatchCreate 批量创建用户角色关联
func (r *SysUsersRolesRepository) BatchCreate(ctx context.Context, userRoles []*model.SysUsersRoles) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if len(userRoles) == 0 {
		return fmt.Errorf("用户角色关联列表不能为空")
	}

	// 验证所有数据
	for i, userRole := range userRoles {
		if err := userRole.ValidateForCreate(); err != nil {
			return fmt.Errorf("validation failed for item %d: %v", i, err)
		}
	}

	logger.Info("Batch creating user role associations", zap.Int("count", len(userRoles)))

	err := r.getDB().WithContext(ctx).CreateInBatches(userRoles, 100).Error
	if err != nil {
		logger.Error("Failed to batch create user role associations",
			zap.Int("count", len(userRoles)),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully batch created user role associations", zap.Int("count", len(userRoles)))
	return nil
}

// BatchUpdateByUserID 批量更新用户的角色关联
func (r *SysUsersRolesRepository) BatchUpdateByUserID(ctx context.Context, userID uint, roleIDs []uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	logger.Info("Batch updating user role associations",
		zap.Uint("user_id", userID),
		zap.Int("role_count", len(roleIDs)))

	// 使用事务确保数据一致性
	return r.getDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除现有关联
		if err := tx.Where("user_id = ?", userID).Delete(&model.SysUsersRoles{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing user role associations: %v", err)
		}

		// 如果没有新的角色ID，直接返回
		if len(roleIDs) == 0 {
			return nil
		}

		// 创建新的关联
		var userRoles []*model.SysUsersRoles
		for _, roleID := range roleIDs {
			if roleID > 0 { // 确保角色ID有效
				userRoles = append(userRoles, &model.SysUsersRoles{
					UserID: userID,
					RoleID: roleID,
				})
			}
		}

		if len(userRoles) > 0 {
			if err := tx.CreateInBatches(userRoles, 100).Error; err != nil {
				return fmt.Errorf("failed to create new user role associations: %v", err)
			}
		}

		return nil
	})
}

// CountByUserID 统计指定用户的角色数量
func (r *SysUsersRolesRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if userID == 0 {
		return 0, fmt.Errorf("用户ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count roles by user %d: %v", userID, err)
	}
	return count, nil
}

// CountByRoleID 统计指定角色的用户数量
func (r *SysUsersRolesRepository) CountByRoleID(ctx context.Context, roleID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if roleID == 0 {
		return 0, fmt.Errorf("角色ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role %d: %v", roleID, err)
	}
	return count, nil
}

// FindAll 查询所有用户角色关联
func (r *SysUsersRolesRepository) FindAll(ctx context.Context) ([]*model.SysUsersRoles, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userRoles []*model.SysUsersRoles
	err := r.getDB().WithContext(ctx).
		Order("user_id ASC, role_id ASC").
		Find(&userRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all user role associations: %v", err)
	}
	return userRoles, nil
}

// FindUsersWithMultipleRoles 查找拥有多个角色的用户
func (r *SysUsersRolesRepository) FindUsersWithMultipleRoles(ctx context.Context) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var userIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Select("user_id").
		Group("user_id").
		Having("COUNT(role_id) > 1").
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users with multiple roles: %v", err)
	}
	return userIDs, nil
}

// FindRolesWithMultipleUsers 查找拥有多个用户的角色
func (r *SysUsersRolesRepository) FindRolesWithMultipleUsers(ctx context.Context) ([]uint, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var roleIDs []uint
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUsersRoles{}).
		Select("role_id").
		Group("role_id").
		Having("COUNT(user_id) > 1").
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles with multiple users: %v", err)
	}
	return roleIDs, nil
}

// HealthCheck 检查数据库连接健康状态
func (r *SysUsersRolesRepository) HealthCheck(ctx context.Context) error {
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
func (r *SysUsersRolesRepository) InitTable() error {
	// 创建表
	mod := &model.SysUsersRoles{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
	}{
		{"idx_user_id", "user_id"},
		{"idx_role_id", "role_id"},
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
