package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"
)

// ResourceBiz 资源数据处理层
type ResourceBiz struct {
	ctx context.Context
}

// GResourceBiz 全局资源数据处理层实例
var GResourceBiz *ResourceBiz

func init() {
	GResourceBiz = NewResourceBiz(context.Background())
}

// NewResourceBiz 创建资源数据处理实例
func NewResourceBiz(ctx context.Context) *ResourceBiz {
	return &ResourceBiz{
		ctx: ctx,
	}
}

// ListPVCs 根据环境ID获取PVC列表
func (biz *ResourceBiz) ListPVCs(environmentID uint) ([]k8s.PVCInfo, error) {
	// 获取环境配置
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %s", err.Error())
	}

	// 调用 Volume 管理器获取 PVC 列表, 指定环境命名空间
	return k8sEntry.Volume.ListPVCs(k8sEntry.Namespace)
}

// ListNodes 根据环境ID获取节点列表
func (biz *ResourceBiz) ListNodes(environmentID uint) ([]k8s.NodeInfo, error) {
	// 获取环境配置
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %s", err.Error())
	}

	// 调用 Node 管理器获取节点列表
	return k8sEntry.Node.ListNodes()
}

// ListStorageClasses 根据环境ID获取存储类列表
func (biz *ResourceBiz) ListStorageClasses(environmentID uint) ([]k8s.StorageClassInfo, error) {
	// 获取环境配置
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %s", err.Error())
	}

	// 调用 Volume 管理器获取存储类列表
	return k8sEntry.Volume.ListStorageClasses()
}

// CreateHostPathPVC 根据环境ID创建基于主机路径的PVC
func (biz *ResourceBiz) CreateHostPathPVC(environmentID uint, name, hostPath, nodeName, accessMode, storageClass string, storageSize int32) (*k8s.PVCInfo, error) {
	// 获取环境配置
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %s", err.Error())
	}

	// 调用 Volume 管理器创建基于主机路径的 PVC 和 PV
	// 使用空标签映射，因为标签已不再需要
	return k8sEntry.Volume.CreateHostPathPVCWithPV(name, hostPath, nodeName, accessMode, storageSize, storageClass, nil)
}

// CreatePVC 根据环境ID创建普通PVC
func (biz *ResourceBiz) CreatePVC(environmentID uint, name, nodeName, accessMode, storageClass string, storageSize int32, labels map[string]string) (*k8s.PVCInfo, error) {
	// 获取环境配置
	k8sEntry, err := biz.getK8sEntryByEnvironmentID(environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %s", err.Error())
	}

	// 调用 Volume 管理器创建普通 PVC
	return k8sEntry.Volume.CreatePVCWithParams(name, nodeName, accessMode, storageClass, storageSize, labels)
}

// getK8sEntryByEnvironmentID 根据环境ID获取K8s Entry
func (biz *ResourceBiz) getK8sEntryByEnvironmentID(environmentID uint) (*k8s.Entry, error) {
	// 获取环境信息
	environment, err := GEnvironmentBiz.GetEnvironment(biz.ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("获取环境信息失败: %s", err.Error())
	}

	// 验证环境类型
	if environment.Environment != model.McpEnvironmentKubernetes {
		return nil, fmt.Errorf("环境类型不是Kubernetes，无法查询K8s资源")
	}

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
		return nil, fmt.Errorf("创建容器运行时入口失败: %s", err.Error())
	}

	// 检查是否为Kubernetes运行时
	if !entry.IsKubernetes() {
		return nil, fmt.Errorf("运行时类型错误，预期Kubernetes运行时")
	}

	// 获取K8s入口
	k8sRuntime := entry.GetK8sRuntime()
	if k8sRuntime == nil {
		return nil, fmt.Errorf("获取K8s入口失败")
	}
	if k8sEntry := k8sRuntime.Entry; k8sEntry != nil {
		return k8sEntry, nil
	}

	return nil, fmt.Errorf("K8s运行时类型断言失败")
}
