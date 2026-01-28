package chunking

import (
	"context"
	"fmt"
)

// LegacyChunkerAdapter 旧版分块器适配器
// 将旧版 Chunker 和 SemanticChunker 适配到新接口
type LegacyChunkerAdapter struct {
	chunkerType string // "fixed" or "semantic"
	legacyChunker interface{}
}

// NewLegacyChunkerAdapter 创建旧版分块器适配器
// 用于平滑过渡，保持向后兼容
func NewLegacyChunkerAdapter(chunkerType string, legacyChunker interface{}) (*LegacyChunkerAdapter, error) {
	if legacyChunker == nil {
		return nil, fmt.Errorf("legacy chunker cannot be nil")
	}

	return &LegacyChunkerAdapter{
		chunkerType:   chunkerType,
		legacyChunker: legacyChunker,
	}, nil
}

// Split 实现 ChunkerStrategy 接口
// 将旧版分块器的结果转换为新格式
func (la *LegacyChunkerAdapter) Split(ctx context.Context, text string) ([]Chunk, error) {
	switch la.chunkerType {
	case "fixed":
		// 旧版 FixedChunker 返回 []string
		if fixedChunker, ok := la.legacyChunker.(interface{ Split(string) []string }); ok {
			texts := fixedChunker.Split(text)
			return convertStringsToChunks(texts, "fixed"), nil
		}
		return nil, fmt.Errorf("invalid fixed chunker type")

	case "semantic":
		// 旧版 SemanticChunker 返回 []string
		if semanticChunker, ok := la.legacyChunker.(interface{ Split(string) []string }); ok {
			texts := semanticChunker.Split(text)
			return convertStringsToChunks(texts, "semantic"), nil
		}
		return nil, fmt.Errorf("invalid semantic chunker type")

	default:
		return nil, fmt.Errorf("unknown legacy chunker type: %s", la.chunkerType)
	}
}

// Name 返回分块器名称
func (la *LegacyChunkerAdapter) Name() string {
	return la.chunkerType
}

// Validate 验证配置
func (la *LegacyChunkerAdapter) Validate() error {
	if la.legacyChunker == nil {
		return fmt.Errorf("legacy chunker is nil")
	}
	return nil
}

// convertStringsToChunks 将旧版的 []string 转换为新版 []Chunk
func convertStringsToChunks(texts []string, chunkType string) []Chunk {
	chunks := make([]Chunk, len(texts))
	position := 0

	for i, text := range texts {
		startPos := position
		endPos := position + len(text)
		position = endPos

		chunks[i] = Chunk{
			Content: text,
			Metadata: ChunkMetadata{
				Index:      i,
				StartPos:   startPos,
				EndPos:     endPos,
				ChunkType:  chunkType,
				TokenCount: estimateTokens(text),
			},
		}
	}

	return chunks
}

// ChunkerManager 分块器管理器
// 提供统一的分块器创建和管理接口
type ChunkerManager struct {
	factory *ChunkerFactory
}

// NewChunkerManager 创建分块器管理器
func NewChunkerManager() *ChunkerManager {
	return &ChunkerManager{
		factory: NewChunkerFactory(),
	}
}

// GetFactory 获取工厂
func (m *ChunkerManager) GetFactory() *ChunkerFactory {
	return m.factory
}

// GetRegistry 获取注册表
func (m *ChunkerManager) GetRegistry() *ChunkerRegistry {
	return m.factory.GetRegistry()
}

// CreateChunker 创建分块器 (便捷方法)
func (m *ChunkerManager) CreateChunker(chunkerType string, config interface{}) (ChunkerStrategy, error) {
	return m.factory.CreateChunker(chunkerType, config)
}

// CreateRecursiveChunker 创建递归分块器 (便捷方法)
func (m *ChunkerManager) CreateRecursiveChunker(chunkSize, overlap int) (*RecursiveCharacterChunker, error) {
	cfg := ChunkerConfig{
		ChunkSize:     chunkSize,
		ChunkOverlap:  overlap,
		MinChunkSize:  chunkSize / 10,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, err := m.factory.CreateChunker("recursive", cfg)
	if err != nil {
		return nil, err
	}

	return chunker.(*RecursiveCharacterChunker), nil
}

// CreateSmallToBigChunker 创建小到大分块器 (便捷方法)
func (m *ChunkerManager) CreateSmallToBigChunker(smallSize, bigSize, overlap int) (*SmallToBigChunker, error) {
	smallConfig := ChunkerConfig{
		ChunkSize:     smallSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	bigConfig := ChunkerConfig{
		ChunkSize:     bigSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, err := m.factory.CreateChunker("small_to_big", map[string]interface{}{
		"small":       smallConfig,
		"big":         bigConfig,
		"parent_merge": 3,
	})
	if err != nil {
		return nil, err
	}

	return chunker.(*SmallToBigChunker), nil
}

// CreateParentDocumentChunker 创建父文档分块器 (便捷方法)
func (m *ChunkerManager) CreateParentDocumentChunker(parentSize, childSize, overlap int) (*ParentDocumentChunker, error) {
	parentConfig := ChunkerConfig{
		ChunkSize:     parentSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	childConfig := ChunkerConfig{
		ChunkSize:     childSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	chunker, err := m.factory.CreateChunker("parent_document", map[string]interface{}{
		"parent":          parentConfig,
		"child":           childConfig,
		"child_per_parent": 5,
	})
	if err != nil {
		return nil, err
	}

	return chunker.(*ParentDocumentChunker), nil
}

// ListAvailableChunkers 列出所有可用的分块器类型
func (m *ChunkerManager) ListAvailableChunkers() []string {
	return m.factory.ListChunkerTypes()
}

// GetChunkerDescription 获取分块器描述
func (m *ChunkerManager) GetChunkerDescription(chunkerType string) string {
	info := m.factory.GetChunkerInfo(chunkerType)
	if desc, ok := info["description"].(string); ok {
		return desc
	}
	return "Unknown chunker type"
}
