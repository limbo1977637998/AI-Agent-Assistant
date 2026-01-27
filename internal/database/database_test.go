package database

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestDatabaseConnection(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	client, err := NewMySQLClient(config)
	if err != nil {
		t.Fatalf("Failed to create MySQL client: %v", err)
	}
	defer client.Close()

	if err := client.GetDB().Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Database connection successful")
}

func TestSessionRepository(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Test Create - use unique session ID
	sessionID := fmt.Sprintf("test-session-%d", time.Now().UnixNano())
	userID := fmt.Sprintf("test-user-%d", time.Now().UnixNano())

	metadata := map[string]interface{}{
		"title":        "Test Session",
		"system_prompt": "You are a helpful assistant",
		"max_tokens":   2000,
		"temperature":  0.7,
		"enable_stream": true,
		"status":       "active",
	}
	metadataBytes, _ := json.Marshal(metadata)

	session := &Session{
		SessionID: sessionID,
		UserID:    userID,
		Model:     "glm-4",
		Metadata:  metadataBytes,
	}

	err = manager.Sessions.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Log("Session created successfully")

	// Test GetByID
	retrieved, err := manager.Sessions.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get session by ID: %v", err)
	}

	if retrieved.SessionID != session.SessionID {
		t.Errorf("Expected session ID %s, got %s", session.SessionID, retrieved.SessionID)
	}

	// Test GetBySessionID
	retrieved2, err := manager.Sessions.GetBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to get session by SessionID: %v", err)
	}

	if retrieved2.ID != session.ID {
		t.Errorf("Expected ID %d, got %d", session.ID, retrieved2.ID)
	}

	// Test Update
	metadata["title"] = "Updated Test Session"
	session.Metadata, _ = json.Marshal(metadata)
	err = manager.Sessions.Update(ctx, session)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Test GetByUserID
	sessions, err := manager.Sessions.GetByUserID(ctx, session.UserID, 10)
	if err != nil {
		t.Fatalf("Failed to get sessions by user ID: %v", err)
	}

	if len(sessions) == 0 {
		t.Error("Expected at least one session")
	}

	// Cleanup
	err = manager.Sessions.Delete(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	t.Log("Session repository test passed")
}

func TestMessageRepository(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Create a session first with unique ID
	sessionID := fmt.Sprintf("test-message-session-%d", time.Now().UnixNano())
	userID := fmt.Sprintf("test-user-%d", time.Now().UnixNano())

	session := &Session{
		SessionID: sessionID,
		UserID:    userID,
		Model:     "glm-4",
		Metadata:  []byte(`{"title":"Message Test Session","status":"active"}`),
	}

	err = manager.Sessions.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer manager.Sessions.Delete(ctx, session.SessionID)

	// Test Create
	message := &Message{
		SessionID:  session.SessionID,
		Role:       "user",
		Content:    "Hello, this is a test message",
		TokensUsed: 10,
	}

	err = manager.Messages.Create(ctx, message)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	t.Log("Message created successfully")

	// Test GetBySessionID
	messages, err := manager.Messages.GetBySessionID(ctx, session.SessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get messages by session ID: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Test CountBySessionID
	count, err := manager.Messages.CountBySessionID(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("Failed to count messages: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test CreateBatch
	messagesBatch := []*Message{
		{
			SessionID:  session.SessionID,
			Role:       "user",
			Content:    "Batch message 1",
			TokensUsed: 5,
		},
		{
			SessionID:  session.SessionID,
			Role:       "assistant",
			Content:    "Batch message 2",
			TokensUsed: 5,
		},
	}

	err = manager.Messages.CreateBatch(ctx, messagesBatch)
	if err != nil {
		t.Fatalf("Failed to create batch messages: %v", err)
	}

	// Verify batch insert
	messages, err = manager.Messages.GetBySessionID(ctx, session.SessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get messages after batch: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages after batch, got %d", len(messages))
	}

	t.Log("Message repository test passed")
}

func TestUserMemoryRepository(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Test Create with unique user ID
	userID := fmt.Sprintf("test-memory-user-%d", time.Now().UnixNano())

	memory := &UserMemory{
		UserID:     userID,
		Memory:     "User prefers Go programming language",
		Topics:     "programming,go",
		Importance: 5,
		MemoryType: "preference",
	}

	err = manager.UserMemories.Create(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to create user memory: %v", err)
	}

	t.Log("User memory created successfully")

	// Test GetByID
	retrieved, err := manager.UserMemories.GetByID(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to get memory by ID: %v", err)
	}

	if retrieved.Memory != memory.Memory {
		t.Errorf("Expected memory %s, got %s", memory.Memory, retrieved.Memory)
	}

	// Test GetByUserID
	memories, err := manager.UserMemories.GetByUserID(ctx, memory.UserID, 10)
	if err != nil {
		t.Fatalf("Failed to get memories by user ID: %v", err)
	}

	if len(memories) == 0 {
		t.Error("Expected at least one memory")
	}

	// Test SearchByTopic
	searchResults, err := manager.UserMemories.SearchByTopic(ctx, memory.UserID, "programming")
	if err != nil {
		t.Fatalf("Failed to search memories by topic: %v", err)
	}

	if len(searchResults) == 0 {
		t.Error("Expected at least one search result")
	}

	// Test Update
	memory.Importance = 8
	err = manager.UserMemories.Update(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	// Cleanup
	err = manager.UserMemories.Delete(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	t.Log("User memory repository test passed")
}

func TestToolCallRepository(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Create a session first with unique ID
	sessionID := fmt.Sprintf("test-tool-session-%d", time.Now().UnixNano())
	userID := fmt.Sprintf("test-user-%d", time.Now().UnixNano())

	session := &Session{
		SessionID: sessionID,
		UserID:    userID,
		Model:     "glm-4",
		Metadata:  []byte(`{"title":"Tool Test Session","status":"active"}`),
	}

	err = manager.Sessions.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer manager.Sessions.Delete(ctx, session.SessionID)

	// Test Create
	arguments := map[string]string{"query": "test search"}
	argumentsBytes, _ := json.Marshal(arguments)

	toolCall := &ToolCall{
		SessionID: session.SessionID,
		UserID:    session.UserID,
		ToolName:  "web_search",
		Arguments: argumentsBytes,
		Result:    `{"results":["result1","result2"]}`,
		Success:   true,
		Duration:  1500,
	}

	err = manager.ToolCalls.Create(ctx, toolCall)
	if err != nil {
		t.Fatalf("Failed to create tool call: %v", err)
	}

	t.Log("Tool call created successfully")

	// Test GetBySessionID
	calls, err := manager.ToolCalls.GetBySessionID(ctx, session.SessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get tool calls by session ID: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(calls))
	}

	// Test GetByToolName
	callsByName, err := manager.ToolCalls.GetByToolName(ctx, "web_search", 10)
	if err != nil {
		t.Fatalf("Failed to get tool calls by name: %v", err)
	}

	if len(callsByName) == 0 {
		t.Error("Expected at least one tool call")
	}

	t.Log("Tool call repository test passed")
}

func TestAgentRunRepository(t *testing.T) {
	config := &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: "1h",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Create a session first with unique ID
	sessionID := fmt.Sprintf("test-run-session-%d", time.Now().UnixNano())
	userID := fmt.Sprintf("test-user-%d", time.Now().UnixNano())

	session := &Session{
		SessionID: sessionID,
		UserID:    userID,
		Model:     "glm-4",
		Metadata:  []byte(`{"title":"Run Test Session","status":"active"}`),
	}

	err = manager.Sessions.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer manager.Sessions.Delete(ctx, session.SessionID)

	// Test Create with unique run ID
	toolsUsed := []string{"web_search", "calculate"}
	toolsUsedBytes, _ := json.Marshal(toolsUsed)

	run := &AgentRun{
		RunID:         fmt.Sprintf("test-run-%d", time.Now().UnixNano()),
		SessionID:     session.SessionID,
		UserID:        session.UserID,
		Input:         "Test input",
		Output:        "Test output",
		Model:         "glm-4",
		InputTokens:   50,
		OutputTokens:  100,
		TotalTokens:   150,
		EstimatedCost: 0.001,
		Latency:       2000,
		Success:       true,
		RAGUsed:       false,
		ToolsUsed:     toolsUsedBytes,
	}

	err = manager.AgentRuns.Create(ctx, run)
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	t.Log("Agent run created successfully")

	// Test GetByRunID
	retrieved, err := manager.AgentRuns.GetByRunID(ctx, run.RunID)
	if err != nil {
		t.Fatalf("Failed to get agent run by RunID: %v", err)
	}

	if retrieved.RunID != run.RunID {
		t.Errorf("Expected RunID %s, got %s", run.RunID, retrieved.RunID)
	}

	// Test GetBySessionID
	runs, err := manager.AgentRuns.GetBySessionID(ctx, session.SessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get runs by session ID: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runs))
	}

	// Test GetStatsByDate
	today := time.Now()
	stats, err := manager.AgentRuns.GetStatsByDate(ctx, today)
	if err != nil {
		t.Fatalf("Failed to get stats by date: %v", err)
	}

	if stats["total_runs"] < 1 {
		t.Error("Expected at least 1 total run in stats")
	}

	t.Logf("Stats: %+v", stats)

	t.Log("Agent run repository test passed")
}
