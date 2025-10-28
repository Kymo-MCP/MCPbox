package biz

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"qm-mcp-server/internal/authz/config"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/jwt"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/redis"
)

// AuthUseCase authentication business logic
type AuthUseCase struct {
	userBiz    *UserBiz
	logger     *zap.Logger
	jwtManager jwt.Manager
}

// NewAuthUseCase creates authentication business logic instance
func NewAuthUseCase() *AuthUseCase {
	uc := &AuthUseCase{
		logger:  logger.L().Logger,
		userBiz: NewUserBiz(),
	}
	// Initialize JWT manager
	jwtConfig := &jwt.Config{
		Secret:  config.GetConfig().Secret,
		Expires: time.Duration(common.AccessTokenExpireTime) * time.Second,
	}
	uc.jwtManager = jwt.NewManager(jwtConfig)
	return uc
}

// LoginData login return data
type LoginData struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	UserInfo     *UserInfo `json:"userInfo"`
}

// UserInfo user information
type UserInfo struct {
	UserID    int64    `json:"userId"`
	Username  string   `json:"username"`
	Nickname  string   `json:"nickname"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Avatar    string   `json:"avatar"`
	DeptID    int64    `json:"deptId"`
	DeptName  string   `json:"deptName"`
	RoleIDs   []uint   `json:"roleIds"`
	RoleNames []string `json:"roleNames"`
}

// TokenData token refresh return data
type TokenData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// ValidateResult token validation result
type ValidateResult struct {
	Valid     bool       `json:"valid"`
	UserInfo  *UserInfo  `json:"userInfo"`
	LoginInfo *LoginInfo `json:"loginInfo"`
}

// LoginInfo login information
type LoginInfo struct {
	LoginTime time.Time `json:"loginTime"`
	LoginIP   string    `json:"loginIp"`
	UserAgent string    `json:"userAgent"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Login user login
func (uc *AuthUseCase) Login(
	ctx context.Context,
	username string,
	plainPassword string,
	timestamp int64,
	clientIP string,
	userAgent string,
) (*LoginData, error) {
	uc.logger.Info("Start user login verification", zap.String("username", username))

	// Find user
	user, err := mysql.SysUserRepo.FindByUsername(ctx, username)
	if err != nil {
		uc.logger.Error("Failed to find user", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// Check user status
	if !user.IsEnabled() {
		uc.logger.Warn("User is disabled", zap.String("username", username))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUserDisabledError))
	}

	// Verify password
	if user.Password == nil {
		uc.logger.Error("User password is empty", zap.String("username", username))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// Double password verification
	if err := uc.userBiz.VerifyPassword(plainPassword, *user.Salt, *user.Password); err != nil {
		uc.logger.Error("Password verification failed", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// Generate token and refreshToken
	userDisplayName := ""
	if user.Username != nil {
		userDisplayName = *user.Username
	}
	token, err := uc.jwtManager.GenerateToken(int64(user.UserID), userDisplayName)
	if err != nil {
		uc.logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken()
	if err != nil {
		uc.logger.Error("Failed to generate refreshToken", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	// Set expiration time
	now := time.Now()
	tokenExpiry := now.Add(common.AccessTokenExpireTime * time.Second)    // 24 hours
	refreshExpiry := now.Add(common.RefreshTokenExpireTime * time.Second) // 7 days

	// Create user session record
	userSession := &redis.UserSession{
		SessionID:        redis.GenerateSessionID(user.UserID, clientIP, userAgent),
		UserID:           user.UserID,
		LoginIP:          clientIP,
		UserAgent:        userAgent,
		Token:            token,
		RefreshToken:     refreshToken,
		ExpiresAt:        &tokenExpiry,
		RefreshExpiresAt: &refreshExpiry,
		CreateTime:       &now,
		UpdateTime:       &now,
	}

	// Save new session to Redis (support multi-browser sessions)
	if err := redis.SaveUserSession(userSession); err != nil {
		uc.logger.Error("Failed to save session", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	// Construct user information
	avatar := ""
	if user.AvatarPath != nil {
		avatar = *user.AvatarPath
	}
	deptName := ""
	deptID := user.GetDeptID()
	if deptID > 0 {
		if dept, derr := mysql.SysDeptRepo.FindByID(ctx, deptID); derr == nil && dept != nil {
			deptName = dept.Name
		} else if derr != nil {
			uc.logger.Warn("Failed to query department name", zap.Uint("deptId", deptID), zap.Error(derr))
		}
	}
	roleIDs := []uint{}
	roleNames := []string{}
	if mysql.SysUsersRolesRepo != nil {
		if ids, rerr := mysql.SysUsersRolesRepo.FindRoleIDsByUserID(ctx, user.UserID); rerr == nil {
			roleIDs = ids
			for _, rid := range ids {
				if role, ferr := mysql.SysRoleRepo.FindByID(ctx, rid); ferr == nil && role != nil {
					roleNames = append(roleNames, role.Name)
				} else if ferr != nil {
					uc.logger.Warn("Failed to query role", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("Failed to query user role IDs", zap.Uint("userId", user.UserID), zap.Error(rerr))
		}
	}

	userInfo := &UserInfo{
		UserID:    int64(user.UserID),
		Username:  user.GetUsername(),
		Nickname:  user.GetNickName(),
		Email:     user.GetEmail(),
		Phone:     user.GetPhone(),
		Avatar:    avatar,
		DeptID:    int64(deptID),
		DeptName:  deptName,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
	}

	loginData := &LoginData{
		Token:        token,
		RefreshToken: refreshToken,
		UserInfo:     userInfo,
	}

	uc.logger.Info("User login successful", zap.String("username", username), zap.Uint("userId", user.UserID))
	return loginData, nil
}

// Logout user logout
func (uc *AuthUseCase) Logout(ctx context.Context, userID int64, token string) error {
	uc.logger.Info("User logout", zap.Int64("userId", userID), zap.String("token", token[:10]+"..."))

	// Delete session record
	if err := redis.DeleteUserSessionByToken(token); err != nil {
		uc.logger.Error("Failed to delete session", zap.Error(err))
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLogoutFailure))
	}

	uc.logger.Info("User logout successful", zap.Int64("userId", userID))
	return nil
}

// RefreshToken refresh token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*TokenData, error) {
	uc.logger.Info("Refresh token request")

	// Find refreshToken record
	sessionRecord, err := redis.GetUserSessionByRefreshToken(refreshToken)
	if err != nil {
		uc.logger.Error("Failed to find refreshToken", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenInvalid))
	}
	if sessionRecord == nil {
		uc.logger.Warn("RefreshToken does not exist")
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenInvalid))
	}

	// Check if refreshToken is expired
	if sessionRecord.RefreshExpiresAt != nil && time.Now().After(*sessionRecord.RefreshExpiresAt) {
		uc.logger.Warn("RefreshToken has expired")
		// Clean up expired session
		redis.DeleteUserSession(sessionRecord.SessionID)
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenExpired))
	}

	// Generate new token and refreshToken
	// Get user information for generating JWT token
	user, err := mysql.SysUserRepo.FindByID(ctx, sessionRecord.UserID)
	if err != nil {
		uc.logger.Error("Failed to find user", zap.Uint("userId", sessionRecord.UserID), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshFailure))
	}

	userDisplayName := ""
	if user.Username != nil {
		userDisplayName = *user.Username
	}

	newToken, err := uc.jwtManager.GenerateToken(int64(user.UserID), userDisplayName)
	if err != nil {
		uc.logger.Error("Failed to generate new JWT token", zap.Error(err))
		return nil, fmt.Errorf("Failed to generate new token: %w", err)
	}

	newRefreshToken, err := uc.jwtManager.GenerateRefreshToken()
	if err != nil {
		uc.logger.Error("Failed to generate new refresh token", zap.Error(err))
		return nil, fmt.Errorf("Failed to generate new refresh token: %w", err)
	}

	// Delete old session
	if err := redis.DeleteUserSession(sessionRecord.SessionID); err != nil {
		uc.logger.Warn("Failed to delete old session", zap.Error(err))
	}

	// Create new session
	now := time.Now()
	tokenExpiry := now.Add(common.AccessTokenExpireTime * time.Second)    // 24 hours
	refreshExpiry := now.Add(common.RefreshTokenExpireTime * time.Second) // 7 days

	newSession := &redis.UserSession{
		SessionID:        redis.GenerateSessionID(sessionRecord.UserID, sessionRecord.LoginIP, sessionRecord.UserAgent),
		UserID:           sessionRecord.UserID,
		LoginIP:          sessionRecord.LoginIP,
		UserAgent:        sessionRecord.UserAgent,
		Token:            newToken,
		RefreshToken:     newRefreshToken,
		ExpiresAt:        &tokenExpiry,
		RefreshExpiresAt: &refreshExpiry,
		CreateTime:       &now,
		UpdateTime:       &now,
	}

	if err := redis.SaveUserSession(newSession); err != nil {
		uc.logger.Error("Failed to save new session", zap.Error(err))
		return nil, fmt.Errorf("Refresh failed, please try again")
	}

	tokenData := &TokenData{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	uc.logger.Info("Token refresh successful", zap.Uint("userId", sessionRecord.UserID))
	return tokenData, nil
}

// ValidateToken validate token
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*ValidateResult, error) {
	uc.logger.Debug("Validate JWT token request")

	// Validate JWT token
	claims, err := uc.jwtManager.ValidateToken(token)
	if err != nil {
		uc.logger.Error("JWT token validation failed", zap.Error(err))
		return &ValidateResult{Valid: false}, nil
	}

	// Check if token exists and is valid in Redis
	sessionRecord, err := redis.GetUserSessionByToken(token)
	if err != nil || sessionRecord == nil {
		uc.logger.Warn("Token is invalid or expired in Redis")
		return &ValidateResult{Valid: false}, nil
	}

	// Check if session is expired
	if sessionRecord.ExpiresAt != nil && time.Now().After(*sessionRecord.ExpiresAt) {
		uc.logger.Warn("Session has expired")
		// Clean up expired session
		redis.DeleteUserSession(sessionRecord.SessionID)
		return &ValidateResult{Valid: false}, nil
	}

	// Get user information
	user, err := mysql.SysUserRepo.FindByID(ctx, uint(claims.UserID))
	if err != nil {
		uc.logger.Error("Failed to find user", zap.Int64("userId", claims.UserID), zap.Error(err))
		return &ValidateResult{Valid: false}, nil
	}

	// Check user status
	if !user.IsEnabled() {
		uc.logger.Warn("User is disabled", zap.Uint("userId", user.UserID))
		return &ValidateResult{Valid: false}, nil
	}

	// Construct user information
	avatar := ""
	if user.AvatarPath != nil {
		avatar = *user.AvatarPath
	}
	deptName := ""
	deptID := user.GetDeptID()
	if deptID > 0 {
		if dept, derr := mysql.SysDeptRepo.FindByID(ctx, deptID); derr == nil && dept != nil {
			deptName = dept.Name
		} else if derr != nil {
			uc.logger.Warn("Failed to query department name", zap.Uint("deptId", deptID), zap.Error(derr))
		}
	}
	roleIDs := []uint{}
	roleNames := []string{}
	if mysql.SysUsersRolesRepo != nil {
		if ids, rerr := mysql.SysUsersRolesRepo.FindRoleIDsByUserID(ctx, user.UserID); rerr == nil {
			roleIDs = ids
			for _, rid := range ids {
				if role, ferr := mysql.SysRoleRepo.FindByID(ctx, rid); ferr == nil && role != nil {
					roleNames = append(roleNames, role.Name)
				} else if ferr != nil {
					uc.logger.Warn("Failed to query role", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("Failed to query user role IDs", zap.Uint("userId", user.UserID), zap.Error(rerr))
		}
	}

	userInfo := &UserInfo{
		UserID:    int64(user.UserID),
		Username:  user.GetUsername(),
		Nickname:  user.GetNickName(),
		Email:     user.GetEmail(),
		Phone:     user.GetPhone(),
		Avatar:    avatar,
		DeptID:    int64(deptID),
		DeptName:  deptName,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
	}

	// Construct login information
	loginInfo := &LoginInfo{
		LoginTime: claims.ExpiresAt.Time, // Get expiration time from JWT claims
		LoginIP:   "",                    // Default value
		UserAgent: "",                    // Default value
		ExpiresAt: claims.ExpiresAt.Time, // Get expiration time from JWT claims
	}

	// Use login information from session record
	if sessionRecord.CreateTime != nil {
		loginInfo.LoginTime = *sessionRecord.CreateTime
	}
	loginInfo.LoginIP = sessionRecord.LoginIP
	loginInfo.UserAgent = sessionRecord.UserAgent
	if sessionRecord.ExpiresAt != nil {
		loginInfo.ExpiresAt = *sessionRecord.ExpiresAt
	}

	result := &ValidateResult{
		Valid:     true,
		UserInfo:  userInfo,
		LoginInfo: loginInfo,
	}

	uc.logger.Debug("Token validation successful", zap.Uint("userId", user.UserID))
	return result, nil
}

// GetUserInfo gets user information
func (uc *AuthUseCase) GetUserInfo(ctx context.Context, userID uint) (*UserInfo, error) {
	uc.logger.Debug("Get user information request")

	// Get user information
	user, err := mysql.SysUserRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to find user", zap.Uint("userId", userID), zap.Error(err))
		return nil, fmt.Errorf("Failed to find user: %w", err)
	}

	// Check user status
	if !user.IsEnabled() {
		uc.logger.Warn("User is disabled", zap.Uint("userId", user.UserID))
		return nil, fmt.Errorf("User is disabled")
	}

	// Construct user information
	avatar := ""
	if user.AvatarPath != nil {
		avatar = *user.AvatarPath
	}
	deptName := ""
	deptID := user.GetDeptID()
	if deptID > 0 {
		if dept, derr := mysql.SysDeptRepo.FindByID(ctx, deptID); derr == nil && dept != nil {
			deptName = dept.Name
		} else if derr != nil {
			uc.logger.Warn("Failed to query department name", zap.Uint("deptId", deptID), zap.Error(derr))
		}
	}
	roleIDs := []uint{}
	roleNames := []string{}
	if mysql.SysUsersRolesRepo != nil {
		if ids, rerr := mysql.SysUsersRolesRepo.FindRoleIDsByUserID(ctx, user.UserID); rerr == nil {
			roleIDs = ids
			for _, rid := range ids {
				if role, ferr := mysql.SysRoleRepo.FindByID(ctx, rid); ferr == nil && role != nil {
					roleNames = append(roleNames, role.Name)
				} else if ferr != nil {
					uc.logger.Warn("Failed to query role", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("Failed to query user role IDs", zap.Uint("userId", user.UserID), zap.Error(rerr))
		}
	}

	userInfo := &UserInfo{
		UserID:    int64(user.UserID),
		Username:  user.GetUsername(),
		Nickname:  user.GetNickName(),
		Email:     user.GetEmail(),
		Phone:     user.GetPhone(),
		Avatar:    avatar,
		DeptID:    int64(deptID),
		DeptName:  deptName,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
	}

	return userInfo, nil
}
