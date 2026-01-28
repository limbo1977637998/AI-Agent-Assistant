package main

import (
	"fmt"
	"log"
	"time"

	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	aiagentworkflow "ai-agent-assistant/internal/workflow"
)

func main() {
	fmt.Println("\n🚀 AI Agent Assistant v0.5 - 基础框架演示")
	fmt.Println("=========================================\n")

	// 1. 演示Agent注册表
	fmt.Println("📋 1. Agent注册表演示")
	demonstrateAgentRegistry()
	fmt.Println()

	// 2. 演示任务调度器
	fmt.Println("📋 2. 任务调度器演示")
	demonstrateTaskScheduler()
	fmt.Println()

	// 3. 演示通信总线
	fmt.Println("📋 3. 通信总线演示")
	demonstrateCommunication()
	fmt.Println()

	// 4. 演示工作流引擎
	fmt.Println("📋 4. 工作流引擎演示")
	demonstrateWorkflow()
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("✅ 所有演示完成！")
}

// demonstrateAgentRegistry 演示Agent注册表
func demonstrateAgentRegistry() {
	registry := aiagentorchestrator.NewAgentRegistry()

	// 注册多个Agent
	agents := []*aiagentorchestrator.AgentInfo{
		{
			ID:           "agent-1",
			Name:         "researcher",
			Type:         "expert",
			Capabilities: []string{"search", "analyze"},
			Endpoint:     "http://localhost:8081",
			Metadata:     map[string]string{"role": "研究专家"},
		},
		{
			ID:           "agent-2",
			Name:         "analyst",
			Type:         "expert",
			Capabilities: []string{"analyze", "report"},
			Endpoint:     "http://localhost:8082",
			Metadata:     map[string]string{"role": "分析专家"},
		},
		{
			ID:           "agent-3",
			Name:         "writer",
			Type:         "expert",
			Capabilities: []string{"write", "summarize"},
			Endpoint:     "http://localhost:8083",
			Metadata:     map[string]string{"role": "写作专家"},
		},
	}

	// 注册所有Agent
	for _, agent := range agents {
		if err := registry.Register(agent); err != nil {
			log.Printf("Failed to register agent %s: %v", agent.Name, err)
		} else {
			fmt.Printf("  ✅ 注册Agent: %s (能力: %v)\n", agent.Name, agent.Capabilities)
		}
	}

	// 查询Agent
	fmt.Printf("\n  📊 Agent总数: %d\n", registry.Count())
	fmt.Printf("  📊 活跃Agent: %d\n", registry.CountByStatus("active"))

	// 按能力查询
	searchAgents := registry.ListByCapability("search")
	fmt.Printf("  🔍 具备search能力的Agent: %d个\n", len(searchAgents))
	for _, agent := range searchAgents {
		fmt.Printf("     - %s\n", agent.Name)
	}

	// 智能选择Agent
	bestAgent, err := registry.FindBestAgent([]string{"analyze", "report"})
	if err != nil {
		log.Printf("  ❌ 找不到匹配的Agent: %v", err)
	} else {
		fmt.Printf("  🎯 最匹配analyze/report能力的Agent: %s\n", bestAgent.Name)
	}
}

// demonstrateTaskScheduler 演示任务调度器
func demonstrateTaskScheduler() {
	registry := aiagentorchestrator.NewAgentRegistry()
	scheduler := aiagentorchestrator.NewTaskScheduler(registry)

	// 注册一个Agent
	agent := &aiagentorchestrator.AgentInfo{
		ID:           "worker-1",
		Name:         "worker",
		Type:         "general",
		Capabilities: []string{"task"},
		Endpoint:     "http://localhost:8081",
		Status:       "active",
		Metadata:     map[string]string{},
	}
	registry.Register(agent)

	// 启动调度器
	scheduler.Start()
	defer scheduler.Stop()

	// 提交不同优先级的任务
	tasks := []*aiagentorchestrator.Task{
		{
			ID:          "task-1",
			Type:        "single",
			Goal:        "低优先级任务",
			Priority:    aiagentorchestrator.TaskPriorityLow,
			Requirements: map[string]interface{}{},
			Metadata:    map[string]interface{}{},
		},
		{
			ID:          "task-2",
			Type:        "single",
			Goal:        "高优先级任务",
			Priority:    aiagentorchestrator.TaskPriorityHigh,
			Requirements: map[string]interface{}{},
			Metadata:    map[string]interface{}{},
		},
		{
			ID:          "task-3",
			Type:        "single",
			Goal:        "普通优先级任务",
			Priority:    aiagentorchestrator.TaskPriorityNormal,
			Requirements: map[string]interface{}{},
			Metadata:    map[string]interface{}{},
		},
	}

	// 提交任务
	for _, task := range tasks {
		if err := scheduler.Submit(task); err != nil {
			log.Printf("  ❌ 提交任务失败: %v", err)
		} else {
			fmt.Printf("  ✅ 提交任务: %s (优先级: %d)\n", task.ID, task.Priority)
		}
	}

	// 等待调度
	time.Sleep(2 * time.Second)

	// 查看队列状态
	fmt.Printf("\n  📊 队列大小: %d\n", scheduler.GetQueueSize())
	fmt.Printf("  📊 运行中任务: %d\n", len(scheduler.GetRunningTasks()))
}

// demonstrateCommunication 演示通信总线
func demonstrateCommunication() {
	bus := aiagentorchestrator.NewCommunicationBus()
	defer bus.Stop()

	// 订阅消息
	receivedCount := 0
	handler := func(msg *aiagentorchestrator.Message) error {
		receivedCount++
		fmt.Printf("  📨 [%s] 收到消息 from %s: %v\n", msg.Type, msg.From, msg.Content)
		return nil
	}

	bus.Subscribe("agent-1", handler)
	bus.SubscribeBroadcast(handler)

	// 发送点对点消息
	msg1 := &aiagentorchestrator.Message{
		Type:    aiagentorchestrator.MessageTypeTask,
		From:    "orchestrator",
		To:      "agent-1",
		Content: "执行任务A",
	}
	if err := bus.Send(msg1); err != nil {
		log.Printf("  ❌ 发送消息失败: %v", err)
	} else {
		fmt.Printf("  ✅ 发送点对点消息: orchestrator -> agent-1\n")
	}

	// 发送广播消息
	msg2 := &aiagentorchestrator.Message{
		Type:    aiagentorchestrator.MessageTypeEvent,
		From:    "system",
		Content: "系统事件: 工作流完成",
	}
	if err := bus.Broadcast(msg2); err != nil {
		log.Printf("  ❌ 广播消息失败: %v", err)
	} else {
		fmt.Printf("  ✅ 发送广播消息: system -> all\n")
	}

	// 等待消息处理
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("\n  📊 共收到 %d 条消息\n", receivedCount)
}

// demonstrateWorkflow 演示工作流引擎
func demonstrateWorkflow() {
	// 创建YAML工作流定义
	yamlDef := `
name: research_workflow
description: 研究和分析工作流
version: 1.0

agents:
  - name: researcher
    type: expert
    role: 搜索信息
    capabilities:
      - search
      - analyze
  - name: analyst
    type: expert
    role: 分析数据
    capabilities:
      - analyze
      - report

steps:
  - id: search
    name: 搜索信息
    type: task
    agent: researcher
    tool: web_search
    config:
      query: "Golang latest features"

  - id: analyze
    name: 分析结果
    type: task
    agent: analyst
    depends_on:
      - search

  - id: report
    name: 生成报告
    type: task
    agent: analyst
    depends_on:
      - analyze
`

	// 解析工作流
	parser := aiagentworkflow.NewParser(".")
	workflow, err := parser.ParseFromString(yamlDef, "yaml")
	if err != nil {
		log.Fatalf("  ❌ 解析工作流失败: %v", err)
	}

	fmt.Printf("  ✅ 解析工作流: %s\n", workflow.Name)
	fmt.Printf("  📝 描述: %s\n", workflow.Description)
	fmt.Printf("  📊 版本: %s\n", workflow.Version)

	// 构建DAG
	dag, err := aiagentworkflow.BuildDAGFromWorkflow(workflow)
	if err != nil {
		log.Fatalf("  ❌ 构建DAG失败: %v", err)
	}

	fmt.Printf("\n  📊 工作流结构:\n")
	fmt.Println(dag.Visualize())

	// 拓扑排序
	order, err := dag.TopologicalSort()
	if err != nil {
		log.Fatalf("  ❌ 拓扑排序失败: %v", err)
	}

	fmt.Printf("  📊 执行顺序: %v\n", order)

	// 获取并行执行组
	groups := dag.GetExecutableSteps(map[string]bool{})
	fmt.Printf("\n  📊 并行执行组:\n")
	for i, group := range groups {
		fmt.Printf("     第%d组: %v\n", i+1, group)
	}

	// 创建执行实例
	inputs := map[string]interface{}{
		"topic": "Golang",
	}
	execution := aiagentworkflow.NewWorkflowExecution(workflow, inputs)
	fmt.Printf("\n  ✅ 创建执行实例: %s\n", execution.ID)
	fmt.Printf("  📊 状态: %s\n", execution.Status)

	// 模拟执行
	for stepID, stepState := range execution.StepStates {
		stepState.Status = aiagentworkflow.StepStatusCompleted
		fmt.Printf("     ✅ 完成: %s\n", stepID)
	}

	execution.MarkCompleted()
	fmt.Printf("  ✅ 执行完成，耗时: %v\n", execution.Duration)
}

func init() {
	// 确保时间格式正确
	time.Now()
}
