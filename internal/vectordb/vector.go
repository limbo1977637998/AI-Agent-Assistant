package vectordb

import (
	"context"
	"fmt"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// VectorData 向量数据
type VectorData struct {
	ID       int64                  // 向量ID
	Vector   []float32              // 向量数据
	Metadata map[string]interface{} // 元数据
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       int64                  // 结果ID
	Score    float32                // 相似度分数
	Vector   []float32              // 向量数据
	Metadata map[string]interface{} // 元数据
}

// VectorOperations 向量操作接口
type VectorOperations struct {
	client    *MilvusClient
	collection string // 集合名称
	dimension  int    // 向量维度
}

// NewVectorOperations 创建向量操作实例
func NewVectorOperations(client *MilvusClient, collectionName string, dimension int) *VectorOperations {
	return &VectorOperations{
		client:     client,
		collection: collectionName,
		dimension:  dimension,
	}
}

// Insert 插入向量数据
func (vo *VectorOperations) Insert(ctx context.Context, vectors []*VectorData) (int64, error) {
	if len(vectors) == 0 {
		return 0, fmt.Errorf("no vectors to insert")
	}

	// 准备插入数据
	ids := make([]int64, 0, len(vectors))
	vectorData := make([][]float32, 0, len(vectors))
	contents := make([]string, 0, len(vectors))

	for _, v := range vectors {
		ids = append(ids, v.ID)
		vectorData = append(vectorData, v.Vector)
		if v.Metadata != nil {
			if content, ok := v.Metadata["content"].(string); ok {
				contents = append(contents, content)
			} else {
				contents = append(contents, "")
			}
		} else {
			contents = append(contents, "")
		}
	}

	// 构建列式数据
	idColumn := entity.NewColumnInt64("id", ids)
	vectorColumn := entity.NewColumnFloatVector("vector", vo.dimension, vectorData)
	contentColumn := entity.NewColumnVarChar("content", contents)

	// 插入数据
	_, err := vo.client.GetClient().Insert(ctx, vo.collection, "", idColumn, vectorColumn, contentColumn)
	if err != nil {
		return 0, fmt.Errorf("failed to insert vectors: %w", err)
	}

	// 刷新以确保数据持久化
	if err := vo.client.GetClient().Flush(ctx, vo.collection, false); err != nil {
		log.Printf("Warning: failed to flush collection: %v", err)
	}

	return int64(len(ids)), nil
}

// InsertWithMetadata 插入带元数据的向量
func (vo *VectorOperations) InsertWithMetadata(ctx context.Context, vectors []*VectorData) (int64, error) {
	return vo.Insert(ctx, vectors)
}

// Search 向量搜索
func (vo *VectorOperations) Search(ctx context.Context, queryVector []float32, topK int) ([]*SearchResult, error) {
	// 构建搜索向量
	vectors := []entity.Vector{entity.FloatVector(queryVector)}

	// 创建搜索参数 - 使用HNSW索引参数
	sp, err := entity.NewIndexHNSWSearchParam(64) // ef参数
	if err != nil {
		return nil, fmt.Errorf("failed to create search param: %w", err)
	}

	// 执行搜索
	searchResult, err := vo.client.GetClient().Search(
		ctx,
		vo.collection,
		[]string{}, // 分区
		"",        // 表达式
		[]string{"id", "content"},
		vectors,
		"vector",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 解析结果
	results := make([]*SearchResult, 0)
	for _, res := range searchResult {
		for i := 0; i < res.ResultCount; i++ {
			id := res.IDs.(*entity.ColumnInt64).Data()[i]
			score := res.Scores[i]

			result := &SearchResult{
				ID:       id,
				Score:    score,
				Metadata: make(map[string]interface{}),
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SearchWithFilter 带过滤条件的向量搜索
func (vo *VectorOperations) SearchWithFilter(ctx context.Context, queryVector []float32, topK int, filter string) ([]*SearchResult, error) {
	// 构建搜索向量
	vectors := []entity.Vector{entity.FloatVector(queryVector)}

	// 创建搜索参数 - 使用HNSW索引参数
	sp, err := entity.NewIndexHNSWSearchParam(64) // ef参数
	if err != nil {
		return nil, fmt.Errorf("failed to create search param: %w", err)
	}

	// 执行搜索
	searchResult, err := vo.client.GetClient().Search(
		ctx,
		vo.collection,
		[]string{}, // 分区
		filter,    // 过滤表达式
		[]string{"id", "content"},
		vectors,
		"vector",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors with filter: %w", err)
	}

	// 解析结果
	results := make([]*SearchResult, 0)
	for _, res := range searchResult {
		for i := 0; i < res.ResultCount; i++ {
			id := res.IDs.(*entity.ColumnInt64).Data()[i]
			score := res.Scores[i]

			result := &SearchResult{
				ID:       id,
				Score:    score,
				Metadata: make(map[string]interface{}),
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// DeleteByID 根据ID删除向量
func (vo *VectorOperations) DeleteByID(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("no ids to delete")
	}

	// 构建删除表达式
	expr := fmt.Sprintf("id in [%v]", ids)

	// Milvus使用表达式删除
	err := vo.client.GetClient().Delete(ctx, vo.collection, expr, "")
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	return nil
}

// GetByID 根据ID获取向量
func (vo *VectorOperations) GetByID(ctx context.Context, ids []int64) ([]*VectorData, error) {
	// 查询向量 - Query返回[]entity.Column
	columns, err := vo.client.GetClient().Query(
		ctx,
		vo.collection,
		[]string{}, // 分区
		fmt.Sprintf("id in %v", ids), // 查询表达式
		[]string{"id", "vector", "content"},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	// 解析结果
	vectors := make([]*VectorData, 0)

	// columns是一个slice，包含请求的各个字段
	// 第一个字段是id
	if len(columns) > 0 {
		if idCol, ok := columns[0].(*entity.ColumnInt64); ok {
			for i := 0; i < idCol.Len(); i++ {
				vectorData := &VectorData{
					ID:       idCol.Data()[i],
					Metadata: make(map[string]interface{}),
				}

				// 如果有content字段
				if len(columns) > 2 {
					if contentCol, ok := columns[2].(*entity.ColumnVarChar); ok && i < contentCol.Len() {
						vectorData.Metadata["content"] = contentCol.Data()[i]
					}
				}

				vectors = append(vectors, vectorData)
			}
		}
	}

	return vectors, nil
}

// Count 统计向量数量
func (vo *VectorOperations) Count(ctx context.Context) (int64, error) {
	stats, err := vo.client.GetClient().GetCollectionStatistics(ctx, vo.collection)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection statistics: %w", err)
	}

	// stats返回的是map[string]string类型
	if rowCount, ok := stats["row_count"]; ok {
		var count int64
		fmt.Sscanf(rowCount, "%d", &count)
		return count, nil
	}

	return 0, fmt.Errorf("failed to get row count from statistics")
}
