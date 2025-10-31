package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/codepackage"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/utils"

	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"

	"go.uber.org/zap"
)

// TaskStatus 任务状态信息
// 移除TaskStatus结构体，不再使用任务管理

// ContainerBiz 容器数据层
type ContainerBiz struct {
	ctx context.Context
}

var GContainerBiz *ContainerBiz

func init() {
	GContainerBiz = NewContainerBiz(context.Background())
}

// NewContainerBiz 创建容器数据处理层实例
func NewContainerBiz(ctx context.Context) *ContainerBiz {
	return &ContainerBiz{
		ctx: ctx,
	}
}

type ContainerOptions struct {
	// 实例ID
	InstanceID string

	// 容器名称
	ContainerName string

	// McpServers 配置
	McpServers string

	// 端口映射配置
	PortMapping map[int]int

	// 初始化脚本内容
	InitScript string

	// 环境变量配置
	EnvironmentVariables map[string]string

	// 卷挂载配置（支持多个卷）
	VolumeMounts []k8s.UnifiedMount

	// 毫秒时间戳，默认 0 表示不检测一直创建，设置值时最大不能超过 1 天
	StartupTimeout int64

	// 毫秒时间戳，默认 0 表示常驻服务，设置值时最大不能超过 1 年（超过 1 年应该设置常驻）
	RunningTimeout int64

	// code package download link
	PackageDownloadLink string
}

// 删除不再使用的存储和节点配置结构体，亲和性逻辑已转移到Create方法中

// ContainerCreateResult 容器创建结果
type ContainerCreateResult struct {
	ContainerName string
	ServiceName   string
	ServicePort   int32
	Message       string
}

// ContainerDeleteParams 容器删除参数
type ContainerDeleteParams struct {
	InstanceID string
}

// ContainerDeleteResult 容器删除结果
type ContainerDeleteResult struct {
	ContainerName string
	ServiceName   string
	Message       string
}

// ContainerStatusParams 容器状态查询参数
type ContainerStatusParams struct {
	InstanceID string
}

// ContainerStatusResult 容器状态查询结果
type ContainerStatusResult struct {
	ContainerName  string
	ServiceName    string
	ErrorMessage   string
	ContainerReady bool                       // 容器是否就绪
	ServiceReady   bool                       // 服务是否就绪
	WarningEvents  []container.ContainerEvent // 警告事件
}

// CreateContainer 创建容器业务逻辑
func (cd *ContainerBiz) CreateHostingContainerForSSEAndSteamableHttp(req *instancepb.CreateRequest, instanceID string) (*ContainerCreateResult, error) {
	var err error
	shortInstanceId := instanceID[:8]

	// 1. 生成容器名称
	containerName := cd.generateContainerName(shortInstanceId)
	serviceName := cd.generateServiceName(shortInstanceId)

	// 2. 代码包下载链接生成
	packageId := req.PackageId
	codepkgInstallScript := ""
	if packageId != "" {
		// 生成代码包安装脚本
		codepkgInstallScript, err = cd.generateCodePkgInstallScript(packageId)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code package install script: %w", err)
		}
	}

	// 4. 生成镜像配置
	imgPms, err := cd.getMcpHostingImageCfgForSSEAndSteamableHttp(req.ImgAddress, req.Port, req.InitScript, req.Command, codepkgInstallScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
	}
	image := imgPms.image
	port := imgPms.port
	command := imgPms.command
	commandArgs := imgPms.commandArgs

	// 5. 设置环境变量
	envVars := make(map[string]string)
	envVars["MCP_INSTANCE_ID"] = instanceID
	envVars["MCP_PORT"] = fmt.Sprintf("%d", imgPms.port)
	envVars["NODE_ENV"] = "production"
	for k, v := range req.EnvironmentVariables {
		envVars[k] = v
	}

	// 6. 设置卷挂载配置（亲和性判断逻辑转移到Create方法中）
	mounts := []k8s.UnifiedMount{}
	if len(req.VolumeMounts) > 0 {
		for _, vm := range req.VolumeMounts {
			mounts = append(mounts, cd.volumeMountFromPb(vm))
		}
	}

	// 7. 设置标签
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	if req.StartupTimeout > 0 {
		labels["mcp.startup.timeout"] = fmt.Sprintf("%d", req.StartupTimeout)
	}
	if req.RunningTimeout > 0 {
		labels["mcp.running.timeout"] = fmt.Sprintf("%d", req.RunningTimeout)
	}

	// 8. 构建容器创建选项
	containerOptions := container.ContainerCreateOptions{
		ImageName:     image,
		ContainerName: containerName,
		Port:          port,
		Command:       command,
		CommandArgs:   commandArgs,
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       envVars,
		Mounts:        mounts,
		WorkingDir:    "/app",
	}

	// 9. 设置超时上下文
	ctx := cd.ctx
	if req.StartupTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(cd.ctx, time.Duration(req.StartupTimeout)*time.Second)
		defer cancel()
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// 统一使用容器管理器创建（简化判断逻辑）
	containerName, err = entry.GetContainerManager().Create(ctx, containerOptions)
	if err != nil {
		// 删除容器（如果容器名称不为空）
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerCreateFailure)+": %v", err)
	}

	// 创建svc
	_, err = entry.GetServiceManager().Create(ctx, serviceName, port, labels)
	if err != nil {
		// 删除容器（如果容器名称不为空）
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceCreateFailure)+": %w", err)
	}

	// 11. 返回创建结果，包含实例更新所需的数据
	return &ContainerCreateResult{
		ContainerName: containerName,
		ServiceName:   serviceName,
		ServicePort:   port,
		Message:       i18n.FormatWithContext(cd.ctx, i18n.CodeContainerCreateSuccess),
	}, nil
}

// CreateContainer 创建容器业务逻辑
func (cd *ContainerBiz) CreateHostingContainerForStdio(req *instancepb.CreateRequest, instanceID string) (*ContainerCreateResult, error) {
	var err error
	// 1. 生成容器名称
	containerName := cd.generateContainerName(instanceID)
	serviceName := cd.generateServiceName(instanceID)

	// 2. 代码包下载链接生成
	packageId := req.PackageId
	codepkgInstallScript := ""
	if packageId != "" {
		// 生成代码包安装脚本
		codepkgInstallScript, err = cd.generateCodePkgInstallScript(packageId)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code package install script: %w", err)
		}
	}

	// 3. 验证MCP配置
	validateInfo, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp config: %w", err)
	}
	// 检查MCP配置是否有效：非有效或者非stdio协议类型
	if validateInfo == nil {
		return nil, fmt.Errorf("mcpServers config is invalid: %s", err)
	}
	if !validateInfo.IsValid || validateInfo.ProtocolType != model.McpProtocolStdio.String() {
		return nil, fmt.Errorf("mcp config is invalid protocol type: %s", validateInfo.ProtocolType)
	}

	// 4. 生成镜像配置
	imgPms, err := cd.getMcpHostingImageCfg(req.ImgAddress, req.Port, req.InitScript, codepkgInstallScript, req.McpServers)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
	}
	image := imgPms.image
	port := imgPms.port
	command := imgPms.command
	commandArgs := imgPms.commandArgs

	// 5. 设置环境变量
	envVars := make(map[string]string)
	envVars["MCP_INSTANCE_ID"] = instanceID
	envVars["MCP_PORT"] = fmt.Sprintf("%d", imgPms.port)
	envVars["NODE_ENV"] = "production"
	for k, v := range req.EnvironmentVariables {
		envVars[k] = v
	}

	// 6. 设置卷挂载配置（亲和性判断逻辑转移到Create方法中）
	mounts := []k8s.UnifiedMount{}
	if len(req.VolumeMounts) > 0 {
		for _, vm := range req.VolumeMounts {
			mounts = append(mounts, cd.volumeMountFromPb(vm))
		}
	}

	// 7. 设置标签
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	if req.StartupTimeout > 0 {
		labels["mcp.startup.timeout"] = fmt.Sprintf("%d", req.StartupTimeout)
	}
	if req.RunningTimeout > 0 {
		labels["mcp.running.timeout"] = fmt.Sprintf("%d", req.RunningTimeout)
	}

	// 8. 构建容器创建选项
	containerOptions := container.ContainerCreateOptions{
		ImageName:     image,
		ContainerName: containerName,
		Port:          port,
		Command:       command,
		CommandArgs:   commandArgs,
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       envVars,
		Mounts:        mounts,
		WorkingDir:    "/app",
	}

	// 9. 设置超时上下文
	ctx := cd.ctx
	if req.StartupTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(cd.ctx, time.Duration(req.StartupTimeout)*time.Second)
		defer cancel()
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// 统一使用容器管理器创建（简化判断逻辑）
	containerName, err = entry.GetContainerManager().Create(ctx, containerOptions)
	if err != nil {
		// 删除容器（如果容器名称不为空）
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerCreateFailure)+": %v", err)
	}

	// 创建svc
	_, err = entry.GetServiceManager().Create(ctx, serviceName, port, labels)
	if err != nil {
		// 删除容器
		_ = entry.GetContainerManager().Delete(ctx, containerName)
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceCreateFailure)+": %w", err)
	}

	// 11. 返回创建结果，包含实例更新所需的数据
	return &ContainerCreateResult{
		ContainerName: containerName,
		ServiceName:   serviceName,
		ServicePort:   port,
		Message:       i18n.FormatWithContext(cd.ctx, i18n.CodeContainerCreateSuccess),
	}, nil
}

// CreateContainer 创建容器业务逻辑
func (cd *ContainerBiz) CreateContainer(containerCreateOptions *container.ContainerCreateOptions, environmentId int32, startupTimeout int32) error {
	// 9. 设置超时上下文
	ctx := cd.ctx
	if startupTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(cd.ctx, time.Duration(startupTimeout)*time.Second)
		defer cancel()
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, uint(environmentId))
	if err != nil {
		return fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// create container
	containerName, err := entry.GetContainerManager().Create(ctx, *containerCreateOptions)
	if err != nil {
		// 删除容器（如果容器名称不为空）
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerCreateFailure)+": %v", err)
	}

	// create service
	_, err = entry.GetServiceManager().Create(ctx, containerCreateOptions.ServiceName, containerCreateOptions.Port, containerCreateOptions.Labels)
	if err != nil {
		// 删除容器（如果容器名称不为空）
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceCreateFailure)+": %w", err)
	}

	return nil
}

// DeleteContainer 删除容器业务逻辑
func (cd *ContainerBiz) DeleteContainer(instance *model.McpInstance) (*ContainerDeleteResult, error) {
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}
	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	message := ""
	// 2. 删除容器
	if err = entry.GetContainerManager().Delete(cd.ctx, instance.ContainerName); err != nil {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeDeleteContainerFailure)+": %v \n", err)
	} else {
		message += i18n.FormatWithContext(cd.ctx, i18n.CodeContainerDeleteSuccess) + " \n"
	}

	// 3. 删除服务
	if err = entry.GetServiceManager().Delete(cd.ctx, instance.ContainerServiceName); err != nil {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceDeleteFailure)+": %v", err.Error())
	} else {
		message += i18n.FormatWithContext(cd.ctx, i18n.CodeServiceDeleteSuccess) + " \n"
	}

	resp := &ContainerDeleteResult{
		ContainerName: instance.ContainerName,
		ServiceName:   instance.ContainerServiceName,
		Message:       message,
	}
	return resp, nil
}

// GetContainerStatus 获取容器详细状态信息，包括容器异常检测和服务探测
func (cd *ContainerBiz) GetContainerStatus(params ContainerStatusParams) (*instancepb.GetStatusResp, error) {
	// 1. 根据 instanceID 获取实例配置
	instance, err := mysql.McpInstanceRepo.FindByInstanceIDAndAccessType(
		context.Background(),
		params.InstanceID,
		model.AccessTypeHosting, // 托管模式才需要查询容器状态
	)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceNotHostingMode)+": %w", err)
	}
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}

	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	message := ""
	warningEvents := make([]container.ContainerEvent, 0)
	// 3. 检查容器就绪状态
	containerReady, runInfo, err := entry.GetContainerManager().IsReady(cd.ctx, instance.ContainerName)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerReadyCheckFailure)+": %w", err)
	}
	if !containerReady {

		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerNotReady)+": %s \n", runInfo)
		// 4. 获取容器警告事件
		warningEvents, err = entry.GetContainerManager().GetWarningEvents(cd.ctx, instance.ContainerName)
		if err != nil {
			message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetContainerWarningEventsFailure)+": %v \n", err)
		}
	}

	// 5. 主动探测服务是否正常运行
	svc, svcErr := entry.GetServiceManager().Get(cd.ctx, instance.ContainerServiceName)
	svcReady := false
	if svcErr == nil {
		// 检查服务配置是否正常
		if svc.ClusterIP != "" {
			// 对于 Headless Service，ClusterIP 为 "None" 也是正常的
			if svc.ClusterIP == "None" || svc.ClusterIP == "docker-network" {
				// Headless Service 或 Docker 网络，检查是否有端口配置
				svcReady = len(svc.Ports) > 0
			} else {
				// 普通 Service，检查 ClusterIP 和端口配置
				svcReady = len(svc.Ports) > 0
			}
		}
	} else {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceStatusAbnormal)+": %v \n", svcErr)
	}

	// 6. 更新实例信息
	if containerReady && svcReady {
		instance.ContainerStatus = model.ContainerStatusRunning
		instance.ContainerIsReady = true
		instance.ContainerLastMessage = message
	} else {
		instance.ContainerIsReady = false
		instance.ContainerLastMessage = message
	}
	err = mysql.McpInstanceRepo.Update(context.Background(), instance)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeUpdateInstanceFailure)+": %w", err)
	}

	events := make([]*instancepb.ContainerEvent, 0, len(warningEvents))
	for _, event := range warningEvents {
		events = append(events, &instancepb.ContainerEvent{
			Type:          event.Type,
			Reason:        event.Reason,
			Message:       event.Message,
			LastTimestamp: event.Timestamp,
		})
	}

	_, _, mcpCfg, err := instance.GetTargetConfig()
	if err != nil {
		return nil, fmt.Errorf("获取目标配置失败: %s", err.Error())
	}
	// Use HTTP probe to check service availability
	probeResult := utils.ProbePortFromURL(cd.ctx, mcpCfg.URL, 5*time.Second)

	probeHttp := false
	if probeResult.Success {
		probeHttp = true
	} else {
		message += fmt.Sprintf("HTTP 探测失败: %s", probeResult.Error)
	}

	resp := &instancepb.GetStatusResp{
		InstanceId:     params.InstanceID,
		Status:         string(instance.Status),
		ContainerName:  instance.ContainerName,
		RuntimeType:    string(entry.GetRuntimeType()),
		ContainerReady: containerReady,
		ServiceReady:   svcReady,
		ProbeHttp:      probeHttp,
		WarningEvents:  events,
		ErrorMessage:   message,
	}

	return resp, nil
}

// generateContainerName 生成容器名称
func (cd *ContainerBiz) generateContainerName(instanceID string) string {
	// 生成基于实例 ID 的容器名称
	instanceID = instanceID[:8]
	return fmt.Sprintf("mcp-instance-%s-container", instanceID)
}

// generateServiceName 生成服务名称
func (cd *ContainerBiz) generateServiceName(instanceID string) string {
	instanceID = instanceID[:8]
	return fmt.Sprintf("mcp-instance-%s-service", instanceID)
}

type imageParams struct {
	image       string
	port        int32
	command     []string
	commandArgs []string
}

func (cd *ContainerBiz) getMcpHostingImageCfg(imgAddress string, port int32, initScript string, codepkgInstallScript string, mcpServerCfg string) (*imageParams, error) {
	if len(imgAddress) == 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeImageAddressRequired))
	}
	if port == 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodePortRequired))
	}
	if len(initScript) == 0 {
		initScript = "echo 'No initialization commands specified'"
	}

	// Build complete startup script
	startupScript := fmt.Sprintf(`
		# Create working directory
		mkdir -p /app/init
		# Generate initialization script dynamically
		cat > /app/init/startup.sh << 'EOF'
#!/bin/sh
set -e

# Download and extract code package
%s

echo "[$(date)] Starting initialization script execution..."
%s
echo "[$(date)] Initialization script execution completed"
EOF
		# Write /app/mcp-servers.json
		cat > /app/mcp-servers.json << 'EOF'
%s
EOF

		# Set script execution permissions
		chmod +x /app/init/startup.sh
		
		# Execute initialization script
		/app/init/startup.sh
		
		# Start main program
		echo "[$(date)] Starting main program: mcp-hosting --port=%d --mcp-servers-config /app/mcp-servers.json"
		mcp-hosting --port=%d --mcp-servers-config /app/mcp-servers.json
	`,
		codepkgInstallScript,
		initScript,
		mcpServerCfg,
		port,
		port)

	imgPms := &imageParams{
		image:       imgAddress,
		port:        port,
		command:     []string{"/bin/sh"},
		commandArgs: []string{"-c", startupScript},
	}

	return imgPms, nil
}

func (cd *ContainerBiz) getMcpHostingImageCfgForSSEAndSteamableHttp(imgAddress string, port int32, initScript string, command string, codepkgInstallScript string) (*imageParams, error) {
	if len(imgAddress) == 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeImageAddressRequired))
	}
	if len(command) == 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeStartupCommandRequired))
	}

	// Build complete startup script
	startupScript := fmt.Sprintf(`
		# Create working directory
		mkdir -p /app/init
		# Generate initialization script dynamically
		cat > /app/init/startup.sh << 'EOF'
#!/bin/sh
set -e
# Download and extract code package
%s

# Execute initialization script
%s

echo "Starting startup command script"
%s
EOF
		# Set script execution permissions
		chmod +x /app/init/startup.sh
		
		# Execute startup command script
		/app/init/startup.sh
	`,
		codepkgInstallScript, initScript, command)

	imgPms := &imageParams{
		image:       imgAddress,
		port:        port,
		command:     []string{"/bin/sh"},
		commandArgs: []string{"-c", startupScript},
	}

	return imgPms, nil
}

func (cd *ContainerBiz) getSupergatewayImage() string {
	return "ccr.ccs.tencentyun.com/itqm-private/supergateway:3.2.0-uvx"
}

// ContainerScaleParams 容器缩放参数
type ContainerScaleParams struct {
	InstanceID string
	Replicas   int32
}

// ContainerScaleResult 容器缩放结果
type ContainerScaleResult struct {
	Message string
}

// ContainerLogsParams 容器日志参数
type ContainerLogsParams struct {
	InstanceID string
	Lines      int64
}

// ContainerRestartResult 容器重启结果
type ContainerRestartResult struct {
	ContainerName string
	Message       string
}

// ScaleContainerToZero 将容器副本数缩放为0
func (cd *ContainerBiz) ScaleContainerToZero(instance *model.McpInstance) (*ContainerScaleResult, error) {
	// 1. 根据 instanceID 获取实例配置
	instance, err := mysql.McpInstanceRepo.FindByInstanceIDAndAccessType(
		context.Background(),
		instance.InstanceID,
		model.AccessTypeHosting, // 托管模式才需要缩放容器
	)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceNotHostingMode)+": %w", err)
	}
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// 获取容器管理器和服务管理器
	containerManager := entry.GetContainerManager()

	// 根据运行时类型选择缩放策略
	if instance.ContainerName != "" {
		// 获取运行时类型
		runtimeType := entry.GetRuntimeType()

		if runtimeType == container.RuntimeKubernetes {
			// Kubernetes: 设置副本数为0
			e1 := containerManager.Scale(cd.ctx, instance.ContainerName, 0)
			if e1 != nil {
				return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero)+": %w", e1)
			}
		} else {
			// Docker: 删除容器
			e2 := containerManager.Delete(cd.ctx, instance.ContainerName)
			if e2 != nil {
				return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeDeleteContainerFailure)+": %w", e2)
			}
		}
	}

	// 更新实例状态
	instance.Status = model.InstanceStatusInactive
	instance.ContainerIsReady = false
	instance.ContainerStatus = model.ContainerStatusManualStop
	instance.ContainerLastMessage = i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero)
	err = mysql.McpInstanceRepo.Update(cd.ctx, instance)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeUpdateInstanceFailure)+": %w", err)
	}

	return &ContainerScaleResult{Message: i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero)}, nil
}

// GetContainerLogs 获取容器日志
func (cd *ContainerBiz) GetContainerLogs(params ContainerLogsParams) (string, error) {
	// 1. 根据 instanceID 获取实例配置
	instance, err := mysql.McpInstanceRepo.FindByInstanceIDAndAccessType(
		context.Background(),
		params.InstanceID,
		model.AccessTypeHosting, // 托管模式才需要获取容器日志
	)
	if err != nil {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceNotHostingMode)+": %w", err)
	}
	if len(instance.ContainerName) <= 0 {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// 设置默认行数
	lines := params.Lines
	if lines <= 0 {
		lines = 100
	}

	// 获取容器日志
	logs, err := entry.GetContainerManager().GetLogs(cd.ctx, instance.ContainerName, lines)
	if err != nil {
		return "", fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetContainerLogsFailure)+": %w", err)
	}

	return logs, nil
}

// RestartContainer 重启容器业务逻辑
func (cd *ContainerBiz) RestartContainer(instance *model.McpInstance) (*ContainerRestartResult, error) {
	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure)+": %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}

	// 解析容器创建选项
	var containerOptions container.ContainerCreateOptions
	if len(instance.ContainerCreateOptions) > 0 {
		if e2 := json.Unmarshal(instance.ContainerCreateOptions, &containerOptions); e2 != nil {
			return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeParseContainerOptionsFailure)+": %w", e2)
		}
	} else {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeMissingContainerOptions))
	}

	// 调用容器管理器的重启方法
	err = entry.GetContainerManager().Restart(cd.ctx, containerOptions)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeRestartContainerFailure)+": %w", err)
	}

	// 获取 service
	err = entry.GetServiceManager().Restart(cd.ctx, containerOptions)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeRestartContainerFailure)+": %w", err)
	}

	return &ContainerRestartResult{
		ContainerName: instance.ContainerName,
		Message:       i18n.FormatWithContext(cd.ctx, i18n.CodeRestartContainerSuccess),
	}, nil
}

// createDownloadLink 创建下载链接
func (cd *ContainerBiz) createDownloadLink(downloadLinkPath string) string {
	mcpMarketSvc := config.GlobalConfig.Services.McpMarket
	if mcpMarketSvc == nil {
		return ""
	}
	return fmt.Sprintf("http://%s:%d/%s/%s",
		mcpMarketSvc.Host,
		mcpMarketSvc.Port,
		strings.TrimPrefix(common.GetMarketRoutePrefix(), "/"),
		strings.TrimPrefix(downloadLinkPath, "/"))
}

func (cd *ContainerBiz) generateDownloadZip(ctx context.Context, codePackage *model.McpCodePackage) (string, string, error) {
	packageManager := codepackage.NewCodePackageManager(&config.GlobalConfig.Code, config.GlobalConfig.Storage.CodePath)
	absPackagePath, err := packageManager.ToAbsolutePath(codePackage.PackagePath)
	if err != nil {
		logger.Error("Failed to convert to absolute path", zap.String("relativePath", codePackage.PackagePath), zap.Error(err))
		return "", "", fmt.Errorf("invalid package path")
	}
	// 将相对路径转换为绝对路径
	absExtractedPath, err := packageManager.ToAbsolutePath(codePackage.ExtractedPath)
	if err != nil {
		logger.Error("Failed to convert to absolute path", zap.String("relativePath", codePackage.ExtractedPath), zap.Error(err))
		return "", "", fmt.Errorf("invalid package path")
	}

	// codePackage.OriginalName 获取文件名，不包括后缀
	zipFilePath := fmt.Sprintf("%s/dl-%s.zip", absPackagePath, strings.TrimSuffix(codePackage.OriginalName, filepath.Ext(codePackage.OriginalName)))

	// 创建压缩包
	if err := utils.CreatePackageZip(absExtractedPath, zipFilePath); err != nil {
		logger.Error("Failed to generate package zip", zap.String("packageId", codePackage.PackageID), zap.Error(err))
		return "", "", fmt.Errorf("failed to generate download package: %v", err)
	}

	return absExtractedPath, zipFilePath, nil
}

// volumeMountFromPb 将 pb 卷挂载转换为本地结构
func (cd *ContainerBiz) volumeMountFromPb(vm *instancepb.VolumeMount) k8s.UnifiedMount {
	unifiedMount := k8s.UnifiedMount{
		Type:      k8s.MountType(vm.Type),
		MountPath: vm.MountPath,
		ReadOnly:  vm.ReadOnly,
		SubPath:   vm.SubPath,
		NodeName:  vm.NodeName,
		HostPath:  vm.HostPath,
		PVCName:   vm.PvcName,
	}
	return unifiedMount
}

// generateCodePkgScript 生成代码包启动脚本
func (cd *ContainerBiz) generateCodePkgInstallScript(packageId string) (string, error) {
	codepkgInstallScript := ""
	// 查找代码包
	codePackage, err := mysql.McpCodePackageRepo.FindByPackageID(cd.ctx, packageId)
	if err != nil {
		return codepkgInstallScript, fmt.Errorf(i18n.FormatWithContext(cd.ctx, i18n.CodeFailedToFindCodePackage)+": %w", err)
	}
	// ext := codePackage.PackageType

	downloadLinkPath := fmt.Sprintf("/code/download/%s", packageId)
	pkgLink := cd.createDownloadLink(downloadLinkPath)
	if codePackage == nil {
		return codepkgInstallScript, fmt.Errorf("code package is nil")
	}
	// 构建下载和解压ZIP包的命令
	if len(pkgLink) > 0 {
		codepkgInstallScript = fmt.Sprintf(`
		# Download and extract ZIP package
		echo "[$(date)] Starting to download package: %s"
		echo "mkdir -p /app/codepkg"
		mkdir -p /app/codepkg
		cd /tmp
		wget -O package.zip "%s" || curl -L -o package.zip "%s"
		echo "[$(date)] Package download completed, starting extraction to /app/codepkg"
		echo "unzip -o package.zip -d /app/codepkg"
		unzip -o package.zip -d /app/codepkg
		ls -al /app/codepkg
		echo "[$(date)] End Download and Extract"
		cd /app
		`, pkgLink, pkgLink, pkgLink)
	}
	return codepkgInstallScript, nil
}

// GetRuntimeEntry 获取环境的运行时入口
func (ed *ContainerBiz) GetRuntimeEntry(ctx context.Context, environmentID uint) (*container.Entry, error) {
	// 根据环境ID获取环境信息
	environment, err := GEnvironmentBiz.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeGetEnvironmentInfoFailure)+": %w", err)
	}

	// 根据环境类型创建不同的运行时配置
	switch environment.Environment {
	case model.McpEnvironmentKubernetes:
		// 创建Kubernetes容器运行时入口
		cfg, err := ed.getKubernetesRuntimeConfig(ctx, environment)
		if err != nil {
			return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeGetK8sRuntimeEntryFailure)+": %w", err)
		}
		// 创建Kubernetes容器运行时入口
		return container.NewEntry(cfg)
	case model.McpEnvironmentDocker:
		// return ed.getDockerRuntimeConfig(ctx, environment)
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeDockerEnvironmentNotSupported))
	default:
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeUnsupportedEnvironmentType))
	}
}

// getKubernetesRuntimeConfig 获取Kubernetes环境的运行时配置
func (ed *ContainerBiz) getKubernetesRuntimeConfig(ctx context.Context, environment *model.McpEnvironment) (container.Config, error) {
	// 创建Kubernetes容器运行时配置
	return container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  environment.Namespace,
		Kubeconfig: common.SetKubeConfig([]byte(environment.Config)),
	}, nil
}

// BuildContainerOptions 构建容器创建选项
func (cd *ContainerBiz) BuildContainerOptions(ctx context.Context, instanceID string, mcpProtocol model.McpProtocol, mcpServices string, packageId string, port int32, initScript string, command string, imgAddress string,
	evs map[string]string, vms []*instancepb.VolumeMount, startupTimeout int32, runningTimeout int32) (*container.ContainerCreateOptions, error) {
	var err error
	containerName := cd.generateContainerName(instanceID)
	serviceName := cd.generateServiceName(instanceID)
	// 代码包下载链接生成
	codepkgInstallScript := ""
	if packageId != "" {
		// 生成代码包安装脚本
		var e1 error
		codepkgInstallScript, e1 = cd.generateCodePkgInstallScript(packageId)
		if e1 != nil {
			return nil, fmt.Errorf("failed to generate code package install script: %w", e1)
		}
	}

	imgPms := &imageParams{}
	if mcpProtocol == model.McpProtocolSSE || mcpProtocol == model.McpProtocolStreamableHttp {
		// 生成镜像配置
		imgPms, err = cd.getMcpHostingImageCfgForSSEAndSteamableHttp(imgAddress, port, initScript, command, codepkgInstallScript)
		if err != nil {
			return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
		}
	} else {
		// 生成镜像配置
		imgPms, err = cd.getMcpHostingImageCfg(imgAddress, port, command, codepkgInstallScript, mcpServices)
		if err != nil {
			return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
		}
	}
	if imgPms.image == "" || len(imgPms.commandArgs) == 0 || imgPms.port == 0 {
		return nil, fmt.Errorf("build container options failed: image or command or port is empty")
	}

	// 设置环境变量
	envVars := make(map[string]string)
	envVars["MCP_INSTANCE_ID"] = instanceID
	envVars["MCP_PORT"] = fmt.Sprintf("%d", imgPms.port)
	envVars["NODE_ENV"] = "production"
	for k, v := range evs {
		envVars[k] = v
	}

	// 设置卷挂载配置（亲和性判断逻辑转移到Create方法中）
	mounts := []k8s.UnifiedMount{}
	if len(vms) > 0 {
		for _, vm := range vms {
			mounts = append(mounts, cd.volumeMountFromPb(vm))
		}
	}

	// 设置标签
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	if startupTimeout > 0 {
		labels["mcp.startup.timeout"] = fmt.Sprintf("%d", startupTimeout)
	}
	if runningTimeout > 0 {
		labels["mcp.running.timeout"] = fmt.Sprintf("%d", runningTimeout)
	}

	// 8. 构建容器创建选项
	containerOptions := container.ContainerCreateOptions{
		ImageName:     imgPms.image,
		ContainerName: containerName,
		ServiceName:   serviceName,
		Port:          imgPms.port,
		Command:       imgPms.command,
		CommandArgs:   imgPms.commandArgs,
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       envVars,
		Mounts:        mounts,
		WorkingDir:    "/app",
	}

	// 创建Kubernetes容器运行时配置
	return &containerOptions, nil
}
