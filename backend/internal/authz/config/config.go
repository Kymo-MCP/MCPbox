package config

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/version"

	"github.com/spf13/viper"
)

// Config represents configuration structure
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

// JWTConfig JWT configuration
type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Expires int    `mapstructure:"expires"`
}

// ServerConfig server configuration
type ServerConfig struct {
	GrpcPort int `mapstructure:"grpcPort"`
	HttpPort int `mapstructure:"httpPort"`
}

var GlobalConfig *Config
var serviceName = "authz"
var cfgFileName = "authz.yaml"

// GetConfig gets global configuration
func GetConfig() *Config {
	return GlobalConfig
}

// Load loads configuration file
func Load() error {
	v := viper.New()
	v.SetConfigType("yaml")

	// If configuration file path is not specified, try to find it automatically
	var err error
	configPath, err := common.FindConfigFile(cfgFileName)
	if err != nil {
		return err
	}

	// Set configuration file path
	v.SetConfigFile(configPath)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Append version information
	config.ServiceName = serviceName
	config.VersionInfo = version.GetVersionInfo()

	GlobalConfig = &config

	return nil
}
