package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// LLMResponseCache LLM响应缓存
type LLMResponseCache struct {
	client   *RedisClient
	ttl      time.Duration
	keySpace string
}

// NewLLMResponseCache 创建LLM响应缓存
func NewLLMResponseCache(client *RedisClient, ttl time.Duration) *LLMResponseCache {
	return &LLMResponseCache{
		client:   client,
		ttl:      ttl,
		keySpace: "llm:response",
	}
}

// LLMRequest LLM请求
type LLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse LLM响应
type LLMResponse struct {
	Content       string        `json:"content"`
	FinishReason  string        `json:"finish_reason"`
	TokensUsed    int           `json:"tokens_used"`
	Model         string        `json:"model"`
	ResponseTime  int64         `json:"response_time"` // 毫秒
	CachedAt      time.Time     `json:"cached_at"`
	PromptTokens  int           `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
}

// Get 获取缓存响应
func (lc *LLMResponseCache) Get(ctx context.Context, request *LLMRequest) (*LLMResponse, bool) {
	key := lc.buildKey(request)

	data, err := lc.client.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var response LLMResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, false
	}

	return &response, true
}

// Set 设置缓存响应
func (lc *LLMResponseCache) Set(ctx context.Context, request *LLMRequest, response *LLMResponse) error {
	key := lc.buildKey(request)
	response.CachedAt = time.Now()

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return lc.client.Set(ctx, key, data, lc.ttl)
}

// Delete 删除缓存
func (lc *LLMResponseCache) Delete(ctx context.Context, request *LLMRequest) error {
	key := lc.buildKey(request)
	return lc.client.Del(ctx, key)
}

// DeleteByModel 删除模型的所有缓存
func (lc *LLMResponseCache) DeleteByModel(ctx context.Context, model string) error {
	pattern := fmt.Sprintf("%s:%s:*", lc.keySpace, model)
	keys, err := lc.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return lc.client.Del(ctx, keys...)
}

// Clear 清空所有LLM缓存
func (lc *LLMResponseCache) Clear(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", lc.keySpace)
	keys, err := lc.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return lc.client.Del(ctx, keys...)
}

// Stats 获取缓存统计
func (lc *LLMResponseCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	pattern := fmt.Sprintf("%s:*", lc.keySpace)
	keys, err := lc.client.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	// 统计每个模型的缓存数量
	modelCounts := make(map[string]int64)
	for _, key := range keys {
		// 从key中提取model name
		// 格式: llm:response:model:hash
		if len(key) > len(lc.keySpace)+1 {
			rest := key[len(lc.keySpace)+1:]
			// 查找第二个冒号的位置
			idx := -1
			for i := 0; i <= len(rest)-1; i++ {
				if rest[i] == ':' {
					idx = i
					break
				}
			}
			if idx > 0 {
				modelName := rest[:idx]
				modelCounts[modelName]++
			}
		}
	}

	return map[string]interface{}{
		"total_keys":  len(keys),
		"model_counts": modelCounts,
		"ttl_seconds":  lc.ttl.Seconds(),
	}, nil
}

// GetMulti 批量获取缓存（用于多个相似的请求）
func (lc *LLMResponseCache) GetMulti(ctx context.Context, requests []*LLMRequest) ([]*LLMResponse, []int) {
	responses := make([]*LLMResponse, len(requests))
	found := make([]int, 0)

	for i, req := range requests {
		if resp, ok := lc.Get(ctx, req); ok {
			responses[i] = resp
			found = append(found, i)
		}
	}

	return responses, found
}

// SetMulti 批量设置缓存
func (lc *LLMResponseCache) SetMulti(ctx context.Context, requests []*LLMRequest, responses []*LLMResponse) error {
	for i, req := range requests {
		if i < len(responses) {
			if err := lc.Set(ctx, req, responses[i]); err != nil {
				return fmt.Errorf("failed to cache response %d: %w", i, err)
			}
		}
	}
	return nil
}

// buildKey 构建缓存键
func (lc *LLMResponseCache) buildKey(request *LLMRequest) string {
	// 创建一个规范化的请求表示用于hash
	cacheKey := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature"`
		MaxTokens   int       `json:"max_tokens"`
	}{
		Model:       request.Model,
		Messages:    request.Messages,
		Temperature: request.Temperature,
		MaxTokens:   request.MaxTokens,
	}

	keyJSON, _ := json.Marshal(cacheKey)
	hash := md5.Sum(keyJSON)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:%s:%s", lc.keySpace, request.Model, hashStr)
}
