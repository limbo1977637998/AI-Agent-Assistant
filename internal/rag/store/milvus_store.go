package store

import (
	"context"
	"fmt"
	"sync"

	"ai-agent-assistant/internal/vectordb"
)

// MilvusVectorStore Milvus向量存储
type MilvusVectorStore struct {
	client       *vectordb.MilvusClient
	collection   string
	ops          *vectordb.VectorOperations
	initialized  bool
	initOnce     sync.Once
	dimension    int
	nextID       int64
	idMutex      sync.Mutex
}

// NewMilvusVectorStore 创建Milvus向量存储
func NewMilvusVectorStore(client *vectordb.MilvusClient, collectionName string, dimension int) *MilvusVectorStore {
	return &MilvusVectorStore{
		client:     client,
		collection: collectionName,
		dimension:  dimension,
		nextID:     1,
	}
}

// initialize 初始化集合
func (s *MilvusVectorStore) initialize(ctx context.Context) error {
	var initErr error
	s.initOnce.Do(func() {
		// 使用CollectionManager创建集合
		manager := vectordb.NewCollectionManager(s.client)

		// 创建简单集合（如果不存在）
		if err := manager.CreateSimpleCollection(ctx, s.collection, s.dimension); err != nil {
			initErr = fmt.Errorf("failed to create collection: %w", err)
			return
		}

		// 创建向量操作实例
		s.ops = vectordb.NewVectorOperations(s.client, s.collection, s.dimension)
		s.initialized = true
	})

	return initErr
}

// Add 添加向量
func (s *MilvusVectorStore) Add(ctx context.Context, vector []float64, text string, metadata map[string]interface{}) error {
	if err := s.initialize(ctx); err != nil {
		return err
	}

	// 将float64转换为float32
	vector32 := make([]float32, len(vector))
	for i, v := range vector {
		vector32[i] = float32(v)
	}

	// 生成ID
	s.idMutex.Lock()
	id := s.nextID
	s.nextID++
	s.idMutex.Unlock()

	// 准备向量数据
	vectorData := &vectordb.VectorData{
		ID:     id,
		Vector: vector32,
		Metadata: map[string]interface{}{
			"content": text,
		},
	}

	// 添加额外的元数据
	if metadata != nil {
		for k, v := range metadata {
			vectorData.Metadata[k] = v
		}
	}

	// 插入向量
	_, err := s.ops.InsertWithMetadata(ctx, []*vectordb.VectorData{vectorData})
	if err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	return nil
}

// Search 搜索最相似的向量
func (s *MilvusVectorStore) Search(ctx context.Context, queryVector []float64, topK int) ([]string, error) {
	if err := s.initialize(ctx); err != nil {
		return nil, err
	}

	// 将float64转换为float32
	vector32 := make([]float32, len(queryVector))
	for i, v := range queryVector {
		vector32[i] = float32(v)
	}

	// 执行搜索
	results, err := s.ops.Search(ctx, vector32, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 提取文本内容
	texts := make([]string, 0, len(results))
	for _, result := range results {
		if result.Metadata != nil {
			if content, ok := result.Metadata["content"].(string); ok {
				texts = append(texts, content)
			}
		}
	}

	return texts, nil
}

// Stats 获取统计信息
func (s *MilvusVectorStore) Stats() map[string]interface{} {
	ctx := context.Background()

	if !s.initialized {
		return map[string]interface{}{
			"type":        "milvus",
			"status":      "not_initialized",
			"vector_count": 0,
			"dimension":   s.dimension,
		}
	}

	count, err := s.ops.Count(ctx)
	if err != nil {
		return map[string]interface{}{
			"type":        "milvus",
			"status":      "error",
			"error":       err.Error(),
			"dimension":   s.dimension,
		}
	}

	return map[string]interface{}{
		"type":         "milvus",
		"collection":   s.collection,
		"vector_count": count,
		"dimension":    s.dimension,
	}
}

// AddBatch 批量添加向量
func (s *MilvusVectorStore) AddBatch(ctx context.Context, vectors []Vector) error {
	if err := s.initialize(ctx); err != nil {
		return err
	}

	if len(vectors) == 0 {
		return nil
	}

	// 准备向量数据
	vectorDataList := make([]*vectordb.VectorData, 0, len(vectors))
	for _, v := range vectors {
		// 将float64转换为float32
		vector32 := make([]float32, len(v.Data))
		for i, val := range v.Data {
			vector32[i] = float32(val)
		}

		s.idMutex.Lock()
		id := s.nextID
		s.nextID++
		s.idMutex.Unlock()

		vectorData := &vectordb.VectorData{
			ID:     id,
			Vector: vector32,
			Metadata: map[string]interface{}{
				"content": v.Text,
			},
		}

		// 添加元数据
		if v.Metadata != nil {
			for k, val := range v.Metadata {
				vectorData.Metadata[k] = val
			}
		}

		vectorDataList = append(vectorDataList, vectorData)
	}

	// 批量插入
	_, err := s.ops.InsertWithMetadata(ctx, vectorDataList)
	if err != nil {
		return fmt.Errorf("failed to insert vectors batch: %w", err)
	}

	return nil
}

// SearchWithMetadata 带元数据的搜索
func (s *MilvusVectorStore) SearchWithMetadata(ctx context.Context, queryVector []float64, topK int) ([]Vector, error) {
	if err := s.initialize(ctx); err != nil {
		return nil, err
	}

	// 将float64转换为float32
	vector32 := make([]float32, len(queryVector))
	for i, v := range queryVector {
		vector32[i] = float32(v)
	}

	// 执行搜索
	results, err := s.ops.Search(ctx, vector32, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 转换为Vector格式
	vectors := make([]Vector, 0, len(results))
	for _, result := range results {
		v := Vector{
			Metadata: make(map[string]interface{}),
		}

		// 提取content到Text
		if result.Metadata != nil {
			if content, ok := result.Metadata["content"].(string); ok {
				v.Text = content
			}
			// 复制其他元数据
			for k, val := range result.Metadata {
				if k != "content" {
					v.Metadata[k] = val
				}
			}
		}

		// 添加相似度分数到元数据
		v.Metadata["score"] = result.Score
		v.Metadata["id"] = result.ID

		vectors = append(vectors, v)
	}

	return vectors, nil
}

// Delete 删除向量
func (s *MilvusVectorStore) Delete(ctx context.Context, ids []int64) error {
	if err := s.initialize(ctx); err != nil {
		return err
	}

	return s.ops.DeleteByID(ctx, ids)
}

// GetCollection 获取集合名称
func (s *MilvusVectorStore) GetCollection() string {
	return s.collection
}

// IsInitialized 检查是否已初始化
func (s *MilvusVectorStore) IsInitialized() bool {
	return s.initialized
}
