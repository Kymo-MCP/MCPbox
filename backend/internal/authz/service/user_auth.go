package service

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/authz/user_auth"
	"qm-mcp-server/internal/authz/biz"
	"qm-mcp-server/pkg/common"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/redis"
	"qm-mcp-server/pkg/utils"
)

// UserAuthService 用户认证HTTP服务
type UserAuthService struct {
	authUseCase *biz.AuthUseCase
	userBiz     *biz.UserBiz
	logger      zap.Logger
}

// NewUserAuthService 创建用户认证服务实例
func NewUserAuthService() *UserAuthService {
	return &UserAuthService{
		authUseCase: biz.NewAuthUseCase(),
		userBiz:     biz.NewUserBiz(),
		logger:      *logger.L().Logger,
	}
}

// Login 用户登录
func (s *UserAuthService) Login(c *gin.Context) {
	var req user_auth.LoginRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 获取客户端IP和User-Agent
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 解密密码
	var plainPassword string
	var err error

	if req.KeyId != "" && req.EncryptedPassword != "" {
		// 使用RSA解密密码
		plainPassword, err = s.decryptPasswordWithRSA(req.KeyId, req.EncryptedPassword)
		if err != nil {
			s.logger.Error("RSA解密密码失败", zap.Error(err), zap.String("keyId", req.KeyId))
			common.GinError(c, i18nresp.CodeInternalError, "用户名或密码错误")
			return
		}
		s.logger.Info("成功解密登录密码", zap.String("keyId", req.KeyId), zap.String("username", req.Username))
	} else {
		s.logger.Error("缺少密钥ID或加密密码", zap.String("username", req.Username))
		common.GinError(c, i18nresp.CodeInternalError, "缺少必要的加密参数")
		return
	}

	// 执行登录
	loginData, err := s.authUseCase.Login(
		c.Request.Context(),
		req.Username,
		plainPassword,
		req.Timestamp,
		clientIP,
		userAgent,
	)
	if err != nil {
		logger.Error("用户登录失败", zap.Error(err), zap.String("username", req.Username))
		common.GinError(c, i18nresp.CodeInternalError, "登录失败: "+err.Error())
		return
	}

	// 转换响应数据
	response := &user_auth.LoginResponse{
		Token:        loginData.Token,
		RefreshToken: loginData.RefreshToken,
		ExpiresIn:    common.AccessTokenExpireTime,
		UserInfo: &user_auth.UserInfo{
			UserId:    loginData.UserInfo.UserID,
			Username:  loginData.UserInfo.Username,
			Nickname:  loginData.UserInfo.Nickname,
			Email:     loginData.UserInfo.Email,
			Phone:     loginData.UserInfo.Phone,
			Avatar:    loginData.UserInfo.Avatar,
			DeptId:    loginData.UserInfo.DeptID,
			DeptName:  loginData.UserInfo.DeptName,
			RoleIds:   s.convertUintToInt64Slice(loginData.UserInfo.RoleIDs),
			RoleNames: loginData.UserInfo.RoleNames,
		},
	}

	common.GinSuccess(c, response)
}

// Logout 用户登出
func (s *UserAuthService) Logout(c *gin.Context) {
	var req user_auth.LogoutRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 执行登出
	if err := s.authUseCase.Logout(c.Request.Context(), req.UserId, req.Token); err != nil {
		logger.Error("用户登出失败", zap.Error(err), zap.Int64("userId", req.UserId))
		common.GinError(c, i18nresp.CodeInternalError, "登出失败: "+err.Error())
		return
	}

	common.GinSuccess(c, nil)
}

// RefreshToken 刷新Token
func (s *UserAuthService) RefreshToken(c *gin.Context) {
	var req user_auth.RefreshTokenRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 执行Token刷新
	tokenData, err := s.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		logger.Error("刷新Token失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "刷新Token失败: "+err.Error())
		return
	}

	// 转换响应数据
	response := &user_auth.RefreshTokenResponse{
		Token:        tokenData.Token,
		RefreshToken: tokenData.RefreshToken,
		ExpiresIn:    common.AccessTokenExpireTime,
	}

	common.GinSuccess(c, response)
}

// ValidateToken 校验Token
func (s *UserAuthService) ValidateToken(c *gin.Context) {
	var req user_auth.ValidateTokenRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 执行Token校验
	validateResult, err := s.authUseCase.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		logger.Error("校验Token失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "校验Token失败: "+err.Error())
		return
	}

	// 转换响应数据
	response := &user_auth.ValidateTokenResponse{
		Valid: validateResult.Valid,
	}

	if validateResult.Valid && validateResult.UserInfo != nil {
		response.UserInfo = &user_auth.UserInfo{
			UserId:    validateResult.UserInfo.UserID,
			Username:  validateResult.UserInfo.Username,
			Nickname:  validateResult.UserInfo.Nickname,
			Email:     validateResult.UserInfo.Email,
			Phone:     validateResult.UserInfo.Phone,
			Avatar:    validateResult.UserInfo.Avatar,
			DeptId:    validateResult.UserInfo.DeptID,
			DeptName:  validateResult.UserInfo.DeptName,
			RoleIds:   s.convertUintToInt64Slice(validateResult.UserInfo.RoleIDs),
			RoleNames: validateResult.UserInfo.RoleNames,
		}
	}

	if validateResult.Valid && validateResult.LoginInfo != nil {
		response.LoginInfo = &user_auth.LoginInfo{
			LoginTime: validateResult.LoginInfo.LoginTime.Unix(),
			LoginIp:   validateResult.LoginInfo.LoginIP,
			UserAgent: validateResult.LoginInfo.UserAgent,
			ExpiresAt: validateResult.LoginInfo.ExpiresAt.Unix(),
		}
	}

	common.GinSuccess(c, response)
}

// GetUserInfo 获取用户信息
func (s *UserAuthService) GetUserInfo(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		common.GinError(c, i18nresp.CodeInternalError, "未获取到用户ID")
		return
	}
	userIdInt, ok := userId.(int64)
	if !ok {
		common.GinError(c, i18nresp.CodeInternalError, "用户ID类型错误")
		return
	}
	// 获取用户信息
	userInfo, err := s.authUseCase.GetUserInfo(c.Request.Context(), uint(userIdInt))
	if err != nil {
		logger.Error("获取用户信息失败", zap.Error(err), zap.Int64("userId", int64(userIdInt)))
		common.GinError(c, i18nresp.CodeInternalError, "获取用户信息失败")
		return
	}
	if userInfo == nil {
		common.GinError(c, i18nresp.CodeInternalError, "未获取到用户信息")
		return
	}

	// 返回默认配置
	response := &user_auth.GetUserInfoResponse{
		TokenExpiry:        common.AccessTokenExpireTime,
		RefreshTokenExpiry: common.RefreshTokenExpireTime,
		Theme:              common.DefaultTheme,
		Language:           common.DefaultLanguage,
		PageSize:           common.DefaultPageSize,
		EnableNotification: common.EnableNotification,
		AutoLogout:         common.AutoLogoutTime,
		UserInfo: &user_auth.UserInfo{
			UserId:    userInfo.UserID,
			Username:  userInfo.Username,
			Nickname:  userInfo.Nickname,
			Email:     userInfo.Email,
			Phone:     userInfo.Phone,
			Avatar:    userInfo.Avatar,
			DeptId:    userInfo.DeptID,
			DeptName:  userInfo.DeptName,
			RoleIds:   s.convertUintToInt64Slice(userInfo.RoleIDs),
			RoleNames: userInfo.RoleNames,
		},
	}

	common.GinSuccess(c, response)
}

// GetEncryptionKey 获取加密密钥
func (s *UserAuthService) GetEncryptionKey(c *gin.Context) {
	var req user_auth.GetEncryptionKeyRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 使用pkg/utils生成RSA密钥对
	keyPair, err := utils.GenerateRSAKeyPair()
	if err != nil {
		s.logger.Error("生成RSA密钥对失败", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "生成加密密钥失败")
		return
	}

	// 构造返回数据
	response := &user_auth.GetEncryptionKeyResponse{
		KeyId:     keyPair.KeyID,
		PublicKey: utils.GetPublicKeyBase64(keyPair.PublicKey),
		Algorithm: utils.AlgorithmRSA2048,
		ExpiresAt: keyPair.ExpiresAt.Unix(),
		IssuedAt:  keyPair.IssuedAt.Unix(),
	}

	// 保存私钥到Redis用于解密
	if privateErr := redis.SetEncryptionPrivateKey(keyPair.KeyID, keyPair.PrivateKey); privateErr != nil {
		s.logger.Error("保存私钥到Redis失败", zap.Error(privateErr), zap.String("keyId", keyPair.KeyID))
		common.GinError(c, i18nresp.CodeInternalError, "保存加密密钥失败")
		return
	}

	// 记录密钥生成日志
	s.logger.Info("成功生成新的加密密钥",
		zap.String("keyId", keyPair.KeyID),
		zap.String("algorithm", utils.AlgorithmRSA2048),
		zap.Time("expiresAt", keyPair.ExpiresAt))

	common.GinSuccess(c, response)
}

// convertUintToInt64Slice 转换uint切片为int64切片
func (s *UserAuthService) convertUintToInt64Slice(uintSlice []uint) []int64 {
	int64Slice := make([]int64, len(uintSlice))
	for i, v := range uintSlice {
		int64Slice[i] = int64(v)
	}
	return int64Slice
}

// decryptPasswordWithRSA 使用RSA私钥解密密码
func (s *UserAuthService) decryptPasswordWithRSA(keyID, encryptedPassword string) (string, error) {
	// 从Redis获取私钥
	privateKeyPEM, err := redis.GetEncryptionPrivateKey(keyID)
	if err != nil {
		return "", fmt.Errorf("获取私钥失败: %v", err)
	}

	// base64 decode
	password, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 使用pkg/utils解密密码
	decryptedBytes, err := utils.RSADecrypt(string(password), privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}

	return string(decryptedBytes), nil
}
