package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"

	"gorm.io/gorm"
)

var McpTemplateRepo *McpTemplateRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewMcpTemplateRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_template table: %v", err))
		}
	})
}

// McpTemplateRepository 封装 mcp_template 表的增删改查操作
type McpTemplateRepository struct{}

// NewMcpTemplateRepository 创建 McpTemplateRepository 实例
func NewMcpTemplateRepository() *McpTemplateRepository {
	return &McpTemplateRepository{}
}

// getDB 获取模型
func (r *McpTemplateRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpTemplate{})
}

// Create 创建模板
func (r *McpTemplateRepository) Create(ctx context.Context, template *model.McpTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	// 执行创建操作
	err := r.getDB().WithContext(ctx).Create(template).Error
	if err != nil {
		// 记录更详细的错误信息，包含模板数据
		return fmt.Errorf("failed to create template [name=%s, mcp_server_id=%v, environment_id=%v]: %v",
			template.Name, template.McpServerID, template.EnvironmentID, err)
	}

	return nil
}

// Update 更新模板
func (r *McpTemplateRepository) Update(ctx context.Context, template *model.McpTemplate) error {
	template.UpdatedAt = time.Now()
	return r.getDB().WithContext(ctx).Where("id = ?", template.ID).Updates(template).Error
}

// Delete 删除模板
func (r *McpTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.McpTemplate{}).Error
}

// FindByID 根据ID查找模板
func (r *McpTemplateRepository) FindByID(ctx context.Context, id uint) (*model.McpTemplate, error) {
	var template model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByName 根据名称查找模板
func (r *McpTemplateRepository) FindByName(ctx context.Context, name string) (*model.McpTemplate, error) {
	var template model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("name = ?", name).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// FindByMcpServerID 根据MCP服务器ID查找模板
func (r *McpTemplateRepository) FindByMcpServerID(ctx context.Context, mcpServerID string) (*model.McpTemplate, error) {
	var template model.McpTemplate
	if err := r.getDB().WithContext(ctx).Where("mcp_server_id = ?", mcpServerID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found: %s", mcpServerID)
		}
		return nil, fmt.Errorf("failed to find template: %v", err)
	}
	return &template, nil
}

// FindAll 查找所有模板
func (r *McpTemplateRepository) FindAll(ctx context.Context) ([]*model.McpTemplate, error) {
	var templates []*model.McpTemplate
	err := r.getDB().WithContext(ctx).Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FindByAccessType 根据访问类型查找模板
func (r *McpTemplateRepository) FindByAccessType(ctx context.Context, accessType model.AccessType) ([]*model.McpTemplate, error) {
	var templates []*model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("access_type = ?", accessType).Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FindBySourceType 根据来源类型查找模板
func (r *McpTemplateRepository) FindBySourceType(ctx context.Context, sourceType model.SourceType) ([]*model.McpTemplate, error) {
	var templates []*model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("source_type = ?", sourceType).Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FindByEnvironmentID 根据环境ID查找模板
func (r *McpTemplateRepository) FindByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpTemplate, error) {
	var templates []*model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("environment_id = ?", environmentID).Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FindByPackageID finds templates by package ID
func (r *McpTemplateRepository) FindByPackageID(ctx context.Context, packageID string) ([]*model.McpTemplate, error) {
	var templates []*model.McpTemplate
	err := r.getDB().WithContext(ctx).Where("package_id = ?", packageID).Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FindWithPagination 分页查询模板
func (r *McpTemplateRepository) FindWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) ([]*model.McpTemplate, int64, error) {
	var templates []*model.McpTemplate
	var total int64

	// 构建查询条件
	query := r.getDB().WithContext(ctx)

	// 应用筛选条件
	for key, value := range filters {
		switch key {
		case "keyword":
			if keyword, ok := value.(string); ok && keyword != "" {
				query = query.Where("name LIKE ? OR mcp_server_id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
			}
		case "environment_id":
			if envId, ok := value.(uint); ok && envId > 0 {
				query = query.Where("environment_id = ?", envId)
			}
		case "environmentId":
			if envId, ok := value.(int64); ok && envId > 0 {
				query = query.Where("environment_id = ?", envId)
			}
		case "template_id":
			if templateId, ok := value.(int32); ok && templateId > 0 {
				query = query.Where("id = ?", templateId)
			}
		case "name":
			if name, ok := value.(string); ok && name != "" {
				query = query.Where("name LIKE ?", "%"+name+"%")
			}
		case "mcpProtocol":
			if mcpProtocol, ok := value.(model.McpProtocol); ok {
				query = query.Where("mcp_protocol = ?", mcpProtocol)
			}
		case "accessType":
			if accessType, ok := value.(model.AccessType); ok {
				query = query.Where("access_type = ?", accessType)
			}
		case "sourceType":
			if sourceType, ok := value.(model.SourceType); ok {
				query = query.Where("source_type = ?", sourceType)
			}
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if sortBy != "" {
		order := "ASC"
		if sortOrder == "desc" {
			order = "DESC"
		}
		switch sortBy {
		case "createdAt":
			query = query.Order(fmt.Sprintf("created_at %s", order))
		case "updatedAt":
			query = query.Order(fmt.Sprintf("updated_at %s", order))
		case "name":
			query = query.Order(fmt.Sprintf("name %s", order))
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// InitTable 初始化表结构
func (r *McpTemplateRepository) InitTable() error {
	// 创建表
	mod := &model.McpTemplate{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查索引是否存在
	var count int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_template_mcp_server_id'", (&model.McpTemplate{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建索引
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_template_mcp_server_id ON %v(mcp_server_id)", (&model.McpTemplate{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// 检查环境ID索引是否存在
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_template_environment_id'", (&model.McpTemplate{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建索引
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_template_environment_id ON %v(environment_id)", (&model.McpTemplate{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// 检查 name 唯一索引是否存在
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_template_name'", (&model.McpTemplate{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE UNIQUE INDEX idx_mcp_template_name ON %v(name)", (&model.McpTemplate{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
