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

// RoleService 角色HTTP服务
type RoleService struct {
	roleData *biz.RoleData
}

// NewRoleService 创建角色服务实例
func NewRoleService() *RoleService {
	return &RoleService{
		roleData: biz.NewRoleData(nil),
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(c *gin.Context) {
	var req role.CreateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 检查角色名称是否已存在
	existingRole, err := s.roleData.GetRoleByName(c.Request.Context(), req.Name)
	if err == nil && existingRole != nil {
		common.GinError(c, i18nresp.CodeInternalError, "角色名称已存在")
		return
	}

	// 转换请求到模型
	roleModel := s.convertCreateRequestToModel(&req)

	// 创建角色
	if err := s.roleData.CreateRole(c.Request.Context(), roleModel); err != nil {
		logger.Error("创建角色失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "创建角色失败")
		return
	}

	// 返回创建的角色信息
	roleProto := s.convertModelToProto(roleModel)
	response := &role.CreateRoleResponse{
		Role: roleProto,
	}
	common.GinSuccess(c, response)
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的角色ID")
		return
	}

	var req role.UpdateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 获取现有角色
	existingRole, err := s.roleData.GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取角色失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "角色不存在")
		return
	}

	// 更新模型
	s.updateModelFromRequest(existingRole, &req)

	// 更新角色
	if err := s.roleData.UpdateRole(c.Request.Context(), existingRole); err != nil {
		logger.Error("更新角色失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "更新角色失败")
		return
	}

	// 返回更新后的角色信息
	roleProto := s.convertModelToProto(existingRole)
	response := &role.UpdateRoleResponse{
		Role: roleProto,
	}
	common.GinSuccess(c, response)
}

// GetRoleById 根据ID获取角色
func (s *RoleService) GetRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的角色ID")
		return
	}

	roleModel, err := s.roleData.GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取角色失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "角色不存在")
		return
	}

	roleProto := s.convertModelToProto(roleModel)
	response := &role.GetRoleByIdResponse{
		Role: roleProto,
	}
	common.GinSuccess(c, response)
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的角色ID")
		return
	}

	if err := s.roleData.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		logger.Error("删除角色失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除角色失败")
		return
	}

	common.GinSuccess(c, &role.DeleteRoleResponse{})
}

// ListRoles 分页获取角色列表
func (s *RoleService) ListRoles(c *gin.Context) {
	var req role.ListRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 转换查询参数
	var name string
	var level *int
	if req.Query != nil {
		name = req.Query.Blurry
	}

	roleList, err := s.roleData.GetRoleList(c.Request.Context(), name, level)
	if err != nil {
		logger.Error("获取角色列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取角色列表失败")
		return
	}

	// 简单分页处理（实际应该在数据层实现）
	page := int(1)
	size := int(10)
	if req.PageInfo != nil {
		if req.PageInfo.Page > 0 {
			page = int(req.PageInfo.Page)
		}
		if req.PageInfo.Size > 0 {
			size = int(req.PageInfo.Size)
		}
	}

	total := int64(len(roleList))
	start := (page - 1) * size
	end := start + size

	if start >= len(roleList) {
		roleList = []*model.SysRole{}
	} else if end > len(roleList) {
		roleList = roleList[start:]
	} else {
		roleList = roleList[start:end]
	}

	// 转换为Proto格式
	listProto := make([]*role.SysRole, 0, len(roleList))
	for _, r := range roleList {
		listProto = append(listProto, s.convertModelToProto(r))
	}

	// 构建分页响应
	pageInfo := &role.PageInfo{
		Page:  int32(page),
		Size:  int32(size),
		Total: total,
		Pages: int32((total + int64(size) - 1) / int64(size)),
	}

	pageData := &role.PageSysRole{
		Roles:    listProto,
		PageInfo: pageInfo,
	}

	response := &role.ListRolesResponse{
		Data: pageData,
	}

	common.GinSuccess(c, response)
}

// GetRolePermissions 获取角色权限
func (s *RoleService) GetRolePermissions(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的角色ID")
		return
	}

	// 获取角色权限
	permissions, err := s.roleData.GetRolePermissions(c.Request.Context(), uint(roleId))
	if err != nil {
		logger.Error("获取角色权限失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取角色权限失败")
		return
	}

	// 转换为权限树格式
	permissionTree := s.convertPermissionsToTree(permissions)

	response := &role.GetRolePermissionsResponse{
		Permissions: permissionTree,
	}

	common.GinSuccess(c, response)
}

// convertCreateRequestToModel 将创建请求转换为模型
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

// updateModelFromRequest 从更新请求更新模型
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

// convertPermissionsToTree 将权限字符串列表转换为权限树结构
func (s *RoleService) convertPermissionsToTree(permissions []string) []*role.PermissionTreeNode {
	permissionTree := make([]*role.PermissionTreeNode, 0, len(permissions))

	// 为每个权限字符串创建一个权限树节点
	for i, perm := range permissions {
		node := &role.PermissionTreeNode{
			Id:     int64(i + 1), // 临时ID，实际应该从权限表获取
			Name:   perm,
			Code:   perm,
			Type:   1, // 默认类型
			Status: "enabled",
			Hidden: false,
		}
		permissionTree = append(permissionTree, node)
	}

	return permissionTree
}

// convertModelToProto 将模型转换为Proto
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

	// 默认状态为启用
	roleProto.Status = role.RoleStatus_RoleStatusEnabled

	if roleModel.CreateTime != nil {
		roleProto.CreatedAt = roleModel.CreateTime.Unix()
	}

	if roleModel.UpdateTime != nil {
		roleProto.UpdatedAt = roleModel.UpdateTime.Unix()
	}

	return roleProto
}
