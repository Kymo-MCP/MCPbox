package model

import (
	"fmt"
	"time"
)

// DeptSource 部门来源类型
type DeptSource string

const (
	// 自建部门
	DeptSourcePlatform DeptSource = "PLATFORM"
	// 飞书部门
	DeptSourceFeishu DeptSource = "FEISHU"
)

// SysDept 部门数据库模型
type SysDept struct {
	DeptID           uint       `gorm:"primarykey;autoIncrement;column:dept_id;comment:ID" json:"deptId"`
	PID              *uint      `gorm:"column:pid;comment:上级部门" json:"pid"`
	SubCount         int        `gorm:"column:sub_count;default:0;comment:子部门数目" json:"subCount"`
	Name             string     `gorm:"column:name;size:255;not null;comment:名称" json:"name"`
	DeptSort         int        `gorm:"column:dept_sort;default:999;comment:排序" json:"deptSort"`
	Enabled          bool       `gorm:"column:enabled;not null;comment:状态" json:"enabled"`
	CreateBy         *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy         *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	CreateTime       *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime       *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	ImageURL         *string    `gorm:"column:image_url;size:255;comment:图片" json:"imageUrl"`
	Source           DeptSource `gorm:"column:source;size:32;not null;comment:部门来源 PLATFORM：自建，FEISHU:飞书" json:"source"`
	CorpID           *string    `gorm:"column:corp_id;size:255;comment:第三方来源配置标识" json:"corpId"`
	OpenDepartmentID *string    `gorm:"column:open_department_id;size:255;comment:第三方部门id" json:"openDepartmentId"`
}

// TableName 指定表名
func (SysDept) TableName() string {
	return "sys_dept"
}

// PrepareForCreate 创建前的准备工作
func (d *SysDept) PrepareForCreate() error {
	now := time.Now()
	d.CreateTime = &now
	d.UpdateTime = &now

	// 设置默认值
	if d.Source == "" {
		d.Source = DeptSourcePlatform
	}

	return nil
}

// PrepareForUpdate 更新前的准备工作
func (d *SysDept) PrepareForUpdate() error {
	now := time.Now()
	d.UpdateTime = &now
	return nil
}

// ValidateForCreate 创建时的验证
func (d *SysDept) ValidateForCreate() error {
	if d.Name == "" {
		return fmt.Errorf("部门名称不能为空")
	}

	if d.Source != DeptSourcePlatform && d.Source != DeptSourceFeishu {
		return fmt.Errorf("无效的部门来源: %s", d.Source)
	}

	// 如果是第三方来源，需要验证相关字段
	if d.Source == DeptSourceFeishu {
		if d.CorpID == nil || *d.CorpID == "" {
			return fmt.Errorf("飞书部门必须提供企业ID")
		}
		if d.OpenDepartmentID == nil || *d.OpenDepartmentID == "" {
			return fmt.Errorf("飞书部门必须提供开放部门ID")
		}
	}

	return nil
}

// ValidateForUpdate 更新时的验证
func (d *SysDept) ValidateForUpdate() error {
	if d.Name == "" {
		return fmt.Errorf("部门名称不能为空")
	}

	if d.Source != DeptSourcePlatform && d.Source != DeptSourceFeishu {
		return fmt.Errorf("无效的部门来源: %s", d.Source)
	}

	return nil
}

// IsThirdParty 判断是否为第三方部门
func (d *SysDept) IsThirdParty() bool {
	return d.Source != DeptSourcePlatform
}

// IsFeishuDept 判断是否为飞书部门
func (d *SysDept) IsFeishuDept() bool {
	return d.Source == DeptSourceFeishu
}

// HasParent 判断是否有上级部门
func (d *SysDept) HasParent() bool {
	return d.PID != nil && *d.PID > 0
}

// HasChildren 判断是否有子部门
func (d *SysDept) HasChildren() bool {
	return d.SubCount > 0
}

// IsActive 判断部门是否激活
func (d *SysDept) IsActive() bool {
	return d.Enabled
}

// GetImageURL 获取图片URL，如果为空则返回默认值
func (d *SysDept) GetImageURL() string {
	if d.ImageURL != nil {
		return *d.ImageURL
	}
	return ""
}

// GetCorpID 获取企业ID，如果为空则返回默认值
func (d *SysDept) GetCorpID() string {
	if d.CorpID != nil {
		return *d.CorpID
	}
	return ""
}

// GetOpenDepartmentID 获取开放部门ID，如果为空则返回默认值
func (d *SysDept) GetOpenDepartmentID() string {
	if d.OpenDepartmentID != nil {
		return *d.OpenDepartmentID
	}
	return ""
}

// Clone 克隆部门对象
func (d *SysDept) Clone() *SysDept {
	clone := &SysDept{
		DeptID:   d.DeptID,
		SubCount: d.SubCount,
		Name:     d.Name,
		DeptSort: d.DeptSort,
		Enabled:  d.Enabled,
		Source:   d.Source,
	}

	// 复制指针字段
	if d.PID != nil {
		pid := *d.PID
		clone.PID = &pid
	}

	if d.CreateBy != nil {
		createBy := *d.CreateBy
		clone.CreateBy = &createBy
	}

	if d.UpdateBy != nil {
		updateBy := *d.UpdateBy
		clone.UpdateBy = &updateBy
	}

	if d.CreateTime != nil {
		createTime := *d.CreateTime
		clone.CreateTime = &createTime
	}

	if d.UpdateTime != nil {
		updateTime := *d.UpdateTime
		clone.UpdateTime = &updateTime
	}

	if d.ImageURL != nil {
		imageURL := *d.ImageURL
		clone.ImageURL = &imageURL
	}

	if d.CorpID != nil {
		corpID := *d.CorpID
		clone.CorpID = &corpID
	}

	if d.OpenDepartmentID != nil {
		openDeptID := *d.OpenDepartmentID
		clone.OpenDepartmentID = &openDeptID
	}

	return clone
}

// SetParent 设置上级部门
func (d *SysDept) SetParent(parentID uint) {
	d.PID = &parentID
}

// ClearParent 清除上级部门
func (d *SysDept) ClearParent() {
	d.PID = nil
}

// SetImageURL 设置图片URL
func (d *SysDept) SetImageURL(url string) {
	if url == "" {
		d.ImageURL = nil
	} else {
		d.ImageURL = &url
	}
}

// SetCorpID 设置企业ID
func (d *SysDept) SetCorpID(corpID string) {
	if corpID == "" {
		d.CorpID = nil
	} else {
		d.CorpID = &corpID
	}
}

// SetOpenDepartmentID 设置开放部门ID
func (d *SysDept) SetOpenDepartmentID(openDeptID string) {
	if openDeptID == "" {
		d.OpenDepartmentID = nil
	} else {
		d.OpenDepartmentID = &openDeptID
	}
}

// SetCreateBy 设置创建者
func (d *SysDept) SetCreateBy(createBy string) {
	if createBy == "" {
		d.CreateBy = nil
	} else {
		d.CreateBy = &createBy
	}
}

// SetUpdateBy 设置更新者
func (d *SysDept) SetUpdateBy(updateBy string) {
	if updateBy == "" {
		d.UpdateBy = nil
	} else {
		d.UpdateBy = &updateBy
	}
}
