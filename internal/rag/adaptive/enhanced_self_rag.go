package adaptive

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// EnhancedSelfRAG 增强版 Self-RAG
//
// 新增功能:
//   1. 动态反思阈值调整
//   2. 多维度质量评估
//   3. 自适应检索策略
//   4. 纠错建议生成
//   5. 性能追踪和优化
//
// 论文:
//   "Self-RAG: Learning to Retrieve, Generate, and Critique through Self-Reflection"
type EnhancedSelfRAG struct {
	SelfReflectiveRAG
	qualityEvaluator QualityEvaluator
	strategyAdapter  StrategyAdapter
	perfTracker      *PerformanceTracker
	config           EnhancedSelfRAGConfig
}

// EnhancedSelfRAGConfig 增强 Self-RAG 配置
type EnhancedSelfRAGConfig struct {
	SelfRAGConfig

	// DynamicThresholding 是否启用动态阈值
	DynamicThresholding bool

	// MultiDimensionalEval 是否启用多维度评估
	MultiDimensionalEval bool

	// AdaptiveStrategy 是否启用自适应策略
	AdaptiveStrategy bool

	// MinImprovementRate 最小改进率
	MinImprovementRate float64

	// ReflectionTokens 反思使用的最大 token 数
	ReflectionTokens int
}

// QualityEvaluator 质量评估器接口
type QualityEvaluator interface {
	EvaluateQuality(ctx context.Context, query string, retrievedDocs []string, answer string) (*QualityMetrics, error)
}

// StrategyAdapter 策略适配器接口
type StrategyAdapter interface {
	AdaptStrategy(ctx context.Context, metrics *QualityMetrics) (string, error)
}

// PerformanceTracker 性能追踪器
type PerformanceTracker struct {
	queryHistory map[string][]QueryPerformance
	mu           sync.RWMutex
}

// QueryPerformance 查询性能记录
type QueryPerformance struct {
	Query         string
	Timestamp     time.Time
	RetrievalScore float64
	AnswerScore    float64
	ReflectionUsed bool
	Strategy       string
	Latency        time.Duration
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	// 检索质量
	RelevanceScore    float64 // 相关性分数
	CoverageScore     float64 // 覆盖率分数
	RankingScore      float64 // 排序质量分数

	// 答案质量
	AccuracyScore     float64 // 准确性分数
	CohesionScore     float64 // 连贯性分数
	CompletenessScore float64 // 完整性分数

	// 综合质量
	OverallScore      float64 // 综合分数

	// 问题诊断
	Issues            []string
	Suggestions       []string
}

// DefaultEnhancedSelfRAGConfig 返回默认配置
func DefaultEnhancedSelfRAGConfig() EnhancedSelfRAGConfig {
	return EnhancedSelfRAGConfig{
		SelfRAGConfig:     DefaultSelfRAGConfig(),
		DynamicThresholding: true,
		MultiDimensionalEval: true,
		AdaptiveStrategy:    true,
		MinImprovementRate:   0.1, // 10% 最小改进
		ReflectionTokens:    200,
	}
}

// NewEnhancedSelfRAG 创建增强版 Self-RAG
func NewEnhancedSelfRAG(llm LLMProvider, config EnhancedSelfRAGConfig) (*EnhancedSelfRAG, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	baseSelfRAG, err := NewSelfReflectiveRAG(llm, config.SelfRAGConfig)
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedSelfRAG{
		SelfReflectiveRAG: *baseSelfRAG,
		qualityEvaluator: &DefaultQualityEvaluator{llm: llm},
		strategyAdapter:  &DefaultStrategyAdapter{},
		perfTracker:      &PerformanceTracker{
			queryHistory: make(map[string][]QueryPerformance),
		},
		config: config,
	}

	return enhanced, nil
}

// EnhancedRetrieve 增强版检索（带自我反思）
func (esr *EnhancedSelfRAG) EnhancedRetrieve(ctx context.Context, query string, initialTopK int) ([]string, *QualityMetrics, error) {
	startTime := time.Now()

	// 1. 初始检索
	retrievedDocs := make([]string, 0)
	var bestDocs []string
	var bestScore float64
	var bestMetrics *QualityMetrics

	currentDocs := esr.performInitialRetrieval(ctx, query, initialTopK)
	retrievedDocs = append(retrievedDocs, currentDocs...)

	// 2. 多轮反思检索
	maxRounds := esr.config.MaxRetries
	if maxRounds <= 0 {
		maxRounds = 2
	}

	for round := 0; round < maxRounds; round++ {
		// 生成临时答案
		tempAnswer := esr.generateTempAnswer(ctx, query, currentDocs)

		// 评估质量
		metrics, err := esr.evaluateQuality(ctx, query, currentDocs, tempAnswer)
		if err != nil {
			break
		}

		// 记录最佳结果
		if metrics.OverallScore > bestScore {
			bestScore = metrics.OverallScore
			bestDocs = currentDocs
			bestMetrics = metrics
		}

		// 检查是否满足质量阈值
		threshold := esr.config.MinScore
		if esr.config.DynamicThresholding {
			threshold = esr.calculateDynamicThreshold(query, round)
		}

		if metrics.OverallScore >= threshold {
			break
		}

		// 生成反思并调整
		_, adjustments := esr.generateReflectionWithAdjustments(ctx, query, metrics)

		// 应用调整
		if adjustments.NeedMoreDocs {
			adjustedTopK := esr.adjustTopK(initialTopK, metrics.OverallScore)
			newDocs := esr.performAdditionalRetrieval(ctx, query, adjustedTopK)
			currentDocs = append(currentDocs, newDocs...)
			retrievedDocs = append(retrievedDocs, newDocs...)
		}

		if adjustments.ChangeStrategy {
			newStrategy, _ := esr.strategyAdapter.AdaptStrategy(ctx, metrics)
			// 可以使用新策略重新检索
			_ = newStrategy
		}
	}

	// 3. 使用最佳结果
	if bestDocs != nil {
		retrievedDocs = bestDocs
	}

	// 4. 记录性能
	latency := time.Since(startTime)
	esr.recordPerformance(query, bestScore, len(retrievedDocs), latency)

	return retrievedDocs, bestMetrics, nil
}

// evaluateQuality 评估检索和答案质量
func (esr *EnhancedSelfRAG) evaluateQuality(ctx context.Context, query string, docs []string, answer string) (*QualityMetrics, error) {
	if !esr.config.MultiDimensionalEval {
		// 简化版本：单一维度评估
		score, _ := esr.EvaluateRetrieval(ctx, query, docs)
		return &QualityMetrics{
			RelevanceScore: score,
			OverallScore:   score,
		}, nil
	}

	// 多维度评估
	return esr.qualityEvaluator.EvaluateQuality(ctx, query, docs, answer)
}

// generateReflectionWithAdjustments 生成反思和调整建议
func (esr *EnhancedSelfRAG) generateReflectionWithAdjustments(ctx context.Context, query string, metrics *QualityMetrics) (string, *ReflectionAdjustments) {
	// 构建反思提示
	prompt := esr.buildReflectionPrompt(query, metrics)

	response, err := esr.llm.Generate(ctx, prompt)
	if err != nil {
		return "", &ReflectionAdjustments{}
	}

	reflection := strings.TrimSpace(response)

	// 解析调整建议
	adjustments := esr.parseAdjustments(reflection, metrics)

	return reflection, adjustments
}

// ReflectionAdjustments 反思调整建议
type ReflectionAdjustments struct {
	NeedMoreDocs   bool
	ChangeStrategy bool
	NewTopK        int
	NewStrategy    string
}

// buildReflectionPrompt 构建反思提示
func (esr *EnhancedSelfRAG) buildReflectionPrompt(query string, metrics *QualityMetrics) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("请评估以下检索和答案的质量，并提供改进建议。\n\n"))
	sb.WriteString(fmt.Sprintf("查询: %s\n\n", query))
	sb.WriteString(fmt.Sprintf("质量指标:\n"))
	sb.WriteString(fmt.Sprintf("- 相关性: %.2f\n", metrics.RelevanceScore))
	sb.WriteString(fmt.Sprintf("- 覆盖率: %.2f\n", metrics.CoverageScore))
	sb.WriteString(fmt.Sprintf("- 准确性: %.2f\n", metrics.AccuracyScore))
	sb.WriteString(fmt.Sprintf("- 完整性: %.2f\n", metrics.CompletenessScore))
	sb.WriteString(fmt.Sprintf("- 综合得分: %.2f\n\n", metrics.OverallScore))

	if len(metrics.Issues) > 0 {
		sb.WriteString("发现的问题:\n")
		for _, issue := range metrics.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("请提供:\n")
	sb.WriteString("1. 问题诊断\n")
	sb.WriteString("2. 具体改进建议\n")
	sb.WriteString("3. 是否需要更多文档 (yes/no)\n")
	sb.WriteString("4. 是否需要改变检索策略 (yes/no)\n")
	sb.WriteString("5. 建议的 Top-K 值 (如果需要调整)\n\n")
	sb.WriteString("回答 (简明扼要):")

	return sb.String()
}

// parseAdjustments 解析调整建议
func (esr *EnhancedSelfRAG) parseAdjustments(reflection string, metrics *QualityMetrics) *ReflectionAdjustments {
	adjustments := &ReflectionAdjustments{}

	reflectionLower := strings.ToLower(reflection)

	// 解析是否需要更多文档
	if strings.Contains(reflectionLower, "需要更多文档") ||
	   strings.Contains(reflectionLower, "增加检索") ||
	   strings.Contains(reflectionLower, "need more") {
		adjustments.NeedMoreDocs = true
	}

	// 解析是否需要改变策略
	if strings.Contains(reflectionLower, "改变策略") ||
	   strings.Contains(reflectionLower, "change strategy") {
		adjustments.ChangeStrategy = true
	}

	// 解析建议的 Top-K
	if strings.Contains(reflectionLower, "top-k:") ||
	   strings.Contains(reflectionLower, "topk:") {
		parts := strings.Split(reflectionLower, ":")
		if len(parts) > 1 {
			var newTopK int
			fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &newTopK)
			if newTopK > 0 {
				adjustments.NewTopK = newTopK
			}
		}
	}

	return adjustments
}

// calculateDynamicThreshold 计算动态阈值
func (esr *EnhancedSelfRAG) calculateDynamicThreshold(query string, round int) float64 {
	// 基于历史性能调整
	esr.perfTracker.mu.RLock()
	defer esr.perfTracker.mu.RUnlock()

	history, exists := esr.perfTracker.queryHistory[query]
	if !exists || len(history) == 0 {
		return esr.config.MinScore
	}

	// 计算历史平均分数
	sum := 0.0
	for _, perf := range history {
		sum += perf.AnswerScore
	}
	avgScore := sum / float64(len(history))

	// 动态阈值：历史平均值 + 最小改进率
	threshold := avgScore + esr.config.MinImprovementRate

	// 限制在合理范围内
	if threshold > 0.95 {
		threshold = 0.95
	}
	if threshold < esr.config.MinScore {
		threshold = esr.config.MinScore
	}

	return threshold
}

// performInitialRetrieval 执行初始检索
func (esr *EnhancedSelfRAG) performInitialRetrieval(ctx context.Context, query string, topK int) []string {
	// 这里应该调用实际的检索器
	// 简化实现：返回模拟数据
	return []string{
		fmt.Sprintf("检索结果 1 for: %s", query),
		fmt.Sprintf("检索结果 2 for: %s", query),
	}
}

// performAdditionalRetrieval 执行额外检索
func (esr *EnhancedSelfRAG) performAdditionalRetrieval(ctx context.Context, query string, topK int) []string {
	// 简化实现
	return []string{
		fmt.Sprintf("额外检索结果 for: %s", query),
	}
}

// generateTempAnswer 生成临时答案
func (esr *EnhancedSelfRAG) generateTempAnswer(ctx context.Context, query string, docs []string) string {
	prompt := fmt.Sprintf(`基于以下文档回答问题:

问题: %s

文档:
%s

请给出一个简洁的回答:`, query, strings.Join(docs, "\n"))

	response, err := esr.llm.Generate(ctx, prompt)
	if err != nil {
		return "临时答案生成失败"
	}

	return response
}

// recordPerformance 记录性能
func (esr *EnhancedSelfRAG) recordPerformance(query string, score float64, docCount int, latency time.Duration) {
	esr.perfTracker.mu.Lock()
	defer esr.perfTracker.mu.Unlock()

	perf := QueryPerformance{
		Query:          query,
		Timestamp:      time.Now(),
		RetrievalScore: score,
		AnswerScore:    score,
		ReflectionUsed: true,
		Strategy:       "enhanced_self_rag",
		Latency:        latency,
	}

	if _, exists := esr.perfTracker.queryHistory[query]; !exists {
		esr.perfTracker.queryHistory[query] = make([]QueryPerformance, 0)
	}

	esr.perfTracker.queryHistory[query] = append(esr.perfTracker.queryHistory[query], perf)
}

// adjustTopK 调整 Top-K
func (esr *EnhancedSelfRAG) adjustTopK(originalTopK int, score float64) int {
	minK := esr.config.TopKRange[0]
	maxK := esr.config.TopKRange[1]

	if score >= 0.8 {
		// 高质量，减少 Top-K
		adjusted := originalTopK * 2 / 3
		if adjusted < minK {
			adjusted = minK
		}
		return int(adjusted)
	} else if score < 0.5 {
		// 低质量，增加 Top-K
		adjusted := originalTopK * 3 / 2
		if adjusted > maxK {
			adjusted = maxK
		}
		return int(adjusted)
	}

	return originalTopK
}

// GetPerformanceHistory 获取性能历史
func (esr *EnhancedSelfRAG) GetPerformanceHistory(query string) []QueryPerformance {
	esr.perfTracker.mu.RLock()
	defer esr.perfTracker.mu.RUnlock()

	if history, exists := esr.perfTracker.queryHistory[query]; exists {
		return history
	}

	return nil
}

// ===== 默认实现 =====

// DefaultQualityEvaluator 默认质量评估器
type DefaultQualityEvaluator struct {
	llm LLMProvider
}

func (e *DefaultQualityEvaluator) EvaluateQuality(ctx context.Context, query string, retrievedDocs []string, answer string) (*QualityMetrics, error) {
	metrics := &QualityMetrics{
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// 1. 评估相关性
	metrics.RelevanceScore = e.evaluateRelevance(ctx, query, retrievedDocs)

	// 2. 评估覆盖率
	metrics.CoverageScore = e.evaluateCoverage(ctx, query, retrievedDocs)

	// 3. 评估准确性
	metrics.AccuracyScore = e.evaluateAccuracy(ctx, query, answer)

	// 4. 评估完整性
	metrics.CompletenessScore = e.evaluateCompleteness(ctx, query, answer)

	// 5. 计算综合分数
	metrics.OverallScore = (metrics.RelevanceScore*0.3 +
		metrics.CoverageScore*0.2 +
		metrics.AccuracyScore*0.3 +
		metrics.CompletenessScore*0.2)

	// 6. 诊断问题
	if metrics.RelevanceScore < 0.6 {
		metrics.Issues = append(metrics.Issues, "检索相关性不足")
		metrics.Suggestions = append(metrics.Suggestions, "增加检索数量或优化查询")
	}

	if metrics.CoverageScore < 0.6 {
		metrics.Issues = append(metrics.Issues, "信息覆盖不全面")
		metrics.Suggestions = append(metrics.Suggestions, "扩展检索范围")
	}

	if metrics.AccuracyScore < 0.6 {
		metrics.Issues = append(metrics.Issues, "答案准确性不足")
		metrics.Suggestions = append(metrics.Suggestions, "改进检索策略或使用更准确的文档")
	}

	if metrics.CompletenessScore < 0.6 {
		metrics.Issues = append(metrics.Issues, "答案不够完整")
		metrics.Suggestions = append(metrics.Suggestions, "添加更多上下文信息")
	}

	return metrics, nil
}

func (e *DefaultQualityEvaluator) evaluateRelevance(ctx context.Context, query string, docs []string) float64 {
	if len(docs) == 0 {
		return 0.0
	}

	queryWords := extractWords(query)
	totalRelevance := 0.0

	for _, doc := range docs {
		docLower := toLower(doc)
		relevance := 0.0
		matchedWords := make(map[string]bool)

		for _, word := range queryWords {
			wordLower := toLower(word)
			if contains(docLower, wordLower) && !matchedWords[wordLower] {
				matchedWords[wordLower] = true
				relevance += 1.0
			}
		}

		if len(queryWords) > 0 {
			totalRelevance += relevance / float64(len(queryWords))
		}
	}

	return math.Min(totalRelevance/float64(len(docs)), 1.0)
}

func (e *DefaultQualityEvaluator) evaluateCoverage(ctx context.Context, query string, docs []string) float64 {
	// 简化实现：基于文档数量和多样性
	if len(docs) == 0 {
		return 0.0
	}

	// 期望的文档数量
	expectedDocs := 5
	if len(docs) >= expectedDocs {
		return 1.0
	}

	return float64(len(docs)) / float64(expectedDocs)
}

func (e *DefaultQualityEvaluator) evaluateAccuracy(ctx context.Context, query, answer string) float64 {
	// 简化实现：基于答案长度和关键词
	if answer == "" {
		return 0.0
	}

	// 答案应该包含查询的关键词
	queryWords := extractWords(query)
	answerLower := toLower(answer)

	matchedCount := 0
	for _, word := range queryWords {
		if contains(answerLower, toLower(word)) {
			matchedCount++
		}
	}

	if len(queryWords) == 0 {
		return 0.5
	}

	return float64(matchedCount) / float64(len(queryWords))
}

func (e *DefaultQualityEvaluator) evaluateCompleteness(ctx context.Context, query, answer string) float64 {
	// 简化实现：基于答案长度
	answerLen := len([]rune(answer))

	// 期望的最小长度
	minLength := 50
	// 期望的最大长度
	maxLength := 500

	if answerLen < minLength {
		return float64(answerLen) / float64(minLength)
	}

	if answerLen > maxLength {
		return 1.0
	}

	return 1.0
}

// DefaultStrategyAdapter 默认策略适配器
type DefaultStrategyAdapter struct{}

func (a *DefaultStrategyAdapter) AdaptStrategy(ctx context.Context, metrics *QualityMetrics) (string, error) {
	// 基于质量指标调整策略

	if metrics.RelevanceScore < 0.5 {
		// 相关性低，改用语义检索
		return "semantic_search", nil
	}

	if metrics.CoverageScore < 0.5 {
		// 覆盖率低，改用全局检索
		return "global_search", nil
	}

	if metrics.AccuracyScore < 0.6 {
		// 准确性低，改用混合检索
		return "hybrid_search", nil
	}

	// 默认保持当前策略
	return "current_strategy", nil
}
