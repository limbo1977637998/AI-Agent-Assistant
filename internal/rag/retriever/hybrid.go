package retriever

import (
	"context"
	"fmt"
	"sort"

	"ai-agent-assistant/internal/llm"
)

// HybridRetriever 混合检索器（向量 + BM25）
type HybridRetriever struct {
	vectorRetriever VectorRetriever
	bm25            *BM25
	embeddingModel  llm.Model
	k               int // RRF融合参数
}

// VectorRetriever 向量检索器接口
type VectorRetriever interface {
	Search(ctx context.Context, queryVector []float64, topK int) ([]VectorSearchResult, error)
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	DocID   string
	Content string
	Score   float64
}

// NewHybridRetriever 创建混合检索器
func NewHybridRetriever(vectorRetriever VectorRetriever, embeddingModel llm.Model, k int) *HybridRetriever {
	if k <= 0 {
		k = 60 // 默认k值
	}

	return &HybridRetriever{
		vectorRetriever: vectorRetriever,
		bm25:            NewBM25(1.5, 0.75), // 默认k1=1.5, b=0.75
		embeddingModel:  embeddingModel,
		k:               k,
	}
}

// IndexDocuments 索引文档（用于BM25）
func (hr *HybridRetriever) IndexDocuments(docs []Document) {
	hr.bm25.Index(docs)
}

// Search 混合搜索
func (hr *HybridRetriever) Search(ctx context.Context, query string, topK int) ([]HybridSearchResult, error) {
	// 1. 向量搜索
	queryVector, err := hr.embeddingModel.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	vectorResults, err := hr.vectorRetriever.Search(ctx, queryVector, topK*2) // 获取更多候选
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 2. BM25关键词搜索
	bm25Results := hr.bm25.Search(query, topK * 2)

	// 3. RRF（Reciprocal Rank Fusion）融合
	fusedResults := hr.rrfFusion(vectorResults, bm25Results, topK)

	return fusedResults, nil
}

// rrfFusion RRF融合算法
func (hr *HybridRetriever) rrfFusion(vectorResults []VectorSearchResult, bm25Results []SearchResult, topK int) []HybridSearchResult {
	// 创建文档ID到得分的映射
	scores := make(map[string]float64)
	contentMap := make(map[string]string)

	// 处理向量搜索结果（按排名计算得分）
	for rank, result := range vectorResults {
		score := 1.0 / float64(hr.k+rank+1)
		scores[result.DocID] += score
		contentMap[result.DocID] = result.Content
	}

	// 处理BM25搜索结果（按排名计算得分）
	for rank, result := range bm25Results {
		score := 1.0 / float64(hr.k+rank+1)
		scores[result.DocID] += score
		if _, exists := contentMap[result.DocID]; !exists {
			contentMap[result.DocID] = result.Content
		}
	}

	// 转换为结果列表
	results := make([]HybridSearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, HybridSearchResult{
			DocID:   docID,
			Content: contentMap[docID],
			Score:   score,
		})
	}

	// 按得分降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回topK
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

// HybridSearchResult 混合搜索结果
type HybridSearchResult struct {
	DocID   string
	Content string
	Score   float64
}

// SetBM25Params 设置BM25参数
func (hr *HybridRetriever) SetBM25Params(k1, b float64) {
	hr.bm25.k1 = k1
	hr.bm25.b = b
}

// SetRRFK 设置RRF的k参数
func (hr *HybridRetriever) SetRRFK(k int) {
	hr.k = k
}
