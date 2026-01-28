package vectordb

import (
	"fmt"
)

// VectorDBFactory 向量数据库工厂
type VectorDBFactory struct {
	creators map[string]func(config VectorDBConfig) (VectorDB, error)
}

// NewVectorDBFactory 创建向量数据库工厂
func NewVectorDBFactory() *VectorDBFactory {
	factory := &VectorDBFactory{
		creators: make(map[string]func(config VectorDBConfig) (VectorDB, error)),
	}

	// 注册默认的向量数据库
	factory.RegisterCreator("milvus", NewMilvusDB)
	factory.RegisterCreator("memory", NewInMemoryDB)

	return factory
}

// RegisterCreator 注册向量数据库创建器
func (f *VectorDBFactory) RegisterCreator(dbType string, creator func(config VectorDBConfig) (VectorDB, error)) {
	f.creators[dbType] = creator
}

// Create 创建向量数据库
func (f *VectorDBFactory) Create(config VectorDBConfig) (VectorDB, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	dbType := config.GetType()
	creator, ok := f.creators[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported vector database type: %s", dbType)
	}

	return creator(config)
}

// ListSupportedTypes 列出支持的向量数据库类型
func (f *VectorDBFactory) ListSupportedTypes() []string {
	types := make([]string, 0, len(f.creators))
	for dbType := range f.creators {
		types = append(types, dbType)
	}
	return types
}

// GetDBInfo 获取数据库信息
func (f *VectorDBFactory) GetDBInfo(dbType string) map[string]interface{} {
	info := map[string]interface{}{
		"type": dbType,
	}

	switch dbType {
	case "milvus":
		info["name"] = "Milvus"
		info["description"] = "开源分布式向量数据库"
		info["features"] = []string{"分布式", "高可用", "GPU 加速", "多种索引类型"}
		info["index_types"] = []string{"IVF_FLAT", "IVF_SQ8", "IVF_PQ", "HNSW", "ANNOY"}
		info["metric_types"] = []string{"L2", "IP", "COSINE"}

	case "memory":
		info["name"] = "In-Memory Vector Store"
		info["description"] = "内存向量数据库，适合开发和小规模部署"
		info["features"] = []string{"快速", "简单", "无依赖"}
		info["use_cases"] = []string{"开发测试", "小规模部署", "< 100K 向量"}

	default:
		info["error"] = "Unknown database type"
	}

	return info
}

// NewMilvusDB 创建 Milvus 数据库 (工厂方法)
func NewMilvusDB(config VectorDBConfig) (VectorDB, error) {
	milvusConfig, ok := config.(*MilvusConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for milvus")
	}

	return NewMilvusClient(milvusConfig)
}

// NewInMemoryDB 创建内存数据库 (工厂方法)
func NewInMemoryDB(config VectorDBConfig) (VectorDB, error) {
	memoryConfig, ok := config.(*InMemoryConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for memory")
	}

	return NewInMemoryVectorDB(memoryConfig)
}
