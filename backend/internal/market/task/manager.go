package task

import (
	"context"
	"fmt"

	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/scheduler"

	"go.uber.org/zap"
)

// TaskManagerImpl 任务管理器实现
type TaskManagerImpl struct {
	// instanceRepo 实例数据库操作
	instanceRepo *mysql.McpInstanceRepository

	// scheduler 调度器
	scheduler scheduler.Scheduler

	// logger 日志记录器
	logger *zap.Logger

	// monitorTaskID 监控任务ID
	monitorTaskID string

	// isRunning 是否正在运行
	isRunning bool
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(
	instanceRepo *mysql.McpInstanceRepository,
	scheduler scheduler.Scheduler,
	logger *zap.Logger,
) TaskManager {
	return &TaskManagerImpl{
		instanceRepo: instanceRepo,
		scheduler:    scheduler,
		logger:       logger,
	}
}

// SetupGlobalTasks 设置全局任务
func (tm *TaskManagerImpl) SetupGlobalTasks(ctx context.Context) error {
	tm.logger.Info("开始设置全局任务")

	// 创建容器监控器
	containerMonitor := NewContainerMonitor(tm.instanceRepo, tm.logger)

	// 创建任务函数适配器
	taskFunc := func(ctx context.Context) error {
		return containerMonitor.MonitorContainers(ctx)
	}

	// 创建容器监控任务 - 使用Cron任务，每30秒执行一次
	// Cron表达式: */30 * * * * * (每30秒执行一次)
	task, err := scheduler.NewCronTask(
		"global_container_monitor",
		"全局容器监控任务",
		"*/30 * * * * *", // 每30秒执行一次
		"container_monitor",
		taskFunc,
	)
	if err != nil {
		tm.logger.Error("创建全局容器监控任务失败",
			zap.Error(err))
		return fmt.Errorf("创建任务失败: %w", err)
	}

	// 添加任务到调度器
	if err := tm.scheduler.AddTask(task); err != nil {
		tm.logger.Error("添加全局容器监控任务失败",
			zap.String("task_id", task.GetID()),
			zap.Error(err))
		return fmt.Errorf("添加任务失败: %w", err)
	}

	// 保存监控任务ID
	tm.monitorTaskID = task.GetID()

	tm.logger.Info("全局容器监控任务设置成功",
		zap.String("task_id", task.GetID()),
		zap.String("task_name", task.GetName()),
		zap.String("cron_expr", "*/30 * * * * *"))

	return nil
}

// StartMonitoring 开始监控
func (tm *TaskManagerImpl) StartMonitoring(ctx context.Context) error {
	if tm.isRunning {
		tm.logger.Warn("任务管理器已在运行中")
		return nil
	}

	tm.logger.Info("启动任务监控")

	// 启动调度器
	err := tm.scheduler.Start(ctx)
	if err != nil {
		tm.logger.Error("启动调度器失败", zap.Error(err))
		return fmt.Errorf("启动调度器失败: %w", err)
	}

	tm.isRunning = true
	tm.logger.Info("任务监控启动成功")

	return nil
}

// StopMonitoring 停止监控
func (tm *TaskManagerImpl) StopMonitoring(ctx context.Context) error {
	if !tm.isRunning {
		tm.logger.Warn("任务管理器未在运行")
		return nil
	}

	tm.logger.Info("停止任务监控")

	// 停止调度器
	err := tm.scheduler.Stop()
	if err != nil {
		tm.logger.Error("停止调度器失败", zap.Error(err))
		return fmt.Errorf("停止调度器失败: %w", err)
	}

	tm.isRunning = false
	tm.logger.Info("任务监控停止成功")

	return nil
}

// IsRunning 检查是否正在运行
func (tm *TaskManagerImpl) IsRunning() bool {
	return tm.isRunning
}

// GetMonitorTaskID 获取监控任务ID
func (tm *TaskManagerImpl) GetMonitorTaskID() string {
	return tm.monitorTaskID
}
