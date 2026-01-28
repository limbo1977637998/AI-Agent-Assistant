package chunking

import (
	"fmt"
	"sync"
)

// ChunkerRegistry 分块器注册表
// 管理所有可用的分块器策略
type ChunkerRegistry struct {
	mu       sync.RWMutex
	strategies map[string]ChunkerStrategy
}

// NewChunkerRegistry 创建分块器注册表
func NewChunkerRegistry() *ChunkerRegistry {
	registry := &ChunkerRegistry{
		strategies: make(map[string]ChunkerStrategy),
	}

	// 注册默认的分块器
	registry.registerDefaultChunkers()

	return registry
}

// registerDefaultChunkers 注册默认的分块器
// 注意: 这里只注册配置模板，实际创建时通过 Factory
func (r *ChunkerRegistry) registerDefaultChunkers() {
	// 预留接口，用于未来注册预配置的分块器实例
}

// Register 注册分块器
func (r *ChunkerRegistry) Register(name string, strategy ChunkerStrategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("chunker name cannot be empty")
	}

	if strategy == nil {
		return fmt.Errorf("chunker strategy cannot be nil")
	}

	// 验证策略
	if err := strategy.Validate(); err != nil {
		return fmt.Errorf("invalid chunker strategy: %w", err)
	}

	r.strategies[name] = strategy

	return nil
}

// Unregister 注销分块器
func (r *ChunkerRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.strategies, name)
}

// Get 获取分块器
func (r *ChunkerRegistry) Get(name string) (ChunkerStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategy, ok := r.strategies[name]
	if !ok {
		return nil, fmt.Errorf("chunker '%s' not found", name)
	}

	return strategy, nil
}

// List 列出所有已注册的分块器名称
func (r *ChunkerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}

	return names
}

// Has 检查分块器是否已注册
func (r *ChunkerRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.strategies[name]
	return ok
}

// Count 返回已注册分块器数量
func (r *ChunkerRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.strategies)
}

// Clear 清空所有注册的分块器
func (r *ChunkerRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.strategies = make(map[string]ChunkerStrategy)
}
