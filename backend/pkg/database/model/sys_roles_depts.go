package model

import (
	"fmt"
)

// SysRolesDepts 角色部门关联表模型
type SysRolesDepts struct {
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:角色ID" json:"roleId"`
	DeptID uint `gorm:"column:dept_id;primaryKey;not null;comment:部门ID" json:"deptId"`
}

// TableName 返回表名
func (SysRolesDepts) TableName() string {
	return "sys_roles_depts"
}

// ValidateForCreate 创建前验证
func (rd *SysRolesDepts) ValidateForCreate() error {
	if rd.RoleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}
	if rd.DeptID == 0 {
		return fmt.Errorf("部门ID不能为空")
	}
	return nil
}

// ValidateForUpdate 更新前验证
func (rd *SysRolesDepts) ValidateForUpdate() error {
	return rd.ValidateForCreate()
}

// Clone 克隆对象
func (rd *SysRolesDepts) Clone() *SysRolesDepts {
	if rd == nil {
		return nil
	}
	return &SysRolesDepts{
		RoleID: rd.RoleID,
		DeptID: rd.DeptID,
	}
}

// GetRoleID 获取角色ID
func (rd *SysRolesDepts) GetRoleID() uint {
	return rd.RoleID
}

// GetDeptID 获取部门ID
func (rd *SysRolesDepts) GetDeptID() uint {
	return rd.DeptID
}

// SetRoleID 设置角色ID
func (rd *SysRolesDepts) SetRoleID(roleID uint) {
	rd.RoleID = roleID
}

// SetDeptID 设置部门ID
func (rd *SysRolesDepts) SetDeptID(deptID uint) {
	rd.DeptID = deptID
}

// IsValid 检查关联是否有效
func (rd *SysRolesDepts) IsValid() bool {
	return rd.RoleID > 0 && rd.DeptID > 0
}

// String 返回字符串表示
func (rd *SysRolesDepts) String() string {
	return fmt.Sprintf("SysRolesDepts{RoleID: %d, DeptID: %d}", rd.RoleID, rd.DeptID)
}