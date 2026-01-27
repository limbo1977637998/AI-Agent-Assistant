package vectordb

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// CollectionManager 集合管理器
type CollectionManager struct {
	client *MilvusClient
}

// NewCollectionManager 创建集合管理器
func NewCollectionManager(client *MilvusClient) *CollectionManager {
	return &CollectionManager{
		client: client,
	}
}

// CreateSimpleCollection 创建简单的向量集合
func (cm *CollectionManager) CreateSimpleCollection(ctx context.Context, collectionName string, dimension int) error {
	// 检查集合是否已存在
	has, err := cm.client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}
	if has {
		return nil
	}

	// 定义schema - 只包含基本字段
	fields := []*entity.Field{
		{
			Name:       "id",
			DataType:   entity.FieldTypeInt64,
			PrimaryKey: true,
			AutoID:     false,
		},
		{
			Name:     "content",
			DataType: entity.FieldTypeVarChar,
			TypeParams: map[string]string{
				"max_length": "65535",
			},
		},
		{
			Name:       "vector",
			DataType:   entity.FieldTypeFloatVector,
			TypeParams: map[string]string{
				"dim": fmt.Sprintf("%d", dimension),
			},
		},
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Fields:         fields,
	}

	// 创建集合
	err = cm.client.GetClient().CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// 创建索引
	idx, _ := entity.NewIndexIvfFlat(entity.L2, 128)
	err = cm.client.CreateIndex(ctx, collectionName, "vector", idx)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 加载集合到内存
	err = cm.client.LoadCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	return nil
}

// DropCollection 删除集合
func (cm *CollectionManager) DropCollection(ctx context.Context, collectionName string) error {
	return cm.client.DropCollection(ctx, collectionName)
}

// GetCollectionInfo 获取集合信息
func (cm *CollectionManager) GetCollectionInfo(ctx context.Context, collectionName string) (*entity.Collection, error) {
	return cm.client.GetClient().DescribeCollection(ctx, collectionName)
}

// LoadCollection 加载集合到内存
func (cm *CollectionManager) LoadCollection(ctx context.Context, collectionName string) error {
	return cm.client.LoadCollection(ctx, collectionName)
}

// ReleaseCollection 释放集合内存
func (cm *CollectionManager) ReleaseCollection(ctx context.Context, collectionName string) error {
	return cm.client.ReleaseCollection(ctx, collectionName)
}

// GetCollectionStats 获取集合统计信息
func (cm *CollectionManager) GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	stats, err := cm.client.GetClient().GetCollectionStatistics(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	// 转换为map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range stats {
		result[k] = v
	}

	return result, nil
}
