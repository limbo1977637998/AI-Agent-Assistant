package routing

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ConditionalRouter 条件路由器
//
// 功能: 基于预定义条件规则选择检索策略
//
// 支持的条件类型:
//   1. 关键词匹配
//   2. 正则表达式匹配
//   3. 查询长度
//   4. 查询类型
//   5. 时间条件
//   6. 自定义条件
//
// 优点:
//   - 灵活的路由规则
//   - 可解释的决策
//   - 易于维护和扩展
//   - 高性能
type ConditionalRouter struct {
	rules      []*RoutingRule
	defaultStrategy string
	llm        LLMProvider
}

// LLMProvider LLM 提供者接口
type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// RoutingRule 路由规则
type RoutingRule struct {
	// Name 规则名称
	Name string

	// Condition 条件表达式
	Condition string

	// ConditionType 条件类型 (keyword, regex, length, type, custom, llm)
	ConditionType string

	// Strategy 目标策略
	Strategy string

	// Priority 优先级 (数字越大优先级越高)
	Priority int

	// Metadata 附加元数据
	Metadata map[string]interface{}

	// ValidateFunc 自定义验证函数
	ValidateFunc func(ctx context.Context, query string) (bool, error)
}

// NewConditionalRouter 创建条件路由器
func NewConditionalRouter(llm LLMProvider, defaultStrategy string) (*ConditionalRouter, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	return &ConditionalRouter{
		rules:          make([]*RoutingRule, 0),
		defaultStrategy: defaultStrategy,
		llm:            llm,
	}, nil
}

// Route 根据条件路由到策略
func (cr *ConditionalRouter) Route(ctx context.Context, query string) (string, error) {
	// 按优先级排序规则
	sortedRules := cr.sortRulesByPriority()

	// 匹配规则
	for _, rule := range sortedRules {
		matched, err := cr.matchRule(ctx, rule, query)
		if err != nil {
			// 规则匹配失败，跳过
			continue
		}

		if matched {
			return rule.Strategy, nil
		}
	}

	// 没有匹配的规则，返回默认策略
	return cr.defaultStrategy, nil
}

// matchRule 匹配规则
func (cr *ConditionalRouter) matchRule(ctx context.Context, rule *RoutingRule, query string) (bool, error) {
	switch rule.ConditionType {
	case "keyword":
		return cr.matchKeyword(rule, query), nil
	case "regex":
		return cr.matchRegex(rule, query)
	case "length":
		return cr.matchLength(rule, query), nil
	case "type":
		return cr.matchQueryType(rule, query), nil
	case "custom":
		if rule.ValidateFunc != nil {
			return rule.ValidateFunc(ctx, query)
		}
		return false, nil
	case "llm":
		return cr.matchWithLLM(ctx, rule, query)
	default:
		return false, nil
	}
}

// matchKeyword 关键词匹配
func (cr *ConditionalRouter) matchKeyword(rule *RoutingRule, query string) bool {
	queryLower := strings.ToLower(query)
	conditionLower := strings.ToLower(rule.Condition)

	// 支持多个关键词（用逗号分隔）
	keywords := strings.Split(conditionLower, ",")

	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}

// matchRegex 正则表达式匹配
func (cr *ConditionalRouter) matchRegex(rule *RoutingRule, query string) (bool, error) {
	matched, err := regexp.MatchString(rule.Condition, query)
	if err != nil {
		return false, fmt.Errorf("regex match failed: %w", err)
	}
	return matched, nil
}

// matchLength 查询长度匹配
// 条件格式: ">10", "<5", "5-20", "==10"
func (cr *ConditionalRouter) matchLength(rule *RoutingRule, query string) bool {
	queryLen := len([]rune(query)) // 使用字符数

	if strings.HasPrefix(rule.Condition, ">") {
		threshold := 0
		fmt.Sscanf(rule.Condition, ">%d", &threshold)
		return queryLen > threshold
	}

	if strings.HasPrefix(rule.Condition, "<") {
		threshold := 0
		fmt.Sscanf(rule.Condition, "<%d", &threshold)
		return queryLen < threshold
	}

	if strings.Contains(rule.Condition, "-") {
		min, max := 0, 0
		fmt.Sscanf(rule.Condition, "%d-%d", &min, &max)
		return queryLen >= min && queryLen <= max
	}

	if strings.HasPrefix(rule.Condition, "==") {
		threshold := 0
		fmt.Sscanf(rule.Condition, "==%d", &threshold)
		return queryLen == threshold
	}

	return false
}

// matchQueryType 查询类型匹配
func (cr *ConditionalRouter) matchQueryType(rule *RoutingRule, query string) bool {
	queryType := cr.detectQueryType(query)
	return queryType == rule.Condition
}

// matchWithLLM 使用 LLM 匹配条件
func (cr *ConditionalRouter) matchWithLLM(ctx context.Context, rule *RoutingRule, query string) (bool, error) {
	prompt := fmt.Sprintf(`请判断以下查询是否满足条件。

查询: %s
条件: %s

要求:
1. 分析查询是否满足条件
2. 只回答 "true" 或 "false"
3. 不要其他内容

判断:`, query, rule.Condition)

	response, err := cr.llm.Generate(ctx, prompt)
	if err != nil {
		return false, fmt.Errorf("LLM generation failed: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "true" || response == "yes" || response == "是", nil
}

// sortRulesByPriority 按优先级排序规则
func (cr *ConditionalRouter) sortRulesByPriority() []*RoutingRule {
	sorted := make([]*RoutingRule, len(cr.rules))
	copy(sorted, cr.rules)

	// 简单冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Priority > sorted[i].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// detectQueryType 检测查询类型
func (cr *ConditionalRouter) detectQueryType(query string) string {
	queryLower := strings.ToLower(query)

	// 定义性查询
	if matched, _ := regexp.MatchString("(什么|什么是|定义|解释|说明)", queryLower); matched {
		return "definition"
	}

	// 程序性查询
	if matched, _ := regexp.MatchString("(如何|怎么|怎样|步骤|流程)", queryLower); matched {
		return "procedure"
	}

	// 推理性查询
	if matched, _ := regexp.MatchString("(为什么|原因|为何)", queryLower); matched {
		return "reasoning"
	}

	// 对比性查询
	if matched, _ := regexp.MatchString("(区别|差异|对比|比较)", queryLower); matched {
		return "comparison"
	}

	// 列举性查询
	if matched, _ := regexp.MatchString("(有哪些|列举|列表|包括)", queryLower); matched {
		return "listing"
	}

	// 全局性查询
	if matched, _ := regexp.MatchString("(整体|总体|所有|总结|架构|概述)", queryLower); matched {
		return "global"
	}

	// 具体性查询
	if matched, _ := regexp.MatchString("(具体|详细|某个)", queryLower); matched {
		return "specific"
	}

	return "general"
}

// AddRule 添加路由规则
func (cr *ConditionalRouter) AddRule(rule *RoutingRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}

	if rule.Priority <= 0 {
		rule.Priority = 5 // 默认优先级
	}

	cr.rules = append(cr.rules, rule)
	return nil
}

// RemoveRule 移除路由规则
func (cr *ConditionalRouter) RemoveRule(ruleName string) bool {
	for i, rule := range cr.rules {
		if rule.Name == ruleName {
			// 删除规则
			cr.rules = append(cr.rules[:i], cr.rules[i+1:]...)
			return true
		}
	}
	return false
}

// GetRules 获取所有规则
func (cr *ConditionalRouter) GetRules() []*RoutingRule {
	return cr.rules
}

// ClearRules 清除所有规则
func (cr *ConditionalRouter) ClearRules() {
	cr.rules = make([]*RoutingRule, 0)
}

// SetDefaultStrategy 设置默认策略
func (cr *ConditionalRouter) SetDefaultStrategy(strategy string) {
	cr.defaultStrategy = strategy
}

// GetDefaultStrategy 获取默认策略
func (cr *ConditionalRouter) GetDefaultStrategy() string {
	return cr.defaultStrategy
}

// ExplainRouting 解释路由决策
func (cr *ConditionalRouter) ExplainRouting(ctx context.Context, query string) (*RoutingExplanation, error) {
	sortedRules := cr.sortRulesByPriority()

	explanation := &RoutingExplanation{
		Query:   query,
		Matched: make([]*MatchedRule, 0),
		Rejected: make([]*RejectedRule, 0),
	}

	for _, rule := range sortedRules {
		matched, err := cr.matchRule(ctx, rule, query)
		if err != nil {
			explanation.Rejected = append(explanation.Rejected, &RejectedRule{
				Rule:    rule,
				Reason:  err.Error(),
			})
			continue
		}

		if matched {
			explanation.Matched = append(explanation.Matched, &MatchedRule{
				Rule:     rule,
				Selected: rule.Strategy,
			})

			// 返回第一个匹配的规则
			explanation.SelectedStrategy = rule.Strategy
			return explanation, nil
		} else {
			explanation.Rejected = append(explanation.Rejected, &RejectedRule{
				Rule:    rule,
				Reason:  "Condition not matched",
			})
		}
	}

	// 没有匹配的规则
	explanation.SelectedStrategy = cr.defaultStrategy
	explanation.UsedDefault = true

	return explanation, nil
}

// RoutingExplanation 路由解释
type RoutingExplanation struct {
	Query            string
	SelectedStrategy string
	Matched          []*MatchedRule
	Rejected         []*RejectedRule
	UsedDefault      bool
}

// MatchedRule 匹配的规则
type MatchedRule struct {
	Rule     *RoutingRule
	Selected string
}

// RejectedRule 拒绝的规则
type RejectedRule struct {
	Rule   *RoutingRule
	Reason string
}

// Name 返回路由器名称
func (cr *ConditionalRouter) Name() string {
	return "conditional_router"
}

// CreateDefaultRules 创建默认规则集
func (cr *ConditionalRouter) CreateDefaultRules() error {
	rules := []*RoutingRule{
		{
			Name:          "全局问题路由",
			Condition:     "(总体|整体|所有|总结|架构|概述)",
			ConditionType: "regex",
			Strategy:      "graph_rag",
			Priority:      10,
		},
		{
			Name:          "定义查询路由",
			Condition:     "definition",
			ConditionType: "type",
			Strategy:      "vector",
			Priority:      8,
		},
		{
			Name:          "复杂推理路由",
			Condition:     "(为什么|原因|分析|比较|关系)",
			ConditionType: "regex",
			Strategy:      "graph_rag",
			Priority:      9,
		},
		{
			Name:          "程序查询路由",
			Condition:     "procedure",
			ConditionType: "type",
			Strategy:      "hybrid",
			Priority:      7,
		},
		{
			Name:          "短查询路由",
			Condition:     "<10",
			ConditionType: "length",
			Strategy:      "vector",
			Priority:      6,
		},
		{
			Name:          "长查询路由",
			Condition:     ">30",
			ConditionType: "length",
			Strategy:      "hyde",
			Priority:      5,
		},
	}

	for _, rule := range rules {
		if err := cr.AddRule(rule); err != nil {
			return err
		}
	}

	return nil
}
