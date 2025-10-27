package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client 封装 Kubernetes 客户端及资源管理入口
// 仅负责 clientset 初始化和 namespace 管理
// 具体资源操作通过子管理器（如 PodManager）实现

type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

// 获取 Pod 管理器，支持创建、删除、等待就绪、获取状态等操作
func (c *Client) Pod() *PodManager {
	return &PodManager{client: c}
}

// 获取 Deployment 管理器，支持 Deployment 的创建、删除、查询等操作
func (c *Client) Deployment() *DeploymentManager {
	return &DeploymentManager{client: c}
}

// 获取 Service 管理器，支持 Service 的创建、删除、查询等操作
func (c *Client) Service() *ServiceManager {
	return &ServiceManager{client: c}
}

// 获取 Volume 管理器，支持 PVC 和 StorageClass 的查询、创建、删除等操作
func (c *Client) Volume() *VolumeManager {
	return &VolumeManager{client: c}
}

// 获取 Node 管理器，支持节点的查询等操作
func (c *Client) Node() *NodeManager {
	return &NodeManager{client: c}
}

// GetNamespace 获取当前命名空间
func (c *Client) GetNamespace() string {
	return c.namespace
}

// CheckNamespaceExists 检查指定命名空间是否存在
func (c *Client) CheckNamespaceExists(namespace string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	return err
}

// ListNamespaces 获取命名空间列表
func (c *Client) ListNamespaces() ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	return namespaceNames, nil
}

// NewClient 通过 kubeconfig 内容和 namespace 初始化 Client
func NewClient(config *rest.Config, namespace string) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// 判断 namespace 是否存在
	nsClient := clientset.CoreV1().Namespaces()
	_, err = nsClient.Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &Client{clientset: clientset, namespace: namespace}, nil
}
