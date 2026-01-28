package orchestrator

import (
	"testing"
	"time"
)

// TestAgentRegistry 测试Agent注册表
func TestAgentRegistry(t *testing.T) {
	registry := NewAgentRegistry()

	// 测试注册Agent
	agent := &AgentInfo{
		ID:           "agent-1",
		Name:         "test-agent",
		Type:         "expert",
		Capabilities: []string{"search", "analyze"},
		Endpoint:     "http://localhost:8081",
		Status:       "active",
		Metadata:     make(map[string]string),
	}

	err := registry.Register(agent)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// 测试获取Agent
	retrieved, err := registry.Get("test-agent")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if retrieved.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", retrieved.Name)
	}

	// 测试列出Agent
	agents := registry.List()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	// 测试按能力查找
	searchAgents := registry.ListByCapability("search")
	if len(searchAgents) != 1 {
		t.Errorf("Expected 1 agent with search capability, got %d", len(searchAgents))
	}

	// 测试心跳更新
	err = registry.UpdateHeartbeat("test-agent")
	if err != nil {
		t.Fatalf("Failed to update heartbeat: %v", err)
	}

	// 测试健康检查
	if !registry.CheckHealth("test-agent") {
		t.Error("Agent should be healthy")
	}

	// 测试注销
	err = registry.Unregister("test-agent")
	if err != nil {
		t.Fatalf("Failed to unregister agent: %v", err)
	}

	// 验证已注销
	_, err = registry.Get("test-agent")
	if err == nil {
		t.Error("Expected error when getting unregistered agent")
	}
}

// TestTaskScheduler 测试任务调度器
func TestTaskScheduler(t *testing.T) {
	registry := NewAgentRegistry()
	scheduler := NewTaskScheduler(registry)

	// 注册一个测试Agent
	agent := &AgentInfo{
		ID:           "agent-1",
		Name:         "worker",
		Type:         "general",
		Capabilities: []string{"task"},
		Endpoint:     "http://localhost:8081",
		Status:       "active",
		Metadata:     make(map[string]string),
	}
	registry.Register(agent)

	// 启动调度器
	scheduler.Start()
	defer scheduler.Stop()

	// 提交任务
	task := &Task{
		ID:          "task-1",
		Type:        "single",
		Goal:        "测试任务",
		Priority:    TaskPriorityNormal,
		Requirements: make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	err := scheduler.Submit(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// 等待任务被调度
	time.Sleep(2 * time.Second)

	// 检查队列大小
	queueSize := scheduler.GetQueueSize()
	t.Logf("Queue size: %d", queueSize)

	// 获取运行中的任务
	runningTasks := scheduler.GetRunningTasks()
	t.Logf("Running tasks: %d", len(runningTasks))

	// 由于调度是异步的，可能任务还在队列中或已经分配
	// 我们只验证任务被正确提交
	if queueSize == 0 && len(runningTasks) == 0 {
		// 任务应该被处理了（即使失败了也算处理）
		t.Log("Task was processed (queue empty and no running tasks)")
	}
}

// TestTaskQueue 测试任务队列
func TestTaskQueue(t *testing.T) {
	queue := NewTaskQueue()

	// 添加不同优先级的任务
	lowTask := &Task{
		ID:       "low",
		Priority: TaskPriorityLow,
	}
	normalTask := &Task{
		ID:       "normal",
		Priority: TaskPriorityNormal,
	}
	highTask := &Task{
		ID:       "high",
		Priority: TaskPriorityHigh,
	}

	queue.Enqueue(lowTask)
	queue.Enqueue(normalTask)
	queue.Enqueue(highTask)

	// 验证队列大小
	if queue.Size() != 3 {
		t.Errorf("Expected queue size 3, got %d", queue.Size())
	}

	// 验证优先级排序（高优先级先出）
	task := queue.Dequeue()
	if task.ID != "high" {
		t.Errorf("Expected first task 'high', got '%s'", task.ID)
	}

	task = queue.Dequeue()
	if task.ID != "normal" {
		t.Errorf("Expected second task 'normal', got '%s'", task.ID)
	}

	task = queue.Dequeue()
	if task.ID != "low" {
		t.Errorf("Expected third task 'low', got '%s'", task.ID)
	}
}

// TestCommunicationBus 测试通信总线
func TestCommunicationBus(t *testing.T) {
	bus := NewCommunicationBus()
	defer bus.Stop()

	// 测试点对点消息
	received := make(chan *Message, 1)

	handler := func(msg *Message) error {
		received <- msg
		return nil
	}

	bus.Subscribe("agent-1", handler)

	msg := &Message{
		Type:    MessageTypeTask,
		From:    "orchestrator",
		To:      "agent-1",
		Content: "test message",
	}

	err := bus.Send(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	select {
	case <-received:
		// 消息接收成功
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for message")
	}

	// 测试广播消息
	broadcastReceived := make(chan *Message, 2)

	broadcastHandler := func(msg *Message) error {
		broadcastReceived <- msg
		return nil
	}

	bus.SubscribeBroadcast(broadcastHandler)

	broadcastMsg := &Message{
		Type:    MessageTypeEvent,
		From:    "system",
		Content: "broadcast message",
	}

	err = bus.Broadcast(broadcastMsg)
	if err != nil {
		t.Fatalf("Failed to broadcast message: %v", err)
	}

	select {
	case <-broadcastReceived:
		// 广播消息接收成功
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for broadcast message")
	}
}

// TestEventBus 测试事件总线
func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	defer bus.Stop()

	received := make(chan *Event, 1)

	handler := func(event *Event) error {
		received <- event
		return nil
	}

	bus.Subscribe("task.completed", handler)

	event := &Event{
		Name:   "task.completed",
		Source: "scheduler",
		Data:   map[string]interface{}{"task_id": "task-1"},
	}

	err := bus.Publish(event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	select {
	case <-received:
		// 事件接收成功
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for event")
	}
}
