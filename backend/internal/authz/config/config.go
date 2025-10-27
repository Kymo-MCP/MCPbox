package config

import (
	"fmt"

	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/version"

	"github.com/spf13/viper"
)

// Config 表示配置结构
type Config struct {
	ServiceName string                `mapstructure:"-"`
	VersionInfo *version.VersionInfo  `mapstructure:"-"`
	Server      ServerConfig          `mapstructure:"server"`
	Storage     common.StorageConfig  `mapstructure:"storage"`
	Domain      string                `mapstructure:"domain"`
	Database    common.DatabaseConfig `mapstructure:"database"`
	Log         common.LogConfig      `mapstructure:"log"`
	Secret      string                `mapstructure:"secret"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Expires int    `mapstructure:"expires"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	GrpcPort int `mapstructure:"grpcPort"`
	HttpPort int `mapstructure:"httpPort"`
}

var GlobalConfig *Config
var serviceName = "authz"
var cfgFileName = "authz.yaml"

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}

// Load 加载配置文件
func Load() error {
	v := viper.New()
	v.SetConfigType("yaml")

	// 如果未指定配置文件路径，尝试自动查找
	var err error
	configPath, err := common.FindConfigFile(cfgFileName)
	if err != nil {
		return err
	}

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 追加 Version 信息
	config.ServiceName = serviceName
	config.VersionInfo = version.GetVersionInfo()

	GlobalConfig = &config

	return nil
}
