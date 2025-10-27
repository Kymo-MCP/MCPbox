package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/redis"
	"qm-mcp-server/pkg/utils"
)

// ListUsersParams 用户列表查询参数
type ListUsersParams struct {
	Keyword  string
	Enabled  *bool
	DeptId   uint
	Page     int
	PageSize int
}

// UserBiz 用户业务逻辑实现
type UserBiz struct {
	userRepo     *mysql.SysUserRepository
	userRoleRepo *mysql.SysUsersRolesRepository
	roleRepo     *mysql.SysRoleRepository
	deptRepo     *mysql.SysDeptRepository
	db           *gorm.DB
	logger       *zap.Logger
}

// NewUserBiz 创建用户业务逻辑实例
func NewUserBiz() *UserBiz {
	return &UserBiz{
		userRepo:     mysql.SysUserRepo,
		userRoleRepo: mysql.SysUsersRolesRepo,
		roleRepo:     mysql.SysRoleRepo,
		deptRepo:     mysql.SysDeptRepo,
		db:           mysql.GetDB(),
		logger:       logger.L().Logger,
	}
}

// CreateUser 创建用户
func (uc *UserBiz) CreateUser(ctx context.Context, user *model.SysUser) error {
	// 检查用户名是否已存在
	var existingUser model.SysUser
	err := uc.db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameAlreadyExists, *user.Username))
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if user.Email != nil && *user.Email != "" {
		err := uc.db.Where("email = ?", *user.Email).First(&existingUser).Error
		if err == nil {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeEmailAlreadyExists, *user.Email))
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
		}
	}

	// 生成随机盐值
	if user.Salt == nil || *user.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeGenerateSaltFailure, err))
		}
		user.Salt = &salt
	}

	// 设置创建时间
	now := time.Now()
	user.CreateTime = &now
	user.UpdateTime = &now

	// 创建用户
	err = uc.db.Create(user).Error
	if err != nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
	}

	var username string
	if user.Username != nil {
		username = *user.Username
	}
	uc.logger.Info("用户创建成功", zap.String("username", username), zap.Uint("userID", user.UserID))
	return nil
}

// UpdateUser 更新用户信息
func (uc *UserBiz) UpdateUser(ctx context.Context, user *model.SysUser) error {
	// 设置更新时间
	now := time.Now()
	user.UpdateTime = &now

	// 更新用户
	err := uc.db.Save(user).Error
	if err != nil {
		return fmt.Errorf("更新用户失败: %v", err)
	}

	uc.logger.Info("用户更新成功", zap.Uint("userId", user.UserID))
	return nil
}

// GetUserById 根据ID获取用户
func (uc *UserBiz) GetUserById(ctx context.Context, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := uc.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		logger.Error("获取用户失败", zap.Error(err), zap.Uint("userId", id))
		return nil, fmt.Errorf("获取用户失败: %v", err)
	}

	return &user, nil
}

// DeleteUser 删除用户
func (uc *UserBiz) DeleteUser(ctx context.Context, id uint) error {
	logger.Info("删除用户", zap.Uint("userId", id))

	// 使用事务删除用户及其关联数据
	err := uc.db.Transaction(func(tx *gorm.DB) error {
		// 先删除用户角色关联
		if err := tx.Where("user_id = ?", id).Delete(&model.SysUsersRoles{}).Error; err != nil {
			return fmt.Errorf("删除用户角色关联失败: %v", err)
		}

		// 删除用户
		if err := tx.Delete(&model.SysUser{}, id).Error; err != nil {
			return fmt.Errorf("删除用户失败: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("删除用户失败", zap.Error(err))
		return err
	}

	logger.Info("用户删除成功", zap.Uint("userId", id))
	return nil
}

// ListUsers 获取用户列表
func (uc *UserBiz) ListUsers(ctx context.Context, params *ListUsersParams) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	// 构建查询条件
	query := uc.db.Model(&model.SysUser{})

	// 关键字搜索
	if params.Keyword != "" {
		keyword := strings.TrimSpace(params.Keyword)
		query = query.Where("username LIKE ? OR nick_name LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 部门筛选
	if params.DeptId > 0 {
		query = query.Where("dept_id = ?", params.DeptId)
	}

	// 状态筛选
	if params.Enabled != nil {
		query = query.Where("enabled = ?", *params.Enabled)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %v", err)
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("user_id DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %v", err)
	}

	return users, total, nil
}

// GetUserListWithPagination 分页获取用户列表
func (uc *UserBiz) GetUserListWithPagination(ctx context.Context, blurry string, deptId uint, status *bool, page, size int) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	// 构建查询条件
	query := uc.db.WithContext(ctx).Model(&model.SysUser{})

	// 模糊搜索
	if blurry != "" {
		blurry = strings.TrimSpace(blurry)
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+blurry+"%", "%"+blurry+"%", "%"+blurry+"%")
	}

	// 部门筛选
	if deptId > 0 {
		query = query.Where("dept_id = ?", deptId)
	}

	// 状态筛选
	if status != nil {
		query = query.Where("enabled = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取用户总数失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取用户总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("user_id DESC").Find(&users).Error; err != nil {
		logger.Error("获取用户列表失败", zap.Error(err))
		return nil, 0, fmt.Errorf("获取用户列表失败: %v", err)
	}

	// 填充部门和角色信息
	for _, user := range users {
		// 获取部门信息
		if user.DeptID != nil && *user.DeptID > 0 {
			dept, err := uc.deptRepo.FindByID(ctx, *user.DeptID)
			if err == nil && dept != nil {
				// 注意：model.SysUser结构体中没有Dept字段，这里需要在返回时单独处理
				// user.Dept = dept
			}
		}

		// 获取角色信息（注意：SysUser模型中没有Roles字段，需要单独返回）
		_, err := uc.GetUserRoles(ctx, user.UserID)
		if err != nil {
			logger.Error("获取用户角色失败", zap.Error(err))
		}
	}

	return users, total, nil
}

// GetCurrentUser 获取当前用户信息
func (uc *UserBiz) GetCurrentUser(ctx context.Context, userId uint) (*model.SysUser, error) {
	return uc.GetUserById(ctx, userId)
}

// GetUserPermissions 获取用户权限
func (uc *UserBiz) GetUserPermissions(ctx context.Context, userId uint) ([]string, error) {
	// 1. 获取用户的角色ID列表
	roleIds, err := uc.userRoleRepo.FindRoleIDsByUserID(ctx, userId)
	if err != nil {
		uc.logger.Error("获取用户角色失败", zap.Error(err), zap.Uint("userId", userId))
		return nil, fmt.Errorf("获取用户角色失败: %v", err)
	}

	if len(roleIds) == 0 {
		uc.logger.Info("用户没有分配角色", zap.Uint("userId", userId))
		return []string{}, nil
	}

	// 2. 获取角色信息
	roles := make([]*model.SysRole, 0, len(roleIds))
	for _, roleId := range roleIds {
		role, err := uc.roleRepo.FindByID(ctx, roleId)
		if err != nil {
			uc.logger.Error("获取角色信息失败", zap.Error(err), zap.Uint("roleId", roleId))
			continue
		}
		if role != nil {
			roles = append(roles, role)
		}
	}

	// 3. 基于角色生成基础权限列表
	permissions := make([]string, 0)
	for _, role := range roles {
		// 根据角色名称生成基础权限
		if role.Name != "" {
			// 添加角色基础权限
			permissions = append(permissions, fmt.Sprintf("role:%s", strings.ToLower(role.Name)))

			// 根据角色名称添加常见权限
			switch strings.ToLower(role.Name) {
			case "admin", "管理员":
				permissions = append(permissions,
					"user:create", "user:update", "user:delete", "user:view",
					"role:create", "role:update", "role:delete", "role:view",
					"dept:create", "dept:update", "dept:delete", "dept:view",
					"system:config", "system:monitor",
				)
			case "user", "用户":
				permissions = append(permissions,
					"user:view", "profile:update",
				)
			case "operator", "操作员":
				permissions = append(permissions,
					"user:view", "user:update",
					"dept:view",
				)
			}
		}
	}

	// 4. 去重
	permissionSet := make(map[string]bool)
	for _, perm := range permissions {
		permissionSet[perm] = true
	}

	uniquePermissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		uniquePermissions = append(uniquePermissions, perm)
	}

	uc.logger.Info("获取用户权限成功",
		zap.Uint("userId", userId),
		zap.Int("roleCount", len(roles)),
		zap.Int("permissionCount", len(uniquePermissions)),
	)

	return uniquePermissions, nil
}

// UpdatePassword 更新用户密码
func (uc *UserBiz) UpdatePassword(ctx context.Context, userId uint, oldPassword, newPassword string) error {
	// 获取用户信息
	user, err := uc.GetUserById(ctx, userId)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 验证旧密码
	if user.Password != nil && user.Salt != nil {
		// 使用盐值验证旧密码
		if err := uc.verifyPasswordWithSalt(oldPassword, *user.Salt, *user.Password); err != nil {
			return fmt.Errorf("旧密码不正确")
		}
	} else if user.Password != nil {
		// 兼容旧的无盐值密码验证
		if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(oldPassword)); err != nil {
			return fmt.Errorf("旧密码不正确")
		}
	}

	// 确保用户有盐值
	if user.Salt == nil || *user.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("生成盐值失败: %v", err)
		}
		user.Salt = &salt
	}

	// 使用盐值哈希新密码
	hashedPassword, err := uc.hashPasswordWithSalt(newPassword, *user.Salt)
	if err != nil {
		return fmt.Errorf("新密码哈希失败: %v", err)
	}

	// 更新密码
	user.Password = &hashedPassword
	now := time.Now()
	user.PwdResetTime = &now

	// 保存用户
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	// 删除用户的所有会话，强制重新登录
	if err := redis.DeleteUserSessionsByUserID(userId); err != nil {
		uc.logger.Warn("删除用户会话失败", zap.Uint("userId", userId), zap.Error(err))
	}

	return nil
}

// AssignRolesToUserOld 为用户分配角色（旧版本）
func (uc *UserBiz) AssignRolesToUserOld(ctx context.Context, userId uint, roleIds []uint) error {
	logger.Info("为用户分配角色", zap.Uint("userId", userId), zap.Uints("roleIds", roleIds))

	// 先删除现有的角色关联
	if err := uc.userRoleRepo.DeleteByUserID(ctx, userId); err != nil {
		logger.Error("删除用户现有角色失败", zap.Error(err))
		return fmt.Errorf("删除用户现有角色失败: %v", err)
	}

	// 添加新的角色关联
	for _, roleId := range roleIds {
		userRole := &model.SysUsersRoles{
			UserID: userId,
			RoleID: roleId,
		}
		if err := uc.userRoleRepo.Create(ctx, userRole); err != nil {
			logger.Error("创建用户角色关联失败", zap.Error(err))
			return fmt.Errorf("创建用户角色关联失败: %v", err)
		}
	}

	logger.Info("用户角色分配成功", zap.Uint("userId", userId))
	return nil
}

// GetUserRoles 获取用户角色列表
func (uc *UserBiz) GetUserRoles(ctx context.Context, userId uint) ([]*model.SysRole, error) {
	var userRoles []*model.SysUsersRoles
	if err := uc.db.WithContext(ctx).Where("user_id = ?", userId).Find(&userRoles).Error; err != nil {
		logger.Error("获取用户角色关联失败", zap.Error(err))
		return nil, fmt.Errorf("获取用户角色关联失败: %v", err)
	}

	var roles []*model.SysRole
	for _, userRole := range userRoles {
		role, err := uc.roleRepo.FindByID(ctx, userRole.RoleID)
		if err != nil {
			logger.Error("获取角色信息失败", zap.Error(err), zap.Uint("roleId", userRole.RoleID))
			continue
		}
		if role != nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

// GetUserDept 获取用户部门信息
func (uc *UserBiz) GetUserDept(ctx context.Context, deptId uint) (*model.SysDept, error) {
	dept, err := uc.deptRepo.FindByID(ctx, deptId)
	if err != nil {
		logger.Error("获取部门信息失败", zap.Error(err), zap.Uint("deptId", deptId))
		return nil, fmt.Errorf("获取部门信息失败: %v", err)
	}
	return dept, nil
}

// CheckUsernameExists 检查用户名是否存在
func (uc *UserBiz) CheckUsernameExists(ctx context.Context, username string, excludeId uint) (bool, error) {
	return uc.userRepo.ExistsByUsername(ctx, username, excludeId)
}

// CheckEmailExists 检查邮箱是否存在
func (uc *UserBiz) CheckEmailExists(ctx context.Context, email string, excludeId uint) (bool, error) {
	return uc.userRepo.ExistsByEmail(ctx, email, excludeId)
}

// AssignRolesToUser 为用户分配角色
func (uc *UserBiz) AssignRolesToUser(ctx context.Context, userId uint, roleIds []uint) error {
	// 先删除用户现有的角色关联
	if err := uc.userRoleRepo.DeleteByUserID(ctx, userId); err != nil {
		uc.logger.Error("删除用户角色关联失败", zap.Error(err), zap.Uint("userId", userId))
		return fmt.Errorf("删除用户角色关联失败: %v", err)
	}

	// 添加新的角色关联
	for _, roleId := range roleIds {
		userRole := &model.SysUsersRoles{
			UserID: userId,
			RoleID: roleId,
		}
		if err := uc.userRoleRepo.Create(ctx, userRole); err != nil {
			uc.logger.Error("创建用户角色关联失败", zap.Error(err), zap.Uint("userId", userId), zap.Uint("roleId", roleId))
			return fmt.Errorf("创建用户角色关联失败: %v", err)
		}
	}

	return nil
}

// hashPasswordWithSalt 使用盐值哈希密码
func (uc *UserBiz) hashPasswordWithSalt(password, salt string) (string, error) {
	// 将密码和盐值组合
	saltedPassword := password + salt

	// 使用bcrypt哈希加盐后的密码
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %v", err)
	}
	return string(hashedBytes), nil
}

// verifyPasswordWithSalt 验证带盐值的密码
func (uc *UserBiz) verifyPasswordWithSalt(password, salt, hashedPassword string) error {
	// 将密码和盐值组合
	saltedPassword := password + salt

	// 使用bcrypt验证
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(saltedPassword))
}

// VerifyPassword 验证密码
func (uc *UserBiz) VerifyPassword(password, salt, hashedPassword string) error {
	return uc.verifyPasswordWithSalt(password, salt, hashedPassword)
}

// EncryptPasswordWithNewAlgorithm 使用新的密码加密算法
func (uc *UserBiz) EncryptPasswordWithNewAlgorithm(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt加密失败: %v", err)
	}

	encryptedPassword := string(hashedBytes)
	logger.Info("新密码加密完成", zap.String("algorithm", "bcrypt"), zap.Int("cost", bcrypt.DefaultCost))
	return encryptedPassword, nil
}

// SetUserPassword 设置用户密码
func (uc *UserBiz) SetUserPassword(ctx context.Context, userModel *model.SysUser, plainPassword string) error {
	// 确保用户有盐值
	if userModel.Salt == nil || *userModel.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("生成盐值失败: %v", err)
		}
		userModel.Salt = &salt
	}

	// 使用盐值哈希密码
	hashedPassword, err := uc.hashPasswordWithSalt(plainPassword, *userModel.Salt)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %v", err)
	}

	// 设置密码
	userModel.Password = &hashedPassword
	now := time.Now()
	userModel.PwdResetTime = &now

	// 保存用户
	if err := uc.userRepo.Update(ctx, userModel); err != nil {
		logger.Error("用户密码设置失败", zap.Error(err), zap.Uint("userId", userModel.UserID))
		return fmt.Errorf("用户密码设置失败: %v", err)
	}

	logger.Info("用户密码设置完成", zap.Uint("userId", userModel.UserID))
	return nil
}
