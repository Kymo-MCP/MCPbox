package model

import (
	"fmt"
)

// SysUsersRoles user role association table model
type SysUsersRoles struct {
	UserID uint `gorm:"column:user_id;primaryKey;not null;comment:用户ID" json:"userId"`
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:角色ID" json:"roleId"`
}

// TableName returns table name
func (SysUsersRoles) TableName() string {
	return "sys_users_roles"
}

// ValidateForCreate validates before creation
func (ur *SysUsersRoles) ValidateForCreate() error {
	if ur.UserID == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if ur.RoleID == 0 {
		return fmt.Errorf("role ID cannot be empty")
	}
	return nil
}

// ValidateForUpdate validates before update
func (ur *SysUsersRoles) ValidateForUpdate() error {
	return ur.ValidateForCreate()
}

// Clone clones object
func (ur *SysUsersRoles) Clone() *SysUsersRoles {
	if ur == nil {
		return nil
	}
	return &SysUsersRoles{
		UserID: ur.UserID,
		RoleID: ur.RoleID,
	}
}

// GetUserID gets user ID
func (ur *SysUsersRoles) GetUserID() uint {
	return ur.UserID
}

// GetRoleID gets role ID
func (ur *SysUsersRoles) GetRoleID() uint {
	return ur.RoleID
}

// SetUserID sets user ID
func (ur *SysUsersRoles) SetUserID(userID uint) {
	ur.UserID = userID
}

// SetRoleID sets role ID
func (ur *SysUsersRoles) SetRoleID(roleID uint) {
	ur.RoleID = roleID
}

// IsValid checks if association is valid
func (ur *SysUsersRoles) IsValid() bool {
	return ur.UserID > 0 && ur.RoleID > 0
}

// String returns string representation
func (ur *SysUsersRoles) String() string {
	return fmt.Sprintf("SysUsersRoles{UserID: %d, RoleID: %d}", ur.UserID, ur.RoleID)
}
