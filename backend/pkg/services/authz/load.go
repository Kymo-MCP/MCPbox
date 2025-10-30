package authz

import (
	"fmt"
	"qm-mcp-server/pkg/common"
)

var AuthzConfig *common.Service

func LoadConfig(service *common.Service) error {
	if service == nil {
		return fmt.Errorf("authz service is nil")
	}
	if service.Host == "" {
		return fmt.Errorf("authz service host is empty")
	}
	if service.Port == 0 {
		return fmt.Errorf("authz service port is 0")
	}
	AuthzConfig = service
	return nil
}
