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

// EnvironmentBiz environment data access layer
type EnvironmentBiz struct {
	ctx  context.Context
	repo *mysql.McpEnvironmentRepository
}

var GEnvironmentBiz *EnvironmentBiz

func init() {
	GEnvironmentBiz = NewEnvironmentBiz(context.Background())
}

// NewEnvironmentBiz create environment data access layer instance
func NewEnvironmentBiz(ctx context.Context) *EnvironmentBiz {
	return &EnvironmentBiz{
		ctx:  ctx,
		repo: mysql.McpEnvironmentRepo,
	}
}

// CreateEnvironment create environment
func (biz *EnvironmentBiz) CreateEnvironment(ctx context.Context, environment *model.McpEnvironment) error {
	return biz.repo.Create(ctx, environment)
}

// UpdateEnvironment update environment
func (biz *EnvironmentBiz) UpdateEnvironment(ctx context.Context, environment *model.McpEnvironment) error {
	return biz.repo.Update(ctx, environment)
}

// DeleteEnvironment delete environment
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

// GetEnvironment get environment by ID
func (biz *EnvironmentBiz) GetEnvironment(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	return biz.repo.FindByID(ctx, id)
}

// GetEnvironmentByName get environment by name
func (biz *EnvironmentBiz) GetEnvironmentByName(ctx context.Context, name string) (*model.McpEnvironment, error) {
	return biz.repo.FindByName(ctx, name)
}

// ListEnvironments get all environment list
func (biz *EnvironmentBiz) ListEnvironments(ctx context.Context) ([]*model.McpEnvironment, error) {
	return biz.repo.FindAll(ctx)
}

// ListEnvironmentsByType get environment list by environment type
func (biz *EnvironmentBiz) ListEnvironmentsByType(ctx context.Context, environmentType model.McpEnvironmentType) ([]*model.McpEnvironment, error) {
	return biz.repo.FindByEnvironment(ctx, environmentType)
}

// GetDeletedEnvironment get deleted environment by ID
func (biz *EnvironmentBiz) GetDeletedEnvironment(ctx context.Context, id uint) (*model.McpEnvironment, error) {
	return biz.repo.FindDeletedByID(ctx, id)
}

// ListAllEnvironments get all environment list (including deleted)
func (biz *EnvironmentBiz) ListAllEnvironments(ctx context.Context) ([]*model.McpEnvironment, error) {
	return biz.repo.FindAllWithDeleted(ctx)
}

// RestoreEnvironment restore deleted environment
func (biz *EnvironmentBiz) RestoreEnvironment(ctx context.Context, id uint) error {
	return biz.repo.RestoreEnvironment(ctx, id)
}

// TestEnvironmentConnectivity perform environment connectivity test
func (biz *EnvironmentBiz) TestEnvironmentConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// Execute different connectivity tests based on environment type
	switch environment.Environment {
	case model.McpEnvironmentKubernetes:
		return biz.testKubernetesConnectivity(ctx, environment)
	case model.McpEnvironmentDocker:
		return biz.testDockerConnectivity(ctx, environment)
	default:
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "unsupported environment type",
		}, nil
	}
}

// testKubernetesConnectivity test Kubernetes connectivity
func (biz *EnvironmentBiz) testKubernetesConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// Create container runtime configuration
	config := container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  environment.Namespace,
		Kubeconfig: common.SetKubeConfig([]byte(environment.Config)),
		Network:    "bridge", // Default network configuration
	}

	// Create container runtime entry
	entry, err := container.NewEntry(config)
	if err != nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Kubernetes client initialization failed",
		}, nil
	}

	// Check if it's Kubernetes runtime
	if !entry.IsKubernetes() {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "runtime type error",
		}, nil
	}

	// Get K8s entry
	k8sRuntime := entry.GetK8sRuntime()
	if k8sRuntime == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Kubernetes client acquisition failed",
		}, nil
	}

	// Test connection - try to get node information
	containerManager := entry.GetContainerManager()
	if containerManager == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "container manager acquisition failed",
		}, nil
	}

	return &mcp_environment.TestConnectivityResponse{
		Success: true,
		Message: "Kubernetes connection test successful",
	}, nil
}

// testDockerConnectivity test Docker connectivity
func (biz *EnvironmentBiz) testDockerConnectivity(ctx context.Context, environment *model.McpEnvironment) (*mcp_environment.TestConnectivityResponse, error) {
	// Create container runtime configuration
	config := container.Config{
		Runtime: container.RuntimeDocker,
		Network: "bridge", // Default Docker network
	}

	// Create container runtime entry
	entry, err := container.NewEntry(config)
	if err != nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Docker client initialization failed",
		}, nil
	}

	// Check if it's Docker runtime
	if !entry.IsDocker() {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "runtime type error",
		}, nil
	}

	// Get container manager for connectivity test
	containerManager := entry.GetContainerManager()
	if containerManager == nil {
		return &mcp_environment.TestConnectivityResponse{
			Success: false,
			Message: "Docker container manager not initialized",
		}, nil
	}

	details := "Docker connection test successful"
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
