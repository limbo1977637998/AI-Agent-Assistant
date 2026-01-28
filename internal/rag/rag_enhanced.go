package rag

import (
	"context"
	"fmt"

	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/internal/rag/chunker"
	"ai-agent-assistant/internal/rag/embedding"
	"ai-agent-assistant/internal/rag/parser"
	"ai-agent-assistant/internal/rag/reranker"
	"ai-agent-assistant/internal/rag/retriever"
	"ai-agent-assistant/internal/rag/store"
	"ai-agent-assistant/internal/vectordb"
)

// RAGEnhanced 增强版RAG系统（支持语义分块、混合检索、重排序）
type RAGEnhanced struct {
	parser         parser.Parser
	chunker        chunker.Chunker
	semanticChunker *chunker.SemanticChunker // 语义分块器
	embedding      llm.Model                 // 使用统一的Model接口
	store          store.VectorStore
	hybridRetriever *retriever.HybridRetriever // 混合检索器
	reranker       reranker.Reranker            // 重排序器
	config         *config.Config
	enableHybrid   bool                        // 是否启用混合检索
	enableRerank   bool                        // 是否启用重排序
}

// NewRAGEnhanced 创建增强版RAG系统
func NewRAGEnhanced(cfg *config.Config, modelManager *llm.ModelManager) (*RAGEnhanced, error) {
	// 初始化各个组件
	p := parser.NewParser()

	// 1. 初始化embedding模型
	embeddingModelName := cfg.Agent.EmbeddingModel
	if embeddingModelName == "" {
		embeddingModelName = "qwen" // 默认使用千问
	}

	embeddingModel, err := modelManager.GetModel(embeddingModelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding model: %w", err)
	}

	if !embeddingModel.SupportsEmbedding() {
		return nil, fmt.Errorf("model %s does not support embedding", embeddingModelName)
	}

	// 2. 初始化分块器
	var semanticChunker *chunker.SemanticChunker
	if cfg.RAG.ChunkSize > 0 {
		// 创建语义分块器
		semanticChunker, err = chunker.NewSemanticChunker(
			embeddingModel,
			0.7,         // 相似度阈值
			cfg.RAG.ChunkSize, // 最大chunk大小
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create semantic chunker: %w", err)
		}
	}

	// 保留旧版chunker作为回退
	var modelConfig config.ModelConfig
	switch embeddingModelName {
	case "qwen":
		modelConfig = cfg.Models.Qwen
	case "glm":
		modelConfig = cfg.Models.GLM
	}

	ep, err := embedding.NewEmbeddingProvider(embeddingModelName, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}
	c := chunker.NewChunker(cfg.RAG.ChunkSize, cfg.RAG.ChunkOverlap)

	// 3. 初始化向量存储
	var vs store.VectorStore
	if cfg.VectorDB.Provider == "milvus" {
		milvusConfig := &vectordb.MilvusConfig{
			Address:  cfg.VectorDB.Milvus.Address,
			Username: "",
			Password: "",
			Database: "default",
		}

		milvusClient, err := vectordb.NewMilvusClient(milvusConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create milvus client: %w", err)
		}

		vs = store.NewMilvusVectorStore(
			milvusClient,
			cfg.VectorDB.Milvus.CollectionName,
			cfg.VectorDB.Milvus.Dimension,
		)
	} else {
		vs = store.NewInMemoryVectorStore(ep)
	}

	// 4. 创建向量检索器适配器
	vectorRetriever := &VectorStoreAdapter{store: vs}

	// 5. 初始化混合检索器
	hybridRetriever := retriever.NewHybridRetriever(vectorRetriever, embeddingModel, 60)

	// 6. 初始化重排序器（可选）
	var r reranker.Reranker
	if cfg.RAG.EnableHybridSearch {
		// 如果启用混合检索，可以使用简单重排序器
		// 如果有CrossEncoder API key，则使用CrossEncoder
		r = reranker.NewSimpleReranker(0.3, 0.7) // 关键词权重0.3，向量权重0.7
	}

	return &RAGEnhanced{
		parser:          p,
		chunker:         *c,
		semanticChunker: semanticChunker,
		embedding:       embeddingModel,
		store:           vs,
		hybridRetriever: hybridRetriever,
		reranker:        r,
		config:          cfg,
		enableHybrid:    cfg.RAG.EnableHybridSearch,
		enableRerank:    true,
	}, nil
}

// VectorStoreAdapter 向量存储适配器（实现VectorRetriever接口）
type VectorStoreAdapter struct {
	store store.VectorStore
}

// Search 实现VectorRetriever接口
func (a *VectorStoreAdapter) Search(ctx context.Context, queryVector []float64, topK int) ([]retriever.VectorSearchResult, error) {
	results, err := a.store.Search(ctx, queryVector, topK)
	if err != nil {
		return nil, err
	}

	// 转换结果格式
	vectorResults := make([]retriever.VectorSearchResult, len(results))
	for i, result := range results {
		vectorResults[i] = retriever.VectorSearchResult{
			DocID:   fmt.Sprintf("doc_%d", i),
			Content: result,
			Score:   1.0, // 简化处理，实际应该从metadata获取
		}
	}

	return vectorResults, nil
}

// AddDocumentWithSemanticChunking 使用语义分块添加文档
func (r *RAGEnhanced) AddDocumentWithSemanticChunking(ctx context.Context, docPath string) error {
	// 1. 解析文档
	text, err := r.parser.Parse(docPath)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}

	// 2. 使用语义分块
	chunks := r.semanticChunker.Split(text)

	// 3. 向量化并存储
	for i, chunk := range chunks {
		vector, err := r.embedding.Embed(ctx, chunk)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		metadata := map[string]interface{}{
			"source":      docPath,
			"chunk":       i,
			"chunk_type":  "semantic",
		}

		if err := r.store.Add(ctx, vector, chunk, metadata); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}
	}

	// 4. 同时索引到BM25（用于混合检索）
	if r.enableHybrid {
		docs := make([]retriever.Document, len(chunks))
		for i, chunk := range chunks {
			docs[i] = retriever.Document{
				ID:      fmt.Sprintf("%s_chunk_%d", docPath, i),
				Content: chunk,
			}
		}
		r.hybridRetriever.IndexDocuments(docs)
	}

	return nil
}

// RetrieveWithHybrid 混合检索
func (r *RAGEnhanced) RetrieveWithHybrid(ctx context.Context, query string, topK int) ([]string, error) {
	if !r.enableHybrid {
		// 回退到普通向量检索
		return r.Retrieve(ctx, query, topK)
	}

	// 使用混合检索
	results, err := r.hybridRetriever.Search(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	// 提取内容
	contents := make([]string, len(results))
	for i, result := range results {
		contents[i] = result.Content
	}

	return contents, nil
}

// RetrieveWithRerank 检索+重排序
func (r *RAGEnhanced) RetrieveWithRerank(ctx context.Context, query string, topK int) ([]string, error) {
	if !r.enableRerank || r.reranker == nil {
		// 回退到普通检索
		return r.Retrieve(ctx, query, topK)
	}

	// 1. 先检索（获取更多候选）
	candidateK := topK * 3
	var contents []string
	var err error

	if r.enableHybrid {
		contents, err = r.RetrieveWithHybrid(ctx, query, candidateK)
	} else {
		contents, err = r.Retrieve(ctx, query, candidateK)
	}

	if err != nil {
		return nil, err
	}

	// 2. 转换为Document格式
	docs := make([]reranker.Document, len(contents))
	for i, content := range contents {
		docs[i] = reranker.Document{
			ID:      fmt.Sprintf("doc_%d", i),
			Content: content,
		}
	}

	// 3. 重排序
	reranked, err := r.reranker.Rerank(ctx, query, docs)
	if err != nil {
		// 重排序失败，返回原始结果
		return contents[:topK], nil
	}

	// 4. 返回topK
	resultCount := topK
	if resultCount > len(reranked) {
		resultCount = len(reranked)
	}

	results := make([]string, resultCount)
	for i := 0; i < resultCount; i++ {
		results[i] = reranked[i].Content
	}

	return results, nil
}

// RetrieveEnhanced 增强检索（结合混合检索和重排序）
func (r *RAGEnhanced) RetrieveEnhanced(ctx context.Context, query string, topK int) ([]string, error) {
	if r.enableRerank && r.reranker != nil {
		return r.RetrieveWithRerank(ctx, query, topK)
	} else if r.enableHybrid {
		return r.RetrieveWithHybrid(ctx, query, topK)
	} else {
		return r.Retrieve(ctx, query, topK)
	}
}

// 以下是兼容旧接口的方法

// AddDocument 添加文档（使用普通分块）
func (r *RAGEnhanced) AddDocument(ctx context.Context, docPath string) error {
	text, err := r.parser.Parse(docPath)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}

	chunks := r.chunker.Split(text)

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

// Retrieve 检索（普通向量检索）
func (r *RAGEnhanced) Retrieve(ctx context.Context, query string, topK int) ([]string, error) {
	queryVector, err := r.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := r.store.Search(ctx, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

// BuildContext 构建上下文
func (r *RAGEnhanced) BuildContext(ctx context.Context, query string, topK int) (string, error) {
	results, err := r.RetrieveEnhanced(ctx, query, topK)
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

// GetStats 获取统计信息
func (r *RAGEnhanced) GetStats() map[string]interface{} {
	return r.store.Stats()
}

// AddText 添加文本知识
func (r *RAGEnhanced) AddText(ctx context.Context, text string, source string) error {
	// 使用语义分块
	chunks := r.semanticChunker.Split(text)

	// 为每个chunk创建embedding并存储
	for i, chunk := range chunks {
		embedding, err := r.embedding.Embed(ctx, chunk)
		if err != nil {
			return fmt.Errorf("failed to embed chunk: %w", err)
		}

		metadata := map[string]interface{}{
			"source": source,
			"chunk":  i,
		}

		if err := r.store.Add(ctx, embedding, chunk, metadata); err != nil {
			return fmt.Errorf("failed to add chunk to store: %w", err)
		}
	}

	return nil
}

// SetReranker 设置重排序器
func (r *RAGEnhanced) SetReranker(reranker reranker.Reranker) {
	r.reranker = reranker
	r.enableRerank = true
}

// EnableHybridSearch 启用混合检索
func (r *RAGEnhanced) EnableHybridSearch(enable bool) {
	r.enableHybrid = enable
}

// EnableRerank 启用重排序
func (r *RAGEnhanced) EnableRerank(enable bool) {
	r.enableRerank = enable
}
