package query

import (
	"context"
	"fmt"
	"strings"
)

// QueryExpander 查询扩展器
//
// 策略说明:
//   使用 LLM 或同义词库扩展查询，增加同义词和相关词
//   提高召回率，避免遗漏相关信息
//
// 优点:
//   - 提高召回率
//   - 覆盖同义词表达
//   - 处理术语变体
//
// 适用场景:
//   - 专业术语检索
//   - 同义词多的查询
//   - 需要高召回率的场景
type QueryExpander struct {
	llm    LLMProvider
	config QueryOptimizerConfig
	name   string
}

// NewQueryExpander 创建查询扩展器
func NewQueryExpander(llm LLMProvider, config QueryOptimizerConfig) (*QueryExpander, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	// 设置默认值
	if config.MaxQueries <= 0 {
		config.MaxQueries = 5
	}
	if config.MinConfidence <= 0 {
		config.MinConfidence = 0.6
	}

	return &QueryExpander{
		llm:    llm,
		config: config,
		name:   "query_expander",
	}, nil
}

// Optimize 实现查询优化接口
func (qe *QueryExpander) Optimize(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 构建扩展提示
	prompt := qe.buildExpandPrompt(query)

	// 调用 LLM 扩展
	response, err := qe.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 解析扩展结果
	queries := qe.parseExpansion(response)

	// 添加原始查询
	queries = append([]string{query}, queries...)

	// 限制数量
	if len(queries) > qe.config.MaxQueries {
		queries = queries[:qe.config.MaxQueries]
	}

	// 构建优化结果
	optimizations := make([]QueryOptimization, len(queries))
	for i, q := range queries {
		score := 1.0
		if i > 0 {
			score = 0.8 - float64(i)*0.1 // 扩展查询的递减置信度
		}

		optimizations[i] = QueryOptimization{
			Query: q,
			Type:  "expand",
			Score: score,
			Metadata: map[string]interface{}{
				"expand_index": i,
				"original_query": query,
			},
		}
	}

	return optimizations, nil
}

// buildExpandPrompt 构建扩展提示
func (qe *QueryExpander) buildExpandPrompt(query string) string {
	if qe.config.Language == "zh" {
		return qe.buildChinesePrompt(query)
	}
	return qe.buildEnglishPrompt(query)
}

// buildChinesePrompt 构建中文提示
func (qe *QueryExpander) buildChinesePrompt(query string) string {
	return fmt.Sprintf(`你是一个查询扩展专家。请为查询生成 3-5 个等价或扩展的查询。

要求:
1. 使用同义词、相关词、上下位词
2. 保持原意不变
3. 覆盖不同的表达方式
4. 每行一个查询
5. 不要解释，只输出扩展后的查询

原始查询: %s

扩展后的查询 (每行一个):`, query)
}

// buildEnglishPrompt 构建英文提示
func (qe *QueryExpander) buildEnglishPrompt(query string) string {
	return fmt.Sprintf(`You are a query expansion expert. Please generate 3-5 equivalent or expanded queries for the given query.

Requirements:
1. Use synonyms, related terms, hypernyms, hyponyms
2. Keep the original meaning
3. Cover different expressions
4. One query per line
5. No explanation, only output the expanded queries

Original query: %s

Expanded queries (one per line):`, query)
}

// parseExpansion 解析扩展结果
func (qe *QueryExpander) parseExpansion(response string) []string {
	response = strings.TrimSpace(response)

	// 按行分割
	lines := strings.Split(response, "\n")
	queries := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行、注释、序号
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		// 移除序号前缀
		if len(line) > 2 && line[1] == '.' {
			if line[0] >= '1' && line[0] <= '9' {
				line = strings.TrimPrefix(line, string(line[0]))
				line = strings.TrimPrefix(line, ".")
				line = strings.TrimPrefix(line, " ")
			}
		}

		line = strings.TrimSpace(line)
		if line != "" {
			queries = append(queries, line)
		}
	}

	return queries
}

// Name 返回优化器名称
func (qe *QueryExpander) Name() string {
	return qe.name
}

// Validate 验证配置
func (qe *QueryExpander) Validate() error {
	if qe.llm == nil {
		return fmt.Errorf("LLM provider is required")
	}
	if qe.config.MaxQueries <= 0 {
		return fmt.Errorf("max_queries must be positive")
	}
	return nil
}

// ExpandSimple 简单的扩展方法（基于规则）
// 用于快速实现或回退方案
func (qe *QueryExpander) ExpandSimple(query string) []string {
	queries := []string{query}

	// 简单的同义词映射（中文）
	synonyms := map[string][]string{
		"人工智能": {"AI", "机器智能", "智能系统"},
		"机器学习": {"ML", "Machine Learning", "算法学习"},
		"深度学习": {"DL", "Deep Learning", "神经网络"},
		"自然语言处理": {"NLP", "Natural Language Processing", "文本处理"},
		"计算机视觉": {"CV", "Computer Vision", "图像识别"},
		"数据": {"信息", "资料", "数据集"},
		"算法": {"方法", "模型", "策略"},
		"系统": {"平台", "框架", "架构"},
		"技术": {"技术栈", "方案", "解决方案"},
	}

	// 查找并替换
	expanded := false
	for term, syns := range synonyms {
		if strings.Contains(query, term) {
			for _, syn := range syns {
				expandedQuery := strings.ReplaceAll(query, term, syn)
				if expandedQuery != query {
					queries = append(queries, expandedQuery)
					expanded = true
				}
			}
			break
		}
	}

	// 如果没有找到同义词，尝试拆分关键词
	if !expanded && len(query) > 5 {
		words := strings.Fields(query)
		if len(words) >= 2 {
			// 生成双词组合
			for i := 0; i < len(words)-1; i++ {
				bigram := words[i] + " " + words[i+1]
				queries = append(queries, bigram)
			}
		}
	}

	return queries
}
