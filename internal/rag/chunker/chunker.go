package chunker

import (
	"strings"
	"unicode"
)

const (
	// DefaultChunkSize 默认分块大小（字符数）
	DefaultChunkSize = 500
	// DefaultOverlap 默认重叠大小
	DefaultOverlap = 50
)

// Chunker 文本分块器
type Chunker struct {
	chunkSize int
	overlap   int
}

// NewChunker 创建分块器
func NewChunker(chunkSize, overlap int) *Chunker {
	return &Chunker{
		chunkSize: chunkSize,
		overlap:   overlap,
	}
}

// Split 将文本分成多个块
func (c *Chunker) Split(text string) []string {
	if len(text) <= c.chunkSize {
		return []string{text}
	}

	var chunks []string
	runes := []rune(text) // 支持中文
	start := 0

	for start < len(runes) {
		end := start + c.chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		// 尝试在句子边界分割
		chunk := c.findSentenceBoundary(runes[start:end])

		chunks = append(chunks, strings.TrimSpace(chunk))

		// 移动到下一个块，考虑重叠
		start += len([]rune(chunk)) - c.overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// findSentenceBoundary 在句子边界分割
func (c *Chunker) findSentenceBoundary(runes []rune) string {
	if len(runes) <= c.chunkSize {
		return string(runes)
	}

	// 从后往前找句子结束符
	sentenceEnders := []rune{'。', '！', '？', '.', '!', '?', '\n', '\r'}

	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		for _, ender := range sentenceEnders {
			if r == ender {
				return string(runes[:i+1])
			}
		}
	}

	// 如果找不到句子边界，尝试在空格分割
	for i := len(runes) - 1; i >= 0; i-- {
		if unicode.IsSpace(runes[i]) {
			return string(runes[:i])
		}
	}

	// 实在找不到，强制截断
	return string(runes[:c.chunkSize])
}

// SplitByParagraph 按段落分割
func (c *Chunker) SplitByParagraph(text string) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果段落过长，继续分割
		if len(para) > c.chunkSize {
			subChunks := c.Split(para)
			chunks = append(chunks, subChunks...)
		} else {
			chunks = append(chunks, para)
		}
	}

	return chunks
}

// GetChunkCount 获取分块数量
func (c *Chunker) GetChunkCount(text string) int {
	return len(c.Split(text))
}

// MergeChunks 合并分块
func (c *Chunker) MergeChunks(chunks []string) string {
	return strings.Join(chunks, "\n\n")
}
