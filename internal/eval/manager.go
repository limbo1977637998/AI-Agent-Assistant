package eval

import (
	"context"
	"fmt"

	"ai-agent-assistant/internal/llm"
)

// Manager 评估管理器
type Manager struct {
	evaluators []Evaluator
}

// NewManager 创建评估管理器
func NewManager() *Manager {
	return &Manager{
		evaluators: make([]Evaluator, 0),
	}
}

// AddEvaluator 添加评估器
func (m *Manager) AddEvaluator(evaluator Evaluator) {
	m.evaluators = append(m.evaluators, evaluator)
}

// RunEvaluations 运行所有评估
func (m *Manager) RunEvaluations(ctx context.Context, model llm.Model, dataset []TestCase) ([]*EvalResult, error) {
	results := make([]*EvalResult, 0, len(m.evaluators))

	for _, evaluator := range m.evaluators {
		result, err := evaluator.Evaluate(ctx, model, dataset)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w", evaluator.GetName(), err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateReport 生成评估报告
func (m *Manager) GenerateReport(results []*EvalResult) string {
	report := "=" + "\n"
	report += "评估报告" + "\n"
	report += "=" + "\n\n"

	for _, result := range results {
		report += fmt.Sprintf("评估器: %s\n", result.EvaluatorName)
		report += fmt.Sprintf("总用例数: %d\n", result.TotalCases)
		report += fmt.Sprintf("通过数: %d\n", result.PassedCases)
		report += fmt.Sprintf("失败数: %d\n", result.FailedCases)
		report += fmt.Sprintf("准确率: %.2f%%\n", result.Accuracy*100)
		report += fmt.Sprintf("得分: %.2f\n", result.Score)
		report += fmt.Sprintf("耗时: %v\n", result.Duration)

		if len(result.Metrics) > 0 {
			report += "\n详细指标:\n"
			for key, value := range result.Metrics {
				report += fmt.Sprintf("  %s: %v\n", key, value)
			}
		}

		report += "\n" + "-" + "\n\n"
	}

	return report
}

// GetOverallScore 获取总体得分
func (m *Manager) GetOverallScore(results []*EvalResult) float64 {
	if len(results) == 0 {
		return 0
	}

	var totalScore float64
	for _, result := range results {
		totalScore += result.Score
	}

	return totalScore / float64(len(results))
}

// EvaluatorBuilder 评估器构建器
type EvaluatorBuilder struct {
	manager *Manager
}

// NewEvaluatorBuilder 创建评估器构建器
func NewEvaluatorBuilder() *EvaluatorBuilder {
	return &EvaluatorBuilder{
		manager: NewManager(),
	}
}

// WithAccuracy 添加准确性评估
func (b *EvaluatorBuilder) WithAccuracy(scoringFunction string, judgeModel llm.Model, threshold float64) *EvaluatorBuilder {
	eval := NewAccuracyEval(scoringFunction, judgeModel, threshold)
	b.manager.AddEvaluator(eval)
	return b
}

// WithPerformance 添加性能评估
func (b *EvaluatorBuilder) WithPerformance(numRuns int) *EvaluatorBuilder {
	eval := NewPerformanceEval(numRuns)
	b.manager.AddEvaluator(eval)
	return b
}

// WithReliability 添加可靠性评估
func (b *EvaluatorBuilder) WithReliability(checkToolCalls, checkMemory bool) *EvaluatorBuilder {
	eval := NewReliabilityEval(checkToolCalls, checkMemory)
	b.manager.AddEvaluator(eval)
	return b
}

// Build 构建评估管理器
func (b *EvaluatorBuilder) Build() *Manager {
	return b.manager
}
