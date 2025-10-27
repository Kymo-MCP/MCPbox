package biz

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"
)

// UserRoleRepo 用户角色关联仓库接口
type UserRoleRepo interface {
	// GetUserRoles 查询用户角色关联
	GetUserRoles(ctx context.Context, userId, roleId uint) ([]*model.SysUsersRoles, error)
	// AddUserRole 添加用户角色关联
	AddUserRole(ctx context.Context, userId, roleId uint) (*model.SysUsersRoles, error)
	// DeleteUserRole 删除用户角色关联
	DeleteUserRole(ctx context.Context, userId, roleId uint) error
	// BatchAddUserRoles 批量添加用户角色关联
	BatchAddUserRoles(ctx context.Context, userRoles []*model.SysUsersRoles) error
	// GetRoleIdsByUserId 根据用户ID获取角色ID列表
	GetRoleIdsByUserId(ctx context.Context, userId uint) ([]uint, error)
	// GetUserIdsByRoleId 根据角色ID获取用户ID列表
	GetUserIdsByRoleId(ctx context.Context, roleId uint) ([]uint, error)
	// DeleteByUserId 删除用户的所有角色关联
	DeleteByUserId(ctx context.Context, userId uint) error
	// DeleteByRoleId 删除角色的所有用户关联
	DeleteByRoleId(ctx context.Context, roleId uint) error
}

// UserRoleUseCase 用户角色关联业务逻辑
type UserRoleUseCase struct {
	userRoleRepo UserRoleRepo
}

// NewUserRoleUseCase 创建用户角色关联业务逻辑实例
func NewUserRoleUseCase(userRoleRepo UserRoleRepo) *UserRoleUseCase {
	return &UserRoleUseCase{
		userRoleRepo: userRoleRepo,
	}
}

// GetUserRoles 查询用户角色关联
func (uc *UserRoleUseCase) GetUserRoles(ctx context.Context, userId, roleId uint) ([]*model.SysUsersRoles, error) {
	logger.Info("查询用户角色关联",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	userRoles, err := uc.userRoleRepo.GetUserRoles(ctx, userId, roleId)
	if err != nil {
		logger.Error("查询用户角色关联失败",
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return nil, fmt.Errorf("查询用户角色关联失败: %v", err)
	}

	logger.Info("查询用户角色关联成功",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId),
		zap.Int("count", len(userRoles)))

	return userRoles, nil
}

// AddUserRole 添加用户角色关联
func (uc *UserRoleUseCase) AddUserRole(ctx context.Context, userId, roleId uint) (*model.SysUsersRoles, error) {
	logger.Info("添加用户角色关联",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	// 验证参数
	if userId == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}
	if roleId == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	// 检查是否已存在关联
	existingRoles, err := uc.userRoleRepo.GetUserRoles(ctx, userId, roleId)
	if err != nil {
		logger.Error("检查用户角色关联失败",
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return nil, fmt.Errorf("检查用户角色关联失败: %v", err)
	}

	if len(existingRoles) > 0 {
		return nil, fmt.Errorf("用户角色关联已存在")
	}

	userRole, err := uc.userRoleRepo.AddUserRole(ctx, userId, roleId)
	if err != nil {
		logger.Error("添加用户角色关联失败",
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return nil, fmt.Errorf("添加用户角色关联失败: %v", err)
	}

	logger.Info("添加用户角色关联成功",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	return userRole, nil
}

// DeleteUserRole 删除用户角色关联
func (uc *UserRoleUseCase) DeleteUserRole(ctx context.Context, userId, roleId uint) error {
	logger.Info("删除用户角色关联",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	// 验证参数
	if userId == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if roleId == 0 {
		return fmt.Errorf("角色ID不能为空")
	}

	err := uc.userRoleRepo.DeleteUserRole(ctx, userId, roleId)
	if err != nil {
		logger.Error("删除用户角色关联失败",
			zap.Uint("userId", userId),
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return fmt.Errorf("删除用户角色关联失败: %v", err)
	}

	logger.Info("删除用户角色关联成功",
		zap.Uint("userId", userId),
		zap.Uint("roleId", roleId))

	return nil
}

// BatchAddUserRoles 批量添加用户角色关联
func (uc *UserRoleUseCase) BatchAddUserRoles(ctx context.Context, userRoles []*model.SysUsersRoles) error {
	logger.Info("批量添加用户角色关联", zap.Int("count", len(userRoles)))

	// 验证参数
	if len(userRoles) == 0 {
		return fmt.Errorf("用户角色关联列表不能为空")
	}

	// 验证每个关联
	for i, userRole := range userRoles {
		if userRole.UserID == 0 {
			return fmt.Errorf("第%d个用户角色关联的用户ID不能为空", i+1)
		}
		if userRole.RoleID == 0 {
			return fmt.Errorf("第%d个用户角色关联的角色ID不能为空", i+1)
		}
	}

	err := uc.userRoleRepo.BatchAddUserRoles(ctx, userRoles)
	if err != nil {
		logger.Error("批量添加用户角色关联失败",
			zap.Int("count", len(userRoles)),
			zap.Error(err))
		return fmt.Errorf("批量添加用户角色关联失败: %v", err)
	}

	logger.Info("批量添加用户角色关联成功", zap.Int("count", len(userRoles)))

	return nil
}

// GetRoleIdsByUserId 根据用户ID获取角色ID列表
func (uc *UserRoleUseCase) GetRoleIdsByUserId(ctx context.Context, userId uint) ([]uint, error) {
	logger.Info("获取用户角色ID列表", zap.Uint("userId", userId))

	// 验证参数
	if userId == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	roleIds, err := uc.userRoleRepo.GetRoleIdsByUserId(ctx, userId)
	if err != nil {
		logger.Error("获取用户角色ID列表失败",
			zap.Uint("userId", userId),
			zap.Error(err))
		return nil, fmt.Errorf("获取用户角色ID列表失败: %v", err)
	}

	logger.Info("获取用户角色ID列表成功",
		zap.Uint("userId", userId),
		zap.Int("count", len(roleIds)))

	return roleIds, nil
}

// GetUserIdsByRoleId 根据角色ID获取用户ID列表
func (uc *UserRoleUseCase) GetUserIdsByRoleId(ctx context.Context, roleId uint) ([]uint, error) {
	logger.Info("获取角色用户ID列表", zap.Uint("roleId", roleId))

	// 验证参数
	if roleId == 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}

	userIds, err := uc.userRoleRepo.GetUserIdsByRoleId(ctx, roleId)
	if err != nil {
		logger.Error("获取角色用户ID列表失败",
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return nil, fmt.Errorf("获取角色用户ID列表失败: %v", err)
	}

	logger.Info("获取角色用户ID列表成功",
		zap.Uint("roleId", roleId),
		zap.Int("count", len(userIds)))

	return userIds, nil
}

// DeleteByUserId 删除用户的所有角色关联
func (uc *UserRoleUseCase) DeleteByUserId(ctx context.Context, userId uint) error {
	logger.Info("删除用户所有角色关联", zap.Uint("userId", userId))

	// 验证参数
	if userId == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	err := uc.userRoleRepo.DeleteByUserId(ctx, userId)
	if err != nil {
		logger.Error("删除用户所有角色关联失败",
			zap.Uint("userId", userId),
			zap.Error(err))
		return fmt.Errorf("删除用户所有角色关联失败: %v", err)
	}

	logger.Info("删除用户所有角色关联成功", zap.Uint("userId", userId))

	return nil
}

// DeleteByRoleId 删除角色的所有用户关联
func (uc *UserRoleUseCase) DeleteByRoleId(ctx context.Context, roleId uint) error {
	logger.Info("删除角色所有用户关联", zap.Uint("roleId", roleId))

	// 验证参数
	if roleId == 0 {
		return fmt.Errorf("角色ID不能为空")
	}

	err := uc.userRoleRepo.DeleteByRoleId(ctx, roleId)
	if err != nil {
		logger.Error("删除角色所有用户关联失败",
			zap.Uint("roleId", roleId),
			zap.Error(err))
		return fmt.Errorf("删除角色所有用户关联失败: %v", err)
	}

	logger.Info("删除角色所有用户关联成功", zap.Uint("roleId", roleId))

	return nil
}
