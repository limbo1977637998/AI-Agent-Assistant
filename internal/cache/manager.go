package cache

import (
	"context"
	"time"

	"ai-agent-assistant/internal/tools"
)

// Manager 缓存管理器
type Manager struct {
	client           *RedisClient
	toolCache        *ToolResultCache
	llmCache         *LLMResponseCache
	enabled          bool
}

// NewManager 创建缓存管理器
func NewManager(client *RedisClient, config *CacheConfig) *Manager {
	if !config.Enabled {
		return &Manager{
			client:  client,
			enabled: false,
		}
	}

	// 解析TTL
	toolTTL, _ := time.ParseDuration(config.ToolResultTTL)
	llmTTL, _ := time.ParseDuration(config.LLMResponseTTL)

	return &Manager{
		client:    client,
		toolCache: NewToolResultCache(client, toolTTL),
		llmCache:  NewLLMResponseCache(client, llmTTL),
		enabled:   true,
	}
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled          bool
	ToolResultTTL    string
	LLMResponseTTL   string
	SessionTTL       string
	KnowledgeCacheTTL string
}

// IsEnabled 检查缓存是否启用
func (m *Manager) IsEnabled() bool {
	return m.enabled
}

// GetToolCache 获取工具缓存
func (m *Manager) GetToolCache() *ToolResultCache {
	return m.toolCache
}

// GetLLMCache 获取LLM缓存
func (m *Manager) GetLLMCache() *LLMResponseCache {
	return m.llmCache
}

// ExecuteToolWithCache 使用缓存执行工具
func (m *Manager) ExecuteToolWithCache(ctx context.Context, tool tools.Tool, args map[string]interface{}) (string, error) {
	if !m.enabled {
		return tool.Execute(ctx, args)
	}

	// 尝试从缓存获取
	if cached, found := m.toolCache.Get(ctx, tool.Name(), args); found {
		if cached.Success {
			return cached.Result.(string), nil
		}
		// 如果缓存的结果是失败，仍然执行工具
	}

	// 执行工具
	start := time.Now()
	result, err := tool.Execute(ctx, args)
	duration := time.Since(start).Milliseconds()

	// 缓存结果
	cacheResult := &ToolResult{
		Result:  result,
		Success: err == nil,
		Duration: duration,
	}

	if err != nil {
		cacheResult.ErrorMsg = err.Error()
	}

	// 异步缓存（不阻塞返回）
	go func() {
		ctx := context.Background()
		_ = m.toolCache.Set(ctx, tool.Name(), args, cacheResult)
	}()

	return result, err
}

// GetStats 获取缓存统计信息
func (m *Manager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if !m.enabled {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	toolStats, _ := m.toolCache.Stats(ctx)
	llmStats, _ := m.llmCache.Stats(ctx)

	return map[string]interface{}{
		"enabled":    true,
		"tool_cache": toolStats,
		"llm_cache":  llmStats,
	}, nil
}

// ClearAll 清空所有缓存
func (m *Manager) ClearAll(ctx context.Context) error {
	if !m.enabled {
		return nil
	}

	if err := m.toolCache.Clear(ctx); err != nil {
		return err
	}

	return m.llmCache.Clear(ctx)
}

// GetClient 获取Redis客户端
func (m *Manager) GetClient() *RedisClient {
	return m.client
}
