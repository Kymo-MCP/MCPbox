package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

// ListUsersParams user list query parameters
type ListUsersParams struct {
	Keyword  string
	Enabled  *bool
	DeptId   uint
	Page     int
	PageSize int
}

// UserBiz user business logic implementation
type UserBiz struct {
	userRepo     *mysql.SysUserRepository
	userRoleRepo *mysql.SysUsersRolesRepository
	roleRepo     *mysql.SysRoleRepository
	deptRepo     *mysql.SysDeptRepository
	db           *gorm.DB
	logger       *zap.Logger
}

// NewUserBiz creates user business logic instance
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

// CreateUser creates user
func (uc *UserBiz) CreateUser(ctx context.Context, user *model.SysUser) error {
	// Check if username already exists
	var existingUser model.SysUser
	err := uc.db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameAlreadyExists, *user.Username))
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
	}

	// Check if email already exists (if email is provided)
	if user.Email != nil && *user.Email != "" {
		err := uc.db.Where("email = ?", *user.Email).First(&existingUser).Error
		if err == nil {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeEmailAlreadyExists, *user.Email))
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
		}
	}

	// Generate random salt
	if user.Salt == nil || *user.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeGenerateSaltFailure, err))
		}
		user.Salt = &salt
	}

	// Set creation time
	now := time.Now()
	user.CreateTime = &now
	user.UpdateTime = &now

	// Create user
	err = uc.db.Create(user).Error
	if err != nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeCreateUserFailure, err))
	}

	var username string
	if user.Username != nil {
		username = *user.Username
	}
	uc.logger.Info("User created successfully", zap.String("username", username), zap.Uint("userID", user.UserID))
	return nil
}

// UpdateUser updates user information
func (uc *UserBiz) UpdateUser(ctx context.Context, user *model.SysUser) error {
	// Set update time
	now := time.Now()
	user.UpdateTime = &now

	// Update user
	err := uc.db.Save(user).Error
	if err != nil {
		return fmt.Errorf("Failed to update user: %v", err)
	}

	uc.logger.Info("User updated successfully", zap.Uint("userId", user.UserID))
	return nil
}

// GetUserById gets user by ID
func (uc *UserBiz) GetUserById(ctx context.Context, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := uc.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("User not found")
		}
		logger.Error("Failed to get user", zap.Error(err), zap.Uint("userId", id))
		return nil, fmt.Errorf("Failed to get user: %v", err)
	}

	return &user, nil
}

// DeleteUser deletes user
func (uc *UserBiz) DeleteUser(ctx context.Context, id uint) error {
	logger.Info("Deleting user", zap.Uint("userId", id))

	// Use transaction to delete user and associated data
	err := uc.db.Transaction(func(tx *gorm.DB) error {
		// First delete user role associations
		if err := tx.Where("user_id = ?", id).Delete(&model.SysUsersRoles{}).Error; err != nil {
			return fmt.Errorf("Failed to delete user role associations: %v", err)
		}

		// Delete user
		if err := tx.Delete(&model.SysUser{}, id).Error; err != nil {
			return fmt.Errorf("Failed to delete user: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return err
	}

	logger.Info("User deleted successfully", zap.Uint("userId", id))
	return nil
}

// ListUsers gets user list
func (uc *UserBiz) ListUsers(ctx context.Context, params *ListUsersParams) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	// Build query conditions
	query := uc.db.Model(&model.SysUser{})

	// Keyword search
	if params.Keyword != "" {
		keyword := strings.TrimSpace(params.Keyword)
		query = query.Where("username LIKE ? OR nick_name LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Department filter
	if params.DeptId > 0 {
		query = query.Where("dept_id = ?", params.DeptId)
	}

	// Status filter
	if params.Enabled != nil {
		query = query.Where("enabled = ?", *params.Enabled)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("Failed to get user count: %v", err)
	}

	// Paginated query
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("user_id DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("Failed to get user list: %v", err)
	}

	return users, total, nil
}

// GetUserListWithPagination gets user list with pagination
func (uc *UserBiz) GetUserListWithPagination(ctx context.Context, blurry string, deptId uint, status *bool, page, size int) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	// Build query conditions
	query := uc.db.WithContext(ctx).Model(&model.SysUser{})

	// Fuzzy search
	if blurry != "" {
		blurry = strings.TrimSpace(blurry)
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+blurry+"%", "%"+blurry+"%", "%"+blurry+"%")
	}

	// Department filter
	if deptId > 0 {
		query = query.Where("dept_id = ?", deptId)
	}

	// Status filter
	if status != nil {
		query = query.Where("enabled = ?", *status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to get user count", zap.Error(err))
		return nil, 0, fmt.Errorf("Failed to get user count: %v", err)
	}

	// Paginated query
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("user_id DESC").Find(&users).Error; err != nil {
		logger.Error("Failed to get user list", zap.Error(err))
		return nil, 0, fmt.Errorf("Failed to get user list: %v", err)
	}

	// Fill department and role information
	for _, user := range users {
		// Get department information
		if user.DeptID != nil && *user.DeptID > 0 {
			dept, err := uc.deptRepo.FindByID(ctx, *user.DeptID)
			if err == nil && dept != nil {
				// Note: SysUser struct doesn't have Dept field, need to handle separately when returning
				// user.Dept = dept
			}
		}

		// Get role information (Note: SysUser model doesn't have Roles field, need to return separately)
		_, err := uc.GetUserRoles(ctx, user.UserID)
		if err != nil {
			logger.Error("Failed to get user roles", zap.Error(err))
		}
	}

	return users, total, nil
}

// GetCurrentUser gets current user information
func (uc *UserBiz) GetCurrentUser(ctx context.Context, userId uint) (*model.SysUser, error) {
	return uc.GetUserById(ctx, userId)
}

// GetUserPermissions gets user permissions
func (uc *UserBiz) GetUserPermissions(ctx context.Context, userId uint) ([]string, error) {
	// 1. Get user's role ID list
	roleIds, err := uc.userRoleRepo.FindRoleIDsByUserID(ctx, userId)
	if err != nil {
		uc.logger.Error("Failed to get user roles", zap.Error(err), zap.Uint("userId", userId))
		return nil, fmt.Errorf("Failed to get user roles: %v", err)
	}

	if len(roleIds) == 0 {
		uc.logger.Info("User has no assigned roles", zap.Uint("userId", userId))
		return []string{}, nil
	}

	// 2. Get role information
	roles := make([]*model.SysRole, 0, len(roleIds))
	for _, roleId := range roleIds {
		role, err := uc.roleRepo.FindByID(ctx, roleId)
		if err != nil {
			uc.logger.Error("Failed to get role information", zap.Error(err), zap.Uint("roleId", roleId))
			continue
		}
		if role != nil {
			roles = append(roles, role)
		}
	}

	// 3. Generate basic permission list based on roles
	permissions := make([]string, 0)
	for _, role := range roles {
		// Generate basic permissions based on role name
		if role.Name != "" {
			// Add role basic permissions
			permissions = append(permissions, fmt.Sprintf("role:%s", strings.ToLower(role.Name)))

			// Add common permissions based on role name
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

	// 4. Remove duplicates
	permissionSet := make(map[string]bool)
	for _, perm := range permissions {
		permissionSet[perm] = true
	}

	uniquePermissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		uniquePermissions = append(uniquePermissions, perm)
	}

	uc.logger.Info("User permissions retrieved successfully",
		zap.Uint("userId", userId),
		zap.Int("roleCount", len(roles)),
		zap.Int("permissionCount", len(uniquePermissions)),
	)

	return uniquePermissions, nil
}

// UpdatePassword updates user password
func (uc *UserBiz) UpdatePassword(ctx context.Context, userId uint, oldPassword, newPassword string) error {
	// Get user information
	user, err := uc.GetUserById(ctx, userId)
	if err != nil {
		return fmt.Errorf("Failed to get user information: %v", err)
	}

	// Verify old password
	if user.Password != nil && user.Salt != nil {
		// Verify old password with salt
		if err := uc.verifyPasswordWithSalt(oldPassword, *user.Salt, *user.Password); err != nil {
			return fmt.Errorf("Old password is incorrect")
		}
	} else if user.Password != nil {
		// Compatible with old password verification without salt
		if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(oldPassword)); err != nil {
			return fmt.Errorf("Old password is incorrect")
		}
	}

	// Ensure user has salt
	if user.Salt == nil || *user.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("Failed to generate salt: %v", err)
		}
		user.Salt = &salt
	}

	// Hash new password with salt
	hashedPassword, err := uc.hashPasswordWithSalt(newPassword, *user.Salt)
	if err != nil {
		return fmt.Errorf("Failed to hash new password: %v", err)
	}

	// Update password
	user.Password = &hashedPassword
	now := time.Now()
	user.PwdResetTime = &now

	// Save user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("Failed to update password: %v", err)
	}

	// Delete all user sessions, force re-login
	if err := redis.DeleteUserSessionsByUserID(userId); err != nil {
		uc.logger.Warn("Failed to delete user sessions", zap.Uint("userId", userId), zap.Error(err))
	}

	return nil
}

// AssignRolesToUserOld assigns roles to user (old version)
func (uc *UserBiz) AssignRolesToUserOld(ctx context.Context, userId uint, roleIds []uint) error {
	logger.Info("Assigning roles to user", zap.Uint("userId", userId), zap.Uints("roleIds", roleIds))

	// First delete existing role associations
	if err := uc.userRoleRepo.DeleteByUserID(ctx, userId); err != nil {
		logger.Error("Failed to delete existing user roles", zap.Error(err))
		return fmt.Errorf("Failed to delete existing user roles: %v", err)
	}

	// Add new role associations
	for _, roleId := range roleIds {
		userRole := &model.SysUsersRoles{
			UserID: userId,
			RoleID: roleId,
		}
		if err := uc.userRoleRepo.Create(ctx, userRole); err != nil {
			logger.Error("Failed to create user role association", zap.Error(err))
			return fmt.Errorf("Failed to create user role association: %v", err)
		}
	}

	logger.Info("User role assignment successful", zap.Uint("userId", userId))
	return nil
}

// GetUserRoles gets user role list
func (uc *UserBiz) GetUserRoles(ctx context.Context, userId uint) ([]*model.SysRole, error) {
	var userRoles []*model.SysUsersRoles
	if err := uc.db.WithContext(ctx).Where("user_id = ?", userId).Find(&userRoles).Error; err != nil {
		logger.Error("Failed to get user role associations", zap.Error(err))
		return nil, fmt.Errorf("Failed to get user role associations: %v", err)
	}

	var roles []*model.SysRole
	for _, userRole := range userRoles {
		role, err := uc.roleRepo.FindByID(ctx, userRole.RoleID)
		if err != nil {
			logger.Error("Failed to get role information", zap.Error(err), zap.Uint("roleId", userRole.RoleID))
			continue
		}
		if role != nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

// GetUserDept gets user department information
func (uc *UserBiz) GetUserDept(ctx context.Context, deptId uint) (*model.SysDept, error) {
	dept, err := uc.deptRepo.FindByID(ctx, deptId)
	if err != nil {
		logger.Error("Failed to get department information", zap.Error(err), zap.Uint("deptId", deptId))
		return nil, fmt.Errorf("Failed to get department information: %v", err)
	}
	return dept, nil
}

// CheckUsernameExists checks if username exists
func (uc *UserBiz) CheckUsernameExists(ctx context.Context, username string, excludeId uint) (bool, error) {
	return uc.userRepo.ExistsByUsername(ctx, username, excludeId)
}

// CheckEmailExists checks if email exists
func (uc *UserBiz) CheckEmailExists(ctx context.Context, email string, excludeId uint) (bool, error) {
	return uc.userRepo.ExistsByEmail(ctx, email, excludeId)
}

// AssignRolesToUser assigns roles to user
func (uc *UserBiz) AssignRolesToUser(ctx context.Context, userId uint, roleIds []uint) error {
	// First delete existing user role associations
	if err := uc.userRoleRepo.DeleteByUserID(ctx, userId); err != nil {
		uc.logger.Error("Failed to delete user role associations", zap.Error(err), zap.Uint("userId", userId))
		return fmt.Errorf("Failed to delete user role associations: %v", err)
	}

	// Add new role associations
	for _, roleId := range roleIds {
		userRole := &model.SysUsersRoles{
			UserID: userId,
			RoleID: roleId,
		}
		if err := uc.userRoleRepo.Create(ctx, userRole); err != nil {
			uc.logger.Error("Failed to create user role association", zap.Error(err), zap.Uint("userId", userId), zap.Uint("roleId", roleId))
			return fmt.Errorf("Failed to create user role association: %v", err)
		}
	}

	return nil
}

// hashPasswordWithSalt hashes password with salt
func (uc *UserBiz) hashPasswordWithSalt(password, salt string) (string, error) {
	// Combine password and salt
	saltedPassword := password + salt

	// Hash salted password with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Password hashing failed: %v", err)
	}
	return string(hashedBytes), nil
}

// verifyPasswordWithSalt verifies password with salt
func (uc *UserBiz) verifyPasswordWithSalt(password, salt, hashedPassword string) error {
	// Combine password and salt
	saltedPassword := password + salt

	// Verify with bcrypt
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(saltedPassword))
}

// VerifyPassword verifies password
func (uc *UserBiz) VerifyPassword(password, salt, hashedPassword string) error {
	return uc.verifyPasswordWithSalt(password, salt, hashedPassword)
}

// EncryptPasswordWithNewAlgorithm encrypts password with new algorithm
func (uc *UserBiz) EncryptPasswordWithNewAlgorithm(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt encryption failed: %v", err)
	}

	encryptedPassword := string(hashedBytes)
	logger.Info("New password encryption completed", zap.String("algorithm", "bcrypt"), zap.Int("cost", bcrypt.DefaultCost))
	return encryptedPassword, nil
}

// SetUserPassword sets user password
func (uc *UserBiz) SetUserPassword(ctx context.Context, userModel *model.SysUser, plainPassword string) error {
	// Ensure user has salt
	if userModel.Salt == nil || *userModel.Salt == "" {
		salt, err := utils.GenerateRandomSalt(32)
		if err != nil {
			return fmt.Errorf("Failed to generate salt: %v", err)
		}
		userModel.Salt = &salt
	}

	// Hash password with salt
	hashedPassword, err := uc.hashPasswordWithSalt(plainPassword, *userModel.Salt)
	if err != nil {
		return fmt.Errorf("Password hashing failed: %v", err)
	}

	// Set password
	userModel.Password = &hashedPassword
	now := time.Now()
	userModel.PwdResetTime = &now

	// Save user
	if err := uc.userRepo.Update(ctx, userModel); err != nil {
		logger.Error("Failed to set user password", zap.Error(err), zap.Uint("userId", userModel.UserID))
		return fmt.Errorf("Failed to set user password: %v", err)
	}

	logger.Info("User password set successfully", zap.Uint("userId", userModel.UserID))
	return nil
}
