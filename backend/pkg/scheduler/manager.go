package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// DefaultTaskManager 默认任务管理器实现
type DefaultTaskManager struct {
	scheduler Scheduler                // 调度器
	taskFuncs map[string]TaskFunc      // 注册的任务函数
	mu        sync.RWMutex             // 读写锁
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(scheduler Scheduler) *DefaultTaskManager {
	return &DefaultTaskManager{
		scheduler: scheduler,
		taskFuncs: make(map[string]TaskFunc),
	}
}

// RegisterTaskFunc 注册任务函数
func (tm *DefaultTaskManager) RegisterTaskFunc(name string, fn TaskFunc) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if name == "" {
		return fmt.Errorf("任务函数名称不能为空")
	}

	if fn == nil {
		return fmt.Errorf("任务函数不能为空")
	}

	if _, exists := tm.taskFuncs[name]; exists {
		return fmt.Errorf("任务函数 %s 已存在", name)
	}

	tm.taskFuncs[name] = fn
	return nil
}

// GetTaskFunc 获取任务函数
func (tm *DefaultTaskManager) GetTaskFunc(name string) (TaskFunc, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	fn, exists := tm.taskFuncs[name]
	if !exists {
		return nil, fmt.Errorf("任务函数 %s 不存在", name)
	}

	return fn, nil
}

// CreateCronTask 创建Cron任务
func (tm *DefaultTaskManager) CreateCronTask(id, name, cronExpr, funcName string) (Task, error) {
	if id == "" {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	if name == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}

	if cronExpr == "" {
		return nil, fmt.Errorf("Cron表达式不能为空")
	}

	if funcName == "" {
		return nil, fmt.Errorf("任务函数名称不能为空")
	}

	// 获取任务函数
	taskFunc, err := tm.GetTaskFunc(funcName)
	if err != nil {
		return nil, fmt.Errorf("获取任务函数失败: %w", err)
	}

	// 创建Cron任务
	task, err := NewCronTask(id, name, cronExpr, funcName, taskFunc)
	if err != nil {
		return nil, fmt.Errorf("创建Cron任务失败: %w", err)
	}

	// 添加到调度器
	err = tm.scheduler.AddTask(task)
	if err != nil {
		return nil, fmt.Errorf("添加任务到调度器失败: %w", err)
	}

	return task, nil
}

// CreateTimerTask 创建定时任务
func (tm *DefaultTaskManager) CreateTimerTask(id, name string, executeAt time.Time, funcName string) (Task, error) {
	if id == "" {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	if name == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}

	if funcName == "" {
		return nil, fmt.Errorf("任务函数名称不能为空")
	}

	if executeAt.Before(time.Now()) {
		return nil, fmt.Errorf("执行时间不能早于当前时间")
	}

	// 获取任务函数
	taskFunc, err := tm.GetTaskFunc(funcName)
	if err != nil {
		return nil, fmt.Errorf("获取任务函数失败: %w", err)
	}

	// 创建定时任务
	task := NewTimerTask(id, name, executeAt, funcName, taskFunc)

	// 添加到调度器
	err = tm.scheduler.AddTask(task)
	if err != nil {
		return nil, fmt.Errorf("添加任务到调度器失败: %w", err)
	}

	return task, nil
}

// GetScheduler 获取调度器
func (tm *DefaultTaskManager) GetScheduler() Scheduler {
	return tm.scheduler
}

// ListTaskFuncs 列出所有注册的任务函数
func (tm *DefaultTaskManager) ListTaskFuncs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	funcs := make([]string, 0, len(tm.taskFuncs))
	for name := range tm.taskFuncs {
		funcs = append(funcs, name)
	}

	return funcs
}

// RemoveTaskFunc 移除任务函数
func (tm *DefaultTaskManager) RemoveTaskFunc(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.taskFuncs[name]; !exists {
		return fmt.Errorf("任务函数 %s 不存在", name)
	}

	delete(tm.taskFuncs, name)
	return nil
}

// GetTaskCount 获取任务数量
func (tm *DefaultTaskManager) GetTaskCount() int {
	tasks := tm.scheduler.ListTasks()
	return len(tasks)
}

// GetTaskFuncCount 获取注册的任务函数数量
func (tm *DefaultTaskManager) GetTaskFuncCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.taskFuncs)
}

// IsSchedulerRunning 检查调度器是否运行中
func (tm *DefaultTaskManager) IsSchedulerRunning() bool {
	return tm.scheduler.IsRunning()
}