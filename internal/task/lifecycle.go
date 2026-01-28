package task

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecycleManager 任务生命周期管理器
type LifecycleManager struct {
	mu         sync.RWMutex
	taskStates map[string]*TaskState // task_id -> state
	history    map[string][]*TaskStateEntry // task_id -> history
	eventBus   *EventBus
}

// TaskState 任务状态
type TaskState struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Stage       string                 `json:"stage"`       // 当前阶段
	Progress    float64                `json:"progress"`    // 进度 0-1
	Input       interface{}            `json:"input"`
	Output      interface{}            `json:"output"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TaskStateEntry 任务状态历史记录
type TaskStateEntry struct {
	State      *TaskState `json:"state"`
	Timestamp  time.Time  `json:"timestamp"`
	ChangedBy  string    `json:"changed_by"`
	Reason     string    `json:"reason"`
}

// TaskEvent 任务事件
type TaskEvent struct {
	Name      string                 `json:"name"`
	TaskID    string                 `json:"task_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewLifecycleManager 创建生命周期管理器
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		taskStates: make(map[string]*TaskState),
		history:    make(map[string][]*TaskStateEntry),
		eventBus:   NewEventBus(),
	}
}

// Create 创建新任务状态
func (m *LifecycleManager) Create(task *Task) (*TaskState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := &TaskState{
		TaskID:    task.ID,
		Status:    task.Status,
		Stage:     "created",
		Progress:  0,
		Input:     task.Requirements,
		Metadata:  task.Metadata,
		UpdatedAt: time.Now(),
	}

	m.taskStates[task.ID] = state

	// 记录历史
	m.recordHistory(state, "system", "task created")

	// 发布事件
	m.publishEvent(&TaskEvent{
		Name:      "task.created",
		TaskID:    task.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"goal":     task.Goal,
			"priority": task.Priority,
		},
	})

	return state, nil
}

// Get 获取任务状态
func (m *LifecycleManager) Get(taskID string) (*TaskState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.taskStates[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return state, nil
}

// UpdateStatus 更新任务状态
func (m *LifecycleManager) UpdateStatus(taskID string, status TaskStatus, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.taskStates[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	oldStatus := state.Status
	state.Status = status
	state.UpdatedAt = time.Now()

	// 根据状态设置时间戳
	switch status {
	case TaskStatusRunning:
		if state.StartedAt == nil {
			now := time.Now()
			state.StartedAt = &now
		}
		state.Stage = "running"
	case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		if state.CompletedAt == nil {
			now := time.Now()
			state.CompletedAt = &now
			if state.StartedAt != nil {
				state.Progress = 1.0
			}
		}
		if status == TaskStatusCompleted {
			state.Stage = "completed"
		} else if status == TaskStatusFailed {
			state.Stage = "failed"
		} else {
			state.Stage = "cancelled"
		}
	}

	// 记录历史
	m.recordHistory(state, "system", fmt.Sprintf("status changed: %s -> %s", oldStatus, status))

	// 发布事件
	eventName := fmt.Sprintf("task.%s", status)
	m.publishEvent(&TaskEvent{
		Name:      eventName,
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": status,
			"reason":     reason,
		},
	})

	return nil
}

// UpdateProgress 更新任务进度
func (m *LifecycleManager) UpdateProgress(taskID string, progress float64, stage string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.taskStates[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	state.Progress = progress
	state.Stage = stage
	state.UpdatedAt = time.Now()

	// 记录历史
	m.recordHistory(state, "system", fmt.Sprintf("progress updated: %.2f%%", progress*100))

	// 发布事件
	m.publishEvent(&TaskEvent{
		Name:      "task.progress",
		TaskID:    taskID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"progress": progress,
			"stage":    stage,
		},
	})

	return nil
}

// SetOutput 设置任务输出
func (m *LifecycleManager) SetOutput(taskID string, output interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.taskStates[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	state.Output = output
	state.UpdatedAt = time.Now()

	return nil
}

// SetError 设置任务错误
func (m *LifecycleManager) SetError(taskID string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.taskStates[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	state.Error = errorMsg
	state.UpdatedAt = time.Now()

	// 记录历史
	m.recordHistory(state, "system", fmt.Sprintf("error set: %s", errorMsg))

	return nil
}

// GetHistory 获取任务状态历史
func (m *LifecycleManager) GetHistory(taskID string) ([]*TaskStateEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.history[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return history, nil
}

// ListByStatus 按状态列出任务
func (m *LifecycleManager) ListByStatus(status TaskStatus) []*TaskState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make([]*TaskState, 0)
	for _, state := range m.taskStates {
		if state.Status == status {
			states = append(states, state)
		}
	}

	return states
}

// GetAll 获取所有任务状态
func (m *LifecycleManager) GetAll() map[string]*TaskState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	result := make(map[string]*TaskState)
	for k, v := range m.taskStates {
		result[k] = v
	}

	return result
}

// Delete 删除任务状态
func (m *LifecycleManager) Delete(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.taskStates[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(m.taskStates, taskID)
	delete(m.history, taskID)

	return nil
}

// recordHistory 记录历史
func (m *LifecycleManager) recordHistory(state *TaskState, changedBy, reason string) {
	entry := &TaskStateEntry{
		State:     state,
		Timestamp: time.Now(),
		ChangedBy:  changedBy,
		Reason:    reason,
	}

	if _, exists := m.history[state.TaskID]; !exists {
		m.history[state.TaskID] = make([]*TaskStateEntry, 0)
	}

	m.history[state.TaskID] = append(m.history[state.TaskID], entry)
}

// publishEvent 发布事件
func (m *LifecycleManager) publishEvent(event *TaskEvent) {
	m.eventBus.Publish(event)
}

// Subscribe 订阅任务事件
func (m *LifecycleManager) Subscribe(eventName string, handler EventHandler) {
	m.eventBus.Subscribe(eventName, handler)
}

// GetTaskCount 获取各状态任务数量统计
func (m *LifecycleManager) GetTaskCount() map[TaskStatus]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := make(map[TaskStatus]int)
	for _, state := range m.taskStates {
		count[state.Status]++
	}

	return count
}

// GetRunningTasks 获取运行中的任务
func (m *LifecycleManager) GetRunningTasks() []*TaskState {
	return m.ListByStatus(TaskStatusRunning)
}

// GetPendingTasks 获取待处理的任务
func (m *LifecycleManager) GetPendingTasks() []*TaskState {
	return m.ListByStatus(TaskStatusPending)
}

// MonitorTask 监控任务（阻塞直到任务完成）
func (m *LifecycleManager) MonitorTask(ctx context.Context, taskID string) (*TaskState, error) {
	ch := make(chan *TaskState, 1)

	// 订阅任务完成事件
	handler := func(event *TaskEvent) error {
		if event.TaskID == taskID {
			state, err := m.Get(taskID)
			if err == nil {
				ch <- state
			}
		}
		return nil
	}

	m.Subscribe("task.completed", handler)
	m.Subscribe("task.failed", handler)
	defer m.Subscribe("task.completed", nil)
	defer m.Subscribe("task.failed", nil)

	// 返回当前状态
	currentState, err := m.Get(taskID)
	if err != nil {
		return nil, err
	}

	// 如果任务已完成，直接返回
	if currentState.Status == TaskStatusCompleted || currentState.Status == TaskStatusFailed {
		return currentState, nil
	}

	// 等待任务完成
	select {
	case state := <-ch:
		return state, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Cleanup 清理已完成的旧任务
func (m *LifecycleManager) Cleanup(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-olderThan)

	for taskID, state := range m.taskStates {
		if (state.Status == TaskStatusCompleted || state.Status == TaskStatusFailed) &&
			state.CompletedAt != nil && state.CompletedAt.Before(cutoff) {

			delete(m.taskStates, taskID)
			delete(m.history, taskID)
			count++
		}
	}

	return count
}
