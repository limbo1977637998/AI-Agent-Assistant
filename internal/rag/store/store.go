package store

import (
	"context"
	"fmt"
	"sort"

	"ai-agent-assistant/internal/rag/embedding"
)

// Vector 向量数据
type Vector struct {
	Data     []float64
	Text     string
	Metadata map[string]interface{}
}

// VectorStore 向量存储接口
type VectorStore interface {
	Add(ctx context.Context, vector []float64, text string, metadata map[string]interface{}) error
	Search(ctx context.Context, queryVector []float64, topK int) ([]string, error)
	Stats() map[string]interface{}
}

// InMemoryVectorStore 内存向量存储
type InMemoryVectorStore struct {
	vectors   []Vector
	embedding embedding.EmbeddingProvider
}

// NewInMemoryVectorStore 创建内存向量存储
func NewInMemoryVectorStore(ep embedding.EmbeddingProvider) *InMemoryVectorStore {
	return &InMemoryVectorStore{
		vectors:   make([]Vector, 0),
		embedding: ep,
	}
}

// Add 添加向量
func (s *InMemoryVectorStore) Add(ctx context.Context, vector []float64, text string, metadata map[string]interface{}) error {
	s.vectors = append(s.vectors, Vector{
		Data:     vector,
		Text:     text,
		Metadata: metadata,
	})
	return nil
}

// Search 搜索最相似的向量
func (s *InMemoryVectorStore) Search(ctx context.Context, queryVector []float64, topK int) ([]string, error) {
	if len(s.vectors) == 0 {
		return []string{}, nil
	}

	// 计算所有向量的相似度
	type Result struct {
		Text       string
		Similarity float64
	}

	results := make([]Result, 0, len(s.vectors))
	for _, v := range s.vectors {
		sim := embedding.CosineSimilarity(queryVector, v.Data)
		results = append(results, Result{
			Text:       v.Text,
			Similarity: sim,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// 返回topK个结果
	if topK > len(results) {
		topK = len(results)
	}

	texts := make([]string, 0, topK)
	for i := 0; i < topK; i++ {
		// 过滤相似度太低的结果
		if results[i].Similarity > 0.3 {
			texts = append(texts, results[i].Text)
		}
	}

	return texts, nil
}

// Stats 获取统计信息
func (s *InMemoryVectorStore) Stats() map[string]interface{} {
	return map[string]interface{}{
		"type":        "memory",
		"vector_count": len(s.vectors),
		"dimension":   s.embedding.GetDimension(),
	}
}

// DeleteAll 清空所有向量
func (s *InMemoryVectorStore) DeleteAll() {
	s.vectors = make([]Vector, 0)
}

// GetVectors 获取所有向量（用于调试）
func (s *InMemoryVectorStore) GetVectors() []Vector {
	return s.vectors
}

// AddBatch 批量添加向量
func (s *InMemoryVectorStore) AddBatch(ctx context.Context, vectors []Vector) error {
	s.vectors = append(s.vectors, vectors...)
	return nil
}

// SearchWithMetadata 带元数据的搜索
func (s *InMemoryVectorStore) SearchWithMetadata(ctx context.Context, queryVector []float64, topK int) ([]Vector, error) {
	if len(s.vectors) == 0 {
		return []Vector{}, nil
	}

	type Result struct {
		Vector     Vector
		Similarity float64
	}

	results := make([]Result, 0, len(s.vectors))
	for _, v := range s.vectors {
		sim := embedding.CosineSimilarity(queryVector, v.Data)
		results = append(results, Result{
			Vector:     v,
			Similarity: sim,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// 返回topK个结果
	if topK > len(results) {
		topK = len(results)
	}

	vectors := make([]Vector, 0, topK)
	for i := 0; i < topK; i++ {
		// 过滤相似度太低的结果
		if results[i].Similarity > 0.3 {
			vectors = append(vectors, results[i].Vector)
		}
	}

	return vectors, nil
}

// FilterByMetadata 根据元数据过滤向量
func (s *InMemoryVectorStore) FilterByMetadata(key string, value interface{}) []Vector {
	filtered := make([]Vector, 0)
	for _, v := range s.vectors {
		if val, ok := v.Metadata[key]; ok && val == value {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// UpdateMetadata 更新元数据
func (s *InMemoryVectorStore) UpdateMetadata(index int, metadata map[string]interface{}) error {
	if index < 0 || index >= len(s.vectors) {
		return fmt.Errorf("index out of bounds")
	}
	s.vectors[index].Metadata = metadata
	return nil
}

// GetTotalCount 获取向量总数
func (s *InMemoryVectorStore) GetTotalCount() int {
	return len(s.vectors)
}
