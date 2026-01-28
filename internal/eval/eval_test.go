package eval

import (
	"context"
	"testing"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// MockEvalModel 模拟评估模型
type MockEvalModel struct{}

func (m *MockEvalModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	return "测试响应", nil
}

func (m *MockEvalModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "测试响应"
	close(ch)
	return ch, nil
}

func (m *MockEvalModel) SupportsToolCalling() bool {
	return false
}

func (m *MockEvalModel) SupportsEmbedding() bool {
	return false
}

func (m *MockEvalModel) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, nil
}

func (m *MockEvalModel) GetModelName() string {
	return "mock-eval"
}

func (m *MockEvalModel) GetProviderName() string {
	return "mock"
}

// TestAccuracyEval 测试准确性评估
func TestAccuracyEval(t *testing.T) {
	model := &MockEvalModel{}
	eval := NewAccuracyEval("exact_match", nil, 0.8)

	dataset := []TestCase{
		{
			Input:    "1+1=?",
			Expected: "2",
		},
		{
			Input:    "2+2=?",
			Expected: "4",
		},
		{
			Input:    "3+3=?",
			Expected: "6",
		},
	}

	result, err := eval.Evaluate(context.Background(), model, dataset)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	t.Logf("Accuracy: %.2f%%", result.Accuracy*100)
	t.Logf("Score: %.2f", result.Score)
	t.Logf("Total cases: %d", result.TotalCases)
	t.Logf("Passed cases: %d", result.PassedCases)
	t.Logf("Failed cases: %d", result.FailedCases)

	if result.TotalCases != 3 {
		t.Errorf("Expected 3 total cases, got %d", result.TotalCases)
	}
}

// TestLevenshteinDistance 测试编辑距离
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1      string
		s2      string
		expect  int
	}{
		{"kitten", "sitting", 3},
		{"test", "test", 0},
		{"", "test", 4},
		{"test", "", 4},
		{"abc", "xyz", 3},
	}

	for _, tc := range tests {
		distance := levenshteinDistance(tc.s1, tc.s2)
		if distance != tc.expect {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d",
				tc.s1, tc.s2, distance, tc.expect)
		}
	}
}

// TestPerformanceEval 测试性能评估
func TestPerformanceEval(t *testing.T) {
	model := &MockEvalModel{}
	eval := NewPerformanceEval(5) // 运行5次

	dataset := []TestCase{
		{Input: "测试输入1", Expected: "测试输出1"},
		{Input: "测试输入2", Expected: "测试输出2"},
	}

	result, err := eval.Evaluate(context.Background(), model, dataset)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	t.Logf("Duration: %v", result.Duration)

	// 检查指标
	if avgLatency, ok := result.Metrics["avg_latency_ms"]; ok {
		t.Logf("Avg latency: %v ms", avgLatency)
	}

	if throughput, ok := result.Metrics["throughput_rps"]; ok {
		t.Logf("Throughput: %v rps", throughput)
	}
}

// TestReliabilityEval 测试可靠性评估
func TestReliabilityEval(t *testing.T) {
	model := &MockEvalModel{}
	eval := NewReliabilityEval(false, false)

	dataset := []TestCase{
		{Input: "测试输入1", Expected: "测试输出1"},
		{Input: "测试输入2", Expected: "测试输出2"},
	}

	result, err := eval.Evaluate(context.Background(), model, dataset)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	t.Logf("Success rate: %.2f%%", result.Accuracy*100)
	t.Logf("Error rate: %.2f%%", result.Metrics["error_rate"])
}

// TestEvaluatorBuilder 测试评估器构建器
func TestEvaluatorBuilder(t *testing.T) {
	model := &MockEvalModel{}

	builder := NewEvaluatorBuilder()
	builder.
		WithAccuracy("similarity", model, 0.7).
		WithPerformance(3).
		WithReliability(false, false)

	manager := builder.Build()

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	evaluators := len(manager.evaluators)
	if evaluators != 3 {
		t.Errorf("Expected 3 evaluators, got %d", evaluators)
	}

	t.Logf("Built manager with %d evaluators", evaluators)
}

// TestPercentile 测试百分位数计算
func TestPercentile(t *testing.T) {
	latencies := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
	}

	// 测试P50（中位数）
	p50 := percentile(latencies, 0.50)
	if p50 != 300*time.Millisecond {
		t.Errorf("Expected P50 to be 300ms, got %v", p50)
	}

	// 测试P95
	p95 := percentile(latencies, 0.95)
	if p95 != 500*time.Millisecond {
		t.Errorf("Expected P95 to be 500ms, got %v", p95)
	}

	t.Logf("P50: %v, P95: %v", p50, p95)
}
