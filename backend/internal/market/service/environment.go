package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kymo-mcp/mcpcan/api/market/mcp_environment"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// EnvironmentService provides environment management functionality
type EnvironmentService struct {
	ctx context.Context
}

// NewEnvironmentService creates a new EnvironmentService instance
func NewEnvironmentService(ctx context.Context) *EnvironmentService {
	return &EnvironmentService{
		ctx: ctx,
	}
}

// modelToMcpEnvironmentInfo converts model to MCP environment info
func modelToMcpEnvironmentInfo(env *model.McpEnvironment) *mcp_environment.McpEnvironmentInfo {
	return &mcp_environment.McpEnvironmentInfo{
		Id:          int32(env.ID),
		Name:        env.Name,
		Environment: string(env.Environment),
		Config:      env.Config,
		Namespace:   env.Namespace,
		CreatedAt:   env.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   env.UpdatedAt.Format(time.RFC3339),
	}
}

// modelToEnvironmentResponse converts model to environment response
func modelToEnvironmentResponse(env *model.McpEnvironment) *mcp_environment.EnvironmentResponse {
	var envType mcp_environment.McpEnvironmentType
	switch env.Environment {
	case model.McpEnvironmentKubernetes:
		envType = mcp_environment.McpEnvironmentType_Kubernetes
	case model.McpEnvironmentDocker:
		envType = mcp_environment.McpEnvironmentType_Docker
	default:
		envType = mcp_environment.McpEnvironmentType_Kubernetes
	}

	return &mcp_environment.EnvironmentResponse{
		Id:          int32(env.ID),
		Name:        env.Name,
		Environment: envType,
		Config:      env.Config,
		Namespace:   env.Namespace,
		CreatedAt:   env.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   env.UpdatedAt.Format(time.RFC3339),
	}
}

// CreateEnvironmentHandler handles environment creation requests
func (s *EnvironmentService) CreateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.CreateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Use EnvironmentService to handle request
	result, err := s.CreateEnvironment(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// CreateEnvironment creates a new environment
func (s *EnvironmentService) CreateEnvironment(req *mcp_environment.CreateEnvironmentRequest) (*mcp_environment.EnvironmentResponse, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	// Validate environment type
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("unsupported environment type, only kubernetes or docker are supported")
	}

	// Check if environment name already exists
	existingEnv, err := biz.GEnvironmentBiz.GetEnvironmentByName(s.ctx, req.Name)
	if err == nil && existingEnv != nil {
		return nil, fmt.Errorf("environment name already exists")
	}

	// Create environment object
	environment := &model.McpEnvironment{
		Name:        req.Name,
		Environment: envType,
		Config:      req.Config,
		Namespace:   req.Namespace,
		CreatorID:   "",
	}

	// Validate and prepare for creation
	if validationErr := environment.ValidateForCreate(); validationErr != nil {
		return nil, fmt.Errorf("environment data validation failed: %s", err.Error())
	}
	environment.PrepareForCreate()

	// Create environment
	err = biz.GEnvironmentBiz.CreateEnvironment(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %s", err.Error())
	}

	// Build response
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// CreateEnvironmentHandler create environment interface (package-level function for backward compatibility)
func CreateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.CreateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment name cannot be empty")
		return
	}

	// Validate environment type
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		common.GinError(c, i18nresp.CodeInternalError, "unsupported environment type, only kubernetes or docker are supported")
		return
	}

	// Check if environment name already exists
	existingEnv, err := biz.GEnvironmentBiz.GetEnvironmentByName(c.Request.Context(), req.Name)
	if err == nil && existingEnv != nil {
		common.GinError(c, i18nresp.CodeInternalError, "environment name already exists")
		return
	}

	// Create environment object
	environment := &model.McpEnvironment{
		Name:        req.Name,
		Environment: envType,
		Config:      req.Config,
		Namespace:   req.Namespace,
		CreatorID:   "",
	}

	// Validate and prepare for creation
	if validationErr := environment.ValidateForCreate(); validationErr != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("environment data validation failed: %s", err.Error()))
		return
	}
	environment.PrepareForCreate()

	// Create environment
	err = biz.GEnvironmentBiz.CreateEnvironment(c.Request.Context(), environment)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create environment: %s", err.Error()))
		return
	}

	// Build response
	response := modelToEnvironmentResponse(environment)

	common.GinSuccess(c, response)
}

// UpdateEnvironmentHandler handles environment update requests
func (s *EnvironmentService) UpdateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.UpdateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid environment ID")
		return
	}
	req.Id = int32(id)

	// Use EnvironmentService to handle request
	result, err := s.UpdateEnvironment(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// UpdateEnvironment updates an existing environment
func (s *EnvironmentService) UpdateEnvironment(req *mcp_environment.UpdateEnvironmentRequest) (*mcp_environment.EnvironmentResponse, error) {
	// Validate environment type
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("unsupported environment type, only kubernetes or docker are supported")
	}

	// Update environment

	// First get existing environment
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, uint(req.Id))
	if err != nil {
		return nil, fmt.Errorf("failed to query environment: %s", err.Error())
	}

	// Update fields
	environment.Name = req.Name
	environment.Environment = envType
	environment.Config = req.Config
	environment.Namespace = req.Namespace

	// Validate and prepare for update
	if validationErr := environment.ValidateForUpdate(); validationErr != nil {
		return nil, fmt.Errorf("environment data validation failed: %s", validationErr.Error())
	}
	environment.PrepareForUpdate()

	// Execute update
	err = biz.GEnvironmentBiz.UpdateEnvironment(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment: %s", err.Error())
	}

	// Build response
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// UpdateEnvironmentHandler update environment interface (package-level function for backward compatibility)
func UpdateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.UpdateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid environment ID")
		return
	}
	req.Id = int32(id)

	// Validate environment type
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		common.GinError(c, i18nresp.CodeInternalError, "unsupported environment type, only kubernetes or docker are supported")
		return
	}

	// Update environment

	// First get existing environment
	environment, err := biz.GEnvironmentBiz.GetEnvironment(c.Request.Context(), uint(req.Id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to query environment: %s", err.Error()))
		return
	}

	// Update fields
	environment.Name = req.Name
	environment.Environment = envType
	environment.Config = req.Config
	environment.Namespace = req.Namespace

	// Validate and prepare for update
	if validationErr := environment.ValidateForUpdate(); validationErr != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("environment data validation failed: %s", validationErr.Error()))
		return
	}
	environment.PrepareForUpdate()

	// Execute update
	err = biz.GEnvironmentBiz.UpdateEnvironment(c.Request.Context(), environment)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to update environment: %s", err.Error()))
		return
	}

	// Build response
	response := modelToEnvironmentResponse(environment)

	common.GinSuccess(c, response)
}

// GetEnvironmentHandler get environment interface Handler
func (s *EnvironmentService) GetEnvironmentHandler(c *gin.Context) {
	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid environment ID")
		return
	}

	// Use EnvironmentService to handle request
	result, err := s.GetEnvironment(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// GetEnvironment get environment business logic
func (s *EnvironmentService) GetEnvironment(id uint) (*mcp_environment.EnvironmentResponse, error) {
	// Get environment
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment: %s", err.Error())
	}

	// Build response
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// DeleteEnvironmentHandler delete environment interface Handler
func (s *EnvironmentService) DeleteEnvironmentHandler(c *gin.Context) {
	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid environment ID")
		return
	}

	// Use EnvironmentService to handle request
	err = s.DeleteEnvironment(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, gin.H{"message": "environment deleted successfully"})
}

// DeleteEnvironment delete environment business logic
func (s *EnvironmentService) DeleteEnvironment(id uint) error {
	// Delete environment
	err := biz.GEnvironmentBiz.DeleteEnvironment(s.ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %s", err.Error())
	}

	return nil
}

// ListEnvironmentsHandler environment list interface Handler
func (s *EnvironmentService) ListEnvironmentsHandler(c *gin.Context) {
	var req mcp_environment.ListEnvironmentsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Use EnvironmentService to handle request
	result, err := s.ListEnvironments(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListEnvironments environment list business logic
func (s *EnvironmentService) ListEnvironments(req *mcp_environment.ListEnvironmentsRequest) (*mcp_environment.ListEnvironmentsResponse, error) {
	// Set default pagination parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit maximum page size
	}

	var environments []*model.McpEnvironment
	var err error

	// Query based on filter conditions
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(s.ctx, envType)
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(s.ctx, envType)
	default:
		// Query all environments
		environments, err = biz.GEnvironmentBiz.ListEnvironments(s.ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query environment list: %s", err.Error())
	}

	// Calculate pagination
	total := int64(len(environments))
	start := (int(req.Page) - 1) * int(req.PageSize)
	end := start + int(req.PageSize)

	if start >= len(environments) {
		environments = []*model.McpEnvironment{}
	} else {
		if end > len(environments) {
			end = len(environments)
		}
		environments = environments[start:end]
	}

	// Build response list
	var responseList []*mcp_environment.McpEnvironmentInfo
	for _, env := range environments {
		responseList = append(responseList, modelToMcpEnvironmentInfo(env))
	}

	response := &mcp_environment.ListEnvironmentsResponse{
		List:     responseList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return response, nil
}

// ListEnvironmentsHandler environment list interface (package-level function for backward compatibility)
func ListEnvironmentsHandler(c *gin.Context) {
	var req mcp_environment.ListEnvironmentsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// Set default pagination parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit maximum page size
	}

	var environments []*model.McpEnvironment
	var err error

	// Query based on filter conditions
	// Note: Since there's no Unspecified value in proto, we need to determine if filtering is needed through other means
	// Here we check if environment type is explicitly specified in the request
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(c.Request.Context(), envType)
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(c.Request.Context(), envType)
	default:
		// Query all environments
		environments, err = biz.GEnvironmentBiz.ListEnvironments(c.Request.Context())
	}

	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Calculate pagination
	total := int64(len(environments))
	start := (int(req.Page) - 1) * int(req.PageSize)
	end := start + int(req.PageSize)

	if start >= len(environments) {
		environments = []*model.McpEnvironment{}
	} else {
		if end > len(environments) {
			end = len(environments)
		}
		environments = environments[start:end]
	}

	// Build response list
	var responseList []*mcp_environment.McpEnvironmentInfo
	for _, env := range environments {
		responseList = append(responseList, modelToMcpEnvironmentInfo(env))
	}

	response := &mcp_environment.ListEnvironmentsResponse{
		List:     responseList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	common.GinSuccess(c, response)
}

// TestConnectivityHandler connectivity test interface Handler
func (s *EnvironmentService) TestConnectivityHandler(c *gin.Context) {
	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid environment ID")
		return
	}

	// Use EnvironmentService to handle request
	result, err := s.TestConnectivity(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// TestConnectivity connectivity test business logic
func (s *EnvironmentService) TestConnectivity(id uint) (*mcp_environment.TestConnectivityResponse, error) {
	// Get environment information
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment: %s", err.Error())
	}
	if environment == nil {
		return nil, fmt.Errorf("environment does not exist")
	}

	// Execute connectivity test
	result, err := testEnvironmentConnectivity(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("connectivity test failed: %s", err.Error())
	}

	return result, nil
}

// testEnvironmentConnectivity execute environment connectivity test
func testEnvironmentConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// Use data layer connectivity test method
	return biz.GEnvironmentBiz.TestEnvironmentConnectivity(ctx, environment)
}

// ListAllEnvironmentsHandler get all environment list (including deleted ones)
func ListAllEnvironmentsHandler(c *gin.Context) {
	environments, err := biz.GEnvironmentBiz.ListAllEnvironments(c.Request.Context())
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	var environmentInfos []*mcp_environment.McpEnvironmentInfo
	for _, env := range environments {
		environmentInfos = append(environmentInfos, modelToMcpEnvironmentInfo(env))
	}

	response := &mcp_environment.ListEnvironmentsResponse{
		List:     environmentInfos,
		Total:    int64(len(environmentInfos)),
		Page:     1,
		PageSize: int32(len(environmentInfos)),
	}

	common.GinSuccess(c, response)
}

// ListNamespacesHandler get namespace list Handler
func (s *EnvironmentService) ListNamespacesHandler(c *gin.Context) {
	// Bind request parameters
	var req mcp_environment.ListNamespacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("parameter binding failed: %s", err.Error()))
		return
	}

	// Use EnvironmentService to handle request
	result, err := s.ListNamespaces(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListNamespaces get namespace list business logic
func (s *EnvironmentService) ListNamespaces(req *mcp_environment.ListNamespacesRequest) (*mcp_environment.ListNamespacesResponse, error) {
	if req.Config == "" {
		return nil, fmt.Errorf("config parameter cannot be empty")
	}

	// Parse environment type
	var environmentType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		environmentType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		environmentType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("unsupported environment type")
	}

	// Call business logic
	namespaces, err := biz.GEnvironmentBiz.ListNamespaces(s.ctx, req.Config, environmentType)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &mcp_environment.ListNamespacesResponse{
		List: namespaces,
	}

	return response, nil
}
