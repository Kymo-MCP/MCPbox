package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kymo-mcp/mcpcan/api/market/resource"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"

	"github.com/gin-gonic/gin"
)

// ResourceService provides resource management functionality
type ResourceService struct {
	ctx context.Context
}

// NewResourceService creates a new ResourceService instance
func NewResourceService(ctx context.Context) *ResourceService {
	return &ResourceService{
		ctx: ctx,
	}
}

// convertPVCInfo converts PVC information to protobuf format
func convertPVCInfo(k8sPVC *k8s.PVCInfo) *resource.PVCInfo {
	return &resource.PVCInfo{
		Name:         k8sPVC.Name,
		Namespace:    k8sPVC.Namespace,
		Status:       k8sPVC.Status,
		VolumeName:   k8sPVC.VolumeName,
		StorageClass: k8sPVC.StorageClass,
		Capacity:     k8sPVC.Capacity,
		AccessModes:  k8sPVC.AccessModes,
		Labels:       k8sPVC.Labels,
		CreationTime: k8sPVC.CreationTime,
		Pods:         k8sPVC.Pods,
	}
}

// convertNodeInfo converts node information to protobuf format
func convertNodeInfo(k8sNode k8s.NodeInfo) *resource.NodeInfo {
	return &resource.NodeInfo{
		Name:              k8sNode.Name,
		Status:            k8sNode.Status,
		Roles:             k8sNode.Roles,
		Version:           k8sNode.Version,
		InternalIp:        k8sNode.InternalIP,
		ExternalIp:        k8sNode.ExternalIP,
		OperatingSystem:   k8sNode.OperatingSystem,
		Architecture:      k8sNode.Architecture,
		KernelVersion:     k8sNode.KernelVersion,
		ContainerRuntime:  k8sNode.ContainerRuntime,
		AllocatableMemory: k8sNode.AllocatableMemory,
		AllocatableCpu:    k8sNode.AllocatableCPU,
		AllocatablePods:   k8sNode.AllocatablePods,
		Labels:            k8sNode.Labels,
		Annotations:       k8sNode.Annotations,
		CreationTime:      k8sNode.CreationTime,
	}
}

// convertStorageClassInfo converts storage class information to protobuf format
func convertStorageClassInfo(k8sSC k8s.StorageClassInfo) *resource.StorageClassInfo {
	return &resource.StorageClassInfo{
		Name:                 k8sSC.Name,
		Provisioner:          k8sSC.Provisioner,
		ReclaimPolicy:        k8sSC.ReclaimPolicy,
		VolumeBindingMode:    k8sSC.VolumeBindingMode,
		Parameters:           k8sSC.Parameters,
		AllowVolumeExpansion: k8sSC.AllowVolumeExpansion,
		MountOptions:         k8sSC.MountOptions,
	}
}

// ListPVCsHandler handles PVC listing requests
func (s *ResourceService) ListPVCsHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListPVCsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 使用 ResourceService 处理请求
	result, err := s.ListPVCs(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListPVCs retrieves a list of PVCs
func (s *ResourceService) ListPVCs(req *resource.ListPVCsRequest) (*resource.ListPVCsResponse, error) {
	// 使用数据处理层获取PVC列表
	pvcList, err := biz.GResourceBiz.ListPVCs(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("获取PVC列表失败: %s", err.Error())
	}

	// 转换为 protobuf 类型
	var pbPVCList []*resource.PVCInfo
	for _, pvc := range pvcList {
		pbPVCList = append(pbPVCList, convertPVCInfo(&pvc))
	}

	// 构建响应
	response := &resource.ListPVCsResponse{
		List: pbPVCList,
	}

	return response, nil
}

// ListPVCsHandler 获取PVC列表接口（包级函数，保持向后兼容）
func ListPVCsHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListPVCsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 使用数据处理层获取PVC列表
	pvcList, err := biz.GResourceBiz.ListPVCs(uint(req.EnvironmentId))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取PVC列表失败: %s", err.Error()))
		return
	}

	// 转换为 protobuf 类型
	var pbPVCList []*resource.PVCInfo
	for _, pvc := range pvcList {
		pbPVCList = append(pbPVCList, convertPVCInfo(&pvc))
	}

	// 构建响应
	response := &resource.ListPVCsResponse{
		List: pbPVCList,
	}

	common.GinSuccess(c, response)
}

// CreatePVCHandler 创建PVC接口Handler
func (s *ResourceService) CreatePVCHandler(c *gin.Context) {
	var req resource.CreatePVCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "请求参数错误: "+err.Error())
		return
	}

	// 使用 ResourceService 处理请求
	result, err := s.CreatePVC(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// CreatePVC creates a new PVC
func (s *ResourceService) CreatePVC(req *resource.CreatePVCRequest) (*resource.CreatePVCResponse, error) {
	// 验证必需参数
	if req.Name == "" {
		return nil, fmt.Errorf("PVC名称不能为空")
	}
	if req.EnvironmentId <= 0 {
		return nil, fmt.Errorf("环境ID必须大于0")
	}
	if req.StorageSize <= 0 {
		return nil, fmt.Errorf("存储大小必须大于0")
	}

	// 验证访问模式
	validAccessModes := map[string]bool{
		"ReadWriteOnce": true,
		"ReadOnlyMany":  true,
		"ReadWriteMany": true,
	}
	if req.AccessMode != "" && !validAccessModes[req.AccessMode] {
		return nil, fmt.Errorf("无效的访问模式，支持: ReadWriteOnce, ReadOnlyMany, ReadWriteMany")
	}

	var pvcInfo *k8s.PVCInfo
	var err error

	// 根据是否提供hostPath选择不同的创建方法
	if req.HostPath != "" {
		// 创建基于主机路径的PVC
		if req.NodeName == "" {
			return nil, fmt.Errorf("创建HostPath类型PVC时，节点名称不能为空")
		}
		pvcInfo, err = biz.GResourceBiz.CreateHostPathPVC(
			uint(req.EnvironmentId),
			req.Name,
			req.HostPath,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
		)
	} else {
		// 创建普通PVC
		pvcInfo, err = biz.GResourceBiz.CreatePVC(
			uint(req.EnvironmentId),
			req.Name,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
			nil, // labels
		)
	}

	if err != nil {
		return nil, fmt.Errorf("创建PVC失败: %s", err.Error())
	}

	// 转换为 protobuf 类型并返回
	return &resource.CreatePVCResponse{
		Pvc: convertPVCInfo(pvcInfo),
	}, nil
}

// CreatePVCHandler 创建PVC接口（包级函数，保持向后兼容）
func CreatePVCHandler(c *gin.Context) {
	var req resource.CreatePVCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必需参数
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PVC名称不能为空"})
		return
	}
	if req.EnvironmentId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "环境ID必须大于0"})
		return
	}
	if req.StorageSize <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "存储大小必须大于0"})
		return
	}

	// 验证访问模式
	validAccessModes := map[string]bool{
		"ReadWriteOnce": true,
		"ReadOnlyMany":  true,
		"ReadWriteMany": true,
	}
	if req.AccessMode != "" && !validAccessModes[req.AccessMode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的访问模式，支持: ReadWriteOnce, ReadOnlyMany, ReadWriteMany"})
		return
	}

	var pvcInfo *k8s.PVCInfo
	var err error

	// 根据是否提供hostPath选择不同的创建方法
	if req.HostPath != "" {
		// 创建基于主机路径的PVC
		if req.NodeName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "创建HostPath类型PVC时，节点名称不能为空"})
			return
		}
		pvcInfo, err = biz.GResourceBiz.CreateHostPathPVC(
			uint(req.EnvironmentId),
			req.Name,
			req.HostPath,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
		)
	} else {
		// 创建普通PVC（使用StorageClass）
		pvcInfo, err = biz.GResourceBiz.CreatePVC(
			uint(req.EnvironmentId),
			req.Name,
			req.NodeName,
			req.AccessMode,
			req.StorageClass,
			req.StorageSize,
			nil, // Labels 字段在 proto 中不存在，传入 nil
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建PVC失败: " + err.Error()})
		return
	}

	// 构建响应
	response := &resource.CreatePVCResponse{
		Pvc: convertPVCInfo(pvcInfo),
	}

	c.JSON(http.StatusOK, response)
}

// ListNodesHandler 获取节点列表接口Handler
func (s *ResourceService) ListNodesHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListNodesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取节点列表失败: %s", err.Error()))
		return
	}

	// 使用 ResourceService 处理请求
	result, err := s.ListNodes(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListNodes 获取节点列表业务逻辑
func (s *ResourceService) ListNodes(req *resource.ListNodesRequest) (*resource.ListNodesResponse, error) {
	// 使用数据处理层获取节点列表
	nodeList, err := biz.GResourceBiz.ListNodes(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %s", err.Error())
	}

	// 转换为 protobuf 类型
	var pbNodeList []*resource.NodeInfo
	for _, node := range nodeList {
		pbNodeList = append(pbNodeList, convertNodeInfo(node))
	}

	// 构建响应
	response := &resource.ListNodesResponse{
		List: pbNodeList,
	}

	return response, nil
}

// ListNodesHandler 获取节点列表接口（包级函数，保持向后兼容）
func ListNodesHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListNodesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取节点列表失败: %s", err.Error()))
		return
	}

	// 使用数据处理层获取节点列表
	nodeList, err := biz.GResourceBiz.ListNodes(uint(req.EnvironmentId))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取节点列表失败: %s", err.Error()))
		return
	}

	// 转换为 protobuf 类型
	var pbNodeList []*resource.NodeInfo
	for _, node := range nodeList {
		pbNodeList = append(pbNodeList, convertNodeInfo(node))
	}

	// 构建响应
	response := &resource.ListNodesResponse{
		List: pbNodeList,
	}

	common.GinSuccess(c, response)
}

// ListStorageClassesHandler 获取存储类列表接口Handler
func (s *ResourceService) ListStorageClassesHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListStorageClassesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 使用 ResourceService 处理请求
	result, err := s.ListStorageClasses(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, result)
}

// ListStorageClasses 获取存储类列表业务逻辑
func (s *ResourceService) ListStorageClasses(req *resource.ListStorageClassesRequest) (*resource.ListStorageClassesResponse, error) {
	// 使用数据处理层获取存储类列表
	storageClassList, err := biz.GResourceBiz.ListStorageClasses(uint(req.EnvironmentId))
	if err != nil {
		return nil, fmt.Errorf("获取存储类列表失败: %s", err.Error())
	}

	// 转换为 protobuf 类型
	var pbStorageClassList []*resource.StorageClassInfo
	for _, sc := range storageClassList {
		pbStorageClassList = append(pbStorageClassList, convertStorageClassInfo(sc))
	}

	// 构建响应
	response := &resource.ListStorageClassesResponse{
		List: pbStorageClassList,
	}

	return response, nil
}

// ListStorageClassesHandler 获取存储类列表接口（包级函数，保持向后兼容）
func ListStorageClassesHandler(c *gin.Context) {
	// 获取环境ID参数
	var req resource.ListStorageClassesRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取存储类列表失败: %s", err.Error()))
		return
	}

	// 使用数据处理层获取存储类列表
	storageClassList, err := biz.GResourceBiz.ListStorageClasses(uint(req.EnvironmentId))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("获取存储类列表失败: %s", err.Error()))
		return
	}

	// 转换为 protobuf 类型
	var pbStorageClassList []*resource.StorageClassInfo
	for _, sc := range storageClassList {
		pbStorageClassList = append(pbStorageClassList, convertStorageClassInfo(sc))
	}

	// 构建响应
	response := &resource.ListStorageClassesResponse{
		List: pbStorageClassList,
	}

	common.GinSuccess(c, response)
}
