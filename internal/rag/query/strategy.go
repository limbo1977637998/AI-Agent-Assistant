package query

import (
	"context"
)

// QueryOptimizerStrategy 查询优化策略接口
// 所有查询优化器都需要实现此接口
type QueryOptimizerStrategy interface {
	// Optimize 优化查询
	// 参数:
	//   ctx: 上下文
	//   query: 原始查询
	// 返回:
	//   []QueryOptimization: 优化后的查询列表
	//   error: 错误信息
	Optimize(ctx context.Context, query string) ([]QueryOptimization, error)

	// Name 返回优化器名称
	Name() string

	// Validate 验证优化器配置
	Validate() error
}

// QueryOptimization 优化后的查询
type QueryOptimization struct {
	// Query 优化后的查询文本
	Query string `json:"query"`

	// Type 优化类型 (rewrite, decompose, expand, hyde)
	Type string `json:"type"`

	// Score 相关性得分或置信度 (可选)
	Score float64 `json:"score,omitempty"`

	// Metadata 额外元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// QueryOptimizerConfig 查询优化器配置
type QueryOptimizerConfig struct {
	// MaxQueries 最大查询数量
	MaxQueries int `json:"max_queries"`

	// MinConfidence 最小置信度阈值
	MinConfidence float64 `json:"min_confidence"`

	// EnableDecomposition 是否启用查询分解
	EnableDecomposition bool `json:"enable_decomposition"`

	// EnableExpansion 是否启用查询扩展
	EnableExpansion bool `json:"enable_expansion"`

	// EnableRewriting 是否启用查询重写
	EnableRewriting bool `json:"enable_rewriting"`

	// Language 语言 (zh, en, etc.)
	Language string `json:"language"`
}

// DefaultQueryOptimizerConfig 返回默认配置
func DefaultQueryOptimizerConfig() QueryOptimizerConfig {
	return QueryOptimizerConfig{
		MaxQueries:          5,
		MinConfidence:       0.6,
		EnableDecomposition: true,
		EnableExpansion:     true,
		EnableRewriting:     true,
		Language:            "zh",
	}
}
