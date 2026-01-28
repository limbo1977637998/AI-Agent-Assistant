package reasoning

import (
	"context"
	"fmt"
	"strings"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// ReasoningManager 推理管理器
// 整合思维链和自我反思，提供完整的推理能力
type ReasoningManager struct {
	cot        *ChainOfThought
	reflection *Reflection
	model      llm.Model
}

// NewReasoningManager 创建推理管理器
func NewReasoningManager(model llm.Model, showReasoning bool, numReflections int) *ReasoningManager {
	return &ReasoningManager{
		cot:        NewChainOfThought(model, showReasoning),
		reflection: NewReflection(model, numReflections),
		model:      model,
	}
}

// ReasonWithCoT 使用思维链推理
func (rm *ReasoningManager) ReasonWithCoT(ctx context.Context, task string) (reasoning string, answer string, err error) {
	return rm.cot.Reason(ctx, task)
}

// ReasonWithReflection 使用反思机制推理
func (rm *ReasoningManager) ReasonWithReflection(ctx context.Context, task string) (finalAnswer string, iterations []Iteration, err error) {
	// 1. 先获取初始答案
	initialAnswer, err := rm.getInitialAnswer(ctx, task)
	if err != nil {
		return "", nil, err
	}

	// 2. 反思并改进
	return rm.reflection.ReflectAndIterate(ctx, task, initialAnswer)
}

// ReasonWithCoTAndReflection 结合思维链和反思
func (rm *ReasoningManager) ReasonWithCoTAndReflection(ctx context.Context, task string) (fullReasoning string, finalAnswer string, err error) {
	// 1. 思维链推理
	reasoning, answer, err := rm.cot.Reason(ctx, task)
	if err != nil {
		return "", "", fmt.Errorf("chain of thought failed: %w", err)
	}

	// 2. 对答案进行反思
	_, improvedAnswer, err := rm.reflection.Reflect(ctx, task, []string{answer})
	if err != nil {
		// 反思失败，返回思维链的结果
		return reasoning, answer, nil
	}

	// 3. 组合完整的推理过程
	fullReasoning = fmt.Sprintf("【初步思考】\n%s\n\n【反思改进】\n经过反思和改进，得出更准确的答案。", reasoning)

	return fullReasoning, improvedAnswer, nil
}

// MultiStepReasoning 多步推理
func (rm *ReasoningManager) MultiStepReasoning(ctx context.Context, task string, steps []string) (string, error) {
	// 逐步推理
	results := make([]string, len(steps))

	for i, step := range steps {
		stepPrompt := fmt.Sprintf("任务总目标：%s\n\n当前步骤（第%d步，共%d步）：%s\n\n请完成这一步。",
			task, i+1, len(steps), step)

		messages := []models.Message{
			{Role: "user", Content: stepPrompt},
		}

		response, err := rm.model.Chat(ctx, messages)
		if err != nil {
			return "", fmt.Errorf("failed to complete step %d: %w", i+1, err)
		}

		results[i] = response
	}

	// 综合所有步骤的结果
	finalPrompt := fmt.Sprintf(`任务：%s

各步骤的结果：
%s

请综合以上所有步骤的结果，给出最终的完整答案。`, task, formatStepResults(steps, results))

	messages := []models.Message{
		{Role: "user", Content: finalPrompt},
	}

	finalAnswer, err := rm.model.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize final answer: %w", err)
	}

	return finalAnswer, nil
}

// formatStepResults 格式化步骤结果
func formatStepResults(steps, results []string) string {
	var sb string
	for i := 0; i < len(steps) && i < len(results); i++ {
		sb += fmt.Sprintf("\n步骤%d：%s\n结果：%s\n", i+1, steps[i], results[i])
	}
	return sb
}

// getInitialAnswer 获取初始答案（不使用推理）
func (rm *ReasoningManager) getInitialAnswer(ctx context.Context, task string) (string, error) {
	messages := []models.Message{
		{Role: "user", Content: task},
	}

	return rm.model.Chat(ctx, messages)
}

// VerifyAnswer 验证答案的正确性
func (rm *ReasoningManager) VerifyAnswer(ctx context.Context, task, answer string) (isCorrect bool, feedback string, err error) {
	// 1. 检查自洽性
	isConsistent, issues, err := rm.reflection.VerifySelfConsistency(ctx, task, answer)
	if err != nil {
		return false, "", err
	}

	if !isConsistent {
		feedback = "答案存在自洽性问题：\n" + formatIssues(issues)
		return false, feedback, nil
	}

	// 2. 批判性评估
	critique, err := rm.reflection.Critique(ctx, task, answer)
	if err != nil {
		return true, "", nil // 批判失败，假设答案正确
	}

	// 简单判断：如果批判中没有严重问题，则认为正确
	isCorrect = !strings.Contains(critique, "严重错误") &&
	           !strings.Contains(critique, "重大问题") &&
	           !strings.Contains(critique, "完全不正确")

	if !isCorrect {
		feedback = critique
	}

	return isCorrect, feedback, nil
}

// formatIssues 格式化问题列表
func formatIssues(issues []string) string {
	result := ""
	for i, issue := range issues {
		result += fmt.Sprintf("%d. %s\n", i+1, issue)
	}
	return result
}

// SetShowReasoning 设置是否展示推理过程
func (rm *ReasoningManager) SetShowReasoning(show bool) {
	rm.cot.ShowReasoning(show)
}

// SetNumReflections 设置反思次数
func (rm *ReasoningManager) SetNumReflections(num int) {
	rm.reflection.SetNumReflections(num)
}

// GetCoT 获取思维链推理器
func (rm *ReasoningManager) GetCoT() *ChainOfThought {
	return rm.cot
}

// GetReflection 获取反思器
func (rm *ReasoningManager) GetReflection() *Reflection {
	return rm.reflection
}
