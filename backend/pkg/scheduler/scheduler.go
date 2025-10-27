package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler 任务调度器实现
type TaskScheduler struct {
	cronScheduler  *cron.Cron            // Cron调度器
	timerTasks     map[string]*TimerTask // 定时任务映射
	tasks          map[string]Task       // 所有任务映射
	running        bool                  // 运行状态
	mu             sync.RWMutex          // 读写锁
	ctx            context.Context       // 上下文
	cancel         context.CancelFunc    // 取消函数
	taskRepository TaskRepository        // 任务存储
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler(taskRepository TaskRepository) *TaskScheduler {
	return &TaskScheduler{
		cronScheduler:  cron.New(cron.WithSeconds()),
		timerTasks:     make(map[string]*TimerTask),
		tasks:          make(map[string]Task),
		running:        false,
		taskRepository: taskRepository,
	}
}

// Start 启动调度器
func (s *TaskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("调度器已经在运行中")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	// 启动Cron调度器
	s.cronScheduler.Start()

	// 启动定时任务监控
	go s.monitorTimerTasks()

	return nil
}

// Stop 停止调度器
func (s *TaskScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("调度器未运行")
	}

	// 停止Cron调度器
	s.cronScheduler.Stop()

	// 取消所有定时任务
	for _, task := range s.timerTasks {
		task.Cancel()
	}

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	return nil
}

// AddTask 添加任务
func (s *TaskScheduler) AddTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID := task.GetID()
	if _, exists := s.tasks[taskID]; exists {
		return fmt.Errorf("任务ID %s 已存在", taskID)
	}

	// 根据任务类型添加到相应的调度器
	switch t := task.(type) {
	case *CronTask:
		err := s.addCronTask(t)
		if err != nil {
			return err
		}
	case *TimerTask:
		err := s.addTimerTask(t)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的任务类型")
	}

	s.tasks[taskID] = task

	// 保存任务到存储
	if s.taskRepository != nil {
		if err := s.taskRepository.SaveTask(s.ctx, task); err != nil {
			return fmt.Errorf("保存任务失败: %w", err)
		}
	}

	return nil
}

// RemoveTask 移除任务
func (s *TaskScheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务ID %s 不存在", taskID)
	}

	// 根据任务类型从相应的调度器中移除
	switch t := task.(type) {
	case *CronTask:
		s.cronScheduler.Remove(t.entryID)
	case *TimerTask:
		t.Cancel()
		delete(s.timerTasks, taskID)
	}

	delete(s.tasks, taskID)

	// 从存储中删除任务
	if s.taskRepository != nil {
		if err := s.taskRepository.DeleteTask(taskID); err != nil {
			return fmt.Errorf("删除任务失败: %w", err)
		}
	}

	return nil
}

// GetTask 获取任务
func (s *TaskScheduler) GetTask(taskID string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务ID %s 不存在", taskID)
	}

	return task, nil
}

// ListTasks 列出所有任务
func (s *TaskScheduler) ListTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// IsRunning 检查调度器是否运行中
func (s *TaskScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// addCronTask 添加Cron任务
func (s *TaskScheduler) addCronTask(task *CronTask) error {
	entryID, err := s.cronScheduler.AddFunc(task.GetCronExpr(), func() {
		s.executeCronTask(task)
	})
	if err != nil {
		return fmt.Errorf("添加Cron任务失败: %w", err)
	}

	task.entryID = entryID
	return nil
}

// addTimerTask 添加定时任务
func (s *TaskScheduler) addTimerTask(task *TimerTask) error {
	now := time.Now()
	executeAt := task.GetExecuteAt()

	if executeAt.Before(now) {
		return fmt.Errorf("执行时间不能早于当前时间")
	}

	duration := executeAt.Sub(now)
	timer := time.NewTimer(duration)
	task.SetTimer(timer)

	s.timerTasks[task.GetID()] = task

	return nil
}

// executeCronTask 执行Cron任务
func (s *TaskScheduler) executeCronTask(task *CronTask) {
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Minute) // 60分钟超时
		defer cancel()

		err := task.Execute(ctx)
		if err != nil {
			// 记录错误日志
			fmt.Printf("Cron任务执行失败 [%s]: %v\n", task.GetName(), err)
		}

		// 更新下次执行时间
		task.UpdateNextRunTime()

		// 更新任务状态到存储
		if s.taskRepository != nil {
			s.taskRepository.UpdateTaskStatus(task.GetID(), task.GetStatus())
		}
	}()
}

// monitorTimerTasks 监控定时任务
func (s *TaskScheduler) monitorTimerTasks() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			s.checkTimerTasks()
			time.Sleep(1 * time.Second) // 每秒检查一次
		}
	}
}

// checkTimerTasks 检查定时任务
func (s *TaskScheduler) checkTimerTasks() {
	s.mu.RLock()
	tasks := make([]*TimerTask, 0, len(s.timerTasks))
	for _, task := range s.timerTasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		select {
		case <-task.timer.C:
			s.executeTimerTask(task)
		default:
			// 任务还未到执行时间
		}
	}
}

// executeTimerTask 执行定时任务
func (s *TaskScheduler) executeTimerTask(task *TimerTask) {
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Minute) // 60分钟超时
		defer cancel()

		err := task.Execute(ctx)
		if err != nil {
			// 记录错误日志
			fmt.Printf("定时任务执行失败 [%s]: %v\n", task.GetName(), err)
		}

		// 更新任务状态到存储
		if s.taskRepository != nil {
			s.taskRepository.UpdateTaskStatus(task.GetID(), task.GetStatus())
		}

		// 移除已完成的定时任务
		s.mu.Lock()
		delete(s.timerTasks, task.GetID())
		delete(s.tasks, task.GetID())
		s.mu.Unlock()
	}()
}
