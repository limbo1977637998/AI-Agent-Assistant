package query

import (
	"context"
	"fmt"
	"strings"
)

// AdvancedQueryExpander 高级查询扩展器
//
// 支持多种扩展策略:
//   1. Synonym Expansion (同义词扩展)
//   2. Semantic Expansion (语义扩展)
//   3. Generative Expansion (生成式扩展)
//   4. HyDE Expansion (假设文档扩展)
//
// 优点:
//   - 提高召回率
//   - 覆盖多种表达方式
//   - 语义理解
//   - 适应不同查询风格
type AdvancedQueryExpander struct {
	llm              LLMProvider
	strategy         string // synonym, semantic, generative, hyde, hybrid
	maxExpansions    int
	synonymDict      map[string][]string
	embeddings       map[string][]float32 // 词向量缓存
}

// NewAdvancedQueryExpander 创建高级查询扩展器
func NewAdvancedQueryExpander(llm LLMProvider, strategy string) (*AdvancedQueryExpander, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	expander := &AdvancedQueryExpander{
		llm:           llm,
		strategy:      strategy,
		maxExpansions: 5,
		synonymDict:   make(map[string][]string),
		embeddings:    make(map[string][]float32),
	}

	// 初始化同义词词典
	expander.initSynonymDict()

	return expander, nil
}

// initSynonymDict 初始化同义词词典
func (e *AdvancedQueryExpander) initSynonymDict() {
	// AI 相关同义词
	e.synonymDict["人工智能"] = []string{"AI", "机器智能", "智能系统"}
	e.synonymDict["机器学习"] = []string{"ML", "Machine Learning", "机器学习算法"}
	e.synonymDict["深度学习"] = []string{"DL", "Deep Learning", "深度神经网络"}
	e.synonymDict["自然语言处理"] = []string{"NLP", "Natural Language Processing", "文本处理"}
	e.synonymDict["计算机视觉"] = []string{"CV", "Computer Vision", "图像识别"}

	// 通用同义词
	e.synonymDict["使用"] = []string{"应用", "采用", "运用", "利用"}
	e.synonymDict["方法"] = []string{"方式", "途径", "手段", "策略"}
	e.synonymDict["问题"] = []string{"疑问", "难题", "困难", "挑战"}
	e.synonymDict["分析"] = []string{"研究", "探讨", "解析", "剖析"}
}

// Expand 扩展查询
func (e *AdvancedQueryExpander) Expand(ctx context.Context, query string) ([]QueryOptimization, error) {
	switch e.strategy {
	case "synonym":
		return e.synonymExpansion(query), nil
	case "semantic":
		return e.semanticExpansion(ctx, query)
	case "generative":
		return e.generativeExpansion(ctx, query)
	case "hyde":
		return e.hydeExpansion(ctx, query)
	case "hybrid":
		return e.hybridExpansion(ctx, query)
	default:
		return e.hybridExpansion(ctx, query)
	}
}

// synonymExpansion 同义词扩展
func (e *AdvancedQueryExpander) synonymExpansion(query string) []QueryOptimization {
	expansions := []QueryOptimization{
		{
			Query: query,
			Type:  "original",
			Score: 1.0,
		},
	}

	// 提取查询中的关键词
	words := e.tokenize(query)

	// 生成变体
	variants := e.generateSynonymVariants(query, words)

	for i, variant := range variants {
		if i >= e.maxExpansions {
			break
		}

		expansions = append(expansions, QueryOptimization{
			Query: variant,
			Type:  "synonym_expansion",
			Score: 0.9 - float64(i)*0.1,
			Metadata: map[string]interface{}{
				"original_query": query,
				"expansion_method": "synonym",
				"variant_index": i,
			},
		})
	}

	return expansions
}

// semanticExpansion 语义扩展（基于 Embedding 相似度）
func (e *AdvancedQueryExpander) semanticExpansion(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 简化实现：使用 LLM 生成语义相关的查询
	prompt := fmt.Sprintf(`请为以下查询生成 3-5 个语义相关的查询变体。

原始查询: %s

要求:
1. 保持原意
2. 使用不同的表达方式
3. 涵盖相关的概念和术语
4. 每行一个查询

相关查询:`, query)

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	queries := e.parseQueries(response)

	optimizations := make([]QueryOptimization, 0, len(queries)+1)
	optimizations = append(optimizations, QueryOptimization{
		Query: query,
		Type:  "original",
		Score: 1.0,
	})

	for i, q := range queries {
		if i >= e.maxExpansions {
			break
		}

		optimizations = append(optimizations, QueryOptimization{
			Query: q,
			Type:  "semantic_expansion",
			Score: 0.95 - float64(i)*0.1,
			Metadata: map[string]interface{}{
				"original_query": query,
				"expansion_method": "semantic",
				"variant_index": i,
			},
		})
	}

	return optimizations, nil
}

// generativeExpansion 生成式扩展（使用 LLM 生成多样化查询）
func (e *AdvancedQueryExpander) generativeExpansion(ctx context.Context, query string) ([]QueryOptimization, error) {
	intent := e.detectQueryIntent(query)

	prompt := fmt.Sprintf(`请基于查询意图，生成多个不同角度的扩展查询。

原始查询: %s
查询意图: %s

要求:
1. 从不同角度扩展（定义、应用、原理、对比等）
2. 每个扩展查询应该独立且有价值
3. 每行一个查询
4. 生成 5 个扩展查询

扩展查询:`, query, intent)

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	queries := e.parseQueries(response)

	optimizations := make([]QueryOptimization, 0, len(queries)+1)
	optimizations = append(optimizations, QueryOptimization{
		Query: query,
		Type:  "original",
		Score: 1.0,
	})

	for i, q := range queries {
		if i >= e.maxExpansions {
			break
		}

		optimizations = append(optimizations, QueryOptimization{
			Query: q,
			Type:  "generative_expansion",
			Score: 0.85 - float64(i)*0.05,
			Metadata: map[string]interface{}{
				"original_query": query,
				"expansion_method": "generative",
				"intent": intent,
				"variant_index": i,
			},
		})
	}

	return optimizations, nil
}

// hydeExpansion HyDE 扩展（假设文档嵌入）
func (e *AdvancedQueryExpander) hydeExpansion(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 生成假设文档
	prompt := fmt.Sprintf(`请为以下查询生成一个理想的答案文档。

查询: %s

要求:
1. 详细、准确地回答查询
2. 包含关键信息和细节
3. 使用专业的语言
4. 100-200 字

假设文档:`, query)

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	hypotheticalDoc := strings.TrimSpace(response)

	// 从假设文档中提取关键短语作为扩展查询
	phrases := e.extractKeyPhrases(hypotheticalDoc)

	optimizations := []QueryOptimization{
		{
			Query: query,
			Type:  "original",
			Score: 1.0,
		},
		{
			Query: hypotheticalDoc,
			Type:  "hyde_document",
			Score: 0.95,
			Metadata: map[string]interface{}{
				"original_query": query,
				"hypothetical_doc": true,
			},
		},
	}

	for i, phrase := range phrases {
		if i >= e.maxExpansions-1 {
			break
		}

		optimizations = append(optimizations, QueryOptimization{
			Query: phrase,
			Type:  "hyde_phrase",
			Score: 0.8 - float64(i)*0.1,
			Metadata: map[string]interface{}{
				"original_query": query,
				"expansion_method": "hyde",
				"phrase_index": i,
			},
		})
	}

	return optimizations, nil
}

// hybridExpansion 混合扩展（结合多种策略）
func (e *AdvancedQueryExpander) hybridExpansion(ctx context.Context, query string) ([]QueryOptimization, error) {
	allOptimizations := make([]QueryOptimization, 0)

	// 1. 同义词扩展
	synonymOpts := e.synonymExpansion(query)
	allOptimizations = append(allOptimizations, synonymOpts...)

	// 2. 语义扩展（如果同义词扩展不够）
	if len(allOptimizations) < 3 {
		semanticOpts, err := e.semanticExpansion(ctx, query)
		if err == nil {
			// 去重
			for _, opt := range semanticOpts {
				if !e.containsQuery(allOptimizations, opt.Query) {
					allOptimizations = append(allOptimizations, opt)
				}
			}
		}
	}

	// 3. HyDE 扩展（作为补充）
	hydeOpts, err := e.hydeExpansion(ctx, query)
	if err == nil {
		for _, opt := range hydeOpts {
			if !e.containsQuery(allOptimizations, opt.Query) {
				allOptimizations = append(allOptimizations, opt)
			}
		}
	}

	// 限制数量
	if len(allOptimizations) > e.maxExpansions {
		allOptimizations = allOptimizations[:e.maxExpansions]
	}

	return allOptimizations, nil
}

// tokenize 分词（简化实现）
func (e *AdvancedQueryExpander) tokenize(text string) []string {
	// 简单按空格和标点分割
	words := strings.Fields(text)
	return words
}

// generateSynonymVariants 生成同义词变体
func (e *AdvancedQueryExpander) generateSynonymVariants(query string, words []string) []string {
	variants := []string{query}

	for _, word := range words {
		if synonyms, ok := e.synonymDict[word]; ok {
			for _, synonym := range synonyms {
				variant := strings.ReplaceAll(query, word, synonym)
				if variant != query {
					variants = append(variants, variant)
				}
			}
		}
	}

	return variants
}

// detectQueryIntent 检测查询意图
func (e *AdvancedQueryExpander) detectQueryIntent(query string) string {
	if strings.Contains(query, "什么") || strings.Contains(query, "定义") {
		return "definition"
	}
	if strings.Contains(query, "如何") || strings.Contains(query, "怎么") {
		return "procedure"
	}
	if strings.Contains(query, "为什么") {
		return "reasoning"
	}
	if strings.Contains(query, "区别") || strings.Contains(query, "对比") {
		return "comparison"
	}
	return "general"
}

// parseQueries 解析查询列表
func (e *AdvancedQueryExpander) parseQueries(response string) []string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	queries := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		if line != "" && !strings.HasPrefix(line, "#") {
			queries = append(queries, line)
		}
	}

	return queries
}

// extractKeyPhrases 提取关键短语
func (e *AdvancedQueryExpander) extractKeyPhrases(text string) []string {
	// 简化实现：按句子分割
	sentences := strings.Split(text, "。")
	phrases := make([]string, 0, len(sentences))

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 && len(sentence) < 100 {
			phrases = append(phrases, sentence)
		}
	}

	if len(phrases) > 3 {
		phrases = phrases[:3]
	}

	return phrases
}

// containsQuery 检查是否已包含查询
func (e *AdvancedQueryExpander) containsQuery(optimizations []QueryOptimization, query string) bool {
	for _, opt := range optimizations {
		if opt.Query == query {
			return true
		}
	}
	return false
}

// Name 返回扩展器名称
func (e *AdvancedQueryExpander) Name() string {
	return fmt.Sprintf("advanced_expander_%s", e.strategy)
}
