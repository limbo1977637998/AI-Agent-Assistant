package orchestrator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeTask      MessageType = "task"
	MessageTypeResult    MessageType = "result"
	MessageTypeEvent     MessageType = "event"
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypeRequest   MessageType = "request"
	MessageTypeResponse  MessageType = "response"
)

// Message 消息定义
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType           `json:"type"`
	From      string                 `json:"from"`      // 发送者Agent名称
	To        string                 `json:"to"`        // 接收者Agent名称（空表示广播）
	Content   interface{}            `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageHandler 消息处理函数
type MessageHandler func(msg *Message) error

// CommunicationBus 通信总线
type CommunicationBus struct {
	mu               sync.RWMutex
	subscribers      map[string][]MessageHandler // agent_name -> handlers
	broadcastSubs    []MessageHandler            // 广播订阅者
	messageHistory   []*Message                  // 消息历史（用于调试）
	maxHistory       int
	eventChan         chan *Message
	stopped          chan struct{}
}

// NewCommunicationBus 创建通信总线
func NewCommunicationBus() *CommunicationBus {
	bus := &CommunicationBus{
		subscribers:    make(map[string][]MessageHandler),
		broadcastSubs: make([]MessageHandler, 0),
		messageHistory: make([]*Message, 0),
		maxHistory:     1000,
		eventChan:      make(chan *Message, 1000),
		stopped:        make(chan struct{}),
	}

	// 启动事件处理协程
	go bus.processEvents()

	return bus
}

// Subscribe 订阅消息
func (b *CommunicationBus) Subscribe(agentName string, handler MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[agentName] = append(b.subscribers[agentName], handler)
}

// SubscribeBroadcast 订阅广播消息
func (b *CommunicationBus) SubscribeBroadcast(handler MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.broadcastSubs = append(b.broadcastSubs, handler)
}

// Unsubscribe 取消订阅
func (b *CommunicationBus) Unsubscribe(agentName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.subscribers, agentName)
}

// Send 发送消息给指定Agent
func (b *CommunicationBus) Send(msg *Message) error {
	if msg.To == "" {
		return fmt.Errorf("message 'to' field is required for direct messages")
	}

	msg.ID = generateMessageID()
	msg.Timestamp = time.Now()

	// 添加到历史记录
	b.addToHistory(msg)

	// 发送到事件通道
	select {
	case b.eventChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message")
	}
}

// Broadcast 广播消息
func (b *CommunicationBus) Broadcast(msg *Message) error {
	msg.ID = generateMessageID()
	msg.Timestamp = time.Now()
	msg.To = "" // 广播消息没有接收者

	// 添加到历史记录
	b.addToHistory(msg)

	// 发送到事件通道
	select {
	case b.eventChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout broadcasting message")
	}
}

// processEvents 处理事件
func (b *CommunicationBus) processEvents() {
	for {
		select {
		case msg := <-b.eventChan:
			b.handleMessage(msg)
		case <-b.stopped:
			return
		}
	}
}

// handleMessage 处理消息
func (b *CommunicationBus) handleMessage(msg *Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 如果是广播消息，通知所有广播订阅者
	if msg.To == "" {
		for _, handler := range b.broadcastSubs {
			go func(h MessageHandler, m *Message) {
				_ = h(m)
			}(handler, msg)
		}
		return
	}

	// 发送给指定Agent的订阅者
	if handlers, exists := b.subscribers[msg.To]; exists {
		for _, handler := range handlers {
			go func(h MessageHandler, m *Message) {
				_ = h(m)
			}(handler, msg)
		}
	}
}

// addToHistory 添加到消息历史
func (b *CommunicationBus) addToHistory(msg *Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messageHistory = append(b.messageHistory, msg)
	if len(b.messageHistory) > b.maxHistory {
		b.messageHistory = b.messageHistory[1:]
	}
}

// GetHistory 获取消息历史
func (b *CommunicationBus) GetHistory(limit int) []*Message {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit <= 0 || limit > len(b.messageHistory) {
		limit = len(b.messageHistory)
	}

	start := len(b.messageHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*Message, limit)
	copy(result, b.messageHistory[start:])
	return result
}

// GetMessagesForAgent 获取特定Agent的消息
func (b *CommunicationBus) GetMessagesForAgent(agentName string, limit int) []*Message {
	b.mu.RLock()
	defer b.mu.RUnlock()

	messages := make([]*Message, 0)
	for i := len(b.messageHistory) - 1; i >= 0; i-- {
		msg := b.messageHistory[i]
		if msg.To == agentName || msg.From == agentName || msg.To == "" {
			messages = append(messages, msg)
			if len(messages) >= limit {
				break
			}
		}
	}

	return messages
}

// Stop 停止通信总线
func (b *CommunicationBus) Stop() {
	close(b.stopped)
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

// Event 事件定义（用于事件驱动）
type Event struct {
	Name      string                 `json:"name"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventBus 事件总线
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler // event_name -> handlers
	eventChan   chan *Event
	stopped     chan struct{}
}

// EventHandler 事件处理函数
type EventHandler func(event *Event) error

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	bus := &EventBus{
		subscribers: make(map[string][]EventHandler),
		eventChan:   make(chan *Event, 1000),
		stopped:     make(chan struct{}),
	}

	go bus.processEvents()

	return bus
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
}

// Publish 发布事件
func (b *EventBus) Publish(event *Event) error {
	event.Timestamp = time.Now()

	select {
	case b.eventChan <- event:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout publishing event")
	}
}

// processEvents 处理事件
func (b *EventBus) processEvents() {
	for {
		select {
		case event := <-b.eventChan:
			b.handleEvent(event)
		case <-b.stopped:
			return
		}
	}
}

// handleEvent 处理事件
func (b *EventBus) handleEvent(event *Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, exists := b.subscribers[event.Name]; exists {
		for _, handler := range handlers {
			go func(h EventHandler, e *Event) {
				_ = h(e)
			}(handler, event)
		}
	}
}

// Stop 停止事件总线
func (b *EventBus) Stop() {
	close(b.stopped)
}

// Helper functions for creating messages

// NewTaskMessage 创建任务消息
func NewTaskMessage(from, to string, task *Task) *Message {
	return &Message{
		Type:    MessageTypeTask,
		From:    from,
		To:      to,
		Content: task,
	}
}

// NewResultMessage 创建结果消息
func NewResultMessage(from, to string, result interface{}) *Message {
	return &Message{
		Type:    MessageTypeResult,
		From:    from,
		To:      to,
		Content: result,
	}
}

// NewEventMessage 创建事件消息
func NewEventMessage(from string, event *Event) *Message {
	return &Message{
		Type:    MessageTypeEvent,
		From:    from,
		Content: event,
	}
}

// Marshal 序列化消息
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal 反序列化消息
func UnmarshalMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
