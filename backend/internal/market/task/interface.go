package task

import (
	"context"

	"qm-mcp-server/pkg/database/model"
)

// TaskManager 任务管理器接口
// 负责管理全局任务的生命周期
type TaskManager interface {
	// SetupGlobalTasks 设置全局任务
	// 初始化所有需要的全局任务
	SetupGlobalTasks(ctx context.Context) error

	// StartMonitoring 开始监控
	// 启动所有监控任务
	StartMonitoring(ctx context.Context) error

	// StopMonitoring 停止监控
	// 停止所有监控任务
	StopMonitoring(ctx context.Context) error
}

// ContainerMonitor 容器监控接口
// 负责容器状态监控和管理
type ContainerMonitor interface {
	// MonitorContainers 监控所有容器
	// 检查所有活跃实例的容器状态
	MonitorContainers(ctx context.Context) error

	// CheckContainer 检查单个容器
	// 检查指定实例的容器状态并在需要时重建
	CheckContainer(ctx context.Context, instance *model.McpInstance) error
}
