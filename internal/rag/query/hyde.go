package query

import (
	"context"
	"fmt"
	"strings"
)

// HyDERetriever HyDE (Hypothetical Document Embeddings) 检索器
//
// 策略说明:
//   1. 生成假设性文档/答案
//   2. 对假设性文档进行向量化
//   3. 使用假设性文档的向量进行检索
//   4. 可以同时使用原始查询和假设性文档检索
//
// 优点:
//   - 召回率提升 20-30%
//   - 缩小查询和文档的语义差距
//   - 特别适用于问答场景
//
// 论文:
//   "Precise Zero-Shot Dense Retrieval without Relevance Labels" (2022)
//
// 适用场景:
//   - 问答系统
//   - 事实查询
//   - 需要高召回率的场景
type HyDERetriever struct {
	llm        LLMProvider
	embedding  EmbeddingProvider // 向量化接口
	config     QueryOptimizerConfig
	name       string
	generateOnly bool            // 是否只生成假设文档（不检索）
}

// EmbeddingProvider 向量化提供者接口
type EmbeddingProvider interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, text string) ([]float64, error)
}

// NewHyDERetriever 创建 HyDE 检索器
func NewHyDERetriever(llm LLMProvider, embedding EmbeddingProvider, config QueryOptimizerConfig) (*HyDERetriever, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}
	if embedding == nil {
		return nil, fmt.Errorf("embedding provider is required")
	}

	// 设置默认值
	if config.MaxQueries <= 0 {
		config.MaxQueries = 3
	}
	if config.MinConfidence <= 0 {
		config.MinConfidence = 0.6
	}

	return &HyDERetriever{
		llm:       llm,
		embedding: embedding,
		config:    config,
		name:      "hyde",
	}, nil
}

// Optimize 实现查询优化接口
// 生成假设性文档
func (hyde *HyDERetriever) Optimize(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 1. 生成假设性文档
	hypotheticalDoc, err := hyde.generateHypotheticalDocument(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hypothetical document: %w", err)
	}

	// 2. 向量化假设性文档
	queryVector, err := hyde.embedding.Embed(ctx, hypotheticalDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to embed hypothetical document: %w", err)
	}

	// 3. 构建优化结果
	optimizations := []QueryOptimization{
		{
			Query: hypotheticalDoc,
			Type:  "hyde_document",
			Score: 0.9,
			Metadata: map[string]interface{}{
				"original_query": query,
				"vector":         queryVector, // 存储向量用于检索
				"document_length": len(hypotheticalDoc),
			},
		},
	}

	// 也可以同时返回原始查询
	optimizations = append(optimizations, QueryOptimization{
		Query: query,
		Type:  "original",
		Score: 1.0,
		Metadata: map[string]interface{}{
			"is_original": true,
		},
	})

	return optimizations, nil
}

// generateHypotheticalDocument 生成假设性文档
func (hyde *HyDERetriever) generateHypotheticalDocument(ctx context.Context, query string) (string, error) {
	prompt := hyde.buildHyDEPrompt(query)

	response, err := hyde.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// buildHyDEPrompt 构建 HyDE 提示
func (hyde *HyDERetriever) buildHyDEPrompt(query string) string {
	if hyde.config.Language == "zh" {
		return hyde.buildChinesePrompt(query)
	}
	return hyde.buildEnglishPrompt(query)
}

// buildChinesePrompt 构建中文提示
func (hyde *HyDERetriever) buildChinesePrompt(query string) string {
	return fmt.Sprintf(`请根据以下查询生成一个简短的假设性文档（约 100-200 字）。

要求:
1. 文档应该直接回答查询的问题
2. 内容应该具体、准确、信息丰富
3. 使用专业的语言风格
4. 不要包含"假设"、"可能"等不确定的表述
5. 只输出文档内容，不要解释

查询: %s

假设性文档:`, query)
}

// buildEnglishPrompt 构建英文提示
func (hyde *HyDERetriever) buildEnglishPrompt(query string) string {
	return fmt.Sprintf(`Please generate a short hypothetical document (about 100-200 words) based on the following query.

Requirements:
1. The document should directly answer the query
2. Content should be specific, accurate, and informative
3. Use professional language style
4. Do not include uncertain expressions like "hypothetical", "possible"
5. Output only the document content, no explanation

Query: %s

Hypothetical document:`, query)
}

// Name 返回优化器名称
func (hyde *HyDERetriever) Name() string {
	return hyde.name
}

// Validate 验证配置
func (hyde *HyDERetriever) Validate() error {
	if hyde.llm == nil {
		return fmt.Errorf("LLM provider is required")
	}
	if hyde.embedding == nil {
		return fmt.Errorf("embedding provider is required")
	}
	if hyde.config.MaxQueries <= 0 {
		return fmt.Errorf("max_queries must be positive")
	}
	return nil
}

// GetHypotheticalDocument 获取假设性文档（便捷方法）
func (hyde *HyDERetriever) GetHypotheticalDocument(ctx context.Context, query string) (string, error) {
	return hyde.generateHypotheticalDocument(ctx, query)
}

// GetHypotheticalVector 获取假设性文档的向量（便捷方法）
func (hyde *HyDERetriever) GetHypotheticalVector(ctx context.Context, query string) ([]float64, error) {
	doc, err := hyde.generateHypotheticalDocument(ctx, query)
	if err != nil {
		return nil, err
	}

	return hyde.embedding.Embed(ctx, doc)
}

// HyDEQuery HyDE 查询结果
type HyDEQuery struct {
	OriginalQuery      string   // 原始查询
	HypotheticalDoc    string   // 假设性文档
	HypotheticalVector []float64 // 假设性文档的向量
	CombinedQueries    []string // 组合查询（原始 + 假设）
}

// GenerateHyDEQuery 生成完整的 HyDE 查询
func (hyde *HyDERetriever) GenerateHyDEQuery(ctx context.Context, query string) (*HyDEQuery, error) {
	// 生成假设性文档
	doc, err := hyde.generateHypotheticalDocument(ctx, query)
	if err != nil {
		return nil, err
	}

	// 向量化
	vector, err := hyde.embedding.Embed(ctx, doc)
	if err != nil {
		return nil, err
	}

	// 构建组合查询
	queries := []string{query, doc}

	return &HyDEQuery{
		OriginalQuery:      query,
		HypotheticalDoc:    doc,
		HypotheticalVector: vector,
		CombinedQueries:    queries,
	}, nil
}
