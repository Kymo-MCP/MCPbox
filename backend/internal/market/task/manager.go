package task

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/scheduler"

	"go.uber.org/zap"
)

// TaskManagerImpl task manager implementation
type TaskManagerImpl struct {
	// instanceRepo instance database operations
	instanceRepo *mysql.McpInstanceRepository

	// scheduler scheduler
	scheduler scheduler.Scheduler

	// logger logger
	logger *zap.Logger

	// monitorTaskID monitor task ID
	monitorTaskID string

	// isRunning whether it is running
	isRunning bool
}

// NewTaskManager creates a new task manager
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

// SetupGlobalTasks sets up global tasks
func (tm *TaskManagerImpl) SetupGlobalTasks(ctx context.Context) error {
	tm.logger.Info("starting to set up global tasks")

	// Create container monitor
	containerMonitor := NewContainerMonitor(tm.instanceRepo, tm.logger)

	// Create task function adapter
	taskFunc := func(ctx context.Context) error {
		return containerMonitor.MonitorContainers(ctx)
	}

	// Create container monitoring task - using Cron task, execute every 30 seconds
	// Cron expression: */30 * * * * * (execute every 30 seconds)
	task, err := scheduler.NewCronTask(
		"global_container_monitor",
		"global container monitoring task",
		"*/30 * * * * *", // execute every 30 seconds
		"container_monitor",
		taskFunc,
	)
	if err != nil {
		tm.logger.Error("failed to create global container monitoring task",
			zap.Error(err))
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Add task to scheduler
	if err := tm.scheduler.AddTask(task); err != nil {
		tm.logger.Error("failed to add global container monitoring task",
			zap.String("task_id", task.GetID()),
			zap.Error(err))
		return fmt.Errorf("failed to add task: %w", err)
	}

	// Save monitor task ID
	tm.monitorTaskID = task.GetID()

	tm.logger.Info("global container monitoring task set up successfully",
		zap.String("task_id", task.GetID()),
		zap.String("task_name", task.GetName()),
		zap.String("cron_expr", "*/30 * * * * *"))

	return nil
}

// StartMonitoring starts monitoring
func (tm *TaskManagerImpl) StartMonitoring(ctx context.Context) error {
	if tm.isRunning {
		tm.logger.Warn("task manager is already running")
		return nil
	}

	tm.logger.Info("starting task monitoring")

	// Start scheduler
	err := tm.scheduler.Start(ctx)
	if err != nil {
		tm.logger.Error("failed to start scheduler", zap.Error(err))
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	tm.isRunning = true
	tm.logger.Info("task monitoring started successfully")

	return nil
}

// StopMonitoring stops monitoring
func (tm *TaskManagerImpl) StopMonitoring(ctx context.Context) error {
	if !tm.isRunning {
		tm.logger.Warn("task manager is not running")
		return nil
	}

	tm.logger.Info("stopping task monitoring")

	// Stop scheduler
	err := tm.scheduler.Stop()
	if err != nil {
		tm.logger.Error("failed to stop scheduler", zap.Error(err))
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	tm.isRunning = false
	tm.logger.Info("task monitoring stopped successfully")

	return nil
}

// IsRunning checks if it is running
func (tm *TaskManagerImpl) IsRunning() bool {
	return tm.isRunning
}

// GetMonitorTaskID gets monitor task ID
func (tm *TaskManagerImpl) GetMonitorTaskID() string {
	return tm.monitorTaskID
}
