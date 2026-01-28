package task

import (
	"context"
	"fmt"
	"time"
)

// Aggregator 结果聚合器接口
type Aggregator interface {
	Aggregate(ctx context.Context, results []*TaskResult) (*AggregateResult, error)
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID     string                 `json:"task_id"`
	TaskGoal   string                 `json:"task_goal"`
	Type       string                 `json:"type"`         // 任务类型
	Status     TaskStatus             `json:"status"`
	Output     interface{}            `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
	AgentUsed  string                 `json:"agent_used,omitempty"`
}

// AggregateResult 聚合结果
type AggregateResult struct {
	Success     bool                   `json:"success"`
	TotalTasks  int                    `json:"total_tasks"`
	Completed   int                    `json:"completed"`
	Failed      int                    `json:"failed"`
	Output      interface{}            `json:"output"`
	Summary     string                 `json:"summary"`
	Details     []*AggregateDetail     `json:"details"`
	Metrics     map[string]interface{} `json:"metrics"`
	AggregatedAt time.Time             `json:"aggregated_at"`
}

// AggregateDetail 聚合详情
type AggregateDetail struct {
	TaskID    string      `json:"task_id"`
	Goal      string      `json:"goal"`
	Status    TaskStatus `json:"status"`
	Output    interface{} `json:"output"`
	Contribution float64   `json:"contribution"` // 贡献度 0-1
}

// SimpleAggregator 简单聚合器（按顺序拼接）
type SimpleAggregator struct{}

// NewSimpleAggregator 创建简单聚合器
func NewSimpleAggregator() *SimpleAggregator {
	return &SimpleAggregator{}
}

// Aggregate 简单聚合（按顺序拼接结果）
func (a *SimpleAggregator) Aggregate(ctx context.Context, results []*TaskResult) (*AggregateResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	completed := 0
	failed := 0
	details := make([]*AggregateDetail, 0, len(results))
	outputs := make([]interface{}, 0)

	for i, result := range results {
		detail := &AggregateDetail{
			TaskID: result.TaskID,
			Goal:   result.TaskGoal,
			Status: result.Status,
			Output: result.Output,
		}

		if result.Status == TaskStatusCompleted {
			completed++
			detail.Contribution = 1.0 / float64(len(results))
			outputs = append(outputs, result.Output)
		} else if result.Status == TaskStatusFailed {
			failed++
			detail.Contribution = 0
		}

		details[i] = detail
	}

	aggResult := &AggregateResult{
		Success:    failed == 0,
		TotalTasks: len(results),
		Completed:  completed,
		Failed:     failed,
		Output:     outputs,
		Summary:    fmt.Sprintf("聚合了%d个任务结果", len(results)),
		Details:    details,
		Metrics: map[string]interface{}{
			"completion_rate": float64(completed) / float64(len(results)),
			"success_rate":    float64(completed) / float64(len(results)),
		},
		AggregatedAt: time.Now(),
	}

	return aggResult, nil
}

// SummaryAggregator 总结聚合器（生成文本摘要）
type SummaryAggregator struct {
	llmModel LLMModel
}

// NewSummaryAggregator 创建总结聚合器
func NewSummaryAggregator(llmModel LLMModel) *SummaryAggregator {
	return &SummaryAggregator{
		llmModel: llmModel,
	}
}

// Aggregate 生成摘要聚合
func (a *SummaryAggregator) Aggregate(ctx context.Context, results []*TaskResult) (*AggregateResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	// 准备结果摘要
	resultsText := ""
	for i, result := range results {
		resultsText += fmt.Sprintf("\n%d. 任务：%s\n", i+1, result.TaskGoal)
		resultsText += fmt.Sprintf("   状态：%s\n", result.Status)
		if result.Status == TaskStatusCompleted {
			resultsText += fmt.Sprintf("   结果：%v\n", result.Output)
		} else if result.Status == TaskStatusFailed {
			resultsText += fmt.Sprintf("   错误：%s\n", result.Error)
		}
	}

	// 使用LLM生成摘要
	prompt := fmt.Sprintf(`请根据以下任务执行结果，生成一个简洁的总结摘要：

任务结果：
%s

请生成一个包含以下内容的JSON格式响应：
{
  "summary": "整体总结",
  "key_findings": ["发现1", "发现2"],
  "final_output": "最终输出"
}`, resultsText)

	messages := []Message{
		{Role: "system", Content: "你是一个结果聚合专家，擅长总结多个任务的执行结果。"},
		{Role: "user", Content: prompt},
	}

	response, err := a.llmModel.Chat(ctx, messages)
	if err != nil {
		// 降级到简单聚合
		return NewSimpleAggregator().Aggregate(ctx, results)
	}

	// 解析LLM响应
	// 简化实现：直接使用响应作为摘要
	completed := 0
	failed := 0
	for _, result := range results {
		if result.Status == TaskStatusCompleted {
			completed++
		} else {
			failed++
		}
	}

	aggResult := &AggregateResult{
		Success:    failed == 0,
		TotalTasks: len(results),
		Completed:  completed,
		Failed:     failed,
		Output:     response,
		Summary:    response,
		Details:    make([]*AggregateDetail, 0),
		Metrics: map[string]interface{}{
			"completion_rate": float64(completed) / float64(len(results)),
		},
		AggregatedAt: time.Now(),
	}

	return aggResult, nil
}

// WeightedAggregator 加权聚合器
type WeightedAggregator struct {
	weights map[string]float64 // task_type -> weight
}

// NewWeightedAggregator 创建加权聚合器
func NewWeightedAggregator() *WeightedAggregator {
	return &WeightedAggregator{
		weights: map[string]float64{
			"research": 0.2,
			"analyze":  0.3,
			"write":   0.5,
		},
	}
}

// Aggregate 加权聚合
func (a *WeightedAggregator) Aggregate(ctx context.Context, results []*TaskResult) (*AggregateResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	completed := 0
	failed := 0
	totalWeight := 0.0
	details := make([]*AggregateDetail, 0)

	for _, result := range results {
		detail := &AggregateDetail{
			TaskID: result.TaskID,
			Goal:   result.TaskGoal,
			Status: result.Status,
			Output: result.Output,
		}

		// 获取权重（基于任务类型）
		weight := a.getWeight(result, len(results))

		if result.Status == TaskStatusCompleted {
			completed++
			detail.Contribution = weight
			totalWeight += weight
		} else {
			failed++
			detail.Contribution = 0
		}

		details = append(details, detail)
	}

	// 归一化贡献度
	for _, detail := range details {
		if totalWeight > 0 {
			detail.Contribution = detail.Contribution / totalWeight
		}
	}

	aggResult := &AggregateResult{
		Success:    failed == 0,
		TotalTasks: len(results),
		Completed:  completed,
		Failed:     failed,
		Summary:    fmt.Sprintf("加权聚合了%d个任务结果", len(results)),
		Details:    details,
		Metrics: map[string]interface{}{
			"total_weight":    totalWeight,
			"completion_rate": float64(completed) / float64(len(results)),
		},
		AggregatedAt: time.Now(),
	}

	return aggResult, nil
}

// getWeight 获取任务权重
func (a *WeightedAggregator) getWeight(result *TaskResult, totalResults int) float64 {
	// 简化实现：根据任务关键词判断权重
	goal := result.TaskGoal

	if weight, ok := a.weights[result.Type]; ok {
		return weight
	}

	// 基于关键词判断
	for keyword, weight := range a.weights {
		if contains(goal, keyword) {
			return weight
		}
	}

	// 默认平均权重
	if totalResults > 0 {
		return 1.0 / float64(totalResults)
	}
	return 0.0
}

// ConsensusAggregator 共识聚合器（多数投票）
type ConsensusAggregator struct{}

// NewConsensusAggregator 创建共识聚合器
func NewConsensusAggregator() *ConsensusAggregator {
	return &ConsensusAggregator{}
}

// Aggregate 共识聚合
func (a *ConsensusAggregator) Aggregate(ctx context.Context, results []*TaskResult) (*AggregateResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	// 统计结果的投票
	votes := make(map[interface{}]int)
	completed := 0
	failed := 0

	for _, result := range results {
		if result.Status == TaskStatusCompleted {
			completed++
			votes[result.Output]++
		} else {
			failed++
		}
	}

	// 找出票数最多的结果
	var maxVotes int
	var consensusOutput interface{}
	for output, votes := range votes {
		if votes > maxVotes {
			maxVotes = votes
			consensusOutput = output
		}
	}

	// 计算共识度
	consensus := float64(maxVotes) / float64(completed)

	aggResult := &AggregateResult{
		Success:    failed == 0,
		TotalTasks: len(results),
		Completed:  completed,
		Failed:     failed,
		Output:     consensusOutput,
		Summary:    fmt.Sprintf("共识聚合：%d个任务，共识度：%.2f%%", len(results), consensus*100),
		Details:    make([]*AggregateDetail, 0),
		Metrics: map[string]interface{}{
			"consensus":       consensus,
			"max_votes":       maxVotes,
			"total_completed": completed,
		},
		AggregatedAt: time.Now(),
	}

	return aggResult, nil
}

// Helper function
func contains(text string, keyword string) bool {
	return len(text) > 0 && (text == keyword ||
		(len(keyword) > 0 && (text == keyword ||
		text[:min(len(text), len(keyword))] == keyword[:min(len(text), len(keyword))])))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
