package reranker

import (
	"context"
	"fmt"
	"strings"
)

// LLMReranker LLM 重排序器
//
// 策略说明:
//   使用 LLM 对检索结果进行重排序
//   通过理解查询和文档的语义关系进行评分
//
// 优点:
//   - 深度语义理解
//   - 上下文感知
//   - 处理复杂查询
//
// 适用场景:
//   - 复杂查询
//   - 需要深度理解的场景
//   - 对准确性要求高的场景
type LLMReranker struct {
	llm        LLMProvider
	topK       int
	batchSize  int
	templating string
}

// LLMProvider LLM 提供者接口
type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// NewLLMReranker 创建 LLM 重排序器
func NewLLMReranker(llm LLMProvider, topK int) (*LLMReranker, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	if topK <= 0 {
		topK = 5
	}

	return &LLMReranker{
		llm:       llm,
		topK:      topK,
		batchSize: 5, // 每次 LLM 调用处理的文档数
	}, nil
}

// Rerank 重排序文档
func (r *LLMReranker) Rerank(ctx context.Context, query string, documents []Document) ([]Document, error) {
	if len(documents) == 0 {
		return documents, nil
	}

	// 如果文档数量 <= topK，直接返回
	if len(documents) <= r.topK {
		return documents, nil
	}

	// 批量处理文档
	scores := make([]float64, len(documents))

	for i := 0; i < len(documents); i += r.batchSize {
		end := i + r.batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		batchScores, err := r.scoreBatch(ctx, query, batch)
		if err != nil {
			// 如果评分失败，使用原始分数
			for j := i; j < end; j++ {
				scores[j] = documents[j].Score
			}
			continue
		}

		for j, score := range batchScores {
			scores[i+j] = score
		}
	}

	// 更新分数并排序
	for i := range documents {
		documents[i].Score = scores[i]
	}

	// 按分数降序排序
	sortedDocs := make([]Document, len(documents))
	copy(sortedDocs, documents)

	// 简单冒泡排序（实际应使用更高效的排序）
	for i := 0; i < len(sortedDocs)-1; i++ {
		for j := i + 1; j < len(sortedDocs); j++ {
			if sortedDocs[j].Score > sortedDocs[i].Score {
				sortedDocs[i], sortedDocs[j] = sortedDocs[j], sortedDocs[i]
			}
		}
	}

	// 返回 Top-K
	if len(sortedDocs) > r.topK {
		sortedDocs = sortedDocs[:r.topK]
	}

	return sortedDocs, nil
}

// scoreBatch 批量评分
func (r *LLMReranker) scoreBatch(ctx context.Context, query string, documents []Document) ([]float64, error) {
	prompt := r.buildScoringPrompt(query, documents)

	response, err := r.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	scores := r.parseScores(response, len(documents))

	return scores, nil
}

// buildScoringPrompt 构建评分提示
func (r *LLMReranker) buildScoringPrompt(query string, documents []Document) string {
	var sb strings.Builder

	sb.WriteString("请为以下文档相对于查询的相关性进行评分。\n\n")
	sb.WriteString(fmt.Sprintf("查询: %s\n\n", query))
	sb.WriteString("文档列表:\n")

	for i, doc := range documents {
		// 限制文档长度
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, content))
	}

	sb.WriteString("\n评分要求:\n")
	sb.WriteString("1. 为每个文档打分 (0-10分)\n")
	sb.WriteString("2. 10分表示高度相关，0分表示完全不相关\n")
	sb.WriteString("3. 按顺序输出分数，用逗号分隔\n")
	sb.WriteString("4. 只输出数字，不要其他内容\n\n")
	sb.WriteString("输出格式: 8,7,9,6,5\n")

	return sb.String()
}

// parseScores 解析分数
func (r *LLMReranker) parseScores(response string, expectedCount int) []float64 {
	// 简化实现：按逗号分割
	parts := strings.Split(strings.TrimSpace(response), ",")

	scores := make([]float64, expectedCount)
	for i := 0; i < expectedCount; i++ {
		if i < len(parts) {
			var score float64
			fmt.Sscanf(strings.TrimSpace(parts[i]), "%f", &score)
			// 归一化到 [0, 1]
			scores[i] = score / 10.0
		} else {
			scores[i] = 0.5 // 默认分数
		}
	}

	return scores
}

// Name 返回重排序器名称
func (r *LLMReranker) Name() string {
	return "llm_reranker"
}
