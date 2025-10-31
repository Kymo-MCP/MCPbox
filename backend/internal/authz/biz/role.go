package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// RoleData role data access layer
type RoleData struct {
	ctx  context.Context
	repo *mysql.SysRoleRepository
}

// NewRoleData creates role data access layer instance
func NewRoleData(ctx context.Context) *RoleData {
	return &RoleData{
		ctx:  ctx,
		repo: mysql.SysRoleRepo,
	}
}

// CreateRole creates role
func (r *RoleData) CreateRole(ctx context.Context, role *model.SysRole) error {
	// Prepare data before creation
	if err := role.PrepareForCreate(); err != nil {
		return fmt.Errorf("failed to prepare role data for creation: %v", err)
	}

	// Validate data
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("role data validation failed: %v", err)
	}

	return r.repo.Create(ctx, role)
}

// UpdateRole updates role
func (r *RoleData) UpdateRole(ctx context.Context, role *model.SysRole) error {
	// Prepare data before update
	if err := role.PrepareForUpdate(); err != nil {
		return fmt.Errorf("failed to prepare role data for update: %v", err)
	}

	// Validate data
	if err := role.ValidateForUpdate(); err != nil {
		return fmt.Errorf("role data validation failed: %v", err)
	}

	return r.repo.Update(ctx, role)
}

// DeleteRole deletes role
func (r *RoleData) DeleteRole(ctx context.Context, id uint) error {
	return r.repo.Delete(ctx, id)
}

// GetRoleByID gets role by ID
func (r *RoleData) GetRoleByID(ctx context.Context, id uint) (*model.SysRole, error) {
	return r.repo.FindByID(ctx, id)
}

// GetRoleByName gets role by name
func (r *RoleData) GetRoleByName(ctx context.Context, name string) (*model.SysRole, error) {
	return r.repo.FindByName(ctx, name)
}

// GetRoleList gets role list
func (r *RoleData) GetRoleList(ctx context.Context, name string, level *int) ([]*model.SysRole, error) {
	if name != "" {
		// Search by name
		role, err := r.repo.FindByName(ctx, name)
		if err != nil {
			return []*model.SysRole{}, nil // Return empty list when not found
		}
		return []*model.SysRole{role}, nil
	}

	if level != nil {
		return r.repo.FindByLevel(ctx, *level)
	}

	return r.repo.FindAll(ctx)
}

// GetRoleListWithPagination gets role list with pagination
func (r *RoleData) GetRoleListWithPagination(ctx context.Context, page, size int, name string, level *int) ([]*model.SysRole, int64, error) {
	// First get all roles that meet the criteria
	allRoles, err := r.GetRoleList(ctx, name, level)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get role list: %v", err)
	}

	total := int64(len(allRoles))

	// Calculate pagination
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

// GetRolePermissions gets role permission list
func (r *RoleData) GetRolePermissions(ctx context.Context, roleID uint) ([]string, error) {
	// This needs to be implemented based on actual permission table structure
	// Currently returns empty list, in actual project need to query role permission association table
	return []string{}, nil
}

// AssignPermissions assigns permissions to role
func (r *RoleData) AssignPermissions(ctx context.Context, roleID uint, permissions []string) error {
	// This needs to be implemented based on actual permission table structure
	// Currently returns nil, in actual project need to operate role permission association table
	return nil
}

// CheckPermission checks if role has specified permission
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
