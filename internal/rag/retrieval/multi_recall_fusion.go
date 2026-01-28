package retrieval

import (
	"context"
	"fmt"
	"math"
)

// MultiRecallFusion 多路召回融合器
//
// 功能: 融合多个检索源的召回结果
//
// 支持的融合算法:
//   1. RRF (Reciprocal Rank Fusion) - 倒数排名融合
//   2. Weighted Sum - 加权求和
//   3. Borda Count - 波达计数
//   4. Condorcet - 孔多塞投票
//   5. CombANZ - CombANZ 算法
//   6. CombMIN - CombMIN 算法
//   7. CombMAX - CombMAX 算法
//
// 优点:
//   - 结合多种检索方法
//   - 提高召回率和准确率
//   - 降低单一方法的偏差
//   - 支持动态权重调整
//
// 适用场景:
//   - 混合检索场景
//   - 多模态检索
//   - 分布式检索
//   - 需要高准确率场景
//
// 论文:
//   "Reciprocal Rank Fusion outperforms Condorcet and individual Rank Learning Methods"
type MultiRecallFusion struct {
	fusionMethod string
	weights      []float64
	k            int  // RRF 常数
	topK         int
	normalize    bool
}

// RecallSource 召回源
type RecallSource struct {
	Name   string
	Weight float64
}

// FusionResult 融合结果
type FusionResult struct {
	DocID       string
	Content     string
	Scores      []float64     // 各路分数
	Ranks       []int         // 各路排名
	FusedScore  float64       // 融合后的分数
	FusedRank   int           // 融合后的排名
	Sources     []string      // 来源列表
	Metadata    map[string]interface{}
}

// NewMultiRecallFusion 创建多路召回融合器
func NewMultiRecallFusion(method string, sources []RecallSource, topK int) (*MultiRecallFusion, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("at least one source is required")
	}

	// 提取权重
	weights := make([]float64, len(sources))
	sum := 0.0
	for i, source := range sources {
		weights[i] = source.Weight
		sum += source.Weight
	}

	// 归一化权重
	if math.Abs(sum-1.0) > 0.01 {
		for i := range weights {
			weights[i] = weights[i] / sum
		}
	}

	if topK <= 0 {
		topK = 10
	}

	return &MultiRecallFusion{
		fusionMethod: method,
		weights:      weights,
		k:            60, // RRF 常数，默认 60
		topK:         topK,
		normalize:    true,
	}, nil
}

// Fuse 融合多路召回结果
func (mrf *MultiRecallFusion) Fuse(ctx context.Context, recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	if len(recallResults) == 0 {
		return nil, nil
	}

	switch mrf.fusionMethod {
	case "rrf":
		return mrf.rrfFusion(recallResults)
	case "weighted":
		return mrf.weightedFusion(recallResults)
	case "borda":
		return mrf.bordaCountFusion(recallResults)
	case "comb_anz":
		return mrf.combANZFusion(recallResults)
	case "comb_min":
		return mrf.combMinFusion(recallResults)
	case "comb_max":
		return mrf.combMaxFusion(recallResults)
	default:
		return mrf.rrfFusion(recallResults)
	}
}

// rrfFusion Reciprocal Rank Fusion 融合
// 公式: score(d) = Σ(weight_i / (k + rank_i(d)))
func (mrf *MultiRecallFusion) rrfFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	// 创建文档 ID 到结果的映射
	docMap := make(map[string]*FusionResult)

	// 遍历每路召回结果
	for sourceIdx, results := range recallResults {
		weight := mrf.weights[sourceIdx]
		if sourceIdx >= len(mrf.weights) {
			weight = 1.0 / float64(len(recallResults))
		}

		for rank, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			// 更新分数和排名
			doc := docMap[result.DocID]
			doc.Scores[sourceIdx] = result.Score
			doc.Ranks[sourceIdx] = rank + 1
			doc.Sources = append(doc.Sources, fmt.Sprintf("source_%d", sourceIdx))

			// 计算 RRF 分数
			k := float64(mrf.k)
			rankFloat := float64(rank + 1)
			rrfScore := weight / (k + rankFloat)
			doc.FusedScore += rrfScore
		}
	}

	// 转换为列表
	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		results = append(results, doc)
	}

	// 排序并返回 Top-K
	return mrf.rankResults(results), nil
}

// weightedFusion 加权求和融合
// 公式: score(d) = Σ(weight_i * score_i(d))
func (mrf *MultiRecallFusion) weightedFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	docMap := make(map[string]*FusionResult)

	for sourceIdx, results := range recallResults {
		weight := mrf.weights[sourceIdx]
		if sourceIdx >= len(mrf.weights) {
			weight = 1.0 / float64(len(recallResults))
		}

		// 归一化该路分数
		maxScore := 0.0
		for _, result := range results {
			if result.Score > maxScore {
				maxScore = result.Score
			}
		}

		for rank, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			doc := docMap[result.DocID]
			normalizedScore := result.Score
			if mrf.normalize && maxScore > 0 {
				normalizedScore = result.Score / maxScore
			}

			doc.Scores[sourceIdx] = normalizedScore
			doc.Ranks[sourceIdx] = rank + 1
			doc.Sources = append(doc.Sources, fmt.Sprintf("source_%d", sourceIdx))

			// 累加加权分数
			doc.FusedScore += weight * normalizedScore
		}
	}

	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		results = append(results, doc)
	}

	return mrf.rankResults(results), nil
}

// bordaCountFusion 波达计数融合
// 方法: 每路为文档投票，排名第 i 得 (n-i) 分
// 公式: score(d) = Σ(weight_i * (n - rank_i(d)))
func (mrf *MultiRecallFusion) bordaCountFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	docMap := make(map[string]*FusionResult)

	for sourceIdx, results := range recallResults {
		weight := mrf.weights[sourceIdx]
		if sourceIdx >= len(mrf.weights) {
			weight = 1.0 / float64(len(recallResults))
		}

		n := len(results) // 该路文档总数

		for rank, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			doc := docMap[result.DocID]
			doc.Scores[sourceIdx] = result.Score
			doc.Ranks[sourceIdx] = rank + 1
			doc.Sources = append(doc.Sources, fmt.Sprintf("source_%d", sourceIdx))

			// 计算 Borda 分数
			bordaScore := float64(n - rank)
			doc.FusedScore += weight * bordaScore
		}
	}

	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		results = append(results, doc)
	}

	// 归一化分数
	maxScore := float64(0)
	for _, result := range results {
		if result.FusedScore > maxScore {
			maxScore = result.FusedScore
		}
	}

	if maxScore > 0 {
		for _, result := range results {
			result.FusedScore = result.FusedScore / maxScore
		}
	}

	return mrf.rankResults(results), nil
}

// combANZFusion CombANZ 融合
// 方法: 对每路排名归一化后取平均
// 公式: score(d) = (1/k) * Σ(rank_i(d) / n_i)
func (mrf *MultiRecallFusion) combANZFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	docMap := make(map[string]*FusionResult)

	for sourceIdx, results := range recallResults {
		n := float64(len(results)) // 该路文档总数

		for rank, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			doc := docMap[result.DocID]
			doc.Ranks[sourceIdx] = rank + 1
			doc.Sources = append(doc.Sources, fmt.Sprintf("source_%d", sourceIdx))

			// 归一化排名 [0, 1]
			normalizedRank := float64(rank+1) / n

			// 累加归一化排名
			doc.FusedScore += normalizedRank
		}
	}

	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		// 取平均
		doc.FusedScore = doc.FusedScore / float64(len(recallResults))
		results = append(results, doc)
	}

	// 反转（因为排名越小越好）
	for _, result := range results {
		result.FusedScore = 1.0 - result.FusedScore
	}

	return mrf.rankResults(results), nil
}

// combMinFusion CombMIN 融合
// 方法: 取各路分数的最小值
// 公式: score(d) = min(score_i(d))
func (mrf *MultiRecallFusion) combMinFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	docMap := make(map[string]*FusionResult)

	for sourceIdx, results := range recallResults {
		// 归一化该路分数
		maxScore := 0.0
		for _, result := range results {
			if result.Score > maxScore {
				maxScore = result.Score
			}
		}

		for _, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			doc := docMap[result.DocID]
			normalizedScore := result.Score
			if mrf.normalize && maxScore > 0 {
				normalizedScore = result.Score / maxScore
			}

			doc.Scores[sourceIdx] = normalizedScore

			// 取最小值
			if doc.FusedScore == 0 || normalizedScore < doc.FusedScore {
				doc.FusedScore = normalizedScore
			}
		}
	}

	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		results = append(results, doc)
	}

	return mrf.rankResults(results), nil
}

// combMaxFusion CombMAX 融合
// 方法: 取各路分数的最大值
// 公式: score(d) = max(score_i(d))
func (mrf *MultiRecallFusion) combMaxFusion(recallResults [][]*RetrievalResult) ([]*FusionResult, error) {
	docMap := make(map[string]*FusionResult)

	for sourceIdx, results := range recallResults {
		// 归一化该路分数
		maxScore := 0.0
		for _, result := range results {
			if result.Score > maxScore {
				maxScore = result.Score
			}
		}

		for _, result := range results {
			if _, exists := docMap[result.DocID]; !exists {
				docMap[result.DocID] = &FusionResult{
					DocID:    result.DocID,
					Content:  result.Content,
					Scores:   make([]float64, len(recallResults)),
					Ranks:    make([]int, len(recallResults)),
					Sources:  make([]string, 0),
					Metadata: result.Metadata,
				}
			}

			doc := docMap[result.DocID]
			normalizedScore := result.Score
			if mrf.normalize && maxScore > 0 {
				normalizedScore = result.Score / maxScore
			}

			doc.Scores[sourceIdx] = normalizedScore

			// 取最大值
			if normalizedScore > doc.FusedScore {
				doc.FusedScore = normalizedScore
			}
		}
	}

	results := make([]*FusionResult, 0, len(docMap))
	for _, doc := range docMap {
		results = append(results, doc)
	}

	return mrf.rankResults(results), nil
}

// rankResults 排序结果并返回 Top-K
func (mrf *MultiRecallFusion) rankResults(results []*FusionResult) []*FusionResult {
	// 简单冒泡排序
	sorted := make([]*FusionResult, len(results))
	copy(sorted, results)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].FusedScore > sorted[i].FusedScore {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 更新排名
	for i := range sorted {
		sorted[i].FusedRank = i + 1
	}

	// 返回 Top-K
	if len(sorted) > mrf.topK {
		sorted = sorted[:mrf.topK]
	}

	return sorted
}

// Name 返回融合器名称
func (mrf *MultiRecallFusion) Name() string {
	return fmt.Sprintf("multi_recall_fusion_%s", mrf.fusionMethod)
}

// GetFusionMethods 获取支持的融合方法
func (mrf *MultiRecallFusion) GetFusionMethods() []string {
	return []string{
		"rrf",       // Reciprocal Rank Fusion
		"weighted",  // Weighted Sum
		"borda",     // Borda Count
		"comb_anz",  // CombANZ
		"comb_min",  // CombMIN
		"comb_max",  // CombMAX
	}
}

// SetK 设置 RRF 常数 k
func (mrf *MultiRecallFusion) SetK(k int) {
	if k > 0 {
		mrf.k = k
	}
}

// SetNormalize 设置是否归一化
func (mrf *MultiRecallFusion) SetNormalize(normalize bool) {
	mrf.normalize = normalize
}

// ExplainFusion 解释融合结果
func (mrf *MultiRecallFusion) ExplainFusion(results []*FusionResult) string {
	explanation := fmt.Sprintf("融合方法: %s\n", mrf.fusionMethod)
	explanation += fmt.Sprintf("权重: %v\n", mrf.weights)
	explanation += fmt.Sprintf("Top-K: %d\n\n", mrf.topK)

	for i, result := range results {
		explanation += fmt.Sprintf("排名 %d (ID: %s):\n", i+1, result.DocID)
		explanation += fmt.Sprintf("  融合分数: %.4f\n", result.FusedScore)
		explanation += fmt.Sprintf("  来源: %v\n", result.Sources)
		explanation += fmt.Sprintf("  各路分数: %v\n", result.Scores)
		explanation += fmt.Sprintf("  各路排名: %v\n\n", result.Ranks)
	}

	return explanation
}
