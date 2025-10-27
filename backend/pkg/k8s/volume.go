package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeManager 负责卷相关操作
// 包括 PVC（持久卷声明）和 StorageClass（存储类）的管理
type VolumeManager struct {
	client *Client
}

// PVCInfo PVC 信息结构
type PVCInfo struct {
	Name         string            `json:"name"`         // PVC 名称
	Namespace    string            `json:"namespace"`    // 命名空间
	Status       string            `json:"status"`       // 状态 (Pending, Bound, Lost)
	VolumeName   string            `json:"volumeName"`   // 绑定的卷名称
	StorageClass string            `json:"storageClass"` // 存储类名称
	Capacity     string            `json:"capacity"`     // 容量
	AccessModes  []string          `json:"accessModes"`  // 访问模式
	Labels       map[string]string `json:"labels"`       // 标签
	CreationTime string            `json:"creationTime"` // 创建时间
	Pods         []string          `json:"pods"`         // 绑定的Pod名称列表
}

// StorageClassInfo 存储类信息结构
type StorageClassInfo struct {
	Name                 string            `json:"name"`                 // 存储类名称
	Provisioner          string            `json:"provisioner"`          // 供应商
	ReclaimPolicy        string            `json:"reclaimPolicy"`        // 回收策略
	VolumeBindingMode    string            `json:"volumeBindingMode"`    // 卷绑定模式
	AllowVolumeExpansion bool              `json:"allowVolumeExpansion"` // 是否允许卷扩展
	Parameters           map[string]string `json:"parameters"`           // 参数
	MountOptions         []string          `json:"mountOptions"`         // 挂载选项
	Labels               map[string]string `json:"labels"`               // 标签
	CreationTime         string            `json:"creationTime"`         // 创建时间
}

// ListPVCs 列出 PVC，支持查询指定命名空间或所有命名空间
// 参数 namespace: 可选，指定命名空间。传入空字符串或""查询所有命名空间
func (vm *VolumeManager) ListPVCs(namespace ...string) ([]PVCInfo, error) {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = metav1.NamespaceAll
	}

	log.Printf("正在查询命名空间 '%s' 中的 PVC...", targetNamespace)

	pvcList, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(targetNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询命名空间 '%s' 中的 PVC 失败: %w", targetNamespace, err)
	}

	log.Printf("成功查询到 %d 个 PVC", len(pvcList.Items))

	var pvcInfos []PVCInfo
	for _, pvc := range pvcList.Items {
		// 获取访问模式
		var accessModes []string
		for _, mode := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		// 获取容量
		capacity := ""
		if pvc.Status.Capacity != nil {
			if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		// 获取存储类名称
		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		// 获取绑定的Pod列表
		boundPods, err := vm.GetPVCBoundPods(pvc.Name, pvc.Namespace)
		if err != nil {
			log.Printf("获取PVC '%s' 绑定的Pod列表失败: %v", pvc.Name, err)
			boundPods = []string{} // 如果获取失败，设置为空列表
		}

		pvcInfo := PVCInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			VolumeName:   pvc.Spec.VolumeName,
			StorageClass: storageClass,
			Capacity:     capacity,
			AccessModes:  accessModes,
			Labels:       pvc.Labels,
			CreationTime: pvc.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Pods:         boundPods,
		}
		pvcInfos = append(pvcInfos, pvcInfo)
	}

	return pvcInfos, nil
}

// GetPVC 获取指定 PVC 详情
func (vm *VolumeManager) GetPVC(name string, namespace ...string) (*PVCInfo, error) {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = vm.client.namespace
	}

	log.Printf("正在查询命名空间 '%s' 中的 PVC '%s'...", targetNamespace, name)

	pvc, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(targetNamespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询命名空间 '%s' 中的 PVC '%s' 失败: %w", targetNamespace, name, err)
	}

	// 获取访问模式
	var accessModes []string
	for _, mode := range pvc.Spec.AccessModes {
		accessModes = append(accessModes, string(mode))
	}

	// 获取容量
	capacity := ""
	if pvc.Status.Capacity != nil {
		if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			capacity = storage.String()
		}
	}

	// 获取存储类名称
	storageClass := ""
	if pvc.Spec.StorageClassName != nil {
		storageClass = *pvc.Spec.StorageClassName
	}

	// 获取绑定的Pod列表
	boundPods, err := vm.GetPVCBoundPods(pvc.Name, pvc.Namespace)
	if err != nil {
		log.Printf("获取PVC '%s' 绑定的Pod列表失败: %v", pvc.Name, err)
		boundPods = []string{} // 如果获取失败，设置为空列表
	}

	pvcInfo := &PVCInfo{
		Name:         pvc.Name,
		Namespace:    pvc.Namespace,
		Status:       string(pvc.Status.Phase),
		VolumeName:   pvc.Spec.VolumeName,
		StorageClass: storageClass,
		Capacity:     capacity,
		AccessModes:  accessModes,
		Labels:       pvc.Labels,
		CreationTime: pvc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Pods:         boundPods,
	}

	return pvcInfo, nil
}

// GetPVCBoundPods 获取PVC绑定的Pod名称列表
func (vm *VolumeManager) GetPVCBoundPods(pvcName, namespace string) ([]string, error) {
	log.Printf("正在查询PVC '%s' 在命名空间 '%s' 中绑定的Pod...", pvcName, namespace)

	// 查询所有Pod
	podList, err := vm.client.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("查询Pod列表失败: %w", err)
	}

	var boundPods []string
	for _, pod := range podList.Items {
		// 检查Pod的卷声明
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				boundPods = append(boundPods, pod.Name)
				break // 找到匹配的PVC后跳出内层循环
			}
		}
	}

	log.Printf("PVC '%s' 绑定了 %d 个Pod: %v", pvcName, len(boundPods), boundPods)
	return boundPods, nil
}

// ListStorageClasses 列出所有存储类
func (vm *VolumeManager) ListStorageClasses() ([]StorageClassInfo, error) {
	scList, err := vm.client.clientset.StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var scInfos []StorageClassInfo
	for _, sc := range scList.Items {
		// 获取回收策略
		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		// 获取卷绑定模式
		volumeBindingMode := ""
		if sc.VolumeBindingMode != nil {
			volumeBindingMode = string(*sc.VolumeBindingMode)
		}

		// 获取是否允许卷扩展
		allowVolumeExpansion := false
		if sc.AllowVolumeExpansion != nil {
			allowVolumeExpansion = *sc.AllowVolumeExpansion
		}

		scInfo := StorageClassInfo{
			Name:                 sc.Name,
			Provisioner:          sc.Provisioner,
			ReclaimPolicy:        reclaimPolicy,
			VolumeBindingMode:    volumeBindingMode,
			AllowVolumeExpansion: allowVolumeExpansion,
			Parameters:           sc.Parameters,
			MountOptions:         sc.MountOptions,
			Labels:               sc.Labels,
			CreationTime:         sc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		scInfos = append(scInfos, scInfo)
	}

	return scInfos, nil
}

// GetStorageClass 获取指定存储类详情
func (vm *VolumeManager) GetStorageClass(name string) (*StorageClassInfo, error) {
	sc, err := vm.client.clientset.StorageV1().StorageClasses().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 获取回收策略
	reclaimPolicy := ""
	if sc.ReclaimPolicy != nil {
		reclaimPolicy = string(*sc.ReclaimPolicy)
	}

	// 获取卷绑定模式
	volumeBindingMode := ""
	if sc.VolumeBindingMode != nil {
		volumeBindingMode = string(*sc.VolumeBindingMode)
	}

	// 获取是否允许卷扩展
	allowVolumeExpansion := false
	if sc.AllowVolumeExpansion != nil {
		allowVolumeExpansion = *sc.AllowVolumeExpansion
	}

	scInfo := &StorageClassInfo{
		Name:                 sc.Name,
		Provisioner:          sc.Provisioner,
		ReclaimPolicy:        reclaimPolicy,
		VolumeBindingMode:    volumeBindingMode,
		AllowVolumeExpansion: allowVolumeExpansion,
		Parameters:           sc.Parameters,
		Labels:               sc.Labels,
		CreationTime:         sc.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return scInfo, nil
}

// CreatePVC 创建 PVC
func (vm *VolumeManager) CreatePVC(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	if pvc.Namespace == "" {
		pvc.Namespace = vm.client.namespace
	}
	return vm.client.clientset.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
}

// CreatePVCWithParams 根据参数创建 PVC
func (vm *VolumeManager) CreatePVCWithParams(name, nodeName, accessMode, storageClass string, storageSize int32, labels map[string]string) (*PVCInfo, error) {
	// 构建访问模式
	var accessModes []corev1.PersistentVolumeAccessMode
	switch accessMode {
	case "ReadWriteOnce":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	case "ReadOnlyMany":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	case "ReadWriteMany":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
	default:
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	// 构建存储资源请求
	storageQuantity := fmt.Sprintf("%dGi", storageSize)
	quantity, err := resource.ParseQuantity(storageQuantity)
	if err != nil {
		return nil, fmt.Errorf("解析存储大小失败: %w", err)
	}

	resources := corev1.VolumeResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: quantity,
		},
	}

	// 构建PVC对象
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: vm.client.namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources:   resources,
		},
	}

	// 如果指定了存储类，则设置存储类
	if storageClass != "" {
		pvc.Spec.StorageClassName = &storageClass
	}

	// 如果指定了节点名称，添加节点亲和性
	if nodeName != "" {
		pvc.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"kubernetes.io/hostname": nodeName,
			},
		}
	}

	log.Printf("正在创建 PVC '%s'，存储大小: %s，访问模式: %s，节点: %s", name, storageQuantity, accessMode, nodeName)

	// 创建PVC
	createdPVC, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(vm.client.namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 PVC '%s' 失败: %w", name, err)
	}

	log.Printf("成功创建 PVC '%s'", name)

	// 转换为PVCInfo结构返回
	var accessModeStrings []string
	for _, mode := range createdPVC.Spec.AccessModes {
		accessModeStrings = append(accessModeStrings, string(mode))
	}

	// 获取存储类名称
	storageClassName := ""
	if createdPVC.Spec.StorageClassName != nil {
		storageClassName = *createdPVC.Spec.StorageClassName
	}

	// 获取绑定的Pod列表（新创建的PVC通常没有绑定的Pod）
	boundPods, err := vm.GetPVCBoundPods(createdPVC.Name, createdPVC.Namespace)
	if err != nil {
		log.Printf("获取PVC '%s' 绑定的Pod列表失败: %v", createdPVC.Name, err)
		boundPods = []string{} // 如果获取失败，设置为空列表
	}

	pvcInfo := &PVCInfo{
		Name:         createdPVC.Name,
		Namespace:    createdPVC.Namespace,
		Status:       string(createdPVC.Status.Phase),
		VolumeName:   createdPVC.Spec.VolumeName,
		StorageClass: storageClassName,
		Capacity:     storageQuantity,
		AccessModes:  accessModeStrings,
		Labels:       createdPVC.Labels,
		CreationTime: createdPVC.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Pods:         boundPods,
	}

	return pvcInfo, nil
}

// DeletePVC 删除指定 PVC，支持指定命名空间
func (vm *VolumeManager) DeletePVC(name string, namespace ...string) error {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = vm.client.namespace
	}

	log.Printf("正在删除命名空间 '%s' 中的 PVC '%s'...", targetNamespace, name)

	err := vm.client.clientset.CoreV1().PersistentVolumeClaims(targetNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除命名空间 '%s' 中的 PVC '%s' 失败: %w", targetNamespace, name, err)
	}

	log.Printf("成功删除 PVC '%s'", name)
	return nil
}

// GetPVCBoundNode 获取 PVC 绑定的节点信息
// 用于本地存储类的自动节点调度
func (vm *VolumeManager) GetPVCBoundNode(pvcName string, namespace ...string) ([]string, error) {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = vm.client.namespace
	}

	log.Printf("正在查询 PVC '%s' 绑定的节点信息...", pvcName)

	// 1. 获取 PVC 信息
	pvc, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(targetNamespace).Get(
		context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PVC '%s' 失败: %w", pvcName, err)
	}

	// 2. 检查 PVC 是否已绑定
	if pvc.Status.Phase != corev1.ClaimBound {
		return nil, fmt.Errorf("PVC '%s' 未绑定到 PV，状态: %s", pvcName, pvc.Status.Phase)
	}

	// 3. 获取绑定的 PV
	pvName := pvc.Spec.VolumeName
	if pvName == "" {
		return nil, fmt.Errorf("PVC '%s' 没有绑定的 PV", pvcName)
	}

	pv, err := vm.client.clientset.CoreV1().PersistentVolumes().Get(
		context.Background(), pvName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PV '%s' 失败: %w", pvName, err)
	}

	// 4. 从 PV 的 NodeAffinity 中提取节点信息
	nodes, err := vm.extractNodesFromPV(pv)
	if err != nil {
		return nil, fmt.Errorf("从 PV '%s' 提取节点信息失败: %w", pvName, err)
	}

	log.Printf("PVC '%s' 绑定到节点: %v", pvcName, nodes)
	return nodes, nil
}

// extractNodesFromPV 从 PV 中提取节点信息
// 支持本地存储和其他有节点亲和性的存储类型
func (vm *VolumeManager) extractNodesFromPV(pv *corev1.PersistentVolume) ([]string, error) {
	var nodes []string

	// 检查是否为本地存储
	if pv.Spec.Local != nil {
		log.Printf("检测到本地存储 PV '%s'，路径: %s", pv.Name, pv.Spec.Local.Path)
	}

	// 从 NodeAffinity 中提取节点信息
	if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required != nil {
		for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
			for _, expr := range term.MatchExpressions {
				// 支持多种节点标识方式
				if (expr.Key == "kubernetes.io/hostname" ||
					expr.Key == "node.kubernetes.io/hostname" ||
					expr.Key == "kubernetes.io/instance") &&
					expr.Operator == corev1.NodeSelectorOpIn {
					nodes = append(nodes, expr.Values...)
				}
			}
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("无法从 PV '%s' 中提取节点信息，可能不是本地存储或缺少节点亲和性配置", pv.Name)
	}

	return nodes, nil
}

// CreateNodeAffinityForPVC 根据 PVC 创建节点亲和性配置
// 用于自动将 Pod 调度到存储所在节点
func (vm *VolumeManager) CreateNodeAffinityForPVC(pvcName string, namespace ...string) (*corev1.NodeAffinity, error) {
	nodes, err := vm.GetPVCBoundNode(pvcName, namespace...)
	if err != nil {
		return nil, err
	}

	// 创建硬亲和性配置，强制调度到存储节点
	nodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   nodes,
						},
					},
				},
			},
		},
	}

	log.Printf("为 PVC '%s' 创建节点亲和性配置，目标节点: %v", pvcName, nodes)
	return nodeAffinity, nil
}

// CreateSoftNodeAffinityForPVC 根据 PVC 创建软节点亲和性配置
// 优先调度到存储节点，但允许调度到其他节点
func (vm *VolumeManager) CreateSoftNodeAffinityForPVC(pvcName string, namespace ...string) (*corev1.NodeAffinity, error) {
	nodes, err := vm.GetPVCBoundNode(pvcName, namespace...)
	if err != nil {
		return nil, err
	}

	// 创建软亲和性配置，优先但不强制调度到存储节点
	nodeAffinity := &corev1.NodeAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
			{
				Weight: 100, // 最高权重
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   nodes,
						},
					},
				},
			},
		},
	}

	log.Printf("为 PVC '%s' 创建软节点亲和性配置，优先节点: %v", pvcName, nodes)
	return nodeAffinity, nil
}

// CheckPermissions 检查当前用户是否有查询 PVC 的权限
func (vm *VolumeManager) CheckPermissions() error {
	log.Printf("检查命名空间 '%s' 的 PVC 查询权限...", vm.client.namespace)

	_, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(vm.client.namespace).List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("权限检查失败，无法查询命名空间 '%s' 中的 PVC: %w", vm.client.namespace, err)
	}

	log.Printf("权限检查通过")
	return nil
}

// ListPVCsWithFilter 支持标签选择器和字段选择器的 PVC 查询
func (vm *VolumeManager) ListPVCsWithFilter(labelSelector, fieldSelector string, namespace ...string) ([]PVCInfo, error) {
	var targetNamespace string
	if len(namespace) > 0 && namespace[0] != "" {
		targetNamespace = namespace[0]
	} else {
		targetNamespace = vm.client.namespace
	}

	// 如果传入 "all"，查询所有命名空间
	if targetNamespace == "all" {
		targetNamespace = metav1.NamespaceAll
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}

	log.Printf("正在使用过滤器查询命名空间 '%s' 中的 PVC (标签: '%s', 字段: '%s')...", targetNamespace, labelSelector, fieldSelector)

	pvcList, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(targetNamespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("使用过滤器查询命名空间 '%s' 中的 PVC 失败: %w", targetNamespace, err)
	}

	log.Printf("成功查询到 %d 个符合条件的 PVC", len(pvcList.Items))

	var pvcInfos []PVCInfo
	for _, pvc := range pvcList.Items {
		// 获取访问模式
		var accessModes []string
		for _, mode := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		// 获取容量
		capacity := ""
		if pvc.Status.Capacity != nil {
			if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		// 获取存储类名称
		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		// 获取绑定的Pod列表
		boundPods, err := vm.GetPVCBoundPods(pvc.Name, pvc.Namespace)
		if err != nil {
			log.Printf("获取PVC '%s' 绑定的Pod列表失败: %v", pvc.Name, err)
			boundPods = []string{} // 如果获取失败，设置为空列表
		}

		pvcInfo := PVCInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			VolumeName:   pvc.Spec.VolumeName,
			StorageClass: storageClass,
			Capacity:     capacity,
			AccessModes:  accessModes,
			Labels:       pvc.Labels,
			CreationTime: pvc.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Pods:         boundPods,
		}
		pvcInfos = append(pvcInfos, pvcInfo)
	}

	return pvcInfos, nil
}

// CreatePV 创建 PV
func (vm *VolumeManager) CreatePV(pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
	return vm.client.clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
}

// CreateHostPathPV 创建基于主机路径的PV
func (vm *VolumeManager) CreateHostPathPV(name, hostPath, nodeName string, storageSize int32, accessModes []corev1.PersistentVolumeAccessMode, labels map[string]string, storageClassName string) (*corev1.PersistentVolume, error) {
	// 检查主机路径是否为空
	if hostPath == "" {
		return nil, fmt.Errorf("主机路径不能为空")
	}

	// 检查节点名称是否为空
	if nodeName == "" {
		return nil, fmt.Errorf("节点名称不能为空")
	}

	// 构建存储大小
	storageQuantity := fmt.Sprintf("%dGi", storageSize)
	quantity, err := resource.ParseQuantity(storageQuantity)
	if err != nil {
		return nil, fmt.Errorf("解析存储大小失败: %w", err)
	}

	// 构建PV对象
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
			AccessModes:                   accessModes,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain, // 使用Retain策略，避免数据丢失
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
					Type: func() *corev1.HostPathType {
						pathType := corev1.HostPathDirectoryOrCreate
						return &pathType
					}(),
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{nodeName},
								},
							},
						},
					},
				},
			},
		},
	}

	// 设置存储类名称（如果提供）
	if storageClassName != "" {
		pv.Spec.StorageClassName = storageClassName
	}

	log.Printf("正在创建 HostPath PV '%s'，路径: %s，节点: %s，存储大小: %s", name, hostPath, nodeName, storageQuantity)

	// 创建PV
	createdPV, err := vm.client.clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 PV '%s' 失败: %w", name, err)
	}

	log.Printf("成功创建 HostPath PV '%s'", name)
	return createdPV, nil
}

// CreateHostPathPVCWithPV 创建基于主机路径的PVC和对应的PV
func (vm *VolumeManager) CreateHostPathPVCWithPV(name, hostPath, nodeName, accessMode string, storageSize int32, storageClass string, labels map[string]string) (*PVCInfo, error) {
	// 构建访问模式
	var accessModes []corev1.PersistentVolumeAccessMode
	switch accessMode {
	case "ReadWriteOnce":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	case "ReadOnlyMany":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	case "ReadWriteMany":
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
	default:
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	// 生成PV名称（PVC名称 + 后缀）
	pvName := fmt.Sprintf("%s-pv", name)

	// 使用传入的存储类名称，如果为空则使用空字符串
	storageClassName := storageClass
	if storageClassName == "" {
		// 对于HostPath类型，默认使用空字符串表示不使用存储类
		storageClassName = ""
	}

	// 创建用于绑定的标签
	bindingLabels := make(map[string]string)
	if labels != nil {
		for k, v := range labels {
			bindingLabels[k] = v
		}
	}
	// 添加特殊的绑定标签，用于PVC选择器
	bindingLabels["pvc-name"] = name
	bindingLabels["volume-type"] = "hostpath"

	// 1. 先创建PV
	_, err := vm.CreateHostPathPV(pvName, hostPath, nodeName, storageSize, accessModes, bindingLabels, storageClassName)
	if err != nil {
		return nil, fmt.Errorf("创建 HostPath PV 失败: %w", err)
	}

	// 2. 创建PVC，使用标签选择器绑定到PV
	storageQuantity := fmt.Sprintf("%dGi", storageSize)
	quantity, err := resource.ParseQuantity(storageQuantity)
	if err != nil {
		return nil, fmt.Errorf("解析存储大小失败: %w", err)
	}

	resources := corev1.VolumeResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: quantity,
		},
	}

	// 构建PVC对象
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: vm.client.namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources:   resources,
			// 使用标签选择器而不是直接指定VolumeName
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pvc-name":    name,
					"volume-type": "hostpath",
				},
			},
		},
	}

	// 设置与PV相同的存储类名称
	if storageClassName != "" {
		pvc.Spec.StorageClassName = &storageClassName
	}

	log.Printf("正在创建 PVC '%s'，绑定到 PV '%s'，存储大小: %s，访问模式: %s，节点: %s，存储类: %s", name, pvName, storageQuantity, accessMode, nodeName, storageClassName)

	// 创建PVC
	createdPVC, err := vm.client.clientset.CoreV1().PersistentVolumeClaims(vm.client.namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		// 如果PVC创建失败，尝试清理已创建的PV
		log.Printf("PVC创建失败，正在清理PV '%s'", pvName)
		if deleteErr := vm.client.clientset.CoreV1().PersistentVolumes().Delete(context.Background(), pvName, metav1.DeleteOptions{}); deleteErr != nil {
			log.Printf("清理PV失败: %v", deleteErr)
		}
		return nil, fmt.Errorf("创建 PVC '%s' 失败: %w", name, err)
	}

	log.Printf("成功创建 PVC '%s'，已绑定到 PV '%s'", name, pvName)

	// 转换为PVCInfo结构返回
	var accessModeStrings []string
	for _, mode := range createdPVC.Spec.AccessModes {
		accessModeStrings = append(accessModeStrings, string(mode))
	}

	// 获取存储类名称
	resultStorageClass := ""
	if createdPVC.Spec.StorageClassName != nil {
		resultStorageClass = *createdPVC.Spec.StorageClassName
	}

	// 获取绑定的Pod列表（新创建的PVC通常没有绑定的Pod）
	boundPods, err := vm.GetPVCBoundPods(createdPVC.Name, createdPVC.Namespace)
	if err != nil {
		log.Printf("获取PVC '%s' 绑定的Pod列表失败: %v", createdPVC.Name, err)
		boundPods = []string{} // 如果获取失败，设置为空列表
	}

	pvcInfo := &PVCInfo{
		Name:         createdPVC.Name,
		Namespace:    createdPVC.Namespace,
		Status:       string(createdPVC.Status.Phase),
		VolumeName:   pvName, // 使用我们创建的PV名称
		StorageClass: resultStorageClass,
		Capacity:     storageQuantity,
		AccessModes:  accessModeStrings,
		Labels:       createdPVC.Labels,
		CreationTime: createdPVC.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Pods:         boundPods,
	}

	return pvcInfo, nil
}

// ... existing code ...
