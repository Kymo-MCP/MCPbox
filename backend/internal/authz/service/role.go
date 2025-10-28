package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/authz/role"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
)

// RoleService role HTTP service
type RoleService struct {
	roleData *biz.RoleData
}

// NewRoleService creates role service instance
func NewRoleService() *RoleService {
	return &RoleService{
		roleData: biz.NewRoleData(nil),
	}
}

// CreateRole creates role
func (s *RoleService) CreateRole(c *gin.Context) {
	var req role.CreateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Check if role name already exists
	existingRole, err := s.roleData.GetRoleByName(c.Request.Context(), req.Name)
	if err == nil && existingRole != nil {
		common.GinError(c, i18nresp.CodeInternalError, "role name already exists")
		return
	}

	// Convert request to model
	roleModel := s.convertCreateRequestToModel(&req)

	// Create role
	if err := s.roleData.CreateRole(c.Request.Context(), roleModel); err != nil {
		logger.Error("create role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "create role failed")
		return
	}

	// Return created role information
	response := s.convertModelToProto(roleModel)
	common.GinSuccess(c, response)
}

// UpdateRole updates role
func (s *RoleService) UpdateRole(c *gin.Context) {
	var req role.UpdateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid role ID")
		return
	}
	req.Id = int64(id)

	// Get existing role
	existingRole, err := s.roleData.GetRoleByID(c.Request.Context(), uint(req.Id))
	if err != nil {
		logger.Error("get role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "role does not exist")
		return
	}

	// Update model
	s.updateModelFromRequest(existingRole, &req)

	// Update role
	if err := s.roleData.UpdateRole(c.Request.Context(), existingRole); err != nil {
		logger.Error("update role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "update role failed")
		return
	}

	// Return updated role information
	response := s.convertModelToProto(existingRole)
	common.GinSuccess(c, response)
}

// GetRoleById gets role by ID
func (s *RoleService) GetRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid role ID")
		return
	}

	roleModel, err := s.roleData.GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("get role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "role does not exist")
		return
	}

	response := s.convertModelToProto(roleModel)
	common.GinSuccess(c, response)
}

// DeleteRole deletes role
func (s *RoleService) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid role ID")
		return
	}

	if err := s.roleData.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		logger.Error("delete role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "delete role failed")
		return
	}

	common.GinSuccess(c, gin.H{"message": "delete successful"})
}

// ListRoles gets role list with pagination
func (s *RoleService) ListRoles(c *gin.Context) {
	var req role.ListRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Convert query parameters
	pageNum := int(req.PageInfo.Page)
	pageSize := int(req.PageInfo.Size)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	roles, total, err := s.roleData.GetRoleListWithPagination(c.Request.Context(), pageNum, pageSize, req.Query.Blurry, nil)
	if err != nil {
		logger.Error("get role list failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "get role list failed")
		return
	}

	// Convert to Proto format
	var protoRoles []*role.SysRole
	for _, r := range roles {
		protoRoles = append(protoRoles, s.convertModelToProto(r))
	}

	// Build pagination response
	response := &role.ListRolesResponse{
		Data: &role.PageSysRole{
			Roles: protoRoles,
			PageInfo: &role.PageInfo{
				Page:  int32(pageNum),
				Size:  int32(pageSize),
				Total: total,
				Pages: int32((total + int64(pageSize) - 1) / int64(pageSize)),
			},
		},
	}

	common.GinSuccess(c, response)
}

// GetRolePermissions gets role permissions
func (s *RoleService) GetRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "invalid role ID")
		return
	}

	// Get role permissions
	permissions, err := s.roleData.GetRolePermissions(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("get role permissions failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "get role permissions failed")
		return
	}

	// Convert to permission tree format
	permissionTree := s.convertPermissionsToTree(permissions)

	response := &role.GetRolePermissionsResponse{
		Permissions: permissionTree,
	}

	common.GinSuccess(c, response)
}

// convertCreateRequestToModel converts create request to model
func (s *RoleService) convertCreateRequestToModel(req *role.CreateRoleRequest) *model.SysRole {
	roleModel := &model.SysRole{
		Name: req.Name,
	}

	if req.Description != "" {
		roleModel.Description = &req.Description
	}

	if req.Sort > 0 {
		level := int(req.Sort)
		roleModel.Level = &level
	}

	return roleModel
}

// updateModelFromRequest updates model from update request
func (s *RoleService) updateModelFromRequest(roleModel *model.SysRole, req *role.UpdateRoleRequest) {
	roleModel.Name = req.Name

	if req.Description != "" {
		roleModel.Description = &req.Description
	} else {
		roleModel.Description = nil
	}

	if req.Sort > 0 {
		level := int(req.Sort)
		roleModel.Level = &level
	} else {
		roleModel.Level = nil
	}
}

// convertPermissionsToTree converts permission string list to permission tree structure
func (s *RoleService) convertPermissionsToTree(permissions []string) []*role.PermissionTreeNode {
	permissionTree := make([]*role.PermissionTreeNode, 0, len(permissions))

	// Create a permission tree node for each permission string
	for i, perm := range permissions {
		node := &role.PermissionTreeNode{
			Id:     int64(i + 1), // Temporary ID, should be obtained from permission table
			Name:   perm,
			Code:   perm,
			Type:   1, // Default type
			Status: "enabled",
			Hidden: false,
		}
		permissionTree = append(permissionTree, node)
	}

	return permissionTree
}

// convertModelToProto converts model to Proto
func (s *RoleService) convertModelToProto(roleModel *model.SysRole) *role.SysRole {
	roleProto := &role.SysRole{
		Id:   int64(roleModel.RoleID),
		Name: roleModel.Name,
	}

	if roleModel.Description != nil {
		roleProto.Description = *roleModel.Description
	}

	if roleModel.Level != nil {
		roleProto.Sort = int32(*roleModel.Level)
	}

	// Default status is enabled
	roleProto.Status = role.RoleStatus_RoleStatusEnabled

	if roleModel.CreateTime != nil {
		roleProto.CreatedAt = roleModel.CreateTime.Unix()
	}

	if roleModel.UpdateTime != nil {
		roleProto.UpdatedAt = roleModel.UpdateTime.Unix()
	}

	return roleProto
}
