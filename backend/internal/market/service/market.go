package service

import (
	"fmt"
	"github.com/kymo-mcp/mcpcan/api/market/market"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/qm"

	"github.com/gin-gonic/gin"
)

// MarketService provides market service functionality
type MarketService struct {
	client *qm.Client
}

// NewMarketService creates a new MarketService instance
func NewMarketService() *MarketService {
	client := qm.NewClientFromGlobalConfig(&config.GlobalConfig.Market)
	if client == nil {
		return nil
	}
	return &MarketService{
		client: client,
	}
}

// MarketAPIRequest represents a market API request
type MarketAPIRequest struct {
	Path   string                 `json:"path" binding:"required"`
	Method string                 `json:"method"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// MarketAPIResponse represents a market API response
type MarketAPIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
}

// ListMarketServices retrieves a list of market services
func (s *MarketService) ListMarketServices(c *gin.Context) {
	// 绑定请求参数
	var req market.ListRequest
	if err := common.BindAndValidateUniversal(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("参数验证失败: %v", err))
		return
	}

	// 调用市场API
	response, err := s.client.ListServices(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("调用市场API失败: %v", err))
		return
	}

	common.GinSuccess(c, response)
}

// GetMarketServiceDetail retrieves detailed information about a market service
func (s *MarketService) GetMarketServiceDetail(c *gin.Context) {
	// 绑定请求参数
	var req market.DetailRequest
	if err := common.BindAndValidateUniversal(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("参数验证失败: %v", err))
		return
	}

	// 调用市场API
	response, err := s.client.GetServiceDetail(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("调用市场API失败: %v", err))
		return
	}

	common.GinSuccess(c, response)
}

// GetMarketCategories retrieves market categories
func (s *MarketService) GetMarketCategories(c *gin.Context) {
	// 绑定请求参数
	var req market.CategoryRequest
	if err := common.BindAndValidateUniversal(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("参数验证失败: %v", err))
		return
	}

	// 调用市场API
	response, err := s.client.GetCategories(&req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("调用市场API失败: %v", err))
		return
	}

	common.GinSuccess(c, response)
}

// GetMarketConfig retrieves market configuration
func (s *MarketService) GetMarketConfig(c *gin.Context) {
	config := s.client.GetConfig()

	// 隐藏敏感信息
	safeConfig := map[string]interface{}{
		"host":          config.Host,
		"customer_uuid": config.CustomerUuid,
		"secret_key":    "***" + config.SecretKey[len(config.SecretKey)-4:], // 只显示后4位
	}

	common.GinSuccess(c, safeConfig)
}

// 删除不再需要的转换函数，因为已经移动到 api.go 中
