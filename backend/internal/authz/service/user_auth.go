package service

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/api/authz/user_auth"
	"github.com/kymo-mcp/mcpcan/internal/authz/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

// UserAuthService user authentication HTTP service
type UserAuthService struct {
	authUseCase *biz.AuthUseCase
	userBiz     *biz.UserBiz
	logger      zap.Logger
}

// NewUserAuthService creates user authentication service instance
func NewUserAuthService() *UserAuthService {
	return &UserAuthService{
		authUseCase: biz.NewAuthUseCase(),
		userBiz:     biz.NewUserBiz(),
		logger:      *logger.L().Logger,
	}
}

// Login user login
func (s *UserAuthService) Login(c *gin.Context) {
	var req user_auth.LoginRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get client IP and User-Agent
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Decrypt password
	var plainPassword string
	var err error

	if req.KeyId != "" && req.EncryptedPassword != "" {
		// Use RSA to decrypt password
		plainPassword, err = s.decryptPasswordWithRSA(req.KeyId, req.EncryptedPassword)
		if err != nil {
			s.logger.Error("RSA password decryption failed", zap.Error(err), zap.String("keyId", req.KeyId))
			common.GinError(c, i18nresp.CodeInternalError, "username or password incorrect")
			return
		}
		s.logger.Info("successfully decrypted login password", zap.String("keyId", req.KeyId), zap.String("username", req.Username))
	} else {
		s.logger.Error("missing key ID or encrypted password", zap.String("username", req.Username))
		common.GinError(c, i18nresp.CodeInternalError, "missing required encryption parameters")
		return
	}

	// Execute login
	loginData, err := s.authUseCase.Login(
		c.Request.Context(),
		req.Username,
		plainPassword,
		req.Timestamp,
		clientIP,
		userAgent,
	)
	if err != nil {
		logger.Error("user login failed", zap.Error(err), zap.String("username", req.Username))
		common.GinError(c, i18nresp.CodeInternalError, "login failed: "+err.Error())
		return
	}

	// Convert response data
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

// Logout user logout
func (s *UserAuthService) Logout(c *gin.Context) {
	var req user_auth.LogoutRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Execute logout
	if err := s.authUseCase.Logout(c.Request.Context(), req.UserId, req.Token); err != nil {
		logger.Error("user logout failed", zap.Error(err), zap.Int64("userId", req.UserId))
		common.GinError(c, i18nresp.CodeInternalError, "logout failed: "+err.Error())
		return
	}

	common.GinSuccess(c, nil)
}

// RefreshToken refresh token
func (s *UserAuthService) RefreshToken(c *gin.Context) {
	var req user_auth.RefreshTokenRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Execute token refresh
	tokenData, err := s.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		logger.Error("refresh token failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "refresh token failed: "+err.Error())
		return
	}

	// Convert response data
	response := &user_auth.RefreshTokenResponse{
		Token:        tokenData.Token,
		RefreshToken: tokenData.RefreshToken,
		ExpiresIn:    common.AccessTokenExpireTime,
	}

	common.GinSuccess(c, response)
}

// ValidateToken validate token
func (s *UserAuthService) ValidateToken(c *gin.Context) {
	var req user_auth.ValidateTokenRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Execute token validation
	validateResult, err := s.authUseCase.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		logger.Error("validate token failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "validate token failed: "+err.Error())
		return
	}

	// Convert response data
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

// GetUserInfo get user information
func (s *UserAuthService) GetUserInfo(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		common.GinError(c, i18nresp.CodeInternalError, "failed to get user ID")
		return
	}
	userIdInt, ok := userId.(int64)
	if !ok {
		common.GinError(c, i18nresp.CodeInternalError, "user ID type error")
		return
	}
	// Get user information
	userInfo, err := s.authUseCase.GetUserInfo(c.Request.Context(), uint(userIdInt))
	if err != nil {
		logger.Error("get user information failed", zap.Error(err), zap.Int64("userId", int64(userIdInt)))
		common.GinError(c, i18nresp.CodeInternalError, "get user information failed")
		return
	}
	if userInfo == nil {
		common.GinError(c, i18nresp.CodeInternalError, "failed to get user information")
		return
	}

	// Return default configuration
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

// GetEncryptionKey get encryption key
func (s *UserAuthService) GetEncryptionKey(c *gin.Context) {
	var req user_auth.GetEncryptionKeyRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Use pkg/utils to generate RSA key pair
	keyPair, err := utils.GenerateRSAKeyPair()
	if err != nil {
		s.logger.Error("generate RSA key pair failed", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "generate encryption key failed")
		return
	}

	// Construct return data
	response := &user_auth.GetEncryptionKeyResponse{
		KeyId:     keyPair.KeyID,
		PublicKey: utils.GetPublicKeyBase64(keyPair.PublicKey),
		Algorithm: utils.AlgorithmRSA2048,
		ExpiresAt: keyPair.ExpiresAt.Unix(),
		IssuedAt:  keyPair.IssuedAt.Unix(),
	}

	// Save private key to Redis for decryption
	if privateErr := redis.SetEncryptionPrivateKey(keyPair.KeyID, keyPair.PrivateKey); privateErr != nil {
		s.logger.Error("save private key to Redis failed", zap.Error(privateErr), zap.String("keyId", keyPair.KeyID))
		common.GinError(c, i18nresp.CodeInternalError, "save encryption key failed")
		return
	}

	// Log key generation
	s.logger.Info("successfully generated new encryption key",
		zap.String("keyId", keyPair.KeyID),
		zap.String("algorithm", utils.AlgorithmRSA2048),
		zap.Time("expiresAt", keyPair.ExpiresAt))

	common.GinSuccess(c, response)
}

// convertUintToInt64Slice convert uint slice to int64 slice
func (s *UserAuthService) convertUintToInt64Slice(uintSlice []uint) []int64 {
	int64Slice := make([]int64, len(uintSlice))
	for i, v := range uintSlice {
		int64Slice[i] = int64(v)
	}
	return int64Slice
}

// decryptPasswordWithRSA decrypt password using RSA private key
func (s *UserAuthService) decryptPasswordWithRSA(keyID, encryptedPassword string) (string, error) {
	// Get private key from Redis
	privateKeyPEM, err := redis.GetEncryptionPrivateKey(keyID)
	if err != nil {
		return "", fmt.Errorf("get private key failed: %v", err)
	}

	// base64 decode
	password, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %v", err)
	}

	// Use pkg/utils to decrypt password
	decryptedBytes, err := utils.RSADecrypt(string(password), privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("RSA decrypt failed: %v", err)
	}

	return string(decryptedBytes), nil
}
