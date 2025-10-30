package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceManager 负责 Service 相关操作
// 通过 Client 组合实现

type ServiceManager struct {
	client *Client
}

// Create 创建一个 Service
func (sm *ServiceManager) Create(svc *corev1.Service) (*corev1.Service, error) {
	if svc.Namespace == "" {
		svc.Namespace = sm.client.namespace
	}
	return sm.client.clientset.CoreV1().Services(sm.client.namespace).Create(context.Background(), svc, metav1.CreateOptions{})
}

// Delete 删除指定 Service
func (sm *ServiceManager) Delete(name string) error {
	return sm.client.clientset.CoreV1().Services(sm.client.namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// Get 获取 Service 详情
func (sm *ServiceManager) Get(name string) (*corev1.Service, error) {
	return sm.client.clientset.CoreV1().Services(sm.client.namespace).Get(context.Background(), name, metav1.GetOptions{})
}
