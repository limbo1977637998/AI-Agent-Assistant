package reasoning

import (
	"context"
	"fmt"
	"strings"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// ChainOfThought 思维链推理
// 让模型逐步思考问题，展示推理过程
type ChainOfThought struct {
	reasoningModel llm.Model
	showReasoning  bool // 是否展示思考过程
}

// NewChainOfThought 创建思维链推理器
func NewChainOfThought(model llm.Model, showReasoning bool) *ChainOfThought {
	return &ChainOfThought{
		reasoningModel: model,
		showReasoning:  showReasoning,
	}
}

// Reason 执行推理
func (cot *ChainOfThought) Reason(ctx context.Context, task string) (reasoning string, answer string, err error) {
	// 构造思维链提示
	prompt := fmt.Sprintf(`请逐步思考以下问题，展示你的推理过程。

要求：
1. 先分析问题的关键点
2. 逐步推理，展示思考过程
3. 考虑多种可能性
4. 得出结论并解释理由

问题：%s

请按以下格式回答：
【思考过程】
（你的思考步骤）

【答案】
（你的结论）`, task)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	// 调用模型
	response, err := cot.reasoningModel.Chat(ctx, messages)
	if err != nil {
		return "", "", fmt.Errorf("failed to reason: %w", err)
	}

	// 解析响应
	return cot.parseResponse(response)
}

// parseResponse 解析模型响应，分离思考过程和答案
func (cot *ChainOfThought) parseResponse(response string) (reasoning string, answer string, err error) {
	// 尝试分离【思考过程】和【答案】
	parts := strings.Split(response, "【答案】")

	if len(parts) >= 2 {
		// 提取思考过程
		reasoningPart := parts[0]
		reasoningPart = strings.TrimPrefix(reasoningPart, "【思考过程】")
		reasoningPart = strings.TrimSpace(reasoningPart)

		// 提取答案
		answerPart := parts[1]
		answerPart = strings.TrimSpace(answerPart)

		return reasoningPart, answerPart, nil
	}

	// 如果无法分离，返回完整响应作为答案
	return "", response, nil
}

// ReasonWithSteps 带步骤的思维链推理
func (cot *ChainOfThought) ReasonWithSteps(ctx context.Context, task string, steps []string) (string, error) {
	// 构造分步推理提示
	stepPrompt := "请按照以下步骤逐步思考：\n"
	for i, step := range steps {
		stepPrompt += fmt.Sprintf("步骤%d：%s\n", i+1, step)
	}
	stepPrompt += fmt.Sprintf("\n最终任务：%s\n\n请逐步完成每个步骤。", task)

	messages := []models.Message{
		{Role: "user", Content: stepPrompt},
	}

	response, err := cot.reasoningModel.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to reason with steps: %w", err)
	}

	return response, nil
}

// GetReasoningPrompt 获取思维链提示模板
func (cot *ChainOfThought) GetReasoningPrompt(task string) string {
	return fmt.Sprintf(`让我们逐步思考这个问题。

问题：%s

请思考：
1. 这个问题的核心是什么？
2. 需要哪些信息来解决它？
3. 如何一步步推理？
4. 最终的答案是什么？`, task)
}

// ShowReasoning 设置是否展示思考过程
func (cot *ChainOfThought) ShowReasoning(show bool) {
	cot.showReasoning = show
}

// ShouldShowReasoning 是否展示思考过程
func (cot *ChainOfThought) ShouldShowReasoning() bool {
	return cot.showReasoning
}
