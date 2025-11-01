package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/kymo-mcp/mcpcan/api/market/instance"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

// TemplateService provides template management functionality
type TemplateService struct {
	templateData *biz.TemplateBiz
	ctx          context.Context
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService(ctx context.Context) *TemplateService {
	return &TemplateService{
		templateData: biz.GTemplateBiz,
		ctx:          ctx,
	}
}

// TemplateCreate creates a new template
func (s *TemplateService) TemplateCreate(ctx context.Context, req *instance.TemplateCreateRequest) (*instance.TemplateCreateResp, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	// Check if template name already exists
	existing, err := s.templateData.GetTemplateByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check template name: %v", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("template name '%s' already exists", req.Name)
	}

	// Create template object
	template := &model.McpTemplate{
		Name:           req.Name,
		Port:           req.Port,
		InitScript:     req.InitScript,
		Command:        req.Command,
		StartupTimeout: req.StartupTimeout,
		RunningTimeout: req.RunningTimeout,
		EnvironmentID:  req.EnvironmentId,
		PackageID:      req.PackageId,
		ImgAddress:     req.ImgAddress,
		McpServerID:    req.McpServerId,
		Notes:          req.Notes,
		IconPath:       req.IconPath,
	}

	// Handle access type
	switch req.AccessType {
	case instance.AccessType_DIRECT:
		template.AccessType = model.AccessTypeDirect
	case instance.AccessType_PROXY:
		template.AccessType = model.AccessTypeProxy
	case instance.AccessType_HOSTING:
		template.AccessType = model.AccessTypeHosting
	default:
		template.AccessType = model.AccessTypeProxy // Default proxy mode
	}

	// Handle MCP protocol
	switch req.McpProtocol {
	case instance.McpProtocol_SSE:
		template.McpProtocol = model.McpProtocolSSE
	case instance.McpProtocol_STEAMABLE_HTTP:
		template.McpProtocol = model.McpProtocolStreamableHttp
	case instance.McpProtocol_STDIO:
		template.McpProtocol = model.McpProtocolStdio
	default:
		template.McpProtocol = model.McpProtocolSSE // Default SSE protocol
	}

	// Handle environment variables
	if len(req.EnvironmentVariables) > 0 {
		envBytes, err := json.Marshal(req.EnvironmentVariables)
		if err != nil {
			logger.Error("failed to marshal environment variables", zap.Error(err))
			return nil, fmt.Errorf("failed to process environment variables: %v", err)
		}
		template.EnvironmentVariables = envBytes
	}

	// Handle volume mount configuration
	if len(req.VolumeMounts) > 0 {
		volumeBytes, err := json.Marshal(req.VolumeMounts)
		if err != nil {
			logger.Error("failed to marshal volume mounts", zap.Error(err))
			return nil, fmt.Errorf("failed to process volume mounts: %v", err)
		}
		template.VolumeMounts = volumeBytes
	}

	// Handle MCP server configuration
	if req.McpServers != "" {
		template.McpServers = json.RawMessage(req.McpServers)
	}

	// Handle tokens
	if len(req.Tokens) > 0 {
		tokens := make([]model.McpToken, 0, len(req.Tokens))
		for _, token := range req.Tokens {
			tokens = append(tokens, model.McpToken{
				Token:     token.Token,
				ExpireAt:  token.ExpireAt,
				PublishAt: token.PublishAt,
				Usages:    token.Usages,
			})
		}
		tokensJSON, err := json.Marshal(tokens)
		if err != nil {
			logger.Error("failed to marshal tokens", zap.Error(err))
			return nil, fmt.Errorf("failed to process tokens: %v", err)
		}
		template.Tokens = json.RawMessage(tokensJSON)
	}

	// Create template
	if err := s.templateData.CreateTemplate(ctx, template); err != nil {
		logger.Error("failed to create template", zap.Error(err), zap.String("name", req.Name))
		return nil, fmt.Errorf("failed to create template: %v", err)
	}

	// Return response
	resp := &instance.TemplateCreateResp{
		TemplateId: int32(template.ID),
	}

	logger.Info("template created successfully", zap.Int32("templateId", resp.TemplateId), zap.String("name", req.Name))
	return resp, nil
}

// TemplateDetail retrieves template details
func (s *TemplateService) TemplateDetail(ctx context.Context, req *instance.TemplateDetailRequest) (*instance.TemplateDetailResp, error) {
	if req.TemplateId == 0 {
		return nil, fmt.Errorf("template ID is required")
	}

	// Query template
	template, err := s.templateData.GetTemplateByID(ctx, uint(req.TemplateId))
	if err != nil {
		logger.Error("failed to get template", zap.Error(err), zap.Int32("templateId", req.TemplateId))
		return nil, fmt.Errorf("failed to get template: %v", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	// Build response
	resp := &instance.TemplateDetailResp{
		TemplateId:     int32(template.ID),
		Name:           template.Name,
		Port:           template.Port,
		InitScript:     template.InitScript,
		Command:        template.Command,
		StartupTimeout: template.StartupTimeout,
		RunningTimeout: template.RunningTimeout,
		EnvironmentId:  int32(template.EnvironmentID),
		PackageId:      template.PackageID,
		ImgAddress:     template.ImgAddress,
		McpServerId:    template.McpServerID,
		Notes:          template.Notes,
		IconPath:       template.IconPath,
		McpServers:     string(template.McpServers),
		CreatedAt:      template.CreatedAt.String(),
		UpdatedAt:      template.UpdatedAt.String(),
		ServicePath:    template.ServicePath,
	}

	// Handle access type
	switch template.AccessType {
	case model.AccessTypeDirect:
		resp.AccessType = instance.AccessType_DIRECT
	case model.AccessTypeProxy:
		resp.AccessType = instance.AccessType_PROXY
	case model.AccessTypeHosting:
		resp.AccessType = instance.AccessType_HOSTING
	default:
		resp.AccessType = instance.AccessType_PROXY
	}

	// Handle MCP protocol
	switch template.McpProtocol {
	case model.McpProtocolSSE:
		resp.McpProtocol = instance.McpProtocol_SSE
	case model.McpProtocolStreamableHttp:
		resp.McpProtocol = instance.McpProtocol_STEAMABLE_HTTP
	case model.McpProtocolStdio:
		resp.McpProtocol = instance.McpProtocol_STDIO
	default:
		resp.McpProtocol = instance.McpProtocol_SSE
	}

	// Handle environment variables
	if len(template.EnvironmentVariables) > 0 {
		envVars := make(map[string]string)
		if err := json.Unmarshal(template.EnvironmentVariables, &envVars); err != nil {
			logger.Error("failed to unmarshal environment variables", zap.Error(err))
		} else {
			resp.EnvironmentVariables = envVars
		}
	}

	// Handle volume mount configuration
	if len(template.VolumeMounts) > 0 {
		volumeMounts := make([]*instance.VolumeMount, 0)
		if err := json.Unmarshal(template.VolumeMounts, &volumeMounts); err != nil {
			logger.Error("failed to unmarshal volume mounts", zap.Error(err))
		} else {
			resp.VolumeMounts = volumeMounts
		}
	}

	// Handle tokens
	if len(template.Tokens) > 0 {
		tokens := make([]*instance.McpToken, 0, len(template.Tokens))
		if err := json.Unmarshal(template.Tokens, &tokens); err != nil {
			logger.Error("failed to unmarshal tokens", zap.Error(err))
		} else {
			resp.Tokens = tokens
		}
	}

	return resp, nil
}

// TemplateEdit edits an existing template
func (s *TemplateService) TemplateEdit(ctx context.Context, req *instance.TemplateEditRequest) (*instance.TemplateEditResp, error) {
	if req.TemplateId == 0 {
		return nil, fmt.Errorf("template ID is required")
	}

	// Query existing template
	template, err := s.templateData.GetTemplateByID(ctx, uint(req.TemplateId))
	if err != nil {
		logger.Error("failed to get template", zap.Error(err), zap.Int32("templateId", req.TemplateId))
		return nil, fmt.Errorf("failed to get template: %v", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	// Update template fields
	template.Name = req.Name
	template.Port = req.Port
	template.InitScript = req.InitScript
	template.Command = req.Command
	template.StartupTimeout = req.StartupTimeout
	template.RunningTimeout = req.RunningTimeout
	template.EnvironmentID = req.EnvironmentId
	template.PackageID = req.PackageId
	template.ImgAddress = req.ImgAddress
	template.McpServerID = req.McpServerId
	template.Notes = req.Notes
	template.IconPath = req.IconPath

	// Handle access type
	switch req.AccessType {
	case instance.AccessType_DIRECT:
		template.AccessType = model.AccessTypeDirect
	case instance.AccessType_PROXY:
		template.AccessType = model.AccessTypeProxy
	case instance.AccessType_HOSTING:
		template.AccessType = model.AccessTypeHosting
	default:
		template.AccessType = model.AccessTypeProxy // Default proxy mode
	}

	// Handle MCP protocol
	switch req.McpProtocol {
	case instance.McpProtocol_SSE:
		template.McpProtocol = model.McpProtocolSSE
	case instance.McpProtocol_STEAMABLE_HTTP:
		template.McpProtocol = model.McpProtocolStreamableHttp
	case instance.McpProtocol_STDIO:
		template.McpProtocol = model.McpProtocolStdio
	default:
		template.McpProtocol = model.McpProtocolSSE // Default SSE protocol
	}

	// Handle environment variables
	if len(req.EnvironmentVariables) > 0 {
		envBytes, err := json.Marshal(req.EnvironmentVariables)
		if err != nil {
			logger.Error("failed to marshal environment variables", zap.Error(err))
			return nil, fmt.Errorf("failed to process environment variables: %v", err)
		}
		template.EnvironmentVariables = envBytes
	}

	// Handle volume mount configuration
	if len(req.VolumeMounts) > 0 {
		volumeBytes, err := json.Marshal(req.VolumeMounts)
		if err != nil {
			logger.Error("failed to marshal volume mounts", zap.Error(err))
			return nil, fmt.Errorf("failed to process volume mounts: %v", err)
		}
		template.VolumeMounts = volumeBytes
	}

	// Handle MCP server configuration
	if req.McpServers != "" {
		template.McpServers = json.RawMessage(req.McpServers)
	}

	// Handle tokens
	if len(req.Tokens) > 0 {
		tokens := make([]model.McpToken, 0, len(req.Tokens))
		for _, token := range req.Tokens {
			tokens = append(tokens, model.McpToken{
				Token:     token.Token,
				ExpireAt:  token.ExpireAt,
				PublishAt: token.PublishAt,
				Usages:    token.Usages,
			})
		}
		tokensJSON, err := json.Marshal(tokens)
		if err != nil {
			logger.Error("failed to marshal tokens", zap.Error(err))
			return nil, fmt.Errorf("failed to process tokens: %v", err)
		}
		template.Tokens = json.RawMessage(tokensJSON)
	}

	// Update template
	if err := s.templateData.UpdateTemplate(ctx, template); err != nil {
		logger.Error("failed to update template", zap.Error(err), zap.Int32("templateId", req.TemplateId))
		return nil, fmt.Errorf("failed to update template: %v", err)
	}

	// Return response
	resp := &instance.TemplateEditResp{
		Message: "Template updated successfully",
	}

	logger.Info("template updated successfully", zap.Int32("templateId", req.TemplateId), zap.String("name", req.Name))
	return resp, nil
}

// TemplateList retrieves a list of templates
func (s *TemplateService) TemplateList(ctx context.Context, req *instance.TemplateListRequest) (*instance.TemplateListResp, error) {
	// Set default pagination parameters
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = int32(common.DefaultPageSize)
	}
	if pageSize > int32(common.MaxPageSize) {
		pageSize = int32(common.MaxPageSize)
	}

	// Build filter conditions
	filters := make(map[string]interface{})

	// Add template ID filter
	if req.TemplateId > 0 {
		filters["template_id"] = req.TemplateId
	}

	// Add name filter
	if req.Name != "" {
		filters["name"] = req.Name
	}

	// Paginated query for template list
	templates, total, err := s.templateData.GetTemplatesWithPagination(ctx, page, pageSize, filters, "id", "desc")
	if err != nil {
		logger.Error("failed to get templates", zap.Error(err))
		return nil, fmt.Errorf("failed to get templates: %v", err)
	}

	// envIds
	envIds := make([]string, 0, len(templates))
	for _, instance := range templates {
		envIds = append(envIds, fmt.Sprintf("%d", instance.EnvironmentID))
	}
	envIds = utils.RemoveDuplicates(envIds)
	envNames, err := mysql.McpEnvironmentRepo.FindNamesByIDs(ctx, envIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment names: %v", err)
	}

	// Build response
	resp := &instance.TemplateListResp{
		List:     make([]*instance.TemplateDetailResp, 0, len(templates)),
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}

	// Process each template
	for _, template := range templates {
		envName, ok := envNames[fmt.Sprintf("%d", template.EnvironmentID)]
		if !ok {
			envName = ""
		}
		templateResp := &instance.TemplateDetailResp{
			TemplateId:      int32(template.ID),
			Name:            template.Name,
			Port:            template.Port,
			InitScript:      template.InitScript,
			Command:         template.Command,
			StartupTimeout:  template.StartupTimeout,
			RunningTimeout:  template.RunningTimeout,
			EnvironmentId:   int32(template.EnvironmentID),
			PackageId:       template.PackageID,
			ImgAddress:      template.ImgAddress,
			McpServerId:     template.McpServerID,
			Notes:           template.Notes,
			IconPath:        template.IconPath,
			McpServers:      string(template.McpServers),
			CreatedAt:       template.CreatedAt.String(),
			UpdatedAt:       template.UpdatedAt.String(),
			EnvironmentName: envName,
			ServicePath:     template.ServicePath,
		}

		// Handle access type
		switch template.AccessType {
		case model.AccessTypeDirect:
			templateResp.AccessType = instance.AccessType_DIRECT
		case model.AccessTypeProxy:
			templateResp.AccessType = instance.AccessType_PROXY
		case model.AccessTypeHosting:
			templateResp.AccessType = instance.AccessType_HOSTING
		default:
			templateResp.AccessType = instance.AccessType_PROXY
		}

		// Handle MCP protocol
		switch template.McpProtocol {
		case model.McpProtocolSSE:
			templateResp.McpProtocol = instance.McpProtocol_SSE
		case model.McpProtocolStreamableHttp:
			templateResp.McpProtocol = instance.McpProtocol_STEAMABLE_HTTP
		case model.McpProtocolStdio:
			templateResp.McpProtocol = instance.McpProtocol_STDIO
		default:
			templateResp.McpProtocol = instance.McpProtocol_SSE
		}

		// Handle environment variables
		if len(template.EnvironmentVariables) > 0 {
			envVars := make(map[string]string)
			if err := json.Unmarshal(template.EnvironmentVariables, &envVars); err != nil {
				logger.Error("failed to unmarshal environment variables", zap.Error(err))
			} else {
				templateResp.EnvironmentVariables = envVars
			}
		}

		// Handle volume mount configuration
		if len(template.VolumeMounts) > 0 {
			volumeMounts := make([]*instance.VolumeMount, 0)
			if err := json.Unmarshal(template.VolumeMounts, &volumeMounts); err != nil {
				logger.Error("failed to unmarshal volume mounts", zap.Error(err))
			} else {
				templateResp.VolumeMounts = volumeMounts
			}
		}

		// Handle tokens
		if len(template.Tokens) > 0 {
			tokens := make([]*instance.McpToken, 0, len(template.Tokens))
			if err := json.Unmarshal(template.Tokens, &tokens); err != nil {
				logger.Error("failed to unmarshal tokens", zap.Error(err))
			} else {
				templateResp.Tokens = tokens
			}
		}

		resp.List = append(resp.List, templateResp)
	}

	return resp, nil
}

// TemplateListWithPagination retrieves a paginated list of templates
func (s *TemplateService) TemplateListWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) ([]*instance.TemplateDetailResp, int64, error) {
	// Paginated query for template list
	templates, total, err := s.templateData.GetTemplatesWithPagination(ctx, page, pageSize, filters, sortBy, sortOrder)
	if err != nil {
		logger.Error("failed to get templates with pagination", zap.Error(err), zap.Int32("page", page), zap.Int32("pageSize", pageSize))
		return nil, 0, fmt.Errorf("failed to get templates: %v", err)
	}

	// Build response
	templateResps := make([]*instance.TemplateDetailResp, 0, len(templates))

	// Process each template
	for _, template := range templates {
		templateResp := &instance.TemplateDetailResp{
			TemplateId:     int32(template.ID),
			Name:           template.Name,
			Port:           template.Port,
			InitScript:     template.InitScript,
			Command:        template.Command,
			StartupTimeout: template.StartupTimeout,
			RunningTimeout: template.RunningTimeout,
			EnvironmentId:  int32(template.EnvironmentID),
			PackageId:      template.PackageID,
			ImgAddress:     template.ImgAddress,
			McpServerId:    template.McpServerID,
			Notes:          template.Notes,
			IconPath:       template.IconPath,
			McpServers:     string(template.McpServers),
		}

		// Handle access type
		switch template.AccessType {
		case model.AccessTypeDirect:
			templateResp.AccessType = instance.AccessType_DIRECT
		case model.AccessTypeProxy:
			templateResp.AccessType = instance.AccessType_PROXY
		case model.AccessTypeHosting:
			templateResp.AccessType = instance.AccessType_HOSTING
		default:
			templateResp.AccessType = instance.AccessType_PROXY
		}

		// Handle MCP protocol
		switch template.McpProtocol {
		case model.McpProtocolSSE:
			templateResp.McpProtocol = instance.McpProtocol_SSE
		case model.McpProtocolStreamableHttp:
			templateResp.McpProtocol = instance.McpProtocol_STEAMABLE_HTTP
		case model.McpProtocolStdio:
			templateResp.McpProtocol = instance.McpProtocol_STDIO
		default:
			templateResp.McpProtocol = instance.McpProtocol_SSE
		}

		// Handle environment variables
		if len(template.EnvironmentVariables) > 0 {
			envVars := make(map[string]string)
			if err := json.Unmarshal(template.EnvironmentVariables, &envVars); err != nil {
				logger.Error("failed to unmarshal environment variables", zap.Error(err))
			} else {
				templateResp.EnvironmentVariables = envVars
			}
		}

		// Handle volume mount configuration
		if len(template.VolumeMounts) > 0 {
			volumeMounts := make([]*instance.VolumeMount, 0)
			if err := json.Unmarshal(template.VolumeMounts, &volumeMounts); err != nil {
				logger.Error("failed to unmarshal volume mounts", zap.Error(err))
			} else {
				templateResp.VolumeMounts = volumeMounts
			}
		}

		// Handle tokens
		if len(template.Tokens) > 0 {
			tokens := make([]*instance.McpToken, 0, len(template.Tokens))
			if err := json.Unmarshal(template.Tokens, &tokens); err != nil {
				logger.Error("failed to unmarshal tokens", zap.Error(err))
			} else {
				templateResp.Tokens = tokens
			}
		}

		templateResps = append(templateResps, templateResp)
	}

	return templateResps, total, nil
}

// TemplateDelete deletes a template
func (s *TemplateService) TemplateDelete(ctx context.Context, req *instance.TemplateDeleteRequest) (*instance.TemplateDeleteResp, error) {
	if req.TemplateId == 0 {
		return nil, fmt.Errorf("template ID is required")
	}

	// Query template
	template, err := s.templateData.GetTemplateByID(ctx, uint(req.TemplateId))
	if err != nil {
		logger.Error("failed to get template", zap.Error(err), zap.Int32("templateId", req.TemplateId))
		return nil, fmt.Errorf("failed to get template: %v", err)
	}

	// Delete template
	if err := s.templateData.DeleteTemplate(ctx, template.ID); err != nil {
		logger.Error("failed to delete template", zap.Error(err), zap.Int32("templateId", req.TemplateId))
		return nil, fmt.Errorf("failed to delete template: %v", err)
	}

	// Return response
	resp := &instance.TemplateDeleteResp{}

	logger.Info("template deleted successfully", zap.Int32("templateId", req.TemplateId))
	return resp, nil
}

// HTTP Handler methods

// TemplateCreateHandler creates template HTTP handler function
func (s *TemplateService) TemplateCreateHandler(c *gin.Context) {
	var req instance.TemplateCreateRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
		return
	}

	// Call create template handler function
	result, err := s.TemplateCreate(c, &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create template: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// TemplateListWithPaginationHandler paginated template list HTTP handler function
func (s *TemplateService) TemplateListWithPaginationHandler(c *gin.Context) {
	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// Convert pagination parameters
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 32)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get sorting parameters
	sortBy := c.DefaultQuery("sortBy", "id")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	// Build filter conditions
	filters := make(map[string]interface{})

	// Handle environment ID filter
	if envIdStr := c.Query("environmentId"); envIdStr != "" {
		if envId, parseErr := strconv.ParseInt(envIdStr, 10, 32); parseErr == nil {
			filters["environment_id"] = envId
		}
	}

	// Handle access type filter
	if accessType := c.Query("accessType"); accessType != "" {
		filters["access_type"] = accessType
	}

	// Handle source type filter
	if sourceType := c.Query("sourceType"); sourceType != "" {
		filters["source_type"] = sourceType
	}

	// Handle name fuzzy search
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	// Call paginated template list handler function
	result, total, err := s.TemplateListWithPagination(c, int32(page), int32(pageSize), filters, sortBy, sortOrder)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get paginated template list: %s", err.Error()))
		return
	}

	// Build pagination response
	response := map[string]interface{}{
		"list":      result,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
		"totalPage": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	// Return success response
	common.GinSuccess(c, response)
}

// TemplateDetailHandler get template details HTTP handler function
func (s *TemplateService) TemplateDetailHandler(c *gin.Context) {
	var req instance.TemplateDetailRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	if req.TemplateId == 0 {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: templateId")
		return
	}

	// Call get template details handler function
	result, err := s.TemplateDetail(c, &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get template details: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// TemplateEditHandler edit template HTTP handler function
func (s *TemplateService) TemplateEditHandler(c *gin.Context) {
	var req instance.TemplateEditRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to bind request body: %s", err.Error()))
		return
	}

	// Validate required fields
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
		return
	}

	// Call edit template handler function
	result, err := s.TemplateEdit(c, &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to edit template: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// TemplateListHandler get template list HTTP handler function
func (s *TemplateService) TemplateListHandler(c *gin.Context) {
	var req instance.TemplateListRequest

	// Bind request body
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Call get template list handler function
	result, err := s.TemplateList(c, &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get template list: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// TemplateDeleteHandler delete template HTTP handler function
func (s *TemplateService) TemplateDeleteHandler(c *gin.Context) {
	var req instance.TemplateDeleteRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to bind request body: %s", err.Error()))
		return
	}
	if req.TemplateId == 0 {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: templateId")
		return
	}

	// Call delete template handler function
	result, err := s.TemplateDelete(c, &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to delete template: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}
