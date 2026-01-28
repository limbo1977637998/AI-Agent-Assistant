package orchestrator

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 0
	TaskPriorityNormal TaskPriority = 1
	TaskPriorityHigh   TaskPriority = 2
	TaskPriorityUrgent TaskPriority = 3
)

// Task 任务定义
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // single, composite, workflow
	Goal        string                 `json:"goal"`        // 任务目标
	Requirements map[string]interface{} `json:"requirements"` // 任务要求
	Priority    TaskPriority           `json:"priority"`    // 优先级
	Status      TaskStatus             `json:"status"`      // 状态
	AssignedTo  string                 `json:"assigned_to"` // 分配给的Agent
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TaskQueue 任务队列（优先队列）
type TaskQueue struct {
	items []*Task
	mu    sync.Mutex
}

// NewTaskQueue 创建任务队列
func NewTaskQueue() *TaskQueue {
	queue := &TaskQueue{
		items: make([]*Task, 0),
	}
	heap.Init(queue)
	return queue
}

// Len 实现heap.Interface
func (q *TaskQueue) Len() int {
	return len(q.items)
}

// Less 实现heap.Interface
func (q *TaskQueue) Less(i, j int) bool {
	// 优先级高的排在前面
	return q.items[i].Priority > q.items[j].Priority
}

// Swap 实现heap.Interface
func (q *TaskQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

// Push 实现heap.Interface
func (q *TaskQueue) Push(x interface{}) {
	q.items = append(q.items, x.(*Task))
}

// Pop 实现heap.Interface
func (q *TaskQueue) Pop() interface{} {
	old := q.items
	n := len(old)
	item := old[n-1]
	q.items = old[0 : n-1]
	return item
}

// Enqueue 入队
func (q *TaskQueue) Enqueue(task *Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(q, task)
}

// Dequeue 出队
func (q *TaskQueue) Dequeue() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Len() == 0 {
		return nil
	}
	return heap.Pop(q).(*Task)
}

// Peek 查看队首任务
func (q *TaskQueue) Peek() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Len() == 0 {
		return nil
	}
	return q.items[0]
}

// Size 队列大小
func (q *TaskQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.Len()
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	registry      *AgentRegistry
	taskQueue     *TaskQueue
	runningTasks  map[string]*Task // task_id -> task
	mu            sync.RWMutex
	stopCh        chan struct{}
	workerStopped chan struct{}
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(registry *AgentRegistry) *TaskScheduler {
	return &TaskScheduler{
		registry:      registry,
		taskQueue:     NewTaskQueue(),
		runningTasks:  make(map[string]*Task),
		stopCh:        make(chan struct{}),
		workerStopped: make(chan struct{}),
	}
}

// Start 启动调度器
func (s *TaskScheduler) Start() {
	go s.worker()
}

// Stop 停止调度器
func (s *TaskScheduler) Stop() {
	close(s.stopCh)
	<-s.workerStopped
}

// Submit 提交任务
func (s *TaskScheduler) Submit(task *Task) error {
	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	s.taskQueue.Enqueue(task)
	return nil
}

// GetTask 获取任务信息
func (s *TaskScheduler) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 先在运行任务中查找
	if task, exists := s.runningTasks[taskID]; exists {
		return task, nil
	}

	// 在队列中查找（需要遍历队列）
	// 注意：这里简化处理，实际应用中可能需要维护一个所有任务的映射

	return nil, fmt.Errorf("task %s not found in running tasks", taskID)
}

// Cancel 取消任务
func (s *TaskScheduler) Cancel(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.runningTasks[taskID]; exists {
		if task.Status == TaskStatusRunning {
			task.Status = TaskStatusCancelled
			return nil
		}
		return fmt.Errorf("task %s cannot be cancelled in current state", taskID)
	}

	return fmt.Errorf("task %s not found in running tasks", taskID)
}

// worker 调度工作协程
func (s *TaskScheduler) worker() {
	defer close(s.workerStopped)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.scheduleTasks()
		}
	}
}

// scheduleTasks 调度任务
func (s *TaskScheduler) scheduleTasks() {
	// 从队列中取出任务
	for {
		task := s.taskQueue.Dequeue()
		if task == nil {
			break
		}

		// 分配任务给Agent
		if err := s.assignTask(task); err != nil {
			// 分配失败，重新入队
			task.RetryCount++
			if task.RetryCount < task.MaxRetries {
				s.taskQueue.Enqueue(task)
			} else {
				task.Status = TaskStatusFailed
				task.Error = fmt.Sprintf("Failed to assign after %d retries: %v", task.RetryCount, err)
			}
		}
	}
}

// assignTask 分配任务给Agent
func (s *TaskScheduler) assignTask(task *Task) error {
	// 查找合适的Agent
	var agent *AgentInfo
	var err error

	if task.AssignedTo != "" {
		// 指定了Agent
		agent, err = s.registry.Get(task.AssignedTo)
		if err != nil {
			return err
		}
		if agent.Status != "active" {
			return fmt.Errorf("agent %s is not active", agent.Name)
		}
	} else {
		// 自动选择Agent
		agent, err = s.registry.GetIdleAgent()
		if err != nil {
			return err
		}
	}

	// 分配任务
	s.mu.Lock()
	task.Status = TaskStatusAssigned
	task.AssignedTo = agent.Name
	s.runningTasks[task.ID] = task
	s.mu.Unlock()

	// 更新Agent状态
	s.registry.UpdateStatus(agent.Name, "busy")

	// 执行任务（这里只是标记，实际执行在其他地方）
	// TODO: 触发Agent执行任务

	return nil
}

// CompleteTask 完成任务
func (s *TaskScheduler) CompleteTask(taskID string, result interface{}, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.runningTasks[taskID]
	if !exists {
		return
	}

	now := time.Now()
	task.CompletedAt = &now

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = TaskStatusCompleted
		task.Result = result
	}

	// 释放Agent
	if task.AssignedTo != "" {
		s.registry.UpdateStatus(task.AssignedTo, "active")
	}

	// 从运行任务中移除
	delete(s.runningTasks, taskID)
}

// GetQueueSize 获取队列大小
func (s *TaskScheduler) GetQueueSize() int {
	return s.taskQueue.Size()
}

// GetRunningTasks 获取运行中的任务
func (s *TaskScheduler) GetRunningTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.runningTasks))
	for _, task := range s.runningTasks {
		tasks = append(tasks, task)
	}
	return tasks
}
