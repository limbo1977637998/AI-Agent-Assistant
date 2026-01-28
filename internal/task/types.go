package task

import (
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityNormal TaskPriority = 1
	PriorityHigh   TaskPriority = 2
	PriorityUrgent TaskPriority = 3
)

// Task 任务定义
type Task struct {
	ID                   string                 `json:"id"`
	Type                 string                 `json:"type"`        // single, composite, workflow
	Goal                 string                 `json:"goal"`        // 任务目标
	Requirements         map[string]interface{} `json:"requirements"` // 任务要求
	Priority             TaskPriority           `json:"priority"`    // 优先级
	Status               TaskStatus             `json:"status"`      // 状态
	AssignedTo           string                 `json:"assigned_to"` // 分配给的Agent
	RequiredCapabilities []string               `json:"required_capabilities"` // 所需能力
	DependsOn            []string               `json:"depends_on"`  // 依赖的任务ID
	CreatedAt            time.Time              `json:"created_at"`
	StartedAt            *time.Time             `json:"started_at,omitempty"`
	CompletedAt          *time.Time             `json:"completed_at,omitempty"`
	Result               interface{}            `json:"result,omitempty"`
	Error                string                 `json:"error,omitempty"`
	RetryCount           int                    `json:"retry_count"`
	MaxRetries           int                    `json:"max_retries"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// EventBus 事件总线
type EventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
}

// EventHandler 事件处理器
type EventHandler func(event *TaskEvent) error

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// Publish 发布事件
func (b *EventBus) Publish(event *TaskEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, ok := b.subscribers[event.Name]; ok {
		for _, handler := range handlers {
			go func(h EventHandler) {
				_ = h(event)
			}(handler)
		}
	}
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handler == nil {
		// 取消订阅
		delete(b.subscribers, eventName)
		return
	}

	if _, exists := b.subscribers[eventName]; !exists {
		b.subscribers[eventName] = make([]EventHandler, 0)
	}
	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
}
