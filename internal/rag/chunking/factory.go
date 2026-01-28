package chunking

import (
	"fmt"
)

// ChunkerFactory 分块器工厂
// 用于创建不同类型的分块器实例
type ChunkerFactory struct {
	registry *ChunkerRegistry
}

// NewChunkerFactory 创建分块器工厂
func NewChunkerFactory() *ChunkerFactory {
	return &ChunkerFactory{
		registry: NewChunkerRegistry(),
	}
}

// CreateChunker 创建分块器的通用方法
//
// 参数:
//   chunkerType: 分块器类型 (recursive, small_to_big, parent_document, fixed, semantic)
//   config: 分块器配置 (可以是 ChunkerConfig 或 map[string]interface{})
//
// 返回:
//   ChunkerStrategy: 分块器实例
//   error: 创建错误
func (f *ChunkerFactory) CreateChunker(chunkerType string, config interface{}) (ChunkerStrategy, error) {
	switch chunkerType {
	case "recursive", "recursive_character":
		return f.createRecursiveChunker(config)

	case "small_to_big", "small-to-big":
		return f.createSmallToBigChunker(config)

	case "parent_document", "parent-document":
		return f.createParentDocumentChunker(config)

	case "fixed", "fixed_size":
		return f.createFixedChunker(config)

	case "semantic":
		return f.createSemanticChunker(config)

	default:
		return nil, fmt.Errorf("unknown chunker type: %s", chunkerType)
	}
}

// createRecursiveChunker 创建递归字符分块器
func (f *ChunkerFactory) createRecursiveChunker(config interface{}) (*RecursiveCharacterChunker, error) {
	cfg, err := parseChunkerConfig(config)
	if err != nil {
		return nil, err
	}

	return NewRecursiveCharacterChunker(cfg)
}

// createSmallToBigChunker 创建小到大分块器
func (f *ChunkerFactory) createSmallToBigChunker(config interface{}) (*SmallToBigChunker, error) {
	// 配置格式: {small: {...}, big: {...}, parent_merge: 3}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config type for small_to_big chunker, expected map")
	}

	// 解析小块配置
	smallConfig, err := parseChunkerConfig(configMap["small"])
	if err != nil {
		return nil, fmt.Errorf("invalid small config: %w", err)
	}

	// 解析大块配置
	bigConfig, err := parseChunkerConfig(configMap["big"])
	if err != nil {
		return nil, fmt.Errorf("invalid big config: %w", err)
	}

	// 解析合并比例
	parentMerge := 3 // 默认值
	if val, ok := configMap["parent_merge"].(int); ok && val > 0 {
		parentMerge = val
	}

	return NewSmallToBigChunker(smallConfig, bigConfig, parentMerge)
}

// createParentDocumentChunker 创建父文档分块器
func (f *ChunkerFactory) createParentDocumentChunker(config interface{}) (*ParentDocumentChunker, error) {
	// 配置格式: {parent: {...}, child: {...}, child_per_parent: 5}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config type for parent_document chunker, expected map")
	}

	// 解析父块配置
	parentConfig, err := parseChunkerConfig(configMap["parent"])
	if err != nil {
		return nil, fmt.Errorf("invalid parent config: %w", err)
	}

	// 解析子块配置
	childConfig, err := parseChunkerConfig(configMap["child"])
	if err != nil {
		return nil, fmt.Errorf("invalid child config: %w", err)
	}

	// 解析每个父块的子块数量
	childPerParent := 5 // 默认值
	if val, ok := configMap["child_per_parent"].(int); ok && val > 0 {
		childPerParent = val
	}

	return NewParentDocumentChunker(parentConfig, childConfig, childPerParent)
}

// createFixedChunker 创建固定大小分块器
func (f *ChunkerFactory) createFixedChunker(config interface{}) (ChunkerStrategy, error) {
	cfg, err := parseChunkerConfig(config)
	if err != nil {
		return nil, err
	}

	// 使用递归字符分块器实现，但配置更简单
	cfg.Separators = []string{"\n\n", "\n", "。", "!", "?", ".", " ", ""}
	return NewRecursiveCharacterChunker(cfg)
}

// createSemanticChunker 创建语义分块器
// 注意: 语义分块器需要 embedding 模型，这里暂时返回占位符
func (f *ChunkerFactory) createSemanticChunker(config interface{}) (ChunkerStrategy, error) {
	// 语义分块器需要 embedding 模型支持
	// 这里返回一个错误，提示使用专门的创建方法
	return nil, fmt.Errorf("semantic chunker requires embedding model, use CreateSemanticChunker method instead")
}

// CreateSemanticChunker 创建语义分块器 (专用方法)
// 需要额外的 embedding 模型参数
func (f *ChunkerFactory) CreateSemanticChunker(config ChunkerConfig, embeddingModel interface{}) (ChunkerStrategy, error) {
	// 这个方法将在重构现有 SemanticChunker 后实现
	// 目前返回占位符
	return nil, fmt.Errorf("semantic chunker not yet integrated, please use chunker.SemanticChunker directly")
}

// GetRegistry 获取注册表
func (f *ChunkerFactory) GetRegistry() *ChunkerRegistry {
	return f.registry
}

// parseChunkerConfig 解析分块器配置
// 支持两种格式: ChunkerConfig 或 map[string]interface{}
func parseChunkerConfig(config interface{}) (ChunkerConfig, error) {
	cfg := DefaultChunkerConfig()

	// 如果是 ChunkerConfig 类型，直接返回
	if chunkerConfig, ok := config.(ChunkerConfig); ok {
		return chunkerConfig, nil
	}

	// 如果是 map 类型，解析字段
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return cfg, fmt.Errorf("invalid config type, expected ChunkerConfig or map")
	}

	// 解析各个字段
	if val, ok := configMap["chunk_size"].(int); ok && val > 0 {
		cfg.ChunkSize = val
	}

	if val, ok := configMap["chunk_overlap"].(int); ok && val >= 0 {
		cfg.ChunkOverlap = val
	}

	if val, ok := configMap["min_chunk_size"].(int); ok && val > 0 {
		cfg.MinChunkSize = val
	}

	if val, ok := configMap["separators"].([]string); ok && len(val) > 0 {
		cfg.Separators = val
	}

	if val, ok := configMap["keep_separator"].(bool); ok {
		cfg.KeepSeparator = val
	}

	return cfg, nil
}

// ListChunkerTypes 列出所有支持的分块器类型
func (f *ChunkerFactory) ListChunkerTypes() []string {
	return []string{
		"recursive",         // 递归字符分块 (推荐)
		"small_to_big",      // 小到大分块
		"parent_document",   // 父文档分块
		"fixed",             // 固定大小分块
		"semantic",          // 语义分块 (需要 embedding)
	}
}

// GetChunkerInfo 获取分块器信息
func (f *ChunkerFactory) GetChunkerInfo(chunkerType string) map[string]interface{} {
	info := map[string]interface{}{
		"type": chunkerType,
	}

	switch chunkerType {
	case "recursive", "recursive_character":
		info["name"] = "Recursive Character Chunker"
		info["description"] = "递归字符分块器，按优先级尝试多个分隔符进行分块"
		info["use_case"] = "结构化文档，需要保持语义完整性"
		info["config_format"] = "ChunkerConfig"

	case "small_to_big":
		info["name"] = "Small to Big Chunker"
		info["description"] = "小到大分块器，先用小块分割再合并成大块"
		info["use_case"] = "需要高精度检索的场景"
		info["config_format"] = "{small: ChunkerConfig, big: ChunkerConfig, parent_merge: int}"

	case "parent_document":
		info["name"] = "Parent Document Chunker"
		info["description"] = "父文档分块器，先分大块再提取小块"
		info["use_case"] = "需要同时支持精确定位和全局理解"
		info["config_format"] = "{parent: ChunkerConfig, child: ChunkerConfig, child_per_parent: int}"

	case "fixed", "fixed_size":
		info["name"] = "Fixed Size Chunker"
		info["description"] = "固定大小分块器，简单的固定长度分割"
		info["use_case"] = "简单场景，对语义要求不高"
		info["config_format"] = "ChunkerConfig"

	case "semantic":
		info["name"] = "Semantic Chunker"
		info["description"] = "语义分块器，基于句子语义相似度分块"
		info["use_case"] = "需要高质量语义分割的场景"
		info["requires_embedding"] = true
		info["config_format"] = "{threshold: float, max_chunk_size: int}"

	default:
		info["error"] = "Unknown chunker type"
	}

	return info
}
