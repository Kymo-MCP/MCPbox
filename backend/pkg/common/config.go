package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ServerConfig server configuration
type ServerConfig struct {
	GrpcPort int `mapstructure:"grpcPort"` // gRPC port
	HttpPort int `mapstructure:"httpPort"` // HTTP port
}

type StorageConfig struct {
	RootPath   string `mapstructure:"rootPath"`
	CodePath   string `mapstructure:"codePath"`
	StaticPath string `mapstructure:"staticPath"`
}

type CodeConfig struct {
	Upload UploadConfig `mapstructure:"upload"`
}

type UploadConfig struct {
	MaxFileSize       int      `mapstructure:"maxFileSize"`
	AllowedExtensions []string `mapstructure:"allowedExtensions"`
}

type PathPrefixConfig struct {
	PathPrefix string `mapstructure:"pathPrefix"`
}

type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

// InitKubernetesConfig initialize Kubernetes configuration
type InitKubernetesConfig struct {
	Namespace             string `mapstructure:"namespace"`
	DefaultConfigFilePath string `mapstructure:"defaultConfigFilePath"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MarketConfig market configuration
type MarketConfig struct {
	// Host address
	Host string `mapstructure:"host"`
	// Secret key
	SecretKey string `mapstructure:"secretKey"`
	// Customer UUID
	CustomerUuid string `mapstructure:"customerUuid"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type Services struct {
	McpMarket *Service `mapstructure:"mcpMarket"`
	McpAuthz  *Service `mapstructure:"mcpAuthz"`
}

type Service struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// findProjectRoot find project root directory
func findProjectRoot() (string, error) {
	// Start searching from current directory upwards
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	// Search up to 10 levels of directories
	for i := 0; i < 10; i++ {
		// Check if go.mod file exists
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Check if .git directory exists
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		// Search upwards
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Already reached root directory
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

// FindConfigFile uses viper to find configuration files in multiple locations.
func FindConfigFile(cfgFileName string) (string, error) {
	v := viper.New()

	// Set configuration file name (without extension)
	configName := cfgFileName
	if ext := filepath.Ext(cfgFileName); ext != "" {
		configName = cfgFileName[:len(cfgFileName)-len(ext)]
		v.SetConfigType(ext[1:]) // Remove the dot
	} else {
		v.SetConfigType("yaml") // Default type
	}

	v.SetConfigName(configName)

	// Add configuration file search paths
	configPaths := getConfigSearchPaths()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Try to find configuration file
	if err := v.ReadInConfig(); err != nil {
		return "", fmt.Errorf("configuration file not found, search paths: %v, error: %v", configPaths, err)
	}

	return v.ConfigFileUsed(), nil
}

// getConfigSearchPaths get configuration file search path list
func getConfigSearchPaths() []string {
	var configPaths []string

	// Configuration root directory specified by environment variable
	if configRoot := os.Getenv("CONFIG_ROOT"); configRoot != "" {
		configPaths = append(configPaths, configRoot)
	} else {
		configPaths = append(configPaths, "/etc/github.com/kymo-mcp/mcpcan")
	}

	// Config folder in project root directory
	if projectRoot, err := findProjectRoot(); err == nil {
		configPaths = append(configPaths, filepath.Join(projectRoot, "config"))
	}

	// Relative paths
	configPaths = append(configPaths,
		"./config",
		"../config",
		"../../config",
		".",
	)

	// Directory where executable file is located
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		configPaths = append(configPaths,
			execDir,
			filepath.Join(execDir, "config"),
		)
	}

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		configPaths = append(configPaths,
			cwd,
			filepath.Join(cwd, "config"),
		)
	}

	return configPaths
}

// LoadConfigWithViper uses viper to load the configuration file into the specified struct.
func LoadConfigWithViper(cfgFileName string, config interface{}) error {
	v := viper.New()

	// Set configuration file name and type
	configName := cfgFileName
	if ext := filepath.Ext(cfgFileName); ext != "" {
		configName = cfgFileName[:len(cfgFileName)-len(ext)]
		v.SetConfigType(ext[1:])
	} else {
		v.SetConfigType("yaml")
	}

	v.SetConfigName(configName)

	// Add search paths
	configPaths := getConfigSearchPaths()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configuration file, search paths: %v, error: %v", configPaths, err)
	}

	// Parse configuration to struct
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return nil
}
