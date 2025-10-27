package scheduler

import (
	"context"
	"sync"
	"time"
)

// GlobalSchedulerEntry 全局调度器入口
type GlobalSchedulerEntry struct {
	ctx         context.Context
	taskManager TaskManager
	once        sync.Once
}

var (
	// globalEntry 全局调度器实例
	globalEntry *GlobalSchedulerEntry
	// globalOnce 确保全局实例只初始化一次
	globalOnce sync.Once
)

// GetGlobalScheduler 获取全局调度器实例
func GetGlobalScheduler() *GlobalSchedulerEntry {
	globalOnce.Do(func() {
		// 创建内存任务存储（可以替换为数据库存储）
		taskRepo := NewMemoryTaskRepository()

		// 创建调度器
		scheduler := NewTaskScheduler(taskRepo)

		// 创建任务管理器
		taskManager := NewTaskManager(scheduler)

		globalEntry = &GlobalSchedulerEntry{
			ctx:         context.Background(),
			taskManager: taskManager,
		}
	})
	return globalEntry
}

// Start 启动全局调度器
func (g *GlobalSchedulerEntry) Start() error {
	return g.taskManager.GetScheduler().Start(g.ctx)
}

// Stop 停止全局调度器
func (g *GlobalSchedulerEntry) Stop() error {
	return g.taskManager.GetScheduler().Stop()
}

// GetTaskManager 获取任务管理器
func (g *GlobalSchedulerEntry) GetTaskManager() TaskManager {
	return g.taskManager
}

// RegisterTaskFunc 注册任务函数
func (g *GlobalSchedulerEntry) RegisterTaskFunc(name string, fn TaskFunc) error {
	return g.taskManager.RegisterTaskFunc(name, fn)
}

// 便捷函数，直接使用全局实例

// Start 启动全局调度器（便捷函数）
func Start(ctx context.Context) error {
	return GetGlobalScheduler().Start()
}

// Stop 停止全局调度器（便捷函数）
func Stop() error {
	return GetGlobalScheduler().Stop()
}

// RegisterTaskFunc 注册任务函数（便捷函数）
func RegisterTaskFunc(name string, fn TaskFunc) error {
	return GetGlobalScheduler().RegisterTaskFunc(name, fn)
}

// GetTaskManager 获取任务管理器（便捷函数）
func GetTaskManager() TaskManager {
	return GetGlobalScheduler().GetTaskManager()
}

// CreateCronTask 创建Cron任务（便捷函数）
func CreateCronTask(id, name, cronExpr, funcName string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().CreateCronTask(id, name, cronExpr, funcName)
}

// CreateTimerTask 创建定时任务（便捷函数）
func CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().CreateTimerTask(id, name, executeAt, funcName)
}

// GetTask 获取任务（便捷函数）
func GetTask(taskID string) (Task, error) {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().GetTask(taskID)
}

// RemoveTask 移除任务（便捷函数）
func RemoveTask(taskID string) error {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().RemoveTask(taskID)
}

// ListTasks 列出所有任务（便捷函数）
func ListTasks() []Task {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().ListTasks()
}

// IsRunning 检查调度器是否运行中（便捷函数）
func IsRunning() bool {
	return GetGlobalScheduler().GetTaskManager().GetScheduler().IsRunning()
}
