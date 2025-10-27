package model

import (
	"fmt"
	"time"
)

// PackageType 代码包类型
type PackageType string

func (p PackageType) String() string {
	return string(p)
}

const (
	PackageTypeUnknown PackageType = "unknown"
	PackageTypeTar     PackageType = "tar"
	PackageTypeZip     PackageType = "zip"
	PackageTypeTarGz   PackageType = "tar.gz"
	PackageTypeDxt     PackageType = "dxt"
	PackageTypeMcpb    PackageType = "mcpb"
)

// McpCodePackage 代码包数据库模型
type McpCodePackage struct {
	ID            uint        `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	PackageID     string      `gorm:"size:100;not null;unique;comment:包ID" json:"packageId"`
	PackageType   PackageType `gorm:"size:10;not null;comment:包类型 (tar/zip)" json:"packageType"`
	PackagePath   string      `gorm:"size:500;not null;comment:包存储目录路径" json:"packagePath"`
	OriginalPath  string      `gorm:"size:500;comment:原始压缩包文件路径" json:"originalPath"`
	ExtractedPath string      `gorm:"size:500;comment:解压后的绝对路径" json:"extractedPath"`
	OriginalName  string      `gorm:"size:255;comment:原始文件名" json:"originalName"`
	FileSize      int64       `gorm:"comment:文件大小(字节)" json:"fileSize"`
	IsDeleted     bool        `gorm:"default:false;comment:是否删除" json:"isDeleted"`
	CreatedAt     time.Time   `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt     time.Time   `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

// TableName 指定表名
func (McpCodePackage) TableName() string {
	return "mcp_code_package"
}

// IsDeletedRecord 检查代码包是否已被删除
func (c *McpCodePackage) IsDeletedRecord() bool {
	return c.IsDeleted
}

// SetCreatedAt 设置创建时间为当前时间
func (c *McpCodePackage) SetCreatedAt() {
	c.CreatedAt = time.Now()
}

// SetUpdatedAt 设置更新时间为当前时间
func (c *McpCodePackage) SetUpdatedAt() {
	c.UpdatedAt = time.Now()
}

// SetDeleted 设置删除状态
func (c *McpCodePackage) SetDeleted() {
	c.IsDeleted = true
	c.UpdatedAt = time.Now()
}

// ClearDeleted 清除删除状态（用于恢复）
func (c *McpCodePackage) ClearDeleted() {
	c.IsDeleted = false
	c.UpdatedAt = time.Now()
}

// PrepareForCreate 准备创建记录（设置创建和更新时间）
func (c *McpCodePackage) PrepareForCreate() {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.IsDeleted = false
}

// PrepareForUpdate 准备更新记录（设置更新时间）
func (c *McpCodePackage) PrepareForUpdate() {
	c.UpdatedAt = time.Now()
}

// PrepareForDelete 准备删除记录（设置删除状态）
func (c *McpCodePackage) PrepareForDelete() {
	c.SetDeleted()
}

// ValidateForCreate 验证创建代码包的必要字段
func (c *McpCodePackage) ValidateForCreate() error {
	if c.PackageID == "" {
		return fmt.Errorf("package ID is required")
	}
	if c.PackageType == "" {
		return fmt.Errorf("package type is required")
	}
	if c.PackagePath == "" {
		return fmt.Errorf("package path is required")
	}

	// 验证包类型
	if c.PackageType != PackageTypeTar && c.PackageType != PackageTypeZip {
		return fmt.Errorf("invalid package type: %s", c.PackageType)
	}

	return nil
}

// ValidateForUpdate 验证更新代码包的必要字段
func (c *McpCodePackage) ValidateForUpdate() error {
	if c.ID == 0 {
		return fmt.Errorf("code package ID is required for update")
	}

	return c.ValidateForCreate()
}
