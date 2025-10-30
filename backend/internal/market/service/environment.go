package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"qm-mcp-server/api/market/mcp_environment"
	"qm-mcp-server/internal/market/biz"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	i18nresp "qm-mcp-server/pkg/i18n"
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

	// 使用 EnvironmentService 处理请求
	result, err := s.CreateEnvironment(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// CreateEnvironment creates a new environment
func (s *EnvironmentService) CreateEnvironment(req *mcp_environment.CreateEnvironmentRequest) (*mcp_environment.EnvironmentResponse, error) {
	// 验证必填字段
	if req.Name == "" {
		return nil, fmt.Errorf("环境名称不能为空")
	}

	// 验证环境类型
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("不支持的环境类型，仅支持 kubernetes 或 docker")
	}

	// 检查环境名称是否已存在
	existingEnv, err := biz.GEnvironmentBiz.GetEnvironmentByName(s.ctx, req.Name)
	if err == nil && existingEnv != nil {
		return nil, fmt.Errorf("环境名称已存在")
	}

	// 创建环境对象
	environment := &model.McpEnvironment{
		Name:        req.Name,
		Environment: envType,
		Config:      req.Config,
		Namespace:   req.Namespace,
		CreatorID:   "",
	}

	// 验证和准备创建
	if validationErr := environment.ValidateForCreate(); validationErr != nil {
		return nil, fmt.Errorf("环境数据验证失败: %s", err.Error())
	}
	environment.PrepareForCreate()

	// 创建环境
	err = biz.GEnvironmentBiz.CreateEnvironment(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("创建环境失败: %s", err.Error())
	}

	// 构建响应
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// CreateEnvironmentHandler 创建环境接口（包级函数，保持向后兼容）
func CreateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.CreateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 验证必填字段
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境名称不能为空")
		return
	}

	// 验证环境类型
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		common.GinError(c, i18nresp.CodeInternalError, "不支持的环境类型，仅支持 kubernetes 或 docker")
		return
	}

	// 检查环境名称是否已存在
	existingEnv, err := biz.GEnvironmentBiz.GetEnvironmentByName(c.Request.Context(), req.Name)
	if err == nil && existingEnv != nil {
		common.GinError(c, i18nresp.CodeInternalError, "环境名称已存在")
		return
	}

	// 创建环境对象
	environment := &model.McpEnvironment{
		Name:        req.Name,
		Environment: envType,
		Config:      req.Config,
		Namespace:   req.Namespace,
		CreatorID:   "",
	}

	// 验证和准备创建
	if validationErr := environment.ValidateForCreate(); validationErr != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("环境数据验证失败: %s", err.Error()))
		return
	}
	environment.PrepareForCreate()

	// 创建环境
	err = biz.GEnvironmentBiz.CreateEnvironment(c.Request.Context(), environment)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("创建环境失败: %s", err.Error()))
		return
	}

	// 构建响应
	response := modelToEnvironmentResponse(environment)

	common.GinSuccess(c, response)
}

// UpdateEnvironmentHandler handles environment update requests
func (s *EnvironmentService) UpdateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.UpdateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 从URL路径参数获取ID
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的环境ID")
		return
	}
	req.Id = int32(id)

	// 使用 EnvironmentService 处理请求
	result, err := s.UpdateEnvironment(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// UpdateEnvironment updates an existing environment
func (s *EnvironmentService) UpdateEnvironment(req *mcp_environment.UpdateEnvironmentRequest) (*mcp_environment.EnvironmentResponse, error) {
	// 验证环境类型
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("不支持的环境类型，仅支持 kubernetes 或 docker")
	}

	// 更新环境

	// 先获取现有环境
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, uint(req.Id))
	if err != nil {
		return nil, fmt.Errorf("查询环境失败: %s", err.Error())
	}

	// 更新字段
	environment.Name = req.Name
	environment.Environment = envType
	environment.Config = req.Config
	environment.Namespace = req.Namespace

	// 验证和准备更新
	if validationErr := environment.ValidateForUpdate(); validationErr != nil {
		return nil, fmt.Errorf("环境数据验证失败: %s", validationErr.Error())
	}
	environment.PrepareForUpdate()

	// 执行更新
	err = biz.GEnvironmentBiz.UpdateEnvironment(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("更新环境失败: %s", err.Error())
	}

	// 构建响应
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// UpdateEnvironmentHandler 更新环境接口（包级函数，保持向后兼容）
func UpdateEnvironmentHandler(c *gin.Context) {
	var req mcp_environment.UpdateEnvironmentRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 从URL路径参数获取ID
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的环境ID")
		return
	}
	req.Id = int32(id)

	// 验证环境类型
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
	default:
		common.GinError(c, i18nresp.CodeInternalError, "不支持的环境类型，仅支持 kubernetes 或 docker")
		return
	}

	// 更新环境

	// 先获取现有环境
	environment, err := biz.GEnvironmentBiz.GetEnvironment(c.Request.Context(), uint(req.Id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("查询环境失败: %s", err.Error()))
		return
	}

	// 更新字段
	environment.Name = req.Name
	environment.Environment = envType
	environment.Config = req.Config
	environment.Namespace = req.Namespace

	// 验证和准备更新
	if validationErr := environment.ValidateForUpdate(); validationErr != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("环境数据验证失败: %s", validationErr.Error()))
		return
	}
	environment.PrepareForUpdate()

	// 执行更新
	err = biz.GEnvironmentBiz.UpdateEnvironment(c.Request.Context(), environment)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("更新环境失败: %s", err.Error()))
		return
	}

	// 构建响应
	response := modelToEnvironmentResponse(environment)

	common.GinSuccess(c, response)
}

// GetEnvironmentHandler 获取环境接口Handler
func (s *EnvironmentService) GetEnvironmentHandler(c *gin.Context) {
	// 从URL路径参数获取ID
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的环境ID")
		return
	}

	// 使用 EnvironmentService 处理请求
	result, err := s.GetEnvironment(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// GetEnvironment 获取环境业务逻辑
func (s *EnvironmentService) GetEnvironment(id uint) (*mcp_environment.EnvironmentResponse, error) {
	// 获取环境
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询环境失败: %s", err.Error())
	}

	// 构建响应
	response := modelToEnvironmentResponse(environment)

	return response, nil
}

// DeleteEnvironmentHandler 删除环境接口Handler
func (s *EnvironmentService) DeleteEnvironmentHandler(c *gin.Context) {
	// 从URL路径参数获取ID
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的环境ID")
		return
	}

	// 使用 EnvironmentService 处理请求
	err = s.DeleteEnvironment(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, gin.H{"message": "环境删除成功"})
}

// DeleteEnvironment 删除环境业务逻辑
func (s *EnvironmentService) DeleteEnvironment(id uint) error {
	// 删除环境
	err := biz.GEnvironmentBiz.DeleteEnvironment(s.ctx, id)
	if err != nil {
		return fmt.Errorf("删除环境失败: %s", err.Error())
	}

	return nil
}

// ListEnvironmentsHandler 环境列表接口Handler
func (s *EnvironmentService) ListEnvironmentsHandler(c *gin.Context) {
	var req mcp_environment.ListEnvironmentsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 使用 EnvironmentService 处理请求
	result, err := s.ListEnvironments(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListEnvironments 环境列表业务逻辑
func (s *EnvironmentService) ListEnvironments(req *mcp_environment.ListEnvironmentsRequest) (*mcp_environment.ListEnvironmentsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页面大小
	}

	var environments []*model.McpEnvironment
	var err error

	// 根据过滤条件查询
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(s.ctx, envType)
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(s.ctx, envType)
	default:
		// 查询所有环境
		environments, err = biz.GEnvironmentBiz.ListEnvironments(s.ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("查询环境列表失败: %s", err.Error())
	}

	// 计算分页
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

	// 构建响应列表
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

// ListEnvironmentsHandler 环境列表接口（包级函数，保持向后兼容）
func ListEnvironmentsHandler(c *gin.Context) {
	var req mcp_environment.ListEnvironmentsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页面大小
	}

	var environments []*model.McpEnvironment
	var err error

	// 根据过滤条件查询
	// 注意：由于proto中没有Unspecified值，我们需要通过其他方式判断是否需要过滤
	// 这里我们检查请求中是否明确指定了环境类型
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(c.Request.Context(), envType)
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
		environments, err = biz.GEnvironmentBiz.ListEnvironmentsByType(c.Request.Context(), envType)
	default:
		// 查询所有环境
		environments, err = biz.GEnvironmentBiz.ListEnvironments(c.Request.Context())
	}

	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 计算分页
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

	// 构建响应列表
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

// TestConnectivityHandler 连通性测试接口Handler
func (s *EnvironmentService) TestConnectivityHandler(c *gin.Context) {
	// 从URL路径参数获取ID
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "环境ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的环境ID")
		return
	}

	// 使用 EnvironmentService 处理请求
	result, err := s.TestConnectivity(uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// TestConnectivity 连通性测试业务逻辑
func (s *EnvironmentService) TestConnectivity(id uint) (*mcp_environment.TestConnectivityResponse, error) {
	// 获取环境信息
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询环境失败: %s", err.Error())
	}
	if environment == nil {
		return nil, fmt.Errorf("环境不存在")
	}

	// 执行连通性测试
	result, err := testEnvironmentConnectivity(s.ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("连通性测试失败: %s", err.Error())
	}

	return result, nil
}

// testEnvironmentConnectivity 执行环境连通性测试
func testEnvironmentConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// 使用数据层的连通性测试方法
	return biz.GEnvironmentBiz.TestEnvironmentConnectivity(ctx, environment)
}

// ListAllEnvironmentsHandler 获取所有环境列表（包括已删除）
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

// ListNamespacesHandler 获取命名空间列表Handler
func (s *EnvironmentService) ListNamespacesHandler(c *gin.Context) {
	// 绑定请求参数
	var req mcp_environment.ListNamespacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("参数绑定失败: %s", err.Error()))
		return
	}

	// 使用 EnvironmentService 处理请求
	result, err := s.ListNamespaces(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListNamespaces 获取命名空间列表业务逻辑
func (s *EnvironmentService) ListNamespaces(req *mcp_environment.ListNamespacesRequest) (*mcp_environment.ListNamespacesResponse, error) {
	if req.Config == "" {
		return nil, fmt.Errorf("config参数不能为空")
	}

	// 解析环境类型
	var environmentType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		environmentType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		environmentType = model.McpEnvironmentDocker
	default:
		return nil, fmt.Errorf("不支持的环境类型")
	}

	// 调用业务逻辑
	namespaces, err := biz.GEnvironmentBiz.ListNamespaces(s.ctx, req.Config, environmentType)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &mcp_environment.ListNamespacesResponse{
		List: namespaces,
	}

	return response, nil
}
