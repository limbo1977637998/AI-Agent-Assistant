package eval

import (
	"context"
	"fmt"
	"strings"
)

// RAGASEvaluator RAGAS 评估器
//
// RAGAS (Retrieval Augmented Generation Assessment) 是一个评估 RAG 系统的框架
//
// 评估指标:
//   1. Context Precision: 上下文精确度（检索到的上下文的相关性）
//   2. Context Recall: 上下文召回率（是否检索到所有相关信息）
//   3. Answer Relevancy: 答案相关性（生成的答案与查询的相关程度）
//   4. Faithfulness: 忠实度（答案是否基于检索到的上下文）
//   5. Context Entity Recall: 上下文实体召回率
//
// 参考资料:
//   https://docs.ragas.io/
//   "RAGAS: Automated Evaluation of Retrieval Augmented Generation" (2023)
type RAGASEvaluator struct {
	llm LLMProvider // LLM 提供者（用于评估）
}

// NewRAGASEvaluator 创建 RAGAS 评估器
func NewRAGASEvaluator(llm LLMProvider) (*RAGASEvaluator, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	return &RAGASEvaluator{
		llm: llm,
	}, nil
}

// RAGASResult RAGAS 评估结果
type RAGASResult struct {
	// ContextPrecision 上下文精确度 (0-1)
	ContextPrecision float64 `json:"context_precision"`

	// ContextRecall 上下文召回率 (0-1)
	ContextRecall float64 `json:"context_recall"`

	// AnswerRelevancy 答案相关性 (0-1)
	AnswerRelevancy float64 `json:"answer_relevancy"`

	// Faithfulness 忠实度 (0-1)
	Faithfulness float64 `json:"faithfulness"`

	// OverallScore 总体得分 (0-1)
	OverallScore float64 `json:"overall_score"`

	// Details 详细信息
	Details map[string]interface{} `json:"details,omitempty"`
}

// Evaluate 评估 RAG 系统
//
// 参数:
//   ctx: 上下文
//   query: 用户查询
//   contexts: 检索到的上下文列表
//   answer: 生成的答案
//   groundTruth: 真实答案（可选，用于计算召回率）
//
// 返回:
//   *RAGASResult: 评估结果
//   error: 错误信息
func (evaluator *RAGASEvaluator) Evaluate(
	ctx context.Context,
	query string,
	contexts []string,
	answer string,
	groundTruth string,
) (*RAGASResult, error) {
	result := &RAGASResult{
		Details: make(map[string]interface{}),
	}

	// 1. 计算 Context Precision
	precision, err := evaluator.evaluateContextPrecision(ctx, query, contexts)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate context precision: %w", err)
	}
	result.ContextPrecision = precision

	// 2. 计算 Context Recall
	recall, err := evaluator.evaluateContextRecall(ctx, query, contexts, groundTruth)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate context recall: %w", err)
	}
	result.ContextRecall = recall

	// 3. 计算 Answer Relevancy
	relevancy, err := evaluator.evaluateAnswerRelevancy(ctx, query, answer)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answer relevancy: %w", err)
	}
	result.AnswerRelevancy = relevancy

	// 4. 计算 Faithfulness
	faithfulness, err := evaluator.evaluateFaithfulness(ctx, contexts, answer)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate faithfulness: %w", err)
	}
	result.Faithfulness = faithfulness

	// 5. 计算总体得分（加权平均）
	result.OverallScore = (
		result.ContextPrecision*0.25 +
			result.ContextRecall*0.25 +
			result.AnswerRelevancy*0.25 +
			result.Faithfulness*0.25)

	return result, nil
}

// evaluateContextPrecision 评估上下文精确度
// 问题: 检索到的上下文是否与查询相关？
func (evaluator *RAGASEvaluator) evaluateContextPrecision(
	ctx context.Context,
	query string,
	contexts []string,
) (float64, error) {
	if len(contexts) == 0 {
		return 0.0, nil
	}

	// 简单实现：计算查询词在上下文中的出现频率
	queryWords := strings.Fields(strings.ToLower(query))
	totalRelevance := 0.0

	for _, context := range contexts {
		contextLower := strings.ToLower(context)
		relevantWords := 0
		for _, word := range queryWords {
			if strings.Contains(contextLower, word) {
				relevantWords++
			}
		}
		totalRelevance += float64(relevantWords) / float64(len(queryWords))
	}

	return totalRelevance / float64(len(contexts)), nil
}

// evaluateContextRecall 评估上下文召回率
// 问题: 是否检索到了所有相关信息？
func (evaluator *RAGASEvaluator) evaluateContextRecall(
	ctx context.Context,
	query string,
	contexts []string,
	groundTruth string,
) (float64, error) {
	if groundTruth == "" {
		// 如果没有真实答案，使用查询作为参考
		groundTruth = query
	}

	// 简单实现：计算真实答案中的关键信息在上下文中的覆盖率
	truthWords := strings.Fields(strings.ToLower(groundTruth))
	foundWords := make(map[string]bool)

	for _, context := range contexts {
		contextLower := strings.ToLower(context)
		for _, word := range truthWords {
			if strings.Contains(contextLower, word) {
				foundWords[word] = true
			}
		}
	}

	if len(truthWords) == 0 {
		return 0.0, nil
	}

	return float64(len(foundWords)) / float64(len(truthWords)), nil
}

// evaluateAnswerRelevancy 评估答案相关性
// 问题: 生成的答案与查询是否相关？
func (evaluator *RAGASEvaluator) evaluateAnswerRelevancy(
	ctx context.Context,
	query string,
	answer string,
) (float64, error) {
	if answer == "" {
		return 0.0, nil
	}

	// 简单实现：计算查询词在答案中的覆盖率
	queryWords := strings.Fields(strings.ToLower(query))
	answerLower := strings.ToLower(answer)

	coveredWords := 0
	for _, word := range queryWords {
		if strings.Contains(answerLower, word) {
			coveredWords++
		}
	}

	if len(queryWords) == 0 {
		return 0.0, nil
	}

	return float64(coveredWords) / float64(len(queryWords)), nil
}

// evaluateFaithfulness 评估忠实度
// 问题: 答案是否基于检索到的上下文？
func (evaluator *RAGASEvaluator) evaluateFaithfulness(
	ctx context.Context,
	contexts []string,
	answer string,
) (float64, error) {
	if answer == "" {
		return 0.0, nil
	}

	if len(contexts) == 0 {
		return 0.0, nil
	}

	// 简单实现：检查答案中的关键陈述是否在上下文中得到支持
	answerWords := strings.Fields(strings.ToLower(answer))
	supportedStatements := 0

	for _, word := range answerWords {
		for _, context := range contexts {
			contextLower := strings.ToLower(context)
			if strings.Contains(contextLower, word) {
				supportedStatements++
				break
			}
		}
	}

	if len(answerWords) == 0 {
		return 0.0, nil
	}

	return float64(supportedStatements) / float64(len(answerWords)), nil
}

// EvaluateBatch 批量评估
func (evaluator *RAGASEvaluator) EvaluateBatch(
	ctx context.Context,
	queries []string,
	contexts [][]string,
	answers []string,
	groundTruths []string,
) ([]*RAGASResult, error) {
	if len(queries) != len(contexts) ||
		len(queries) != len(answers) ||
		len(queries) != len(groundTruths) {
		return nil, fmt.Errorf("input arrays must have the same length")
	}

	results := make([]*RAGASResult, len(queries))

	for i := range queries {
		result, err := evaluator.Evaluate(
			ctx,
			queries[i],
			contexts[i],
			answers[i],
			groundTruths[i],
		)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate query %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// GenerateReport 生成评估报告
func (evaluator *RAGASEvaluator) GenerateReport(results []*RAGASResult) string {
	var report strings.Builder

	report.WriteString("# RAGAS 评估报告\n\n")
	report.WriteString(fmt.Sprintf("评估样本数: %d\n\n", len(results)))

	// 计算平均得分
	avgPrecision := 0.0
	avgRecall := 0.0
	avgRelevancy := 0.0
	avgFaithfulness := 0.0
	avgOverall := 0.0

	for _, result := range results {
		avgPrecision += result.ContextPrecision
		avgRecall += result.ContextRecall
		avgRelevancy += result.AnswerRelevancy
		avgFaithfulness += result.Faithfulness
		avgOverall += result.OverallScore
	}

	if len(results) > 0 {
		avgPrecision /= float64(len(results))
		avgRecall /= float64(len(results))
		avgRelevancy /= float64(len(results))
		avgFaithfulness /= float64(len(results))
		avgOverall /= float64(len(results))
	}

	report.WriteString("## 平均得分\n\n")
	report.WriteString(fmt.Sprintf("- Context Precision: %.2f%%\n", avgPrecision*100))
	report.WriteString(fmt.Sprintf("- Context Recall: %.2f%%\n", avgRecall*100))
	report.WriteString(fmt.Sprintf("- Answer Relevancy: %.2f%%\n", avgRelevancy*100))
	report.WriteString(fmt.Sprintf("- Faithfulness: %.2f%%\n", avgFaithfulness*100))
	report.WriteString(fmt.Sprintf("- Overall Score: %.2f%%\n\n", avgOverall*100))

	// 生成建议
	report.WriteString("## 改进建议\n\n")

	if avgPrecision < 0.7 {
		report.WriteString("- ⚠️ Context Precision 偏低，建议改进检索策略或使用更好的向量化模型\n")
	}
	if avgRecall < 0.7 {
		report.WriteString("- ⚠️ Context Recall 偏低，建议增加检索数量或使用查询扩展\n")
	}
	if avgRelevancy < 0.7 {
		report.WriteString("- ⚠️ Answer Relevancy 偏低，建议优化提示词或使用更好的生成模型\n")
	}
	if avgFaithfulness < 0.7 {
		report.WriteString("- ⚠️ Faithfulness 偏低，建议确保答案基于检索到的上下文\n")
	}

	if avgPrecision >= 0.7 && avgRecall >= 0.7 && avgRelevancy >= 0.7 && avgFaithfulness >= 0.7 {
		report.WriteString("- ✅ 所有指标表现良好！\n")
	}

	return report.String()
}

// LLMProvider LLM 提供者接口（用于评估）
type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
