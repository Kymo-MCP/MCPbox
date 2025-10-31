package config

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"github.com/kymo-mcp/mcpcan/pkg/version"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

// Config 表示配置结构
type Config struct {
	ServiceName string                `mapstructure:"-"`
	VersionInfo *version.VersionInfo  `mapstructure:"-"`
	Server      common.ServerConfig   `mapstructure:"server"`
	Services    common.Services       `mapstructure:"services"`
	Domain      string                `mapstructure:"domain"`
	Database    common.DatabaseConfig `mapstructure:"database"`
	Code        common.CodeConfig     `mapstructure:"code"`
	Market      common.MarketConfig   `mapstructure:"market"`
	Log         common.LogConfig      `mapstructure:"log"`
	Secret      string                `mapstructure:"secret"`
	Storage     common.StorageConfig  `mapstructure:"storage"`
}

var serviceName = "market"
var cfgFileName = "market.yaml"

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}

// Load 加载配置文件
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// 如果未指定配置文件路径，尝试自动查找
	var err error
	configPath, err := common.FindConfigFile(cfgFileName)
	if err != nil {
		return nil, err
	}

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if config.Code.Upload.MaxFileSize == 0 {
		config.Code.Upload.MaxFileSize = 100
	}

	if config.Code.Upload.AllowedExtensions == nil {
		config.Code.Upload.AllowedExtensions = []string{".zip", ".tar.gz", ".tar", ".rar"}
	}

	if config.Storage.RootPath == "" {
		config.Storage.RootPath = "/app/data"
	}
	utils.MkdirP(config.Storage.RootPath)

	if config.Storage.CodePath == "" {
		config.Storage.CodePath = "/app/data/code-package"
	}
	utils.MkdirP(config.Storage.CodePath)

	if config.Storage.StaticPath == "" {
		config.Storage.StaticPath = "/app/data/static"
	}
	utils.MkdirP(config.Storage.StaticPath)

	// 追加 Version 信息
	config.ServiceName = serviceName
	config.VersionInfo = version.GetVersionInfo()

	GlobalConfig = &config

	return &config, nil
}
