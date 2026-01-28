package chunking

import (
	"context"
	"fmt"
)

// ParentDocumentChunker 父文档分块器
//
// 策略说明:
//   1. 先将文档分成大块 (parent chunks)
//   2. 从每个父块中提取多个子块 (child chunks) 用于检索
//   3. 检索时返回子块，但可以追溯到父块获取完整上下文
//
// 与 SmallToBigChunker 的区别:
//   - SmallToBig: 先小块 → 后合并大块
//   - ParentDocument: 先大块 → 后提取小块
//
// 优点:
//   - 父块保持完整语义 (适合全文理解)
//   - 子块提高检索精度 (更精准匹配)
//   - 灵活性高 (可配置父子块大小)
//
// 适用场景:
//   - 需要同时支持精确定位和全局理解
//   - 文档问答系统
//   - 长文档摘要和检索
type ParentDocumentChunker struct {
	parentConfig ChunkerConfig // 父块配置
	childConfig  ChunkerConfig // 子块配置
	name         string
	childPerParent int         // 每个父块提取多少个子块
}

// NewParentDocumentChunker 创建父文档分块器
//
// 参数:
//   parentConfig: 父块配置 (大块，用于上下文)
//   childConfig: 子块配置 (小块，用于检索)
//   childPerParent: 每个父块提取多少个子块 (0 = 自动计算)
//
// 返回:
//   *ParentDocumentChunker: 分块器实例
//   error: 配置验证错误
func NewParentDocumentChunker(parentConfig, childConfig ChunkerConfig, childPerParent int) (*ParentDocumentChunker, error) {
	// 验证父块配置
	if parentConfig.ChunkSize <= 0 {
		return nil, fmt.Errorf("parent chunk_size must be positive")
	}
	if parentConfig.ChunkOverlap < 0 {
		return nil, fmt.Errorf("parent chunk_overlap cannot be negative")
	}

	// 验证子块配置
	if childConfig.ChunkSize <= 0 {
		return nil, fmt.Errorf("child chunk_size must be positive")
	}
	if childConfig.ChunkOverlap < 0 {
		return nil, fmt.Errorf("child chunk_overlap cannot be negative")
	}

	// 父块必须大于子块
	if parentConfig.ChunkSize <= childConfig.ChunkSize {
		return nil, fmt.Errorf("parent chunk_size must be greater than child chunk_size")
	}

	// 设置默认 childPerParent (自动计算)
	if childPerParent <= 0 {
		// 估算: 父块大小 / 子块大小
		childPerParent = parentConfig.ChunkSize / childConfig.ChunkSize
		if childPerParent < 1 {
			childPerParent = 1
		}
		// 限制上限，避免子块过多
		if childPerParent > 10 {
			childPerParent = 10
		}
	}

	return &ParentDocumentChunker{
		parentConfig:    parentConfig,
		childConfig:     childConfig,
		name:            "parent_document",
		childPerParent: childPerParent,
	}, nil
}

// Split 实现分块逻辑
// 返回的是子块 (child chunks)，每个子块记录其所属的父块索引
func (pdc *ParentDocumentChunker) Split(ctx context.Context, text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	// 1. 先创建父块 (大块)
	parentChunks := pdc.createParentChunks(text)
	if len(parentChunks) == 0 {
		return []Chunk{}, nil
	}

	// 2. 从每个父块中提取子块
	var allChildChunks []ChildChunkInfo
	globalChunkIndex := 0

	for parentIndex, parentChunk := range parentChunks {
		childChunks := pdc.createChildChunks(parentChunk.content, parentIndex)
		allChildChunks = append(allChildChunks, childChunks...)

		// 更新全局位置
		for i := range childChunks {
			childChunks[i].globalIndex = globalChunkIndex
			globalChunkIndex++
		}
	}

	// 3. 构建返回结果 (子块)
	result := make([]Chunk, len(allChildChunks))
	for i, childInfo := range allChildChunks {
		parentContent := parentChunks[childInfo.parentIndex].content

		result[i] = Chunk{
			Content: childInfo.content,
			Metadata: ChunkMetadata{
				Index:            childInfo.globalIndex,
				StartPos:         childInfo.startPos,
				EndPos:           childInfo.endPos,
				ChunkType:        pdc.name + "_child",
				TokenCount:       estimateTokens(childInfo.content),
				ParentChunkIndex: childInfo.parentIndex,
				AdditionalMetadata: map[string]interface{}{
					"parent_content":      parentContent, // 存储父块内容
					"parent_start_pos":    parentChunks[childInfo.parentIndex].startPos,
					"parent_end_pos":      parentChunks[childInfo.parentIndex].endPos,
					"is_child_chunk":      true,
					"chunk_size_type":     "child",
				},
			},
		}
	}

	return result, nil
}

// ParentChunkInfo 父块信息
type ParentChunkInfo struct {
	content  string
	startPos int
	endPos   int
}

// ChildChunkInfo 子块信息
type ChildChunkInfo struct {
	content     string
	parentIndex int
	globalIndex int
	startPos    int
	endPos      int
}

// createParentChunks 创建父块
func (pdc *ParentDocumentChunker) createParentChunks(text string) []ParentChunkInfo {
	// 使用递归字符分块器创建父块
	recursiveConfig := ChunkerConfig{
		ChunkSize:     pdc.parentConfig.ChunkSize,
		ChunkOverlap:  pdc.parentConfig.ChunkOverlap,
		MinChunkSize:  pdc.parentConfig.MinChunkSize,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, _ := NewRecursiveCharacterChunker(recursiveConfig)
	chunks, _ := chunker.Split(context.Background(), text)

	result := make([]ParentChunkInfo, len(chunks))
	position := 0
	for i, chunk := range chunks {
		startPos := position
		endPos := position + len(chunk.Content)
		position = endPos - pdc.parentConfig.ChunkOverlap

		result[i] = ParentChunkInfo{
			content:  chunk.Content,
			startPos: startPos,
			endPos:   endPos,
		}
	}

	return result
}

// createChildChunks 从父块中创建子块
func (pdc *ParentDocumentChunker) createChildChunks(parentText string, parentIndex int) []ChildChunkInfo {
	// 使用递归字符分块器创建子块
	recursiveConfig := ChunkerConfig{
		ChunkSize:     pdc.childConfig.ChunkSize,
		ChunkOverlap:  pdc.childConfig.ChunkOverlap,
		MinChunkSize:  pdc.childConfig.MinChunkSize,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, _ := NewRecursiveCharacterChunker(recursiveConfig)
	chunks, _ := chunker.Split(context.Background(), parentText)

	// 限制子块数量
	maxChildren := pdc.childPerParent
	if len(chunks) > maxChildren {
		// 均匀选择子块
		step := float64(len(chunks)) / float64(maxChildren)
		selectedChunks := make([]Chunk, 0, maxChildren)
		for i := 0; i < maxChildren; i++ {
			idx := int(float64(i) * step)
			selectedChunks = append(selectedChunks, chunks[idx])
		}
		chunks = selectedChunks
	}

	result := make([]ChildChunkInfo, len(chunks))
	position := 0
	for i, chunk := range chunks {
		startPos := position
		endPos := position + len(chunk.Content)
		position = endPos - pdc.childConfig.ChunkOverlap

		result[i] = ChildChunkInfo{
			content:     chunk.Content,
			parentIndex: parentIndex,
			startPos:    startPos,
			endPos:      endPos,
		}
	}

	return result
}

// GetParentChunks 获取父块 (用于单独索引父块)
func (pdc *ParentDocumentChunker) GetParentChunks(ctx context.Context, text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	parentChunks := pdc.createParentChunks(text)

	result := make([]Chunk, len(parentChunks))
	for i, parentInfo := range parentChunks {
		result[i] = Chunk{
			Content: parentInfo.content,
			Metadata: ChunkMetadata{
				Index:      i,
				StartPos:   parentInfo.startPos,
				EndPos:     parentInfo.endPos,
				ChunkType:  pdc.name + "_parent",
				TokenCount: estimateTokens(parentInfo.content),
				AdditionalMetadata: map[string]interface{}{
					"is_parent_chunk": true,
					"child_count":     pdc.childPerParent,
					"chunk_size_type": "parent",
				},
			},
		}
	}

	return result, nil
}

// Name 返回分块器名称
func (pdc *ParentDocumentChunker) Name() string {
	return pdc.name
}

// Validate 验证配置
func (pdc *ParentDocumentChunker) Validate() error {
	if pdc.parentConfig.ChunkSize <= 0 {
		return fmt.Errorf("parent chunk_size must be positive")
	}
	if pdc.childConfig.ChunkSize <= 0 {
		return fmt.Errorf("child chunk_size must be positive")
	}
	if pdc.parentConfig.ChunkSize <= pdc.childConfig.ChunkSize {
		return fmt.Errorf("parent chunk_size must be greater than child chunk_size")
	}
	if pdc.childPerParent <= 0 {
		return fmt.Errorf("child_per_parent must be positive")
	}
	return nil
}

// GetParentConfig 返回父块配置
func (pdc *ParentDocumentChunker) GetParentConfig() ChunkerConfig {
	return pdc.parentConfig
}

// GetChildConfig 返回子块配置
func (pdc *ParentDocumentChunker) GetChildConfig() ChunkerConfig {
	return pdc.childConfig
}

// GetChildPerParent 返回每个父块的子块数量
func (pdc *ParentDocumentChunker) GetChildPerParent() int {
	return pdc.childPerParent
}
