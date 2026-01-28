package chunking

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

// RecursiveCharacterChunker 递归字符分块器
//
// 策略说明:
//   1. 按优先级尝试多个分隔符进行分块
//   2. 如果某个分隔符产生的块仍然过大，则尝试下一个分隔符
//   3. 递归进行直到所有块都满足大小要求
//
// 优点:
//   - 保持语义完整性 (优先在段落、句子边界分割)
//   - 灵活配置分隔符优先级
//   - 广泛验证和使用 (LangChain 核心分块器)
//
// 适用场景:
//   - 结构化文档 (有明确的段落、句子结构)
//   - 需要保持语义完整性的场景
//   - 通用文本分块
type RecursiveCharacterChunker struct {
	config ChunkerConfig
	name   string
}

// NewRecursiveCharacterChunker 创建递归字符分块器
//
// 参数:
//   config: 分块器配置
//
// 返回:
//   *RecursiveCharacterChunker: 分块器实例
//   error: 配置验证错误
func NewRecursiveCharacterChunker(config ChunkerConfig) (*RecursiveCharacterChunker, error) {
	// 验证配置
	if config.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk_size must be positive")
	}
	if config.ChunkOverlap < 0 {
		return nil, fmt.Errorf("chunk_overlap cannot be negative")
	}
	if config.ChunkOverlap >= config.ChunkSize {
		return nil, fmt.Errorf("chunk_overlap must be less than chunk_size")
	}

	// 设置默认分隔符 (优先级从高到低)
	if len(config.Separators) == 0 {
		config.Separators = []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""}
	}

	// 设置默认最小分块大小
	if config.MinChunkSize <= 0 {
		config.MinChunkSize = config.ChunkSize / 10
		if config.MinChunkSize < 50 {
			config.MinChunkSize = 50
		}
	}

	return &RecursiveCharacterChunker{
		config: config,
		name:   "recursive_character",
	}, nil
}

// Split 实现分块逻辑
func (rc *RecursiveCharacterChunker) Split(ctx context.Context, text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	// 如果文本本身小于等于分块大小，直接返回
	if len(text) <= rc.config.ChunkSize {
		return []Chunk{{
			Content: text,
			Metadata: ChunkMetadata{
				Index:      0,
				StartPos:   0,
				EndPos:     len(text),
				ChunkType:  rc.name,
				TokenCount: estimateTokens(text),
			},
		}}, nil
	}

	// 递归分块
	splits := rc.recursiveSplit(text, rc.config.Separators)

	// 合并分块 (考虑 overlap)
	chunks := rc.mergeSplits(splits)

	// 构建带元数据的 Chunk 对象
	result := make([]Chunk, len(chunks))
	position := 0
	for i, chunkText := range chunks {
		startPos := position
		endPos := position + len(chunkText)
		position = endPos - rc.config.ChunkOverlap // 考虑重叠

		result[i] = Chunk{
			Content: chunkText,
			Metadata: ChunkMetadata{
				Index:      i,
				StartPos:   startPos,
				EndPos:     endPos,
				ChunkType:  rc.name,
				TokenCount: estimateTokens(chunkText),
			},
		}
	}

	return result, nil
}

// recursiveSplit 递归分割文本
func (rc *RecursiveCharacterChunker) recursiveSplit(text string, separators []string) []string {
	// 基础情况: 如果文本足够小，直接返回
	if len(text) <= rc.config.ChunkSize {
		return []string{text}
	}

	// 如果没有分隔符了，强制分割
	if len(separators) == 0 {
		return rc.forceSplit(text)
	}

	// 获取当前分隔符
	separator := separators[0]

	// 尝试用当前分隔符分割
	var splits []string
	if separator == "" {
		// 字符级分割
		splits = rc.splitByCharacter(text)
	} else {
		// 按分隔符分割
		splits = strings.Split(text, separator)
	}

	// 处理分割结果
	var result []string
	currentChunk := ""
	currentSeparator := ""

	for i, split := range splits {
		// 如果需要保留分隔符
		if rc.config.KeepSeparator && i < len(splits)-1 {
			split += separator
		}

		// 如果添加这个 split 后不超过大小限制
		if len(currentChunk)+len(split)+len(currentSeparator) <= rc.config.ChunkSize {
			currentChunk += currentSeparator + split
			currentSeparator = separator
		} else {
			// 当前 chunk 已满，保存它
			if currentChunk != "" {
				result = append(result, currentChunk)
			}

			// 如果单个 split 就超过大小限制，递归处理
			if len(split) > rc.config.ChunkSize {
				// 使用剩余的分隔符递归分割
				recursiveSplits := rc.recursiveSplit(split, separators[1:])
				result = append(result, recursiveSplits...)
				currentChunk = ""
				currentSeparator = ""
			} else {
				// 开始新的 chunk
				currentChunk = split
				currentSeparator = separator
			}
		}
	}

	// 添加最后一个 chunk
	if currentChunk != "" {
		result = append(result, currentChunk)
	}

	return result
}

// splitByCharacter 按字符分割
func (rc *RecursiveCharacterChunker) splitByCharacter(text string) []string {
	runes := []rune(text)
	var result []string

	// 保持单词完整性
	currentWord := ""
	currentChunk := ""

	for _, r := range runes {
		if unicode.IsSpace(r) {
			// 遇到空格，保存当前单词
			if currentWord != "" {
				if len(currentChunk)+len(currentWord) <= rc.config.ChunkSize {
					currentChunk += currentWord
				} else {
					if currentChunk != "" {
						result = append(result, currentChunk)
					}
					// 如果单个单词就超过大小，强制分割
					if len(currentWord) > rc.config.ChunkSize {
						for i := 0; i < len(currentWord); i += rc.config.ChunkSize {
							end := i + rc.config.ChunkSize
							if end > len(currentWord) {
								end = len(currentWord)
							}
							result = append(result, currentWord[i:end])
						}
					} else {
						currentChunk = currentWord
					}
				}
				currentWord = ""
			}
			currentChunk += string(r)
		} else {
			currentWord += string(r)
		}
	}

	// 添加最后一个单词
	if currentWord != "" {
		currentChunk += currentWord
	}

	if currentChunk != "" {
		result = append(result, currentChunk)
	}

	return result
}

// forceSplit 强制分割 (当所有分隔符都无效时)
func (rc *RecursiveCharacterChunker) forceSplit(text string) []string {
	var result []string

	for i := 0; i < len(text); i += rc.config.ChunkSize {
		end := i + rc.config.ChunkSize
		if end > len(text) {
			end = len(text)
		}
		result = append(result, text[i:end])
	}

	return result
}

// mergeSplits 合并分割结果，考虑 overlap
func (rc *RecursiveCharacterChunker) mergeSplits(splits []string) []string {
	if len(splits) == 0 {
		return []string{}
	}

	// 如果没有 overlap，直接返回
	if rc.config.ChunkOverlap == 0 {
		return splits
	}

	var result []string
	currentChunk := splits[0]

	for i := 1; i < len(splits); i++ {
		// 如果当前 chunk + 下一个 split - overlap 不超过限制
		if len(currentChunk)+len(splits[i])-rc.config.ChunkOverlap <= rc.config.ChunkSize {
			// 合并 (保留 overlap 部分)
			currentChunk += splits[i][rc.config.ChunkOverlap:]
		} else {
			// 保存当前 chunk
			result = append(result, currentChunk)
			// 开始新的 chunk
			currentChunk = splits[i]
		}
	}

	// 添加最后一个 chunk
	result = append(result, currentChunk)

	return result
}

// Name 返回分块器名称
func (rc *RecursiveCharacterChunker) Name() string {
	return rc.name
}

// Validate 验证配置
func (rc *RecursiveCharacterChunker) Validate() error {
	if rc.config.ChunkSize <= 0 {
		return fmt.Errorf("chunk_size must be positive")
	}
	if rc.config.ChunkOverlap < 0 {
		return fmt.Errorf("chunk_overlap cannot be negative")
	}
	if rc.config.ChunkOverlap >= rc.config.ChunkSize {
		return fmt.Errorf("chunk_overlap must be less than chunk_size")
	}
	return nil
}

// estimateTokens 估算 Token 数量
// 粗略估计: 中文约 1.5 字符 = 1 token, 英文约 4 字符 = 1 token
func estimateTokens(text string) int {
	runes := []rune(text)
	chineseChars := 0
	otherChars := 0

	for _, r := range runes {
		if unicode.Is(unicode.Han, r) {
			chineseChars++
		} else {
			otherChars++
		}
	}

	// 中文: 1.5 字符 ≈ 1 token
	// 英文: 4 字符 ≈ 1 token
	chineseTokens := int(float64(chineseChars) / 1.5)
	otherTokens := otherChars / 4

	return chineseTokens + otherTokens
}
