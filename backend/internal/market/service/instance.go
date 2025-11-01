package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

// InstanceService struct for instance service
type InstanceService struct {
	ctx context.Context
}

// ContainerInstanceUpdateParams parameters for container instance update
type ContainerInstanceUpdateParams struct {
	Instance                   *model.McpInstance
	ContainerName              string
	ServiceName                string
	ContainerStatus            model.ContainerStatus
	ContainerLastMessage       string
	ContainerCreateOptions     []byte
	ContainerInitTimeoutStopAt int64
	ContainerRunTimeoutStopAt  int64
	Status                     model.InstanceStatus
	ContainerIsReady           bool
}

// NewInstanceService creates a new instance service
func NewInstanceService(ctx context.Context) *InstanceService {
	return &InstanceService{
		ctx: ctx,
	}
}

// CreateHandler creates instance HTTP handler function
func (s *InstanceService) CreateHandler(c *gin.Context) {
	var req instancepb.CreateRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
		return
	}
	// Call write instance handler function
	result, err := s.create(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to write instance: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// DetailHandler gets instance details HTTP handler function
func (s *InstanceService) DetailHandler(c *gin.Context) {
	var req instancepb.DetailRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Call get instance details handler function
	result, err := s.detail(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get instance details: %s", err.Error()))
		return
	}

	// Return success response
	common.GinSuccess(c, result)
}

// EditHandler instance edit
func (s *InstanceService) EditHandler(c *gin.Context) {
	var req instancepb.EditRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Get original instance information
	oriInstance, err := biz.GInstanceBiz.GetInstance(req.InstanceId)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get instance information: %s", err.Error()))
		return
	}
	if oriInstance == nil {
		common.GinError(c, i18nresp.CodeInternalError, "instance does not exist")
		return
	}

	var resp *instancepb.EditResp
	switch oriInstance.AccessType {
	case model.AccessTypeDirect:
		// validate instance name
		if len(req.Name) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
			return
		}
		// validate instance mcpServers
		if len(req.McpServers) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: mcpServers")
			return
		}
		resp, err = biz.GInstanceBiz.UpdateInstanceForDirect(c.Request.Context(), &req, oriInstance)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to edit instance: %s", err.Error()))
			return
		}
	case model.AccessTypeProxy:
		// validate instance name
		if len(req.Name) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
			return
		}
		// validate instance mcpServers
		if len(req.McpServers) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: mcpServers")
			return
		}
		resp, err = biz.GInstanceBiz.UpdateInstanceForProxy(c.Request.Context(), &req, oriInstance)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to edit instance: %s", err.Error()))
			return
		}
	case model.AccessTypeHosting:
		// validate instance name
		if len(req.Name) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: name")
			return
		}
		// validate instance port
		if req.Port <= 0 {
			common.GinError(c, i18nresp.CodeInternalError, "missing required field: port")
			return
		}
		resp, err = biz.GInstanceBiz.UpdateInstanceForHosting(c.Request.Context(), &req, oriInstance)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to edit instance: %s", err.Error()))
			return
		}
	default:
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("unknown access type: %s", oriInstance.AccessType))
		return
	}

	common.GinSuccess(c, resp)
}

// ListHandler instance list
func (s *InstanceService) ListHandler(c *gin.Context) {
	var req instancepb.ListRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Use InstanceService to handle request
	result, err := s.list(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get instance list: %s", err.Error()))
		return
	}

	common.GinSuccess(c, result)
}

// DisabledHandler disable instance handler
func (s *InstanceService) DisabledHandler(c *gin.Context) {
	var req instancepb.DisabledRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Use InstanceService to handle request
	result, err := s.disable(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// RestartHandler restart instance handler
func (s *InstanceService) RestartHandler(c *gin.Context) {
	var req instancepb.RestartRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Use InstanceService to handle request
	result, err := s.restart(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// DeleteHandler delete instance handler
func (s *InstanceService) DeleteHandler(c *gin.Context) {
	var req instancepb.DeleteRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Use InstanceService to handle request
	result, err := s.delete(req.InstanceId)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// StatusHandler query instance status handler
func (s *InstanceService) StatusHandler(c *gin.Context) {
	var req instancepb.GetStatusRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Use InstanceService to handle request
	result, err := s.getStatus(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// LogsHandler get managed instance logs handler
func (s *InstanceService) LogsHandler(c *gin.Context) {
	var req instancepb.LogsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Validate required fields
	if req.InstanceId == "" {
		common.GinError(c, i18nresp.CodeInternalError, "missing required field: instanceId")
		return
	}

	// Use InstanceService to handle request
	result, err := s.getLogs(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// create writes instance method
func (s *InstanceService) create(req *instancepb.CreateRequest) (*instancepb.CreateResp, error) {

	// Generate instance ID (UUID)
	instanceID := uuid.New().String()

	// Hosting mode, Stdio protocol
	switch req.AccessType {
	case instancepb.AccessType_DIRECT:
		return s.createInstanceDirectMode(req, instanceID)
	case instancepb.AccessType_PROXY:
		return s.createInstanceProxyMode(req, instanceID)
	case instancepb.AccessType_HOSTING:
		return s.createInstanceHosting(req, instanceID)
	default:
		return nil, fmt.Errorf("unsupported access type: %v", req.AccessType)
	}
}

// detail gets instance details
func (s *InstanceService) detail(req *instancepb.DetailRequest) (*instancepb.DetailResp, error) {
	// Get instance information
	instance, err := biz.GInstanceBiz.GetInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("instance does not exist")
	}

	// Convert access type
	pbAccessType, err := common.ConvertToProtoAccessType(instance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}

	// Convert MCP protocol type
	pbMcpProtocol, err := common.ConvertToProtoMcpProtocol(instance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert MCP protocol type: %w", err)
	}

	// Build response
	resp := &instancepb.DetailResp{
		InstanceId:  instance.InstanceID,
		Name:        instance.InstanceName,
		Status:      string(instance.Status),
		AccessType:  pbAccessType,
		McpProtocol: pbMcpProtocol,
		Notes:       instance.Notes,
		IconPath:    instance.IconPath,
	}

	// Add specific fields based on access type
	switch instance.AccessType {
	case model.AccessTypeHosting:
		resp.PackageId = instance.PackageID
		resp.EnvironmentId = int32(instance.EnvironmentID)
		resp.McpServerId = instance.McpServerID
		resp.TemplateId = int32(instance.TemplateID)
		resp.ImgAddress = instance.ImgAddr
		resp.McpServers = string(instance.SourceConfig)
		resp.Port = instance.Port
		resp.InitScript = instance.InitScript
		resp.Command = instance.Command
		resp.ServicePath = instance.ServicePath
		resp.ContainerName = instance.ContainerName
		resp.ContainerServiceName = instance.ContainerServiceName
		resp.ContainerStatus = string(instance.ContainerStatus)
		resp.ContainerLastMessage = instance.ContainerLastMessage
		resp.ContainerIsReady = instance.ContainerIsReady

		// Convert environment variables
		if len(instance.EnvironmentVariables) > 0 {
			envVarsMap := make(map[string]string)
			if err := json.Unmarshal(instance.EnvironmentVariables, &envVarsMap); err == nil {
				resp.EnvironmentVariables = envVarsMap
			}
		}

		// Convert volume mounts
		if len(instance.VolumeMounts) > 0 {
			var volumeMounts []*instancepb.VolumeMount
			if err := json.Unmarshal(instance.VolumeMounts, &volumeMounts); err == nil {
				resp.VolumeMounts = volumeMounts
			}
		}

		// Convert tokens
		resp.Tokens = common.ConvertToProtoMcpToken(instance.Tokens)

		// Convert public proxy configuration
		resp.PublicProxyConfig = string(instance.PublicProxyConfig)

	case model.AccessTypeDirect, model.AccessTypeProxy:
		// For direct and proxy mode, add MCP servers configuration
		if len(instance.SourceConfig) > 0 {
			resp.McpServers = string(instance.SourceConfig)
		}
	}

	return resp, nil
}

func (s *InstanceService) list(req *instancepb.ListRequest) (*instancepb.ListResp, error) {
	// Parameter validation
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
	if req.InstanceName != "" {
		filters["instanceName"] = req.InstanceName
	}
	if req.EnvironmentId > 0 {
		filters["environmentId"] = req.EnvironmentId
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.ContainerStatus != "" {
		filters["containerStatus"] = req.ContainerStatus
	}
	if req.AccessType > 0 {
		accessType, err := common.ConvertToModelAccessType(req.AccessType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert access type: %w", err)
		}
		filters["accessType"] = accessType
	}
	if req.McpProtocol > 0 {
		mcpProtocol, err := common.ConvertToModelMcpProtocol(req.McpProtocol)
		if err != nil {
			return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
		}
		filters["mcpProtocol"] = mcpProtocol
	}

	// Sort parameters
	sortBy := "createdAt"
	sortOrder := "desc"

	return biz.GInstanceBiz.ListInstance(page, pageSize, filters, sortBy, sortOrder)
}

// GetLogs get instance logs
func (s *InstanceService) getLogs(req *instancepb.LogsRequest) (*instancepb.LogsResp, error) {
	// Set default number of lines
	lines := req.Lines
	if lines <= 0 {
		lines = 100
	}

	instance, err := biz.GInstanceBiz.GetInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("instance does not exist")
	}

	var response instancepb.LogsResp
	response.InstanceId = req.InstanceId

	// Check if it is a managed instance
	if instance.AccessType != model.AccessTypeHosting {
		response.IsManaged = false
		response.Message = "Instance is not of managed type"
		return &response, nil
	}

	response.IsManaged = true

	// Get environment information
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, instance.EnvironmentID)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to get environment information: %v", err)
		return &response, nil
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		response.Message = "Environment type error, only Kubernetes environment is supported"
		return &response, nil
	}

	// Get container logs
	logs, err := biz.GContainerBiz.GetContainerLogs(biz.ContainerLogsParams{
		InstanceID: req.InstanceId,
		Lines:      int64(lines),
	})
	if err != nil {
		response.Message = fmt.Sprintf("Failed to get container logs: %v", err)
		return &response, nil
	}

	response.Logs = logs
	response.Message = "Logs retrieved successfully"

	return &response, nil
}

// GetStatus retrieves the status of an instance
func (s *InstanceService) getStatus(req *instancepb.GetStatusRequest) (*instancepb.GetStatusResp, error) {
	instance, err := biz.GInstanceBiz.GetInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("instance does not exist")
	}

	var response *instancepb.GetStatusResp

	// Use different status query strategies based on access type
	switch instance.AccessType {
	case model.AccessTypeHosting:
		// Hosting type: query container status
		params := biz.ContainerStatusParams{
			InstanceID: req.InstanceId,
		}
		result, err := biz.GContainerBiz.GetContainerStatus(params)
		if err != nil {
			return nil, fmt.Errorf("failed to get container status: %s", err.Error())
		}

		response = result
	case model.AccessTypeProxy:
		_, _, tMcpConfig, err := instance.GetTargetConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get target configuration: %s", err.Error())
		}

		// Use HTTP probe to check service availability
		probeResult := utils.ProbePortFromURL(s.ctx, tMcpConfig.URL, 5*time.Second)

		// Build response
		response = &instancepb.GetStatusResp{
			InstanceId: req.InstanceId,
			Status:     string(instance.Status),
		}

		// If probe fails, add error message
		if !probeResult.Success {
			response.ProbeHttp = false
		} else {
			response.ProbeHttp = true
		}
	case model.AccessTypeDirect:
		_, _, sMcpConfig, err := instance.GetTargetConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get target configuration: %s", err.Error())
		}

		// Use HTTP probe to check service availability
		probeResult := utils.ProbePortFromURL(s.ctx, sMcpConfig.URL, 5*time.Second)

		// Build response
		response = &instancepb.GetStatusResp{
			InstanceId: req.InstanceId,
			Status:     string(instance.Status),
		}

		// If probe fails, add error message
		if !probeResult.Success {
			response.ProbeHttp = false
		} else {
			response.ProbeHttp = true
		}
	default:
		return nil, fmt.Errorf("unsupported access type")
	}

	return response, nil
}

// delete deletes an instance
func (s *InstanceService) delete(instanceID string) (*instancepb.DeleteResp, error) {
	req := &instancepb.DeleteRequest{
		InstanceId: instanceID,
	}

	// Get instance information directly
	instance, err := biz.GInstanceBiz.GetInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("instance does not exist")
	}

	switch instance.AccessType {
	case model.AccessTypeHosting:
		_, err = biz.GContainerBiz.DeleteContainer(instance)
		if err != nil {
			return nil, fmt.Errorf("failed to delete container: %v", err)
		}
	}

	// Disable the instance and set deletion time
	err = biz.GInstanceBiz.DeleteInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to disable instance: %v", err)
	}

	return &instancepb.DeleteResp{Message: "Instance deleted successfully"}, nil
}

// restart restarts an instance
func (s *InstanceService) restart(req *instancepb.RestartRequest) (*instancepb.RestartResp, error) {
	// 1. Query instance data by ID
	instance, err := s.getInstanceByID(req.InstanceId)
	if err != nil {
		return nil, err
	}

	switch instance.AccessType {
	case model.AccessTypeHosting:
		_, err = biz.GContainerBiz.RestartContainer(instance)
		if err != nil {
			return nil, fmt.Errorf("failed to restart container: %w", err)
		}
	default:
		return nil, fmt.Errorf("this service does not need to be restarted")
	}

	// 3. Update container status to pending
	if err = s.updateInstanceStatusToPending(instance); err != nil {
		return nil, err
	}

	pbAccessType, err := common.ConvertToProtoAccessType(instance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}

	// 4. Return restart result
	return &instancepb.RestartResp{
		InstanceId:        instance.InstanceID,
		Name:              instance.InstanceName,
		Status:            string(instance.Status),
		AccessType:        pbAccessType,
		AccessConfig:      s.convertMcpConfigToProto(instance.TargetConfig),
		PublicProxyConfig: s.convertMcpConfigToProto(instance.PublicProxyConfig),
		Message:           "Instance restarted successfully",
	}, nil
}

// disable disables an instance
func (s *InstanceService) disable(req *instancepb.DisabledRequest) (*instancepb.DisabledResp, error) {
	// Disable the instance and set deletion time
	msg, err := biz.GInstanceBiz.DisableInstance(req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to disable instance: %v", err)
	}

	return &instancepb.DisabledResp{Message: msg}, nil
}

// getInstanceByID retrieves an instance by its ID
func (s *InstanceService) getInstanceByID(instanceID string) (*model.McpInstance, error) {
	instance, err := biz.GInstanceBiz.GetInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance information: %v", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("instance does not exist")
	}
	return instance, nil
}

// updateInstanceStatusToPending updates instance status to pending
func (s *InstanceService) updateInstanceStatusToPending(instance *model.McpInstance) error {
	instance.Status = model.InstanceStatusActive
	instance.ContainerStatus = model.ContainerStatusPending
	instance.ContainerLastMessage = "Instance is restarting"
	if err := mysql.McpInstanceRepo.Update(s.ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance status: %v", err)
	}
	return nil
}

// createInstanceDirectMode direct connection mode handler function
func (s *InstanceService) createInstanceDirectMode(req *instancepb.CreateRequest, instanceID string) (*instancepb.CreateResp, error) {
	accessType, err := common.ConvertToModelAccessType(req.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToModelMcpProtocol(req.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}
	sourceType, err := common.ConvertToModelSourceType(req.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source type: %w", err)
	}
	if len(req.McpServers) == 0 {
		return nil, fmt.Errorf("missing required field: mcpServers")
	}
	// Validate MCP configuration format
	validationResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !validationResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", validationResult.ErrorMessage)
	}
	if validationResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}
	if validationResult.ProtocolType != string(mcpProtocol) {
		return nil, fmt.Errorf("mcp servers config is invalid: protocol type is %s, expected %s", validationResult.ProtocolType, mcpProtocol)
	}

	sourceConfig := json.RawMessage([]byte(req.McpServers))
	// Create new instance record
	instance := &model.McpInstance{
		InstanceID:        instanceID,
		InstanceName:      req.Name,
		AccessType:        accessType,
		McpProtocol:       mcpProtocol,
		SourceType:        sourceType,
		SourceConfig:      sourceConfig,
		TargetConfig:      sourceConfig,
		PublicProxyConfig: sourceConfig,
		Status:            model.InstanceStatusActive,
		IconPath:          req.IconPath,         // Add iconPath field handling
		Notes:             req.Notes,            // Add notes field handling
		McpServerID:       req.McpServerId,      // Add mcpServerId field handling
		TemplateID:        uint(req.TemplateId), // Add templateId field handling
		ServicePath:       req.ServicePath,      // Add servicePath field handling
	}

	// Save instance to database
	if err := biz.GInstanceBiz.CreateInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return &instancepb.CreateResp{
		InstanceId:  instanceID,
		Name:        req.Name,
		Status:      string(model.InstanceStatusActive),
		AccessType:  req.AccessType,
		McpProtocol: req.McpProtocol,
	}, nil
}

// createInstanceProxyMode proxy mode handler function
func (s *InstanceService) createInstanceProxyMode(req *instancepb.CreateRequest, instanceID string) (*instancepb.CreateResp, error) {
	accessType, err := common.ConvertToModelAccessType(req.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToModelMcpProtocol(req.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}
	sourceType, err := common.ConvertToModelSourceType(req.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source type: %w", err)
	}
	if len(req.McpServers) == 0 {
		return nil, fmt.Errorf("missing required field: mcpServers")
	}
	// Validate MCP configuration format
	validationResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !validationResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", validationResult.ErrorMessage)
	}
	if validationResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}
	if validationResult.ProtocolType != string(mcpProtocol) {
		return nil, fmt.Errorf("mcp servers config is invalid: protocol type is %s, expected %s", validationResult.ProtocolType, mcpProtocol)
	}

	// Create proxy configuration
	publicProxyConfig := biz.GInstanceBiz.CreatePublicProxyConfig(instanceID, mcpProtocol)
	pb, _ := common.MarshalAndAssignConfig(publicProxyConfig)

	// Create new instance record
	instance := &model.McpInstance{
		InstanceID:        instanceID,
		InstanceName:      req.Name,
		AccessType:        accessType,
		McpProtocol:       mcpProtocol,
		SourceType:        sourceType,
		SourceConfig:      json.RawMessage([]byte(req.McpServers)),
		PublicProxyConfig: pb,
		TargetConfig:      json.RawMessage([]byte(req.McpServers)),
		Status:            model.InstanceStatusActive,
		IconPath:          req.IconPath,         // Add iconPath field handling
		Notes:             req.Notes,            // Add notes field handling
		McpServerID:       req.McpServerId,      // Add mcpServerId field handling
		TemplateID:        uint(req.TemplateId), // Add templateId field handling
		ServicePath:       req.ServicePath,      // Add servicePath field handling
	}

	// Save instance to database
	if err := biz.GInstanceBiz.CreateInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return &instancepb.CreateResp{
		InstanceId:  instanceID,
		Name:        req.Name,
		Status:      string(model.InstanceStatusActive),
		AccessType:  req.AccessType,
		McpProtocol: req.McpProtocol,
	}, nil
}

// createInstanceHosting Hosting mode handler function
func (s *InstanceService) createInstanceHosting(req *instancepb.CreateRequest, instanceID string) (*instancepb.CreateResp, error) {

	// Validate timeout parameters
	if err := s.validateTimeoutParams(int(req.StartupTimeout), int(req.RunningTimeout)); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}
	mcpProtocol, err := common.ConvertToModelMcpProtocol(req.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}
	sourceType, err := common.ConvertToModelSourceType(req.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source type: %w", err)
	}

	if req.Port <= 0 {
		return nil, fmt.Errorf("missing required field: port")
	}
	// Validate environment ID
	if req.EnvironmentId == 0 {
		return nil, fmt.Errorf("hosting type instance requires environment ID")
	}
	if req.ImgAddress == "" {
		return nil, fmt.Errorf("missing required field: imgAddress")
	}
	// Query Kubernetes configuration and namespace based on environment ID
	environment, err := biz.GEnvironmentBiz.GetEnvironment(s.ctx, uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get environment information: %w", err)
	}

	// Validate environment type
	if environment.Environment != model.McpEnvironmentKubernetes {
		return nil, fmt.Errorf("environment type is not Kubernetes, cannot create container")
	}

	if mcpProtocol == model.McpProtocolStdio {
		mcpServers := req.McpServers
		if len(mcpServers) == 0 {
			return nil, fmt.Errorf("mcp servers config is empty")
		}
		reqMcpResult, err2 := utils.ValidateMcpConfig([]byte(mcpServers))
		if err2 != nil {
			return nil, fmt.Errorf("failed to validate mcp servers: %w", err2)
		}
		if !reqMcpResult.IsValid {
			return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
		}
		if !reqMcpResult.HasCommand {
			return nil, fmt.Errorf("mcp servers config is invalid: command is required")
		}
	}
	containerOptions, err := biz.GContainerBiz.BuildContainerOptions(s.ctx, instanceID, mcpProtocol, req.McpServers, req.PackageId, req.Port,
		req.InitScript, req.Command, req.ImgAddress, req.EnvironmentVariables, req.VolumeMounts, int32(req.StartupTimeout), int32(req.RunningTimeout))
	if err != nil {
		return nil, fmt.Errorf("failed to build container options: %w", err)
	}
	err = biz.GContainerBiz.CreateContainer(containerOptions, req.EnvironmentId, req.StartupTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Create target configuration
	toMcpProtocol := mcpProtocol
	if mcpProtocol == model.McpProtocolStdio {
		toMcpProtocol = model.McpProtocolSSE
	}
	// Call data layer to create container
	tb := []byte{}
	switch mcpProtocol {
	case model.McpProtocolStdio:
		if strings.Contains(req.ImgAddress, common.DefatuleHostingImg) {
			targetConfig := common.CreateTargetProxyConfigForDefatuleHostingImg(containerOptions.ServiceName, containerOptions.Port, containerOptions.ContainerName, toMcpProtocol)
			tb, _ = common.MarshalAndAssignConfig(targetConfig)
		}
	case model.McpProtocolSSE, model.McpProtocolStreamableHttp:
		targetConfig := common.CreateTargetProxyConfigForHttp(containerOptions.ServiceName, containerOptions.Port, containerOptions.ContainerName, mcpProtocol, req.ServicePath)
		tb, _ = common.MarshalAndAssignConfig(targetConfig)
	default:
		return nil, fmt.Errorf("unsupported mcp protocol: %v", mcpProtocol)
	}
	// Create proxy configuration
	publicProxyConfig := biz.GInstanceBiz.CreatePublicProxyConfig(instanceID, toMcpProtocol)
	pb, _ := common.MarshalAndAssignConfig(publicProxyConfig)

	// Create new instance record
	containerCreateOptions, err := common.MarshalAndAssignConfig(containerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal container create options: %w", err)
	}
	evs, err := common.MarshalAndAssignConfig(req.EnvironmentVariables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment variables: %w", err)
	}
	vms, err := common.MarshalAndAssignConfig(req.VolumeMounts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volume mounts: %w", err)
	}
	instance := &model.McpInstance{
		InstanceID:             instanceID,
		InstanceName:           req.Name,
		AccessType:             model.AccessTypeHosting,
		McpProtocol:            mcpProtocol,
		Status:                 model.InstanceStatusActive,
		PackageID:              req.PackageId,
		ContainerStatus:        model.ContainerStatusPending,
		EnvironmentID:          uint(req.EnvironmentId),
		SourceType:             sourceType,
		McpServerID:            req.McpServerId,
		TemplateID:             uint(req.TemplateId),
		Tokens:                 common.ConvertProtoTokensToModel(req.Tokens),
		ImgAddr:                req.ImgAddress,
		Port:                   req.Port,
		InitScript:             req.InitScript,
		Command:                req.Command,
		EnvironmentVariables:   evs,
		VolumeMounts:           vms,
		ContainerName:          containerOptions.ContainerName,
		ContainerServiceName:   containerOptions.ServiceName,
		ContainerIsReady:       false,
		ContainerCreateOptions: containerCreateOptions,
		ContainerLastMessage:   "container is pending",
		StartupTimeout:         int64(req.StartupTimeout),
		RunningTimeout:         int64(req.RunningTimeout),
		SourceConfig:           json.RawMessage(req.McpServers),
		TargetConfig:           tb,
		PublicProxyConfig:      pb,
		ServicePath:            req.ServicePath,
		Notes:                  req.Notes,
		IconPath:               req.IconPath,
	}

	// Save instance to database
	if err := biz.GInstanceBiz.CreateInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return &instancepb.CreateResp{
		InstanceId:  instanceID,
		Name:        req.Name,
		Status:      string(model.InstanceStatusActive),
		AccessType:  req.AccessType,
		McpProtocol: req.McpProtocol,
	}, nil
}

// validateTimeoutParams validates timeout parameters
func (s *InstanceService) validateTimeoutParams(startupTimeout, runningTimeout int) error {
	// Startup timeout validation
	if startupTimeout < 0 {
		return fmt.Errorf("startup timeout cannot be negative")
	}
	if startupTimeout > 0 && startupTimeout < 30 {
		return fmt.Errorf("startup timeout cannot be less than 30 seconds")
	}
	if startupTimeout > 3600 {
		return fmt.Errorf("startup timeout cannot exceed 3600 seconds")
	}

	// Running timeout validation
	if runningTimeout < 0 {
		return fmt.Errorf("running timeout cannot be negative")
	}
	if runningTimeout > 0 && runningTimeout < 60 {
		return fmt.Errorf("running timeout cannot be less than 60 seconds")
	}
	if runningTimeout > 86400 {
		return fmt.Errorf("running timeout cannot exceed 86400 seconds")
	}

	return nil
}

// convertMcpConfigToProto converts JSON configuration from database to proto structure
func (s *InstanceService) convertMcpConfigToProto(configData json.RawMessage) *instancepb.McpServersConfig {
	if len(configData) == 0 {
		return nil
	}

	// Parse configuration from database
	_, mcpConfig, err := model.ParseMcpServersConfig(configData)
	if err != nil || mcpConfig == nil {
		return nil
	}

	// Extract host and port information from URL
	host := "localhost"
	port := int32(8080)
	protocol := "http"

	if mcpConfig.URL != "" {
		// Simple URL parsing to get host and port
		if strings.HasPrefix(mcpConfig.URL, "http://") {
			protocol = "http"
			urlPart := strings.TrimPrefix(mcpConfig.URL, "http://")
			if parts := strings.Split(urlPart, ":"); len(parts) >= 2 {
				host = parts[0]
				if portStr := strings.Split(parts[1], "/")[0]; portStr != "" {
					if p, err := strconv.Atoi(portStr); err == nil {
						port = int32(p)
					}
				}
			} else if len(parts) == 1 {
				host = strings.Split(parts[0], "/")[0]
			}
		} else if strings.HasPrefix(mcpConfig.URL, "https://") {
			protocol = "https"
			urlPart := strings.TrimPrefix(mcpConfig.URL, "https://")
			if parts := strings.Split(urlPart, ":"); len(parts) >= 2 {
				host = parts[0]
				if portStr := strings.Split(parts[1], "/")[0]; portStr != "" {
					if p, err := strconv.Atoi(portStr); err == nil {
						port = int32(p)
					}
				}
			} else if len(parts) == 1 {
				host = strings.Split(parts[0], "/")[0]
				port = 443
			}
		}
	}

	// If Transport field exists, use it first
	if mcpConfig.Transport != "" {
		protocol = mcpConfig.Transport
	}

	return &instancepb.McpServersConfig{
		Host:     host,
		Port:     port,
		Protocol: protocol,
	}
}
