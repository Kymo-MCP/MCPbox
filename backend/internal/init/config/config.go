package config

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/version"

	"github.com/spf13/viper"
)

var GlobalConfig *InitConfig

// InitConfig 表示初始化配置结构
type InitConfig struct {
	ServiceName string                      `mapstructure:"-"`
	VersionInfo *version.VersionInfo        `mapstructure:"-"`
	Database    common.DatabaseConfig       `mapstructure:"database"`
	Storage     common.StorageConfig        `mapstructure:"storage"`
	Log         common.LogConfig            `mapstructure:"log"`
	Init        InitUserConfig              `mapstructure:"init"`
	Kubernetes  common.InitKubernetesConfig `mapstructure:"kubernetes"`
}

// InitUserConfig 初始化用户配置
type InitUserConfig struct {
	AdminUsername        string `mapstructure:"admin_username"`
	AdminPassword        string `mapstructure:"admin_password"`
	AdminNickname        string `mapstructure:"admin_nickname"`
	AdminRoleName        string `mapstructure:"admin_role_name"`
	AdminRoleDescription string `mapstructure:"admin_role_description"`
	AdminRoleLevel       int    `mapstructure:"admin_role_level"`
	AdminDataScope       string `mapstructure:"admin_data_scope"`
}

// GetInitConfig 获取全局初始化配置
func GetInitConfig() *InitConfig {
	return GlobalConfig
}

var serviceName = "init"
var cfgFileName = "init.yaml"

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
	var config InitConfig
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 追加 Version 信息
	config.ServiceName = serviceName
	config.VersionInfo = version.GetVersionInfo()

	GlobalConfig = &config

	return nil
}
