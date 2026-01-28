package reasoning

import (
	"context"
	"testing"

	"ai-agent-assistant/internal/llm"
)

// MockReasoningModel 模拟推理模型
type MockReasoningModel struct {
	// 用于模拟不同类型的响应
	cotResponse    string
	reflectionResponse string
}

func (m *MockReasoningModel) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	content := messages[0].Content

	// 检测是否是思维链请求
	if contains(content, "思考") || contains(content, "逐步") {
		if m.cotResponse != "" {
			return m.cotResponse, nil
		}
		return `【思考过程】
首先分析问题的核心是数学计算。
然后计算 25 * 4 = 100。

【答案】
25乘以4等于100。`, nil
	}

	// 检测是否是反思请求
	if contains(content, "检查") || contains(content, "反思") {
		if m.reflectionResponse != "" {
			return m.reflectionResponse, nil
		}
		return `【反思】
之前的答案计算正确，步骤清晰。

【改进后的答案】
答案正确，无需改进。`, nil
	}

	return "默认响应", nil
}

func (m *MockReasoningModel) ChatStream(ctx context.Context, messages []llm.Message) (<-chan string, error) {
	ch := make(chan string, 1)
	resp, _ := m.Chat(ctx, messages)
	ch <- resp
	close(ch)
	return ch, nil
}

func (m *MockReasoningModel) SupportsToolCalling() bool {
	return false
}

func (m *MockReasoningModel) SupportsEmbedding() bool {
	return false
}

func (m *MockReasoningModel) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, nil
}

func (m *MockReasoningModel) GetModelName() string {
	return "mock-reasoning"
}

func (m *MockReasoningModel) GetProviderName() string {
	return "mock"
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr
}

// TestChainOfThought 测试思维链推理
func TestChainOfThought(t *testing.T) {
	model := &MockReasoningModel{
		cotResponse: `【思考过程】
第一步：理解题目要求计算 25 * 4
第二步：进行乘法运算
第三步：得出结果 100

【答案】
25乘以4等于100`,
	}

	cot := NewChainOfThought(model, true)

	// 测试推理
	reasoning, answer, err := cot.Reason(context.Background(), "25 * 4 = ?")
	if err != nil {
		t.Fatalf("Reason failed: %v", err)
	}

	t.Logf("Reasoning: %s", reasoning)
	t.Logf("Answer: %s", answer)

	// 验证思考过程
	if reasoning == "" {
		t.Error("Reasoning should not be empty")
	}

	// 验证答案
	if answer == "" {
		t.Error("Answer should not be empty")
	}
}

// TestReflection 测试自我反思
func TestReflection(t *testing.T) {
	model := &MockReasoningModel{
		reflectionResponse: `【反思】
答案"25 * 4 = 100"是正确的。

【改进后的答案】
确认答案是100。`,
	}

	reflection := NewReflection(model, 1)

	previousAttempts := []string{
		"25 * 4 = 100",
	}

	// 测试反思
	reflectionText, improvedAnswer, err := reflection.Reflect(context.Background(), "25 * 4 = ?", previousAttempts)
	if err != nil {
		t.Fatalf("Reflect failed: %v", err)
	}

	t.Logf("Reflection: %s", reflectionText)
	t.Logf("Improved answer: %s", improvedAnswer)

	if reflectionText == "" {
		t.Error("Reflection should not be empty")
	}

	if improvedAnswer == "" {
		t.Error("Improved answer should not be empty")
	}
}

// TestReflectionAndIterate 测试反思迭代
func TestReflectionAndIterate(t *testing.T) {
	model := &MockReasoningModel{
		reflectionResponse: `【反思】
需要更详细的解释。

【改进后的答案】
25乘以4等于100，因为25乘4就是25加4次，即25+25+25+25=100。`,
	}

	reflection := NewReflection(model, 2)

	initialAnswer := "25 * 4 = 100"

	// 测试反思迭代
	finalAnswer, iterations, err := reflection.ReflectAndIterate(context.Background(), "25 * 4 = ?", initialAnswer)
	if err != nil {
		t.Fatalf("ReflectAndIterate failed: %v", err)
	}

	t.Logf("Final answer: %s", finalAnswer)
	t.Logf("Iterations: %d", len(iterations))

	if len(iterations) != reflection.GetNumReflections() {
		t.Errorf("Expected %d iterations, got %d", reflection.GetNumReflections(), len(iterations))
	}
}

// TestReasoningManager 测试推理管理器
func TestReasoningManager(t *testing.T) {
	model := &MockReasoningModel{
		cotResponse: `【思考过程】
分析问题...
计算...
得出结论

【答案】
答案是42`,
	}

	manager := NewReasoningManager(model, true, 1)

	// 测试思维链推理
	reasoning, answer, err := manager.ReasonWithCoT(context.Background(), "测试任务")
	if err != nil {
		t.Fatalf("ReasonWithCoT failed: %v", err)
	}

	t.Logf("CoT Reasoning: %s", reasoning)
	t.Logf("CoT Answer: %s", answer)

	// 测试反思
	finalAnswer, iterations, err := manager.ReasonWithReflection(context.Background(), "测试任务")
	if err != nil {
		t.Fatalf("ReasonWithReflection failed: %v", err)
	}

	t.Logf("Final Answer: %s", finalAnswer)
	t.Logf("Iterations: %d", len(iterations))
}
