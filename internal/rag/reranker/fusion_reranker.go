package reranker

import (
	"context"
	"fmt"
	"math"
)

// FusionReranker 融合重排序器
//
// 策略说明:
//   融合多个检索结果并进行重排序
//   支持多种融合算法：RRF, Weighted Sum, Borda Count
//
// 优点:
//   - 结合多种检索源的优势
//   - 提高召回和准确率
//   - 降低单一方法的偏差
//
// 适用场景:
//   - 多路召回场景
//   - 需要高准确率场景
//   - 混合检索场景
type FusionReranker struct {
	method    string      // 融合方法: rrf, weighted, borda
	weights   []float64   // 各路权重
	topK      int
	normalize bool        // 是否归一化分数
}

// FusionResult 融合结果
type FusionResult struct {
	Document Document
	Scores   []float64 // 各路分数
	FusedScore float64 // 融合后的分数
	Rank     int       // 融合后的排名
}

// NewFusionReranker 创建融合重排序器
func NewFusionReranker(method string, weights []float64, topK int) (*FusionReranker, error) {
	if len(weights) == 0 {
		weights = []float64{0.5, 0.5} // 默认等权重
	}

	// 验证权重
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	if math.Abs(sum-1.0) > 0.01 {
		return nil, fmt.Errorf("weights must sum to 1.0, got %f", sum)
	}

	if topK <= 0 {
		topK = 10
	}

	return &FusionReranker{
		method:    method,
		weights:   weights,
		topK:      topK,
		normalize: true,
	}, nil
}

// Rerank 重排序并融合多个检索结果
func (f *FusionReranker) Rerank(ctx context.Context, query string, documents []Document) ([]Document, error) {
	if len(documents) == 0 {
		return documents, nil
	}

	// 单路检索，直接返回
	if f.method == "single" {
		return documents, nil
	}

	// 这里简化实现，实际应该接收多路结果
	// 假设 documents 已经包含多路信息（通过元数据区分）

	// 使用 RRF 算法
	if f.method == "rrf" || f.method == "" {
		return f.rrfFusion(documents), nil
	}

	// 使用加权平均
	if f.method == "weighted" {
		return f.weightedFusion(documents), nil
	}

	// 使用 Borda Count
	if f.method == "borda" {
		return f.bordaCountFusion(documents), nil
	}

	return documents, nil
}

// rrfFusion Reciprocal Rank Fusion 融合
// 论文: "Reciprocal Rank Fusion outperforms condorcet and individual Rank Learning Methods"
// 公式: score(d) = Σ(weight_i / (k + rank_i))
// 其中 k 是常数（通常为 60）
func (f *FusionReranker) rrfFusion(documents []Document) []Document {
	k := 60.0 // RRF 常数

	// 创建文档 ID 到分数的映射
	scoreMap := make(map[string]float64)

	// 假设文档按来源分组（简化实现）
	// 实际应该接收多路结果: [][]Document

	// 为每个文档计算 RRF 分数
	for i, doc := range documents {
		rank := float64(i + 1)
		rrfScore := 1.0 / (k + rank)

		// 累加分数
		scoreMap[doc.ID] += rrfScore
	}

	// 更新文档分数
	results := make([]Document, len(documents))
	copy(results, documents)

	for i := range results {
		if score, ok := scoreMap[results[i].ID]; ok {
			results[i].Score = score
		}
	}

	// 排序并返回 Top-K
	return f.topKResults(results)
}

// weightedFusion 加权平均融合
func (f *FusionReranker) weightedFusion(documents []Document) []Document {
	// 简化实现：使用原始分数和权重
	results := make([]Document, len(documents))
	copy(results, documents)

	// 归一化分数
	if f.normalize {
		maxScore := 0.0
		for _, doc := range results {
			if doc.Score > maxScore {
				maxScore = doc.Score
			}
		}

		if maxScore > 0 {
			for i := range results {
				results[i].Score = results[i].Score / maxScore
			}
		}
	}

	// 应用权重（简化：单路检索权重为 1.0）
	for i := range results {
		results[i].Score = results[i].Score * f.weights[0]
	}

	return f.topKResults(results)
}

// bordaCountFusion Borda Count 融合
// 投票法：每个检索方法为文档投票，排名越前得分越高
func (f *FusionReranker) bordaCountFusion(documents []Document) []Document {
	n := len(documents)

	// Borda 分数：排名第 i 得 (n-i) 分
	scoreMap := make(map[string]float64)

	for i, doc := range documents {
		bordaScore := float64(n - i)
		scoreMap[doc.ID] += bordaScore
	}

	// 更新分数
	results := make([]Document, len(documents))
	copy(results, documents)

	for i := range results {
		if score, ok := scoreMap[results[i].ID]; ok {
			results[i].Score = score
		}
	}

	// 归一化
	maxScore := float64(n * n) // 最大可能分数
	for i := range results {
		results[i].Score = results[i].Score / maxScore
	}

	return f.topKResults(results)
}

// topKResults 返回 Top-K 结果
func (f *FusionReranker) topKResults(documents []Document) []Document {
	// 排序
	sorted := make([]Document, len(documents))
	copy(sorted, documents)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 返回 Top-K
	if len(sorted) > f.topK {
		sorted = sorted[:f.topK]
	}

	return sorted
}

// Name 返回重排序器名称
func (f *FusionReranker) Name() string {
	return fmt.Sprintf("fusion_reranker_%s", f.method)
}
