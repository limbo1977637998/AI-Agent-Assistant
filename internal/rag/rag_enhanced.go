package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/internal/rag/adaptive"
	"ai-agent-assistant/internal/rag/chunking"
	"ai-agent-assistant/internal/rag/chunker"
	"ai-agent-assistant/internal/rag/embedding"
	"ai-agent-assistant/internal/rag/eval"
	"ai-agent-assistant/internal/rag/graph"
	"ai-agent-assistant/internal/rag/parser"
	"ai-agent-assistant/internal/rag/query"
	"ai-agent-assistant/internal/rag/reranker"
	"ai-agent-assistant/internal/rag/retriever"
	"ai-agent-assistant/internal/rag/store"
	"ai-agent-assistant/internal/vectordb"
	"ai-agent-assistant/pkg/models"
)

// RAGResult RAG 查询结果
type RAGResult struct {
	Answer  string   // 生成的答案
	Context []string // 检索到的上下文
	Query   string   // 原始查询
}

// RAGEnhanced 增强版RAG系统（支持语义分块、混合检索、重排序）
type RAGEnhanced struct {
	parser         parser.Parser
	chunker        chunker.Chunker
	semanticChunker *chunker.SemanticChunker // 语义分块器 (旧版，保持兼容)
	chunkerManager *chunking.ChunkerManager // 新版分块器管理器
	queryOptimizer *query.QueryOptimizerManager // 查询优化器管理器
	ragasEvaluator *eval.RAGASEvaluator        // RAGAS 评估器
	graphRAG       *graph.GraphRAG              // Graph RAG 检索器
	knowledgeGraph *graph.KnowledgeGraph       // 知识图谱
	selfRAG        *adaptive.SelfReflectiveRAG // Self-RAG 自我反思系统
	queryRouter    *adaptive.QueryRouter       // 查询路由器
	parameterOptimizer *adaptive.ParameterOptimizer // 参数优化器
	abTesting      *adaptive.ABTestingFramework   // A/B 测试框架
	embedding      llm.Model                 // 使用统一的Model接口
	store          store.VectorStore
	hybridRetriever *retriever.HybridRetriever // 混合检索器
	reranker       reranker.Reranker            // 重排序器
	crossEncoder   *reranker.CrossEncoderReranker // CrossEncoder 重排序器
	config         *config.Config
	enableHybrid   bool                        // 是否启用混合检索
	enableRerank   bool                        // 是否启用重排序
	enableQueryOpt bool                        // 是否启用查询优化
	enableGraphRAG  bool                       // 是否启用 Graph RAG
	enableSelfRAG   bool                       // 是否启用 Self-RAG
	enableAdaptive  bool                       // 是否启用自适应路由
	currentChunker chunking.ChunkerStrategy    // 当前使用的分块器 (新版)
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

	// 2.5 初始化新版分块器管理器
	chunkerManager := chunking.NewChunkerManager()

	// 2.6 初始化查询优化器管理器
	queryOptimizer := query.NewQueryOptimizerManager()

	// 2.7 初始化 RAGAS 评估器
	var ragasEvaluator *eval.RAGASEvaluator
	if embeddingModel != nil {
		// 创建 LLMProvider 适配器
		llmProvider := &ModelLLMAdapter{model: embeddingModel}
		ragasEvaluator, _ = eval.NewRAGASEvaluator(llmProvider)
	}

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
		parser:             p,
		chunker:            *c,
		semanticChunker:    semanticChunker,
		chunkerManager:     chunkerManager,
		queryOptimizer:     queryOptimizer,
		ragasEvaluator:     ragasEvaluator,
		graphRAG:           nil, // 可选，需要单独初始化
		knowledgeGraph:     nil,
		selfRAG:            nil, // 可选，需要单独初始化
		queryRouter:        nil, // 可选，需要单独初始化
		parameterOptimizer: nil, // 可选，需要单独初始化
		abTesting:          nil, // 可选，需要单独初始化
		embedding:          embeddingModel,
		store:              vs,
		hybridRetriever:    hybridRetriever,
		reranker:           r,
		crossEncoder:       nil, // 可选，需要单独初始化
		config:             cfg,
		enableHybrid:       cfg.RAG.EnableHybridSearch,
		enableRerank:       true,
		enableQueryOpt:     false, // 默认关闭查询优化
		enableGraphRAG:     false, // 默认关闭 Graph RAG
		enableSelfRAG:      false, // 默认关闭 Self-RAG
		enableAdaptive:     false, // 默认关闭自适应路由
		currentChunker:     nil,  // 默认使用旧版分块器
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

// ==================== 新版分块器系统方法 ====================

// SetChunker 设置当前使用的分块器 (新版)
func (r *RAGEnhanced) SetChunker(chunkerType string, config interface{}) error {
	chunker, err := r.chunkerManager.CreateChunker(chunkerType, config)
	if err != nil {
		return fmt.Errorf("failed to create chunker: %w", err)
	}

	r.currentChunker = chunker
	return nil
}

// GetChunkerManager 获取分块器管理器
func (r *RAGEnhanced) GetChunkerManager() *chunking.ChunkerManager {
	return r.chunkerManager
}

// AddDocumentWithChunker 使用指定分块器添加文档 (新版)
func (r *RAGEnhanced) AddDocumentWithChunker(ctx context.Context, docPath string) error {
	if r.currentChunker == nil {
		return fmt.Errorf("no chunker set, please call SetChunker first")
	}

	// 1. 解析文档
	text, err := r.parser.Parse(docPath)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}

	// 2. 使用当前分块器分块
	chunks, err := r.currentChunker.Split(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to split document: %w", err)
	}

	// 3. 向量化并存储
	for i, chunk := range chunks {
		vector, err := r.embedding.Embed(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		metadata := map[string]interface{}{
			"source":      docPath,
			"chunk":       chunk.Metadata.Index,
			"chunk_type":  chunk.Metadata.ChunkType,
			"start_pos":   chunk.Metadata.StartPos,
			"end_pos":     chunk.Metadata.EndPos,
			"token_count": chunk.Metadata.TokenCount,
		}

		// 添加额外的元数据
		if chunk.Metadata.AdditionalMetadata != nil {
			for k, v := range chunk.Metadata.AdditionalMetadata {
				metadata[k] = v
			}
		}

		if err := r.store.Add(ctx, vector, chunk.Content, metadata); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}
	}

	// 4. 同时索引到BM25（用于混合检索）
	if r.enableHybrid {
		docs := make([]retriever.Document, len(chunks))
		for i, chunk := range chunks {
			docs[i] = retriever.Document{
				ID:      fmt.Sprintf("%s_chunk_%d", docPath, chunk.Metadata.Index),
				Content: chunk.Content,
			}
		}
		r.hybridRetriever.IndexDocuments(docs)
	}

	return nil
}

// AddDocumentWithRecursiveChunker 使用递归分块器添加文档
func (r *RAGEnhanced) AddDocumentWithRecursiveChunker(ctx context.Context, docPath string, chunkSize, overlap int) error {
	cfg := chunking.ChunkerConfig{
		ChunkSize:     chunkSize,
		ChunkOverlap:  overlap,
		MinChunkSize:  chunkSize / 10,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	if err := r.SetChunker("recursive", cfg); err != nil {
		return err
	}

	return r.AddDocumentWithChunker(ctx, docPath)
}

// AddDocumentWithSmallToBigChunker 使用小到大分块器添加文档
func (r *RAGEnhanced) AddDocumentWithSmallToBigChunker(ctx context.Context, docPath string, smallSize, bigSize, overlap int) error {
	smallConfig := chunking.ChunkerConfig{
		ChunkSize:     smallSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	bigConfig := chunking.ChunkerConfig{
		ChunkSize:     bigSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	config := map[string]interface{}{
		"small":       smallConfig,
		"big":         bigConfig,
		"parent_merge": 3,
	}

	if err := r.SetChunker("small_to_big", config); err != nil {
		return err
	}

	return r.AddDocumentWithChunker(ctx, docPath)
}

// AddDocumentWithParentDocumentChunker 使用父文档分块器添加文档
func (r *RAGEnhanced) AddDocumentWithParentDocumentChunker(ctx context.Context, docPath string, parentSize, childSize, overlap int) error {
	parentConfig := chunking.ChunkerConfig{
		ChunkSize:     parentSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	childConfig := chunking.ChunkerConfig{
		ChunkSize:     childSize,
		ChunkOverlap:  overlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
		KeepSeparator: false,
	}

	config := map[string]interface{}{
		"parent":          parentConfig,
		"child":           childConfig,
		"child_per_parent": 5,
	}

	if err := r.SetChunker("parent_document", config); err != nil {
		return err
	}

	return r.AddDocumentWithChunker(ctx, docPath)
}

// ListAvailableChunkers 列出所有可用的分块器类型
func (r *RAGEnhanced) ListAvailableChunkers() []string {
	return r.chunkerManager.ListAvailableChunkers()
}

// GetChunkerDescription 获取分块器描述
func (r *RAGEnhanced) GetChunkerDescription(chunkerType string) string {
	return r.chunkerManager.GetChunkerDescription(chunkerType)
}

// ==================== 查询优化系统方法 ====================

// ModelLLMAdapter LLM Model 适配器
// 将 llm.Model 适配为 query.LLMProvider 接口
type ModelLLMAdapter struct {
	model llm.Model
}

// Generate 实现 LLMProvider 接口
func (adapter *ModelLLMAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	response, err := adapter.model.Chat(ctx, messages)
	if err != nil {
		return "", err
	}
	return response, nil
}

// SetQueryOptimizer 设置查询优化器
func (r *RAGEnhanced) SetQueryOptimizer(optimizerName string, optimizerType string) error {
	if r.queryOptimizer == nil {
		return fmt.Errorf("query optimizer manager not initialized")
	}

	// 创建 LLMProvider 适配器
	llmProvider := &ModelLLMAdapter{model: r.embedding}

	config := query.DefaultQueryOptimizerConfig()
	return r.queryOptimizer.CreateOptimizer(optimizerName, optimizerType, llmProvider, nil, config)
}

// QueryWithOptimization 使用查询优化进行检索
func (r *RAGEnhanced) QueryWithOptimization(ctx context.Context, query string, optimizerName string, topK int) (*RAGResult, error) {
	if !r.enableQueryOpt {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 优化查询
	optimizations, err := r.queryOptimizer.Optimize(ctx, optimizerName, query)
	if err != nil {
		return nil, fmt.Errorf("query optimization failed: %w", err)
	}

	// 2. 使用所有优化后的查询检索
	allContexts := make([]string, 0)
	for _, opt := range optimizations {
		contexts, err := r.RetrieveEnhanced(ctx, opt.Query, topK)
		if err != nil {
			continue
		}
		allContexts = append(allContexts, contexts...)
	}

	// 3. 去重并限制数量
	uniqueContexts := r.deduplicateStrings(allContexts)
	if len(uniqueContexts) > topK {
		uniqueContexts = uniqueContexts[:topK]
	}

	// 4. 构建上下文并生成答案
	context := strings.Join(uniqueContexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", context, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: uniqueContexts,
		Query:   query,
	}, nil
}

// deduplicateStrings 去重字符串切片
func (r *RAGEnhanced) deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(strs))

	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// EnableQueryOptimization 启用/禁用查询优化
func (r *RAGEnhanced) EnableQueryOptimization(enable bool) {
	r.enableQueryOpt = enable
}

// ==================== CrossEncoder 重排序方法 ====================

// SetCrossEncoder 设置 CrossEncoder 重排序器
func (r *RAGEnhanced) SetCrossEncoder(apiKey, baseURL, model string) error {
	crossEncoder, err := reranker.NewCrossEncoderReranker(apiKey, baseURL, model)
	if err != nil {
		return err
	}

	r.crossEncoder = crossEncoder
	r.enableRerank = true
	return nil
}

// QueryWithCrossEncoder 使用 CrossEncoder 重排序的查询
func (r *RAGEnhanced) QueryWithCrossEncoder(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if r.crossEncoder == nil {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 先检索更多候选 (topK * 3)
	candidateK := topK * 3
	contexts, err := r.RetrieveEnhanced(ctx, query, candidateK)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// 2. 转换为 Document 格式
	docs := make([]reranker.Document, len(contexts))
	for i, ctx := range contexts {
		docs[i] = reranker.Document{
			ID:      fmt.Sprintf("doc_%d", i),
			Content: ctx,
		}
	}

	// 3. 使用 CrossEncoder 重排序
	reranked, err := r.crossEncoder.Rerank(ctx, query, docs)
	if err != nil {
		// 重排序失败，返回原始结果
		return r.QueryWithContext(ctx, query, topK)
	}

	// 4. 取 Top-K
	finalContexts := make([]string, 0, topK)
	for i := 0; i < topK && i < len(reranked); i++ {
		finalContexts = append(finalContexts, reranked[i].Content)
	}

	// 5. 生成答案
	context := strings.Join(finalContexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", context, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: finalContexts,
		Query:   query,
	}, nil
}

// ==================== RAGAS 评估方法 ====================

// EvaluateRAG 评估 RAG 系统性能
func (r *RAGEnhanced) EvaluateRAG(ctx context.Context, query string, groundTruth string) (*eval.RAGASResult, error) {
	if r.ragasEvaluator == nil {
		return nil, fmt.Errorf("RAGAS evaluator not initialized")
	}

	// 1. 检索上下文
	topK := 5
	contexts, err := r.RetrieveEnhanced(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// 2. 生成答案
	answer, err := r.QueryWithContext(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// 3. 评估
	result, err := r.ragasEvaluator.Evaluate(ctx, query, contexts, answer.Answer, groundTruth)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	return result, nil
}

// EvaluateRAGBatch 批量评估 RAG 系统
func (r *RAGEnhanced) EvaluateRAGBatch(ctx context.Context, queries []string, groundTruths []string) ([]*eval.RAGASResult, string, error) {
	if r.ragasEvaluator == nil {
		return nil, "", fmt.Errorf("RAGAS evaluator not initialized")
	}

	if len(queries) != len(groundTruths) {
		return nil, "", fmt.Errorf("queries and groundTruths count mismatch")
	}

	results := make([]*eval.RAGASResult, len(queries))
	contextsList := make([][]string, len(queries))
	answers := make([]string, len(queries))

	// 1. 执行查询并生成答案
	for i, query := range queries {
		topK := 5
		contexts, err := r.RetrieveEnhanced(ctx, query, topK)
		if err != nil {
			return nil, "", fmt.Errorf("retrieval failed for query %d: %w", i, err)
		}
		contextsList[i] = contexts

		answer, err := r.QueryWithContext(ctx, query, topK)
		if err != nil {
			return nil, "", fmt.Errorf("generation failed for query %d: %w", i, err)
		}
		answers[i] = answer.Answer
	}

	// 2. 批量评估
	results, err := r.ragasEvaluator.EvaluateBatch(ctx, queries, contextsList, answers, groundTruths)
	if err != nil {
		return nil, "", fmt.Errorf("batch evaluation failed: %w", err)
	}

	// 3. 生成报告
	report := r.ragasEvaluator.GenerateReport(results)

	return results, report, nil
}

// GetQueryOptimizer 获取查询优化器管理器
func (r *RAGEnhanced) GetQueryOptimizer() *query.QueryOptimizerManager {
	return r.queryOptimizer
}

// GetRAGASEvaluator 获取 RAGAS 评估器
func (r *RAGEnhanced) GetRAGASEvaluator() *eval.RAGASEvaluator {
	return r.ragasEvaluator
}

// QueryWithContext 使用上下文查询（新增方法）
func (r *RAGEnhanced) QueryWithContext(ctx context.Context, query string, topK int) (*RAGResult, error) {
	// 1. 检索上下文
	contexts, err := r.RetrieveEnhanced(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// 2. 构建提示
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	// 3. 生成答案
	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// ==================== Graph RAG 方法 ====================

// InitGraphRAG 初始化 Graph RAG
func (r *RAGEnhanced) InitGraphRAG() error {
	if r.embedding == nil {
		return fmt.Errorf("embedding model is required for Graph RAG")
	}

	// 创建 LLMProvider 适配器
	llmProvider := &ModelLLMAdapter{model: r.embedding}

	config := graph.DefaultGraphRAGConfig()
	graphRAG, err := graph.NewGraphRAG(llmProvider, config)
	if err != nil {
		return fmt.Errorf("failed to create Graph RAG: %w", err)
	}

	r.graphRAG = graphRAG
	r.enableGraphRAG = true

	return nil
}

// BuildKnowledgeGraph 构建知识图谱
func (r *RAGEnhanced) BuildKnowledgeGraph(ctx context.Context, documents []string) error {
	if r.graphRAG == nil {
		return fmt.Errorf("Graph RAG not initialized, call InitGraphRAG first")
	}

	graph, err := r.graphRAG.BuildGraph(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to build knowledge graph: %w", err)
	}

	r.knowledgeGraph = graph
	return nil
}

// QueryWithGraphRAG 使用 Graph RAG 检索
func (r *RAGEnhanced) QueryWithGraphRAG(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if !r.enableGraphRAG || r.graphRAG == nil || r.knowledgeGraph == nil {
		// 回退到普通检索
		return r.QueryWithContext(ctx, query, topK)
	}

	// 社区检索（结合全局和局部）
	contexts, err := r.graphRAG.CommunitySearch(ctx, r.knowledgeGraph, query, topK)
	if err != nil {
		return nil, fmt.Errorf("Graph RAG search failed: %w", err)
	}

	if len(contexts) == 0 {
		// 如果 Graph RAG 没有结果，回退到普通检索
		return r.QueryWithContext(ctx, query, topK)
	}

	// 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下知识图谱信息回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// QueryGlobalGraph 使用全局图检索
func (r *RAGEnhanced) QueryGlobalGraph(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if r.graphRAG == nil || r.knowledgeGraph == nil {
		return nil, fmt.Errorf("knowledge graph not built")
	}

	contexts, err := r.graphRAG.GlobalSearch(ctx, r.knowledgeGraph, query, topK)
	if err != nil {
		return nil, fmt.Errorf("global graph search failed: %w", err)
	}

	if len(contexts) == 0 {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下全局信息回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// QueryLocalGraph 使用局部图检索
func (r *RAGEnhanced) QueryLocalGraph(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if r.graphRAG == nil || r.knowledgeGraph == nil {
		return nil, fmt.Errorf("knowledge graph not built")
	}

	contexts, err := r.graphRAG.LocalSearch(ctx, r.knowledgeGraph, query, topK)
	if err != nil {
		return nil, fmt.Errorf("local graph search failed: %w", err)
	}

	if len(contexts) == 0 {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下实体关系信息回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// GetGraphHierarchy 获取图谱层次结构
func (r *RAGEnhanced) GetGraphHierarchy(ctx context.Context) *graph.CommunityHierarchy {
	if r.graphRAG == nil || r.knowledgeGraph == nil {
		return nil
	}

	return r.graphRAG.GetHierarchicalSummaries(ctx, r.knowledgeGraph)
}

// EnableGraphRAG 启用/禁用 Graph RAG
func (r *RAGEnhanced) EnableGraphRAG(enable bool) {
	r.enableGraphRAG = enable
}

// GetGraphRAG 获取 Graph RAG 检索器
func (r *RAGEnhanced) GetGraphRAG() *graph.GraphRAG {
	return r.graphRAG
}

// GetKnowledgeGraph 获取知识图谱
func (r *RAGEnhanced) GetKnowledgeGraph() *graph.KnowledgeGraph {
	return r.knowledgeGraph
}

// ==================== Self-RAG 自适应方法 ====================

// InitSelfRAG 初始化 Self-RAG
func (r *RAGEnhanced) InitSelfRAG() error {
	if r.embedding == nil {
		return fmt.Errorf("embedding model is required for Self-RAG")
	}

	// 创建 LLMProvider 适配器
	llmProvider := &ModelLLMAdapter{model: r.embedding}

	config := adaptive.DefaultSelfRAGConfig()
	selfRAG, err := adaptive.NewSelfReflectiveRAG(llmProvider, config)
	if err != nil {
		return fmt.Errorf("failed to create Self-RAG: %w", err)
	}

	r.selfRAG = selfRAG
	r.enableSelfRAG = true

	return nil
}

// QueryWithSelfRAG 使用 Self-RAG 进行自我反思检索
func (r *RAGEnhanced) QueryWithSelfRAG(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if !r.enableSelfRAG || r.selfRAG == nil {
		// 回退到普通检索
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 初始检索
	contexts, err := r.RetrieveEnhanced(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// 2. 评估检索质量
	score, err := r.selfRAG.EvaluateRetrieval(ctx, query, contexts)
	if err != nil {
		// 评估失败，继续使用当前结果
		score = 0.5
	}

	// 3. 判断是否需要重新检索
	maxRetries := 2
	retryCount := 0

	for r.selfRAG.ShouldRetrieveMore(score) && retryCount < maxRetries {
		// 调整 Top-K
		adjustedTopK := r.selfRAG.AdjustTopK(topK, score)

		// 重新检索
		newContexts, err := r.RetrieveEnhanced(ctx, query, adjustedTopK)
		if err != nil {
			break
		}

		// 合并上下文
		contexts = append(contexts, newContexts...)

		// 重新评估
		score, err = r.selfRAG.EvaluateRetrieval(ctx, query, contexts)
		if err != nil {
			break
		}

		retryCount++
	}

	// 4. 去重并限制数量
	uniqueContexts := r.deduplicateStrings(contexts)
	if len(uniqueContexts) > topK {
		uniqueContexts = uniqueContexts[:topK]
	}

	// 5. 生成答案（可选：包含反思）
	reflection := ""
	if score < 0.7 {
		reflection, _ = r.selfRAG.GenerateReflection(ctx, query, score, uniqueContexts)
	}

	// 6. 构建提示
	contextText := strings.Join(uniqueContexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	if reflection != "" {
		prompt += fmt.Sprintf("\n\n改进建议:\n%s", reflection)
	}

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Context: uniqueContexts,
		Query:   query,
	}, nil
}

// EnableSelfRAG 启用/禁用 Self-RAG
func (r *RAGEnhanced) EnableSelfRAG(enable bool) {
	r.enableSelfRAG = enable
}

// GetSelfRAG 获取 Self-RAG 系统
func (r *RAGEnhanced) GetSelfRAG() *adaptive.SelfReflectiveRAG {
	return r.selfRAG
}

// ==================== 查询路由器方法 ====================

// InitQueryRouter 初始化查询路由器
func (r *RAGEnhanced) InitQueryRouter(config adaptive.RouterConfig) error {
	if r.embedding == nil {
		return fmt.Errorf("embedding model is required for query router")
	}

	// 创建 LLMProvider 适配器
	llmProvider := &ModelLLMAdapter{model: r.embedding}

	router, err := adaptive.NewQueryRouter(llmProvider, config)
	if err != nil {
		return fmt.Errorf("failed to create query router: %w", err)
	}

	r.queryRouter = router
	r.enableAdaptive = true

	return nil
}

// QueryWithRouting 使用自适应路由检索
func (r *RAGEnhanced) QueryWithRouting(ctx context.Context, query string, topK int) (*RAGResult, error) {
	if !r.enableAdaptive || r.queryRouter == nil {
		// 回退到普通检索
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 选择策略
	strategy, err := r.queryRouter.SelectStrategy(ctx, query)
	if err != nil {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 2. 根据策略检索
	var contexts []string

	switch strategy {
	case "vector":
		contexts, _ = r.Retrieve(ctx, query, topK)
	case "hybrid":
		contexts, _ = r.RetrieveWithHybrid(ctx, query, topK)
	case "graph_rag":
		if r.enableGraphRAG && r.knowledgeGraph != nil {
			contexts, _ = r.graphRAG.CommunitySearch(ctx, r.knowledgeGraph, query, topK)
		} else {
			contexts, _ = r.RetrieveEnhanced(ctx, query, topK)
		}
	case "hyde":
		// 使用 HyDE 优化
		optimizations, _ := r.queryOptimizer.Optimize(ctx, "hyde", query)
		if len(optimizations) > 0 {
			contexts, _ = r.RetrieveEnhanced(ctx, optimizations[0].Query, topK)
		} else {
			contexts, _ = r.RetrieveEnhanced(ctx, query, topK)
		}
	default:
		contexts, _ = r.RetrieveEnhanced(ctx, query, topK)
	}

	// 3. 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 4. 记录性能数据
	result := &adaptive.RAGExecutionResult{
		Strategy:      strategy,
		Query:         query,
		Answer:        answer,
		Contexts:      contexts,
		Score:         0.7, // 简化：默认得分
		Latency:       100, // 简化：默认延迟
		UserFeedback:  0.7, // 简化：默认反馈
		Success:       err == nil,
	}
	r.queryRouter.RecordFeedback(ctx, query, strategy, result)

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// GetQueryRouter 获取查询路由器
func (r *RAGEnhanced) GetQueryRouter() *adaptive.QueryRouter {
	return r.queryRouter
}

// EnableAdaptiveRouting 启用/禁用自适应路由
func (r *RAGEnhanced) EnableAdaptiveRouting(enable bool) {
	r.enableAdaptive = enable
}

// ==================== 参数优化器方法 ====================

// InitParameterOptimizer 初始化参数优化器
func (r *RAGEnhanced) InitParameterOptimizer(config adaptive.OptimizerConfig) error {
	if r.embedding == nil {
		return fmt.Errorf("embedding model is required for parameter optimizer")
	}

	// 创建 LLMProvider 适配器
	llmProvider := &ModelLLMAdapter{model: r.embedding}

	optimizer, err := adaptive.NewParameterOptimizer(llmProvider, config)
	if err != nil {
		return fmt.Errorf("failed to create parameter optimizer: %w", err)
	}

	r.parameterOptimizer = optimizer

	return nil
}

// QueryWithOptimizedParams 使用优化参数检索
func (r *RAGEnhanced) QueryWithOptimizedParams(ctx context.Context, query, strategy string, topK int) (*RAGResult, error) {
	if r.parameterOptimizer == nil {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 获取优化参数
	params, err := r.parameterOptimizer.OptimizeParameters(ctx, strategy)
	if err != nil {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 2. 应用参数（简化：只支持 top_k）
	optimizedTopK := topK
	if tk, ok := params["top_k"].(int); ok {
		optimizedTopK = tk
	}

	// 3. 检索
	contexts, err := r.RetrieveEnhanced(ctx, query, optimizedTopK)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// 4. 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 5. 记录性能
	result := &adaptive.RAGExecutionResult{
		Strategy:      strategy,
		Query:         query,
		Answer:        answer,
		Contexts:      contexts,
		Score:         0.7,
		Latency:       time.Now().UnixMilli(),
		UserFeedback:  0.7,
		Success:       err == nil,
	}
	r.parameterOptimizer.RecordPerformance(ctx, strategy, result)

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// GetParameterOptimizer 获取参数优化器
func (r *RAGEnhanced) GetParameterOptimizer() *adaptive.ParameterOptimizer {
	return r.parameterOptimizer
}

// ==================== A/B 测试方法 ====================

// InitABTesting 初始化 A/B 测试框架
func (r *RAGEnhanced) InitABTesting(config adaptive.ABTestConfig) error {
	r.abTesting = adaptive.NewABTestingFramework(config)
	return nil
}

// CreateABTest 创建 A/B 测试
func (r *RAGEnhanced) CreateABTest(ctx context.Context, name, description string, variants []*adaptive.Variant) error {
	if r.abTesting == nil {
		return fmt.Errorf("AB testing framework not initialized")
	}

	return r.abTesting.CreateExperiment(ctx, name, description, variants)
}

// QueryWithABTest 使用 A/B 测试检索
func (r *RAGEnhanced) QueryWithABTest(ctx context.Context, experimentName, query string, topK int) (*RAGResult, error) {
	if r.abTesting == nil {
		return r.QueryWithContext(ctx, query, topK)
	}

	// 1. 选择变体
	variant, err := r.abTesting.SelectVariant(ctx, experimentName)
	if err != nil {
		return nil, fmt.Errorf("failed to select variant: %w", err)
	}

	// 2. 根据变体检索
	startTime := time.Now()
	var contexts []string

	switch variant.Strategy {
	case "vector":
		contexts, _ = r.Retrieve(ctx, query, topK)
	case "hybrid":
		contexts, _ = r.RetrieveWithHybrid(ctx, query, topK)
	case "graph_rag":
		if r.enableGraphRAG && r.knowledgeGraph != nil {
			contexts, _ = r.graphRAG.CommunitySearch(ctx, r.knowledgeGraph, query, topK)
		} else {
			contexts, _ = r.RetrieveEnhanced(ctx, query, topK)
		}
	default:
		contexts, _ = r.RetrieveEnhanced(ctx, query, topK)
	}

	latency := time.Now().Sub(startTime).Milliseconds()

	// 3. 生成答案
	contextText := strings.Join(contexts, "\n\n")
	prompt := fmt.Sprintf("基于以下上下文回答问题:\n\n上下文:\n%s\n\n问题: %s\n\n回答:", contextText, query)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}
	answer, err := r.embedding.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 4. 记录结果
	result := &adaptive.VariantResult{
		Query:        query,
		Contexts:     contexts,
		Answer:       answer,
		Score:        0.7, // 简化：默认得分
		Latency:      latency,
		UserFeedback: 0.7, // 简化：默认反馈
	}
	r.abTesting.RecordResult(ctx, experimentName, variant.Name, result)

	return &RAGResult{
		Answer:  answer,
		Context: contexts,
		Query:   query,
	}, nil
}

// GetABTestReport 获取 A/B 测试报告
func (r *RAGEnhanced) GetABTestReport(name string) string {
	if r.abTesting == nil {
		return ""
	}

	return r.abTesting.GenerateReport(name)
}

// GetABTestingFramework 获取 A/B 测试框架
func (r *RAGEnhanced) GetABTestingFramework() *adaptive.ABTestingFramework {
	return r.abTesting
}
