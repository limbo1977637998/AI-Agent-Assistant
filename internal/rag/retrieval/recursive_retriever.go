package retrieval

import (
	"context"
	"fmt"
)

// RecursiveRetriever 递归检索器
//
// 策略说明:
//   递归地在不同粒度的文档层级中检索
//   从小到大或从大到小逐层深入，直到找到满意的结果
//
// 优点:
//   - 平衡精确度和上下文
//   - 自适应检索深度
//   - 支持多粒度文档
//
// 适用场景:
//   - 层次化文档结构
//   - 需要不同粒度信息的场景
//   - 长文档检索
//
// 论文:
//   "Improving Retrieval Performance in RAG with Recursive Retrieval"
type RecursiveRetriever struct {
	childRetriever  Retriever  // 子文档检索器
	parentRetriever Retriever  // 父文档检索器
	config          RecursiveConfig
}

// RecursiveConfig 递归检索配置
type RecursiveConfig struct {
	// StartStrategy 起始策略 (small_to_big, big_to_small)
	StartStrategy string

	// MaxDepth 最大递归深度
	MaxDepth int

	// MinScore 最小质量分数
	MinScore float64

	// ScoreThreshold 分数阈值，低于此值则继续深入
	ScoreThreshold float64

	// TopK 每层检索的文档数
	TopK int

	// EnableAutoMerging 是否启用自动合并
	EnableAutoMerging bool

	// MergeThreshold 合并阈值
	MergeThreshold float64
}

// Retriever 检索器接口
type Retriever interface {
	Retrieve(ctx context.Context, query string, topK int) ([]*RetrievalResult, error)
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	DocID     string
	Content   string
	Score     float64
	Metadata  map[string]interface{}
	ParentID  string // 父文档 ID
	Children  []*RetrievalResult // 子文档
}

// DefaultRecursiveConfig 返回默认配置
func DefaultRecursiveConfig() RecursiveConfig {
	return RecursiveConfig{
		StartStrategy:     "small_to_big",
		MaxDepth:          3,
		MinScore:          0.7,
		ScoreThreshold:    0.6,
		TopK:              5,
		EnableAutoMerging: true,
		MergeThreshold:    0.75,
	}
}

// NewRecursiveRetriever 创建递归检索器
func NewRecursiveRetriever(child, parent Retriever, config RecursiveConfig) (*RecursiveRetriever, error) {
	if child == nil {
		return nil, fmt.Errorf("child retriever is required")
	}

	if config.MaxDepth <= 0 {
		config.MaxDepth = 3
	}
	if config.MinScore <= 0 {
		config.MinScore = 0.7
	}
	if config.ScoreThreshold <= 0 {
		config.ScoreThreshold = 0.6
	}
	if config.TopK <= 0 {
		config.TopK = 5
	}

	return &RecursiveRetriever{
		childRetriever:  child,
		parentRetriever: parent,
		config:          config,
	}, nil
}

// Retrieve 递归检索
func (rr *RecursiveRetriever) Retrieve(ctx context.Context, query string, topK int) ([]*RetrievalResult, error) {
	switch rr.config.StartStrategy {
	case "small_to_big":
		return rr.smallToBigRetrieve(ctx, query, topK, 0)
	case "big_to_small":
		return rr.bigToSmallRetrieve(ctx, query, topK, 0)
	default:
		return rr.smallToBigRetrieve(ctx, query, topK, 0)
	}
}

// smallToBigRetrieve 从小到大检索
// 先检索小的、精确的文档，如果质量不够再检索大的、上下文丰富的文档
func (rr *RecursiveRetriever) smallToBigRetrieve(ctx context.Context, query string, topK, depth int) ([]*RetrievalResult, error) {
	if depth >= rr.config.MaxDepth {
		return nil, nil
	}

	// 1. 从子检索器检索
	childResults, err := rr.childRetriever.Retrieve(ctx, query, rr.config.TopK)
	if err != nil {
		return nil, fmt.Errorf("child retrieval failed: %w", err)
	}

	// 2. 评估检索质量
	avgScore := rr.calculateAverageScore(childResults)

	// 3. 如果质量足够高，直接返回
	if avgScore >= rr.config.MinScore {
		return childResults, nil
	}

	// 4. 质量不够，获取父文档
	parentResults := make([]*RetrievalResult, 0)
	for _, child := range childResults {
		if child.ParentID != "" && rr.parentRetriever != nil {
			// 检索父文档
			parents, err := rr.retrieveParent(ctx, child.ParentID, query)
			if err != nil {
				continue
			}

			parentResults = append(parentResults, parents...)
		}
	}

	// 5. 合并子文档和父文档
	merged := rr.mergeResults(childResults, parentResults)

	// 6. 如果启用了自动合并且父文档质量好
	if rr.config.EnableAutoMerging && len(parentResults) > 0 {
		merged = rr.autoMerge(ctx, query, merged, depth)
	}

	// 7. 递归继续
	if avgScore < rr.config.ScoreThreshold && depth < rr.config.MaxDepth-1 {
		deeperResults, err := rr.smallToBigRetrieve(ctx, query, topK, depth+1)
		if err == nil && len(deeperResults) > 0 {
			merged = append(merged, deeperResults...)
		}
	}

	// 8. 返回 Top-K
	return rr.topKResults(merged, topK), nil
}

// bigToSmallRetrieve 从大到小检索
// 先检索大的、上下文丰富的文档，再深入检索相关的子文档
func (rr *RecursiveRetriever) bigToSmallRetrieve(ctx context.Context, query string, topK, depth int) ([]*RetrievalResult, error) {
	if depth >= rr.config.MaxDepth {
		return nil, nil
	}

	// 1. 如果有父检索器，先从父文档开始
	var parentResults []*RetrievalResult
	if rr.parentRetriever != nil {
		var err error
		parentResults, err = rr.parentRetriever.Retrieve(ctx, query, rr.config.TopK)
		if err != nil {
			return nil, fmt.Errorf("parent retrieval failed: %w", err)
		}
	}

	// 2. 从父文档中检索相关的子文档
	childResults := make([]*RetrievalResult, 0)
	for _, parent := range parentResults {
		// 假设父文档的 Metadata 中包含子文档 ID 列表
		if childrenIDs, ok := parent.Metadata["children_ids"].([]string); ok {
			for _, childID := range childrenIDs {
				// 检索子文档
				children, err := rr.retrieveChildren(ctx, childID, query)
				if err != nil {
					continue
				}

				childResults = append(childResults, children...)
			}
		}
	}

	// 3. 合并结果
	merged := rr.mergeResults(parentResults, childResults)

	// 4. 评估是否需要更深入的检索
	if rr.needDeeperRetrieval(merged) && depth < rr.config.MaxDepth-1 {
		// 继续检索更小的文档
		deeperResults, err := rr.bigToSmallRetrieve(ctx, query, topK, depth+1)
		if err == nil && len(deeperResults) > 0 {
			merged = append(merged, deeperResults...)
		}
	}

	// 5. 返回 Top-K
	return rr.topKResults(merged, topK), nil
}

// retrieveParent 检索父文档
func (rr *RecursiveRetriever) retrieveParent(ctx context.Context, parentID, query string) ([]*RetrievalResult, error) {
	// 简化实现：使用父检索器
	if rr.parentRetriever == nil {
		return nil, nil
	}

	// 实际应该根据 parentID 直接获取文档，而不是检索
	// 这里简化为检索
	results, err := rr.parentRetriever.Retrieve(ctx, query, 1)
	if err != nil {
		return nil, err
	}

	// 过滤出匹配的父文档
	filtered := make([]*RetrievalResult, 0)
	for _, result := range results {
		if result.DocID == parentID {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// retrieveChildren 检索子文档
func (rr *RecursiveRetriever) retrieveChildren(ctx context.Context, childID, query string) ([]*RetrievalResult, error) {
	results, err := rr.childRetriever.Retrieve(ctx, query, 1)
	if err != nil {
		return nil, err
	}

	// 过滤出匹配的子文档
	filtered := make([]*RetrievalResult, 0)
	for _, result := range results {
		if result.DocID == childID {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// mergeResults 合并检索结果
func (rr *RecursiveRetriever) mergeResults(results ...[]*RetrievalResult) []*RetrievalResult {
	merged := make([]*RetrievalResult, 0)

	for _, resultList := range results {
		seen := make(map[string]bool)
		for _, result := range resultList {
			if !seen[result.DocID] {
				merged = append(merged, result)
				seen[result.DocID] = true
			}
		}
	}

	return merged
}

// autoMerge 自动合并文档
func (rr *RecursiveRetriever) autoMerge(ctx context.Context, query string, results []*RetrievalResult, depth int) []*RetrievalResult {
	// 简化实现：基于相似度合并
	// 实际应该使用更复杂的合并策略

	// 如果父文档的分数高于子文档，使用父文档
	// 否则保留子文档
	merged := make([]*RetrievalResult, 0)

	for _, result := range results {
		// 检查是否有子文档
		if len(result.Children) > 0 {
			// 计算子文档的平均分数
			childAvgScore := rr.calculateAverageScore(result.Children)

			// 如果父文档分数显著高于子文档，使用父文档
			if result.Score > childAvgScore*rr.config.MergeThreshold {
				merged = append(merged, result)
			} else {
				// 否则使用子文档
				merged = append(merged, result.Children...)
			}
		} else {
			merged = append(merged, result)
		}
	}

	return merged
}

// needDeeperRetrieval 判断是否需要更深入的检索
func (rr *RecursiveRetriever) needDeeperRetrieval(results []*RetrievalResult) bool {
	if len(results) == 0 {
		return true
	}

	avgScore := rr.calculateAverageScore(results)
	return avgScore < rr.config.ScoreThreshold
}

// calculateAverageScore 计算平均分数
func (rr *RecursiveRetriever) calculateAverageScore(results []*RetrievalResult) float64 {
	if len(results) == 0 {
		return 0
	}

	sum := 0.0
	for _, result := range results {
		sum += result.Score
	}

	return sum / float64(len(results))
}

// topKResults 返回 Top-K 结果
func (rr *RecursiveRetriever) topKResults(results []*RetrievalResult, topK int) []*RetrievalResult {
	// 排序（简单冒泡排序）
	sorted := make([]*RetrievalResult, len(results))
	copy(sorted, results)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 返回 Top-K
	if len(sorted) > topK {
		sorted = sorted[:topK]
	}

	return sorted
}

// Name 返回检索器名称
func (rr *RecursiveRetriever) Name() string {
	return fmt.Sprintf("recursive_retriever_%s", rr.config.StartStrategy)
}
