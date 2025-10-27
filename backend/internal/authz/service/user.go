package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/authz/user"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/internal/authz/config"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/utils"
)

// UserService 用户HTTP服务
type UserService struct {
	userBiz *biz.UserBiz
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		userBiz: biz.NewUserBiz(),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 转换请求为模型
	userModel := s.convertCreateRequestToModel(&req)

	// 如果提供了密码，进行哈希处理
	if req.Password != "" {
		if err := s.userBiz.SetUserPassword(c.Request.Context(), userModel, req.Password); err != nil {
			logger.Error("设置用户密码失败", zap.Error(err))
			common.GinError(c, i18nresp.CodeInternalError, "设置用户密码失败")
			return
		}
	}

	// 创建用户
	if err := s.userBiz.CreateUser(c.Request.Context(), userModel); err != nil {
		logger.Error("创建用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 返回创建的用户信息
	response := s.convertModelToProto(userModel)
	common.GinSuccess(c, response)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的用户ID")
		return
	}

	var req user.UpdateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 获取现有用户
	existingUser, err := s.userBiz.GetUserById(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 更新模型
	s.updateModelFromRequest(existingUser, &req)

	// 更新用户
	if err := s.userBiz.UpdateUser(c.Request.Context(), existingUser); err != nil {
		logger.Error("更新用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 返回更新后的用户信息
	userProto := s.convertModelToProto(existingUser)
	common.GinSuccess(c, userProto)
}

// GetUserById 根据ID获取用户
func (s *UserService) GetUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的用户ID")
		return
	}

	userModel, err := s.userBiz.GetUserById(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("获取用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	userProto := s.convertModelToProto(userModel)
	common.GinSuccess(c, userProto)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "无效的用户ID")
		return
	}

	if err := s.userBiz.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("删除用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, gin.H{"message": "用户删除成功"})
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(c *gin.Context) {
	var req user.ListUsersRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 设置默认分页参数
	if req.PageInfo.Page <= 0 {
		req.PageInfo.Page = 1
	}
	if req.PageInfo.Size <= 0 {
		req.PageInfo.Size = 10
	}
	if req.PageInfo.Size > 100 {
		req.PageInfo.Size = 100
	}

	// 构建查询参数
	params := &biz.ListUsersParams{
		Page:     int(req.PageInfo.Page),
		PageSize: int(req.PageInfo.Size),
		Keyword:  req.Query.Blurry,
		DeptId:   uint(req.Query.DeptId),
	}

	// 处理状态参数
	switch req.Query.Status {
	case user.UserStatus_UserStatusEnabled:
		enabled := true
		params.Enabled = &enabled
	case user.UserStatus_UserStatusDisabled:
		enabled := false
		params.Enabled = &enabled
	}

	// 获取用户列表
	users, total, err := s.userBiz.ListUsers(c.Request.Context(), params)
	if err != nil {
		logger.Error("获取用户列表失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 转换为响应格式
	userProtos := make([]*user.SysUser, len(users))
	for i, u := range users {
		userProtos[i] = s.convertModelToProto(u)
	}

	response := &user.ListUsersResponse{
		Data: &user.PageSysUser{
			Users: userProtos,
			PageInfo: &user.PageInfo{
				Page:  req.PageInfo.Page,
				Size:  req.PageInfo.Size,
				Total: total,
			},
		},
	}

	common.GinSuccess(c, response)
}

// convertCreateRequestToModel 将创建请求转换为模型
func (s *UserService) convertCreateRequestToModel(req *user.CreateUserRequest) *model.SysUser {
	userModel := &model.SysUser{}
	userModel.SetUsername(req.Username)
	userModel.SetNickName(req.FullName)
	userModel.SetEmail(req.Email)
	userModel.SetPhone(req.Phone)

	if req.DeptId > 0 {
		deptId := uint(req.DeptId)
		userModel.DeptID = &deptId
	}

	if req.Status == user.UserStatus_UserStatusEnabled {
		userModel.SetEnabled(true)
	} else {
		userModel.SetEnabled(false)
	}

	// 处理密码
	if req.Password != "" {
		// 这里应该调用密码加密逻辑
		userModel.SetPassword(req.Password) // 临时直接设置，实际应该加密
	}

	return userModel
}

// updateModelFromRequest 从更新请求更新模型
func (s *UserService) updateModelFromRequest(userModel *model.SysUser, req *user.UpdateUserRequest) {
	if req.FullName != "" {
		userModel.SetNickName(req.FullName)
	}
	if req.Email != "" {
		userModel.SetEmail(req.Email)
	}
	if req.Phone != "" {
		userModel.SetPhone(req.Phone)
	}
	if req.DeptId > 0 {
		deptId := uint(req.DeptId)
		userModel.DeptID = &deptId
	}
	switch req.Status {
	case user.UserStatus_UserStatusEnabled:
		userModel.SetEnabled(true)
	case user.UserStatus_UserStatusDisabled:
		userModel.SetEnabled(false)
	}
}

// GetCurrentUser 获取当前用户信息
func (s *UserService) GetCurrentUser(c *gin.Context) {
	// TODO: 从JWT token中获取用户ID
	// 这里暂时返回错误，需要实现JWT中间件后才能获取当前用户
	common.GinError(c, i18nresp.CodeInternalError, "获取当前用户功能暂未实现")
}

// GetUserPermissions 获取用户权限
func (s *UserService) GetUserPermissions(c *gin.Context) {
	var req user.GetUserPermissionsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// TODO: 实现获取用户权限逻辑
	_ = req.UserId
	permissions := []*user.PermissionTreeNode{}

	response := &user.GetUserPermissionsResponse{
		Permissions: permissions,
	}

	common.GinSuccess(c, response)
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(c *gin.Context) {
	var req user.UpdatePasswordRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userId := c.GetInt64("userId")

	// 基本参数验证
	if userId <= 0 {
		common.GinError(c, i18nresp.CodeUserIDInvalid, "")
		return
	}

	if req.OldPassword == "" {
		common.GinError(c, i18nresp.CodeOldPasswordEmpty, "")
		return
	}

	if req.NewPassword == "" {
		common.GinError(c, i18nresp.CodeNewPasswordEmpty, "")
		return
	}

	if req.ConfirmPassword == "" {
		common.GinError(c, i18nresp.CodeConfirmPasswordEmpty, "")
		return
	}

	// 验证新密码和确认密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		common.GinError(c, i18nresp.CodePasswordMismatch, "")
		return
	}

	// 验证新密码强度
	if isValid, errorCode := common.ValidatePasswordStrengthWithI18n(req.NewPassword); !isValid {
		common.GinError(c, errorCode, "")
		return
	}

	// 验证新密码不能与旧密码相同
	if req.OldPassword == req.NewPassword {
		common.GinError(c, i18nresp.CodePasswordSameAsOld, "")
		return
	}

	// 使用业务层的UpdatePassword方法进行密码更新
	if err := s.userBiz.UpdatePassword(c.Request.Context(), uint(userId), req.OldPassword, req.NewPassword); err != nil {
		logger.Error("更新密码失败", zap.Error(err), zap.Int64("userId", userId))
		// 根据错误类型返回相应的错误码
		if strings.Contains(err.Error(), "旧密码不正确") || strings.Contains(err.Error(), "old password") {
			common.GinError(c, i18nresp.CodeOldPasswordIncorrect, "")
		} else if strings.Contains(err.Error(), "用户不存在") || strings.Contains(err.Error(), "user not found") {
			common.GinError(c, i18nresp.CodeUserNotFoundError, "")
		} else {
			common.GinError(c, i18nresp.CodeUpdatePasswordFailure, "")
		}
		return
	}

	response := &user.UpdatePasswordResponse{}
	common.GinSuccess(c, response)
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(c *gin.Context) {
	userId := c.GetInt64("userId")
	if userId <= 0 {
		common.GinError(c, i18nresp.CodeUserIDInvalid, "")
		return
	}
	// 获取上传的文件
	imageFile, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "No image file provided")
		return
	}

	// 打开文件
	file, err := imageFile.Open()
	if err != nil {
		logger.Error("Failed to open image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to open image file")
		return
	}
	defer file.Close()

	// 读取文件内容
	imageData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to read image file")
		return
	}

	// 验证图片类型
	if !utils.IsValidImageType(imageData) {
		logger.Error("Invalid image type")
		common.GinError(c, i18nresp.CodeInternalError, "Unsupported image type")
		return
	}

	// 验证文件大小 (5MB限制)
	maxSize := int64(5 * 1024 * 1024)
	if int64(len(imageData)) > maxSize {
		logger.Error("Image file too large", zap.Int("size", len(imageData)))
		common.GinError(c, i18nresp.CodeInternalError, "Image file too large")
		return
	}

	// 获取用户信息
	userModel, err := s.userBiz.GetUserById(c.Request.Context(), uint(userId))
	if err != nil {
		logger.Error("获取用户失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// 生成文件名
	ext := utils.GetImageFileExtension(imageData)
	if ext == "" {
		ext = "jpg"
	}
	fileName := fmt.Sprintf("%d.%s", userId, ext)
	// 构建存储路径
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, strings.Trim(common.AvatarPath, "/"))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create storage directory")
		return
	}
	// 保存文件
	filePath := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		logger.Error("Failed to save image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to save image file")
		return
	}

	// 更新头像路径
	imagePath := filepath.Join(common.StaticPrefix, common.AvatarPath, fileName)
	userModel.AvatarPath = &imagePath

	if err := s.userBiz.UpdateUser(c.Request.Context(), userModel); err != nil {
		logger.Error("更新头像失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	resp := &user.UpdateAvatarResponse{
		Path: imagePath,
		Size: int64(len(imageData)),
		Mime: fmt.Sprintf("image/%s", ext),
	}
	common.GinSuccess(c, resp)
}

// convertModelToProto 将模型转换为Proto
func (s *UserService) convertModelToProto(userModel *model.SysUser) *user.SysUser {
	userProto := &user.SysUser{
		Id:       int64(userModel.UserID),
		Username: userModel.GetUsername(),
		FullName: userModel.GetNickName(),
		Email:    userModel.GetEmail(),
		Phone:    userModel.GetPhone(),
	}

	if userModel.DeptID != nil {
		userProto.DeptId = int64(*userModel.DeptID)
	}

	if userModel.IsEnabled() {
		userProto.Status = user.UserStatus_UserStatusEnabled
	} else {
		userProto.Status = user.UserStatus_UserStatusDisabled
	}

	if userModel.CreateTime != nil {
		userProto.CreatedAt = userModel.CreateTime.Unix()
	}
	if userModel.UpdateTime != nil {
		userProto.UpdatedAt = userModel.UpdateTime.Unix()
	}

	return userProto
}
