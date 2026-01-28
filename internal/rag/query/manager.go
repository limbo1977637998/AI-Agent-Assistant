package query

import (
	"context"
	"fmt"
	"sync"
)

// QueryOptimizerManager 查询优化器管理器
// 管理多个查询优化器，提供统一的优化接口
type QueryOptimizerManager struct {
	factory    *QueryOptimizerFactory
	optimizers map[string]QueryOptimizerStrategy
	mu         sync.RWMutex
}

// NewQueryOptimizerManager 创建查询优化器管理器
func NewQueryOptimizerManager() *QueryOptimizerManager {
	return &QueryOptimizerManager{
		factory:    NewQueryOptimizerFactory(),
		optimizers: make(map[string]QueryOptimizerStrategy),
	}
}

// RegisterOptimizer 注册优化器
func (m *QueryOptimizerManager) RegisterOptimizer(name string, optimizer QueryOptimizerStrategy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return fmt.Errorf("optimizer name cannot be empty")
	}

	if optimizer == nil {
		return fmt.Errorf("optimizer cannot be nil")
	}

	if err := optimizer.Validate(); err != nil {
		return fmt.Errorf("invalid optimizer: %w", err)
	}

	m.optimizers[name] = optimizer
	return nil
}

// GetOptimizer 获取优化器
func (m *QueryOptimizerManager) GetOptimizer(name string) (QueryOptimizerStrategy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	optimizer, ok := m.optimizers[name]
	if !ok {
		return nil, fmt.Errorf("optimizer '%s' not found", name)
	}

	return optimizer, nil
}

// CreateOptimizer 创建优化器
func (m *QueryOptimizerManager) CreateOptimizer(
	name string,
	optimizerType string,
	llm LLMProvider,
	embedding EmbeddingProvider,
	config QueryOptimizerConfig,
) error {
	optimizer, err := m.factory.CreateOptimizer(optimizerType, llm, embedding, config)
	if err != nil {
		return err
	}

	return m.RegisterOptimizer(name, optimizer)
}

// Optimize 使用指定优化器优化查询
func (m *QueryOptimizerManager) Optimize(ctx context.Context, optimizerName string, query string) ([]QueryOptimization, error) {
	optimizer, err := m.GetOptimizer(optimizerName)
	if err != nil {
		return nil, err
	}

	return optimizer.Optimize(ctx, query)
}

// OptimizeWithAll 使用所有已注册的优化器优化查询
func (m *QueryOptimizerManager) OptimizeWithAll(ctx context.Context, query string) (map[string][]QueryOptimization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string][]QueryOptimization)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(m.optimizers))

	for name, optimizer := range m.optimizers {
		wg.Add(1)
		go func(optName string, opt QueryOptimizerStrategy) {
			defer wg.Done()

			optimizations, err := opt.Optimize(ctx, query)
			if err != nil {
				errChan <- fmt.Errorf("optimizer %s failed: %w", optName, err)
				return
			}

			mu.Lock()
			results[optName] = optimizations
			mu.Unlock()
		}(name, optimizer)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if err := <-errChan; err != nil {
		return results, err
	}

	return results, nil
}

// OptimizeSequential 顺序使用多个优化器优化查询
// 前一个优化器的输出作为后一个优化器的输入
func (m *QueryOptimizerManager) OptimizeSequential(
	ctx context.Context,
	optimizerNames []string,
	query string,
) ([]QueryOptimization, error) {
	currentQuery := query
	allOptimizations := make([]QueryOptimization, 0)

	for _, optimizerName := range optimizerNames {
		optimizer, err := m.GetOptimizer(optimizerName)
		if err != nil {
			return nil, err
		}

		optimizations, err := optimizer.Optimize(ctx, currentQuery)
		if err != nil {
			return nil, fmt.Errorf("optimizer %s failed: %w", optimizerName, err)
		}

		allOptimizations = append(allOptimizations, optimizations...)

		// 使用第一个优化后的查询作为下一个优化器的输入
		if len(optimizations) > 0 {
			currentQuery = optimizations[0].Query
		}
	}

	return allOptimizations, nil
}

// ListOptimizers 列出所有已注册的优化器
func (m *QueryOptimizerManager) ListOptimizers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.optimizers))
	for name := range m.optimizers {
		names = append(names, name)
	}

	return names
}

// GetFactory 获取工厂
func (m *QueryOptimizerManager) GetFactory() *QueryOptimizerFactory {
	return m.factory
}

// ListAvailableTypes 列出所有可用的优化器类型
func (m *QueryOptimizerManager) ListAvailableTypes() []string {
	return m.factory.ListOptimizers()
}

// GetOptimizerDescription 获取优化器描述
func (m *QueryOptimizerManager) GetOptimizerDescription(optimizerType string) string {
	info := m.factory.GetOptimizerInfo(optimizerType)
	if desc, ok := info["description"].(string); ok {
		return desc
	}
	return "Unknown optimizer type"
}

// Clear 清空所有优化器
func (m *QueryOptimizerManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.optimizers = make(map[string]QueryOptimizerStrategy)
}

// Count 返回已注册优化器数量
func (m *QueryOptimizerManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.optimizers)
}
