package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// PodManager 负责 Pod 相关操作
// 通过 Client 组合实现

type PodManager struct {
	client *Client
}

// MountType 挂载类型枚举
type MountType string

const (
	// MountTypeHostPath HostPath 类型挂载
	MountTypeHostPath MountType = "hostPath"
	// MountTypePVC PVC 类型挂载
	MountTypePVC MountType = "pvc"
	// MountTypeConfigMap ConfigMap 类型挂载
	MountTypeConfigMap MountType = "configMap"
)

// UnifiedMount 统一挂载配置结构
// 支持 HostPath、PVC 和 ConfigMap 三种挂载类型
type UnifiedMount struct {
	// 通用字段
	Type      MountType `json:"type"`               // 挂载类型
	MountPath string    `json:"mountPath"`          // 容器内挂载路径
	SubPath   string    `json:"subPath,omitempty"`  // 子路径（可选）
	ReadOnly  bool      `json:"readOnly,omitempty"` // 是否只读

	// HostPath 专用字段
	HostPath string `json:"hostPath,omitempty"` // 主机路径（HostPath 类型使用）

	// PVC 专用字段
	PVCName string `json:"pvcName,omitempty"` // PVC 名称（PVC 类型使用）

	// ConfigMap 专用字段
	ConfigMapName string `json:"configMapName,omitempty"` // ConfigMap 名称（ConfigMap 类型使用）

	// Node Name 专用字段
	NodeName string `json:"nodeName,omitempty"` // 节点名称（HostPath 类型使用）
}

// 为了保持向后兼容性，保留原有结构体定义
// ConfigMapMount ConfigMap 挂载配置（已废弃，请使用 UnifiedMount）
// Deprecated: 使用 UnifiedMount 替代
type ConfigMapMount struct {
	ConfigMapName string // ConfigMap 名称
	MountPath     string // 挂载路径
	SubPath       string // 子路径（可选）
	ReadOnly      bool   // 是否只读（默认 true）
}

// VolumeMount 卷挂载配置（已废弃，请使用 UnifiedMount）
// Deprecated: 使用 UnifiedMount 替代
type VolumeMount struct {
	HostPath   string // 来源卷路径
	TargetPath string // 目标挂载路径
}

// PVCMount PVC 挂载配置（已废弃，请使用 UnifiedMount）
// Deprecated: 使用 UnifiedMount 替代
type PVCMount struct {
	PVCName   string // PVC 名称
	MountPath string // 挂载路径
	SubPath   string // 子路径（可选）
	ReadOnly  bool   // 是否只读（默认 false）
}

// FileCopy 文件拷贝配置
type FileCopy struct {
	HostPath   string // 源文件路径
	TargetPath string // 目标文件路径
}

// PodEvent Pod 事件信息
type PodEvent struct {
	Type      string    // 事件类型 (Normal, Warning)
	Reason    string    // 事件原因
	Message   string    // 事件消息
	Timestamp time.Time // 事件时间
	Count     int32     // 事件发生次数
}

// NodeAffinityMode 节点亲和性模式
type NodeAffinityMode string

const (
	// NodeAffinityHard 硬亲和性，强制调度到指定节点
	NodeAffinityHard NodeAffinityMode = "hard"
	// NodeAffinitySoft 软亲和性，优先调度到指定节点
	NodeAffinitySoft NodeAffinityMode = "soft"
	// NodeAffinityDisabled 禁用自动节点亲和性
	NodeAffinityDisabled NodeAffinityMode = "disabled"
)

// PodCreateOptions Pod 创建选项
type PodCreateOptions struct {
	ImageName string `json:"imageName"`
	PodName   string `json:"podName"`
	Namespace string `json:"namespace,omitempty"`
	Port      int32  `json:"port,omitempty"`

	// 新的统一挂载配置（推荐使用）
	Mounts []UnifiedMount `json:"mounts,omitempty"`

	// 向后兼容的挂载配置（已废弃，建议迁移到 Mounts 字段）
	VolumeMounts    []UnifiedMount   `json:"volumeMounts,omitempty"`    // Deprecated: 使用 Mounts 替代
	PVCMounts       []PVCMount       `json:"pvcMounts,omitempty"`       // Deprecated: 使用 Mounts 替代
	ConfigMapMounts []ConfigMapMount `json:"configMapMounts,omitempty"` // Deprecated: 使用 Mounts 替代

	ReadinessProbe   *corev1.Probe     `json:"readinessProbe,omitempty"`
	Command          []string          `json:"command,omitempty"`
	CommandArgs      []string          `json:"commandArgs,omitempty"`
	EnvVars          map[string]string `json:"envVars,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	RestartPolicy    string            `json:"restartPolicy,omitempty"`
	WorkingDir       string            `json:"workingDir,omitempty"`
	ImagePullSecrets []string          `json:"imagePullSecrets,omitempty"`

	// 自动节点亲和性配置
	AutoNodeAffinity   bool                 `json:"autoNodeAffinity,omitempty"` // 是否启用自动节点亲和性
	NodeAffinityMode   NodeAffinityMode     `json:"nodeAffinityMode,omitempty"` // 节点亲和性模式
	NodeSelector       map[string]string    `json:"nodeSelector,omitempty"`     // 手动节点选择器
	CustomNodeAffinity *corev1.NodeAffinity `json:"-"`                          // 自定义节点亲和性（不序列化）
}

// Create 创建一个 Pod，支持镜像、端口、卷挂载、文件拷贝、资源限制等配置
func (pm *PodManager) Create(options PodCreateOptions) (string, error) {
	// 验证参数
	if err := pm.validateCreateOptions(options); err != nil {
		return "", err
	}

	// 构建卷挂载
	volumes, volumeMounts, err := pm.buildVolumes(options)
	if err != nil {
		return "", err
	}

	// 构建容器
	container := pm.buildContainer(options, volumeMounts)

	// 确定命名空间
	targetNamespace := pm.getTargetNamespace(options.Namespace)

	// 构建节点亲和性
	nodeAffinity, nodeSelector, err := pm.buildNodeAffinity(options, targetNamespace)
	if err != nil {
		return "", err
	}

	// 构建 Pod 规格
	podSpec := pm.buildPodSpec(container, volumes, options, nodeAffinity, nodeSelector)

	// 创建 Pod 对象
	pod := pm.buildPodObject(options, podSpec, targetNamespace)

	// 执行创建
	return pm.createPod(pod, targetNamespace)
}

// Delete 删除指定 Pod
// Delete 删除指定 Pod（保持向后兼容）
func (pm *PodManager) Delete(podName string) error {
	return pm.DeletePod(podName)
}

// DeletePod 删除指定 Pod，支持指定命名空间
func (pm *PodManager) DeletePod(podName string, namespace ...string) error {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = pm.client.namespace
	}

	return pm.client.clientset.CoreV1().Pods(targetNamespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
}

// CreatePod 创建 Pod，支持指定命名空间
func (pm *PodManager) CreatePod(options PodCreateOptions) (*corev1.Pod, error) {
	podName, err := pm.Create(options)
	if err != nil {
		return nil, err
	}

	// 获取创建的 Pod 对象
	var targetNamespace string
	if options.Namespace != "" {
		targetNamespace = options.Namespace
	} else {
		targetNamespace = pm.client.namespace
	}

	return pm.client.clientset.CoreV1().Pods(targetNamespace).Get(context.Background(), podName, metav1.GetOptions{})
}

// WaitReady 等待 Pod 进入 Running 状态并返回 Pod IP
func (pm *PodManager) IsReady(podName string) (isReady bool, runInfo string, err error) {
	// 获取 pod 详情
	pod, err := pm.GetPod(podName)
	if err != nil {
		return false, "", fmt.Errorf("获取 Pod 详细状态失败: %w", err)
	}

	// 全面检测容器异常状态
	var runInfos []string
	for _, cs := range pod.Status.ContainerStatuses {
		// 检查容器等待状态（未启动）
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			message := cs.State.Waiting.Message
			runInfos = append(runInfos, fmt.Sprintf("%s: 等待状态 - %s (%s)", cs.Name, reason, message))
			continue
		}

		// 检查容器终止状态
		if cs.State.Terminated != nil {
			reason := cs.State.Terminated.Reason
			message := cs.State.Terminated.Message
			exitCode := cs.State.Terminated.ExitCode
			runInfos = append(runInfos, fmt.Sprintf("%s: 已终止 - %s (退出码: %d, %s)", cs.Name, reason, exitCode, message))
			continue
		}

		// 检查运行中但异常的容器
		if cs.State.Running != nil {
			// 重启次数过多
			if cs.RestartCount > 3 {
				runInfos = append(runInfos, fmt.Sprintf("%s: 重启次数过多 (%d次)，可能存在异常", cs.Name, cs.RestartCount))
			}

			// 检查就绪状态
			if !cs.Ready {
				runInfos = append(runInfos, fmt.Sprintf("%s: 容器运行中但未就绪，可能健康检查失败", cs.Name))
			}
		}
	}

	// 检查初始化容器状态
	for _, ics := range pod.Status.InitContainerStatuses {
		if ics.State.Waiting != nil {
			runInfos = append(runInfos, fmt.Sprintf("初始化容器 %s: 等待状态 - %s (%s)", ics.Name, ics.State.Waiting.Reason, ics.State.Waiting.Message))
		}
		if ics.State.Terminated != nil && ics.State.Terminated.ExitCode != 0 {
			runInfos = append(runInfos, fmt.Sprintf("初始化容器 %s: 执行失败 (退出码: %d)", ics.Name, ics.State.Terminated.ExitCode))
		}
	}

	// 合并所有错误信息
	if len(runInfos) > 0 {
		return false, fmt.Sprintf("Pod 异常: %s", strings.Join(runInfos, "\n")), nil
	}

	// 全面检测 Pod 运行状态
	podReady := false
	podScheduled := false
	podInitialized := false
	containersReady := false

	for _, cond := range pod.Status.Conditions {
		switch cond.Type {
		case corev1.PodReady:
			if cond.Status == corev1.ConditionTrue {
				podReady = true
			}
		case corev1.PodScheduled:
			if cond.Status == corev1.ConditionTrue {
				podScheduled = true
			}
		case corev1.PodInitialized:
			if cond.Status == corev1.ConditionTrue {
				podInitialized = true
			}
		case corev1.ContainersReady:
			if cond.Status == corev1.ConditionTrue {
				containersReady = true
			}
		}
	}

	// 如果 Pod 未完全就绪，添加详细状态信息
	if !podReady {
		if !podScheduled {
			runInfos = append(runInfos, "Pod未调度")
		}
		if !podInitialized {
			runInfos = append(runInfos, "Pod未初始化")
		}
		if !containersReady {
			runInfos = append(runInfos, "容器未就绪")
		}

		// 获取 Pod 告警事件
		events, err := pm.GetEvents(podName, string(pod.UID))
		if err != nil {
			return false, fmt.Sprintf("Pod 异常: %s", strings.Join(runInfos, "\n")), err
		}
		for _, event := range events {
			runInfos = append(runInfos, fmt.Sprintf("警告事件: %v", event))
		}
	}
	return podReady, fmt.Sprintf("Pod 异常: %s", strings.Join(runInfos, "\n")), nil
}

// GetStatus 获取 Pod 当前状态
func (pm *PodManager) GetStatus(podName string) (corev1.PodPhase, error) {
	pod, err := pm.client.clientset.CoreV1().Pods(pm.client.namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pod.Status.Phase, nil
}

// GetEvents 获取 Pod 相关的事件信息
func (pm *PodManager) GetEvents(podName, podUID string) ([]PodEvent, error) {
	// 构建字段选择器，查找与该 Pod 相关的事件
	fieldSelector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.kind", "Pod"),
		fields.OneTermEqualSelector("involvedObject.name", podName),
		fields.OneTermEqualSelector("involvedObject.namespace", pm.client.namespace),
		fields.OneTermEqualSelector("involvedObject.uid", string(podUID)),
	)

	// 获取事件列表
	events, err := pm.client.clientset.CoreV1().Events(pm.client.namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取事件列表失败: %v", err)
	}

	// 转换为自定义的 PodEvent 结构
	var podEvents []PodEvent
	for _, event := range events.Items {
		podEvent := PodEvent{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.FirstTimestamp.Time,
			Count:     event.Count,
		}
		// 如果 FirstTimestamp 为空，使用 EventTime
		if event.FirstTimestamp.IsZero() && !event.EventTime.IsZero() {
			podEvent.Timestamp = event.EventTime.Time
		}
		podEvents = append(podEvents, podEvent)
	}

	// 按时间排序（最新的在前）
	sort.Slice(podEvents, func(i, j int) bool {
		return podEvents[i].Timestamp.After(podEvents[j].Timestamp)
	})

	return podEvents, nil
}

// GetWarningEvents 获取 Pod 的告警事件（只返回 Warning 类型的事件）
func (pm *PodManager) GetWarningEvents(podName, podUID string) ([]PodEvent, error) {
	allEvents, err := pm.GetEvents(podName, podUID)
	if err != nil {
		return nil, err
	}

	// 过滤出 Warning 类型的事件
	var warningEvents []PodEvent
	for _, event := range allEvents {
		if event.Type == "Warning" {
			warningEvents = append(warningEvents, event)
		}
	}

	return warningEvents, nil
}

func (pm *PodManager) GetPod(podName string) (*corev1.Pod, error) {
	return pm.client.clientset.CoreV1().Pods(pm.client.namespace).Get(context.Background(), podName, metav1.GetOptions{})
}

// 提取容器构建逻辑
func (pm *PodManager) buildContainer(options PodCreateOptions, volumeMounts []corev1.VolumeMount) corev1.Container {
	container := corev1.Container{
		Name:            "main",
		Image:           options.ImageName,
		ImagePullPolicy: corev1.PullAlways,
		Ports: []corev1.ContainerPort{{
			ContainerPort: options.Port,
		}},
		VolumeMounts: volumeMounts,
		Resources:    pm.buildResourceRequirements(),
	}

	// 设置可选字段
	pm.setContainerOptionalFields(&container, options)

	return container
}

func (pm *PodManager) setContainerOptionalFields(container *corev1.Container, options PodCreateOptions) {
	if options.Command != nil {
		container.Command = options.Command
	}
	if options.CommandArgs != nil {
		container.Args = options.CommandArgs
	}
	if options.ReadinessProbe != nil {
		container.ReadinessProbe = options.ReadinessProbe
	}
	if options.WorkingDir != "" {
		container.WorkingDir = options.WorkingDir
	}
	// 设置环境变量
	pm.setEnvironmentVariables(container, options.EnvVars)
}

// validateCreateOptions 验证创建选项
func (pm *PodManager) validateCreateOptions(options PodCreateOptions) error {
	if options.ImageName == "" {
		return fmt.Errorf("镜像名称不能为空")
	}
	if options.PodName == "" {
		return fmt.Errorf("Pod 名称不能为空")
	}
	return nil
}

// buildVolumes 构建卷和卷挂载（支持新的统一挂载结构和向后兼容）
func (pm *PodManager) buildVolumes(options PodCreateOptions) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// 处理新的统一挂载配置
	for i, mount := range options.Mounts {
		vol, volMount, err := pm.buildSingleMount(mount, i)
		if err != nil {
			return nil, nil, fmt.Errorf("处理统一挂载配置失败: %w", err)
		}
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, volMount)
	}

	// 处理统一的卷挂载
	for i, vm := range options.VolumeMounts {
		switch vm.Type {
		case "hostPath":
			// 验证必需字段
			if vm.HostPath == "" {
				return nil, nil, fmt.Errorf("HostPath 卷挂载的 HostPath 不能为空")
			}
			if vm.MountPath == "" {
				return nil, nil, fmt.Errorf("HostPath 卷挂载的 MountPath 不能为空")
			}

			volumeName := fmt.Sprintf("hostpath-volume-%d", i)
			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: vm.HostPath,
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			})
		case "pvc":
			// 验证必需字段
			if vm.PVCName == "" {
				return nil, nil, fmt.Errorf("PVC 卷挂载的 PVCName 不能为空")
			}
			if vm.MountPath == "" {
				return nil, nil, fmt.Errorf("PVC 卷挂载的 MountPath 不能为空")
			}

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

	// 向后兼容：处理旧的 PVC 卷挂载
	for _, pvc := range options.PVCMounts {
		// 验证必需字段
		if pvc.PVCName == "" {
			return nil, nil, fmt.Errorf("PVC 卷挂载的 PVCName 不能为空")
		}
		if pvc.MountPath == "" {
			return nil, nil, fmt.Errorf("PVC 卷挂载的 MountPath 不能为空")
		}

		volumeName := fmt.Sprintf("legacy-pvc-%s", pvc.PVCName)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.PVCName,
					ReadOnly:  pvc.ReadOnly,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: pvc.MountPath,
			SubPath:   pvc.SubPath,
			ReadOnly:  pvc.ReadOnly,
		})
	}

	// 向后兼容：处理旧的 ConfigMap 卷挂载
	for _, cm := range options.ConfigMapMounts {
		// 验证必需字段
		if cm.ConfigMapName == "" {
			return nil, nil, fmt.Errorf("ConfigMap 卷挂载的 ConfigMapName 不能为空")
		}
		if cm.MountPath == "" {
			return nil, nil, fmt.Errorf("ConfigMap 卷挂载的 MountPath 不能为空")
		}

		volumeName := fmt.Sprintf("legacy-configmap-%s", cm.ConfigMapName)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cm.ConfigMapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: cm.MountPath,
			SubPath:   cm.SubPath,
			ReadOnly:  cm.ReadOnly,
		})
	}

	return volumes, volumeMounts, nil
}

// buildSingleMount 构建单个统一挂载配置
func (pm *PodManager) buildSingleMount(mount UnifiedMount, index int) (corev1.Volume, corev1.VolumeMount, error) {
	// 验证通用字段
	if mount.MountPath == "" {
		return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("挂载路径不能为空")
	}

	switch mount.Type {
	case MountTypeHostPath:
		return pm.buildHostPathMount(mount, index)
	case MountTypePVC:
		return pm.buildPVCMount(mount)
	case MountTypeConfigMap:
		return pm.buildConfigMapMount(mount)
	default:
		return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("不支持的挂载类型: %s", mount.Type)
	}
}

// buildHostPathMount 构建 HostPath 挂载
func (pm *PodManager) buildHostPathMount(mount UnifiedMount, index int) (corev1.Volume, corev1.VolumeMount, error) {
	if mount.HostPath == "" {
		return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("HostPath 挂载的源路径不能为空")
	}

	volumeName := fmt.Sprintf("hostpath-volume-%d", index)
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: mount.HostPath,
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mount.MountPath,
		SubPath:   mount.SubPath,
		ReadOnly:  mount.ReadOnly,
	}

	return volume, volumeMount, nil
}

// buildPVCMount 构建 PVC 挂载
func (pm *PodManager) buildPVCMount(mount UnifiedMount) (corev1.Volume, corev1.VolumeMount, error) {
	if mount.PVCName == "" {
		return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("PVC 挂载的 PVC 名称不能为空")
	}

	volumeName := fmt.Sprintf("pvc-%s", mount.PVCName)
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: mount.PVCName,
				ReadOnly:  mount.ReadOnly,
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mount.MountPath,
		SubPath:   mount.SubPath,
		ReadOnly:  mount.ReadOnly,
	}

	return volume, volumeMount, nil
}

// buildConfigMapMount 构建 ConfigMap 挂载
func (pm *PodManager) buildConfigMapMount(mount UnifiedMount) (corev1.Volume, corev1.VolumeMount, error) {
	if mount.ConfigMapName == "" {
		return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("ConfigMap 挂载的 ConfigMap 名称不能为空")
	}

	volumeName := fmt.Sprintf("configmap-%s", mount.ConfigMapName)
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: mount.ConfigMapName,
				},
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mount.MountPath,
		SubPath:   mount.SubPath,
		ReadOnly:  mount.ReadOnly,
	}

	return volume, volumeMount, nil
}

// getTargetNamespace 获取目标命名空间
func (pm *PodManager) getTargetNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}
	return pm.client.namespace
}

// buildNodeAffinity 构建节点亲和性
func (pm *PodManager) buildNodeAffinity(options PodCreateOptions, targetNamespace string) (*corev1.NodeAffinity, map[string]string, error) {
	// 如果有自定义节点亲和性，直接使用
	if options.CustomNodeAffinity != nil {
		return options.CustomNodeAffinity, nil, nil
	}

	// 如果有手动节点选择器，使用节点选择器
	if len(options.NodeSelector) > 0 {
		return nil, options.NodeSelector, nil
	}

	// 如果禁用自动节点亲和性，返回空
	if options.NodeAffinityMode == NodeAffinityDisabled {
		return nil, nil, nil
	}

	// 如果启用自动节点亲和性且有 PVC 挂载，构建基于 PVC 的节点亲和性
	if options.AutoNodeAffinity && len(options.PVCMounts) > 0 {
		return pm.buildPVCBasedNodeAffinity(options.PVCMounts, options.NodeAffinityMode, targetNamespace)
	}

	return nil, nil, nil
}

// buildPVCBasedNodeAffinity 构建基于 PVC 的节点亲和性
func (pm *PodManager) buildPVCBasedNodeAffinity(pvcMounts []PVCMount, mode NodeAffinityMode, namespace string) (*corev1.NodeAffinity, map[string]string, error) {
	// 这里可以根据 PVC 的存储类或其他属性来构建节点亲和性
	// 简化实现，返回空的节点亲和性
	return nil, nil, nil
}

// buildPodSpec 构建 Pod 规格
func (pm *PodManager) buildPodSpec(container corev1.Container, volumes []corev1.Volume, options PodCreateOptions, nodeAffinity *corev1.NodeAffinity, nodeSelector map[string]string) *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{container},
		Volumes:    volumes,
	}

	// 设置重启策略
	if options.RestartPolicy != "" {
		podSpec.RestartPolicy = corev1.RestartPolicy(options.RestartPolicy)
	} else {
		podSpec.RestartPolicy = corev1.RestartPolicyAlways
	}

	// 设置镜像拉取密钥
	if len(options.ImagePullSecrets) > 0 {
		for _, secret := range options.ImagePullSecrets {
			podSpec.ImagePullSecrets = append(podSpec.ImagePullSecrets, corev1.LocalObjectReference{
				Name: secret,
			})
		}
	}

	// 设置节点选择器
	if len(nodeSelector) > 0 {
		podSpec.NodeSelector = nodeSelector
	}

	// 设置节点亲和性
	if nodeAffinity != nil {
		podSpec.Affinity = &corev1.Affinity{
			NodeAffinity: nodeAffinity,
		}
	}

	return podSpec
}

// buildPodObject 构建 Pod 对象
func (pm *PodManager) buildPodObject(options PodCreateOptions, podSpec *corev1.PodSpec, targetNamespace string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.PodName,
			Namespace: targetNamespace,
			Labels:    options.Labels,
		},
		Spec: *podSpec,
	}

	return pod
}

// createPod 执行 Pod 创建
func (pm *PodManager) createPod(pod *corev1.Pod, targetNamespace string) (string, error) {
	createdPod, err := pm.client.clientset.CoreV1().Pods(targetNamespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("创建 Pod 失败: %w", err)
	}
	return createdPod.Name, nil
}

// buildResourceRequirements 构建资源需求
func (pm *PodManager) buildResourceRequirements() corev1.ResourceRequirements {
	// 返回默认的资源需求，可以根据需要进行配置
	return corev1.ResourceRequirements{}
}

// setEnvironmentVariables 设置环境变量
func (pm *PodManager) setEnvironmentVariables(container *corev1.Container, envVars map[string]string) {
	if len(envVars) == 0 {
		return
	}

	for key, value := range envVars {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
}

// GetLogs 获取 Pod 日志
func (pm *PodManager) GetLogs(podName string, lines int64) (string, error) {
	return pm.GetLogsWithNamespace(podName, pm.client.namespace, lines)
}

// GetLogsWithNamespace 获取指定命名空间中 Pod 的日志
func (pm *PodManager) GetLogsWithNamespace(podName, namespace string, lines int64) (string, error) {
	// 设置默认行数
	if lines <= 0 {
		lines = 100
	}

	// 构建日志获取选项
	logOptions := &corev1.PodLogOptions{
		TailLines: &lines,
		Follow:    false, // 不跟踪，只获取现有日志
	}

	// 获取日志请求
	req := pm.client.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)

	// 执行请求
	logs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("获取 Pod 日志失败: %w", err)
	}
	defer logs.Close()

	// 读取日志内容
	buf := make([]byte, 1024*1024) // 1MB 缓冲区
	var result strings.Builder

	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return result.String(), nil
}
