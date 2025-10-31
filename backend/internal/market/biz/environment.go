package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/api/market/mcp_environment"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"

	"gopkg.in/yaml.v3"
)

// EnvironmentBiz 环境数据访问层
type EnvironmentBiz struct {
	ctx  context.Context
	repo *mysql.McpEnvironmentRepository
}

var GEnvironmentBiz *EnvironmentBiz

func init() {
	GEnvironmentBiz = NewEnvironmentBiz(context.Background())
}

// NewEnvironmentBiz 创建环境数据访问层实例
func NewEnvironmentBiz(ctx context.Context) *EnvironmentBiz {
	return &EnvironmentBiz{
		ctx:  ctx,
		repo: mysql.McpEnvironmentRepo,
	}
}

// CreateEnvironment 创建环境
func (biz *EnvironmentBiz) CreateEnvironment(ctx context.Context, environment *model.McpEnvironment) error {
	return biz.repo.Create(ctx, environment)
}

// UpdateEnvironment 更新环境
func (biz *EnvironmentBiz) UpdateEnvironment(ctx context.Context, environment *model.McpEnvironment) error {
	return biz.repo.Update(ctx, environment)
}

// DeleteEnvironment 删除环境
func (biz *EnvironmentBiz) DeleteEnvironment(ctx context.Context, id uint) error {
	// Check if there are templates associated with this environment
	templates, err := GTemplateBiz.GetTemplatesByEnvironmentID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check templates: %w", err)
	}
	if len(templates) > 0 {
		return fmt.Errorf("cannot delete environment: %d templates are still associated with this environment", len(templates))
	}

	// Check if there are instances associated with this environment
	instances, err := GInstanceBiz.GetInstancesByEnvironmentID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check instances: %w", err)
	}
	if len(instances) > 0 {
		return fmt.Errorf("cannot delete environment: %d instances are still associated with this environment", len(instances))
	}

	return biz.repo.Delete(ctx, id)
}

// GetEnvironment 根据ID获取环境
func (biz *EnvironmentBiz) GetEnvironment(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	return biz.repo.FindByID(ctx, id)
}

// GetEnvironmentByName 根据名称获取环境
func (biz *EnvironmentBiz) GetEnvironmentByName(ctx context.Context, name string) (*model.McpEnvironment, error) {
	return biz.repo.FindByName(ctx, name)
}

// ListEnvironments 获取所有环境列表
func (biz *EnvironmentBiz) ListEnvironments(ctx context.Context) ([]*model.McpEnvironment, error) {
	return biz.repo.FindAll(ctx)
}

// ListEnvironmentsByType 根据环境类型获取环境列表
func (biz *EnvironmentBiz) ListEnvironmentsByType(ctx context.Context, environmentType model.McpEnvironmentType) ([]*model.McpEnvironment, error) {
	return biz.repo.FindByEnvironment(ctx, environmentType)
}

// GetDeletedEnvironment 根据ID获取已删除的环境
func (biz *EnvironmentBiz) GetDeletedEnvironment(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	return biz.repo.FindDeletedByID(ctx, id)
}

// ListAllEnvironments 获取所有环境列表（包括已删除）
func (biz *EnvironmentBiz) ListAllEnvironments(ctx context.Context) ([]*model.McpEnvironment, error) {
	return biz.repo.FindAllWithDeleted(ctx)
}

// RestoreEnvironment 恢复已删除的环境
func (biz *EnvironmentBiz) RestoreEnvironment(ctx context.Context, id uint) error {
	return biz.repo.RestoreEnvironment(ctx, id)
}

// TestEnvironmentConnectivity 执行环境连通性测试
func (biz *EnvironmentBiz) TestEnvironmentConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// 根据环境类型执行不同的连通性测试
	switch environment.Environment {
	case model.McpEnvironmentKubernetes:
		return biz.testKubernetesConnectivity(ctx, environment)
	case model.McpEnvironmentDocker:
		return biz.testDockerConnectivity(ctx, environment)
	default:
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "不支持的环境类型",
		}, nil
	}
}

// testKubernetesConnectivity 测试Kubernetes连通性
func (biz *EnvironmentBiz) testKubernetesConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// 创建容器运行时配置
	config := container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  environment.Namespace,
		Kubeconfig: common.SetKubeConfig([]byte(environment.Config)),
		Network:    "bridge", // 默认网络配置
	}

	// 创建容器运行时入口
	entry, err := container.NewEntry(config)
	if err != nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Kubernetes客户端初始化失败",
		}, nil
	}

	// 检查是否为Kubernetes运行时
	if !entry.IsKubernetes() {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "运行时类型错误",
		}, nil
	}

	// 获取K8s入口
	k8sRuntime := entry.GetK8sRuntime()
	if k8sRuntime == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Kubernetes客户端获取失败",
		}, nil
	}

	// 测试连接 - 尝试获取节点信息
	containerManager := entry.GetContainerManager()
	if containerManager == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "容器管理器获取失败",
		}, nil
	}

	return &mcp_environment.TestConnectivityResponse{
		Success: true,
		Message: "Kubernetes连接测试成功",
	}, nil
}

// testDockerConnectivity 测试Docker连通性
func (biz *EnvironmentBiz) testDockerConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// 创建容器运行时配置
	config := container.Config{
		Runtime: container.RuntimeDocker,
		Network: "bridge", // 默认Docker网络
	}

	// 创建容器运行时入口
	entry, err := container.NewEntry(config)
	if err != nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Docker客户端初始化失败",
		}, nil
	}

	// 检查是否为Docker运行时
	if !entry.IsDocker() {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "运行时类型错误",
		}, nil
	}

	// 获取容器管理器进行连通性测试
	containerManager := entry.GetContainerManager()
	if containerManager == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Docker容器管理器未初始化",
		}, nil
	}

	details := "Docker连接测试通过"
	if environment.Config != "" {
		details += fmt.Sprintf("，使用配置: %s", environment.Config)
	}

	return &mcp_environment.TestConnectivityResponse{
		Success: true,
		Message: i18n.FormatWithContext(ctx, i18n.CodeDockerConnectionSuccess),
	}, nil
}

// ListNamespaces 获取命名空间列表（仅支持Kubernetes环境）
func (biz *EnvironmentBiz) ListNamespaces(ctx context.Context, config string, environmentType model.McpEnvironmentType) ([]string, error) {
	if environmentType != model.McpEnvironmentKubernetes {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeOnlyK8sSupportNamespace))
	}

	// 验证 config 数据是否为有效的 YAML 格式
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(config), &yamlData); err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigFormatError)+": %w", err)
	}

	// 验证是否为有效的 kubeconfig 结构
	var kubeconfigStruct map[string]interface{}
	if err := yaml.Unmarshal([]byte(config), &kubeconfigStruct); err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigParseFailure)+": %w", err)
	}

	// 检查必要的 kubeconfig 字段
	if _, exists := kubeconfigStruct["apiVersion"]; !exists {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigMissingField, "apiVersion"))
	}
	if _, exists := kubeconfigStruct["kind"]; !exists {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigMissingField, "kind"))
	}
	if _, exists := kubeconfigStruct["clusters"]; !exists {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigMissingField, "clusters"))
	}
	if _, exists := kubeconfigStruct["contexts"]; !exists {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigMissingField, "contexts"))
	}
	if _, exists := kubeconfigStruct["users"]; !exists {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigMissingField, "users"))
	}

	// kubeconfigStruct 转换为 YAML 字符串
	configYAML, err := yaml.Marshal(kubeconfigStruct)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigYamlConversionFailure)+": %w", err)
	}

	// 使用修复后的 SetKubeConfig 函数
	kubeconfig := common.SetKubeConfig([]byte(configYAML))
	if kubeconfig == nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeKubeconfigConversionFailure))
	}

	// 创建容器运行时配置
	containerConfig := container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  "default", // 使用默认命名空间来连接集群
		Kubeconfig: kubeconfig,
		Network:    "bridge",
	}

	// 创建容器运行时入口
	entry, err := container.NewEntry(containerConfig)
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeK8sClientInitFailure)+": %w", err)
	}

	// 检查是否为Kubernetes运行时
	if !entry.IsKubernetes() {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeRuntimeTypeError))
	}

	// 获取K8s入口
	namespaces, err := entry.ListNamespaces()
	if err != nil {
		return nil, fmt.Errorf(i18n.FormatWithContext(ctx, i18n.CodeListNamespacesFailure)+": %w", err)
	}
	return namespaces, nil
}
