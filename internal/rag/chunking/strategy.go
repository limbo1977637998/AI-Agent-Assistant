package chunking

import (
	"context"
)

// ChunkerStrategy 分块器策略接口
// 所有分块器实现都需要实现此接口
type ChunkerStrategy interface {
	// Split 将文本分割成多个块
	// 参数:
	//   ctx: 上下文
	//   text: 待分块的文本
	// 返回:
	//   []Chunk: 分块结果
	//   error: 错误信息
	Split(ctx context.Context, text string) ([]Chunk, error)

	// Name 返回分块器名称
	Name() string

	// Validate 验证分块器配置是否有效
	Validate() error
}

// Chunk 文本分块
// 包含文本内容和元数据
type Chunk struct {
	// Content 分块内容
	Content string `json:"content"`

	// Metadata 分块元数据
	Metadata ChunkMetadata `json:"metadata"`
}

// ChunkMetadata 分块元数据
type ChunkMetadata struct {
	// Index 分块索引 (从0开始)
	Index int `json:"index"`

	// StartPos 在原文档中的起始位置
	StartPos int `json:"start_pos,omitempty"`

	// EndPos 在原文档中的结束位置
	EndPos int `json:"end_pos,omitempty"`

	// TokenCount Token 数量 (估计值)
	TokenCount int `json:"token_count,omitempty"`

	// ChunkType 分块类型 (fixed, semantic, recursive, parent, etc.)
	ChunkType string `json:"chunk_type"`

	// ParentChunkIndex 父分块索引 (用于 ParentDocumentChunker)
	ParentChunkIndex int `json:"parent_chunk_index,omitempty"`

	// Score 相关性得分或质量得分 (可选)
	Score float64 `json:"score,omitempty"`

	// AdditionalMetadata 额外的元数据字段
	AdditionalMetadata map[string]interface{} `json:"additional_metadata,omitempty"`
}

// ChunkerConfig 分块器配置 (通用)
type ChunkerConfig struct {
	// ChunkSize 分块大小 (字符数或 token 数)
	ChunkSize int `json:"chunk_size"`

	// ChunkOverlap 分块重叠大小
	ChunkOverlap int `json:"chunk_overlap"`

	// MinChunkSize 最小分块大小 (某些策略使用)
	MinChunkSize int `json:"min_chunk_size,omitempty"`

	// Separator 分隔符 (某些策略使用)
	Separators []string `json:"separators,omitempty"`

	// KeepSeparator 是否保留分隔符
	KeepSeparator bool `json:"keep_separator,omitempty"`
}

// DefaultChunkerConfig 返回默认配置
func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		ChunkSize:     500,
		ChunkOverlap: 50,
		MinChunkSize:  50,
		Separators:    []string{"\n\n", "\n", "。", "!", "?", ".", " ", ""},
		KeepSeparator: false,
	}
}
