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

// AuthUseCase 认证业务逻辑
type AuthUseCase struct {
	userBiz    *UserBiz
	logger     *zap.Logger
	jwtManager jwt.Manager
}

// NewAuthUseCase 创建认证业务逻辑实例
func NewAuthUseCase() *AuthUseCase {
	uc := &AuthUseCase{
		logger:  logger.L().Logger,
		userBiz: NewUserBiz(),
	}
	// 初始化JWT管理器
	jwtConfig := &jwt.Config{
		Secret:  config.GetConfig().Secret,
		Expires: time.Duration(common.AccessTokenExpireTime) * time.Second,
	}
	uc.jwtManager = jwt.NewManager(jwtConfig)
	return uc
}

// LoginData 登录返回数据
type LoginData struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	UserInfo     *UserInfo `json:"userInfo"`
}

// UserInfo 用户信息
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

// TokenData token刷新返回数据
type TokenData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// ValidateResult token验证结果
type ValidateResult struct {
	Valid     bool       `json:"valid"`
	UserInfo  *UserInfo  `json:"userInfo"`
	LoginInfo *LoginInfo `json:"loginInfo"`
}

// LoginInfo 登录信息
type LoginInfo struct {
	LoginTime time.Time `json:"loginTime"`
	LoginIP   string    `json:"loginIp"`
	UserAgent string    `json:"userAgent"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Login 用户登录
func (uc *AuthUseCase) Login(
	ctx context.Context,
	username string,
	plainPassword string,
	timestamp int64,
	clientIP string,
	userAgent string,
) (*LoginData, error) {
	uc.logger.Info("开始用户登录验证", zap.String("username", username))

	// 查找用户
	user, err := mysql.SysUserRepo.FindByUsername(ctx, username)
	if err != nil {
		uc.logger.Error("查找用户失败", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// 检查用户状态
	if !user.IsEnabled() {
		uc.logger.Warn("用户已禁用", zap.String("username", username))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUserDisabledError))
	}

	// 验证密码
	if user.Password == nil {
		uc.logger.Error("用户密码为空", zap.String("username", username))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// 双重密码验证
	if err := uc.userBiz.VerifyPassword(plainPassword, *user.Salt, *user.Password); err != nil {
		uc.logger.Error("密码验证失败", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUsernameOrPasswordIncorrect))
	}

	// 生成token和refreshToken
	userDisplayName := ""
	if user.Username != nil {
		userDisplayName = *user.Username
	}
	token, err := uc.jwtManager.GenerateToken(int64(user.UserID), userDisplayName)
	if err != nil {
		uc.logger.Error("生成JWT token失败", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken()
	if err != nil {
		uc.logger.Error("生成refreshToken失败", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	// 设置过期时间
	now := time.Now()
	tokenExpiry := now.Add(common.AccessTokenExpireTime * time.Second)    // 24小时
	refreshExpiry := now.Add(common.RefreshTokenExpireTime * time.Second) // 7天

	// 创建用户会话记录
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

	// 保存新会话到Redis（支持多浏览器会话）
	if err := redis.SaveUserSession(userSession); err != nil {
		uc.logger.Error("保存会话失败", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLoginFailure))
	}

	// 构造用户信息
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
			uc.logger.Warn("查询部门名称失败", zap.Uint("deptId", deptID), zap.Error(derr))
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
					uc.logger.Warn("查询角色失败", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("查询用户角色ID失败", zap.Uint("userId", user.UserID), zap.Error(rerr))
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

	uc.logger.Info("用户登录成功", zap.String("username", username), zap.Uint("userId", user.UserID))
	return loginData, nil
}

// Logout 用户退出
func (uc *AuthUseCase) Logout(ctx context.Context, userID int64, token string) error {
	uc.logger.Info("用户退出", zap.Int64("userId", userID), zap.String("token", token[:10]+"..."))

	// 删除会话记录
	if err := redis.DeleteUserSessionByToken(token); err != nil {
		uc.logger.Error("删除会话失败", zap.Error(err))
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeLogoutFailure))
	}

	uc.logger.Info("用户退出成功", zap.Int64("userId", userID))
	return nil
}

// RefreshToken 刷新token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*TokenData, error) {
	uc.logger.Info("刷新token请求")

	// 查找refreshToken记录
	sessionRecord, err := redis.GetUserSessionByRefreshToken(refreshToken)
	if err != nil {
		uc.logger.Error("查找refreshToken失败", zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenInvalid))
	}
	if sessionRecord == nil {
		uc.logger.Warn("refreshToken不存在")
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenInvalid))
	}

	// 检查refreshToken是否过期
	if sessionRecord.RefreshExpiresAt != nil && time.Now().After(*sessionRecord.RefreshExpiresAt) {
		uc.logger.Warn("refreshToken已过期")
		// 清理过期会话
		redis.DeleteUserSession(sessionRecord.SessionID)
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshTokenExpired))
	}

	// 生成新的token和refreshToken
	// 获取用户信息用于生成JWT token
	user, err := mysql.SysUserRepo.FindByID(ctx, sessionRecord.UserID)
	if err != nil {
		uc.logger.Error("查找用户失败", zap.Uint("userId", sessionRecord.UserID), zap.Error(err))
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeRefreshFailure))
	}

	userDisplayName := ""
	if user.Username != nil {
		userDisplayName = *user.Username
	}

	newToken, err := uc.jwtManager.GenerateToken(int64(user.UserID), userDisplayName)
	if err != nil {
		uc.logger.Error("生成新JWT token失败", zap.Error(err))
		return nil, fmt.Errorf("生成新token失败: %w", err)
	}

	newRefreshToken, err := uc.jwtManager.GenerateRefreshToken()
	if err != nil {
		uc.logger.Error("生成新refresh token失败", zap.Error(err))
		return nil, fmt.Errorf("生成新refresh token失败: %w", err)
	}

	// 删除旧会话
	if err := redis.DeleteUserSession(sessionRecord.SessionID); err != nil {
		uc.logger.Warn("删除旧会话失败", zap.Error(err))
	}

	// 创建新会话
	now := time.Now()
	tokenExpiry := now.Add(common.AccessTokenExpireTime * time.Second)    // 24小时
	refreshExpiry := now.Add(common.RefreshTokenExpireTime * time.Second) // 7天

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
		uc.logger.Error("保存新会话失败", zap.Error(err))
		return nil, fmt.Errorf("刷新失败，请重试")
	}

	tokenData := &TokenData{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	uc.logger.Info("token刷新成功", zap.Uint("userId", sessionRecord.UserID))
	return tokenData, nil
}

// ValidateToken 验证token
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*ValidateResult, error) {
	uc.logger.Debug("验证JWT token请求")

	// 验证JWT token
	claims, err := uc.jwtManager.ValidateToken(token)
	if err != nil {
		uc.logger.Error("JWT token验证失败", zap.Error(err))
		return &ValidateResult{Valid: false}, nil
	}

	// 检查token是否在Redis中存在且有效
	sessionRecord, err := redis.GetUserSessionByToken(token)
	if err != nil || sessionRecord == nil {
		uc.logger.Warn("token在Redis中无效或已过期")
		return &ValidateResult{Valid: false}, nil
	}

	// 检查会话是否过期
	if sessionRecord.ExpiresAt != nil && time.Now().After(*sessionRecord.ExpiresAt) {
		uc.logger.Warn("会话已过期")
		// 清理过期会话
		redis.DeleteUserSession(sessionRecord.SessionID)
		return &ValidateResult{Valid: false}, nil
	}

	// 获取用户信息
	user, err := mysql.SysUserRepo.FindByID(ctx, uint(claims.UserID))
	if err != nil {
		uc.logger.Error("查找用户失败", zap.Int64("userId", claims.UserID), zap.Error(err))
		return &ValidateResult{Valid: false}, nil
	}

	// 检查用户状态
	if !user.IsEnabled() {
		uc.logger.Warn("用户已禁用", zap.Uint("userId", user.UserID))
		return &ValidateResult{Valid: false}, nil
	}

	// 构造用户信息
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
			uc.logger.Warn("查询部门名称失败", zap.Uint("deptId", deptID), zap.Error(derr))
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
					uc.logger.Warn("查询角色失败", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("查询用户角色ID失败", zap.Uint("userId", user.UserID), zap.Error(rerr))
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

	// 构造登录信息
	loginInfo := &LoginInfo{
		LoginTime: claims.ExpiresAt.Time, // 从JWT claims获取过期时间
		LoginIP:   "",                    // 默认值
		UserAgent: "",                    // 默认值
		ExpiresAt: claims.ExpiresAt.Time, // 从JWT claims获取过期时间
	}

	// 使用会话记录中的登录信息
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

	uc.logger.Debug("token验证成功", zap.Uint("userId", user.UserID))
	return result, nil
}

// UserInfo
func (uc *AuthUseCase) GetUserInfo(ctx context.Context, userID uint) (*UserInfo, error) {
	uc.logger.Debug("获取用户信息请求")

	// 获取用户信息
	user, err := mysql.SysUserRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error("查找用户失败", zap.Uint("userId", userID), zap.Error(err))
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}

	// 检查用户状态
	if !user.IsEnabled() {
		uc.logger.Warn("用户已禁用", zap.Uint("userId", user.UserID))
		return nil, fmt.Errorf("用户已禁用")
	}

	// 构造用户信息
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
			uc.logger.Warn("查询部门名称失败", zap.Uint("deptId", deptID), zap.Error(derr))
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
					uc.logger.Warn("查询角色失败", zap.Uint("roleId", rid), zap.Error(ferr))
				}
			}
		} else {
			uc.logger.Warn("查询用户角色ID失败", zap.Uint("userId", user.UserID), zap.Error(rerr))
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
