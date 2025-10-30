package app

import (
	"context"
	"fmt"
	"os"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	"strconv"
)

func (a *App) initDefaultKubernetesEnvironment(ctx context.Context, adminUser *model.SysUser) (*model.McpEnvironment, error) {
	if len(a.config.Kubernetes.DefaultConfigFilePath) == 0 || len(a.config.Kubernetes.Namespace) == 0 {
		return nil, fmt.Errorf("default config file path or namespace is empty")
	}
	// 检查是否存在名为 Default-Kubernetes-Env 的环境；不存在则创建
	const defaultName = common.EnvironmentDefaultName

	// 先按名称查询
	existingEnv, err := mysql.McpEnvironmentRepo.FindByName(ctx, defaultName)
	if err == nil && existingEnv != nil {
		// 已存在，无需处理
		return existingEnv, nil
	}

	namespace := a.config.Kubernetes.Namespace
	defaultConfigFilePath := a.config.Kubernetes.DefaultConfigFilePath

	// 读取默认配置文件内容
	configContent, err := os.ReadFile(defaultConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read default config file: %v", err)
	}

	strConfig := string(configContent)
	if strConfig == "" {
		return nil, fmt.Errorf("default config file is empty")
	}

	// 构造默认环境（Kubernetes）
	env := &model.McpEnvironment{
		Name:        defaultName,
		Environment: model.McpEnvironmentKubernetes,
		Config:      strConfig,
		CreatorID:   strconv.FormatUint(uint64(adminUser.UserID), 10),
		Namespace:   namespace,
	}

	// 验证与准备
	if vErr := env.ValidateForCreate(); vErr != nil {
		return nil, fmt.Errorf("default environment validation failed: %v", vErr)
	}
	env.PrepareForCreate()

	// 创建记录
	if err := mysql.McpEnvironmentRepo.Create(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to create default environment: %v", err)
	}

	fmt.Printf("Default environment created successfully with ID: %d\n", env.ID)
	return env, nil
}
