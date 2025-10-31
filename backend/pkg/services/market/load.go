package market

import (
	"fmt"
	"github.com/kymo-mcp/mcpcan/pkg/common"
)

var MarketConfig *common.Service

func LoadConfig(service *common.Service) error {
	if service == nil {
		return fmt.Errorf("market service is nil")
	}
	if service.Host == "" {
		return fmt.Errorf("market service host is empty")
	}
	if service.Port == 0 {
		return fmt.Errorf("market service port is 0")
	}
	MarketConfig = service
	return nil
}
