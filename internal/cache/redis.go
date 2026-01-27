package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string // Redis地址，格式：localhost:6379
	Password string // 密码
	DB       int    // 数据库编号
	PoolSize int    // 连接池大小
}

// DefaultRedisConfig 返回默认配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
}

// RedisClient Redis客户端
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	rc := &RedisClient{
		client: client,
		config: config,
	}

	return rc, nil
}

// Close 关闭Redis客户端
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// GetClient 获取原始Redis客户端
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

// Set 设置键值
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Del 删除键
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (rc *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.client.TTL(ctx, key).Result()
}

// SetJSON 设置JSON值（序列化）
func (rc *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// GetJSON 获取JSON值（反序列化）
func (rc *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return rc.client.Get(ctx, key).Scan(dest)
}

// HSet 设置哈希字段
func (rc *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return rc.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (rc *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return rc.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (rc *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rc.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (rc *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return rc.client.HDel(ctx, key, fields...).Err()
}

// HExists 检查哈希字段是否存在
func (rc *RedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return rc.client.HExists(ctx, key, field).Result()
}

// HIncrBy 哈希字段增加整数值
func (rc *RedisClient) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return rc.client.HIncrBy(ctx, key, field, incr).Result()
}

// SAdd 添加到集合
func (rc *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return rc.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (rc *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return rc.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否是集合成员
func (rc *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return rc.client.SIsMember(ctx, key, member).Result()
}

// SRem 从集合移除
func (rc *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return rc.client.SRem(ctx, key, members...).Err()
}

// ZAdd 添加到有序集合
func (rc *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return rc.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围
func (rc *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围（带分数）
func (rc *RedisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return rc.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem 从有序集合移除
func (rc *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return rc.client.ZRem(ctx, key, members...).Err()
}

// Incr 递增
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

// Decr 递减
func (rc *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return rc.client.Decr(ctx, key).Result()
}

// FlushDB 清空当前数据库
func (rc *RedisClient) FlushDB(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// Keys 获取匹配的键
func (rc *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return rc.client.Keys(ctx, pattern).Result()
}

// Info 获取Redis信息
func (rc *RedisClient) Info(ctx context.Context, section ...string) (string, error) {
	return rc.client.Info(ctx, section...).Result()
}
