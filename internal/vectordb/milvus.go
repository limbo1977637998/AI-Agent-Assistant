package vectordb

import (
	"context"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// MilvusConfig Milvus配置
type MilvusConfig struct {
	Address        string        // Milvus地址，格式：localhost:19530
	Username       string        // 用户名
	Password       string        // 密码
	Database       string        // 数据库名称
	ConnectTimeout time.Duration // 连接超时时间
}

// DefaultMilvusConfig 返回默认配置
func DefaultMilvusConfig() *MilvusConfig {
	return &MilvusConfig{
		Address:        "localhost:19530",
		Username:       "",
		Password:       "",
		Database:       "default",
		ConnectTimeout: 10 * time.Second,
	}
}

// MilvusClient Milvus客户端
type MilvusClient struct {
	client client.Client
	config *MilvusConfig
}

// NewMilvusClient 创建Milvus客户端
func NewMilvusClient(config *MilvusConfig) (*MilvusClient, error) {
	if config == nil {
		config = DefaultMilvusConfig()
	}

	// 构建Milvus客户端配置
	clientCfg := client.Config{
		Address:  config.Address,
		Username: config.Username,
		Password: config.Password,
	}

	// 连接到Milvus
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	milvusClient, err := client.NewClient(ctx, clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %w", err)
	}

	mc := &MilvusClient{
		client: milvusClient,
		config: config,
	}

	// 测试连接
	if err := mc.Ping(ctx); err != nil {
		milvusClient.Close()
		return nil, fmt.Errorf("failed to ping milvus: %w", err)
	}

	return mc, nil
}

// Ping 测试Milvus连接
func (mc *MilvusClient) Ping(ctx context.Context) error {
	// 使用CheckHealth来测试连接
	health, err := mc.client.CheckHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}

	if !health.IsHealthy {
		return fmt.Errorf("milvus is not healthy")
	}

	return nil
}

// Close 关闭Milvus客户端
func (mc *MilvusClient) Close() error {
	return mc.client.Close()
}

// GetClient 获取原始Milvus客户端
func (mc *MilvusClient) GetClient() client.Client {
	return mc.client
}

// ListCollections 列出所有集合
func (mc *MilvusClient) ListCollections(ctx context.Context) ([]string, error) {
	collections, err := mc.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	names := make([]string, 0, len(collections))
	for _, coll := range collections {
		names = append(names, coll.Name)
	}

	return names, nil
}

// HasCollection 检查集合是否存在
func (mc *MilvusClient) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	has, err := mc.client.HasCollection(ctx, collectionName)
	if err != nil {
		return false, fmt.Errorf("failed to check collection: %w", err)
	}
	return has, nil
}

// DropCollection 删除集合
func (mc *MilvusClient) DropCollection(ctx context.Context, collectionName string) error {
	return mc.client.DropCollection(ctx, collectionName)
}

// CreateIndex 为集合创建索引
func (mc *MilvusClient) CreateIndex(ctx context.Context, collectionName string, fieldName string, index entity.Index) error {
	return mc.client.CreateIndex(ctx, collectionName, fieldName, index, false)
}

// GetIndex 获取集合的索引信息
func (mc *MilvusClient) GetIndex(ctx context.Context, collectionName string, fieldName string) ([]entity.Index, error) {
	indexes, err := mc.client.DescribeIndex(ctx, collectionName, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}
	return indexes, nil
}

// LoadCollection 加载集合到内存
func (mc *MilvusClient) LoadCollection(ctx context.Context, collectionName string) error {
	return mc.client.LoadCollection(ctx, collectionName, false)
}

// ReleaseCollection 释放集合内存
func (mc *MilvusClient) ReleaseCollection(ctx context.Context, collectionName string) error {
	return mc.client.ReleaseCollection(ctx, collectionName)
}
