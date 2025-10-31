package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var SysUserRepo *SysUserRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysUserRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_user table: %v", err))
		}
	})
}

// SysUserRepository 系统用户仓库
type SysUserRepository struct{}

// NewSysUserRepository 创建系统用户仓库实例
func NewSysUserRepository(db *gorm.DB) *SysUserRepository {
	SysUserRepo = &SysUserRepository{}
	return SysUserRepo
}

// getDB 获取数据库连接
func (r *SysUserRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysUser{})
}

// Create 创建用户
func (r *SysUserRepository) Create(ctx context.Context, user *model.SysUser) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 创建前的准备工作
	if err := user.PrepareForCreate(); err != nil {
		return fmt.Errorf("prepare for create failed: %v", err)
	}

	// 验证数据
	if err := user.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Creating user",
		zap.String("username", user.GetUsername()),
		zap.String("email", user.GetEmail()),
		zap.String("source", user.GetSource()))

	err := r.getDB().WithContext(ctx).Create(user).Error
	if err != nil {
		logger.Error("Failed to create user",
			zap.String("username", user.GetUsername()),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully created user",
		zap.Uint("id", user.UserID),
		zap.String("username", user.GetUsername()))

	return nil
}

// Update 更新用户
func (r *SysUserRepository) Update(ctx context.Context, user *model.SysUser) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 更新前的准备工作
	if err := user.PrepareForUpdate(); err != nil {
		return fmt.Errorf("prepare for update failed: %v", err)
	}

	// 验证数据
	if err := user.ValidateForUpdate(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	logger.Info("Updating user",
		zap.Uint("id", user.UserID),
		zap.String("username", user.GetUsername()))

	err := r.getDB().WithContext(ctx).Where("user_id = ?", user.UserID).Save(user).Error
	if err != nil {
		logger.Error("Failed to update user",
			zap.Uint("id", user.UserID),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully updated user",
		zap.Uint("id", user.UserID),
		zap.String("username", user.GetUsername()))

	return nil
}

// Delete 删除用户（物理删除）
func (r *SysUserRepository) Delete(ctx context.Context, id uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	logger.Info("Deleting user", zap.Uint("id", id))

	err := r.getDB().WithContext(ctx).Delete(&model.SysUser{}, id).Error
	if err != nil {
		logger.Error("Failed to delete user",
			zap.Uint("id", id),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully deleted user", zap.Uint("id", id))
	return nil
}

// FindByID 根据ID查找用户
func (r *SysUserRepository) FindByID(ctx context.Context, id uint) (*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *SysUserRepository) FindByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to find user by username '%s': %v", username, err)
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *SysUserRepository) FindByEmail(ctx context.Context, email string) (*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to find user by email '%s': %v", email, err)
	}
	return &user, nil
}

// FindByDeptID 根据部门ID查找用户
func (r *SysUserRepository) FindByDeptID(ctx context.Context, deptID uint) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return nil, fmt.Errorf("部门ID不能为空")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("dept_id = ?", deptID).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by dept %d: %v", deptID, err)
	}
	return users, nil
}

// FindByEnabled 根据启用状态查找用户
func (r *SysUserRepository) FindByEnabled(ctx context.Context, enabled bool) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("enabled = ?", enabled).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by enabled status %v: %v", enabled, err)
	}
	return users, nil
}

// FindBySource 根据用户来源查找用户
func (r *SysUserRepository) FindBySource(ctx context.Context, source string) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if source == "" {
		return nil, fmt.Errorf("用户来源不能为空")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("source = ?", source).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by source %s: %v", source, err)
	}
	return users, nil
}

// FindByCorpID 根据企业ID查找用户
func (r *SysUserRepository) FindByCorpID(ctx context.Context, corpID string) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if corpID == "" {
		return nil, fmt.Errorf("企业ID不能为空")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("corp_id = ?", corpID).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by corp ID %s: %v", corpID, err)
	}
	return users, nil
}

// FindByFeishuID 根据飞书ID查找用户
func (r *SysUserRepository) FindByFeishuID(ctx context.Context, feishuID string) (*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if feishuID == "" {
		return nil, fmt.Errorf("飞书ID不能为空")
	}

	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("feishu_id = ?", feishuID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with feishu ID: %s", feishuID)
		}
		return nil, fmt.Errorf("failed to find user by feishu ID '%s': %v", feishuID, err)
	}
	return &user, nil
}

// FindByThirdPartyOpenID 根据第三方平台OpenID查找用户
func (r *SysUserRepository) FindByThirdPartyOpenID(ctx context.Context, openID string) (*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if openID == "" {
		return nil, fmt.Errorf("第三方平台OpenID不能为空")
	}

	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("third_party_open_id = ?", openID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with third party open ID: %s", openID)
		}
		return nil, fmt.Errorf("failed to find user by third party open ID '%s': %v", openID, err)
	}
	return &user, nil
}

// FindAdmins 查找管理员用户
func (r *SysUserRepository) FindAdmins(ctx context.Context) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("is_admin = ?", true).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find admin users: %v", err)
	}
	return users, nil
}

// FindAll 查找所有用户
func (r *SysUserRepository) FindAll(ctx context.Context) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all users: %v", err)
	}
	return users, nil
}

// UpdateEnabled 更新用户启用状态
func (r *SysUserRepository) UpdateEnabled(ctx context.Context, id uint, enabled bool) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id = ?", id).
		Updates(map[string]interface{}{
			"enabled":     enabled,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update enabled status for user %d: %v", id, err)
	}
	return nil
}

// UpdatePassword 更新用户密码
func (r *SysUserRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id = ?", id).
		Updates(map[string]interface{}{
			"password":       password,
			"pwd_reset_time": &now,
			"update_time":    &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update password for user %d: %v", id, err)
	}
	return nil
}

// UpdateDept 更新用户部门
func (r *SysUserRepository) UpdateDept(ctx context.Context, id uint, deptID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id = ?", id).
		Updates(map[string]interface{}{
			"dept_id":     deptID,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update dept for user %d: %v", id, err)
	}
	return nil
}

// ExistsByUsername 检查指定用户名是否存在
func (r *SysUserRepository) ExistsByUsername(ctx context.Context, username string, excludeID ...uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if username == "" {
		return false, fmt.Errorf("用户名不能为空")
	}

	query := r.getDB().WithContext(ctx).Model(&model.SysUser{}).Where("username = ?", username)

	// 排除指定ID（用于更新时检查重名）
	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("user_id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %v", err)
	}
	return count > 0, nil
}

// ExistsByEmail 检查指定邮箱是否存在
func (r *SysUserRepository) ExistsByEmail(ctx context.Context, email string, excludeID ...uint) (bool, error) {
	if r.getDB() == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if email == "" {
		return false, fmt.Errorf("邮箱不能为空")
	}

	query := r.getDB().WithContext(ctx).Model(&model.SysUser{}).Where("email = ?", email)

	// 排除指定ID（用于更新时检查重名）
	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("user_id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %v", err)
	}
	return count > 0, nil
}

// CountByDeptID 统计指定部门的用户数量
func (r *SysUserRepository) CountByDeptID(ctx context.Context, deptID uint) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if deptID == 0 {
		return 0, fmt.Errorf("部门ID不能为空")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("dept_id = ?", deptID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by dept %d: %v", deptID, err)
	}
	return count, nil
}

// CountByEnabled 统计指定启用状态的用户数量
func (r *SysUserRepository) CountByEnabled(ctx context.Context, enabled bool) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("enabled = ?", enabled).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by enabled status %v: %v", enabled, err)
	}
	return count, nil
}

// CountBySource 统计指定来源的用户数量
func (r *SysUserRepository) CountBySource(ctx context.Context, source string) (int64, error) {
	if r.getDB() == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("source = ?", source).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by source %s: %v", source, err)
	}
	return count, nil
}

// SearchByKeyword 根据关键词搜索用户（用户名、昵称、邮箱）
func (r *SysUserRepository) SearchByKeyword(ctx context.Context, keyword string) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if keyword == "" {
		return r.FindAll(ctx)
	}

	var users []*model.SysUser
	searchPattern := "%" + keyword + "%"
	err := r.getDB().WithContext(ctx).
		Where("username LIKE ? OR nick_name LIKE ? OR email LIKE ?", searchPattern, searchPattern, searchPattern).
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search users by keyword '%s': %v", keyword, err)
	}
	return users, nil
}

// FindUsersWithoutDept 查找没有部门的用户
func (r *SysUserRepository) FindUsersWithoutDept(ctx context.Context) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("dept_id IS NULL OR dept_id = 0").
		Order("user_id ASC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users without dept: %v", err)
	}
	return users, nil
}

// FindRecentUsers 查找最近创建的用户
func (r *SysUserRepository) FindRecentUsers(ctx context.Context, days int) ([]*model.SysUser, error) {
	if r.getDB() == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if days <= 0 {
		days = 7 // 默认7天
	}

	since := time.Now().AddDate(0, 0, -days)
	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("create_time >= ?", since).
		Order("create_time DESC").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find recent users: %v", err)
	}
	return users, nil
}

// BatchUpdateDept 批量更新用户部门
func (r *SysUserRepository) BatchUpdateDept(ctx context.Context, userIDs []uint, deptID uint) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if len(userIDs) == 0 {
		return fmt.Errorf("用户ID列表不能为空")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id IN ?", userIDs).
		Updates(map[string]interface{}{
			"dept_id":     deptID,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to batch update dept for users: %v", err)
	}
	return nil
}

// BatchUpdateEnabled 批量更新用户启用状态
func (r *SysUserRepository) BatchUpdateEnabled(ctx context.Context, userIDs []uint, enabled bool) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	if len(userIDs) == 0 {
		return fmt.Errorf("用户ID列表不能为空")
	}

	now := time.Now()
	err := r.getDB().WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id IN ?", userIDs).
		Updates(map[string]interface{}{
			"enabled":     enabled,
			"update_time": &now,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to batch update enabled status for users: %v", err)
	}
	return nil
}

// HealthCheck 检查数据库连接健康状态
func (r *SysUserRepository) HealthCheck(ctx context.Context) error {
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
func (r *SysUserRepository) InitTable() error {
	// 创建表
	mod := &model.SysUser{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
	}{
		{"idx_dept_id", "dept_id"},
		{"idx_enabled", "enabled"},
		{"idx_nick_name", "nick_name"},
		{"uniq_username", "username"},
		{"uniq_email", "email"},
	}

	for _, idx := range indexes {
		var count int64
		r.getDB().Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
			mod.TableName(), idx.name).Scan(&count)

		if count == 0 {
			// 对于唯一索引，使用不同的创建语句
			if strings.Contains(idx.name, "uniq") || strings.Contains(idx.name, "UK_") {
				if err := r.getDB().Exec(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", idx.name, mod.TableName(), idx.column)).Error; err != nil {
					return fmt.Errorf("failed to create unique index %s: %v", idx.name, err)
				}
			} else {
				if err := r.getDB().Exec(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.name, mod.TableName(), idx.column)).Error; err != nil {
					return fmt.Errorf("failed to create index %s: %v", idx.name, err)
				}
			}
		}
	}

	return nil
}
