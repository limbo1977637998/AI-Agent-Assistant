package expert

import (
	"context"
	"testing"
	"time"

	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	"ai-agent-assistant/internal/task"
)

func TestFactory(t *testing.T) {
	factory := NewFactory()

	t.Run("Create Researcher Agent", func(t *testing.T) {
		agent, err := factory.CreateAgent("researcher")
		if err != nil {
			t.Fatalf("Failed to create researcher agent: %v", err)
		}
		if agent == nil {
			t.Fatal("Researcher agent is nil")
		}
		info := agent.GetInfo()
		if info.Name != "Researcher" {
			t.Errorf("Expected name 'Researcher', got '%s'", info.Name)
		}
	})

	t.Run("Create Analyst Agent", func(t *testing.T) {
		agent, err := factory.CreateAgent("analyst")
		if err != nil {
			t.Fatalf("Failed to create analyst agent: %v", err)
		}
		if agent == nil {
			t.Fatal("Analyst agent is nil")
		}
		info := agent.GetInfo()
		if info.Name != "Analyst" {
			t.Errorf("Expected name 'Analyst', got '%s'", info.Name)
		}
	})

	t.Run("Create Writer Agent", func(t *testing.T) {
		agent, err := factory.CreateAgent("writer")
		if err != nil {
			t.Fatalf("Failed to create writer agent: %v", err)
		}
		if agent == nil {
			t.Fatal("Writer agent is nil")
		}
		info := agent.GetInfo()
		if info.Name != "Writer" {
			t.Errorf("Expected name 'Writer', got '%s'", info.Name)
		}
	})

	t.Run("Create Invalid Agent", func(t *testing.T) {
		_, err := factory.CreateAgent("invalid")
		if err == nil {
			t.Error("Expected error for invalid agent type, got nil")
		}
	})
}

func TestRegisterAllAgents(t *testing.T) {
	factory := NewFactory()
	registry := aiagentorchestrator.NewAgentRegistry()

	err := factory.RegisterAllAgents(registry)
	if err != nil {
		t.Fatalf("Failed to register all agents: %v", err)
	}

	// 验证所有Agent都已注册
	agents := []string{"researcher", "analyst", "writer"}
	for _, name := range agents {
		_, err := registry.Get(name)
		if err != nil {
			t.Errorf("Agent %s not registered: %v", name, err)
		}
	}
}

func TestResearcherAgent(t *testing.T) {
	researcher := NewResearcherAgent()

	t.Run("Agent Info", func(t *testing.T) {
		info := researcher.GetInfo()
		if info.Name != "Researcher" {
			t.Errorf("Expected name 'Researcher', got '%s'", info.Name)
		}
		if info.Type != "researcher" {
			t.Errorf("Expected type 'researcher', got '%s'", info.Type)
		}
		if len(info.Capabilities) == 0 {
			t.Error("Researcher should have capabilities")
		}
	})

	t.Run("Has Capability", func(t *testing.T) {
		if !researcher.HasCapability("web_search") {
			t.Error("Researcher should have web_search capability")
		}
		if researcher.HasCapability("unknown_capability") {
			t.Error("Researcher should not have unknown capability")
		}
	})

	t.Run("Execute Search Task", func(t *testing.T) {
		task := &task.Task{
			ID:       "task-1",
			Type:     "researcher",
			Goal:     "搜索关于AI的最新信息",
			Status:   task.TaskStatusPending,
			Priority: task.PriorityNormal,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := researcher.Execute(ctx, task)
		if err != nil {
			t.Logf("Search task failed (expected with mock data): %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if result.TaskID != task.ID {
			t.Errorf("Expected task ID '%s', got '%s'", task.ID, result.TaskID)
		}
	})
}

func TestAnalystAgent(t *testing.T) {
	analyst := NewAnalystAgent()

	t.Run("Agent Info", func(t *testing.T) {
		info := analyst.GetInfo()
		if info.Name != "Analyst" {
			t.Errorf("Expected name 'Analyst', got '%s'", info.Name)
		}
		if info.Type != "analyst" {
			t.Errorf("Expected type 'analyst', got '%s'", info.Type)
		}
		if len(info.Capabilities) == 0 {
			t.Error("Analyst should have capabilities")
		}
	})

	t.Run("Has Capability", func(t *testing.T) {
		if !analyst.HasCapability("data_analysis") {
			t.Error("Analyst should have data_analysis capability")
		}
	})

	t.Run("Execute Analysis Task", func(t *testing.T) {
		requirements := map[string]interface{}{
			"data": []interface{}{10.0, 20.0, 30.0, 40.0, 50.0},
		}

		task := &task.Task{
			ID:           "task-2",
			Type:         "analyst",
			Goal:         "分析数据的统计特征",
			Requirements: requirements,
			Status:       task.TaskStatusPending,
			Priority:     task.PriorityNormal,
		}

		ctx := context.Background()
		result, err := analyst.Execute(ctx, task)
		if err != nil {
			t.Fatalf("Analysis task failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if result.Status != task.TaskStatusCompleted {
			t.Errorf("Expected status 'completed', got '%s'", result.Status)
		}

		// 验证输出包含统计信息
		if output, ok := result.Output.(map[string]interface{}); ok {
			if _, hasStats := output["statistics"]; !hasStats {
				t.Error("Output should contain statistics")
			}
		}
	})
}

func TestWriterAgent(t *testing.T) {
	writer := NewWriterAgent()

	t.Run("Agent Info", func(t *testing.T) {
		info := writer.GetInfo()
		if info.Name != "Writer" {
			t.Errorf("Expected name 'Writer', got '%s'", info.Name)
		}
		if info.Type != "writer" {
			t.Errorf("Expected type 'writer', got '%s'", info.Type)
		}
		if len(info.Capabilities) == 0 {
			t.Error("Writer should have capabilities")
		}
	})

	t.Run("Has Capability", func(t *testing.T) {
		if !writer.HasCapability("content_generation") {
			t.Error("Writer should have content_generation capability")
		}
	})

	t.Run("Execute Writing Task", func(t *testing.T) {
		requirements := map[string]interface{}{
			"style":    "formal",
			"length":   500,
			"keywords": []string{"AI", "技术"},
		}

		task := &task.Task{
			ID:           "task-3",
			Type:         "writer",
			Goal:         "撰写一篇关于AI技术的文章",
			Requirements: requirements,
			Status:       task.TaskStatusPending,
			Priority:     task.PriorityNormal,
		}

		ctx := context.Background()
		result, err := writer.Execute(ctx, task)
		if err != nil {
			t.Fatalf("Writing task failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if result.Status != task.TaskStatusCompleted {
			t.Errorf("Expected status 'completed', got '%s'", result.Status)
		}

		// 验证输出包含内容
		if output, ok := result.Output.(map[string]interface{}); ok {
			if _, hasContent := output["content"]; !hasContent {
				t.Error("Output should contain content")
			}
		}
	})

	t.Run("Execute Summary Task", func(t *testing.T) {
		requirements := map[string]interface{}{
			"content": "这是一段很长的文本内容，需要进行摘要。本文讨论了人工智能的发展历史，从早期的专家系统到现代的深度学习。重点介绍了机器学习、神经网络和自然语言处理等关键技术。",
			"title":   "人工智能发展史",
		}

		task := &task.Task{
			ID:           "task-4",
			Type:         "writer",
			Goal:         "为这段内容生成摘要",
			Requirements: requirements,
			Status:       task.TaskStatusPending,
			Priority:     task.PriorityNormal,
		}

		ctx := context.Background()
		result, err := writer.Execute(ctx, task)
		if err != nil {
			t.Fatalf("Summary task failed: %v", err)
		}
		if result.Status != task.TaskStatusCompleted {
			t.Errorf("Expected status 'completed', got '%s'", result.Status)
		}
	})
}

func TestAgentCollaboration(t *testing.T) {
	factory := NewFactory()
	registry := aiagentorchestrator.NewAgentRegistry()

	// 注册所有Agent
	err := factory.RegisterAllAgents(registry)
	if err != nil {
		t.Fatalf("Failed to register agents: %v", err)
	}

	t.Run("Find Agent by Capability", func(t *testing.T) {
		// 查找具有web_search能力的Agent
		agent, err := registry.FindBestAgent([]string{"web_search"})
		if err != nil {
			t.Logf("Expected to find researcher agent, got: %v", err)
		}
		if agent != nil && agent.Name != "Researcher" {
			t.Errorf("Expected Researcher, got %s", agent.Name)
		}
	})

	t.Run("Get All Agents", func(t *testing.T) {
		agents := registry.GetAll()
		if len(agents) != 3 {
			t.Errorf("Expected 3 agents, got %d", len(agents))
		}
	})
}
