package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/authz/dept"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
)

// DeptService 部门HTTP服务
type DeptService struct {
	deptData *biz.DeptData
}

// NewDeptService 创建部门服务实例
func NewDeptService() *DeptService {
	return &DeptService{
		deptData: biz.NewDeptData(nil),
	}
}

// CreateDept 创建部门
func (s *DeptService) CreateDept(c *gin.Context) {
	var req dept.CreateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 转换请求到模型
	deptModel := s.convertCreateRequestToModel(&req)

	// 创建部门
	if err := s.deptData.CreateDept(c.Request.Context(), deptModel); err != nil {
		logger.Error("创建部门失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "创建部门失败")
		return
	}

	// 返回创建的部门信息
	deptProto := s.convertModelToProto(deptModel)
	common.GinSuccess(c, deptProto)
}

// UpdateDept 更新部门
func (s *DeptService) UpdateDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的部门ID")
		return
	}

	var req dept.UpdateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 获取现有部门
	existingDept, err := s.deptData.GetDeptByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取部门失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "部门不存在")
		return
	}

	// 更新模型
	s.updateModelFromRequest(existingDept, &req)

	// 更新部门
	if err := s.deptData.UpdateDept(c.Request.Context(), existingDept); err != nil {
		logger.Error("更新部门失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "更新部门失败")
		return
	}

	// 返回更新后的部门信息
	deptProto := s.convertModelToProto(existingDept)
	common.GinSuccess(c, deptProto)
}

// GetDeptById 根据ID获取部门
func (s *DeptService) GetDeptById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的部门ID")
		return
	}

	deptModel, err := s.deptData.GetDeptByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取部门失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "部门不存在")
		return
	}

	deptProto := s.convertModelToProto(deptModel)
	common.GinSuccess(c, deptProto)
}

// DeleteDept 删除部门
func (s *DeptService) DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的部门ID")
		return
	}

	if err := s.deptData.DeleteDept(c.Request.Context(), uint(id)); err != nil {
		logger.Error("删除部门失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除部门失败")
		return
	}

	common.GinSuccess(c, nil)
}

// GetDeptTree 获取部门树形结构
func (s *DeptService) GetDeptTree(c *gin.Context) {
	deptTree, err := s.deptData.GetDeptTree(c.Request.Context())
	if err != nil {
		logger.Error("获取部门树失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取部门树失败")
		return
	}

	// 转换为Proto格式
	treeProto := make([]*dept.SysDept, 0, len(deptTree))
	for _, d := range deptTree {
		treeProto = append(treeProto, s.convertModelToProto(d))
	}

	common.GinSuccess(c, treeProto)
}

// ListDepts 获取部门列表
func (s *DeptService) ListDepts(c *gin.Context) {
	var req dept.ListDeptsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 转换状态参数
	var status *bool
	if req.Status != 0 {
		enabled := req.Status == 1
		status = &enabled
	}

	deptList, err := s.deptData.GetDeptList(c.Request.Context(), req.Name, status)
	if err != nil {
		logger.Error("获取部门列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取部门列表失败")
		return
	}

	// 转换为Proto格式
	listProto := make([]*dept.SysDept, 0, len(deptList))
	for _, d := range deptList {
		listProto = append(listProto, s.convertModelToProto(d))
	}

	common.GinSuccess(c, listProto)
}

// PageDepts 分页获取部门列表
func (s *DeptService) PageDepts(c *gin.Context) {
	var req dept.PageDeptsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 转换状态参数
	var status *bool
	if req.Query != nil && req.Query.Status != 0 {
		enabled := req.Query.Status == 1
		status = &enabled
	}

	var name string
	if req.Query != nil {
		name = req.Query.Name
	}

	deptList, err := s.deptData.GetDeptList(c.Request.Context(), name, status)
	if err != nil {
		logger.Error("获取部门列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取部门列表失败")
		return
	}

	// 简单分页处理（实际应该在数据层实现）
	page := int(req.Query.Page)
	size := int(req.Query.Size)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	total := int64(len(deptList))
	start := (page - 1) * size
	end := start + size

	if start >= len(deptList) {
		deptList = []*model.SysDept{}
	} else if end > len(deptList) {
		deptList = deptList[start:]
	} else {
		deptList = deptList[start:end]
	}

	// 转换为Proto格式
	listProto := make([]*dept.SysDept, 0, len(deptList))
	for _, d := range deptList {
		listProto = append(listProto, s.convertModelToProto(d))
	}

	// 构建分页响应
	pageInfo := &dept.PageInfo{
		TotalElements:    total,
		TotalPages:       int32((total + int64(size) - 1) / int64(size)),
		First:            page == 1,
		Last:             int64(end) >= total,
		Size:             int32(size),
		Number:           int32(page),
		NumberOfElements: int32(len(listProto)),
		Empty:            len(listProto) == 0,
	}

	response := &dept.PageSysDept{
		Depts:    listProto,
		PageInfo: pageInfo,
	}

	common.GinSuccess(c, response)
}

// convertCreateRequestToModel 将创建请求转换为模型
func (s *DeptService) convertCreateRequestToModel(req *dept.CreateDeptRequest) *model.SysDept {
	deptModel := &model.SysDept{}
	deptModel.Name = req.Dept.Name
	deptModel.DeptSort = int(req.Dept.Sort)

	if req.Dept.ParentId > 0 {
		parentId := uint(req.Dept.ParentId)
		deptModel.PID = &parentId
	}

	if req.Dept.Status == dept.DeptStatus_DeptStatusEnabled {
		deptModel.Enabled = true
	} else {
		deptModel.Enabled = false
	}

	return deptModel
}

// updateModelFromRequest 从更新请求更新模型
func (s *DeptService) updateModelFromRequest(deptModel *model.SysDept, req *dept.UpdateDeptRequest) {
	if req.Dept.Name != "" {
		deptModel.Name = req.Dept.Name
	}
	if req.Dept.Sort > 0 {
		deptModel.DeptSort = int(req.Dept.Sort)
	}
	if req.Dept.ParentId > 0 {
		parentId := uint(req.Dept.ParentId)
		deptModel.PID = &parentId
	}
	if req.Dept.Status == dept.DeptStatus_DeptStatusEnabled {
		deptModel.Enabled = true
	} else if req.Dept.Status == dept.DeptStatus_DeptStatusDisabled {
		deptModel.Enabled = false
	}
}

// convertModelToProto 将模型转换为Proto
func (s *DeptService) convertModelToProto(deptModel *model.SysDept) *dept.SysDept {
	deptProto := &dept.SysDept{
		Id:   int64(deptModel.DeptID),
		Name: deptModel.Name,
		Sort: int32(deptModel.DeptSort),
	}

	if deptModel.PID != nil {
		deptProto.ParentId = int64(*deptModel.PID)
	}

	if deptModel.Enabled {
		deptProto.Status = dept.DeptStatus_DeptStatusEnabled
	} else {
		deptProto.Status = dept.DeptStatus_DeptStatusDisabled
	}

	if deptModel.CreateTime != nil {
		deptProto.CreatedAt = deptModel.CreateTime.Unix()
	}
	if deptModel.UpdateTime != nil {
		deptProto.UpdatedAt = deptModel.UpdateTime.Unix()
	}

	return deptProto
}
