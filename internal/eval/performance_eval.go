package eval

import (
	"context"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// PerformanceEval 性能评估器
type PerformanceEval struct {
	numRuns int
}

// NewPerformanceEval 创建性能评估器
func NewPerformanceEval(numRuns int) *PerformanceEval {
	if numRuns <= 0 {
		numRuns = 10 // 默认运行10次
	}

	return &PerformanceEval{
		numRuns: numRuns,
	}
}

// Evaluate 评估性能
func (pe *PerformanceEval) Evaluate(ctx context.Context, model llm.Model, dataset []TestCase) (*EvalResult, error) {
	startTime := time.Now()

	result := &EvalResult{
		EvaluatorName: "Performance",
		TotalCases:    len(dataset) * pe.numRuns,
		Metrics:       make(map[string]interface{}),
		Details:       make([]CaseDetail, 0),
	}

	// 对每个测试用例运行多次
	var totalLatency time.Duration
	latencies := make([]time.Duration, 0)
	var totalTokens int
	tokenUsages := make([]int, 0)

	for _, testCase := range dataset {
		for run := 0; run < pe.numRuns; run++ {
			caseStart := time.Now()

			// 执行
			messages := []models.Message{
				{Role: "user", Content: testCase.Input},
			}

			actual, err := model.Chat(ctx, messages)
			latency := time.Since(caseStart)

			if err != nil {
				result.FailedCases++
				continue
			}

			result.PassedCases++
			latencies = append(latencies, latency)
			totalLatency += latency

			// 估算token数量（简化版：假设1个字符≈0.5个token）
			estimatedTokens := len(testCase.Input) + len(actual)
			totalTokens += estimatedTokens
			tokenUsages = append(tokenUsages, estimatedTokens)

			result.Details = append(result.Details, CaseDetail{
				Input:    testCase.Input,
				Expected: testCase.GetExpected(),
				Actual:   actual,
				Passed:   true,
				Duration: latency,
			})
		}
	}

	// 计算统计指标
	if len(latencies) > 0 {
		// 平均延迟
		avgLatency := totalLatency / time.Duration(len(latencies))

		// P50, P95, P99延迟
		p50Latency := percentile(latencies, 0.50)
		p95Latency := percentile(latencies, 0.95)
		p99Latency := percentile(latencies, 0.99)

		// 吞吐量（每秒请求数）
		totalDuration := time.Since(startTime)
		throughput := float64(len(latencies)) / totalDuration.Seconds()

		// Token相关
		avgTokens := float64(totalTokens) / float64(len(latencies))
		tokensPerSecond := float64(totalTokens) / totalDuration.Seconds()

		// 填充结果
		result.Metrics["avg_latency_ms"] = avgLatency.Milliseconds()
		result.Metrics["p50_latency_ms"] = p50Latency.Milliseconds()
		result.Metrics["p95_latency_ms"] = p95Latency.Milliseconds()
		result.Metrics["p99_latency_ms"] = p99Latency.Milliseconds()
		result.Metrics["throughput_rps"] = throughput
		result.Metrics["avg_tokens"] = avgTokens
		result.Metrics["tokens_per_second"] = tokensPerSecond
		result.Metrics["total_requests"] = len(latencies)
		result.Metrics["num_runs"] = pe.numRuns

		result.Score = throughput // 使用吞吐量作为得分
	}

	result.Duration = time.Since(startTime)

	return result, nil
}

// GetName 获取评估器名称
func (pe *PerformanceEval) GetName() string {
	return "PerformanceEval"
}

// percentile 计算百分位数
func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// 排序
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	// 计算索引
	index := int(float64(len(latencies)-1) * p)

	return latencies[index]
}

// ReliabilityEval 可靠性评估器
type ReliabilityEval struct {
	checkToolCalls bool
	checkMemory    bool
}

// NewReliabilityEval 创建可靠性评估器
func NewReliabilityEval(checkToolCalls, checkMemory bool) *ReliabilityEval {
	return &ReliabilityEval{
		checkToolCalls: checkToolCalls,
		checkMemory:    checkMemory,
	}
}

// Evaluate 评估可靠性
func (re *ReliabilityEval) Evaluate(ctx context.Context, model llm.Model, dataset []TestCase) (*EvalResult, error) {
	startTime := time.Now()

	result := &EvalResult{
		EvaluatorName: "Reliability",
		TotalCases:    len(dataset),
		Metrics:       make(map[string]interface{}),
		Details:       make([]CaseDetail, 0, len(dataset)),
	}

	var errorCount int
	var successCount int

	for _, testCase := range dataset {
		caseStart := time.Now()

		messages := []models.Message{
			{Role: "user", Content: testCase.Input},
		}

		_, err := model.Chat(ctx, messages)
		latency := time.Since(caseStart)

		if err != nil {
			errorCount++
			result.FailedCases++
			result.Details = append(result.Details, CaseDetail{
				Input:    testCase.Input,
				Expected: testCase.GetExpected(),
				Actual:   "",
				Passed:   false,
				Score:    0,
				Error:    err.Error(),
				Duration: latency,
			})
		} else {
			successCount++
			result.PassedCases++
			result.Details = append(result.Details, CaseDetail{
				Input:    testCase.Input,
				Expected: testCase.GetExpected(),
				Actual:   "success",
				Passed:   true,
				Score:    1,
				Duration: latency,
			})
		}
	}

	// 计算可靠性指标
	successRate := float64(successCount) / float64(result.TotalCases)
	errorRate := float64(errorCount) / float64(result.TotalCases)

	result.Metrics["success_rate"] = successRate
	result.Metrics["error_rate"] = errorRate
	result.Metrics["error_count"] = errorCount
	result.Metrics["success_count"] = successCount

	result.Accuracy = successRate
	result.Score = successRate
	result.Duration = time.Since(startTime)

	return result, nil
}

// GetName 获取评估器名称
func (re *ReliabilityEval) GetName() string {
	return "ReliabilityEval"
}
