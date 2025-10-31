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

	"github.com/kymo-mcp/mcpcan/api/authz/user"
	"github.com/kymo-mcp/mcpcan/internal/authz/biz"
	"github.com/kymo-mcp/mcpcan/internal/authz/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

// UserService user HTTP service
type UserService struct {
	userBiz *biz.UserBiz
}

// NewUserService creates user service instance
func NewUserService() *UserService {
	return &UserService{
		userBiz: biz.NewUserBiz(),
	}
}

// CreateUser creates user
func (s *UserService) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Convert request to model
	userModel := s.convertCreateRequestToModel(&req)

	// If password is provided, hash it
	if req.Password != "" {
		if err := s.userBiz.SetUserPassword(c.Request.Context(), userModel, req.Password); err != nil {
			logger.Error("Failed to set user password", zap.Error(err))
			common.GinError(c, i18nresp.CodeInternalError, "Failed to set user password")
			return
		}
	}

	// Create user
	if err := s.userBiz.CreateUser(c.Request.Context(), userModel); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Return created user information
	response := s.convertModelToProto(userModel)
	common.GinSuccess(c, response)
}

// UpdateUser updates user
func (s *UserService) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid user ID")
		return
	}

	var req user.UpdateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get existing user
	existingUser, err := s.userBiz.GetUserById(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Update model
	s.updateModelFromRequest(existingUser, &req)

	// Update user
	if err := s.userBiz.UpdateUser(c.Request.Context(), existingUser); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Return updated user information
	userProto := s.convertModelToProto(existingUser)
	common.GinSuccess(c, userProto)
}

// GetUserById gets user by ID
func (s *UserService) GetUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid user ID")
		return
	}

	userModel, err := s.userBiz.GetUserById(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	userProto := s.convertModelToProto(userModel)
	common.GinSuccess(c, userProto)
}

// DeleteUser deletes user
func (s *UserService) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid user ID")
		return
	}

	if err := s.userBiz.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, gin.H{"message": "User deleted successfully"})
}

// ListUsers gets user list
func (s *UserService) ListUsers(c *gin.Context) {
	var req user.ListUsersRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Set default pagination parameters
	if req.PageInfo.Page <= 0 {
		req.PageInfo.Page = 1
	}
	if req.PageInfo.Size <= 0 {
		req.PageInfo.Size = 10
	}
	if req.PageInfo.Size > 100 {
		req.PageInfo.Size = 100
	}

	// Build query parameters
	params := &biz.ListUsersParams{
		Page:     int(req.PageInfo.Page),
		PageSize: int(req.PageInfo.Size),
		Keyword:  req.Query.Blurry,
		DeptId:   uint(req.Query.DeptId),
	}

	// Handle status parameter
	switch req.Query.Status {
	case user.UserStatus_UserStatusEnabled:
		enabled := true
		params.Enabled = &enabled
	case user.UserStatus_UserStatusDisabled:
		enabled := false
		params.Enabled = &enabled
	}

	// Get user list
	users, total, err := s.userBiz.ListUsers(c.Request.Context(), params)
	if err != nil {
		logger.Error("Failed to get user list", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Convert to response format
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

// convertCreateRequestToModel converts create request to model
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

	// Handle password
	if req.Password != "" {
		// This should call password encryption logic
		userModel.SetPassword(req.Password) // Temporary direct setting, should be encrypted in practice
	}

	return userModel
}

// updateModelFromRequest updates model from update request
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

// GetCurrentUser gets current user information
func (s *UserService) GetCurrentUser(c *gin.Context) {
	// TODO: Get user ID from JWT token
	// This temporarily returns an error, need to implement JWT middleware to get current user
	common.GinError(c, i18nresp.CodeInternalError, "Get current user feature not implemented yet")
}

// GetUserPermissions gets user permissions
func (s *UserService) GetUserPermissions(c *gin.Context) {
	var req user.GetUserPermissionsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// TODO: Implement get user permissions logic
	_ = req.UserId
	permissions := []*user.PermissionTreeNode{}

	response := &user.GetUserPermissionsResponse{
		Permissions: permissions,
	}

	common.GinSuccess(c, response)
}

// UpdatePassword updates user password
func (s *UserService) UpdatePassword(c *gin.Context) {
	var req user.UpdatePasswordRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userId := c.GetInt64("userId")

	// Basic parameter validation
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

	// Verify new password and confirm password match
	if req.NewPassword != req.ConfirmPassword {
		common.GinError(c, i18nresp.CodePasswordMismatch, "")
		return
	}

	// Verify new password strength
	if isValid, errorCode := common.ValidatePasswordStrengthWithI18n(req.NewPassword); !isValid {
		common.GinError(c, errorCode, "")
		return
	}

	// Verify new password can't be the same as old password
	if req.OldPassword == req.NewPassword {
		common.GinError(c, i18nresp.CodePasswordSameAsOld, "")
		return
	}

	// Use business layer's UpdatePassword method for password update
	if err := s.userBiz.UpdatePassword(c.Request.Context(), uint(userId), req.OldPassword, req.NewPassword); err != nil {
		logger.Error("Failed to update password", zap.Error(err), zap.Int64("userId", userId))
		// Return corresponding error code based on error type
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

// UpdateAvatar updates user avatar
func (s *UserService) UpdateAvatar(c *gin.Context) {
	userId := c.GetInt64("userId")
	if userId <= 0 {
		common.GinError(c, i18nresp.CodeUserIDInvalid, "")
		return
	}
	// Get uploaded file
	imageFile, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "No image file provided")
		return
	}

	// Open file
	file, err := imageFile.Open()
	if err != nil {
		logger.Error("Failed to open image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to open image file")
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to read image file")
		return
	}

	// Validate image type
	if !utils.IsValidImageType(imageData) {
		logger.Error("Invalid image type")
		common.GinError(c, i18nresp.CodeInternalError, "Unsupported image type")
		return
	}

	// Validate file size (5MB limit)
	maxSize := int64(5 * 1024 * 1024)
	if int64(len(imageData)) > maxSize {
		logger.Error("Image file too large", zap.Int("size", len(imageData)))
		common.GinError(c, i18nresp.CodeInternalError, "Image file too large")
		return
	}

	// Get user information
	userModel, err := s.userBiz.GetUserById(c.Request.Context(), uint(userId))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Generate file name
	ext := utils.GetImageFileExtension(imageData)
	if ext == "" {
		ext = "jpg"
	}
	fileName := fmt.Sprintf("%d.%s", userId, ext)
	// Build storage path
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, strings.Trim(common.AvatarPath, "/"))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create storage directory")
		return
	}
	// Save file
	filePath := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		logger.Error("Failed to save image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to save image file")
		return
	}

	// Update avatar path
	imagePath := filepath.Join(common.StaticPrefix, common.AvatarPath, fileName)
	userModel.AvatarPath = &imagePath

	if err := s.userBiz.UpdateUser(c.Request.Context(), userModel); err != nil {
		logger.Error("Failed to update avatar", zap.Error(err))
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

// convertModelToProto converts model to Proto
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
