package vectordb

import (
	"context"
	"fmt"
)

// VectorDB 向量数据库接口
// 统一的向量数据库接口，支持多种向量数据库实现
type VectorDB interface {
	// Insert 插入向量
	Insert(ctx context.Context, vectors []*VectorData) error

	// Search 搜索向量
	Search(ctx context.Context, vector []float32, topK int, opts *SearchParams) ([]*VectorSearchResult, error)

	// Delete 删除向量
	Delete(ctx context.Context, ids []string) error

	// Update 更新向量
	Update(ctx context.Context, vectors []*VectorData) error

	// Get 获取向量
	Get(ctx context.Context, ids []string) ([]*VectorData, error)

	// Count 统计向量数量
	Count(ctx context.Context) (int64, error)

	// Clear 清空所有向量
	Clear(ctx context.Context) error

	// CreateCollection 创建集合
	CreateCollection(ctx context.Context, collection string, dimension int, opts *CollectionOptions) error

	// DeleteCollection 删除集合
	DeleteCollection(ctx context.Context, collection string) error

	// HasCollection 检查集合是否存在
	HasCollection(ctx context.Context, collection string) (bool, error)

	// ListCollections 列出所有集合
	ListCollections(ctx context.Context) ([]string, error)

	// Close 关闭连接
	Close() error
}

// VectorData 向量数据
type VectorData struct {
	ID       string                 // 唯一标识
	Vector   []float32              // 向量数据
	Metadata map[string]interface{} // 元数据
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID       string                 // 向量 ID
	Score    float64                // 相似度得分
	Metadata map[string]interface{} // 元数据
}

// SearchParams 搜索参数
type SearchParams struct {
	// MetricType 相似度度量类型 (L2, IP, COSINE)
	MetricType string

	// 过滤表达式 (可选)
	Filter string

	// 额外参数
	Params map[string]interface{}
}

// CollectionOptions 集合选项
type CollectionOptions struct {
	// Description 集合描述
	Description string

	// IndexType 索引类型 (IVF_FLAT, IVF_SQ8, HNSW, etc.)
	IndexType string

	// MetricType 相似度度量
	MetricType string

	// IndexParams 索引参数
	IndexParams map[string]interface{}

	// ConsistencyLevel 一致性级别 (可选)
	ConsistencyLevel string
}

// VectorDBConfig 向量数据库配置
type VectorDBConfig interface {
	// Validate 验证配置
	Validate() error

	// GetType 获取数据库类型
	GetType() string
}

// MilvusConfig Milvus 配置
type MilvusConfig struct {
	Address  string // Milvus 地址
	Username string // 用户名
	Password string // 密码
	Database string // 数据库名
	APIKey   string // API Key (如果使用 Milvus Cloud)
}

// Validate 验证 Milvus 配置
func (c *MilvusConfig) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("milvus address is required")
	}
	return nil
}

// GetType 获取数据库类型
func (c *MilvusConfig) GetType() string {
	return "milvus"
}

// InMemoryConfig 内存数据库配置
type InMemoryConfig struct {
	// MaxVectors 最大向量数量 (0 = 无限制)
	MaxVectors int
}

// Validate 验证内存数据库配置
func (c *InMemoryConfig) Validate() error {
	return nil
}

// GetType 获取数据库类型
func (c *InMemoryConfig) GetType() string {
	return "memory"
}
