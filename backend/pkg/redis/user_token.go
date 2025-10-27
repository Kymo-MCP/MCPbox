package redis

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"
)

// UserToken 用户令牌结构
type UserToken struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"userId"`
	Token            string     `json:"token"`
	RefreshToken     string     `json:"refreshToken"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	RefreshExpiresAt *time.Time `json:"refreshExpiresAt"`
	LoginIP          *string    `json:"loginIp"`
	UserAgent        *string    `json:"userAgent"`
	CreateTime       *time.Time `json:"createTime"`
	UpdateTime       *time.Time `json:"updateTime"`
	SessionID        string     `json:"sessionId"` // 会话ID，基于 userId+IP+userAgent 生成
}

// UserSession 用户会话结构
type UserSession struct {
	SessionID        string     `json:"sessionId"`
	UserID           uint       `json:"userId"`
	LoginIP          string     `json:"loginIp"`
	UserAgent        string     `json:"userAgent"`
	Token            string     `json:"token"`
	RefreshToken     string     `json:"refreshToken"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	RefreshExpiresAt *time.Time `json:"refreshExpiresAt"`
	CreateTime       *time.Time `json:"createTime"`
	UpdateTime       *time.Time `json:"updateTime"`
}

const (
	// UserTokenPrefix 用户令牌Redis键前缀
	UserTokenPrefix = "user_token:"
	// RefreshTokenPrefix 刷新令牌Redis键前缀
	RefreshTokenPrefix = "refresh_token:"
	// UserTokenByUserIDPrefix 按用户ID索引的令牌键前缀
	UserTokenByUserIDPrefix = "user_tokens_by_user:"
	// UserSessionPrefix 用户会话Redis键前缀
	UserSessionPrefix = "user_session:"
	// UserSessionsByUserIDPrefix 按用户ID索引的会话键前缀
	UserSessionsByUserIDPrefix = "user_sessions_by_user:"
)

// GenerateSessionID 生成会话ID
func GenerateSessionID(userID uint, loginIP, userAgent string) string {
	data := fmt.Sprintf("%d_%s_%s", userID, loginIP, userAgent)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// SaveUserSession 保存用户会话到Redis
func SaveUserSession(session *UserSession) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()

	// 序列化会话数据
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	// 计算过期时间
	var expiration time.Duration
	if session.ExpiresAt != nil {
		expiration = time.Until(*session.ExpiresAt)
		if expiration <= 0 {
			return fmt.Errorf("session already expired")
		}
	} else {
		return fmt.Errorf("session expiresAt is nil")
	}

	// 保存会话
	sessionKey := UserSessionPrefix + session.SessionID
	err = client.client.Set(ctx, sessionKey, sessionData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	// 保存访问令牌映射
	tokenKey := UserTokenPrefix + session.Token
	err = client.client.Set(ctx, tokenKey, session.SessionID, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save token mapping: %v", err)
	}

	// 保存刷新令牌映射
	if session.RefreshToken != "" {
		refreshKey := RefreshTokenPrefix + session.RefreshToken
		var refreshExpiration time.Duration
		if session.RefreshExpiresAt != nil {
			refreshExpiration = time.Until(*session.RefreshExpiresAt)
		} else {
			return fmt.Errorf("session refreshExpiresAt is nil")
		}
		err = client.client.Set(ctx, refreshKey, session.SessionID, refreshExpiration).Err()
		if err != nil {
			return fmt.Errorf("failed to save refresh token mapping: %v", err)
		}
	}

	// 保存用户ID到会话的映射
	userSessionsKey := fmt.Sprintf("%s%d", UserSessionsByUserIDPrefix, session.UserID)
	err = client.client.SAdd(ctx, userSessionsKey, session.SessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to save user session mapping: %v", err)
	}

	// 设置用户会话集合的过期时间
	client.client.Expire(ctx, userSessionsKey, expiration)

	return nil
}

// GetUserSessionByToken 根据访问令牌获取用户会话信息
func GetUserSessionByToken(token string) (*UserSession, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	tokenKey := UserTokenPrefix + token

	// 获取会话ID
	sessionID, err := client.client.Get(ctx, tokenKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // 令牌不存在
		}
		return nil, fmt.Errorf("failed to get token mapping: %v", err)
	}

	// 根据会话ID获取会话信息
	return GetUserSessionByID(sessionID)
}

// GetUserSessionByRefreshToken 根据刷新令牌获取用户会话信息
func GetUserSessionByRefreshToken(refreshToken string) (*UserSession, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	refreshKey := RefreshTokenPrefix + refreshToken

	// 获取会话ID
	sessionID, err := client.client.Get(ctx, refreshKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // 刷新令牌不存在
		}
		return nil, fmt.Errorf("failed to get refresh token mapping: %v", err)
	}

	// 根据会话ID获取会话信息
	return GetUserSessionByID(sessionID)
}

// GetUserSessionByID 根据会话ID获取用户会话信息
func GetUserSessionByID(sessionID string) (*UserSession, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	sessionKey := UserSessionPrefix + sessionID

	sessionData, err := client.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // 会话不存在
		}
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	var userSession UserSession
	err = json.Unmarshal([]byte(sessionData), &userSession)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return &userSession, nil
}

// GetUserSessionsByUserID 获取用户的所有会话
func GetUserSessionsByUserID(userID uint) ([]*UserSession, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	userSessionsKey := fmt.Sprintf("%s%d", UserSessionsByUserIDPrefix, userID)

	// 获取用户的所有会话ID
	sessionIDs, err := client.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %v", err)
	}

	var sessions []*UserSession
	for _, sessionID := range sessionIDs {
		session, err := GetUserSessionByID(sessionID)
		if err != nil {
			// 记录错误但继续处理其他会话
			fmt.Printf("failed to get session %s: %v\n", sessionID, err)
			continue
		}
		if session != nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// DeleteUserSession 删除用户会话
func DeleteUserSession(sessionID string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()

	// 先获取会话信息以便清理相关数据
	userSession, err := GetUserSessionByID(sessionID)
	if err != nil {
		return err
	}
	if userSession == nil {
		return nil // 会话不存在
	}

	// 删除会话
	sessionKey := UserSessionPrefix + sessionID
	err = client.client.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	// 删除访问令牌映射
	if userSession.Token != "" {
		tokenKey := UserTokenPrefix + userSession.Token
		client.client.Del(ctx, tokenKey)
	}

	// 删除刷新令牌映射
	if userSession.RefreshToken != "" {
		refreshKey := RefreshTokenPrefix + userSession.RefreshToken
		client.client.Del(ctx, refreshKey)
	}

	// 从用户会话集合中移除
	userSessionsKey := fmt.Sprintf("%s%d", UserSessionsByUserIDPrefix, userSession.UserID)
	client.client.SRem(ctx, userSessionsKey, sessionID)

	return nil
}

// DeleteUserSessionByToken 根据访问令牌删除用户会话
func DeleteUserSessionByToken(token string) error {
	session, err := GetUserSessionByToken(token)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // 会话不存在
	}
	return DeleteUserSession(session.SessionID)
}

// DeleteUserSessionsByUserID 删除用户的所有会话
func DeleteUserSessionsByUserID(userID uint) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	userSessionsKey := fmt.Sprintf("%s%d", UserSessionsByUserIDPrefix, userID)

	// 获取用户的所有会话ID
	sessionIDs, err := client.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %v", err)
	}

	// 删除每个会话
	for _, sessionID := range sessionIDs {
		err = DeleteUserSession(sessionID)
		if err != nil {
			// 记录错误但继续删除其他会话
			fmt.Printf("failed to delete session %s: %v\n", sessionID, err)
		}
	}

	// 删除用户会话集合
	err = client.client.Del(ctx, userSessionsKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user sessions set: %v", err)
	}

	return nil
}

// UpdateUserSession 更新用户会话
func UpdateUserSession(session *UserSession) error {
	// 先删除旧的会话数据
	err := DeleteUserSession(session.SessionID)
	if err != nil {
		return err
	}

	// 保存新的会话数据
	return SaveUserSession(session)
}

// IsTokenValid 检查令牌是否有效
func IsTokenValid(token string) bool {
	userSession, err := GetUserSessionByToken(token)
	if err != nil || userSession == nil {
		return false
	}

	// 检查是否过期
	if userSession.ExpiresAt != nil && time.Now().After(*userSession.ExpiresAt) {
		// 令牌已过期，删除会话
		DeleteUserSession(userSession.SessionID)
		return false
	}

	return true
}

// IsRefreshTokenValid 检查刷新令牌是否有效
func IsRefreshTokenValid(refreshToken string) bool {
	userSession, err := GetUserSessionByRefreshToken(refreshToken)
	if err != nil || userSession == nil {
		return false
	}

	// 检查刷新令牌是否过期
	if userSession.RefreshExpiresAt != nil && time.Now().After(*userSession.RefreshExpiresAt) {
		// 刷新令牌已过期，删除会话
		DeleteUserSession(userSession.SessionID)
		return false
	}

	return true
}

// 兼容性函数 - 保持向后兼容
// SaveUserToken 保存用户令牌到Redis (兼容性函数)
func SaveUserToken(token *UserToken) error {
	// 转换为会话格式
	session := &UserSession{
		SessionID:        token.SessionID,
		UserID:           token.UserID,
		Token:            token.Token,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        token.ExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
		CreateTime:       token.CreateTime,
		UpdateTime:       token.UpdateTime,
	}

	if token.LoginIP != nil {
		session.LoginIP = *token.LoginIP
	}
	if token.UserAgent != nil {
		session.UserAgent = *token.UserAgent
	}

	// 如果没有会话ID，生成一个
	if session.SessionID == "" {
		session.SessionID = GenerateSessionID(token.UserID, session.LoginIP, session.UserAgent)
	}

	return SaveUserSession(session)
}

// GetUserTokenByToken 根据访问令牌获取用户令牌信息 (兼容性函数)
func GetUserTokenByToken(token string) (*UserToken, error) {
	session, err := GetUserSessionByToken(token)
	if err != nil || session == nil {
		return nil, err
	}

	// 转换为令牌格式
	userToken := &UserToken{
		UserID:           session.UserID,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		CreateTime:       session.CreateTime,
		UpdateTime:       session.UpdateTime,
		SessionID:        session.SessionID,
	}

	if session.LoginIP != "" {
		userToken.LoginIP = &session.LoginIP
	}
	if session.UserAgent != "" {
		userToken.UserAgent = &session.UserAgent
	}

	return userToken, nil
}

// GetUserTokenByRefreshToken 根据刷新令牌获取用户令牌信息 (兼容性函数)
func GetUserTokenByRefreshToken(refreshToken string) (*UserToken, error) {
	session, err := GetUserSessionByRefreshToken(refreshToken)
	if err != nil || session == nil {
		return nil, err
	}

	// 转换为令牌格式
	userToken := &UserToken{
		UserID:           session.UserID,
		Token:            session.Token,
		RefreshToken:     session.RefreshToken,
		ExpiresAt:        session.ExpiresAt,
		RefreshExpiresAt: session.RefreshExpiresAt,
		CreateTime:       session.CreateTime,
		UpdateTime:       session.UpdateTime,
		SessionID:        session.SessionID,
	}

	if session.LoginIP != "" {
		userToken.LoginIP = &session.LoginIP
	}
	if session.UserAgent != "" {
		userToken.UserAgent = &session.UserAgent
	}

	return userToken, nil
}

// DeleteUserToken 删除用户令牌 (兼容性函数)
func DeleteUserToken(token string) error {
	return DeleteUserSessionByToken(token)
}

// DeleteUserTokensByUserID 删除用户的所有令牌 (兼容性函数)
func DeleteUserTokensByUserID(userID uint) error {
	return DeleteUserSessionsByUserID(userID)
}

// UpdateUserToken 更新用户令牌 (兼容性函数)
func UpdateUserToken(token *UserToken) error {
	// 转换为会话格式
	session := &UserSession{
		SessionID:        token.SessionID,
		UserID:           token.UserID,
		Token:            token.Token,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        token.ExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
		CreateTime:       token.CreateTime,
		UpdateTime:       token.UpdateTime,
	}

	if token.LoginIP != nil {
		session.LoginIP = *token.LoginIP
	}
	if token.UserAgent != nil {
		session.UserAgent = *token.UserAgent
	}

	// 如果没有会话ID，生成一个
	if session.SessionID == "" {
		session.SessionID = GenerateSessionID(token.UserID, session.LoginIP, session.UserAgent)
	}

	return UpdateUserSession(session)
}
