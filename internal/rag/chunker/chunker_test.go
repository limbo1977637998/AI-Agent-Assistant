package chunker

import (
	"context"
	"testing"

	"ai-agent-assistant/pkg/models"
)

// MockEmbeddingModel 模拟Embedding模型
type MockEmbeddingModel struct{}

func (m *MockEmbeddingModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	return "test", nil
}

func (m *MockEmbeddingModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "test"
	close(ch)
	return ch, nil
}

func (m *MockEmbeddingModel) SupportsToolCalling() bool {
	return false
}

func (m *MockEmbeddingModel) SupportsEmbedding() bool {
	return true
}

func (m *MockEmbeddingModel) Embed(ctx context.Context, text string) ([]float64, error) {
	// 返回固定长度的向量
	dim := 1024
	vector := make([]float64, dim)
	for i := range vector {
		vector[i] = 0.1
	}
	return vector, nil
}

func (m *MockEmbeddingModel) GetModelName() string {
	return "mock-embedding"
}

func (m *MockEmbeddingModel) GetProviderName() string {
	return "mock"
}

// TestSemanticChunker 测试语义分块器
func TestSemanticChunker(t *testing.T) {
	mockModel := &MockEmbeddingModel{}

	chunker, err := NewSemanticChunker(mockModel, 0.7, 500)
	if err != nil {
		t.Fatalf("Failed to create semantic chunker: %v", err)
	}

	// 测试分句
	sentences := chunker.splitSentences("这是第一句。这是第二句！这是第三句？")
	if len(sentences) != 3 {
		t.Errorf("Expected 3 sentences, got %d", len(sentences))
	}

	t.Logf("Sentences: %v", sentences)

	// 测试短文本
	shortText := "短文本"
	chunks := chunker.Split(shortText)
	if len(chunks) != 1 {
		t.Errorf("Short text should result in 1 chunk, got %d", len(chunks))
	}

	// 测试长文本
	longText := "这是第一句话。" +
		"这是第二句话，内容更多一些。" +
		"这是第三句话，继续增加内容。" +
		"这是第四句话，还要更多内容。" +
		"这是第五句话，确保足够长。"

	chunks = chunker.Split(longText)
	if len(chunks) == 0 {
		t.Error("Long text should result in at least 1 chunk")
	}

	t.Logf("Long text split into %d chunks", len(chunks))
	for i, chunk := range chunks {
		t.Logf("Chunk %d: %s", i+1, chunk[:min(50, len(chunk))])
	}
}

// TestFixedChunker 测试固定大小分块器
func TestFixedChunker(t *testing.T) {
	chunker := &FixedChunker{
		maxSize: 50,
		overlap:  10,
	}

	text := "这是一个测试文本。" +
		"内容应该足够长以被分割成多个块。" +
		"每个块的长度应该不会超过maxSize。" +
		"相邻的块之间应该有重叠内容。" +
		"这样可以确保上下文不会在块边界处丢失。"

	chunks := chunker.Split(text)

	if len(chunks) == 0 {
		t.Error("Should have at least one chunk")
	}

	// 验证每个chunk的大小
	for i, chunk := range chunks {
		if len(chunk) > chunker.maxSize {
			t.Errorf("Chunk %d size %d exceeds maxSize %d", i, len(chunk), chunker.maxSize)
		}
	}

	t.Logf("Fixed chunker created %d chunks", len(chunks))
}

// TestCosineSimilarity 测试余弦相似度计算
func TestCosineSimilarity(t *testing.T) {
	// 相同向量
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{1.0, 2.0, 3.0}

	sim := cosineSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("Identical vectors should have similarity 1.0, got %f", sim)
	}

	// 正交向量
	c := []float64{1.0, 0.0, 0.0}
	d := []float64{0.0, 1.0, 0.0}

	sim = cosineSimilarity(c, d)
	if sim != 0.0 {
		t.Errorf("Orthogonal vectors should have similarity 0.0, got %f", sim)
	}

	// 测试不同长度的向量
	e := []float64{1.0, 2.0}
	f := []float64{1.0, 2.0, 3.0}

	sim = cosineSimilarity(e, f)
	if sim == 0 {
		t.Error("Similarity should be calculated for different length vectors")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
