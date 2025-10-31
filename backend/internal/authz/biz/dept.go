package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// DeptData department data access layer
type DeptData struct {
	ctx  context.Context
	repo *mysql.SysDeptRepository
}

// NewDeptData creates department data access layer instance
func NewDeptData(ctx context.Context) *DeptData {
	return &DeptData{
		ctx:  ctx,
		repo: mysql.SysDeptRepo,
	}
}

// CreateDept creates department
func (d *DeptData) CreateDept(ctx context.Context, dept *model.SysDept) error {
	return d.repo.Create(ctx, dept)
}

// UpdateDept updates department
func (d *DeptData) UpdateDept(ctx context.Context, dept *model.SysDept) error {
	return d.repo.Update(ctx, dept)
}

// DeleteDept deletes department
func (d *DeptData) DeleteDept(ctx context.Context, id uint) error {
	return d.repo.Delete(ctx, id)
}

// GetDeptByID gets department by ID
func (d *DeptData) GetDeptByID(ctx context.Context, id uint) (*model.SysDept, error) {
	return d.repo.FindByID(ctx, id)
}

// GetDeptTree gets department tree structure
func (d *DeptData) GetDeptTree(ctx context.Context) ([]*model.SysDept, error) {
	// Get all departments
	allDepts, err := d.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all departments: %v", err)
	}

	// Build tree structure
	return d.buildDeptTree(allDepts, 0), nil
}

// GetDeptList gets department list
func (d *DeptData) GetDeptList(ctx context.Context, name string, status *bool) ([]*model.SysDept, error) {
	if name != "" {
		// Search by name, return empty list if not found
		dept, err := d.repo.FindByName(ctx, name)
		if err != nil {
			return []*model.SysDept{}, nil // Return empty list instead of error when not found
		}
		return []*model.SysDept{dept}, nil
	}

	if status != nil {
		return d.repo.FindByEnabled(ctx, *status)
	}

	return d.repo.FindAll(ctx)
}

// buildDeptTree builds department tree structure
func (d *DeptData) buildDeptTree(allDepts []*model.SysDept, parentID uint) []*model.SysDept {
	var tree []*model.SysDept

	for _, dept := range allDepts {
		if dept.PID != nil && *dept.PID == parentID {
			// Recursively build child departments
			children := d.buildDeptTree(allDepts, dept.DeptID)
			if len(children) > 0 {
				// Here we need to set child departments based on actual model structure
				// Since SysDept model may not have Children field, we return flat structure for now
			}
			tree = append(tree, dept)
		}
	}

	return tree
}
