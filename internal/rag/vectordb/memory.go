package vectordb

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryVectorDB 内存向量数据库
type InMemoryVectorDB struct {
	config    *InMemoryConfig
	vectors   map[string]*VectorData
	mu        sync.RWMutex
	nextID    int
}

// NewInMemoryVectorDB 创建内存向量数据库
func NewInMemoryVectorDB(config *InMemoryConfig) (VectorDB, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	return &InMemoryVectorDB{
		config:  config,
		vectors: make(map[string]*VectorData),
		nextID:  0,
	}, nil
}

// Insert 插入向量
func (m *InMemoryVectorDB) Insert(ctx context.Context, vectors []*VectorData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, vec := range vectors {
		if vec.ID == "" {
			vec.ID = fmt.Sprintf("vec_%d", m.nextID)
			m.nextID++
		}
		m.vectors[vec.ID] = vec
	}

	return nil
}

// Search 搜索向量
func (m *InMemoryVectorDB) Search(ctx context.Context, queryVector []float32, topK int, opts *SearchParams) ([]*VectorSearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 简化实现：直接返回前 topK 个向量
	results := make([]*VectorSearchResult, 0, topK)
	count := 0
	for _, vec := range m.vectors {
		if count >= topK {
			break
		}
		results = append(results, &VectorSearchResult{
			ID:       vec.ID,
			Metadata: vec.Metadata,
			Score:    1.0, // 简化：固定得分
		})
		count++
	}

	return results, nil
}

// Delete 删除向量
func (m *InMemoryVectorDB) Delete(ctx context.Context, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range ids {
		delete(m.vectors, id)
	}

	return nil
}

// Update 更新向量
func (m *InMemoryVectorDB) Update(ctx context.Context, vectors []*VectorData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, vec := range vectors {
		if _, exists := m.vectors[vec.ID]; exists {
			m.vectors[vec.ID] = vec
		}
	}

	return nil
}

// Get 获取向量
func (m *InMemoryVectorDB) Get(ctx context.Context, ids []string) ([]*VectorData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*VectorData, 0, len(ids))
	for _, id := range ids {
		if vec, exists := m.vectors[id]; exists {
			results = append(results, vec)
		}
	}

	return results, nil
}

// CreateCollection 创建集合 (内存数据库不需要)
func (m *InMemoryVectorDB) CreateCollection(ctx context.Context, collection string, dimension int, opts *CollectionOptions) error {
	return nil
}

// DeleteCollection 删除集合 (内存数据库不需要)
func (m *InMemoryVectorDB) DeleteCollection(ctx context.Context, collection string) error {
	return nil
}

// HasCollection 检查集合是否存在 (内存数据库不需要)
func (m *InMemoryVectorDB) HasCollection(ctx context.Context, collection string) (bool, error) {
	return true, nil
}

// ListCollections 列出所有集合 (内存数据库不需要)
func (m *InMemoryVectorDB) ListCollections(ctx context.Context) ([]string, error) {
	return []string{"default"}, nil
}

// Count 计数
func (m *InMemoryVectorDB) Count(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return int64(len(m.vectors)), nil
}

// Clear 清空
func (m *InMemoryVectorDB) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.vectors = make(map[string]*VectorData)
	m.nextID = 0

	return nil
}

// Close 关闭
func (m *InMemoryVectorDB) Close() error {
	return nil
}
