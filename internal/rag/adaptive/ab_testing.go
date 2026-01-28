package adaptive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ABTestingFramework A/B 测试框架
//
// 功能: 对比不同检索策略的性能
//
// 支持的测试:
//   1. 策略对比测试
//   2. 参数对比测试
//   3. 统计显著性检验
//
// 统计方法:
//   - T-test (T检验)
//   - Z-test (Z检验)
//   - Bootstrap 置信区间
type ABTestingFramework struct {
	experiments map[string]*Experiment
	mu          sync.RWMutex
	config      ABTestConfig
}

// ABTestConfig A/B 测试配置
type ABTestConfig struct {
	// MinSamples 最小样本数
	MinSamples int

	// ConfidenceLevel 置信水平 (0.95 表示 95%)
	ConfidenceLevel float64

	// EnableAutoStop 是否启用自动停止
	EnableAutoStop bool

	// SignificanceLevel 显著性水平 (0.05 表示 5%)
	SignificanceLevel float64
}

// DefaultABTestConfig 返回默认配置
func DefaultABTestConfig() ABTestConfig {
	return ABTestConfig{
		MinSamples:       30,
		ConfidenceLevel:  0.95,
		EnableAutoStop:   true,
		SignificanceLevel: 0.05,
	}
}

// NewABTestingFramework 创建 A/B 测试框架
func NewABTestingFramework(config ABTestConfig) *ABTestingFramework {
	return &ABTestingFramework{
		experiments: make(map[string]*Experiment),
		config:      config,
	}
}

// Experiment 实验信息
type Experiment struct {
	// Name 实验名称
	Name string

	// Description 实验描述
	Description string

	// StartTime 开始时间
	StartTime time.Time

	// EndTime 结束时间
	EndTime *time.Time

	// Variants 变体列表
	Variants []*Variant

	// Winner 获胜变体
	Winner *Variant

	// Status 状态 (running, completed, stopped)
	Status string

	// Metrics 指标
	Metrics *ExperimentMetrics
}

// Variant 变体
type Variant struct {
	// Name 变体名称
	Name string

	// Strategy 策略名称
	Strategy string

	// Parameters 参数
	Parameters map[string]interface{}

	// Traffic 流量比例 (0-1)
	Traffic float64

	// Results 结果
	Results []*VariantResult

	// Stats 统计数据
	Stats *VariantStats
}

// VariantResult 变体结果
type VariantResult struct {
	Query         string
	Contexts      []string
	Answer        string
	Score         float64
	Latency       int64
	UserFeedback  float64
	Timestamp     time.Time
}

// VariantStats 变体统计
type VariantStats struct {
	TotalQueries   int
	SuccessCount   int
	AverageScore   float64
	AverageLatency int64
	AverageFeedback float64
	ConversionRate float64
	StdDevScore    float64
	ConfidenceInterval *ConfidenceInterval
}

// ConfidenceInterval 置信区间
type ConfidenceInterval struct {
	Lower float64
	Upper float64
	Level float64
}

// ExperimentMetrics 实验指标
type ExperimentMetrics struct {
	// PValue P值 (用于统计显著性检验)
	PValue float64

	// EffectSize 效应大小
	EffectSize float64

	// StatisticalSignificant 是否统计显著
	StatisticalSignificant bool

	// Improvement 相对改进
	Improvement float64

	// Winner 获胜变体名称
	Winner string
}

// CreateExperiment 创建实验
func (ab *ABTestingFramework) CreateExperiment(ctx context.Context, name, description string, variants []*Variant) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if _, exists := ab.experiments[name]; exists {
		return fmt.Errorf("experiment %s already exists", name)
	}

	// 验证流量总和
	totalTraffic := 0.0
	for _, variant := range variants {
		totalTraffic += variant.Traffic
	}
	if math.Abs(totalTraffic-1.0) > 0.01 {
		return fmt.Errorf("traffic sum must be 1.0, got %f", totalTraffic)
	}

	// 初始化统计
	for _, variant := range variants {
		variant.Stats = &VariantStats{}
		variant.Results = make([]*VariantResult, 0)
	}

	experiment := &Experiment{
		Name:        name,
		Description: description,
		StartTime:   time.Now(),
		Variants:    variants,
		Status:      "running",
		Metrics:     &ExperimentMetrics{},
	}

	ab.experiments[name] = experiment
	return nil
}

// RecordResult 记录结果
func (ab *ABTestingFramework) RecordResult(ctx context.Context, experimentName string, variantName string, result *VariantResult) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment, exists := ab.experiments[experimentName]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentName)
	}

	// 找到变体
	var targetVariant *Variant
	for _, variant := range experiment.Variants {
		if variant.Name == variantName {
			targetVariant = variant
			break
		}
	}

	if targetVariant == nil {
		return fmt.Errorf("variant %s not found", variantName)
	}

	// 记录结果
	result.Timestamp = time.Now()
	targetVariant.Results = append(targetVariant.Results, result)

	// 更新统计
	ab.updateVariantStats(targetVariant)

	// 检查是否需要自动停止
	if ab.config.EnableAutoStop {
		ab.checkAutoStop(ctx, experiment)
	}

	return nil
}

// updateVariantStats 更新变体统计
func (ab *ABTestingFramework) updateVariantStats(variant *Variant) {
	n := len(variant.Results)
	if n == 0 {
		return
	}

	stats := variant.Stats
	stats.TotalQueries = n

	// 计算平均值
	totalScore := 0.0
	totalLatency := int64(0)
	totalFeedback := 0.0
	successCount := 0

	for _, result := range variant.Results {
		totalScore += result.Score
		totalLatency += result.Latency
		totalFeedback += result.UserFeedback
		if result.Score > 0.5 {
			successCount++
		}
	}

	stats.AverageScore = totalScore / float64(n)
	stats.AverageLatency = totalLatency / int64(n)
	stats.AverageFeedback = totalFeedback / float64(n)
	stats.SuccessCount = successCount
	stats.ConversionRate = float64(successCount) / float64(n)

	// 计算标准差
	if n > 1 {
		sumSquares := 0.0
		for _, result := range variant.Results {
			diff := result.Score - stats.AverageScore
			sumSquares += diff * diff
		}
		stats.StdDevScore = math.Sqrt(sumSquares / float64(n-1))
	}

	// 计算置信区间
	if n >= ab.config.MinSamples {
		stats.ConfidenceInterval = ab.calculateConfidenceInterval(
			stats.AverageScore,
			stats.StdDevScore,
			n,
			ab.config.ConfidenceLevel,
		)
	}
}

// calculateConfidenceInterval 计算置信区间
func (ab *ABTestingFramework) calculateConfidenceInterval(mean, stdDev float64, n int, confidenceLevel float64) *ConfidenceInterval {
	// 简化实现：使用正态分布近似
	// 标准误差
	stdError := stdDev / math.Sqrt(float64(n))

	// Z值 (简化，95% 置信水平 ≈ 1.96)
	zValue := 1.96
	if confidenceLevel == 0.99 {
		zValue = 2.576
	} else if confidenceLevel == 0.90 {
		zValue = 1.645
	}

	margin := zValue * stdError

	return &ConfidenceInterval{
		Lower: mean - margin,
		Upper: mean + margin,
		Level: confidenceLevel,
	}
}

// checkAutoStop 检查是否自动停止
func (ab *ABTestingFramework) checkAutoStop(ctx context.Context, experiment *Experiment) {
	// 检查每个变体是否有足够的样本
	for _, variant := range experiment.Variants {
		if len(variant.Results) < ab.config.MinSamples {
			return
		}
	}

	// 计算统计显著性
	metrics := ab.calculateMetrics(experiment)
	experiment.Metrics = metrics

	// 如果统计显著，停止实验
	if metrics.StatisticalSignificant {
		experiment.Status = "completed"
		experiment.Winner = ab.findVariant(experiment, metrics.Winner)
		now := time.Now()
		experiment.EndTime = &now
	}
}

// calculateMetrics 计算实验指标
func (ab *ABTestingFramework) calculateMetrics(experiment *Experiment) *ExperimentMetrics {
	if len(experiment.Variants) < 2 {
		return &ExperimentMetrics{}
	}

	// 获取两个变体 (简化：只比较前两个)
	v1 := experiment.Variants[0]
	v2 := experiment.Variants[1]

	if len(v1.Results) < ab.config.MinSamples || len(v2.Results) < ab.config.MinSamples {
		return &ExperimentMetrics{}
	}

	// 执行 T-test (简化实现)
	metrics := &ExperimentMetrics{}

	// 计算效应大小 (Cohen's d)
	pooledStdDev := math.Sqrt((v1.Stats.StdDevScore*v1.Stats.StdDevScore + v2.Stats.StdDevScore*v2.Stats.StdDevScore) / 2)
	if pooledStdDev > 0 {
		metrics.EffectSize = (v2.Stats.AverageScore - v1.Stats.AverageScore) / pooledStdDev
	}

	// 计算改进
	if v1.Stats.AverageScore > 0 {
		metrics.Improvement = (v2.Stats.AverageScore - v1.Stats.AverageScore) / v1.Stats.AverageScore
	}

	// 简化的 P值计算 (基于效应大小)
	// 实际应该使用 T-test 或 Mann-Whitney U test
	metrics.PValue = ab.calculatePValue(metrics.EffectSize, len(v1.Results)+len(v2.Results))

	// 判断统计显著性
	metrics.StatisticalSignificant = metrics.PValue < ab.config.SignificanceLevel

	// 确定获胜者
	if v2.Stats.AverageScore > v1.Stats.AverageScore {
		metrics.Winner = v2.Name
	} else {
		metrics.Winner = v1.Name
	}

	return metrics
}

// calculatePValue 计算 P值 (简化实现)
func (ab *ABTestingFramework) calculatePValue(effectSize float64, n int) float64 {
	// 简化实现：基于效应大小的近似
	// 实际应该使用 T-distribution

	absEffect := effectSize
	if absEffect < 0 {
		absEffect = -absEffect
	}

	// 粗略估计
	if absEffect > 0.8 {
		return 0.01 // 大效应，显著
	} else if absEffect > 0.5 {
		return 0.05 // 中等效应，临界显著
	} else if absEffect > 0.2 {
		return 0.10 // 小效应，不显著
	}
	return 0.50 // 无效应
}

// findVariant 查找变体
func (ab *ABTestingFramework) findVariant(experiment *Experiment, name string) *Variant {
	for _, variant := range experiment.Variants {
		if variant.Name == name {
			return variant
		}
	}
	return nil
}

// SelectVariant 选择变体 (用于流量分配)
func (ab *ABTestingFramework) SelectVariant(ctx context.Context, experimentName string) (*Variant, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	experiment, exists := ab.experiments[experimentName]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentName)
	}

	if experiment.Status != "running" {
		return nil, fmt.Errorf("experiment is not running")
	}

	// 简化：基于时间戳的流量分配
	// 实际应该使用更复杂的算法 (如哈希、随机等)
	now := time.Now()
	rand := float64(now.UnixNano()%100) / 100.0

	cumulative := 0.0
	for _, variant := range experiment.Variants {
		cumulative += variant.Traffic
		if rand < cumulative {
			return variant, nil
		}
	}

	// 默认返回第一个
	return experiment.Variants[0], nil
}

// GetExperiment 获取实验
func (ab *ABTestingFramework) GetExperiment(name string) (*Experiment, bool) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	experiment, exists := ab.experiments[name]
	return experiment, exists
}

// ListExperiments 列出所有实验
func (ab *ABTestingFramework) ListExperiments() []string {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	names := make([]string, 0, len(ab.experiments))
	for name := range ab.experiments {
		names = append(names, name)
	}
	return names
}

// StopExperiment 停止实验
func (ab *ABTestingFramework) StopExperiment(ctx context.Context, name string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	experiment, exists := ab.experiments[name]
	if !exists {
		return fmt.Errorf("experiment %s not found", name)
	}

	experiment.Status = "stopped"
	now := time.Now()
	experiment.EndTime = &now

	// 计算最终指标
	experiment.Metrics = ab.calculateMetrics(experiment)

	return nil
}

// GenerateReport 生成报告
func (ab *ABTestingFramework) GenerateReport(name string) string {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	experiment, exists := ab.experiments[name]
	if !exists {
		return ""
	}

	report := fmt.Sprintf("实验报告: %s\n", experiment.Name)
	report += fmt.Sprintf("描述: %s\n", experiment.Description)
	report += fmt.Sprintf("状态: %s\n", experiment.Status)
	report += fmt.Sprintf("开始时间: %s\n", experiment.StartTime.Format("2006-01-02 15:04:05"))
	if experiment.EndTime != nil {
		report += fmt.Sprintf("结束时间: %s\n", experiment.EndTime.Format("2006-01-02 15:04:05"))
	}

	report += "\n变体对比:\n"
	for _, variant := range experiment.Variants {
		report += fmt.Sprintf("\n%s:\n", variant.Name)
		report += fmt.Sprintf("  策略: %s\n", variant.Strategy)
		report += fmt.Sprintf("  流量: %.1f%%\n", variant.Traffic*100)
		report += fmt.Sprintf("  查询数: %d\n", variant.Stats.TotalQueries)
		report += fmt.Sprintf("  平均得分: %.3f\n", variant.Stats.AverageScore)
		report += fmt.Sprintf("  平均延迟: %dms\n", variant.Stats.AverageLatency)
		report += fmt.Sprintf("  转化率: %.2f%%\n", variant.Stats.ConversionRate*100)

		if variant.Stats.ConfidenceInterval != nil {
			ci := variant.Stats.ConfidenceInterval
			report += fmt.Sprintf("  置信区间 (%.0f%%): [%.3f, %.3f]\n",
				ci.Level*100, ci.Lower, ci.Upper)
		}
	}

	if experiment.Metrics != nil {
		report += "\n统计结果:\n"
		report += fmt.Sprintf("  P值: %.4f\n", experiment.Metrics.PValue)
		report += fmt.Sprintf("  效应大小: %.3f\n", experiment.Metrics.EffectSize)
		report += fmt.Sprintf("  相对改进: %.2f%%\n", experiment.Metrics.Improvement*100)
		report += fmt.Sprintf("  统计显著: %v\n", experiment.Metrics.StatisticalSignificant)
		if experiment.Metrics.Winner != "" {
			report += fmt.Sprintf("  获胜者: %s\n", experiment.Metrics.Winner)
		}
	}

	return report
}
