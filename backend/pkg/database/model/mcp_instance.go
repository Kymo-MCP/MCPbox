package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type AccessType string

func (accessType AccessType) String() string {
	return string(accessType)
}

const (
	// 直连模式，代理模式，托管模式
	AccessTypeDirect  AccessType = "direct"
	AccessTypeProxy   AccessType = "proxy"
	AccessTypeHosting AccessType = "hosting"
)

type McpProtocol string

func (mcpProtocol McpProtocol) String() string {
	return string(mcpProtocol)
}

const (
	// SSE 协议
	McpProtocolSSE McpProtocol = "sse"
	// StreamableHttp 协议
	McpProtocolStreamableHttp McpProtocol = "streamable-http"
	// Stdio 协议
	McpProtocolStdio McpProtocol = "stdio"
)

type InstanceStatus string

const (
	// 激活
	InstanceStatusActive InstanceStatus = "active"
	// 停用
	InstanceStatusInactive InstanceStatus = "inactive"
)

// 容器状态
type ContainerStatus string

const (
	// 启动中
	ContainerStatusPending ContainerStatus = "pending"
	// 运行中
	ContainerStatusRunning ContainerStatus = "running"
	// 运行中但未就绪
	ContainerStatusRunningUnready ContainerStatus = "running-unready"
	// 启动超时停止
	ContainerStatusInitTimeoutStop ContainerStatus = "init-timeout-stop"
	// 运行超时停止
	ContainerStatusRunTimeoutStop ContainerStatus = "run-timeout-stop"
	// 异常强制停止
	ContainerStatusExceptionForceStop ContainerStatus = "exception-force-stop"
	// 手动停止
	ContainerStatusManualStop ContainerStatus = "manual-stop"
	// 创建失败
	ContainerStatusCreateFailed ContainerStatus = "create-failed"
)

const DefaultMcpType = "sse"

type SourceType string

const (
	// MCP 市场
	SourceTypeMarket SourceType = "market"
	// 实例模版
	SourceTypeTemplate SourceType = "template"
	// 自定义
	SourceTypeCustom SourceType = "custom"
)

// McpInstance MCP 实例数据库模型
type McpInstance struct {
	ID                     uint            `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	InstanceID             string          `gorm:"size:100;not null;comment:实例ID" json:"instanceID"`
	InstanceName           string          `gorm:"size:200;not null;comment:实例名称" json:"instanceName"`
	Notes                  string          `gorm:"type:text;comment:备注" json:"notes"`
	AccessType             AccessType      `gorm:"size:20;not null;comment:访问类型 (直连-direct/代理-proxy/托管-hosting)" json:"accessType"`
	McpProtocol            McpProtocol     `gorm:"size:20;not null;comment:MCP 协议 (sse/streamableHttp/stdio)" json:"mcpProtocol"`
	Status                 InstanceStatus  `gorm:"size:20;not null;default:active;comment:实例状态 (活跃-active/不活跃-inactive)" json:"status"`
	PackageID              string          `gorm:"size:100;not null;comment:实例所属套餐ID" json:"packageID"`
	EnvironmentID          uint            `gorm:"default:0;comment:环境ID" json:"environmentID"`
	SourceType             SourceType      `gorm:"size:20;not null;comment:实例来源 (MCP 市场-market/实例模版-template/自定义-custom)" json:"sourceType"`
	McpServerID            string          `gorm:"size:100;not null;comment:MCP 服务器ID" json:"mcpServerID"`
	TemplateID             uint            `gorm:"size:100;not null;comment:实例模版ID" json:"templateID"`
	Tokens                 []McpToken      `gorm:"type:json;comment:MCP 实例令牌 (JSON格式)" json:"tokens"`
	ImgAddr                string          `gorm:"size:100;not null;default:'';comment:镜像地址" json:"imgAddr"`
	Port                   int32           `gorm:"default:0;comment:端口号" json:"port"`
	InitScript             string          `gorm:"type:text;comment:初始化脚本" json:"initScript"`
	Command                string          `gorm:"type:text;comment:启动命令" json:"command"`
	EnvironmentVariables   json.RawMessage `gorm:"type:json;comment:环境变量 (JSON格式)" json:"environmentVariables"`
	VolumeMounts           json.RawMessage `gorm:"type:json;comment:卷挂载配置列表 (JSON格式)" json:"volumeMounts"`
	StartupTimeout         int64           `gorm:"type:bigint;default:0;comment:容器启动超时时间 (毫秒时间戳)" json:"startupTimeout"`
	RunningTimeout         int64           `gorm:"type:bigint;default:0;comment:容器运行超时时间 (毫秒时间戳)" json:"runningTimeout"`
	ContainerCreateOptions json.RawMessage `gorm:"type:json;comment:容器创建选项 (JSON格式)" json:"containerCreateOptions"`
	ContainerStatus        ContainerStatus `gorm:"size:20;not null;default:pending;comment:容器状态 (启动中-pending/运行中-running/启动超时停止-init-timeout-stop/运行超时停止-run-timeout-stop/异常强制停止-exception-force-stop/手动停止-manual-stop)" json:"containerStatus"`
	ContainerName          string          `gorm:"size:100;not null;comment:容器名称" json:"containerName"`
	ContainerServiceName   string          `gorm:"size:100;not null;comment:容器服务名称" json:"containerServiceName"`
	ContainerIsReady       bool            `gorm:"not null;comment:容器服务名称" json:"containerIsReady"`
	ContainerLastMessage   string          `gorm:"type:text;comment:容器上次状态信息" json:"containerLastMessage"`
	SourceConfig           json.RawMessage `gorm:"type:json;comment:MCP 来源服务配置 (JSON格式)" json:"sourceConfig"`
	TargetConfig           json.RawMessage `gorm:"type:json;comment:MCP 目标服务配置 (JSON格式)" json:"targetConfig"`
	PublicProxyConfig      json.RawMessage `gorm:"type:json;comment:MCP 公网代理服务配置 (JSON格式)" json:"publicProxyConfig"`
	ServicePath            string          `gorm:"size:100;not null;default:'';comment:MCP 服务路径" json:"servicePath"`
	IconPath               string          `gorm:"size:100;not null;default:'';comment:MCP 图标路径" json:"iconPath"`
	CreatedAt              time.Time       `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt              time.Time       `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

type McpToken struct {
	Token     string   `json:"token"`
	ExpireAt  int64    `json:"expireAt"`
	PublishAt int64    `json:"publishAt"`
	Usages    []string `json:"usages"`
}

// McpConfig 表示单个 MCP 服务器配置
type McpConfig struct {
	URL            string            `json:"url"`
	Type           string            `json:"type,omitempty"`
	Command        string            `json:"command,omitempty"`
	Transport      string            `json:"transport,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Timeout        int               `json:"timeout,omitempty"`
	SseReadTimeout int               `json:"sseReadTimeout,omitempty"`
}

// McpServersConfig 统一的 MCP 服务器配置结构
type McpServersConfig struct {
	McpServers map[string]*McpConfig `json:"mcpServers"`
}

// GetMcpConfig 获取第一个 MCP 配置
func (m *McpServersConfig) GetMcpConfig() (*McpConfig, error) {
	if m == nil || len(m.McpServers) == 0 {
		return nil, fmt.Errorf("no mcp servers found in config")
	}
	for _, cfg := range m.McpServers {
		return cfg, nil
	}
	return nil, fmt.Errorf("no valid mcp server config found")
}

// GetMcpConfigByName 根据名称获取 MCP 配置
func (m *McpServersConfig) GetMcpConfigByName(name string) (*McpConfig, error) {
	if m == nil || len(m.McpServers) == 0 {
		return nil, fmt.Errorf("no mcp servers found in config")
	}
	if cfg, exists := m.McpServers[name]; exists {
		return cfg, nil
	}
	return nil, fmt.Errorf("mcp server config not found for name: %s", name)
}

// 为了向后兼容，保留原有类型别名
type SourceConfig = McpServersConfig
type TargetConfig = McpServersConfig
type PublicProxyConfig = McpServersConfig
type InnerProxyConfig = McpServersConfig

// TableName 指定表名
func (McpInstance) TableName() string {
	return "mcp_instance"
}

// ParseMcpServersConfig 通用解析 MCP 服务器配置
func ParseMcpServersConfig(rawConfig json.RawMessage) (string, *McpConfig, error) {
	var tempConfig struct {
		McpServers map[string]json.RawMessage `json:"mcpServers"`
	}

	if err := json.Unmarshal(rawConfig, &tempConfig); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if len(tempConfig.McpServers) == 0 {
		return "", nil, fmt.Errorf("no mcpServers found in config")
	}

	// 遍历服务器配置，只取第一个有效的配置
	for mcpName, serverRaw := range tempConfig.McpServers {
		mcpCfg := &McpConfig{}
		if err := json.Unmarshal(serverRaw, mcpCfg); err != nil {
			// 忽略无法解析的配置，继续下一个
			continue
		}

		if len(mcpCfg.URL) == 0 {
			continue // URL 是必要字段
		}

		if len(mcpCfg.Type) == 0 && len(mcpCfg.Transport) == 0 {
			mcpCfg.Type = DefaultMcpType
		}
		return mcpName, mcpCfg, nil
	}

	return "", nil, fmt.Errorf("no valid server config found")
}

// parseMcpServersConfig 解析 MCP 服务器配置
func parseMcpServersConfig(data json.RawMessage) (string, *McpServersConfig, *McpConfig, error) {
	var cfg McpServersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", nil, nil, fmt.Errorf("failed to unmarshal mcp servers config: %w", err)
	}

	// 获取第一个有效的 MCP 配置
	for mcpName, mcpConfig := range cfg.McpServers {
		if mcpConfig != nil {
			return mcpName, &cfg, mcpConfig, nil
		}
	}

	return "", &cfg, nil, nil
}

// GetSourceConfig 获取源配置
func (m *McpInstance) GetSourceConfig() (string, *McpServersConfig, *McpConfig, error) {
	return parseMcpServersConfig(m.SourceConfig)
}

// GetTargetConfig 获取目标配置
func (m *McpInstance) GetTargetConfig() (string, *McpServersConfig, *McpConfig, error) {
	return parseMcpServersConfig(m.TargetConfig)
}

// GetPublicProxyConfig 获取公共代理配置
func (m *McpInstance) GetPublicProxyConfig() (string, *McpServersConfig, *McpConfig, error) {
	return parseMcpServersConfig(m.PublicProxyConfig)
}
