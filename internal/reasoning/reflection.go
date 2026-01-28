package reasoning

import (
	"context"
	"fmt"
	"strings"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// Reflection 自我反思
// 让模型检查之前的回答，识别错误并改进
type Reflection struct {
	reflectionModel llm.Model
	numReflections  int // 反思次数
}

// NewReflection 创建反思器
func NewReflection(model llm.Model, numReflections int) *Reflection {
	if numReflections <= 0 {
		numReflections = 1 // 默认反思1次
	}

	return &Reflection{
		reflectionModel: model,
		numReflections:  numReflections,
	}
}

// Reflect 对之前的回答进行反思并改进
func (r *Reflection) Reflect(ctx context.Context, task string, previousAttempts []string) (reflection string, improvedAnswer string, err error) {
	if len(previousAttempts) == 0 {
		return "", "", fmt.Errorf("no previous attempts to reflect on")
	}

	// 构造反思提示
	prompt := r.buildReflectionPrompt(task, previousAttempts)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	response, err := r.reflectionModel.Chat(ctx, messages)
	if err != nil {
		return "", "", fmt.Errorf("failed to reflect: %w", err)
	}

	// 解析响应
	return r.parseReflectionResponse(response)
}

// buildReflectionPrompt 构建反思提示
func (r *Reflection) buildReflectionPrompt(task string, previousAttempts []string) string {
	var sb strings.Builder

	sb.WriteString("请检查之前的回答，识别可能的错误或改进点。\n\n")
	sb.WriteString(fmt.Sprintf("任务：%s\n\n", task))
	sb.WriteString("之前的尝试：\n")

	for i, attempt := range previousAttempts {
		sb.WriteString(fmt.Sprintf("尝试%d：%s\n\n", i+1, attempt))
	}

	sb.WriteString(`请按以下格式回答：

【反思】
（分析之前回答的问题和不足）

【改进后的答案】
（基于反思，给出更好的答案）`)

	return sb.String()
}

// parseReflectionResponse 解析反思响应
func (r *Reflection) parseReflectionResponse(response string) (reflection string, improvedAnswer string, err error) {
	// 尝试分离【反思】和【改进后的答案】
	parts := strings.Split(response, "【改进后的答案】")

	if len(parts) >= 2 {
		// 提取反思
		reflectionPart := parts[0]
		reflectionPart = strings.TrimPrefix(reflectionPart, "【反思】")
		reflectionPart = strings.TrimSpace(reflectionPart)

		// 提取改进后的答案
		answerPart := parts[1]
		answerPart = strings.TrimSpace(answerPart)

		return reflectionPart, answerPart, nil
	}

	// 如果无法分离，返回完整响应
	return "", response, nil
}

// ReflectAndIterate 反思并迭代改进
func (r *Reflection) ReflectAndIterate(ctx context.Context, task string, initialAnswer string) (finalAnswer string, iterations []Iteration, err error) {
	currentAnswer := initialAnswer
	attempts := []string{currentAnswer}
	iterations = make([]Iteration, 0)

	for i := 0; i < r.numReflections; i++ {
		// 反思
		reflection, improvedAnswer, err := r.Reflect(ctx, task, attempts)
		if err != nil {
			// 反思失败，使用当前答案
			return currentAnswer, iterations, nil
		}

		// 记录迭代
		iteration := Iteration{
			Number:     i + 1,
			Previous:   currentAnswer,
			Reflection: reflection,
			Improved:   improvedAnswer,
		}
		iterations = append(iterations, iteration)

		// 更新当前答案
		currentAnswer = improvedAnswer
		attempts = append(attempts, currentAnswer)
	}

	return currentAnswer, iterations, nil
}

// Critique 批判答案（找问题）
func (r *Reflection) Critique(ctx context.Context, task string, answer string) (critique string, err error) {
	prompt := fmt.Sprintf(`请批判性地评估以下答案，找出可能的问题：

任务：%s

答案：%s

请评估：
1. 答案是否完整？
2. 逻辑是否正确？
3. 是否有遗漏或错误？
4. 如何改进？

请给出详细的批判意见。`, task, answer)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	response, err := r.reflectionModel.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to critique: %w", err)
	}

	return response, nil
}

// VerifySelfConsistency 验证自洽性
func (r *Reflection) VerifySelfConsistency(ctx context.Context, task string, answer string) (isConsistent bool, issues []string, err error) {
	prompt := fmt.Sprintf(`请检查以下答案的自洽性：

任务：%s

答案：%s

请检查：
1. 答案内部是否有矛盾？
2. 前提和结论是否一致？
3. 是否有逻辑跳跃？

请列出发现的问题（如果没有问题，请回答"无问题"）。`, task, answer)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	response, err := r.reflectionModel.Chat(ctx, messages)
	if err != nil {
		return false, nil, fmt.Errorf("failed to verify self-consistency: %w", err)
	}

	// 解析响应
	isConsistent = !strings.Contains(response, "问题")
	issues = parseIssues(response)

	return isConsistent, issues, nil
}

// parseIssues 解析问题列表
func parseIssues(response string) []string {
	lines := strings.Split(response, "\n")
	issues := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "无问题") {
			issues = append(issues, line)
		}
	}

	return issues
}

// Iteration 迭代记录
type Iteration struct {
	Number     int
	Previous   string
	Reflection string
	Improved   string
}

// SetNumReflections 设置反思次数
func (r *Reflection) SetNumReflections(num int) {
	if num > 0 {
		r.numReflections = num
	}
}

// GetNumReflections 获取反思次数
func (r *Reflection) GetNumReflections() int {
	return r.numReflections
}
