package rag

import (
	"context"
	"fmt"

	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/rag/chunker"
	"ai-agent-assistant/internal/rag/embedding"
	"ai-agent-assistant/internal/rag/parser"
	"ai-agent-assistant/internal/rag/store"
)

// RAG RAG系统
type RAG struct {
	parser    parser.Parser
	chunker   chunker.Chunker
	embedding embedding.EmbeddingProvider
	store     store.VectorStore
	config    *config.Config
}

// NewRAG 创建RAG系统
func NewRAG(cfg *config.Config) (*RAG, error) {
	// 初始化各个组件
	p := parser.NewParser()
	c := chunker.NewChunker(chunker.DefaultChunkSize, chunker.DefaultOverlap)

	// 初始化embedding提供者
	ep, err := embedding.NewEmbeddingProvider(cfg.Models.GLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}

	// 初始化向量存储（内存存储）
	vs := store.NewInMemoryVectorStore(ep)

	return &RAG{
		parser:    p,
		chunker:   *c,
		embedding: ep,
		store:     vs,
		config:    cfg,
	}, nil
}

// AddDocument 添加文档到知识库
func (r *RAG) AddDocument(ctx context.Context, docPath string) error {
	// 1. 解析文档
	text, err := r.parser.Parse(docPath)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}

	// 2. 分块
	chunks := r.chunker.Split(text)

	// 3. 向量化并存储
	for i, chunk := range chunks {
		vector, err := r.embedding.Embed(ctx, chunk)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		metadata := map[string]interface{}{
			"source": docPath,
			"chunk":  i,
		}

		if err := r.store.Add(ctx, vector, chunk, metadata); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}
	}

	return nil
}

// AddText 直接添加文本到知识库
func (r *RAG) AddText(ctx context.Context, text string, source string) error {
	// 1. 分块
	chunks := r.chunker.Split(text)

	// 2. 向量化并存储
	for i, chunk := range chunks {
		vector, err := r.embedding.Embed(ctx, chunk)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		metadata := map[string]interface{}{
			"source": source,
			"chunk":  i,
		}

		if err := r.store.Add(ctx, vector, chunk, metadata); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}
	}

	return nil
}

// Retrieve 检索相关内容
func (r *RAG) Retrieve(ctx context.Context, query string, topK int) ([]string, error) {
	// 1. 将查询向量化
	queryVector, err := r.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. 检索最相似的内容
	results, err := r.store.Search(ctx, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

// BuildContext 构建增强上下文
func (r *RAG) BuildContext(ctx context.Context, query string, topK int) (string, error) {
	results, err := r.Retrieve(ctx, query, topK)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	context := "参考信息：\n"
	for i, result := range results {
		context += fmt.Sprintf("\n[%d] %s", i+1, result)
	}

	return context, nil
}

// GetStats 获取知识库统计信息
func (r *RAG) GetStats() map[string]interface{} {
	return r.store.Stats()
}
