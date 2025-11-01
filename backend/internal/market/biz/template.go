package biz

import (
	"context"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// TemplateBiz template data access layer
type TemplateBiz struct {
	ctx context.Context
}

// GTemplateBiz global template data access layer instance
var GTemplateBiz *TemplateBiz

func init() {
	GTemplateBiz = NewTemplateBiz(context.Background())
}

// NewTemplateBiz create template data access layer instance
func NewTemplateBiz(ctx context.Context) *TemplateBiz {
	return &TemplateBiz{
		ctx: ctx,
	}
}

// CreateTemplate create template
func (biz *TemplateBiz) CreateTemplate(ctx context.Context, template *model.McpTemplate) error {
	return mysql.McpTemplateRepo.Create(ctx, template)
}

// GetTemplateByID get template by ID
func (biz *TemplateBiz) GetTemplateByID(ctx context.Context, id uint) (*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByID(ctx, id)
}

// GetTemplateByName get template by name
func (biz *TemplateBiz) GetTemplateByName(ctx context.Context, name string) (*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByName(ctx, name)
}

// UpdateTemplate update template
func (biz *TemplateBiz) UpdateTemplate(ctx context.Context, template *model.McpTemplate) error {
	template.UpdatedAt = time.Now()
	return mysql.McpTemplateRepo.Update(ctx, template)
}

// DeleteTemplate delete template
func (biz *TemplateBiz) DeleteTemplate(ctx context.Context, id uint) error {
	return mysql.McpTemplateRepo.Delete(ctx, id)
}

// GetAllTemplates get all templates
func (biz *TemplateBiz) GetAllTemplates(ctx context.Context) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindAll(ctx)
}

// GetTemplatesByEnvironmentID get template list by environment ID
func (biz *TemplateBiz) GetTemplatesByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByEnvironmentID(ctx, environmentID)
}

// GetTemplatesByAccessType get template list by access type
func (biz *TemplateBiz) GetTemplatesByAccessType(ctx context.Context, accessType model.AccessType) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindByAccessType(ctx, accessType)
}

// GetTemplatesBySourceType get template list by source type
func (biz *TemplateBiz) GetTemplatesBySourceType(ctx context.Context, sourceType model.SourceType) ([]*model.McpTemplate, error) {
	return mysql.McpTemplateRepo.FindBySourceType(ctx, sourceType)
}

// GetTemplatesWithPagination get template list with pagination
func (biz *TemplateBiz) GetTemplatesWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) ([]*model.McpTemplate, int64, error) {
	return mysql.McpTemplateRepo.FindWithPagination(ctx, page, pageSize, filters, sortBy, sortOrder)
}
