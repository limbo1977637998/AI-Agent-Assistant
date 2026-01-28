package query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// QueryDecomposer 查询分解器
//
// 策略说明:
//   将复杂查询分解为多个简单的子查询
//   每个子查询可以独立检索，最后合并结果
//
// 优点:
//   - 处理复杂多面问题
//   - 提高召回率
//   - 支持多步骤推理
//
// 适用场景:
//   - 复杂问题
//   - 多个信息需求
//   - 需要多角度检索的问题
type QueryDecomposer struct {
	llm    LLMProvider
	config QueryOptimizerConfig
	name   string
}

// NewQueryDecomposer 创建查询分解器
func NewQueryDecomposer(llm LLMProvider, config QueryOptimizerConfig) (*QueryDecomposer, error) {
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

	return &QueryDecomposer{
		llm:    llm,
		config: config,
		name:   "query_decomposer",
	}, nil
}

// Optimize 实现查询优化接口
func (qd *QueryDecomposer) Optimize(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 构建分解提示
	prompt := qd.buildDecomposePrompt(query)

	// 调用 LLM 分解
	response, err := qd.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 解析分解结果
	queries, err := qd.parseDecomposition(response)
	if err != nil {
		// 如果解析失败，返回原始查询
		return []QueryOptimization{{
			Query: query,
			Type:  "original",
			Score: 1.0,
		}}, nil
	}

	// 限制子查询数量
	if len(queries) > qd.config.MaxQueries {
		queries = queries[:qd.config.MaxQueries]
	}

	// 构建优化结果
	optimizations := make([]QueryOptimization, len(queries))
	for i, q := range queries {
		optimizations[i] = QueryOptimization{
			Query: q,
			Type:  "decompose",
			Score: 1.0 - float64(i)*0.1, // 递减的置信度
			Metadata: map[string]interface{}{
				"sub_query_index": i,
				"original_query":   query,
				"total_sub_queries": len(queries),
			},
		}
	}

	return optimizations, nil
}

// buildDecomposePrompt 构建分解提示
func (qd *QueryDecomposer) buildDecomposePrompt(query string) string {
	if qd.config.Language == "zh" {
		return qd.buildChinesePrompt(query)
	}
	return qd.buildEnglishPrompt(query)
}

// buildChinesePrompt 构建中文提示
func (qd *QueryDecomposer) buildChinesePrompt(query string) string {
	return fmt.Sprintf(`你是一个查询分解专家。请将复杂的查询分解为多个简单的子查询。

要求:
1. 每个子查询应该简单、明确
2. 子查询之间应该相互独立
3. 子查询数量 2-5 个
4. 确保所有子查询合起来能完整回答原始查询
5. 以 JSON 数组格式输出: ["子查询1", "子查询2", ...]

原始查询: %s

分解后的子查询 (JSON 数组):`, query)
}

// buildEnglishPrompt 构建英文提示
func (qd *QueryDecomposer) buildEnglishPrompt(query string) string {
	return fmt.Sprintf(`You are a query decomposition expert. Please break down the complex query into multiple simple sub-queries.

Requirements:
1. Each sub-query should be simple and clear
2. Sub-queries should be independent
3. Number of sub-queries: 2-5
4. All sub-queries together should fully answer the original query
5. Output as JSON array: ["subquery1", "subquery2", ...]

Original query: %s

Decomposed sub-queries (JSON array):`, query)
}

// parseDecomposition 解析分解结果
func (qd *QueryDecomposer) parseDecomposition(response string) ([]string, error) {
	response = strings.TrimSpace(response)

	// 尝试解析 JSON
	var queries []string
	if err := json.Unmarshal([]byte(response), &queries); err == nil {
		return queries, nil
	}

	// 如果 JSON 解析失败，尝试按行分割
	lines := strings.Split(response, "\n")
	queries = make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		// 移除序号前缀 (1. 2. 3. 或 - - -)
		if len(line) > 2 && (line[1] == '.' || line[1] == ' ') {
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

	if len(queries) == 0 {
		return nil, fmt.Errorf("failed to parse decomposition")
	}

	return queries, nil
}

// Name 返回优化器名称
func (qd *QueryDecomposer) Name() string {
	return qd.name
}

// Validate 验证配置
func (qd *QueryDecomposer) Validate() error {
	if qd.llm == nil {
		return fmt.Errorf("LLM provider is required")
	}
	if qd.config.MaxQueries <= 0 {
		return fmt.Errorf("max_queries must be positive")
	}
	return nil
}

// DecomposeSimple 简单的分解方法（基于规则）
// 用于快速实现或回退方案
func (qd *QueryDecomposer) DecomposeSimple(query string) []string {
	queries := []string{query}

	// 检测是否包含多个问题
	questionMarks := strings.Count(query, "？") + strings.Count(query, "?")
	if questionMarks > 1 {
		// 按问号分割
		parts := strings.FieldsFunc(query, func(r rune) bool {
			return r == '？' || r == '?'
		})

		queries = make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				queries = append(queries, part)
			}
		}
	}

	// 检测关键词：和、以及、或
	connectors := []string{" 和 ", " 以及 ", " 或 ", " and ", " or "}
	for _, connector := range connectors {
		if strings.Contains(query, connector) {
			parts := strings.Split(query, connector)
			if len(parts) > 1 {
				queries = make([]string, 0, len(parts))
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part != "" {
						queries = append(queries, part)
					}
				}
				break
			}
		}
	}

	return queries
}
