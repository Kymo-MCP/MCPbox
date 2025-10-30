package qm

import (
	"fmt"
	"qm-mcp-server/pkg/common"
	"strconv"
	"time"
)

// Client market client
type Client struct {
	config *common.MarketConfig
}

// NewClient creates a new market client
func NewClient(config *common.MarketConfig) *Client {
	return &Client{
		config: config,
	}
}

// NewClientFromGlobalConfig creates a market client from global configuration
func NewClientFromGlobalConfig(config *common.MarketConfig) *Client {
	if config == nil {
		return nil
	}
	return NewClient(config)
}

// GetConfig retrieves the client configuration
func (c *Client) GetConfig() *common.MarketConfig {
	return c.config
}

// GenerateAuthHeaders generates authentication headers
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

// GetBaseURL retrieves the base URL
func (c *Client) GetBaseURL() string {
	return c.config.Host
}

// BuildURL builds the full URL
func (c *Client) BuildURL(path string) string {
	return fmt.Sprintf("%s%s", c.config.Host, path)
}
