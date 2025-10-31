package utils

import (
	"encoding/json"
	"fmt"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"strings"
	"unicode"
)

// McpServerConfig MCP service configuration structure
type McpServerConfig struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}

// McpServersConfig MCP server configuration root structure
type McpServersConfig struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

// McpValidationResult MCP configuration validation result
type McpValidationResult struct {
	IsValid      bool   `json:"isValid"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	ServiceName  string `json:"serviceName,omitempty"`
	ProtocolType string `json:"protocolType,omitempty"`
	HasArgs      bool   `json:"hasArgs"`
	HasCommand   bool   `json:"hasCommand"`
	HasType      bool   `json:"hasType"`
	HasTransport bool   `json:"hasTransport"`
	HasURL       bool   `json:"hasURL"`
	Url          string `json:"url,omitempty"`
}

// ValidateMcpConfigFromString validate MCP configuration format from string
func ValidateMcpConfigFromString(configStr string) (*McpValidationResult, error) {
	return ValidateMcpConfig([]byte(configStr))
}

// ValidateMcpConfigFromMap validate MCP configuration format from map
func ValidateMcpConfigFromMap(configMap map[string]interface{}) (*McpValidationResult, error) {
	configData, err := json.Marshal(configMap)
	if err != nil {
		return &McpValidationResult{
			ErrorMessage: fmt.Sprintf("failed to serialize configuration: %v", err),
		}, nil
	}
	return ValidateMcpConfig(configData)
}

// ValidateMcpConfig validates MCP configuration format
func ValidateMcpConfig(configData []byte) (*McpValidationResult, error) {
	result := &McpValidationResult{}

	// Parse JSON data
	var config McpServersConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		result.ErrorMessage = fmt.Sprintf("JSON parsing failed: %v", err)
		return result, nil
	}

	// Check if mcpServers field exists
	if config.McpServers == nil {
		result.ErrorMessage = "missing mcpServers field"
		return result, nil
	}

	// Check if there is at least one service configuration
	if len(config.McpServers) == 0 || len(config.McpServers) > 1 {
		result.ErrorMessage = "mcpServers cannot be empty and must contain exactly one service configuration"
		return result, nil
	}

	// Get the first service name and configuration
	var serviceName string
	var serviceConfig McpServerConfig
	for name, cfg := range config.McpServers {
		// Validate name: must be letters
		if !isValidServiceName(name) {
			result.ErrorMessage = fmt.Sprintf("invalid service name: %s, service name must consist of letters", name)
			return result, nil
		}
		serviceName = name
		serviceConfig = cfg
		break // Only process the first service
	}

	result.ServiceName = serviceName

	// Check if fields exist
	result.HasArgs = len(serviceConfig.Args) > 0
	result.HasCommand = serviceConfig.Command != ""
	result.HasType = serviceConfig.Type != ""
	result.HasTransport = serviceConfig.Transport != ""
	result.HasURL = serviceConfig.URL != ""
	if result.HasURL {
		result.Url = serviceConfig.URL
	}

	// Determine protocol type logic
	protocolType := determineProtocolType(serviceConfig)
	result.ProtocolType = protocolType

	// Validate if protocol type is valid
	validProtocol := false
	for _, validT := range []string{model.McpProtocolStdio.String(), model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String()} {
		if protocolType == validT {
			validProtocol = true
			break
		}
	}
	if !validProtocol {
		result.ErrorMessage = fmt.Sprintf("invalid protocol type: %s, valid values are: %v", protocolType, []string{model.McpProtocolStdio.String(), model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String()})
		return result, nil
	}

	// Validate required fields based on protocol type
	if err := validateProtocolFields(protocolType, serviceConfig); err != nil {
		result.ErrorMessage = err.Error()
		return result, nil
	}

	// Validation successful
	result.IsValid = true
	return result, nil
}

// determineProtocolType determines the protocol type
func determineProtocolType(config McpServerConfig) string {
	// 1. Prioritize checking type and transport fields
	if config.Type != "" {
		return config.Type
	}
	if config.Transport != "" {
		return config.Transport
	}

	// 2. Check url field
	if config.URL != "" {
		if strings.Contains(strings.ToLower(config.URL), "sse") {
			return model.McpProtocolSSE.String()
		}
		return model.McpProtocolStreamableHttp.String()
	}

	// 3. Check command field
	if config.Command != "" {
		return model.McpProtocolStdio.String()
	}

	// Return empty string by default if unable to determine
	return ""
}

// validateProtocolFields validates protocol fields
func validateProtocolFields(protocolType string, config McpServerConfig) error {
	switch protocolType {
	case model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String():
		if config.URL == "" {
			return fmt.Errorf("%s protocol must contain a valid url field", protocolType)
		}
	case model.McpProtocolStdio.String():
		if config.Command == "" {
			return fmt.Errorf("%s protocol must contain a valid command field", protocolType)
		}
	default:
		return fmt.Errorf("unknown protocol type: %s", protocolType)
	}
	return nil
}

// isValidServiceName validates service name: letters, digits, underscore, hyphen, cannot start with digit
func isValidServiceName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with a letter
	if !unicode.IsLetter(rune(name[0])) {
		return false
	}
	// Can only contain letters, digits, underscores, and hyphens
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// CompareMcpValidationResult compares two McpValidationResult objects
func CompareMcpValidationResult(a, b *McpValidationResult) bool {
	if a.IsValid != b.IsValid {
		return false
	}
	if a.ErrorMessage != b.ErrorMessage {
		return false
	}
	if a.ServiceName != b.ServiceName {
		return false
	}
	if a.ProtocolType != b.ProtocolType {
		return false
	}
	if a.HasArgs != b.HasArgs {
		return false
	}
	if a.HasCommand != b.HasCommand {
		return false
	}
	if a.HasType != b.HasType {
		return false
	}
	if a.HasTransport != b.HasTransport {
		return false
	}
	if a.HasURL != b.HasURL {
		return false
	}
	if a.Url != b.Url {
		return false
	}
	return true
}
