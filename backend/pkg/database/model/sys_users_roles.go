package model

import (
	"fmt"
)

// SysUsersRoles 用户角色关联表模型
type SysUsersRoles struct {
	UserID uint `gorm:"column:user_id;primaryKey;not null;comment:用户ID" json:"userId"`
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:角色ID" json:"roleId"`
}

// TableName 返回表名
func (SysUsersRoles) TableName() string {
	return "sys_users_roles"
}

// ValidateForCreate 创建前验证
func (ur *SysUsersRoles) ValidateForCreate() error {
	if ur.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if ur.RoleID == 0 {
		return fmt.Errorf("角色ID不能为空")
	}
	return nil
}

// ValidateForUpdate 更新前验证
func (ur *SysUsersRoles) ValidateForUpdate() error {
	return ur.ValidateForCreate()
}

// Clone 克隆对象
func (ur *SysUsersRoles) Clone() *SysUsersRoles {
	if ur == nil {
		return nil
	}
	return &SysUsersRoles{
		UserID: ur.UserID,
		RoleID: ur.RoleID,
	}
}

// GetUserID 获取用户ID
func (ur *SysUsersRoles) GetUserID() uint {
	return ur.UserID
}

// GetRoleID 获取角色ID
func (ur *SysUsersRoles) GetRoleID() uint {
	return ur.RoleID
}

// SetUserID 设置用户ID
func (ur *SysUsersRoles) SetUserID(userID uint) {
	ur.UserID = userID
}

// SetRoleID 设置角色ID
func (ur *SysUsersRoles) SetRoleID(roleID uint) {
	ur.RoleID = roleID
}

// IsValid 检查关联是否有效
func (ur *SysUsersRoles) IsValid() bool {
	return ur.UserID > 0 && ur.RoleID > 0
}

// String 返回字符串表示
func (ur *SysUsersRoles) String() string {
	return fmt.Sprintf("SysUsersRoles{UserID: %d, RoleID: %d}", ur.UserID, ur.RoleID)
}
