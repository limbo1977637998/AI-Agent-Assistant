package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ToolResultCache 工具结果缓存
type ToolResultCache struct {
	client   *RedisClient
	ttl      time.Duration
	keySpace string
}

// NewToolResultCache 创建工具结果缓存
func NewToolResultCache(client *RedisClient, ttl time.Duration) *ToolResultCache {
	return &ToolResultCache{
		client:   client,
		ttl:      ttl,
		keySpace: "tool:result",
	}
}

// ToolCall 工具调用记录
type ToolCall struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult 工具结果
type ToolResult struct {
	Result   interface{} `json:"result"`
	Success  bool        `json:"success"`
	ErrorMsg string      `json:"error_msg,omitempty"`
	Duration int64       `json:"duration"` // 毫秒
	CachedAt time.Time   `json:"cached_at"`
}

// Get 获取缓存结果
func (tc *ToolResultCache) Get(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolResult, bool) {
	key := tc.buildKey(toolName, arguments)

	data, err := tc.client.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var result ToolResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, false
	}

	return &result, true
}

// Set 设置缓存结果
func (tc *ToolResultCache) Set(ctx context.Context, toolName string, arguments map[string]interface{}, result *ToolResult) error {
	key := tc.buildKey(toolName, arguments)
	result.CachedAt = time.Now()

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	return tc.client.Set(ctx, key, data, tc.ttl)
}

// Delete 删除缓存
func (tc *ToolResultCache) Delete(ctx context.Context, toolName string, arguments map[string]interface{}) error {
	key := tc.buildKey(toolName, arguments)
	return tc.client.Del(ctx, key)
}

// DeleteByTool 删除工具的所有缓存
func (tc *ToolResultCache) DeleteByTool(ctx context.Context, toolName string) error {
	pattern := fmt.Sprintf("%s:%s:*", tc.keySpace, toolName)
	keys, err := tc.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return tc.client.Del(ctx, keys...)
}

// Clear 清空所有工具缓存
func (tc *ToolResultCache) Clear(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", tc.keySpace)
	keys, err := tc.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return tc.client.Del(ctx, keys...)
}

// Stats 获取缓存统计
func (tc *ToolResultCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	pattern := fmt.Sprintf("%s:*", tc.keySpace)
	keys, err := tc.client.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	// 统计每个工具的缓存数量
	toolCounts := make(map[string]int64)
	for _, key := range keys {
		// 从key中提取tool name
		// 格式: tool:result:tool_name:hash
		if len(key) > len(tc.keySpace)+1 {
			rest := key[len(tc.keySpace)+1:]
			// 查找第二个冒号的位置
			idx := -1
			for i := 0; i <= len(rest)-1; i++ {
				if rest[i] == ':' {
					idx = i
					break
				}
			}
			if idx > 0 {
				toolName := rest[:idx]
				toolCounts[toolName]++
			}
		}
	}

	return map[string]interface{}{
		"total_keys": len(keys),
		"tool_counts": toolCounts,
		"ttl_seconds": tc.ttl.Seconds(),
	}, nil
}

// buildKey 构建缓存键
func (tc *ToolResultCache) buildKey(toolName string, arguments map[string]interface{}) string {
	// 将参数序列化为JSON以生成hash
	argsJSON, _ := json.Marshal(arguments)
	hash := md5.Sum(argsJSON)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:%s:%s", tc.keySpace, toolName, hashStr)
}
