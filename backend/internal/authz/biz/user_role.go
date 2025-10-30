package biz

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"
)

// UserRoleRepo user role association repository interface
type UserRoleRepo interface {
	// GetUserRoles query user role associations
	GetUserRoles(ctx context.Context, userId, roleId uint) ([]*model.SysUsersRoles, error)
	// AddUserRole add user role association
	AddUserRole(ctx context.Context, userId, roleId uint) (*model.SysUsersRoles, error)
	// DeleteUserRole delete user role association
	DeleteUserRole(ctx context.Context, userId, roleId uint) error
	// BatchAddUserRoles batch add user role associations
	BatchAddUserRoles(ctx context.Context, userRoles []*model.SysUsersRoles) error
	// GetRoleIdsByUserId get role ID list by user ID
	GetRoleIdsByUserId(ctx context.Context, userId uint) ([]uint, error)
	// GetUserIdsByRoleId get user ID list by role ID
	GetUserIdsByRoleId(ctx context.Context, roleId uint) ([]uint, error)
	// DeleteByUserId delete all role associations of user
	DeleteByUserId(ctx context.Context, userId uint) error
	// DeleteByRoleId delete all user associations of role
	DeleteByRoleId(ctx context.Context, roleId uint) error
}

// UserRoleUseCase user role association business logic
type UserRoleUseCase struct {
	userRoleRepo UserRoleRepo
}

// NewUserRoleUseCase create user role association business logic instance
func NewUserRoleUseCase(userRoleRepo UserRoleRepo) *UserRoleUseCase {
	return &UserRoleUseCase{
		userRoleRepo: userRoleRepo,
	}
}

// GetUserRoles query user role associations
func (uc *UserRoleUseCase) GetUserRoles(ctx context.Context, userId, roleId uint) ([]*model.SysUsersRoles, error) {
	logger.Info("query user role associations",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	userRoles, err := uc.userRoleRepo.GetUserRoles(ctx, userId, roleId)
	if err != nil {
		logger.Error("query user role associations failed",
			zap.Error(err),
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId))
		return nil, fmt.Errorf("query user role associations failed: %v", err)
	}

	logger.Info("query user role associations success",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId),
		zap.Int("count", len(userRoles)))

	return userRoles, nil
}

// AddUserRole add user role association
func (uc *UserRoleUseCase) AddUserRole(ctx context.Context, userId, roleId uint) (*model.SysUsersRoles, error) {
	logger.Info("add user role association",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	// Validate parameters
	if userId == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if roleId == 0 {
		return nil, fmt.Errorf("role ID cannot be empty")
	}

	// Check if association already exists
	existingRoles, err := uc.userRoleRepo.GetUserRoles(ctx, userId, roleId)
	if err != nil {
		logger.Error("check user role association failed",
			zap.Error(err),
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId))
		return nil, fmt.Errorf("check user role association failed: %v", err)
	}

	if len(existingRoles) > 0 {
		return nil, fmt.Errorf("user role association already exists")
	}

	userRole, err := uc.userRoleRepo.AddUserRole(ctx, userId, roleId)
	if err != nil {
		logger.Error("add user role association failed",
			zap.Error(err),
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId))
		return nil, fmt.Errorf("add user role association failed: %v", err)
	}

	logger.Info("add user role association success",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	return userRole, nil
}

// DeleteUserRole delete user role association
func (uc *UserRoleUseCase) DeleteUserRole(ctx context.Context, userId, roleId uint) error {
	logger.Info("delete user role association",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	// Validate parameters
	if userId == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if roleId == 0 {
		return fmt.Errorf("role ID cannot be empty")
	}

	err := uc.userRoleRepo.DeleteUserRole(ctx, userId, roleId)
	if err != nil {
		logger.Error("delete user role association failed",
			zap.Error(err),
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId))
		return fmt.Errorf("delete user role association failed: %v", err)
	}

	logger.Info("delete user role association success",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	return nil
}

// BatchAddUserRoles batch add user role associations
func (uc *UserRoleUseCase) BatchAddUserRoles(ctx context.Context, userRoles []*model.SysUsersRoles) error {
	logger.Info("batch add user role associations", zap.Int("count", len(userRoles)))

	// Validate parameters
	if len(userRoles) == 0 {
		return fmt.Errorf("user role association list cannot be empty")
	}

	// Validate each association
	for i, userRole := range userRoles {
		if userRole.UserID == 0 {
			return fmt.Errorf("user ID of the %d-th user role association cannot be empty", i+1)
		}
		if userRole.RoleID == 0 {
			return fmt.Errorf("role ID of the %d-th user role association cannot be empty", i+1)
		}
	}

	err := uc.userRoleRepo.BatchAddUserRoles(ctx, userRoles)
	if err != nil {
		logger.Error("batch add user role associations failed",
			zap.Error(err))
		return fmt.Errorf("batch add user role associations failed: %v", err)
	}

	logger.Info("batch add user role associations success", zap.Int("count", len(userRoles)))

	return nil
}

// GetRoleIdsByUserId get role ID list by user ID
func (uc *UserRoleUseCase) GetRoleIdsByUserId(ctx context.Context, userId uint) ([]uint, error) {
	logger.Info("get user role ID list", zap.Uint("userId", userId))

	// Validate parameters
	if userId == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	roleIds, err := uc.userRoleRepo.GetRoleIdsByUserId(ctx, userId)
	if err != nil {
		logger.Error("get user role ID list failed",
			zap.Error(err),
			zap.Uint("userId", userId))
		return nil, fmt.Errorf("get user role ID list failed: %v", err)
	}

	logger.Info("get user role ID list success",
		zap.Uint("userId", userId),
		zap.Int("count", len(roleIds)))

	return roleIds, nil
}

// GetUserIdsByRoleId get user ID list by role ID
func (uc *UserRoleUseCase) GetUserIdsByRoleId(ctx context.Context, roleId uint) ([]uint, error) {
	logger.Info("get role user ID list", zap.Uint("roleId", roleId))

	// Validate parameters
	if roleId == 0 {
		return nil, fmt.Errorf("role ID cannot be empty")
	}

	userIds, err := uc.userRoleRepo.GetUserIdsByRoleId(ctx, roleId)
	if err != nil {
		logger.Error("get role user ID list failed",
			zap.Error(err),
			zap.Uint("roleId", roleId))
		return nil, fmt.Errorf("get role user ID list failed: %v", err)
	}

	logger.Info("get role user ID list success",
		zap.Uint("roleId", roleId),
		zap.Int("count", len(userIds)))

	return userIds, nil
}

// DeleteByUserId delete all role associations of user
func (uc *UserRoleUseCase) DeleteByUserId(ctx context.Context, userId uint) error {
	logger.Info("delete all role associations of user", zap.Uint("userId", userId))

	// Validate parameters
	if userId == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}

	err := uc.userRoleRepo.DeleteByUserId(ctx, userId)
	if err != nil {
		logger.Error("delete all role associations of user failed",
			zap.Error(err),
			zap.Uint("userId", userId))
		return fmt.Errorf("delete all role associations of user failed: %v", err)
	}

	logger.Info("delete all role associations of user success", zap.Uint("userId", userId))

	return nil
}

// DeleteByRoleId delete all user associations of role
func (uc *UserRoleUseCase) DeleteByRoleId(ctx context.Context, roleId uint) error {
	logger.Info("delete all user associations of role", zap.Uint("roleId", roleId))

	// Validate parameters
	if roleId == 0 {
		return fmt.Errorf("role ID cannot be empty")
	}

	err := uc.userRoleRepo.DeleteByRoleId(ctx, roleId)
	if err != nil {
		logger.Error("delete all user associations of role failed",
			zap.Error(err),
			zap.Uint("roleId", roleId))
		return fmt.Errorf("delete all user associations of role failed: %v", err)
	}

	logger.Info("delete all user associations of role success", zap.Uint("roleId", roleId))

	return nil
}
