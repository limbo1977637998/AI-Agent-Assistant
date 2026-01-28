package adaptive

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// QueryRouter 查询路由器
//
// 功能: 根据查询特征自动选择最优的检索策略
//
// 支持的策略:
//   - vector: 纯向量检索
//   - hybrid: 混合检索 (Dense + Sparse)
//   - graph_rag: Graph RAG
//   - hyde: HyDE 检索
//   - semantic: 语义分块检索
//   - recursive: 递归分块检索
type QueryRouter struct {
	strategies       map[string]RetrievalStrategy
	performanceData map[string]*StrategyPerformance
	llm             LLMProvider
	config          RouterConfig
}

// RouterConfig 路由器配置
type RouterConfig struct {
	// DefaultStrategy 默认策略
	DefaultStrategy string

	// EnableMLRouting 是否启用基于 ML 的路由
	EnableMLRouting bool

	// EnableAHeuristicRouting 是否启用启发式路由
	EnableAHeuristicRouting bool

	// RoutingRules 路由规则
	RoutingRules []RoutingRule
}

// RoutingRule 路由规则
type RoutingRule struct {
	// Name 规则名称
	Name string

	// Condition 匹配条件
	Condition string // 规则表达式或关键词

	// Strategy 目标策略
	Strategy string

	// Priority 优先级
	Priority int
}

// RetrievalStrategy 检索策略
type RetrievalStrategy struct {
	Name        string
	Description string
	Config      map[string]interface{}
}

// StrategyPerformance 策略性能数据
type StrategyPerformance struct {
	Strategy          string
	TotalQueries       int
	SuccessCount       int
	AverageLatency     int64  // 毫秒
	AverageScore       float64
	UserSatisfaction  float64
	LastUpdated        int64
}

// DefaultRouterConfig 返回默认配置
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		DefaultStrategy:         "hybrid",
		EnableMLRouting:          false,
		EnableAHeuristicRouting:  true,
		RoutingRules: []RoutingRule{
			{
				Name:     "全局性问题",
				Condition: "(总体|整体|所有|总结|架构|概述)",
				Strategy: "graph_rag",
				Priority: 10,
			},
			{
				Name:     "事实查询",
				Condition: "(什么是|如何|怎么|哪个|哪些|列表)",
				Strategy: "vector",
				Priority: 8,
			},
			{
				Name:     "复杂推理",
				Condition: "(为什么|原因|分析|比较|关系)",
				Strategy: "graph_rag",
				Priority: 9,
			},
			{
				Name:     "模糊查询",
				Condition: "(大概|可能|或许|应该)",
				Strategy: "hyde",
				Priority: 7,
			},
			{
				Name:     "技术术语",
				Condition: "(算法|模型|架构|API|协议)",
				Strategy: "hybrid",
				Priority: 6,
			},
		},
	}
}

// NewQueryRouter 创建查询路由器
func NewQueryRouter(llm LLMProvider, config RouterConfig) (*QueryRouter, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	router := &QueryRouter{
		strategies:       make(map[string]RetrievalStrategy),
		performanceData:   make(map[string]*StrategyPerformance),
		llm:             llm,
		config:          config,
	}

	// 注册默认策略
	router.registerDefaultStrategies()

	return router, nil
}

// registerDefaultStrategies 注册默认策略
func (qr *QueryRouter) registerDefaultStrategies() {
	strategies := []RetrievalStrategy{
		{
			Name:        "vector",
			Description: "纯向量检索",
			Config:      map[string]interface{}{"top_k": 10},
		},
		{
			Name:        "hybrid",
			Description: "混合检索 (Dense + Sparse)",
			Config:      map[string]interface{}{"top_k": 10, "dense_weight": 0.7, "sparse_weight": 0.3},
		},
		{
			Name:        "graph_rag",
			Description: "Graph RAG 检索",
			Config:      map[string]interface{}{"top_k": 5, "search_type": "community"},
		},
		{
			Name:        "hyde",
			Description: "HyDE 检索",
			Config:      map[string]interface{}{"top_k": 10},
		},
		{
			Name:        "semantic",
			Description: "语义分块检索",
			Config:      map[string]interface{}{"chunk_size": 500},
		},
		{
			Name:        "recursive",
			Description: "递归分块检索",
			Config:      map[string]interface{}{"chunk_size": 500, "overlap": 50},
		},
	}

	for _, strategy := range strategies {
		qr.strategies[strategy.Name] = strategy
	}
}

// SelectStrategy 选择最优策略
func (qr *QueryRouter) SelectStrategy(ctx context.Context, query string) (string, error) {
	// 方法 1: 启发式路由 (基于规则)
	if qr.config.EnableAHeuristicRouting {
		strategy := qr.routeByHeuristics(query)
		if strategy != "" {
			return strategy, nil
		}
	}

	// 方法 2: 基于 ML 的路由 (可选)
	if qr.config.EnableMLRouting {
		strategy, err := qr.routeByML(ctx, query)
		if err == nil && strategy != "" {
			return strategy, nil
		}
	}

	// 方法 3: 基于历史性能
	strategy := qr.routeByPerformance()

	// 方法 4: 使用默认策略
	if strategy == "" {
		strategy = qr.config.DefaultStrategy
	}

	return strategy, nil
}

// routeByHeuristics 基于启发式规则路由
func (qr *QueryRouter) routeByHeuristics(query string) string {
	// 按优先级匹配规则
	bestMatch := ""
	bestPriority := -1

	for _, rule := range qr.config.RoutingRules {
		// 检查条件
		matched, _ := regexp.MatchString(rule.Condition, query)
		if matched && rule.Priority > bestPriority {
			bestMatch = rule.Strategy
			bestPriority = rule.Priority
		}
	}

	return bestMatch
}

// routeByML 基于 ML 路由 (可选)
func (qr *QueryRouter) routeByML(ctx context.Context, query string) (string, error) {
	// 使用 LLM 分类查询类型
	prompt := fmt.Sprintf(`请分析以下查询的类型，并选择最适合的检索策略。

查询: %s

可选策略:
1. vector - 纯向量检索，适合精确匹配
2. hybrid - 混合检索，平衡精度和召回
3. graph_rag - 图检索，适合全局性问题
4. hyde - 假设文档嵌入，适合模糊查询

请只返回策略名称 (vector/hybrid/graph_rag/hyde):`, query)

	response, err := qr.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	// 提取策略名称
	response = strings.ToLower(strings.TrimSpace(response))
	for _, strategy := range []string{"vector", "hybrid", "graph_rag", "hyde"} {
		if strings.Contains(response, strategy) {
			return strategy, nil
		}
	}

	// 如果无法识别，返回空
	return "", nil
}

// routeByPerformance 基于历史性能路由
func (qr *QueryRouter) routeByPerformance() string {
	bestStrategy := ""
	bestScore := 0.0

	for strategy, perf := range qr.performanceData {
		if perf.TotalQueries < 5 {
			// 样本太少，跳过
			continue
		}

		// 综合得分 = 成功率 (40%) + 用户满意度 (40%) + 性能 (20%)
		score := float64(perf.SuccessCount)/float64(perf.TotalQueries)*0.4 +
			perf.UserSatisfaction*0.4 +
			(1.0-float64(perf.AverageLatency)/1000.0)*0.2

		if score > bestScore {
			bestScore = score
			bestStrategy = strategy
		}
	}

	return bestStrategy
}

// OptimizeParameters 优化检索参数
func (qr *QueryRouter) OptimizeParameters(ctx context.Context, queryType string) (map[string]interface{}, error) {
	// 获取策略的性能数据
	perf, exists := qr.performanceData[queryType]
	if !exists || perf.TotalQueries < 10 {
		// 数据不足，返回默认参数
		return qr.getDefaultParameters(queryType), nil
	}

	// 基于历史性能优化参数
	params := qr.getDefaultParameters(queryType)

	// 根据用户满意度调整
	if perf.UserSatisfaction < 0.6 {
		// 用户不满意，增加 Top-K
		if topK, ok := params["top_k"].(int); ok {
			params["top_k"] = topK * 2
		}
	} else if perf.UserSatisfaction > 0.8 {
		// 用户满意，可以减少 Top-K 提高性能
		if topK, ok := params["top_k"].(int); ok {
			params["top_k"] = max(5, topK/2)
		}
	}

	return params, nil
}

// getDefaultParameters 获取默认参数
func (qr *QueryRouter) getDefaultParameters(strategy string) map[string]interface{} {
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

	case "semantic":
		params["chunk_size"] = 500
		params["overlap"] = 50

	case "recursive":
		params["chunk_size"] = 500
		params["overlap"] = 50

	default:
		params["top_k"] = 10
	}

	return params
}

// RecordFeedback 记录反馈
func (qr *QueryRouter) RecordFeedback(ctx context.Context, query string, strategy string, result *RAGExecutionResult) error {
	perf, exists := qr.performanceData[strategy]
	if !exists {
		perf = &StrategyPerformance{
			Strategy: strategy,
		}
		qr.performanceData[strategy] = perf
	}

	// 更新统计数据
	perf.TotalQueries++
	if result.Success {
		perf.SuccessCount++
	}
	perf.AverageLatency = (perf.AverageLatency*int64(perf.TotalQueries-1) + result.Latency) / int64(perf.TotalQueries)
	perf.AverageScore = (perf.AverageScore*float64(perf.TotalQueries-1) + result.Score) / float64(perf.TotalQueries)
	perf.UserSatisfaction = (perf.UserSatisfaction*float64(perf.TotalQueries-1) + result.UserFeedback) / float64(perf.TotalQueries)
	perf.LastUpdated = 0 // 简化实现

	return nil
}

// GetStrategyPerformance 获取策略性能
func (qr *QueryRouter) GetStrategyPerformance(strategy string) (*StrategyPerformance, bool) {
	perf, exists := qr.performanceData[strategy]
	return perf, exists
}

// GetAllPerformance 获取所有策略性能
func (qr *QueryRouter) GetAllPerformance() map[string]*StrategyPerformance {
	return qr.performanceData
}

// GetStrategy 获取策略信息
func (qr *QueryRouter) GetStrategy(name string) (RetrievalStrategy, bool) {
	strategy, ok := qr.strategies[name]
	return strategy, ok
}

// ListStrategies 列出所有策略
func (qr *QueryRouter) ListStrategies() []string {
	strategies := make([]string, 0, len(qr.strategies))
	for name := range qr.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}
