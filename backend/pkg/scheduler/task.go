package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// BaseTask 基础任务结构体
type BaseTask struct {
	id          string      // 任务ID
	name        string      // 任务名称
	taskType    TaskType    // 任务类型
	status      TaskStatus  // 任务状态
	funcName    string      // 任务函数名称
	taskFunc    TaskFunc    // 任务执行函数
	createdAt   time.Time   // 创建时间
	lastRunTime *time.Time  // 上次执行时间
	nextRunTime *time.Time  // 下次执行时间
	mu          sync.RWMutex // 读写锁
}

// GetID 获取任务ID
func (t *BaseTask) GetID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

// GetName 获取任务名称
func (t *BaseTask) GetName() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.name
}

// GetType 获取任务类型
func (t *BaseTask) GetType() TaskType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.taskType
}

// GetStatus 获取任务状态
func (t *BaseTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetNextRunTime 获取下次执行时间
func (t *BaseTask) GetNextRunTime() *time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nextRunTime
}

// GetLastRunTime 获取上次执行时间
func (t *BaseTask) GetLastRunTime() *time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastRunTime
}

// GetCreatedAt 获取创建时间
func (t *BaseTask) GetCreatedAt() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.createdAt
}

// setStatus 设置任务状态（内部方法）
func (t *BaseTask) setStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

// setLastRunTime 设置上次执行时间（内部方法）
func (t *BaseTask) setLastRunTime(runTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRunTime = &runTime
}

// setNextRunTime 设置下次执行时间（内部方法）
func (t *BaseTask) setNextRunTime(nextTime *time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextRunTime = nextTime
}

// Execute 执行任务
func (t *BaseTask) Execute(ctx context.Context) error {
	if t.taskFunc == nil {
		return fmt.Errorf("任务函数未设置")
	}

	t.setStatus(TaskStatusRunning)
	t.setLastRunTime(time.Now())

	err := t.taskFunc(ctx)
	if err != nil {
		t.setStatus(TaskStatusFailed)
		return fmt.Errorf("任务执行失败: %w", err)
	}

	t.setStatus(TaskStatusCompleted)
	return nil
}

// Cancel 取消任务
func (t *BaseTask) Cancel() error {
	t.setStatus(TaskStatusCancelled)
	return nil
}

// CronTask Cron定时任务
type CronTask struct {
	*BaseTask
	cronExpr string           // Cron表达式
	schedule cron.Schedule    // Cron调度器
	entryID  cron.EntryID     // Cron条目ID
}

// NewCronTask 创建新的Cron任务
func NewCronTask(id, name, cronExpr, funcName string, taskFunc TaskFunc) (*CronTask, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("无效的Cron表达式: %w", err)
	}

	now := time.Now()
	nextTime := schedule.Next(now)

	task := &CronTask{
		BaseTask: &BaseTask{
			id:          id,
			name:        name,
			taskType:    TaskTypeCron,
			status:      TaskStatusPending,
			funcName:    funcName,
			taskFunc:    taskFunc,
			createdAt:   now,
			nextRunTime: &nextTime,
		},
		cronExpr: cronExpr,
		schedule: schedule,
	}

	return task, nil
}

// GetCronExpr 获取Cron表达式
func (ct *CronTask) GetCronExpr() string {
	return ct.cronExpr
}

// UpdateNextRunTime 更新下次执行时间
func (ct *CronTask) UpdateNextRunTime() {
	now := time.Now()
	nextTime := ct.schedule.Next(now)
	ct.setNextRunTime(&nextTime)
}

// TimerTask 一次性定时任务
type TimerTask struct {
	*BaseTask
	executeAt time.Time // 执行时间
	timer     *time.Timer // 定时器
}

// NewTimerTask 创建新的定时任务
func NewTimerTask(id, name string, executeAt time.Time, funcName string, taskFunc TaskFunc) *TimerTask {
	now := time.Now()
	task := &TimerTask{
		BaseTask: &BaseTask{
			id:          id,
			name:        name,
			taskType:    TaskTypeTimer,
			status:      TaskStatusPending,
			funcName:    funcName,
			taskFunc:    taskFunc,
			createdAt:   now,
			nextRunTime: &executeAt,
		},
		executeAt: executeAt,
	}

	return task
}

// GetExecuteAt 获取执行时间
func (tt *TimerTask) GetExecuteAt() time.Time {
	return tt.executeAt
}

// Cancel 取消定时任务
func (tt *TimerTask) Cancel() error {
	if tt.timer != nil {
		tt.timer.Stop()
	}
	return tt.BaseTask.Cancel()
}

// SetTimer 设置定时器
func (tt *TimerTask) SetTimer(timer *time.Timer) {
	tt.timer = timer
}