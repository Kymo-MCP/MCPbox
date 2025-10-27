package utils

import (
	"encoding/json"
	"fmt"
	"qm-mcp-server/pkg/database/model"
	"strings"
	"unicode"
)

// McpServerConfig MCP服务配置结构
type McpServerConfig struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}

// McpServersConfig MCP服务器配置根结构
type McpServersConfig struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

// McpValidationResult MCP配置验证结果
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

// ValidateMcpConfigFromString 从字符串验证MCP配置格式
func ValidateMcpConfigFromString(configStr string) (*McpValidationResult, error) {
	return ValidateMcpConfig([]byte(configStr))
}

// ValidateMcpConfigFromMap 从map验证MCP配置格式
func ValidateMcpConfigFromMap(configMap map[string]interface{}) (*McpValidationResult, error) {
	configData, err := json.Marshal(configMap)
	if err != nil {
		return &McpValidationResult{
			ErrorMessage: fmt.Sprintf("序列化配置失败: %v", err),
		}, nil
	}
	return ValidateMcpConfig(configData)
}

// ValidateMcpConfig 验证MCP配置格式
func ValidateMcpConfig(configData []byte) (*McpValidationResult, error) {
	result := &McpValidationResult{}

	// 解析JSON数据
	var config McpServersConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		result.ErrorMessage = fmt.Sprintf("JSON解析失败: %v", err)
		return result, nil
	}

	// 检查mcpServers字段是否存在
	if config.McpServers == nil {
		result.ErrorMessage = "缺少mcpServers字段"
		return result, nil
	}

	// 检查是否至少有一个服务配置
	if len(config.McpServers) == 0 || len(config.McpServers) > 1 {
		result.ErrorMessage = "mcpServers不能为空，必须包含且仅包含一个服务配置"
		return result, nil
	}

	// 获取第一个服务名称和配置
	var serviceName string
	var serviceConfig McpServerConfig
	for name, cfg := range config.McpServers {
		// name 校验，必须是大小写字母
		if !isValidServiceName(name) {
			result.ErrorMessage = fmt.Sprintf("无效的服务名称: %s，服务名称必须是大小写字母组成", name)
			return result, nil
		}
		serviceName = name
		serviceConfig = cfg
		break // 只处理第一个服务
	}

	result.ServiceName = serviceName

	// 检查各字段是否存在
	result.HasArgs = len(serviceConfig.Args) > 0
	result.HasCommand = serviceConfig.Command != ""
	result.HasType = serviceConfig.Type != ""
	result.HasTransport = serviceConfig.Transport != ""
	result.HasURL = serviceConfig.URL != ""
	if result.HasURL {
		result.Url = serviceConfig.URL
	}

	// 协议类型判断逻辑
	protocolType := determineProtocolType(serviceConfig)
	result.ProtocolType = protocolType

	// 验证协议类型是否有效
	validProtocol := false
	for _, validT := range []string{model.McpProtocolStdio.String(), model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String()} {
		if protocolType == validT {
			validProtocol = true
			break
		}
	}
	if !validProtocol {
		result.ErrorMessage = fmt.Sprintf("无效的协议类型: %s，有效值为: %v", protocolType, []string{model.McpProtocolStdio.String(), model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String()})
		return result, nil
	}

	// 根据协议类型验证必要字段
	if err := validateProtocolFields(protocolType, serviceConfig); err != nil {
		result.ErrorMessage = err.Error()
		return result, nil
	}

	// 验证成功
	result.IsValid = true
	return result, nil
}

// determineProtocolType 确定协议类型
func determineProtocolType(config McpServerConfig) string {
	// 1. 优先检查 type 和 transport 字段
	if config.Type != "" {
		return config.Type
	}
	if config.Transport != "" {
		return config.Transport
	}

	// 2. 检查 url 字段
	if config.URL != "" {
		if strings.Contains(strings.ToLower(config.URL), "sse") {
			return model.McpProtocolSSE.String()
		}
		return model.McpProtocolStreamableHttp.String()
	}

	// 3. 检查 command 字段
	if config.Command != "" {
		return model.McpProtocolStdio.String()
	}

	// 默认返回空字符串表示无法确定
	return ""
}

// validateProtocolFields 验证协议字段
func validateProtocolFields(protocolType string, config McpServerConfig) error {
	switch protocolType {
	case model.McpProtocolSSE.String(), model.McpProtocolStreamableHttp.String():
		if config.URL == "" {
			return fmt.Errorf("%s协议必须包含有效的url字段", protocolType)
		}
	case model.McpProtocolStdio.String():
		if config.Command == "" {
			return fmt.Errorf("%s协议必须包含有效的command字段", protocolType)
		}
	default:
		return fmt.Errorf("未知的协议类型: %s", protocolType)
	}
	return nil
}

// isValidServiceName validates service name: letters, digits, underscore, hyphen, cannot start with digit
func isValidServiceName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Check first character: must be letter or underscore or hyphen
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' && firstChar != '-' {
		return false
	}

	// Check remaining characters: letters, digits, underscore, hyphen
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

// CompareMcpValidationResult 对比两个McpValidationResult是否相等
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
