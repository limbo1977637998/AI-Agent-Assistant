package chunker

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"

	"ai-agent-assistant/internal/llm"
)

// SemanticChunker 语义分块器
// 基于句子语义相似度进行分块，相关联的句子会被分到同一块
type SemanticChunker struct {
	embeddingProvider llm.Model
	threshold         float64 // 相似度阈值，低于此值则开始新块
	maxChunkSize     int     // 单个chunk的最大字符数
}

// NewSemanticChunker 创建语义分块器
func NewSemanticChunker(embeddingModel llm.Model, threshold float64, maxChunkSize int) (*SemanticChunker, error) {
	if embeddingModel == nil {
		return nil, fmt.Errorf("embedding model is required")
	}
	if !embeddingModel.SupportsEmbedding() {
		return nil, fmt.Errorf("model does not support embedding")
	}
	if threshold <= 0 || threshold > 1 {
		threshold = 0.7 // 默认阈值
	}
	if maxChunkSize <= 0 {
		maxChunkSize = 500 // 默认最大chunk大小
	}

	return &SemanticChunker{
		embeddingProvider: embeddingModel,
		threshold:         threshold,
		maxChunkSize:      maxChunkSize,
	}, nil
}

// Split 将文本按语义分块
func (sc *SemanticChunker) Split(text string) []string {
	// 1. 分句
	sentences := sc.splitSentences(text)
	if len(sentences) == 0 {
		return []string{text}
	}

	// 2. 如果只有一个句子，直接返回
	if len(sentences) == 1 {
		return sentences
	}

	// 3. 计算所有句子的embedding
	ctx := context.Background()
	embeddings, err := sc.computeEmbeddings(ctx, sentences)
	if err != nil {
		// 如果embedding失败，回退到固定大小分块
		fallback := &FixedChunker{maxSize: sc.maxChunkSize, overlap: 50}
		return fallback.Split(text)
	}

	// 4. 基于相似度分块
	chunks := sc.groupBySimilarity(sentences, embeddings)

	// 5. 合并过小的chunk并拆分过大的chunk
	chunks = sc.optimizeChunks(chunks)

	return chunks
}

// splitSentences 将文本分割成句子
func (sc *SemanticChunker) splitSentences(text string) []string {
	// 使用正则表达式分句（支持中英文）
	re := regexp.MustCompile(`([。！？.!?]+)\s*`)
	matches := re.Split(text, -1)

	sentences := make([]string, 0, len(matches))
	for _, sentence := range matches {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// computeEmbeddings 计算所有句子的embedding
func (sc *SemanticChunker) computeEmbeddings(ctx context.Context, sentences []string) ([][]float64, error) {
	embeddings := make([][]float64, len(sentences))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(sentences))

	// 批量计算embedding（并发）
	for i, sentence := range sentences {
		wg.Add(1)
		go func(idx int, sent string) {
			defer wg.Done()

			embedding, err := sc.embeddingProvider.Embed(ctx, sent)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			embeddings[idx] = embedding
			mu.Unlock()
		}(i, sentence)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if err := <-errChan; err != nil {
		return nil, err
	}

	return embeddings, nil
}

// groupBySimilarity 基于相似度将句子分组
func (sc *SemanticChunker) groupBySimilarity(sentences []string, embeddings [][]float64) []string {
	chunks := make([]string, 0)
	currentChunk := make([]string, 0)
	currentChunkSize := 0

	// 添加第一个句子
	if len(sentences) > 0 {
		currentChunk = append(currentChunk, sentences[0])
		currentChunkSize = len(sentences[0])
	}

	// 遍历后续句子
	for i := 1; i < len(sentences); i++ {
		// 计算当前句子与前一个句子的相似度
		similarity := cosineSimilarity(embeddings[i-1], embeddings[i])

		// 判断是否应该开始新的chunk
		shouldStartNewChunk := similarity < sc.threshold ||
			currentChunkSize+len(sentences[i]) > sc.maxChunkSize

		if shouldStartNewChunk {
			// 保存当前chunk
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.Join(currentChunk, ""))
			}

			// 开始新的chunk
			currentChunk = []string{sentences[i]}
			currentChunkSize = len(sentences[i])
		} else {
			// 添加到当前chunk
			currentChunk = append(currentChunk, sentences[i])
			currentChunkSize += len(sentences[i])
		}
	}

	// 添加最后一个chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, ""))
	}

	return chunks
}

// optimizeChunks 优化chunks：合并过小的，拆分过大的
func (sc *SemanticChunker) optimizeChunks(chunks []string) []string {
	optimized := make([]string, 0, len(chunks))

	for _, chunk := range chunks {
		if len(chunk) > sc.maxChunkSize {
			// 拆分过大的chunk
			subChunks := sc.splitLargeChunk(chunk)
			optimized = append(optimized, subChunks...)
		} else if len(chunk) < 100 {
			// 合并过小的chunk到前一个或后一个
			if len(optimized) > 0 && len(optimized[len(optimized)-1])+len(chunk) < sc.maxChunkSize {
				// 合并到前一个chunk
				optimized[len(optimized)-1] += chunk
			} else {
				// 不能合并，直接添加
				optimized = append(optimized, chunk)
			}
		} else {
			// 大小合适的chunk
			optimized = append(optimized, chunk)
		}
	}

	return optimized
}

// splitLargeChunk 拆分过大的chunk
func (sc *SemanticChunker) splitLargeChunk(chunk string) []string {
	sentences := sc.splitSentences(chunk)
	result := make([]string, 0)
	currentChunk := ""
	currentSize := 0

	for _, sentence := range sentences {
		if currentSize+len(sentence) > sc.maxChunkSize && currentChunk != "" {
			result = append(result, currentChunk)
			currentChunk = sentence
			currentSize = len(sentence)
		} else {
			currentChunk += sentence
			currentSize += len(sentence)
		}
	}

	if currentChunk != "" {
		result = append(result, currentChunk)
	}

	return result
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// FixedChunker 固定大小分块器（作为回退方案）
type FixedChunker struct {
	maxSize int
	overlap int
}

// Split 固定大小分块
func (fc *FixedChunker) Split(text string) []string {
	if len(text) <= fc.maxSize {
		return []string{text}
	}

	chunks := make([]string, 0)
	start := 0

	for start < len(text) {
		end := start + fc.maxSize
		if end > len(text) {
			end = len(text)
		}

		chunks = append(chunks, text[start:end])

		// 移动start位置，保留overlap
		start = end - fc.overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}
