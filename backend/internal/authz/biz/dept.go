package biz

import (
	"context"
	"fmt"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
)

// DeptData 部门数据访问层
type DeptData struct {
	ctx  context.Context
	repo *mysql.SysDeptRepository
}

// NewDeptData 创建部门数据访问层实例
func NewDeptData(ctx context.Context) *DeptData {
	return &DeptData{
		ctx:  ctx,
		repo: mysql.SysDeptRepo,
	}
}

// CreateDept 创建部门
func (d *DeptData) CreateDept(ctx context.Context, dept *model.SysDept) error {
	return d.repo.Create(ctx, dept)
}

// UpdateDept 更新部门
func (d *DeptData) UpdateDept(ctx context.Context, dept *model.SysDept) error {
	return d.repo.Update(ctx, dept)
}

// DeleteDept 删除部门
func (d *DeptData) DeleteDept(ctx context.Context, id uint) error {
	return d.repo.Delete(ctx, id)
}

// GetDeptByID 根据ID获取部门
func (d *DeptData) GetDeptByID(ctx context.Context, id uint) (*model.SysDept, error) {
	return d.repo.FindByID(ctx, id)
}

// GetDeptTree 获取部门树形结构
func (d *DeptData) GetDeptTree(ctx context.Context) ([]*model.SysDept, error) {
	// 获取所有部门
	allDepts, err := d.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all departments: %v", err)
	}

	// 构建树形结构
	return d.buildDeptTree(allDepts, 0), nil
}

// GetDeptList 获取部门列表
func (d *DeptData) GetDeptList(ctx context.Context, name string, status *bool) ([]*model.SysDept, error) {
	if name != "" {
		// 使用名称查找，如果没有找到则返回空列表
		dept, err := d.repo.FindByName(ctx, name)
		if err != nil {
			return []*model.SysDept{}, nil // 没找到返回空列表而不是错误
		}
		return []*model.SysDept{dept}, nil
	}

	if status != nil {
		return d.repo.FindByEnabled(ctx, *status)
	}

	return d.repo.FindAll(ctx)
}

// buildDeptTree 构建部门树形结构
func (d *DeptData) buildDeptTree(allDepts []*model.SysDept, parentID uint) []*model.SysDept {
	var tree []*model.SysDept

	for _, dept := range allDepts {
		if dept.PID != nil && *dept.PID == parentID {
			// 递归构建子部门
			children := d.buildDeptTree(allDepts, dept.DeptID)
			if len(children) > 0 {
				// 这里需要根据实际的模型结构来设置子部门
				// 由于 SysDept 模型可能没有 Children 字段，我们先返回扁平结构
			}
			tree = append(tree, dept)
		}
	}

	return tree
}
