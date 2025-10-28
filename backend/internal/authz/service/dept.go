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

// DeptService department HTTP service
type DeptService struct {
	deptData *biz.DeptData
}

// NewDeptService creates department service instance
func NewDeptService() *DeptService {
	return &DeptService{
		deptData: biz.NewDeptData(nil),
	}
}

// CreateDept creates department
func (s *DeptService) CreateDept(c *gin.Context) {
	var req dept.CreateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Convert request to model
	deptModel := s.convertCreateRequestToModel(&req)

	// Create department
	if err := s.deptData.CreateDept(c.Request.Context(), deptModel); err != nil {
		logger.Error("Failed to create department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create department")
		return
	}

	// Return created department information
	deptProto := s.convertModelToProto(deptModel)
	common.GinSuccess(c, deptProto)
}

// UpdateDept updates department
func (s *DeptService) UpdateDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid department ID")
		return
	}

	var req dept.UpdateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get existing department
	existingDept, err := s.deptData.GetDeptByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Department not found")
		return
	}

	// Update model
	s.updateModelFromRequest(existingDept, &req)

	// Update department
	if err := s.deptData.UpdateDept(c.Request.Context(), existingDept); err != nil {
		logger.Error("Failed to update department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to update department")
		return
	}

	// Return updated department information
	deptProto := s.convertModelToProto(existingDept)
	common.GinSuccess(c, deptProto)
}

// GetDeptById gets department by ID
func (s *DeptService) GetDeptById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid department ID")
		return
	}

	deptModel, err := s.deptData.GetDeptByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Department not found")
		return
	}

	deptProto := s.convertModelToProto(deptModel)
	common.GinSuccess(c, deptProto)
}

// DeleteDept deletes department
func (s *DeptService) DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid department ID")
		return
	}

	if err := s.deptData.DeleteDept(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Failed to delete department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to delete department")
		return
	}

	common.GinSuccess(c, nil)
}

// GetDeptTree gets department tree structure
func (s *DeptService) GetDeptTree(c *gin.Context) {
	deptTree, err := s.deptData.GetDeptTree(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get department tree", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to get department tree")
		return
	}

	// Convert to Proto format
	treeProto := make([]*dept.SysDept, 0, len(deptTree))
	for _, d := range deptTree {
		treeProto = append(treeProto, s.convertModelToProto(d))
	}

	common.GinSuccess(c, treeProto)
}

// ListDepts gets department list
func (s *DeptService) ListDepts(c *gin.Context) {
	var req dept.ListDeptsRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// Convert status parameter
	var status *bool
	if req.Status != 0 {
		enabled := req.Status == 1
		status = &enabled
	}

	deptList, err := s.deptData.GetDeptList(c.Request.Context(), req.Name, status)
	if err != nil {
		logger.Error("Failed to get department list", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to get department list")
		return
	}

	// Convert to Proto format
	listProto := make([]*dept.SysDept, 0, len(deptList))
	for _, d := range deptList {
		listProto = append(listProto, s.convertModelToProto(d))
	}

	common.GinSuccess(c, listProto)
}

// PageDepts gets department list with pagination
func (s *DeptService) PageDepts(c *gin.Context) {
	var req dept.PageDeptsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Convert status parameter
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
		logger.Error("Failed to get department list", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to get department list")
		return
	}

	// Simple pagination handling (should be implemented in data layer)
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

	// Convert to Proto format
	listProto := make([]*dept.SysDept, 0, len(deptList))
	for _, d := range deptList {
		listProto = append(listProto, s.convertModelToProto(d))
	}

	// Build pagination response
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

// convertCreateRequestToModel converts create request to model
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

// updateModelFromRequest updates model from update request
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

// convertModelToProto converts model to Proto
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
