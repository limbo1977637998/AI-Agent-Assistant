package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// MockMemoryModel 模拟记忆模型
type MockMemoryModel struct {
	summaryResponse string
}

func (m *MockMemoryModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	if m.summaryResponse != "" {
		return m.summaryResponse, nil
	}
	return "摘要：用户讨论了技术问题", nil
}

func (m *MockMemoryModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	ch := make(chan string, 1)
	resp, _ := m.Chat(ctx, messages)
	ch <- resp
	close(ch)
	return ch, nil
}

func (m *MockMemoryModel) SupportsToolCalling() bool {
	return false
}

func (m *MockMemoryModel) SupportsEmbedding() bool {
	return true
}

func (m *MockMemoryModel) Embed(ctx context.Context, text string) ([]float64, error) {
	// 返回固定向量
	vector := make([]float64, 1024)
	for i := range vector {
		vector[i] = 0.1
	}
	return vector, nil
}

func (m *MockMemoryModel) GetModelName() string {
	return "mock-memory"
}

func (m *MockMemoryModel) GetProviderName() string {
	return "mock"
}

// TestEnhancedSessionManager 测试增强版会话管理
func TestEnhancedSessionManager(t *testing.T) {
	model := &MockMemoryModel{}
	manager := NewEnhancedSessionManager(10, "memory", model)

	// 测试创建会话
	session, err := manager.GetOrCreateSession("test-session", "qwen")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", session.ID)
	}

	if session.Model != "qwen" {
		t.Errorf("Expected model 'qwen', got '%s'", session.Model)
	}

	// 测试添加消息
	message := models.Message{
		Role:    "user",
		Content: "测试消息",
	}

	err = manager.AddMessage("test-session", message)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// 获取历史
	history, err := manager.GetHistory("test-session")
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 message in history, got %d", len(history))
	}

	// 测试状态管理
	updates := map[string]interface{}{
		"username": "testuser",
		"theme":    "dark",
	}

	version, err := manager.UpdateState("test-session", updates)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}

	// 获取状态
	state, err := manager.GetState("test-session")
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if state.Data["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%v'", state.Data["username"])
	}

	t.Logf("Session state version: %d", state.Version)
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	model := &MockMemoryModel{}
	manager := NewEnhancedSessionManager(10, "memory", model)

	// 并发创建会话
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			_, err := manager.GetOrCreateSession(sessionID, "qwen")
			if err != nil {
				t.Errorf("Concurrent session creation failed: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证会话数量
	count := manager.GetSessionCount()
	if count != 10 {
		t.Errorf("Expected 10 sessions, got %d", count)
	}

	t.Logf("Successfully handled %d concurrent session creations", count)
}

// TestAutoSummary 测试自动摘要
func TestAutoSummary(t *testing.T) {
	model := &MockMemoryModel{
		summaryResponse: "摘要：用户询问了多个问题，讨论了AI技术",
	}

	manager := NewEnhancedSessionManager(5, "memory", model)
	manager.SetSummaryThreshold(3) // 3条消息后摘要
	manager.EnableAutoSummary(true)

	sessionID := "test-summary"
	session, _ := manager.GetOrCreateSession(sessionID, "qwen")

	// 添加消息直到触发摘要
	for i := 0; i < 5; i++ {
		message := models.Message{
			Role:    "user",
			Content: fmt.Sprintf("消息 %d", i+1),
		}
		manager.AddMessage(sessionID, message)
	}

	// 等待异步摘要完成
	time.Sleep(100 * time.Millisecond)

	// 检查会话
	session, _ = manager.GetSession(sessionID)
	if session.Summary == "" {
		t.Log("Summary not yet generated (async operation)")
	} else {
		t.Logf("Auto-generated summary: %s", session.Summary)
	}
}

// TestEnhancedMemoryManager 测试增强版记忆管理
func TestEnhancedMemoryManager(t *testing.T) {
	model := &MockMemoryModel{}
	memoryMgr := NewEnhancedMemoryManager(model)

	userID := "test-user"

	// 创建测试记忆
	memory := &UserMemory{
		ID:        "mem-1",
		UserID:    userID,
		Content:   "用户喜欢Go语言编程",
		Topics:    []string{"编程", "Go"},
		Importance: 0.8,
		CreatedAt: time.Now(),
	}

	// 测试添加记忆
	err := memoryMgr.AddMemory(context.Background(), memory)
	if err != nil {
		t.Fatalf("Failed to add memory: %v", err)
	}

	// 测试获取记忆
	memories := memoryMgr.GetMemories(userID, 10)
	if len(memories) == 0 {
		t.Error("Should have at least one memory")
	}

	t.Logf("Retrieved %d memories", len(memories))

	// 测试语义检索
	searchResults, err := memoryMgr.SemanticSearch(context.Background(), userID, "编程", 5)
	if err != nil {
		t.Fatalf("Semantic search failed: %v", err)
	}

	t.Logf("Semantic search returned %d results", len(searchResults))
}

// TestMemoryOptimization 测试记忆优化
func TestMemoryOptimization(t *testing.T) {
	model := &MockMemoryModel{}
	memoryMgr := NewEnhancedMemoryManager(model)

	userID := "test-opt"

	// 添加多条记忆
	memories := []*UserMemory{
		{
			ID:        "mem-1",
			UserID:    userID,
			Content:   "旧记忆1",
			Topics:    []string{"topic1"},
			Importance: 0.3,
			CreatedAt: time.Now().Add(-24 * time.Hour), // 1天前
			Vector:    make([]float64, 1024),
		},
		{
			ID:        "mem-2",
			UserID:    userID,
			Content:   "重要记忆",
			Topics:    []string{"topic2"},
			Importance: 0.9,
			CreatedAt: time.Now(),
			Vector:    make([]float64, 1024),
		},
		{
			ID:        "mem-3",
			UserID:    userID,
			Content:   "旧记忆2",
			Topics:    []string{"topic3"},
			Importance: 0.4,
			CreatedAt: time.Now().Add(-12 * time.Hour), // 12小时前
			Vector:    make([]float64, 1024),
		},
	}

	for _, mem := range memories {
		memoryMgr.AddMemory(context.Background(), mem)
	}

	// 测试重要性优化
	memoryMgr.SetOptimizationStrategy("importance")
	optimized := memoryMgr.GetMemories(userID, 10)

	if len(optimized) == 0 {
		t.Error("Optimization should not remove all memories")
	}

	// 验证排序（重要性高的应该在前）
	if len(optimized) > 1 {
		if optimized[0].Importance < optimized[1].Importance {
			t.Error("Memories should be sorted by importance (descending)")
		}
	}

	t.Logf("Optimized from %d to %d memories", len(memories), len(optimized))
}
