package services

import (
	"fmt"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/services/authz"
	"qm-mcp-server/pkg/services/market"
)

func LoadServices(cfgs *common.Services) error {
	if cfgs == nil {
		return nil
	}
	err := authz.LoadConfig(cfgs.McpAuthz)
	if err != nil {
		return fmt.Errorf("load authz config failed: %w", err)
	}
	err = market.LoadConfig(cfgs.McpMarket)
	if err != nil {
		return fmt.Errorf("load market config failed: %w", err)
	}
	return nil
}
