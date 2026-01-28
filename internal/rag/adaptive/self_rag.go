package adaptive

import (
	"context"
	"fmt"
)

// SelfRAGStrategy Self-RAG 策略接口
//
// Self-RAG (Self-Reflective Retrieval-Augmented Generation)
// 通过自我反思机制动态调整检索策略
//
// 论文: "Self-RAG: Learning to Retrieve, Generate, and Critique through Self-Reflection"
//
// 核心思想:
//   1. 检索后自我评估检索质量
//   2. 根据评估结果生成反思
//   3. 决定是否需要重新检索或调整策略
type SelfRAGStrategy interface {
	// EvaluateRetrieval 评估检索质量
	// 返回 0-1 的分数，表示检索质量
	EvaluateRetrieval(ctx context.Context, query string, retrievedDocs []string) (float64, error)

	// GenerateReflection 生成反思
	// 根据评估结果生成反思文本
	GenerateReflection(ctx context.Context, query string, score float64, retrievedDocs []string) (string, error)

	// ShouldRetrieveMore 是否需要检索更多
	// 根据评估分数决定是否需要检索更多文档
	ShouldRetrieveMore(score float64) bool

	// AdjustTopK 调整 Top-K 参数
	// 根据评估分数动态调整 Top-K
	AdjustTopK(originalTopK int, score float64) int
}

// AdaptiveRAG 自适应 RAG 接口
type AdaptiveRAG interface {
	// SelectStrategy 选择最优检索策略
	// 根据查询特征选择最合适的检索策略
	SelectStrategy(ctx context.Context, query string) (string, error)

	// OptimizeParameters 优化检索参数
	// 根据历史性能优化检索参数
	OptimizeParameters(ctx context.Context, queryType string) (map[string]interface{}, error)

	// RecordFeedback 记录反馈
	// 记录查询和结果，用于学习优化
	RecordFeedback(ctx context.Context, query string, strategy string, result *RAGExecutionResult) error
}

// RAGExecutionResult RAG 执行结果
type RAGExecutionResult struct {
	Strategy      string   // 使用的策略
	Query         string   // 查询
	Answer        string   // 生成的答案
	Contexts      []string // 检索到的上下文
	Score         float64  // 质量得分
	Latency       int64    // 延迟 (毫秒)
	UserFeedback float64  // 用户反馈 (0-1)
	Success       bool     // 是否成功
}

// SelfReflectiveRAG 自我反思 RAG 实现
type SelfReflectiveRAG struct {
	strategy SelfRAGStrategy
	llm      LLMProvider
	config   SelfRAGConfig
}

// SelfRAGConfig Self-RAG 配置
type SelfRAGConfig struct {
	// MinScore 最小质量分数
	MinScore float64

	// MaxRetries 最大重试次数
	MaxRetries int

	// TopKRange Top-K 范围 [min, max]
	TopKRange [2]int

	// EnableReflection 是否启用反思
	EnableReflection bool
}

// DefaultSelfRAGConfig 返回默认配置
func DefaultSelfRAGConfig() SelfRAGConfig {
	return SelfRAGConfig{
		MinScore:         0.7,
		MaxRetries:       2,
		TopKRange:        [2]int{5, 20},
		EnableReflection: true,
	}
}

// NewSelfReflectiveRAG 创建自我反思 RAG
func NewSelfReflectiveRAG(llm LLMProvider, config SelfRAGConfig) (*SelfReflectiveRAG, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	// 设置默认值
	if config.MinScore <= 0 {
		config.MinScore = 0.7
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 2
	}
	if config.TopKRange[0] <= 0 {
		config.TopKRange = [2]int{5, 20}
	}
	if config.TopKRange[1] <= config.TopKRange[0] {
		config.TopKRange[1] = config.TopKRange[0] + 15
	}

	return &SelfReflectiveRAG{
		llm:    llm,
		config: config,
	}, nil
}

// EvaluateRetrieval 评估检索质量
func (sr *SelfReflectiveRAG) EvaluateRetrieval(ctx context.Context, query string, retrievedDocs []string) (float64, error) {
	if len(retrievedDocs) == 0 {
		return 0.0, nil
	}

	// 简单实现：基于查询词覆盖率
	queryWords := extractWords(query)
	totalRelevance := 0.0

	for _, doc := range retrievedDocs {
		docLower := toLower(doc)
		relevance := 0.0
		matchedWords := make(map[string]bool)

		for _, word := range queryWords {
			wordLower := toLower(word)
			if contains(docLower, wordLower) {
				if !matchedWords[wordLower] {
					matchedWords[wordLower] = true
					relevance += 1.0
				}
			}
		}

		// 归一化
		if len(queryWords) > 0 {
			totalRelevance += relevance / float64(len(queryWords))
		}
	}

	// 平均得分
	score := totalRelevance / float64(len(retrievedDocs))

	// 限制在 [0, 1] 范围内
	if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// GenerateReflection 生成反思
func (sr *SelfReflectiveRAG) GenerateReflection(ctx context.Context, query string, score float64, retrievedDocs []string) (string, error) {
	if !sr.config.EnableReflection {
		return "", nil
	}

	prompt := fmt.Sprintf(`请评估以下检索结果的质量，并生成改进建议。

查询: %s
检索质量得分: %.2f (0-1分)

检索到的文档:
%s

要求:
1. 分析检索结果是否与查询相关
2. 如果得分较低，指出问题所在
3. 提供具体的改进建议
4. 简明扼要，不超过100字

反思:`, query, score, formatDocuments(retrievedDocs))

	response, err := sr.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

// ShouldRetrieveMore 是否需要检索更多
func (sr *SelfReflectiveRAG) ShouldRetrieveMore(score float64) bool {
	return score < sr.config.MinScore
}

// AdjustTopK 调整 Top-K 参数
func (sr *SelfReflectiveRAG) AdjustTopK(originalTopK int, score float64) int {
	minK := sr.config.TopKRange[0]
	maxK := sr.config.TopKRange[1]

	// 根据得分线性调整
	// 得分越低，Top-K 越大
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

	// 中等质量，保持不变
	return originalTopK
}

// formatDocuments 格式化文档列表
func formatDocuments(docs []string) string {
	if len(docs) == 0 {
		return "(无文档)"
	}

	formatted := ""
	for i, doc := range docs {
		// 限制每个文档显示长度
		maxLen := 100
		if len(doc) > maxLen {
			doc = doc[:maxLen] + "..."
		}
		formatted += fmt.Sprintf("%d. %s\n", i+1, doc)
	}

	return formatted
}

// extractWords 提取单词
func extractWords(text string) []string {
	// 简单实现：按空格分割
	words := make([]string, 0)
	currentWord := ""

	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(r)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// toLower 转换为小写
func toLower(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

// indexOf 查找子串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if len(substr) == 0 {
			return i
		}
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// LLMProvider LLM 提供者接口
type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
