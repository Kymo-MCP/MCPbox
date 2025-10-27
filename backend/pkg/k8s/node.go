package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeManager 负责节点相关操作
type NodeManager struct {
	client *Client
}

// NodeInfo 节点信息结构体
type NodeInfo struct {
	Name              string            `json:"name"`              // 节点名称
	Status            string            `json:"status"`            // 节点状态 (Ready, NotReady, Unknown)
	Roles             []string          `json:"roles"`             // 节点角色 (master, worker)
	Version           string            `json:"version"`           // Kubelet 版本
	InternalIP        string            `json:"internalIP"`        // 内部IP
	ExternalIP        string            `json:"externalIP"`        // 外部IP
	OperatingSystem   string            `json:"operatingSystem"`   // 操作系统
	Architecture      string            `json:"architecture"`      // 架构
	KernelVersion     string            `json:"kernelVersion"`     // 内核版本
	ContainerRuntime  string            `json:"containerRuntime"`  // 容器运行时
	AllocatableMemory string            `json:"allocatableMemory"` // 可分配内存
	AllocatableCPU    string            `json:"allocatableCpu"`    // 可分配CPU
	AllocatablePods   string            `json:"allocatablePods"`   // 可分配Pod数量
	Labels            map[string]string `json:"labels"`            // 标签
	Annotations       map[string]string `json:"annotations"`       // 注解
	CreationTime      string            `json:"creationTime"`      // 创建时间
}

// ListNodes 获取所有节点列表
func (nm *NodeManager) ListNodes() ([]NodeInfo, error) {
	log.Printf("正在查询集群中的所有节点...")

	nodeList, err := nm.client.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询节点列表失败: %w", err)
	}

	log.Printf("成功查询到 %d 个节点", len(nodeList.Items))

	var nodeInfos []NodeInfo
	for _, node := range nodeList.Items {
		// 获取节点状态
		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// 获取节点角色
		var roles []string
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
				roles = append(roles, "master")
			} else if label == "node-role.kubernetes.io/worker" {
				roles = append(roles, "worker")
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker") // 默认为worker节点
		}

		// 获取IP地址
		var internalIP, externalIP string
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalIP:
				internalIP = address.Address
			case corev1.NodeExternalIP:
				externalIP = address.Address
			}
		}

		// 构建节点信息
		nodeInfo := NodeInfo{
			Name:              node.Name,
			Status:            status,
			Roles:             roles,
			Version:           node.Status.NodeInfo.KubeletVersion,
			InternalIP:        internalIP,
			ExternalIP:        externalIP,
			OperatingSystem:   node.Status.NodeInfo.OperatingSystem,
			Architecture:      node.Status.NodeInfo.Architecture,
			KernelVersion:     node.Status.NodeInfo.KernelVersion,
			ContainerRuntime:  node.Status.NodeInfo.ContainerRuntimeVersion,
			AllocatableMemory: node.Status.Allocatable.Memory().String(),
			AllocatableCPU:    node.Status.Allocatable.Cpu().String(),
			AllocatablePods:   node.Status.Allocatable.Pods().String(),
			Labels:            node.Labels,
			Annotations:       node.Annotations,
			CreationTime:      node.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	return nodeInfos, nil
}

// GetNode 获取指定节点信息
func (nm *NodeManager) GetNode(name string) (*NodeInfo, error) {
	log.Printf("正在查询节点 '%s'...", name)

	node, err := nm.client.clientset.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询节点 '%s' 失败: %w", name, err)
	}

	// 获取节点状态
	status := "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	// 获取节点角色
	var roles []string
	for label := range node.Labels {
		if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
			roles = append(roles, "master")
		} else if label == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker") // 默认为worker节点
	}

	// 获取IP地址
	var internalIP, externalIP string
	for _, address := range node.Status.Addresses {
		switch address.Type {
		case corev1.NodeInternalIP:
			internalIP = address.Address
		case corev1.NodeExternalIP:
			externalIP = address.Address
		}
	}

	// 构建节点信息
	nodeInfo := &NodeInfo{
		Name:              node.Name,
		Status:            status,
		Roles:             roles,
		Version:           node.Status.NodeInfo.KubeletVersion,
		InternalIP:        internalIP,
		ExternalIP:        externalIP,
		OperatingSystem:   node.Status.NodeInfo.OperatingSystem,
		Architecture:      node.Status.NodeInfo.Architecture,
		KernelVersion:     node.Status.NodeInfo.KernelVersion,
		ContainerRuntime:  node.Status.NodeInfo.ContainerRuntimeVersion,
		AllocatableMemory: node.Status.Allocatable.Memory().String(),
		AllocatableCPU:    node.Status.Allocatable.Cpu().String(),
		AllocatablePods:   node.Status.Allocatable.Pods().String(),
		Labels:            node.Labels,
		Annotations:       node.Annotations,
		CreationTime:      node.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return nodeInfo, nil
}
