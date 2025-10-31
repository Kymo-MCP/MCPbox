package biz

import (
	"context"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// TemplateBiz 模板数据访问层
type TemplateBiz struct {
	ctx context.Context
}

// GTemplateBiz 全局模板数据访问层实例
var GTemplateBiz *TemplateBiz

func init() {
	GTemplateBiz = NewTemplateBiz(context.Background())
}

// NewTemplateBiz 创建模板数据访问层实例
func NewTemplateBiz(ctx context.Context) *TemplateBiz {
	return &TemplateBiz{
		ctx: ctx,
	}
}

// CreateTemplate 创建模板
func (biz *TemplateBiz) CreateTemplate(ctx context.Context, template *model.McpTemplate) error {
	return mysql.McpTemplateRepo.Create(ctx, template)
}

// GetTemplateByID 根据ID获取模板
func (biz *TemplateBiz) GetTemplateByID(ctx context.Context, id uint) (*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByID(ctx, id)
}

// GetTemplateByName 根据名称获取模板
func (biz *TemplateBiz) GetTemplateByName(ctx context.Context, name string) (*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByName(ctx, name)
}

// UpdateTemplate 更新模板
func (biz *TemplateBiz) UpdateTemplate(ctx context.Context, template *model.McpTemplate) error {
	template.UpdatedAt = time.Now()
	return mysql.McpTemplateRepo.Update(ctx, template)
}

// DeleteTemplate 删除模板
func (biz *TemplateBiz) DeleteTemplate(ctx context.Context, id uint) error {
	return mysql.McpTemplateRepo.Delete(ctx, id)
}

// GetAllTemplates 获取所有模板
func (biz *TemplateBiz) GetAllTemplates(ctx context.Context) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindAll(ctx)
}

// GetTemplatesByEnvironmentID 根据环境ID获取模板列表
func (biz *TemplateBiz) GetTemplatesByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByEnvironmentID(ctx, environmentID)
}

// GetTemplatesByAccessType 根据访问类型获取模板列表
func (biz *TemplateBiz) GetTemplatesByAccessType(ctx context.Context, accessType model.AccessType) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByAccessType(ctx, accessType)
}

// GetTemplatesBySourceType 根据来源类型获取模板列表
func (biz *TemplateBiz) GetTemplatesBySourceType(ctx context.Context, sourceType model.SourceType) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindBySourceType(ctx, sourceType)
}

// GetTemplatesWithPagination 分页获取模板列表
func (biz *TemplateBiz) GetTemplatesWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) ([]*model.McpTemplate, int64, error) {
	return mysql.McpTemplateRepo.FindWithPagination(ctx, page, pageSize, filters, sortBy, sortOrder)
}
