package cache

import (
	"context"
	"testing"
	"time"
)

func TestRedisConnection(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "redis_pass_1977637998",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.GetClient().Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	t.Log("Redis connection successful")
}

func TestToolCache(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "redis_pass_1977637998",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	cache := NewToolResultCache(client, 5*time.Minute)

	// Test Set
	toolName := "test_tool"
	arguments := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}

	result := &ToolResult{
		Result:   "test result",
		Success:  true,
		Duration: 100,
	}

	err = cache.Set(ctx, toolName, arguments, result)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	t.Log("Tool cache set successful")

	// Test Get
	cached, found := cache.Get(ctx, toolName, arguments)
	if !found {
		t.Fatal("Cache not found")
	}

	if cached.Result != "test result" {
		t.Errorf("Expected 'test result', got '%v'", cached.Result)
	}

	t.Log("Tool cache get successful")

	// Test Delete
	err = cache.Delete(ctx, toolName, arguments)
	if err != nil {
		t.Fatalf("Failed to delete cache: %v", err)
	}

	// Verify deletion
	_, found = cache.Get(ctx, toolName, arguments)
	if found {
		t.Error("Cache should be deleted")
	}

	t.Log("Tool cache delete successful")
}

func TestLLMCache(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "redis_pass_1977637998",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	cache := NewLLMResponseCache(client, 10*time.Minute)

	// Test Set
	request := &LLMRequest{
		Model: "glm-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	response := &LLMResponse{
		Content:           "Hello! How can I help you?",
		FinishReason:      "stop",
		TokensUsed:        10,
		Model:             "glm-4",
		ResponseTime:      500,
		PromptTokens:      5,
		CompletionTokens: 5,
	}

	err = cache.Set(ctx, request, response)
	if err != nil {
		t.Fatalf("Failed to set LLM cache: %v", err)
	}

	t.Log("LLM cache set successful")

	// Test Get
	cached, found := cache.Get(ctx, request)
	if !found {
		t.Fatal("LLM cache not found")
	}

	if cached.Content != "Hello! How can I help you?" {
		t.Errorf("Expected 'Hello! How can I help you?', got '%s'", cached.Content)
	}

	t.Log("LLM cache get successful")

	// Test Stats
	stats, err := cache.Stats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	t.Logf("LLM cache stats: %+v", stats)

	// Test Delete
	err = cache.Delete(ctx, request)
	if err != nil {
		t.Fatalf("Failed to delete LLM cache: %v", err)
	}

	t.Log("LLM cache delete successful")
}

func TestCacheManager(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "redis_pass_1977637998",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	cacheConfig := &CacheConfig{
		Enabled:        true,
		ToolResultTTL:  "5m",
		LLMResponseTTL: "10m",
	}

	manager := NewManager(client, cacheConfig)

	ctx := context.Background()

	// Test IsEnabled
	if !manager.IsEnabled() {
		t.Error("Cache should be enabled")
	}

	// Test GetStats
	stats, err := manager.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	t.Logf("Cache manager stats: %+v", stats)

	// Test GetToolCache
	toolCache := manager.GetToolCache()
	if toolCache == nil {
		t.Error("Tool cache should not be nil")
	}

	// Test GetLLMCache
	llmCache := manager.GetLLMCache()
	if llmCache == nil {
		t.Error("LLM cache should not be nil")
	}

	t.Log("Cache manager test passed")
}

func TestBasicRedisOperations(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "redis_pass_1977637998",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test Set and Get
	key := "test:key"
	value := "test_value"

	err = client.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	retrieved, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if retrieved != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}

	// Test Exists
	exists, err := client.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}

	if exists != 1 {
		t.Error("Key should exist")
	}

	// Test Del
	err = client.Del(ctx, key)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify deletion
	exists, _ = client.Exists(ctx, key)
	if exists != 0 {
		t.Error("Key should not exist")
	}

	t.Log("Basic Redis operations test passed")
}
