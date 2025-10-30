package model

import (
	"encoding/json"
	"time"
)

type McpTemplate struct {
	ID                   uint            `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	Name                 string          `gorm:"size:200;not null;comment:实例名称" json:"Name"`
	Port                 int32           `gorm:"default:0;comment:端口号" json:"Port"`
	InitScript           string          `gorm:"type:text;comment:初始化脚本" json:"InitScript"`
	Command              string          `gorm:"type:text;comment:启动命令" json:"Command"`
	EnvironmentVariables json.RawMessage `gorm:"type:json;comment:环境变量 (JSON格式)" json:"EnvironmentVariables"`
	VolumeMounts         json.RawMessage `gorm:"type:json;comment:卷挂载配置列表 (JSON格式)" json:"VolumeMounts"`
	StartupTimeout       int32           `gorm:"default:0;comment:启动超时时间（秒）" json:"StartupTimeout"`
	RunningTimeout       int32           `gorm:"default:0;comment:运行超时时间（秒）" json:"RunningTimeout"`
	EnvironmentID        int32           `gorm:"default:0;comment:环境ID" json:"environmentID"`
	PackageID            string          `gorm:"size:100;comment:包ID" json:"packageID"`
	AccessType           AccessType      `gorm:"size:20;not null;comment:访问类型 (直连-direct/代理-proxy/托管-hosting)" json:"accessType"`
	McpProtocol          McpProtocol     `gorm:"size:20;not null;comment:MCP协议 (SSE-1/StreamableHttp-2/Stdio-3)" json:"mcpProtocol"`
	McpServers           json.RawMessage `gorm:"type:json;comment:MCP服务器配置 (JSON格式)" json:"mcpServers"`
	ImgAddress           string          `gorm:"size:100;not null;comment:镜像地址" json:"imgAddress"`
	McpServerID          string          `gorm:"size:100;comment:MCP 服务器ID" json:"mcpServerID"`
	Tokens               json.RawMessage `gorm:"type:json;comment:MCP 实例令牌 (JSON格式)" json:"tokens"`
	Notes                string          `gorm:"type:text;comment:备注" json:"notes"`
	ServicePath          string          `gorm:"size:100;not null;default:'';comment:MCP 服务路径" json:"servicePath"`
	IconPath             string          `gorm:"size:100;not null;default:'';comment:MCP 图标路径" json:"iconPath"`
	CreatedAt            time.Time       `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt            time.Time       `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

func (McpTemplate) TableName() string {
	return "mcp_template"
}
