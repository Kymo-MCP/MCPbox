package biz

import (
	"context"
	"fmt"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
)

// RoleData 角色数据访问层
type RoleData struct {
	ctx  context.Context
	repo *mysql.SysRoleRepository
}

// NewRoleData 创建角色数据访问层实例
func NewRoleData(ctx context.Context) *RoleData {
	return &RoleData{
		ctx:  ctx,
		repo: mysql.SysRoleRepo,
	}
}

// CreateRole 创建角色
func (r *RoleData) CreateRole(ctx context.Context, role *model.SysRole) error {
	// 准备创建前的数据
	if err := role.PrepareForCreate(); err != nil {
		return fmt.Errorf("准备创建角色数据失败: %v", err)
	}

	// 验证数据
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("角色数据验证失败: %v", err)
	}

	return r.repo.Create(ctx, role)
}

// UpdateRole 更新角色
func (r *RoleData) UpdateRole(ctx context.Context, role *model.SysRole) error {
	// 准备更新前的数据
	if err := role.PrepareForUpdate(); err != nil {
		return fmt.Errorf("准备更新角色数据失败: %v", err)
	}

	// 验证数据
	if err := role.ValidateForUpdate(); err != nil {
		return fmt.Errorf("角色数据验证失败: %v", err)
	}

	return r.repo.Update(ctx, role)
}

// DeleteRole 删除角色
func (r *RoleData) DeleteRole(ctx context.Context, id uint) error {
	return r.repo.Delete(ctx, id)
}

// GetRoleByID 根据ID获取角色
func (r *RoleData) GetRoleByID(ctx context.Context, id uint) (*model.SysRole, error) {
	return r.repo.FindByID(ctx, id)
}

// GetRoleByName 根据名称获取角色
func (r *RoleData) GetRoleByName(ctx context.Context, name string) (*model.SysRole, error) {
	return r.repo.FindByName(ctx, name)
}

// GetRoleList 获取角色列表
func (r *RoleData) GetRoleList(ctx context.Context, name string, level *int) ([]*model.SysRole, error) {
	if name != "" {
		// 使用名称查找
		role, err := r.repo.FindByName(ctx, name)
		if err != nil {
			return []*model.SysRole{}, nil // 没找到返回空列表
		}
		return []*model.SysRole{role}, nil
	}

	if level != nil {
		return r.repo.FindByLevel(ctx, *level)
	}

	return r.repo.FindAll(ctx)
}

// GetRoleListWithPagination 分页获取角色列表
func (r *RoleData) GetRoleListWithPagination(ctx context.Context, page, size int, name string, level *int) ([]*model.SysRole, int64, error) {
	// 先获取所有符合条件的角色
	allRoles, err := r.GetRoleList(ctx, name, level)
	if err != nil {
		return nil, 0, fmt.Errorf("获取角色列表失败: %v", err)
	}

	total := int64(len(allRoles))

	// 计算分页
	offset := (page - 1) * size
	if offset >= len(allRoles) {
		return []*model.SysRole{}, total, nil
	}

	end := offset + size
	if end > len(allRoles) {
		end = len(allRoles)
	}

	return allRoles[offset:end], total, nil
}

// GetRolePermissions 获取角色权限列表
func (r *RoleData) GetRolePermissions(ctx context.Context, roleID uint) ([]string, error) {
	// 这里需要根据实际的权限表结构来实现
	// 暂时返回空列表，实际项目中需要查询角色权限关联表
	return []string{}, nil
}

// AssignPermissions 为角色分配权限
func (r *RoleData) AssignPermissions(ctx context.Context, roleID uint, permissions []string) error {
	// 这里需要根据实际的权限表结构来实现
	// 暂时返回nil，实际项目中需要操作角色权限关联表
	return nil
}

// CheckPermission 检查角色是否有指定权限
func (r *RoleData) CheckPermission(ctx context.Context, roleID uint, permission string) (bool, error) {
	permissions, err := r.GetRolePermissions(ctx, roleID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}
