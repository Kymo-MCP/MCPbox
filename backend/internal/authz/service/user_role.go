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

// UserRoleService user role service
type UserRoleService struct {
	userRoleUseCase *biz.UserRoleUseCase
}

// NewUserRoleService creates user role service instance
func NewUserRoleService(userRoleUseCase *biz.UserRoleUseCase) *UserRoleService {
	return &UserRoleService{
		userRoleUseCase: userRoleUseCase,
	}
}

// GetUserRoles queries user role associations
func (s *UserRoleService) GetUserRoles(c *gin.Context) {
	var req user_role.GetUserRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userRoles, err := s.userRoleUseCase.GetUserRoles(c.Request.Context(), uint(req.UserId), uint(req.RoleId))
	if err != nil {
		logger.Error("query user role associations failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "query user role associations failed")
		return
	}

	// Convert to Proto format
	userRoleProtos := make([]*user_role.UserRole, 0, len(userRoles))
	for _, ur := range userRoles {
		userRoleProtos = append(userRoleProtos, s.convertModelToProto(ur))
	}

	response := &user_role.GetUserRolesResponse{
		UserRoles: userRoleProtos,
	}

	common.GinSuccess(c, response)
}

// AddUserRole adds user role association
func (s *UserRoleService) AddUserRole(c *gin.Context) {
	var req user_role.AddUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userRoleModel, err := s.userRoleUseCase.AddUserRole(c.Request.Context(), uint(req.UserId), uint(req.RoleId))
	if err != nil {
		logger.Error("add user role association failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "add user role association failed")
		return
	}

	userRoleProto := s.convertModelToProto(userRoleModel)
	response := &user_role.AddUserRoleResponse{
		UserRole: userRoleProto,
	}

	common.GinSuccess(c, response)
}

// DeleteUserRole deletes user role association
func (s *UserRoleService) DeleteUserRole(c *gin.Context) {
	var req user_role.DeleteUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	if err := s.userRoleUseCase.DeleteUserRole(c.Request.Context(), uint(req.UserId), uint(req.RoleId)); err != nil {
		logger.Error("delete user role association failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "delete user role association failed")
		return
	}

	response := &user_role.CommonResponse{
		Message: "delete successful",
	}

	common.GinSuccess(c, response)
}

// BatchAddUserRole batch adds user role associations
func (s *UserRoleService) BatchAddUserRole(c *gin.Context) {
	var req user_role.BatchAddUserRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Delete existing role associations of user first
	if err := s.userRoleUseCase.DeleteByUserId(c.Request.Context(), uint(req.UserId)); err != nil {
		logger.Error("delete existing role associations of user failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "delete existing role associations of user failed")
		return
	}

	// Batch add new role associations
	userRoleModels := make([]*model.SysUsersRoles, 0, len(req.RoleIds))
	for _, roleId := range req.RoleIds {
		userRoleModels = append(userRoleModels, &model.SysUsersRoles{
			UserID: uint(req.UserId),
			RoleID: uint(roleId),
		})
	}

	if err := s.userRoleUseCase.BatchAddUserRoles(c.Request.Context(), userRoleModels); err != nil {
		logger.Error("batch add user role associations failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "batch add user role associations failed")
		return
	}

	// Convert to Proto format
	userRoleProtos := make([]*user_role.UserRole, 0, len(userRoleModels))
	for _, ur := range userRoleModels {
		userRoleProtos = append(userRoleProtos, s.convertModelToProto(ur))
	}

	response := &user_role.BatchAddUserRoleResponse{
		Data: userRoleProtos,
	}

	common.GinSuccess(c, response)
}

// GetRoleIdsByUserId gets role ID list by user ID
func (s *UserRoleService) GetRoleIdsByUserId(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "user ID format error")
		return
	}

	roleIds, err := s.userRoleUseCase.GetRoleIdsByUserId(c.Request.Context(), uint(userId))
	if err != nil {
		logger.Error("get user role ID list failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "get user role ID list failed")
		return
	}

	// Convert to uint64 slice
	roleIdList := make([]uint64, 0, len(roleIds))
	for _, roleId := range roleIds {
		roleIdList = append(roleIdList, uint64(roleId))
	}

	response := &user_role.GetRoleIdsByUserIdResponse{
		Data: roleIdList,
	}

	common.GinSuccess(c, response)
}

// GetUserIdsByRoleId gets user ID list by role ID
func (s *UserRoleService) GetUserIdsByRoleId(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "role ID format error")
		return
	}

	userIds, err := s.userRoleUseCase.GetUserIdsByRoleId(c.Request.Context(), uint(roleId))
	if err != nil {
		logger.Error("get role user ID list failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "get role user ID list failed")
		return
	}

	// Convert to uint64 slice
	userIdList := make([]uint64, 0, len(userIds))
	for _, userId := range userIds {
		userIdList = append(userIdList, uint64(userId))
	}

	response := &user_role.GetUserIdsByRoleIdResponse{
		Data: userIdList,
	}

	common.GinSuccess(c, response)
}

// DeleteByUserId deletes all role associations of user
func (s *UserRoleService) DeleteByUserId(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "user ID format error")
		return
	}

	if err := s.userRoleUseCase.DeleteByUserId(c.Request.Context(), uint(userId)); err != nil {
		logger.Error("delete all role associations of user failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "delete all role associations of user failed")
		return
	}

	response := &user_role.CommonResponse{
		Message: "delete successful",
	}

	common.GinSuccess(c, response)
}

// DeleteByRoleId deletes all user associations of role
func (s *UserRoleService) DeleteByRoleId(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "role ID format error")
		return
	}

	if err := s.userRoleUseCase.DeleteByRoleId(c.Request.Context(), uint(roleId)); err != nil {
		logger.Error("delete all user associations of role failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "delete all user associations of role failed")
		return
	}

	response := &user_role.CommonResponse{
		Message: "delete successful",
	}

	common.GinSuccess(c, response)
}

// convertModelToProto converts model to Proto
func (s *UserRoleService) convertModelToProto(userRoleModel *model.SysUsersRoles) *user_role.UserRole {
	return &user_role.UserRole{
		UserId: uint64(userRoleModel.UserID),
		RoleId: uint64(userRoleModel.RoleID),
	}
}
