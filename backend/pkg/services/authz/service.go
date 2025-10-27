package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"qm-mcp-server/api/authz/user_auth"
)

type AuthzService struct {
	Url           string
	Authorization string
	ContentType   string
}

func NewAuthzService(bearerToken string) *AuthzService {
	return &AuthzService{
		Url:           fmt.Sprintf("http://%s:%d", AuthzConfig.Host, AuthzConfig.Port),
		Authorization: bearerToken,
		ContentType:   ContentType,
	}
}

// GetUserInfo 获取用户信息
func (s *AuthzService) GetUserInfo(ctx context.Context) (*user_auth.GetUserInfoResponse, error) {
	req := &user_auth.GetUserInfoRequest{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}
	reqCtx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	url := fmt.Sprintf("%s%s", s.Url, ApiAuthzGetUserInfo)
	// GET请求不应该带body，应该使用query parameters
	values := url + "?" + string(reqBody)
	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, values, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", s.ContentType)
	httpReq.Header.Set("Authorization", s.Authorization)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// 解析响应体
	var respBody struct {
		Code    int                           `json:"code"`
		Message string                        `json:"message"`
		Data    user_auth.GetUserInfoResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &respBody); err != nil {
		return nil, fmt.Errorf("unmarshal response body failed: %w", err)
	}
	if respBody.Code != 0 {
		return nil, fmt.Errorf("unexpected code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return &respBody.Data, nil
}
