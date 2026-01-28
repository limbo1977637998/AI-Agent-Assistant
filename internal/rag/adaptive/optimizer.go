package adaptive

import (
	"context"
	"fmt"
	"math"
	"sync"
)

// ParameterOptimizer 参数优化器
//
// 功能: 根据历史性能自动优化检索参数
//
// 优化方法:
//   1. 网格搜索
//   2. 贝叶斯优化
//   3. 强化学习 (可选)
type ParameterOptimizer struct {
	performanceData map[string]*StrategyPerformance
	optimalParams  map[string]map[string]interface{}
	llm             LLMProvider
	config          OptimizerConfig
	mu              sync.RWMutex
}

// OptimizerConfig 优化器配置
type OptimizerConfig struct {
	// OptimizationMethod 优化方法 (grid, bayesian, reinforcement)
	OptimizationMethod string

	// MinSamples 最小样本数
	MinSamples int

	// OptimizationInterval 优化间隔 (查询次数)
	OptimizationInterval int

	// EnableAutoOptimization 是否启用自动优化
	EnableAutoOptimization bool
}

// DefaultOptimizerConfig 返回默认配置
func DefaultOptimizerConfig() OptimizerConfig {
	return OptimizerConfig{
		OptimizationMethod:     "grid", // 简单但有效
		MinSamples:             50,
		OptimizationInterval:   100,
		EnableAutoOptimization: true,
	}
}

// NewParameterOptimizer 创建参数优化器
func NewParameterOptimizer(llm LLMProvider, config OptimizerConfig) (*ParameterOptimizer, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	return &ParameterOptimizer{
		performanceData: make(map[string]*StrategyPerformance),
		optimalParams:  make(map[string]map[string]interface{}),
		llm:             llm,
		config:          config,
	}, nil
}

// OptimizeParameters 优化参数
func (po *ParameterOptimizer) OptimizeParameters(ctx context.Context, strategy string) (map[string]interface{}, error) {
	po.mu.Lock()
	defer po.mu.Unlock()

	perf, exists := po.performanceData[strategy]
	if !exists || perf.TotalQueries < po.config.MinSamples {
		// 数据不足，返回默认参数
		return po.getDefaultParameters(strategy), nil
	}

	// 检查是否需要优化
	queryCount := perf.TotalQueries
	if queryCount%po.config.OptimizationInterval != 0 {
		// 还未到优化时间
		return po.getCurrentParams(strategy), nil
	}

	// 根据优化方法选择优化算法
	switch po.config.OptimizationMethod {
	case "grid":
		return po.gridSearchOptimize(ctx, strategy, perf)
	case "bayesian":
		return po.bayesianOptimize(ctx, strategy, perf)
	default:
		return po.getDefaultParameters(strategy), nil
	}
}

// gridSearchOptimize 网格搜索优化
func (po *ParameterOptimizer) gridSearchOptimize(ctx context.Context, strategy string, perf *StrategyPerformance) (map[string]interface{}, error) {
	// 定义搜索空间
	searchSpace := po.defineSearchSpace(strategy)

	bestParams := make(map[string]interface{})

	// 网格搜索
	for paramName, paramRange := range searchSpace {
		bestValue := po.testParameterRange(ctx, strategy, paramName, paramRange, perf)
		bestParams[paramName] = bestValue
	}

	// 保存最优参数
	po.optimalParams[strategy] = bestParams

	return bestParams, nil
}

// bayesianOptimize 贝叶斯优化 (简化版)
func (po *ParameterOptimizer) bayesianOptimize(ctx context.Context, strategy string, perf *StrategyPerformance) (map[string]interface{}, error) {
	// 简化实现：使用网格搜索作为替代
	// 完整的贝叶斯优化需要更复杂的实现
	return po.gridSearchOptimize(ctx, strategy, perf)
}

// testParameterRange 测试参数范围
func (po *ParameterOptimizer) testParameterRange(ctx context.Context, strategy, paramName string, paramRange ParameterRange, perf *StrategyPerformance) interface{} {
	bestValue := paramRange.Default
	bestScore := perf.AverageScore

	// 测试范围内的值
	step := (paramRange.Max - paramRange.Min) / 10
	for i := 0; i <= 10; i++ {
		value := paramRange.Min + float64(i)*step

		// 更新参数
		params := po.getCurrentParams(strategy)
		params[paramName] = value

		// 估算新得分 (简化实现)
	estimatedScore := po.estimateScore(ctx, strategy, params, perf)

		if estimatedScore > bestScore {
			bestScore = estimatedScore
			bestValue = value
		}
	}

	return bestValue
}

// defineSearchSpace 定义搜索空间
func (po *ParameterOptimizer) defineSearchSpace(strategy string) map[string]ParameterRange {
	searchSpace := make(map[string]ParameterRange)

	switch strategy {
	case "vector":
		searchSpace["top_k"] = ParameterRange{
			Min:     5,
			Max:     20,
			Default: 10,
		}

	case "hybrid":
		searchSpace["top_k"] = ParameterRange{
			Min:     5,
			Max:     20,
			Default: 10,
		}
		searchSpace["dense_weight"] = ParameterRange{
			Min:     0.1,
			Max:     0.9,
			Default: 0.7,
		}
		searchSpace["sparse_weight"] = ParameterRange{
			Min:     0.1,
			Max:     0.9,
			Default: 0.3,
		}

	case "graph_rag":
		searchSpace["top_k"] = ParameterRange{
			Min:     3,
			Max:     15,
			Default: 5,
		}

	default:
		// 默认搜索空间
		searchSpace["top_k"] = ParameterRange{
			Min:     5,
			Max:     20,
			Default: 10,
		}
	}

	return searchSpace
}

// ParameterRange 参数范围
type ParameterRange struct {
	Min    float64
	Max    float64
	Default float64
}

// getCurrentParams 获取当前参数
func (po *ParameterOptimizer) getCurrentParams(strategy string) map[string]interface{} {
	if params, ok := po.optimalParams[strategy]; ok {
		return params
	}

	return po.getDefaultParameters(strategy)
}

// getDefaultParameters 获取默认参数
func (po *ParameterOptimizer) getDefaultParameters(strategy string) map[string]interface{} {
	params := make(map[string]interface{})

	switch strategy {
	case "vector":
		params["top_k"] = 10
		params["metric_type"] = "COSINE"

	case "hybrid":
		params["top_k"] = 10
		params["dense_weight"] = 0.7
		params["sparse_weight"] = 0.3

	case "graph_rag":
		params["top_k"] = 5
		params["search_type"] = "community"

	case "hyde":
		params["top_k"] = 10

	default:
		params["top_k"] = 10
	}

	return params
}

// estimateScore 估算参数得分
func (po *ParameterOptimizer) estimateScore(ctx context.Context, strategy string, params map[string]interface{}, perf *StrategyPerformance) float64 {
	// 简化实现：基于参数距离最优值的程度
	optimalParams := po.optimalParams[strategy]
	if optimalParams == nil {
		return perf.AverageScore
	}

	// 计算参数相似度
	similarity := 0.0
	paramCount := 0

	for key, optimalValue := range optimalParams {
		currentValue, exists := params[key]
		if !exists {
			continue
		}

		// 计算归一化距离
		var distance float64
		switch v := optimalValue.(type) {
		case int:
			currentInt, _ := currentValue.(int)
			optimalInt := int(v)
			distance = math.Abs(float64(currentInt - optimalInt))
		case float64:
			currentFloat, _ := currentValue.(float64)
			distance = math.Abs(currentFloat - v)
		default:
			distance = 0
		}

		// 归一化 [0, 1]
		similarity += 1.0 - distance
		paramCount++
	}

	if paramCount > 0 {
		similarity /= float64(paramCount)
	}

	// 加权平均：70% 历史性能 + 30% 参数相似度
	score := perf.AverageScore*0.7 + similarity*0.3

	return score
}

// RecordPerformance 记录性能数据
func (po *ParameterOptimizer) RecordPerformance(ctx context.Context, strategy string, executionResult *RAGExecutionResult) error {
	po.mu.Lock()
	defer po.mu.Unlock()

	perf, exists := po.performanceData[strategy]
	if !exists {
		perf = &StrategyPerformance{
			Strategy: strategy,
		}
		po.performanceData[strategy] = perf
	}

	// 更新统计数据
	perf.TotalQueries++
	if executionResult.Success {
		perf.SuccessCount++
	}
	perf.AverageLatency = (perf.AverageLatency*int64(perf.TotalQueries-1) + executionResult.Latency) / int64(perf.TotalQueries)
	perf.AverageScore = (perf.AverageScore*float64(perf.TotalQueries-1) + executionResult.Score) / float64(perf.TotalQueries)
	perf.UserSatisfaction = (perf.UserSatisfaction*float64(perf.TotalQueries-1) + executionResult.UserFeedback) / float64(perf.TotalQueries)

	// 检查是否需要优化
	if perf.TotalQueries%po.config.OptimizationInterval == 0 && po.config.EnableAutoOptimization {
		// 异步优化参数
		go po.optimizeAsync(ctx, strategy, perf)
	}

	return nil
}

// optimizeAsync 异步优化参数
func (po *ParameterOptimizer) optimizeAsync(ctx context.Context, strategy string, perf *StrategyPerformance) {
	// 简化实现：直接调用同步优化
	_, err := po.OptimizeParameters(ctx, strategy)
	if err != nil {
		// 记录错误
		fmt.Printf("参数优化失败: %v\n", err)
	}
}

// GetOptimalParams 获取最优参数
func (po *ParameterOptimizer) GetOptimalParams(strategy string) (map[string]interface{}, bool) {
	po.mu.RLock()
	defer po.mu.RUnlock()

	params, ok := po.optimalParams[strategy]
	return params, ok
}

// GetAllPerformance 获取所有性能数据
func (po *ParameterOptimizer) GetAllPerformance() map[string]*StrategyPerformance {
	po.mu.RLock()
	defer po.mu.RUnlock()

	// 返回副本
	result := make(map[string]*StrategyPerformance)
	for k, v := range po.performanceData {
		result[k] = v
	}

	return result
}
