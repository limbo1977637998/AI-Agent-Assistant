package chunking

import (
	"context"
	"fmt"
	"strings"
)

// SmallToBigChunker 小到大分块器
//
// 策略说明:
//   1. 先用小块 (small chunk) 分割文档 (提高检索精度)
//   2. 将多个小块合并成大块 (parent chunk) 用于索引
//   3. 检索时返回小块，但可以追溯到父块获取更多上下文
//
// 优点:
//   - 检索精度高 (小块更精准匹配查询)
//   - 上下文丰富 (可以追溯到父块)
//   - 平衡精度和召回率
//
// 适用场景:
//   - 需要高精度检索的场景
//   - 长文档理解
//   - 问答系统
type SmallToBigChunker struct {
	smallConfig  ChunkerConfig // 小块配置
	bigConfig    ChunkerConfig // 大块配置
	name         string
	parentMerge  int           // 多少个小块合并成一个大块
}

// NewSmallToBigChunker 创建小到大分块器
//
// 参数:
//   smallConfig: 小块配置 (用于检索)
//   bigConfig: 大块配置 (用于索引和上下文)
//   parentMerge: 多少个小块合并成一个大块 (默认 3)
//
// 返回:
//   *SmallToBigChunker: 分块器实例
//   error: 配置验证错误
func NewSmallToBigChunker(smallConfig, bigConfig ChunkerConfig, parentMerge int) (*SmallToBigChunker, error) {
	// 验证小块配置
	if smallConfig.ChunkSize <= 0 {
		return nil, fmt.Errorf("small chunk_size must be positive")
	}
	if smallConfig.ChunkOverlap < 0 {
		return nil, fmt.Errorf("small chunk_overlap cannot be negative")
	}

	// 验证大块配置
	if bigConfig.ChunkSize <= 0 {
		return nil, fmt.Errorf("big chunk_size must be positive")
	}
	if bigConfig.ChunkOverlap < 0 {
		return nil, fmt.Errorf("big chunk_overlap cannot be negative")
	}

	// 大块必须大于小块
	if bigConfig.ChunkSize <= smallConfig.ChunkSize {
		return nil, fmt.Errorf("big chunk_size must be greater than small chunk_size")
	}

	// 设置默认 parentMerge
	if parentMerge <= 0 {
		parentMerge = 3 // 默认 3 个小块合并成 1 个大块
	}

	// 确保 parentMerge 合理
	if bigConfig.ChunkSize < smallConfig.ChunkSize*parentMerge {
		// 调整 parentMerge 以避免大块过小
		parentMerge = bigConfig.ChunkSize / smallConfig.ChunkSize
		if parentMerge < 2 {
			parentMerge = 2
		}
	}

	return &SmallToBigChunker{
		smallConfig: smallConfig,
		bigConfig:   bigConfig,
		name:        "small_to_big",
		parentMerge: parentMerge,
	}, nil
}

// Split 实现分块逻辑
// 返回的是小块 (small chunks)，每个小块记录其所属的父块索引
func (stb *SmallToBigChunker) Split(ctx context.Context, text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	// 1. 先用小块分割
	smallChunks := stb.createSmallChunks(text)
	if len(smallChunks) == 0 {
		return []Chunk{}, nil
	}

	// 2. 将小块合并成大块
	parentChunks := stb.createParentChunks(smallChunks)

	// 3. 为每个小块添加父块引用
	result := make([]Chunk, len(smallChunks))
	position := 0
	for i, smallChunk := range smallChunks {
		// 找到这个小块属于哪个父块
		parentIndex := i / stb.parentMerge
		if parentIndex >= len(parentChunks) {
			parentIndex = len(parentChunks) - 1
		}

		startPos := position
		endPos := position + len(smallChunk)
		position = endPos - stb.smallConfig.ChunkOverlap

		result[i] = Chunk{
			Content: smallChunk,
			Metadata: ChunkMetadata{
				Index:            i,
				StartPos:         startPos,
				EndPos:           endPos,
				ChunkType:        stb.name + "_small",
				TokenCount:       estimateTokens(smallChunk),
				ParentChunkIndex: parentIndex,
				AdditionalMetadata: map[string]interface{}{
					"parent_content":    parentChunks[parentIndex], // 存储父块内容
					"is_small_chunk":    true,
					"chunk_size_type":   "small",
				},
			},
		}
	}

	return result, nil
}

// createSmallChunks 创建小块
func (stb *SmallToBigChunker) createSmallChunks(text string) []string {
	// 使用递归字符分块器创建小块
	recursiveConfig := ChunkerConfig{
		ChunkSize:     stb.smallConfig.ChunkSize,
		ChunkOverlap:  stb.smallConfig.ChunkOverlap,
		MinChunkSize:  stb.smallConfig.MinChunkSize,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, _ := NewRecursiveCharacterChunker(recursiveConfig)
	chunks, _ := chunker.Split(context.Background(), text)

	contents := make([]string, len(chunks))
	for i, chunk := range chunks {
		contents[i] = chunk.Content
	}

	return contents
}

// createParentChunks 创建父块 (大块)
func (stb *SmallToBigChunker) createParentChunks(smallChunks []string) []string {
	var parentChunks []string
	currentParent := ""
	currentSize := 0

	for i, smallChunk := range smallChunks {
		// 计算是否应该开始新的父块
		shouldStartNewParent := currentSize > 0 && // 不是第一个
			(i%stb.parentMerge == 0 || // 达到合并数量
				currentSize+len(smallChunk) > stb.bigConfig.ChunkSize) // 超过大小限制

		if shouldStartNewParent {
			parentChunks = append(parentChunks, strings.TrimSpace(currentParent))
			currentParent = smallChunk
			currentSize = len(smallChunk)
		} else {
			// 添加到当前父块
			if currentParent != "" {
				currentParent += "\n\n" + smallChunk
			} else {
				currentParent = smallChunk
			}
			currentSize += len(smallChunk) + 2 // +2 for "\n\n"
		}
	}

	// 添加最后一个父块
	if currentParent != "" {
		parentChunks = append(parentChunks, strings.TrimSpace(currentParent))
	}

	return parentChunks
}

// GetParentChunks 获取父块 (用于单独索引大块)
func (stb *SmallToBigChunker) GetParentChunks(ctx context.Context, text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	// 1. 创建小块
	smallChunks := stb.createSmallChunks(text)
	if len(smallChunks) == 0 {
		return []Chunk{}, nil
	}

	// 2. 创建父块
	parentChunks := stb.createParentChunks(smallChunks)

	// 3. 构建父块 Chunk 对象
	result := make([]Chunk, len(parentChunks))
	position := 0
	for i, parentContent := range parentChunks {
		startPos := position
		endPos := position + len(parentContent)
		position = endPos

		result[i] = Chunk{
			Content: parentContent,
			Metadata: ChunkMetadata{
				Index:      i,
				StartPos:   startPos,
				EndPos:     endPos,
				ChunkType:  stb.name + "_parent",
				TokenCount: estimateTokens(parentContent),
				AdditionalMetadata: map[string]interface{}{
					"is_parent_chunk":  true,
					"child_count":      stb.parentMerge,
					"chunk_size_type":  "parent",
				},
			},
		}
	}

	return result, nil
}

// Name 返回分块器名称
func (stb *SmallToBigChunker) Name() string {
	return stb.name
}

// Validate 验证配置
func (stb *SmallToBigChunker) Validate() error {
	if stb.smallConfig.ChunkSize <= 0 {
		return fmt.Errorf("small chunk_size must be positive")
	}
	if stb.bigConfig.ChunkSize <= stb.smallConfig.ChunkSize {
		return fmt.Errorf("big chunk_size must be greater than small chunk_size")
	}
	if stb.parentMerge <= 0 {
		return fmt.Errorf("parent_merge must be positive")
	}
	return nil
}

// GetSmallConfig 返回小块配置
func (stb *SmallToBigChunker) GetSmallConfig() ChunkerConfig {
	return stb.smallConfig
}

// GetBigConfig 返回大块配置
func (stb *SmallToBigChunker) GetBigConfig() ChunkerConfig {
	return stb.bigConfig
}

// GetParentMerge 返回合并比例
func (stb *SmallToBigChunker) GetParentMerge() int {
	return stb.parentMerge
}
