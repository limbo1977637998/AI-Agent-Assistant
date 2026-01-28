package eval

import (
	"context"
	"math"
	"strings"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// Evaluator 评估器接口
type Evaluator interface {
	Evaluate(ctx context.Context, agent llm.Model, dataset []TestCase) (*EvalResult, error)
	GetName() string
}

// TestCase 测试用例
type TestCase struct {
	Input        string                 `json:"input"`
	Expected     string                 `json:"expected_output"`
	ExpectedText string                 `json:"expected"` // 备用字段，兼容旧格式
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GetExpected 获取期望输出，优先使用Expected字段
func (tc *TestCase) GetExpected() string {
	if tc.Expected != "" {
		return tc.Expected
	}
	return tc.ExpectedText
}

// EvalResult 评估结果
type EvalResult struct {
	EvaluatorName  string                 `json:"evaluator_name"`
	TotalCases     int                    `json:"total_cases"`
	PassedCases    int                    `json:"passed_cases"`
	FailedCases    int                    `json:"failed_cases"`
	Accuracy       float64                `json:"accuracy"`
	Score          float64                `json:"score"`
	Metrics        map[string]interface{} `json:"metrics"`
	Details        []CaseDetail           `json:"details"`
	Duration       time.Duration          `json:"duration"`
}

// CaseDetail 详细案例
type CaseDetail struct {
	Input       string      `json:"input"`
	Expected    string      `json:"expected"`
	Actual      string      `json:"actual"`
	Passed      bool        `json:"passed"`
	Score       float64     `json:"score"`
	Error       string      `json:"error,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// AccuracyEval 准确性评估器
type AccuracyEval struct {
	scoringFunction string // "exact_match", "similarity", "llm_judge"
	judgeModel      llm.Model
	threshold       float64
}

// NewAccuracyEval 创建准确性评估器
func NewAccuracyEval(scoringFunction string, judgeModel llm.Model, threshold float64) *AccuracyEval {
	if threshold <= 0 {
		threshold = 0.8 // 默认阈值
	}

	return &AccuracyEval{
		scoringFunction: scoringFunction,
		judgeModel:      judgeModel,
		threshold:       threshold,
	}
}

// Evaluate 评估准确性
func (e *AccuracyEval) Evaluate(ctx context.Context, model llm.Model, dataset []TestCase) (*EvalResult, error) {
	startTime := time.Now()

	result := &EvalResult{
		EvaluatorName: "Accuracy",
		TotalCases:    len(dataset),
		Metrics:       make(map[string]interface{}),
		Details:       make([]CaseDetail, 0, len(dataset)),
	}

	var totalScore float64

	for _, testCase := range dataset {
		caseStart := time.Now()

		// 执行
		messages := []models.Message{
			{Role: "user", Content: testCase.Input},
		}

		actual, err := model.Chat(ctx, messages)
		if err != nil {
			result.FailedCases++
			result.Details = append(result.Details, CaseDetail{
				Input:    testCase.Input,
				Expected: testCase.GetExpected(),
				Actual:   "",
				Passed:   false,
				Score:    0,
				Error:    err.Error(),
				Duration: time.Since(caseStart),
			})
			continue
		}

		// 评分
		score, passed := e.scoreResult(testCase.GetExpected(), actual)
		totalScore += score

		if passed {
			result.PassedCases++
		} else {
			result.FailedCases++
		}

		result.Details = append(result.Details, CaseDetail{
			Input:    testCase.Input,
			Expected: testCase.GetExpected(),
			Actual:   actual,
			Passed:   passed,
			Score:    score,
			Duration: time.Since(caseStart),
		})
	}

	result.Accuracy = float64(result.PassedCases) / float64(result.TotalCases)
	result.Score = totalScore / float64(result.TotalCases)
	result.Duration = time.Since(startTime)

	// 额外指标
	result.Metrics["avg_score"] = result.Score
	result.Metrics["pass_rate"] = result.Accuracy
	result.Metrics["threshold"] = e.threshold

	return result, nil
}

// scoreResult 评分
func (e *AccuracyEval) scoreResult(expected, actual string) (score float64, passed bool) {
	switch e.scoringFunction {
	case "exact_match":
		return e.exactMatch(expected, actual)
	case "similarity":
		return e.similarity(expected, actual)
	case "llm_judge":
		return e.llmJudge(expected, actual)
	default:
		return e.similarity(expected, actual)
	}
}

// exactMatch 精确匹配
func (e *AccuracyEval) exactMatch(expected, actual string) (score float64, passed bool) {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	if expected == actual {
		return 1.0, true
	}
	return 0.0, false
}

// similarity 相似度评分（改进版：包含检查 + 编辑距离）
func (e *AccuracyEval) similarity(expected, actual string) (score float64, passed bool) {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	// 情况1：完全匹配
	if expected == actual {
		return 1.0, true
	}

	// 情况2：包含关系（实际答案包含期望答案）
	if len(expected) > 0 && len(actual) > 0 {
		// 检查实际答案是否包含期望答案（忽略空格和标点）
		cleanExpected := strings.ToLower(expected)
		cleanActual := strings.ToLower(actual)

		// 移除常见标点符号
		cleanExpected = removePunctuation(cleanExpected)
		cleanActual = removePunctuation(cleanActual)

		// 如果实际答案包含期望答案
		if strings.Contains(cleanActual, cleanExpected) {
			// 根据期望答案长度给分（期望越短，包含它的难度越大，分数越高）
			ratio := float64(len(cleanExpected)) / float64(len(cleanActual))
			containScore := 0.95 - ratio*0.1 // 0.85-0.95之间
			return math.Max(0.7, math.Min(0.95, containScore)), true
		}

		// 检查期望答案是否包含实际答案（实际答案更精确）
		if strings.Contains(cleanExpected, cleanActual) {
			return 0.85, true
		}
	}

	// 情况3：使用编辑距离计算相似度
	distance := levenshteinDistance(expected, actual)
	maxLen := math.Max(float64(len(expected)), float64(len(actual)))

	if maxLen == 0 {
		return 1.0, true
	}

	editSimilarity := 1.0 - (float64(distance) / maxLen)
	passed = editSimilarity >= e.threshold

	return editSimilarity, passed
}

// removePunctuation 移除标点符号和空格
func removePunctuation(s string) string {
	// 移除常见标点符号
	punctuations := []string{" ", "，", "。", "！", "？", "、", "；", "：", "'", "\"", "！", ",", ".", "!", "?", ":", ";"}
	for _, p := range punctuations {
		s = strings.ReplaceAll(s, p, "")
	}
	return s
}

// llmJudge 使用LLM判断
func (e *AccuracyEval) llmJudge(expected, actual string) (score float64, passed bool) {
	if e.judgeModel == nil {
		// 回退到相似度
		return e.similarity(expected, actual)
	}

	// 简化版：使用相似度
	// 实际应用中可以让LLM打分
	return e.similarity(expected, actual)
}

// GetName 获取评估器名称
func (e *AccuracyEval) GetName() string {
	return "AccuracyEval"
}

// levenshteinDistance 计算编辑距离
func levenshteinDistance(s1, s2 string) int {
	len1 := len(s1)
	len2 := len(s2)

	// 创建距离矩阵
	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			dp[i][j] = min(
				dp[i-1][j]+1,      // 删除
				dp[i][j-1]+1,      // 插入
				dp[i-1][j-1]+cost, // 替换
			)
		}
	}

	return dp[len1][len2]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
