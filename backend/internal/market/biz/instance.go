package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"strings"

	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"
)

// InstanceBiz instance data processing layer
type InstanceBiz struct {
	ctx context.Context
}

// GInstanceBiz global instance data processing layer instance
var GInstanceBiz *InstanceBiz

func init() {
	GInstanceBiz = NewInstanceBiz(context.Background())
}

// NewInstanceBiz create instance data processing layer instance
func NewInstanceBiz(ctx context.Context) *InstanceBiz {
	return &InstanceBiz{
		ctx: ctx,
	}
}

// GetInstance get instance information
func (biz *InstanceBiz) GetInstance(instanceID string) (*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByInstanceID(biz.ctx, instanceID)
}

// DisableInstance disable instance
func (biz *InstanceBiz) DisableInstance(instanceID string) (string, error) {
	instance, err := biz.GetInstance(instanceID)
	if err != nil {
		return "", err
	}
	msg := "Instance has been disabled"
	if instance.AccessType == model.AccessTypeHosting {
		res, err := GContainerBiz.DeleteContainer(instance)
		if err != nil {
			return "", err
		}
		msg = res.Message
	}
	instance.Status = model.InstanceStatusInactive
	instance.ContainerIsReady = false
	instance.ContainerStatus = model.ContainerStatusManualStop
	instance.ContainerLastMessage = msg
	return msg, mysql.McpInstanceRepo.Update(biz.ctx, instance)
}

// DeleteInstance delete instance
func (biz *InstanceBiz) DeleteInstance(instanceID string) error {
	// Get instance by access type
	_, err := mysql.McpInstanceRepo.FindByInstanceID(biz.ctx, instanceID)
	if err != nil {
		return err
	}
	return mysql.McpInstanceRepo.Delete(biz.ctx, instanceID)
}

// ListInstance get instance list
func (biz *InstanceBiz) ListInstance(page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) (*instancepb.ListResp, error) {
	// Query data
	instances, total, err := mysql.McpInstanceRepo.FindWithPagination(biz.ctx, page, pageSize, filters, sortBy, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to query instance list: %v", err)
	}

	// envIds
	envIds := make([]string, 0, len(instances))
	for _, instance := range instances {
		envIds = append(envIds, fmt.Sprintf("%d", instance.EnvironmentID))
	}
	envNames, err := mysql.McpEnvironmentRepo.FindNamesByIDs(biz.ctx, envIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment names: %v", err)
	}

	// Convert to proto response
	instanceInfos := make([]*instancepb.ListResp_InstanceInfo, 0, len(instances))
	for _, instance := range instances {
		instanceInfo := common.ConvertToInstanceInfo(instance)
		if envName, ok := envNames[fmt.Sprintf("%d", instance.EnvironmentID)]; ok {
			instanceInfo.EnvironmentName = envName
		}
		instanceInfos = append(instanceInfos, instanceInfo)
	}

	return &instancepb.ListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     instanceInfos,
	}, nil
}

// CreateInstance create instance
func (biz *InstanceBiz) CreateInstance(instance *model.McpInstance) error {
	if instance.InstanceName == "" {
		return fmt.Errorf("instance name cannot be empty")
	}
	// Check if name already exists
	existingInstance, err := mysql.McpInstanceRepo.FindByName(biz.ctx, instance.InstanceName)
	if err == nil && existingInstance != nil {
		return fmt.Errorf("instance name %s already exists", instance.InstanceName)
	}
	return mysql.McpInstanceRepo.Create(biz.ctx, instance)
}

// UpdateInstanceForDirect update instance
func (biz *InstanceBiz) UpdateInstanceForDirect(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	// Update basic information
	if req.Name != "" {
		oriInstance.InstanceName = req.Name
	}
	if req.Notes != "" {
		oriInstance.Notes = req.Notes
	}

	// Validate MCP configuration format
	reqMcpResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !reqMcpResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
	}
	if reqMcpResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}

	oriMcpResult, err := utils.ValidateMcpConfig([]byte(oriInstance.SourceConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !utils.CompareMcpValidationResult(reqMcpResult, oriMcpResult) {
		sourceConfig := json.RawMessage([]byte(req.McpServers))
		oriInstance.SourceConfig = sourceConfig
		oriInstance.TargetConfig = sourceConfig
		oriInstance.PublicProxyConfig = sourceConfig
	}

	// Save to database
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %v", err)
	}

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}

	return resp, nil
}

// UpdateInstanceForProxy update instance
func (biz *InstanceBiz) UpdateInstanceForProxy(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	// Update basic information
	if req.Name != "" {
		oriInstance.InstanceName = req.Name
	}
	if req.Notes != "" {
		oriInstance.Notes = req.Notes
	}
	// Validate MCP configuration format
	reqMcpResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !reqMcpResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
	}
	if reqMcpResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}

	oriMcpResult, err := utils.ValidateMcpConfig([]byte(oriInstance.SourceConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !utils.CompareMcpValidationResult(reqMcpResult, oriMcpResult) {
		sourceConfig := json.RawMessage([]byte(req.McpServers))
		oriInstance.SourceConfig = sourceConfig
		oriInstance.TargetConfig = sourceConfig
		// Create proxy configuration
		publicProxyConfig := biz.CreatePublicProxyConfig(oriInstance.InstanceID, oriInstance.McpProtocol)
		pb, e2 := common.MarshalAndAssignConfig(publicProxyConfig)
		if e2 != nil {
			return nil, fmt.Errorf("failed to marshal public proxy config: %w", e2)
		}
		oriInstance.PublicProxyConfig = pb
	}

	// 保存到数据库
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("更新实例失败: %v", err)
	}

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}

	return resp, nil
}

// UpdateInstanceForHosting 更新实例
func (biz *InstanceBiz) UpdateInstanceForHosting(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	var err error
	port := req.Port
	instanceID := req.InstanceId
	packageID := req.PackageId
	initScript := req.InitScript
	command := req.Command
	imgAddress := req.ImgAddress
	envs := req.EnvironmentVariables
	vms := req.VolumeMounts
	startupTimeout := req.StartupTimeout
	runningTimeout := req.RunningTimeout
	mcpServers := req.McpServers

	if oriInstance.McpProtocol == model.McpProtocolStdio {
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
		oriInstance.SourceConfig = json.RawMessage([]byte(mcpServers))
	}

	newContainerCreateOptions, err := GContainerBiz.BuildContainerOptions(ctx, instanceID, oriInstance.McpProtocol, mcpServers, packageID, port, initScript,
		command, imgAddress, envs, vms, startupTimeout, runningTimeout)
	if err != nil {
		return nil, fmt.Errorf("构建容器配置失败: %v", err)
	}
	containerCreateOptions, err := common.MarshalAndAssignConfig(newContainerCreateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal container create containerCreateOptions: %w", err)
	}

	// 删除旧的容器和svc服务
	_, err = GContainerBiz.DeleteContainer(oriInstance)
	if err != nil {
		return nil, fmt.Errorf("删除容器失败: %v", err)
	}

	// Create target configuration
	toMcpProtocol := oriInstance.McpProtocol
	if oriInstance.McpProtocol == model.McpProtocolStdio {
		toMcpProtocol = model.McpProtocolSSE
	}
	// Call data layer to create container
	tb := []byte{}
	switch oriInstance.McpProtocol {
	case model.McpProtocolStdio:
		if strings.Contains(req.ImgAddress, common.DefatuleHostingImg) {
			targetConfig := common.CreateTargetProxyConfigForDefatuleHostingImg(newContainerCreateOptions.ServiceName, newContainerCreateOptions.Port, newContainerCreateOptions.ContainerName, toMcpProtocol)
			tb, _ = common.MarshalAndAssignConfig(targetConfig)
		}
	case model.McpProtocolSSE, model.McpProtocolStreamableHttp:
		targetConfig := common.CreateTargetProxyConfigForHttp(newContainerCreateOptions.ServiceName, newContainerCreateOptions.Port, newContainerCreateOptions.ContainerName, oriInstance.McpProtocol, req.ServicePath)
		tb, _ = common.MarshalAndAssignConfig(targetConfig)
	default:
		return nil, fmt.Errorf("unsupported mcp protocol: %v", oriInstance.McpProtocol)
	}
	// Create proxy configuration
	publicProxyConfig := GInstanceBiz.CreatePublicProxyConfig(instanceID, toMcpProtocol)
	pb, _ := common.MarshalAndAssignConfig(publicProxyConfig)

	// 更新
	oriInstance.InstanceName = req.Name
	oriInstance.Notes = req.Notes
	oriInstance.Port = int32(port)
	oriInstance.InitScript = initScript
	oriInstance.Command = command
	oriInstance.ImgAddr = imgAddress
	oriInstance.EnvironmentVariables, _ = common.MarshalAndAssignConfig(envs)
	oriInstance.VolumeMounts, _ = common.MarshalAndAssignConfig(vms)
	oriInstance.StartupTimeout = int64(startupTimeout)
	oriInstance.RunningTimeout = int64(runningTimeout)
	oriInstance.ContainerCreateOptions = containerCreateOptions
	oriInstance.ContainerStatus = model.ContainerStatusPending
	oriInstance.ContainerIsReady = false
	oriInstance.SourceConfig = json.RawMessage([]byte(mcpServers))
	oriInstance.TargetConfig = tb
	oriInstance.PublicProxyConfig = pb
	oriInstance.ServicePath = req.ServicePath
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("更新实例失败: %v", err)
	}

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}
	return resp, nil
}

// CreatePublicProxyConfig creates public proxy configuration
func (biz *InstanceBiz) CreatePublicProxyConfig(instanceID string, mcpProtocol model.McpProtocol) *model.McpServersConfig {
	mcpName := fmt.Sprintf("mcp-%s", instanceID[:8])
	addr, _ := url.JoinPath(config.GlobalConfig.Domain, strings.TrimPrefix(common.GetGatewayRoutePrefix(), "/"), instanceID)
	if mcpProtocol == model.McpProtocolSSE {
		addr += fmt.Sprintf("/%s", mcpProtocol.String())
	}
	return &model.McpServersConfig{
		McpServers: map[string]*model.McpConfig{
			mcpName: {
				Type: mcpProtocol.String(),
				URL:  addr,
			},
		},
	}
}

// GetInstancesByEnvironmentID 根据环境ID获取实例列表
func (biz *InstanceBiz) GetInstancesByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByEnvironmentID(ctx, environmentID)
}
