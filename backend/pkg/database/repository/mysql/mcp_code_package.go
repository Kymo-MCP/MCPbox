package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"

	"gorm.io/gorm"
)

var McpCodePackageRepo *McpCodePackageRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewMcpCodePackageRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize code_package table: %v", err))
		}
	})
}

// McpCodePackageRepository 封装 code_package 表的增删改查操作
type McpCodePackageRepository struct {
	db *gorm.DB
}

// NewMcpCodePackageRepository 创建 McpCodePackageRepository 实例
func NewMcpCodePackageRepository(db *gorm.DB) *McpCodePackageRepository {
	if McpCodePackageRepo == nil {
		// 优先使用传入的db参数，如果为nil则使用GetDB()
		var dbInstance *gorm.DB
		if db != nil {
			dbInstance = db
		} else {
			dbInstance = GetDB()
		}

		// 确保数据库实例不为nil
		if dbInstance == nil {
			panic("database instance is nil, please initialize database first")
		}

		m := &model.McpCodePackage{}
		McpCodePackageRepo = &McpCodePackageRepository{
			db: dbInstance.Table(m.TableName()).Model(m),
		}
	}
	return McpCodePackageRepo
}

// Create 创建代码包记录
func (r *McpCodePackageRepository) Create(ctx context.Context, pkg *model.McpCodePackage) error {
	pkg.PrepareForCreate()
	return r.db.WithContext(ctx).Create(pkg).Error
}

// FindByPackageID 根据包ID查找代码包
func (r *McpCodePackageRepository) FindByPackageID(ctx context.Context, packageID string) (*model.McpCodePackage, error) {
	var pkg model.McpCodePackage
	if err := r.db.WithContext(ctx).Where("package_id = ? AND is_deleted = false", packageID).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("package not found: %s", packageID)
		}
		return nil, fmt.Errorf("failed to find package: %v", err)
	}
	return &pkg, nil
}

// FindByOriginalName finds code package by original name
func (r *McpCodePackageRepository) FindByOriginalName(ctx context.Context, originalName string) (*model.McpCodePackage, error) {
	var pkg model.McpCodePackage
	if err := r.db.WithContext(ctx).Where("original_name = ? AND is_deleted = false", originalName).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("package not found with original name: %s", originalName)
		}
		return nil, fmt.Errorf("failed to find package by original name: %v", err)
	}
	return &pkg, nil
}

// Update 更新代码包记录
func (r *McpCodePackageRepository) Update(ctx context.Context, pkg *model.McpCodePackage) error {
	pkg.PrepareForUpdate()
	return r.db.WithContext(ctx).Save(pkg).Error
}

// Delete 软删除代码包记录
func (r *McpCodePackageRepository) Delete(ctx context.Context, pkg *model.McpCodePackage) error {
	pkg.PrepareForDelete()
	return r.db.WithContext(ctx).Save(pkg).Error
}

// DeleteByInstanceID 根据实例ID软删除代码包记录
func (r *McpCodePackageRepository) DeleteByInstanceID(ctx context.Context, instanceID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("instance_id = ? AND is_deleted = false", instanceID).
		Updates(map[string]interface{}{
			"updated_at": now,
			"is_deleted": true,
		}).Error
}

// DeleteByPackageID soft deletes a code package by package ID
func (r *McpCodePackageRepository) DeleteByPackageID(ctx context.Context, packageID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("package_id = ? AND is_deleted = false", packageID).
		Updates(map[string]interface{}{
			"updated_at": now,
			"is_deleted": true,
		}).Error
}

// FindAll 查找所有有效的代码包记录
func (r *McpCodePackageRepository) FindAll(ctx context.Context) ([]*model.McpCodePackage, error) {
	var packages []*model.McpCodePackage
	err := r.db.WithContext(ctx).Where("is_deleted = false").Find(&packages).Error
	if err != nil {
		return nil, err
	}
	return packages, nil
}

// FindWithPagination 分页查询代码包记录
func (r *McpCodePackageRepository) FindWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}) ([]*model.McpCodePackage, int64, error) {
	var packages []*model.McpCodePackage
	var total int64

	query := r.db.WithContext(ctx).Where("is_deleted = false")

	// 如果有关键词，添加搜索条件
	for key, value := range filters {
		switch key {
		case "keyword":
			if keyword, ok := value.(string); ok && keyword != "" {
				query = query.Where("original_name LIKE ? OR package_id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
			}
		case "packageType":
			if packageTypes, ok := value.([]model.PackageType); ok && len(packageTypes) > 0 {
				query = query.Where("package_type IN ?", packageTypes)
			}
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&packages).Error
	if err != nil {
		return nil, 0, err
	}

	return packages, total, nil
}

// InitTable 初始化表结构
func (r *McpCodePackageRepository) InitTable() error {
	// 创建表
	mod := &model.McpCodePackage{}
	if err := r.db.AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查包ID索引是否存在
	var count int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_code_package_package_id'", mod.TableName())
	r.db.Raw(sql).Count(&count)
	if count == 0 {
		// 创建包ID索引
		sql2 := fmt.Sprintf("CREATE UNIQUE INDEX idx_code_package_package_id ON %v(package_id)", mod.TableName())
		if err := r.db.Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create package_id index: %v", err)
		}
	}

	return nil
}
