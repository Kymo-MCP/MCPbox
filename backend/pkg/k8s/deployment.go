package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// DeploymentManager 负责 Deployment 相关操作
type DeploymentManager struct {
	client *Client
}

// DeploymentCreateOptions Deployment 创建选项
type DeploymentCreateOptions struct {
	ImageName string `json:"imageName"`
	AppName   string `json:"appName"` // 应用名称，用作 Deployment 名称
	Namespace string `json:"namespace,omitempty"`
	Port      int32  `json:"port,omitempty"`
	Replicas  int32  `json:"replicas,omitempty"` // 副本数，默认为 1

	// 容器配置
	Command     []string          `json:"command,omitempty"`
	CommandArgs []string          `json:"commandArgs,omitempty"`
	EnvVars     map[string]string `json:"envVars,omitempty"`
	WorkingDir  string            `json:"workingDir,omitempty"`

	// 标签和选择器
	Labels map[string]string `json:"labels,omitempty"`

	// 重启策略（对于 Deployment 总是 Always）
	RestartPolicy string `json:"restartPolicy,omitempty"`

	// 卷挂载配置
	VolumeMounts []UnifiedMount `json:"volumeMounts,omitempty"`

	// 健康检查
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	LivenessProbe  *corev1.Probe `json:"livenessProbe,omitempty"`

	// 镜像拉取
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// 资源限制
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
}

// Create 创建 Deployment
func (dm *DeploymentManager) Create(options DeploymentCreateOptions) (string, error) {
	// 验证参数
	if err := dm.validateCreateOptions(options); err != nil {
		return "", err
	}

	// 设置默认值和资源优化
	if options.Replicas <= 0 {
		options.Replicas = 1
	} else {
		options.Replicas = int32(options.Replicas)
	}

	// 确定命名空间
	targetNamespace := dm.getTargetNamespace(options.Namespace)

	// 构建标签
	labels := dm.buildLabels(options)

	// 构建卷和卷挂载
	volumes, volumeMounts, err := dm.buildVolumes(options)
	if err != nil {
		return "", err
	}

	// 构建容器
	container := dm.buildContainer(options, volumeMounts)

	// 构建节点亲和性
	nodeAffinity, err := dm.buildAutoNodeAffinity(options, targetNamespace)
	if err != nil {
		return "", err
	}

	// 构建 Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.AppName,
			Namespace: targetNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &options.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers:       []corev1.Container{container},
					Volumes:          volumes,
					RestartPolicy:    corev1.RestartPolicyAlways, // Deployment 中总是 Always
					ImagePullSecrets: dm.buildImagePullSecrets(options.ImagePullSecrets),
				},
			},
		},
	}

	// 如果有节点亲和性，设置到 PodSpec 中
	if nodeAffinity != nil {
		deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: nodeAffinity,
		}
	}

	// 创建 Deployment
	createdDeployment, err := dm.client.clientset.AppsV1().Deployments(targetNamespace).Create(
		context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("创建 Deployment 失败: %w", err)
	}

	return createdDeployment.Name, nil
}

// Delete 删除 Deployment
func (dm *DeploymentManager) Delete(deploymentName string) error {
	// 设置级联删除策略，确保删除所有相关资源（ReplicaSet、Pod等）
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	return dm.client.clientset.AppsV1().Deployments(dm.client.namespace).Delete(
		context.Background(), deploymentName, deleteOptions)
}

// Get 获取 Deployment
func (dm *DeploymentManager) Get(deploymentName string) (*appsv1.Deployment, error) {
	return dm.client.clientset.AppsV1().Deployments(dm.client.namespace).Get(
		context.Background(), deploymentName, metav1.GetOptions{})
}

// Scale 设置 Deployment 副本数
func (dm *DeploymentManager) Scale(deploymentName string, replicas int32) error {
	// 获取当前 Deployment
	deployment, err := dm.Get(deploymentName)
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 更新副本数
	deployment.Spec.Replicas = &replicas

	// 更新 Deployment
	_, err = dm.client.clientset.AppsV1().Deployments(dm.client.namespace).Update(
		context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Deployment 副本数失败: %w", err)
	}

	return nil
}

// GetStatus 获取 Deployment 状态
func (dm *DeploymentManager) GetStatus(deploymentName string) (*appsv1.DeploymentStatus, error) {
	deployment, err := dm.Get(deploymentName)
	if err != nil {
		return nil, err
	}
	return &deployment.Status, nil
}

// IsReady 检查 Deployment 是否就绪
func (dm *DeploymentManager) IsReady(deploymentName string) (bool, error) {
	status, err := dm.GetStatus(deploymentName)
	if err != nil {
		return false, err
	}

	// 检查是否所有副本都已就绪
	return status.ReadyReplicas == status.Replicas && status.Replicas > 0, nil
}

// GetEvents 获取 Deployment 相关事件
func (dm *DeploymentManager) GetEvents(deploymentName string) ([]corev1.Event, error) {
	// 获取 Deployment 对象
	deployment, err := dm.Get(deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 构建字段选择器，获取与该 Deployment 相关的事件
	fieldSelector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.kind", "Deployment"),
		fields.OneTermEqualSelector("involvedObject.name", deploymentName),
		fields.OneTermEqualSelector("involvedObject.namespace", deployment.Namespace),
	)

	// 获取事件列表
	eventList, err := dm.client.clientset.CoreV1().Events(deployment.Namespace).List(
		context.Background(),
		metav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 事件失败: %w", err)
	}

	// 同时获取 ReplicaSet 相关事件
	replicaSets, err := dm.client.clientset.AppsV1().ReplicaSets(deployment.Namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
		},
	)
	if err == nil {
		for _, rs := range replicaSets.Items {
			// 检查 ReplicaSet 是否属于该 Deployment
			for _, ownerRef := range rs.OwnerReferences {
				if ownerRef.Kind == "Deployment" && ownerRef.Name == deploymentName {
					// 获取 ReplicaSet 事件
					rsFieldSelector := fields.AndSelectors(
						fields.OneTermEqualSelector("involvedObject.kind", "ReplicaSet"),
						fields.OneTermEqualSelector("involvedObject.name", rs.Name),
						fields.OneTermEqualSelector("involvedObject.namespace", rs.Namespace),
					)
					rsEvents, err := dm.client.clientset.CoreV1().Events(rs.Namespace).List(
						context.Background(),
						metav1.ListOptions{
							FieldSelector: rsFieldSelector.String(),
						},
					)
					if err == nil {
						eventList.Items = append(eventList.Items, rsEvents.Items...)
					}
					break
				}
			}
		}
	}

	// 获取 Pod 相关事件
	pods, err := dm.client.clientset.CoreV1().Pods(deployment.Namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
		},
	)
	if err == nil {
		for _, pod := range pods.Items {
			// 检查 Pod 是否属于该 Deployment（通过 ReplicaSet）
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "ReplicaSet" {
					// 获取 Pod 事件
					podFieldSelector := fields.AndSelectors(
						fields.OneTermEqualSelector("involvedObject.kind", "Pod"),
						fields.OneTermEqualSelector("involvedObject.name", pod.Name),
						fields.OneTermEqualSelector("involvedObject.namespace", pod.Namespace),
					)
					podEvents, err := dm.client.clientset.CoreV1().Events(pod.Namespace).List(
						context.Background(),
						metav1.ListOptions{
							FieldSelector: podFieldSelector.String(),
						},
					)
					if err == nil {
						eventList.Items = append(eventList.Items, podEvents.Items...)
					}
					break
				}
			}
		}
	}

	return eventList.Items, nil
}

// GetWarningEvents 获取 Deployment 相关警告事件
func (dm *DeploymentManager) GetWarningEvents(deploymentName string) ([]corev1.Event, error) {
	allEvents, err := dm.GetEvents(deploymentName)
	if err != nil {
		return nil, err
	}

	// 过滤出警告事件
	var warningEvents []corev1.Event
	for _, event := range allEvents {
		if event.Type == "Warning" {
			warningEvents = append(warningEvents, event)
		}
	}

	return warningEvents, nil
}

// GetPods 获取 Deployment 管理的 Pod 列表
func (dm *DeploymentManager) GetPods(deploymentName string) ([]corev1.Pod, error) {
	deployment, err := dm.Get(deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 通过标签选择器获取 Pod
	podList, err := dm.client.clientset.CoreV1().Pods(deployment.Namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	return podList.Items, nil
}

// GetPodIPs 获取 Deployment 管理的 Pod IP 列表
func (dm *DeploymentManager) GetPodIPs(deploymentName string) ([]string, error) {
	pods, err := dm.GetPods(deploymentName)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, pod := range pods {
		if pod.Status.PodIP != "" && pod.Status.Phase == corev1.PodRunning {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return ips, nil
}

// validateCreateOptions 验证创建选项
func (dm *DeploymentManager) validateCreateOptions(options DeploymentCreateOptions) error {
	if options.ImageName == "" {
		return fmt.Errorf("镜像名称不能为空")
	}
	if options.AppName == "" {
		return fmt.Errorf("应用名称不能为空")
	}
	return nil
}

// getTargetNamespace 获取目标命名空间
func (dm *DeploymentManager) getTargetNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}
	return dm.client.namespace
}

// buildLabels 构建标签
func (dm *DeploymentManager) buildLabels(options DeploymentCreateOptions) map[string]string {
	labels := map[string]string{
		"app": options.AppName,
	}

	// 添加自定义标签
	for k, v := range options.Labels {
		labels[k] = v
	}

	return labels
}

// buildContainer 构建容器配置
func (dm *DeploymentManager) buildContainer(options DeploymentCreateOptions, volumeMounts []corev1.VolumeMount) corev1.Container {
	container := corev1.Container{
		Name:         options.AppName,
		Image:        options.ImageName,
		VolumeMounts: volumeMounts,
	}

	// 设置命令
	if len(options.Command) > 0 {
		container.Command = options.Command
	}

	// 设置命令参数
	if len(options.CommandArgs) > 0 {
		container.Args = options.CommandArgs
	}

	// 设置工作目录
	if options.WorkingDir != "" {
		container.WorkingDir = options.WorkingDir
	}

	// 设置环境变量
	if len(options.EnvVars) > 0 {
		for key, value := range options.EnvVars {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	// 设置端口
	if options.Port > 0 {
		container.Ports = []corev1.ContainerPort{
			{
				ContainerPort: options.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	}

	// 设置健康检查
	if options.ReadinessProbe != nil {
		container.ReadinessProbe = options.ReadinessProbe
	}
	if options.LivenessProbe != nil {
		container.LivenessProbe = options.LivenessProbe
	}

	// 设置资源限制
	container.Resources = dm.buildResourceRequirements(options)

	return container
}

// buildVolumes 构建卷和卷挂载
func (dm *DeploymentManager) buildVolumes(options DeploymentCreateOptions) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// 处理 VolumeMounts，根据SourcePath判断是HostPath还是PVC
	for i, vm := range options.VolumeMounts {
		if len(vm.Type) == 0 {
			return nil, nil, fmt.Errorf("卷挂载类型不能为空")
		}
		// 如果SourcePath以"/"开头，认为是HostPath
		switch vm.Type {
		case "hostPath":
			// HostPath 卷
			volumeName := fmt.Sprintf("hostpath-volume-%d", i)
			hostPathType := corev1.HostPathDirectoryOrCreate

			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: vm.HostPath,
						Type: &hostPathType,
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			})
		case "pvc":
			// PVC 卷
			volumeName := fmt.Sprintf("pvc-volume-%d", i)

			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: vm.PVCName,
						ReadOnly:  vm.ReadOnly,
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: vm.MountPath,
				SubPath:   vm.SubPath,
				ReadOnly:  vm.ReadOnly,
			})
		}
	}

	return volumes, volumeMounts, nil
}

// buildImagePullSecrets 构建镜像拉取密钥
func (dm *DeploymentManager) buildImagePullSecrets(secrets []string) []corev1.LocalObjectReference {
	var imagePullSecrets []corev1.LocalObjectReference
	for _, secret := range secrets {
		imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{
			Name: secret,
		})
	}
	return imagePullSecrets
}

// buildResourceRequirements 构建资源需求
func (dm *DeploymentManager) buildResourceRequirements(options DeploymentCreateOptions) corev1.ResourceRequirements {
	requirements := corev1.ResourceRequirements{}

	// 设置资源请求
	if len(options.ResourceRequests) > 0 {
		requirements.Requests = corev1.ResourceList{}
		for k, v := range options.ResourceRequests {
			requirements.Requests[corev1.ResourceName(k)] = parseQuantity(v)
		}
	}

	// 设置资源限制
	if len(options.ResourceLimits) > 0 {
		requirements.Limits = corev1.ResourceList{}
		for k, v := range options.ResourceLimits {
			requirements.Limits[corev1.ResourceName(k)] = parseQuantity(v)
		}
	}

	return requirements
}

// parseQuantity 解析资源数量字符串
func parseQuantity(value string) resource.Quantity {
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		// 如果解析失败，返回零值
		return resource.Quantity{}
	}
	return quantity
}

// buildAutoNodeAffinity 构建自动节点亲和性
// 根据卷挂载类型设置节点亲和性：
// 1. hostPath 类型：检查 SourcePath 是否为节点 ID，设置对应节点亲和性
// 2. pvc 类型：检查存储类型是否为 local-storage，设置对应节点亲和性
func (dm *DeploymentManager) buildAutoNodeAffinity(options DeploymentCreateOptions, targetNamespace string) (*corev1.NodeAffinity, error) {
	var nodeNames []string
	var needsNodeAffinity bool

	// 遍历卷挂载配置
	for _, vm := range options.VolumeMounts {
		switch vm.Type {
		case MountTypeHostPath:
			if vm.NodeName != "" {
				// 查询节点是否存在
				node, err := dm.client.Node().GetNode(vm.NodeName)
				if err != nil {
					return nil, fmt.Errorf("节点 %s 不存在: %w", vm.NodeName, err)
				}
				nodeNames = append(nodeNames, node.Name)
				needsNodeAffinity = true
			} else {
				return nil, fmt.Errorf("hostPath 类型必须指定节点名称")
			}
		case MountTypePVC:
			// pvc 类型：检查是否为 local-storage 类型
			if vm.PVCName != "" {
				isLocalStorage, err := dm.isPVCLocalStorage(vm.PVCName, targetNamespace)
				if err != nil {
					// 检查失败，跳过此 PVC
					continue
				}

				// 只对 local-storage 类型的 PVC 设置节点亲和性
				if isLocalStorage {
					needsNodeAffinity = true
					// 获取 PVC 绑定的节点
					boundNodes, err := dm.client.Volume().GetPVCBoundNode(vm.PVCName, targetNamespace)
					if err != nil {
						// 获取失败，跳过此 PVC
						continue
					}

					// 添加绑定的节点到列表中
					for _, boundNode := range boundNodes {
						if boundNode != "" && !contains(nodeNames, boundNode) {
							nodeNames = append(nodeNames, boundNode)
						}
					}
				}
			} else {
				return nil, fmt.Errorf("pvc 类型必须指定 pvcName")
			}
		}
	}

	var nodeAffinity *corev1.NodeAffinity
	if needsNodeAffinity {
		nodeAffinity = dm.buildFlexibleNodeAffinity(nodeNames)
		return nodeAffinity, nil
	}

	// 不需要节点亲和性时返回 nil
	return nil, nil
}

// buildFlexibleNodeAffinity 构建节点亲和性策略
// 使用硬亲和性，必须调度到指定节点
func (dm *DeploymentManager) buildFlexibleNodeAffinity(nodeNames []string) *corev1.NodeAffinity {
	terms := []corev1.NodeSelectorTerm{}
	for _, nodeName := range nodeNames {
		terms = append(terms, corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      "kubernetes.io/hostname",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{nodeName},
				},
			},
		})
	}
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: terms,
		},
	}
}

// isPVCLocalStorage 检查 PVC 是否使用 local-storage 存储类
func (dm *DeploymentManager) isPVCLocalStorage(pvcName, namespace string) (bool, error) {
	// 获取 PVC 信息
	pvc, err := dm.client.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(
		context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取 PVC '%s' 失败: %w", pvcName, err)
	}

	// 检查存储类名称
	if pvc.Spec.StorageClassName == nil {
		return false, nil
	}

	storageClassName := *pvc.Spec.StorageClassName
	// log.Printf("PVC '%s' 使用存储类: %s", pvcName, storageClassName)

	// 检查是否为 local-storage 类型
	return storageClassName == "local-storage", nil
}

// contains 检查字符串切片中是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
