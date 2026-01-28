package query

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// AdvancedQueryRewriter 高级查询重写器
//
// 支持多种重写策略:
//   1. Rule-based Rewriting (基于规则)
//   2. LLM-based Rewriting (基于 LLM)
//   3. Multi-round Rewriting (多轮重写)
//   4. Intent-based Rewriting (基于意图)
//
// 优点:
//   - 多策略融合
//   - 提高查询质量
//   - 处理模糊查询
//   - 理解用户意图
type AdvancedQueryRewriter struct {
	llm           LLMProvider
	strategy      string // rule, llm, hybrid, multi_round
	maxRounds     int
	intentRules   map[string]string
	rewriteRules  []*RewriteRule
}

// RewriteRule 重写规则
type RewriteRule struct {
	Pattern   string // 匹配模式
	Template  string // 重写模板
	Priority  int    // 优先级
	Intent    string // 意图类型
}

// NewAdvancedQueryRewriter 创建高级查询重写器
func NewAdvancedQueryRewriter(llm LLMProvider, strategy string) (*AdvancedQueryRewriter, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	rewriter := &AdvancedQueryRewriter{
		llm:          llm,
		strategy:     strategy,
		maxRounds:    3,
		intentRules:  make(map[string]string),
		rewriteRules: make([]*RewriteRule, 0),
	}

	// 初始化默认规则
	rewriter.initDefaultRules()

	return rewriter, nil
}

// initDefaultRules 初始化默认重写规则
func (r *AdvancedQueryRewriter) initDefaultRules() {
	// 意图识别规则
	r.intentRules = map[string]string{
		"definition": "(什么|什么是|定义|解释|说明)",
		"comparison": "(区别|差异|对比|比较|区别在哪)",
		"procedure":  "(如何|怎么|怎样|步骤|流程|方法)",
		"reasoning":  "(为什么|原因|为何|理由)",
		"listing":    "(有哪些|列举|列表|包括)",
		"global":     "(整体|总体|所有|总结|架构|概述)",
	}

	// 重写规则
	r.rewriteRules = []*RewriteRule{
		{
			Pattern:  `(.+)是什么`,
			Template: "请详细解释 $1 的定义、特点和作用",
			Priority: 10,
			Intent:   "definition",
		},
		{
			Pattern:  `(.+)和(.+)的区别`,
			Template: "请对比分析 $1 和 $2 的异同点、优缺点和适用场景",
			Priority: 10,
			Intent:   "comparison",
		},
		{
			Pattern:  `如何(.+)`,
			Template: "请详细说明如何 $1，包括步骤、方法和注意事项",
			Priority: 9,
			Intent:   "procedure",
		},
		{
			Pattern:  `为什么(.+)`,
			Template: "请分析 $1 的原因、背景和影响因素",
			Priority: 9,
			Intent:   "reasoning",
		},
		{
			Pattern:  `(.+)有哪些`,
			Template: "请列举 $1 的主要类型、特点和示例",
			Priority: 8,
			Intent:   "listing",
		},
	}
}

// Rewrite 重写查询
func (r *AdvancedQueryRewriter) Rewrite(ctx context.Context, query string) ([]QueryOptimization, error) {
	switch r.strategy {
	case "rule":
		return r.ruleBasedRewrite(query), nil
	case "llm":
		return r.llmBasedRewrite(ctx, query)
	case "hybrid":
		return r.hybridRewrite(ctx, query)
	case "multi_round":
		return r.multiRoundRewrite(ctx, query)
	default:
		return r.hybridRewrite(ctx, query)
	}
}

// ruleBasedRewrite 基于规则的重写
func (r *AdvancedQueryRewriter) ruleBasedRewrite(query string) []QueryOptimization {
	optimizations := make([]QueryOptimization, 0)

	// 按优先级匹配规则
	matchedRule := r.matchBestRule(query)

	if matchedRule != nil {
		rewritten := r.applyRule(query, matchedRule)

		optimizations = append(optimizations, QueryOptimization{
			Query: rewritten,
			Type:  "rule_rewrite",
			Score: 0.9,
			Metadata: map[string]interface{}{
				"original_query": query,
				"intent":         matchedRule.Intent,
				"rule_priority":  matchedRule.Priority,
			},
		})
	}

	// 始终包含原始查询
	if len(optimizations) == 0 || matchedRule == nil {
		optimizations = append(optimizations, QueryOptimization{
			Query: query,
			Type:  "original",
			Score: 1.0,
		})
	}

	return optimizations
}

// llmBasedRewrite 基于 LLM 的重写
func (r *AdvancedQueryRewriter) llmBasedRewrite(ctx context.Context, query string) ([]QueryOptimization, error) {
	intent := r.detectIntent(query)

	prompt := r.buildRewritePrompt(query, intent)

	response, err := r.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 解析 LLM 响应
	rewrittenQueries := r.parseLLMResponse(response)

	optimizations := make([]QueryOptimization, 0, len(rewrittenQueries))
	for i, q := range rewrittenQueries {
		score := 1.0 - float64(i)*0.1
		optimizations = append(optimizations, QueryOptimization{
			Query: q,
			Type:  "llm_rewrite",
			Score: score,
			Metadata: map[string]interface{}{
				"original_query": query,
				"intent":         intent,
				"rewrite_index":  i,
			},
		})
	}

	return optimizations, nil
}

// hybridRewrite 混合重写（规则 + LLM）
func (r *AdvancedQueryRewriter) hybridRewrite(ctx context.Context, query string) ([]QueryOptimization, error) {
	// 先尝试规则重写
	ruleResults := r.ruleBasedRewrite(query)

	// 如果规则匹配成功且置信度高，直接返回
	if len(ruleResults) > 0 && ruleResults[0].Score > 0.85 {
		return ruleResults, nil
	}

	// 规则匹配失败或置信度低，使用 LLM
	llmResults, err := r.llmBasedRewrite(ctx, query)
	if err != nil {
		// LLM 失败，返回规则结果
		return ruleResults, nil
	}

	// 合并结果
	return append(ruleResults, llmResults...), nil
}

// multiRoundRewrite 多轮重写
func (r *AdvancedQueryRewriter) multiRoundRewrite(ctx context.Context, query string) ([]QueryOptimization, error) {
	allOptimizations := make([]QueryOptimization, 0)

	currentQuery := query
	for round := 0; round < r.maxRounds; round++ {
		// 每轮使用 LLM 重写
		prompt := fmt.Sprintf(`请改进以下查询，使其更清晰、更具体。

当前查询: %s

要求:
1. 保持原意
2. 添加必要的上下文
3. 明确模糊的表达
4. 只返回改进后的查询，不要解释

改进后的查询:`, currentQuery)

		response, err := r.llm.Generate(ctx, prompt)
		if err != nil {
			break
		}

		rewritten := strings.TrimSpace(response)

		// 检查是否有改进
		if rewritten == currentQuery || rewritten == "" {
			break
		}

		allOptimizations = append(allOptimizations, QueryOptimization{
			Query: rewritten,
			Type:  "multi_round_rewrite",
			Score: 1.0 - float64(round)*0.15,
			Metadata: map[string]interface{}{
				"original_query": query,
				"round":          round + 1,
				"previous_query": currentQuery,
			},
		})

		currentQuery = rewritten
	}

	// 添加原始查询
	if len(allOptimizations) == 0 {
		allOptimizations = append(allOptimizations, QueryOptimization{
			Query: query,
			Type:  "original",
			Score: 1.0,
		})
	}

	return allOptimizations, nil
}

// matchBestRule 匹配最佳规则
func (r *AdvancedQueryRewriter) matchBestRule(query string) *RewriteRule {
	var bestRule *RewriteRule
	bestPriority := -1

	for _, rule := range r.rewriteRules {
		matched, _ := regexp.MatchString(rule.Pattern, query)
		if matched && rule.Priority > bestPriority {
			bestRule = rule
			bestPriority = rule.Priority
		}
	}

	return bestRule
}

// applyRule 应用重写规则
func (r *AdvancedQueryRewriter) applyRule(query string, rule *RewriteRule) string {
	// 简化实现：使用正则替换
	re := regexp.MustCompile(rule.Pattern)
	result := re.ReplaceAllString(query, rule.Template)

	return result
}

// detectIntent 检测查询意图
func (r *AdvancedQueryRewriter) detectIntent(query string) string {
	for intent, pattern := range r.intentRules {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return intent
		}
	}
	return "general"
}

// buildRewritePrompt 构建重写提示
func (r *AdvancedQueryRewriter) buildRewritePrompt(query, intent string) string {
	prompt := fmt.Sprintf(`请改进以下查询，使其更加清晰和具体。

原始查询: %s
查询意图: %s

要求:
1. 保持原意不变
2. 添加必要的上下文信息
3. 明确模糊的表达
4. 使查询更易于检索
5. 提供 3 个改进版本，每行一个

改进后的查询（每行一个）:`, query, intent)

	return prompt
}

// parseLLMResponse 解析 LLM 响应
func (r *AdvancedQueryRewriter) parseLLMResponse(response string) []string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	queries := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和标记
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		queries = append(queries, line)
	}

	if len(queries) == 0 {
		queries = []string{response}
	}

	return queries
}

// Name 返回重写器名称
func (r *AdvancedQueryRewriter) Name() string {
	return fmt.Sprintf("advanced_rewriter_%s", r.strategy)
}
