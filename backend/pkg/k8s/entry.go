package k8s

import (
	"k8s.io/client-go/rest"
)

// Entry 统一入口，支持通过接口快速调用 Pod、Service、Volume 管理能力
type Entry struct {
	kubeconfig *rest.Config
	Namespace  string
	Client     *Client
	Pod        *PodManager
	Service    *ServiceManager
	Volume     *VolumeManager
	Node       *NodeManager
}

var K8sEntry *Entry

// NewEntry 初始化 Entry，注入 PodManager、ServiceManager 和 VolumeManager
func NewEntry(kubeconfig *rest.Config, namespace string) (*Entry, error) {
	client, err := NewClient(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}
	return &Entry{
		kubeconfig: kubeconfig,
		Namespace:  namespace,
		Client:     client,
		Pod:        client.Pod(),
		Service:    client.Service(),
		Volume:     client.Volume(),
		Node:       client.Node(),
	}, nil
}
