package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	GrpcPort int `mapstructure:"grpcPort"` // gRPC 端口
	HttpPort int `mapstructure:"httpPort"` // HTTP 端口
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

// InitKubernetesConfig 初始化Kubernetes配置
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

// MarketConfig 市场配置
type MarketConfig struct {
	// 主机地址
	Host string `mapstructure:"host"`
	// 密钥
	SecretKey string `mapstructure:"secretKey"`
	// 客户UUID
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

// findProjectRoot 查找项目根目录
func findProjectRoot() (string, error) {
	// 从当前目录开始向上查找
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	// 最多向上查找 10 层目录
	for i := 0; i < 10; i++ {
		// 检查是否存在 go.mod 文件
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// 检查是否存在 .git 目录
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		// 向上查找
		parent := filepath.Dir(dir)
		if parent == dir {
			break // 已经到达根目录
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

// FindConfigFile 使用 viper 在多个位置查找配置文件
func FindConfigFile(cfgFileName string) (string, error) {
	v := viper.New()

	// 设置配置文件名（不包含扩展名）
	configName := cfgFileName
	if ext := filepath.Ext(cfgFileName); ext != "" {
		configName = cfgFileName[:len(cfgFileName)-len(ext)]
		v.SetConfigType(ext[1:]) // 去掉点号
	} else {
		v.SetConfigType("yaml") // 默认类型
	}

	v.SetConfigName(configName)

	// 添加配置文件搜索路径
	configPaths := getConfigSearchPaths()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 尝试查找配置文件
	if err := v.ReadInConfig(); err != nil {
		return "", fmt.Errorf("配置文件未找到，搜索路径: %v, 错误: %v", configPaths, err)
	}

	return v.ConfigFileUsed(), nil
}

// getConfigSearchPaths 获取配置文件搜索路径列表
func getConfigSearchPaths() []string {
	var configPaths []string

	// 环境变量指定的配置根目录
	if configRoot := os.Getenv("CONFIG_ROOT"); configRoot != "" {
		configPaths = append(configPaths, configRoot)
	} else {
		configPaths = append(configPaths, "/etc/qm-mcp-server")
	}

	// 项目根目录的 config 文件夹
	if projectRoot, err := findProjectRoot(); err == nil {
		configPaths = append(configPaths, filepath.Join(projectRoot, "config"))
	}

	// 相对路径
	configPaths = append(configPaths,
		"./config",
		"../config",
		"../../config",
		".",
	)

	// 可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		configPaths = append(configPaths,
			execDir,
			filepath.Join(execDir, "config"),
		)
	}

	// 当前工作目录
	if cwd, err := os.Getwd(); err == nil {
		configPaths = append(configPaths,
			cwd,
			filepath.Join(cwd, "config"),
		)
	}

	return configPaths
}

// LoadConfigWithViper 使用 viper 加载配置文件到指定结构体
func LoadConfigWithViper(cfgFileName string, config interface{}) error {
	v := viper.New()

	// 设置配置文件名和类型
	configName := cfgFileName
	if ext := filepath.Ext(cfgFileName); ext != "" {
		configName = cfgFileName[:len(cfgFileName)-len(ext)]
		v.SetConfigType(ext[1:])
	} else {
		v.SetConfigType("yaml")
	}

	v.SetConfigName(configName)

	// 添加搜索路径
	configPaths := getConfigSearchPaths()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败，搜索路径: %v, 错误: %v", configPaths, err)
	}

	// 解析配置到结构体
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}
