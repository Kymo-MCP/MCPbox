package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// MemoryTaskRepository 内存任务存储实现
type MemoryTaskRepository struct {
	tasks map[string]Task // 任务映射
	mu    sync.RWMutex    // 读写锁
}

// NewMemoryTaskRepository 创建新的内存任务存储
func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]Task),
	}
}

// SaveTask 保存任务
func (r *MemoryTaskRepository) SaveTask(ctx context.Context, task Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task == nil {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeTaskCannotBeEmpty))
	}

	taskID := task.GetID()
	if taskID == "" {
		return fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeTaskIDCannotBeEmpty))
	}

	r.tasks[taskID] = task
	return nil
}

// GetTask 获取任务
func (r *MemoryTaskRepository) GetTask(taskID string) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if taskID == "" {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	task, exists := r.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务ID %s 不存在", taskID)
	}

	return task, nil
}

// ListTasks 列出任务
func (r *MemoryTaskRepository) ListTasks() ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (r *MemoryTaskRepository) DeleteTask(taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if taskID == "" {
		return fmt.Errorf("任务ID不能为空")
	}

	if _, exists := r.tasks[taskID]; !exists {
		return fmt.Errorf("任务ID %s 不存在", taskID)
	}

	delete(r.tasks, taskID)
	return nil
}

// UpdateTaskStatus 更新任务状态
func (r *MemoryTaskRepository) UpdateTaskStatus(taskID string, status TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if taskID == "" {
		return fmt.Errorf("任务ID不能为空")
	}

	task, exists := r.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务ID %s 不存在", taskID)
	}

	// 由于Task接口没有SetStatus方法，这里只能通过类型断言来更新状态
	switch t := task.(type) {
	case *CronTask:
		t.setStatus(status)
	case *TimerTask:
		t.setStatus(status)
	default:
		return fmt.Errorf("不支持的任务类型")
	}

	return nil
}

// GetTaskCount 获取任务数量
func (r *MemoryTaskRepository) GetTaskCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tasks)
}

// Clear 清空所有任务
func (r *MemoryTaskRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = make(map[string]Task)
	return nil
}
