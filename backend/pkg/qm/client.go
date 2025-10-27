package qm

import (
	"fmt"
	"qm-mcp-server/pkg/common"
	"strconv"
	"time"
)

// Client 市场客户端
type Client struct {
	config *common.MarketConfig
}

// NewClient 创建新的市场客户端
func NewClient(config *common.MarketConfig) *Client {
	return &Client{
		config: config,
	}
}

// NewClientFromGlobalConfig 从全局配置创建市场客户端
func NewClientFromGlobalConfig(config *common.MarketConfig) *Client {
	if config == nil {
		return nil
	}
	return NewClient(config)
}

// GetConfig 获取配置
func (c *Client) GetConfig() *common.MarketConfig {
	return c.config
}

// GenerateAuthHeaders 生成认证头信息
func (c *Client) GenerateAuthHeaders() map[string]string {
	timestampStr := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := GenerateSignature(c.config.CustomerUuid, timestampStr, c.config.SecretKey)
	_ = signature
	return map[string]string{
		// "X-Customer-UUID": c.config.CustomerUuid,
		// "X-Timestamp":     timestampStr,Ω
		// "X-Signature":     signature,
		// "Authorization": "Bearer ---",
	}
}

// GetBaseURL 获取基础URL
func (c *Client) GetBaseURL() string {
	return c.config.Host
}

// BuildURL 构建完整URL
func (c *Client) BuildURL(path string) string {
	return fmt.Sprintf("%s%s", c.config.Host, path)
}
