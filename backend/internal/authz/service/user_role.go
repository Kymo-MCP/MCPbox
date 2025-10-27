package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/authz/user_role"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
)

// UserRoleService 用户角色服务
type UserRoleService struct {
	userRoleUseCase *biz.UserRoleUseCase
}

// NewUserRoleService 创建用户角色服务实例
func NewUserRoleService(userRoleUseCase *biz.UserRoleUseCase) *UserRoleService {
	return &UserRoleService{
		userRoleUseCase: userRoleUseCase,
	}
}

// GetUserRoles 查询用户角色关联
func (s *UserRoleService) GetUserRoles(c *gin.Context) {
	var req user_role.GetUserRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userRoles, err := s.userRoleUseCase.GetUserRoles(c.Request.Context(), uint(req.UserId), uint(req.RoleId))
	if err != nil {
		logger.Error("查询用户角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "查询用户角色关联失败")
		return
	}

	// 转换为Proto格式
	userRoleProtos := make([]*user_role.UserRole, 0, len(userRoles))
	for _, ur := range userRoles {
		userRoleProtos = append(userRoleProtos, s.convertModelToProto(ur))
	}

	response := &user_role.GetUserRolesResponse{
		UserRoles: userRoleProtos,
	}

	common.GinSuccess(c, response)
}

// AddUserRole 添加用户角色关联
func (s *UserRoleService) AddUserRole(c *gin.Context) {
	var req user_role.AddUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userRoleModel, err := s.userRoleUseCase.AddUserRole(c.Request.Context(), uint(req.UserId), uint(req.RoleId))
	if err != nil {
		logger.Error("添加用户角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "添加用户角色关联失败")
		return
	}

	userRoleProto := s.convertModelToProto(userRoleModel)
	response := &user_role.AddUserRoleResponse{
		UserRole: userRoleProto,
	}

	common.GinSuccess(c, response)
}

// DeleteUserRole 删除用户角色关联
func (s *UserRoleService) DeleteUserRole(c *gin.Context) {
	var req user_role.DeleteUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	if err := s.userRoleUseCase.DeleteUserRole(c.Request.Context(), uint(req.UserId), uint(req.RoleId)); err != nil {
		logger.Error("删除用户角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除用户角色关联失败")
		return
	}

	response := &user_role.CommonResponse{
		Message: "删除成功",
	}

	common.GinSuccess(c, response)
}

// BatchAddUserRole 批量添加用户角色关联
func (s *UserRoleService) BatchAddUserRole(c *gin.Context) {
	var req user_role.BatchAddUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 先删除用户现有的角色关联
	if err := s.userRoleUseCase.DeleteByUserId(c.Request.Context(), uint(req.UserId)); err != nil {
		logger.Error("删除用户现有角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除用户现有角色关联失败")
		return
	}

	// 批量添加新的角色关联
	userRoleModels := make([]*model.SysUsersRoles, 0, len(req.RoleIds))
	for _, roleId := range req.RoleIds {
		userRoleModels = append(userRoleModels, &model.SysUsersRoles{
			UserID: uint(req.UserId),
			RoleID: uint(roleId),
		})
	}

	if err := s.userRoleUseCase.BatchAddUserRoles(c.Request.Context(), userRoleModels); err != nil {
		logger.Error("批量添加用户角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "批量添加用户角色关联失败")
		return
	}

	// 转换为Proto格式
	userRoleProtos := make([]*user_role.UserRole, 0, len(userRoleModels))
	for _, ur := range userRoleModels {
		userRoleProtos = append(userRoleProtos, s.convertModelToProto(ur))
	}

	response := &user_role.BatchAddUserRoleResponse{
		Data: userRoleProtos,
	}

	common.GinSuccess(c, response)
}

// GetRoleIdsByUserId 根据用户ID获取角色ID列表
func (s *UserRoleService) GetRoleIdsByUserId(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "用户ID格式错误")
		return
	}

	roleIds, err := s.userRoleUseCase.GetRoleIdsByUserId(c.Request.Context(), uint(userId))
	if err != nil {
		logger.Error("获取用户角色ID列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取用户角色ID列表失败")
		return
	}

	// 转换为uint64切片
	roleIdList := make([]uint64, 0, len(roleIds))
	for _, roleId := range roleIds {
		roleIdList = append(roleIdList, uint64(roleId))
	}

	response := &user_role.GetRoleIdsByUserIdResponse{
		Data: roleIdList,
	}

	common.GinSuccess(c, response)
}

// GetUserIdsByRoleId 根据角色ID获取用户ID列表
func (s *UserRoleService) GetUserIdsByRoleId(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "角色ID格式错误")
		return
	}

	userIds, err := s.userRoleUseCase.GetUserIdsByRoleId(c.Request.Context(), uint(roleId))
	if err != nil {
		logger.Error("获取角色用户ID列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "获取角色用户ID列表失败")
		return
	}

	// 转换为uint64切片
	userIdList := make([]uint64, 0, len(userIds))
	for _, userId := range userIds {
		userIdList = append(userIdList, uint64(userId))
	}

	response := &user_role.GetUserIdsByRoleIdResponse{
		Data: userIdList,
	}

	common.GinSuccess(c, response)
}

// DeleteByUserId 删除用户所有角色关联
func (s *UserRoleService) DeleteByUserId(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "用户ID格式错误")
		return
	}

	if err := s.userRoleUseCase.DeleteByUserId(c.Request.Context(), uint(userId)); err != nil {
		logger.Error("删除用户所有角色关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除用户所有角色关联失败")
		return
	}

	response := &user_role.CommonResponse{
		Message: "删除成功",
	}

	common.GinSuccess(c, response)
}

// DeleteByRoleId 删除角色所有用户关联
func (s *UserRoleService) DeleteByRoleId(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "角色ID格式错误")
		return
	}

	if err := s.userRoleUseCase.DeleteByRoleId(c.Request.Context(), uint(roleId)); err != nil {
		logger.Error("删除角色所有用户关联失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "删除角色所有用户关联失败")
		return
	}

	response := &user_role.CommonResponse{
		Message: "删除成功",
	}

	common.GinSuccess(c, response)
}

// convertModelToProto 将模型转换为Proto
func (s *UserRoleService) convertModelToProto(userRoleModel *model.SysUsersRoles) *user_role.UserRole {
	return &user_role.UserRole{
		UserId: uint64(userRoleModel.UserID),
		RoleId: uint64(userRoleModel.RoleID),
	}
}
