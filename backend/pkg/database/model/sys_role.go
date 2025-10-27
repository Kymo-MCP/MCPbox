package model

import (
	"fmt"
	"time"
)

// DataScope 数据权限范围类型
type DataScope string

const (
	// 全部数据权限
	DataScopeAll DataScope = "all"
	// 自定义数据权限
	DataScopeCustom DataScope = "custom"
	// 本部门数据权限
	DataScopeDept DataScope = "dept"
	// 本部门及以下数据权限
	DataScopeDeptAndChild DataScope = "dept_and_child"
	// 仅本人数据权限
	DataScopeSelf DataScope = "self"
)

// SysRole 角色数据库模型
type SysRole struct {
	RoleID      uint       `gorm:"primarykey;autoIncrement;column:role_id;comment:ID" json:"roleId"`
	Name        string     `gorm:"column:name;size:100;not null;uniqueIndex:uniq_name;index:role_name_index;comment:名称" json:"name"`
	Level       *int       `gorm:"column:level;comment:角色级别" json:"level"`
	Description *string    `gorm:"column:description;size:255;comment:描述" json:"description"`
	DataScope   *string    `gorm:"column:data_scope;size:255;comment:数据权限" json:"dataScope"`
	CreateBy    *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy    *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	CreateTime  *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime  *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
}

// TableName 指定表名
func (SysRole) TableName() string {
	return "sys_role"
}

// PrepareForCreate 创建前的准备工作
func (r *SysRole) PrepareForCreate() error {
	now := time.Now()
	r.CreateTime = &now
	r.UpdateTime = &now

	// 设置默认数据权限
	if r.DataScope == nil || *r.DataScope == "" {
		defaultScope := string(DataScopeDept)
		r.DataScope = &defaultScope
	}

	return nil
}

// PrepareForUpdate 更新前的准备工作
func (r *SysRole) PrepareForUpdate() error {
	now := time.Now()
	r.UpdateTime = &now
	return nil
}

// ValidateForCreate 创建时的验证
func (r *SysRole) ValidateForCreate() error {
	if r.Name == "" {
		return fmt.Errorf("角色名称不能为空")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("角色名称长度不能超过100个字符")
	}

	// 验证数据权限范围
	if r.DataScope != nil && *r.DataScope != "" {
		if !r.isValidDataScope(*r.DataScope) {
			return fmt.Errorf("无效的数据权限范围: %s", *r.DataScope)
		}
	}

	// 验证角色级别
	if r.Level != nil && *r.Level < 0 {
		return fmt.Errorf("角色级别不能为负数")
	}

	return nil
}

// ValidateForUpdate 更新时的验证
func (r *SysRole) ValidateForUpdate() error {
	if r.Name == "" {
		return fmt.Errorf("角色名称不能为空")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("角色名称长度不能超过100个字符")
	}

	// 验证数据权限范围
	if r.DataScope != nil && *r.DataScope != "" {
		if !r.isValidDataScope(*r.DataScope) {
			return fmt.Errorf("无效的数据权限范围: %s", *r.DataScope)
		}
	}

	// 验证角色级别
	if r.Level != nil && *r.Level < 0 {
		return fmt.Errorf("角色级别不能为负数")
	}

	return nil
}

// isValidDataScope 验证数据权限范围是否有效
func (r *SysRole) isValidDataScope(scope string) bool {
	validScopes := []DataScope{
		DataScopeAll,
		DataScopeCustom,
		DataScopeDept,
		DataScopeDeptAndChild,
		DataScopeSelf,
	}

	for _, validScope := range validScopes {
		if scope == string(validScope) {
			return true
		}
	}
	return false
}

// HasLevel 判断是否设置了角色级别
func (r *SysRole) HasLevel() bool {
	return r.Level != nil
}

// GetLevel 获取角色级别，如果未设置则返回默认值
func (r *SysRole) GetLevel() int {
	if r.Level != nil {
		return *r.Level
	}
	return 0
}

// GetDescription 获取描述，如果为空则返回默认值
func (r *SysRole) GetDescription() string {
	if r.Description != nil {
		return *r.Description
	}
	return ""
}

// GetDataScope 获取数据权限范围，如果为空则返回默认值
func (r *SysRole) GetDataScope() string {
	if r.DataScope != nil {
		return *r.DataScope
	}
	return string(DataScopeDept)
}

// IsDataScopeAll 判断是否为全部数据权限
func (r *SysRole) IsDataScopeAll() bool {
	return r.GetDataScope() == string(DataScopeAll)
}

// IsDataScopeCustom 判断是否为自定义数据权限
func (r *SysRole) IsDataScopeCustom() bool {
	return r.GetDataScope() == string(DataScopeCustom)
}

// IsDataScopeDept 判断是否为本部门数据权限
func (r *SysRole) IsDataScopeDept() bool {
	return r.GetDataScope() == string(DataScopeDept)
}

// IsDataScopeDeptAndChild 判断是否为本部门及以下数据权限
func (r *SysRole) IsDataScopeDeptAndChild() bool {
	return r.GetDataScope() == string(DataScopeDeptAndChild)
}

// IsDataScopeSelf 判断是否为仅本人数据权限
func (r *SysRole) IsDataScopeSelf() bool {
	return r.GetDataScope() == string(DataScopeSelf)
}

// Clone 克隆角色对象
func (r *SysRole) Clone() *SysRole {
	clone := &SysRole{
		RoleID: r.RoleID,
		Name:   r.Name,
	}

	// 复制指针字段
	if r.Level != nil {
		level := *r.Level
		clone.Level = &level
	}

	if r.Description != nil {
		description := *r.Description
		clone.Description = &description
	}

	if r.DataScope != nil {
		dataScope := *r.DataScope
		clone.DataScope = &dataScope
	}

	if r.CreateBy != nil {
		createBy := *r.CreateBy
		clone.CreateBy = &createBy
	}

	if r.UpdateBy != nil {
		updateBy := *r.UpdateBy
		clone.UpdateBy = &updateBy
	}

	if r.CreateTime != nil {
		createTime := *r.CreateTime
		clone.CreateTime = &createTime
	}

	if r.UpdateTime != nil {
		updateTime := *r.UpdateTime
		clone.UpdateTime = &updateTime
	}

	return clone
}

// SetLevel 设置角色级别
func (r *SysRole) SetLevel(level int) {
	r.Level = &level
}

// ClearLevel 清除角色级别
func (r *SysRole) ClearLevel() {
	r.Level = nil
}

// SetDescription 设置描述
func (r *SysRole) SetDescription(description string) {
	if description == "" {
		r.Description = nil
	} else {
		r.Description = &description
	}
}

// SetDataScope 设置数据权限范围
func (r *SysRole) SetDataScope(dataScope string) {
	if dataScope == "" {
		r.DataScope = nil
	} else {
		r.DataScope = &dataScope
	}
}

// SetCreateBy 设置创建者
func (r *SysRole) SetCreateBy(createBy string) {
	if createBy == "" {
		r.CreateBy = nil
	} else {
		r.CreateBy = &createBy
	}
}

// SetUpdateBy 设置更新者
func (r *SysRole) SetUpdateBy(updateBy string) {
	if updateBy == "" {
		r.UpdateBy = nil
	} else {
		r.UpdateBy = &updateBy
	}
}

// IsHigherLevelThan 判断当前角色级别是否高于指定角色
func (r *SysRole) IsHigherLevelThan(other *SysRole) bool {
	if r.Level == nil || other.Level == nil {
		return false
	}
	return *r.Level > *other.Level
}

// IsLowerLevelThan 判断当前角色级别是否低于指定角色
func (r *SysRole) IsLowerLevelThan(other *SysRole) bool {
	if r.Level == nil || other.Level == nil {
		return false
	}
	return *r.Level < *other.Level
}

// IsSameLevelAs 判断当前角色级别是否与指定角色相同
func (r *SysRole) IsSameLevelAs(other *SysRole) bool {
	if r.Level == nil && other.Level == nil {
		return true
	}
	if r.Level == nil || other.Level == nil {
		return false
	}
	return *r.Level == *other.Level
}
