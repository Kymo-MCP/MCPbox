package scheduler

import (
	"context"
	"time"
)

// TaskFunc 任务执行函数类型
type TaskFunc func(ctx context.Context) error

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 待执行
	TaskStatusRunning                     // 执行中
	TaskStatusCompleted                   // 已完成
	TaskStatusFailed                      // 执行失败
	TaskStatusCancelled                   // 已取消
)

// TaskType 任务类型
type TaskType int

const (
	TaskTypeCron  TaskType = iota // Cron 定时任务
	TaskTypeTimer                 // 一次性定时任务
)

// Task 任务接口
type Task interface {
	// GetID 获取任务ID
	GetID() string
	// GetName 获取任务名称
	GetName() string
	// GetType 获取任务类型
	GetType() TaskType
	// GetStatus 获取任务状态
	GetStatus() TaskStatus
	// Execute 执行任务
	Execute(ctx context.Context) error
	// Cancel 取消任务
	Cancel() error
	// GetNextRunTime 获取下次执行时间
	GetNextRunTime() *time.Time
	// GetLastRunTime 获取上次执行时间
	GetLastRunTime() *time.Time
	// GetCreatedAt 获取创建时间
	GetCreatedAt() time.Time
}

// Scheduler 调度器接口
type Scheduler interface {
	// Start 启动调度器
	Start(ctx context.Context) error
	// Stop 停止调度器
	Stop() error
	// AddTask 添加任务
	AddTask(task Task) error
	// RemoveTask 移除任务
	RemoveTask(taskID string) error
	// GetTask 获取任务
	GetTask(taskID string) (Task, error)
	// ListTasks 列出所有任务
	ListTasks() []Task
	// IsRunning 检查调度器是否运行中
	IsRunning() bool
}

// TaskManager 任务管理器接口
type TaskManager interface {
	// RegisterTaskFunc 注册任务函数
	RegisterTaskFunc(name string, fn TaskFunc) error
	// GetTaskFunc 获取任务函数
	GetTaskFunc(name string) (TaskFunc, error)
	// CreateCronTask 创建Cron任务
	CreateCronTask(id, name, cronExpr, funcName string) (Task, error)
	// CreateTimerTask 创建定时任务
	CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error)
	// GetScheduler 获取调度器
	GetScheduler() Scheduler
}

// TaskRepository 任务存储接口
type TaskRepository interface {
	// SaveTask 保存任务
	SaveTask(ctx context.Context, task Task) error
	// GetTask 获取任务
	GetTask(taskID string) (Task, error)
	// ListTasks 列出任务
	ListTasks() ([]Task, error)
	// DeleteTask 删除任务
	DeleteTask(taskID string) error
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(taskID string, status TaskStatus) error
}
