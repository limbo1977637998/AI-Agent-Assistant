package query

import (
	"context"
	"fmt"
	"strings"
)

// QueryRewriter 查询重写器
//
// 策略说明:
//   使用 LLM 改写用户查询，使其更清晰、更具体、更适合检索
//
// 优点:
//   - 提高查询质量
//   - 消除歧义
//   - 补充缺失信息
//
// 适用场景:
//   - 模糊查询
//   - 不完整查询
//   - 需要澄清的查询
type QueryRewriter struct {
	llm    LLMProvider // LLM 接口
	config QueryOptimizerConfig
	name   string
}

// LLMProvider LLM 提供者接口
// 定义 LLM 需要实现的方法
type LLMProvider interface {
	// Generate 生成文本
	Generate(ctx context.Context, prompt string) (string, error)
}

// NewQueryRewriter 创建查询重写器
func NewQueryRewriter(llm LLMProvider, config QueryOptimizerConfig) (*QueryRewriter, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	// 设置默认值
	if config.MaxQueries <= 0 {
		config.MaxQueries = 3
	}
	if config.MinConfidence <= 0 {
		config.MinConfidence = 0.6
	}
	if config.Language == "" {
		config.Language = "zh"
	}

	return &QueryRewriter{
		llm:    llm,
		config: config,
		name:   "query_rewriter",
	}, nil
}

// Optimize 实现查询优化接口
func (qr *QueryRewriter) Optimize(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 构建重写提示
	prompt := qr.buildRewritePrompt(query)

	// 调用 LLM 重写
	rewritten, err := qr.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 解析重写结果
	rewritten = strings.TrimSpace(rewritten)
	if rewritten == "" {
		// 如果重写失败，返回原始查询
		return []QueryOptimization{{
			Query: query,
			Type:  "rewrite",
			Score: 1.0,
		}}, nil
	}

	// 返回优化结果
	optimizations := []QueryOptimization{
		{
			Query: rewritten,
			Type:  "rewrite",
			Score: 0.9, // 重写查询的默认置信度
			Metadata: map[string]interface{}{
				"original_query": query,
				"rewrite_type":   "llm_based",
			},
		},
	}

	// 也可以同时返回原始查询（用于对比）
	optimizations = append(optimizations, QueryOptimization{
		Query: query,
		Type:  "original",
		Score: 1.0, // 原始查询的得分
	})

	return optimizations, nil
}

// buildRewritePrompt 构建重写提示
func (qr *QueryRewriter) buildRewritePrompt(query string) string {
	if qr.config.Language == "zh" {
		return qr.buildChinesePrompt(query)
	}
	return qr.buildEnglishPrompt(query)
}

// buildChinesePrompt 构建中文提示
func (qr *QueryRewriter) buildChinesePrompt(query string) string {
	return fmt.Sprintf(`你是一个查询优化专家。请改写用户查询，使其更清晰、更具体、更适合信息检索。

要求:
1. 保持原意不变
2. 补充缺失的上下文
3. 消除歧义
4. 使用更准确的关键词
5. 只输出改写后的查询，不要解释

原始查询: %s

改写后的查询:`, query)
}

// buildEnglishPrompt 构建英文提示
func (qr *QueryRewriter) buildEnglishPrompt(query string) string {
	return fmt.Sprintf(`You are a query optimization expert. Please rewrite the user query to make it clearer, more specific, and more suitable for information retrieval.

Requirements:
1. Keep the original meaning
2. Add missing context
3. Eliminate ambiguity
4. Use more accurate keywords
5. Output only the rewritten query, no explanation

Original query: %s

Rewritten query:`, query)
}

// Name 返回优化器名称
func (qr *QueryRewriter) Name() string {
	return qr.name
}

// Validate 验证配置
func (qr *QueryRewriter) Validate() error {
	if qr.llm == nil {
		return fmt.Errorf("LLM provider is required")
	}
	if qr.config.MaxQueries <= 0 {
		return fmt.Errorf("max_queries must be positive")
	}
	return nil
}

// RewriteSimple 简单的重写方法（不使用 LLM）
// 用于快速实现或回退方案
func (qr *QueryRewriter) RewriteSimple(query string) string {
	// 简单的查询扩展
	rewritten := query

	// 移除常见的无意义前缀
	prefixes := []string{"帮我", "请", "能否", "可以", "帮我查一下", "我想知道"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(rewritten, prefix) {
			rewritten = strings.TrimPrefix(rewritten, prefix)
			break
		}
	}

	// 添加关键词（根据查询内容）
	if strings.Contains(rewritten, "怎么") || strings.Contains(rewritten, "如何") {
		rewritten = strings.ReplaceAll(rewritten, "怎么", "方法")
		rewritten = strings.ReplaceAll(rewritten, "如何", "方法")
	}

	return rewritten
}
