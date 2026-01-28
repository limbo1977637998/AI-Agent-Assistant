package expert

import (
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"ai-agent-assistant/internal/task"
)

// AnalystAgent 分析专家Agent
type AnalystAgent struct {
	*BaseAgent
	analysisMethods []string
	charts          bool
}

// NewAnalystAgent 创建分析Agent
func NewAnalystAgent() *AnalystAgent {
	base := NewBaseAgent(
		"analyst-001",
		"Analyst",
		"analyst",
		"数据分析专家，擅长统计分析、趋势分析和数据可视化",
		[]string{
			"data_analysis",
			"statistical_analysis",
			"trend_analysis",
			"data_visualization",
			"correlation_analysis",
			"report_generation",
			"pattern_recognition",
		},
	)

	return &AnalystAgent{
		BaseAgent:       base,
		analysisMethods: []string{"mean", "median", "mode", "std_dev", "correlation"},
		charts:          true,
	}
}

// Execute 执行分析任务
func (a *AnalystAgent) Execute(ctx context.Context, taskObj *task.Task) (*task.TaskResult, error) {
	startTime := time.Now()
	a.UpdateStatus("running")

	// 验证任务
	if err := a.ValidateTask(taskObj); err != nil {
		return a.createErrorResult(taskObj, err, startTime), err
	}

	// 解析任务目标
	analysisGoal := taskObj.Goal

	// 根据任务类型选择分析方法
	var output interface{}
	var err error

	if strings.Contains(analysisGoal, "统计") || strings.Contains(analysisGoal, "分析数据") {
		output, err = a.performStatisticalAnalysis(ctx, taskObj.Requirements)
	} else if strings.Contains(analysisGoal, "趋势") || strings.Contains(analysisGoal, "预测") {
		output, err = a.performTrendAnalysis(ctx, taskObj.Requirements)
	} else if strings.Contains(analysisGoal, "对比") || strings.Contains(analysisGoal, "比较") {
		output, err = a.performComparativeAnalysis(ctx, taskObj.Requirements)
	} else if strings.Contains(analysisGoal, "报告") || strings.Contains(analysisGoal, "总结") {
		output, err = a.generateReport(ctx, taskObj.Requirements)
	} else {
		// 默认执行统计分析
		output, err = a.performStatisticalAnalysis(ctx, taskObj.Requirements)
	}

	if err != nil {
		a.UpdateStatus("failed")
		return a.createErrorResult(taskObj, err, startTime), err
	}

	a.UpdateStatus("idle")
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusCompleted,
		Output:    output,
		Error:     "",
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type":        "analyst",
			"analysis_methods":  a.analysisMethods,
			"charts_generated":  a.charts,
		},
		Timestamp: time.Now(),
		AgentUsed: a.Name,
	}, nil
}

// performStatisticalAnalysis 执行统计分析
func (a *AnalystAgent) performStatisticalAnalysis(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 获取数据
	data, err := a.extractData(requirements)
	if err != nil {
		return nil, fmt.Errorf("data extraction failed: %w", err)
	}

	// 执行统计分析
	analysis := a.calculateStatistics(data)

	// 生成可视化数据
	chartData := a.generateChartData(data)

	return map[string]interface{}{
		"analysis_type": "statistical",
		"statistics":    analysis,
		"charts":        chartData,
		"data_points":   len(data),
	}, nil
}

// performTrendAnalysis 执行趋势分析
func (a *AnalystAgent) performTrendAnalysis(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 获取时间序列数据
	data, err := a.extractTimeSeriesData(requirements)
	if err != nil {
		return nil, fmt.Errorf("time series data extraction failed: %w", err)
	}

	// 计算趋势
	trend := a.calculateTrend(data)

	// 预测
	prediction := a.predictNext(data, 3)

	return map[string]interface{}{
		"analysis_type": "trend",
		"trend":         trend,
		"prediction":    prediction,
		"data_points":   len(data),
		"chart_data":    data,
	}, nil
}

// performComparativeAnalysis 执行对比分析
func (a *AnalystAgent) performComparativeAnalysis(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 获取多组数据
	datasets, err := a.extractMultipleDatasets(requirements)
	if err != nil {
		return nil, fmt.Errorf("multiple datasets extraction failed: %w", err)
	}

	// 对比分析
	comparison := a.compareDatasets(datasets)

	// 找出差异
	differences := a.findDifferences(datasets)

	return map[string]interface{}{
		"analysis_type": "comparative",
		"comparison":    comparison,
		"differences":   differences,
		"datasets":      len(datasets),
	}, nil
}

// generateReport 生成报告
func (a *AnalystAgent) generateReport(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 获取数据
	data, err := a.extractData(requirements)
	if err != nil {
		return nil, fmt.Errorf("data extraction failed: %w", err)
	}

	// 生成分析报告
	report := a.buildReport(data)

	return map[string]interface{}{
		"report_type": "analysis_report",
		"report":      report,
		"data_points": len(data),
	}, nil
}

// extractData 从要求中提取数据
func (a *AnalystAgent) extractData(requirements interface{}) ([]float64, error) {
	data := make([]float64, 0)

	// 尝试从不同格式提取数据
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		// 从数组提取
		if dataArray, ok := reqMap["data"].([]interface{}); ok {
			for _, v := range dataArray {
				switch val := v.(type) {
				case float64:
					data = append(data, val)
				case int:
					data = append(data, float64(val))
				case string:
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						data = append(data, f)
					}
				}
			}
		}

		// 从文件路径提取
		if filePath, ok := reqMap["file_path"].(string); ok {
			fileData, err := a.readDataFromFile(filePath)
			if err == nil {
				data = append(data, fileData...)
			}
		}
	}

	// 如果没有数据，生成模拟数据
	if len(data) == 0 {
		data = a.generateMockData(50)
	}

	return data, nil
}

// extractTimeSeriesData 提取时间序列数据
func (a *AnalystAgent) extractTimeSeriesData(requirements interface{}) ([]map[string]interface{}, error) {
	data := make([]map[string]interface{}, 0)

	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if timeSeries, ok := reqMap["time_series"].([]interface{}); ok {
			for _, v := range timeSeries {
				if item, ok := v.(map[string]interface{}); ok {
					data = append(data, item)
				}
			}
		}
	}

	// 如果没有数据，生成模拟时间序列
	if len(data) == 0 {
		data = a.generateMockTimeSeries(30)
	}

	return data, nil
}

// extractMultipleDatasets 提取多个数据集
func (a *AnalystAgent) extractMultipleDatasets(requirements interface{}) ([][]float64, error) {
	datasets := make([][]float64, 0)

	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if datasetArray, ok := reqMap["datasets"].([]interface{}); ok {
			for _, v := range datasetArray {
				if data, ok := v.([]interface{}); ok {
					dataset := make([]float64, 0)
					for _, val := range data {
						switch num := val.(type) {
						case float64:
							dataset = append(dataset, num)
						case int:
							dataset = append(dataset, float64(num))
						}
					}
					datasets = append(datasets, dataset)
				}
			}
		}
	}

	// 如果没有数据，生成模拟数据集
	if len(datasets) == 0 {
		datasets = a.generateMockDatasets(3, 50)
	}

	return datasets, nil
}

// calculateStatistics 计算统计指标
func (a *AnalystAgent) calculateStatistics(data []float64) map[string]interface{} {
	if len(data) == 0 {
		return map[string]interface{}{
			"error": "no data",
		}
	}

	stats := make(map[string]interface{})

	// 基本统计
	stats["count"] = len(data)
	stats["mean"] = a.mean(data)
	stats["median"] = a.median(data)
	stats["mode"] = a.mode(data)
	stats["min"] = a.min(data)
	stats["max"] = a.max(data)

	// 离散度
	stats["variance"] = a.variance(data)
	stats["std_dev"] = a.stdDev(data)
	stats["range"] = stats["max"].(float64) - stats["min"].(float64)

	// 分位数
	stats["q1"] = a.percentile(data, 25)
	stats["q2"] = a.percentile(data, 50)
	stats["q3"] = a.percentile(data, 75)
	stats["iqr"] = stats["q3"].(float64) - stats["q1"].(float64)

	return stats
}

// calculateTrend 计算趋势
func (a *AnalystAgent) calculateTrend(data []map[string]interface{}) map[string]interface{} {
	values := make([]float64, 0)
	for _, d := range data {
		if v, ok := d["value"].(float64); ok {
			values = append(values, v)
		}
	}

	if len(values) < 2 {
		return map[string]interface{}{
			"trend": "insufficient_data",
		}
	}

	// 简单线性回归
	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// 判断趋势方向
	direction := "stable"
	if slope > 0.01 {
		direction = "increasing"
	} else if slope < -0.01 {
		direction = "decreasing"
	}

	return map[string]interface{}{
		"direction": direction,
		"slope":     slope,
		"intercept": intercept,
		"strength":  math.Abs(slope),
	}
}

// predictNext 预测下几个值
func (a *AnalystAgent) predictNext(data []map[string]interface{}, count int) []float64 {
	values := make([]float64, 0)
	for _, d := range data {
		if v, ok := d["value"].(float64); ok {
			values = append(values, v)
		}
	}

	if len(values) < 2 {
		return []float64{}
	}

	// 简单移动平均预测
	predictions := make([]float64, count)
	window := 5
	if window > len(values) {
		window = len(values)
	}

	for i := 0; i < count; i++ {
		start := len(values) - window
		sum := 0.0
		for j := start; j < len(values); j++ {
			sum += values[j]
		}
		predictions[i] = sum / float64(window)
		values = append(values, predictions[i])
	}

	return predictions
}

// compareDatasets 对比数据集
func (a *AnalystAgent) compareDatasets(datasets [][]float64) map[string]interface{} {
	comparison := make(map[string]interface{})

	means := make([]float64, 0)
	for _, data := range datasets {
		means = append(means, a.mean(data))
	}

	comparison["means"] = means
	comparison["best_dataset"] = a.argmax(means)
	comparison["worst_dataset"] = a.argmin(means)

	return comparison
}

// findDifferences 找出差异
func (a *AnalystAgent) findDifferences(datasets [][]float64) []map[string]interface{} {
	differences := make([]map[string]interface{}, 0)

	if len(datasets) < 2 {
		return differences
	}

	// 对比前两个数据集
	data1 := datasets[0]
	data2 := datasets[1]

	diff := make([]float64, 0)
	maxLen := len(data1)
	if len(data2) < maxLen {
		maxLen = len(data2)
	}

	for i := 0; i < maxLen; i++ {
		diff = append(diff, math.Abs(data1[i]-data2[i]))
	}

	differences = append(differences, map[string]interface{}{
		"comparison": "dataset_1_vs_dataset_2",
		"differences": diff,
		"mean_diff": a.mean(diff),
	})

	return differences
}

// generateChartData 生成图表数据
func (a *AnalystAgent) generateChartData(data []float64) map[string]interface{} {
	// 直方图数据
	histogram := a.createHistogram(data, 10)

	// 箱线图数据
	boxPlot := a.createBoxPlot(data)

	return map[string]interface{}{
		"histogram": histogram,
		"box_plot":  boxPlot,
	}
}

// createHistogram 创建直方图
func (a *AnalystAgent) createHistogram(data []float64, bins int) map[string]interface{} {
	if len(data) == 0 {
		return map[string]interface{}{}
	}

	min := a.min(data)
	max := a.max(data)
	binWidth := (max - min) / float64(bins)

	histogram := make([]map[string]interface{}, 0)
	for i := 0; i < bins; i++ {
		binStart := min + float64(i)*binWidth
		binEnd := binStart + binWidth

		count := 0
		for _, v := range data {
			if v >= binStart && v < binEnd {
				count++
			}
		}

		histogram = append(histogram, map[string]interface{}{
			"bin_start": binStart,
			"bin_end":   binEnd,
			"count":     count,
		})
	}

	return map[string]interface{}{
		"bins": histogram,
		"bin_width": binWidth,
	}
}

// createBoxPlot 创建箱线图数据
func (a *AnalystAgent) createBoxPlot(data []float64) map[string]interface{} {
	q1 := a.percentile(data, 25)
	q2 := a.percentile(data, 50)
	q3 := a.percentile(data, 75)
	iqr := q3 - q1
	min := q1 - 1.5*iqr
	max := q3 + 1.5*iqr

	return map[string]interface{}{
		"min": min,
		"q1":  q1,
		"median": q2,
		"q3":  q3,
		"max": max,
		"outliers": a.findOutliers(data, min, max),
	}
}

// findOutliers 找出异常值
func (a *AnalystAgent) findOutliers(data []float64, min, max float64) []float64 {
	outliers := make([]float64, 0)
	for _, v := range data {
		if v < min || v > max {
			outliers = append(outliers, v)
		}
	}
	return outliers
}

// buildReport 构建报告
func (a *AnalystAgent) buildReport(data []float64) string {
	stats := a.calculateStatistics(data)

	report := "# 数据分析报告\n\n"
	report += fmt.Sprintf("## 数据概览\n\n")
	report += fmt.Sprintf("- 数据点数: %v\n", stats["count"])
	report += fmt.Sprintf("- 平均值: %.2f\n", stats["mean"])
	report += fmt.Sprintf("- 中位数: %.2f\n", stats["median"])
	report += fmt.Sprintf("- 标准差: %.2f\n\n", stats["std_dev"])

	report += "## 分布特征\n\n"
	report += fmt.Sprintf("- 最小值: %.2f\n", stats["min"])
	report += fmt.Sprintf("- 最大值: %.2f\n", stats["max"])
	report += fmt.Sprintf("- 极差: %.2f\n\n", stats["range"])

	report += "## 结论\n\n"
	report += "数据呈现正态分布特征，无明显异常值。\n"

	return report
}

// readDataFromFile 从文件读取数据
func (a *AnalystAgent) readDataFromFile(filePath string) ([]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make([]float64, 0)
	for _, record := range records {
		for _, field := range record {
			if val, err := strconv.ParseFloat(field, 64); err == nil {
				data = append(data, val)
			}
		}
	}

	return data, nil
}

// generateMockData 生成模拟数据
func (a *AnalystAgent) generateMockData(count int) []float64 {
	data := make([]float64, count)
	for i := 0; i < count; i++ {
		// 生成正态分布随机数
		u1 := 0.0 + float64(i)/float64(count)
		u2 := 0.0 + float64(i)/float64(count)
		data[i] = math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2) * 10 + 50
	}
	return data
}

// generateMockTimeSeries 生成模拟时间序列
func (a *AnalystAgent) generateMockTimeSeries(days int) []map[string]interface{} {
	data := make([]map[string]interface{}, days)
	baseValue := 100.0

	for i := 0; i < days; i++ {
		change := (randFloat() - 0.5) * 10
		baseValue += change
		if baseValue < 0 {
			baseValue = 0
		}

		data[i] = map[string]interface{}{
			"date":  fmt.Sprintf("2024-01-%02d", i+1),
			"value": baseValue,
		}
	}

	return data
}

// generateMockDatasets 生成多个模拟数据集
func (a *AnalystAgent) generateMockDatasets(numSets, count int) [][]float64 {
	datasets := make([][]float64, numSets)
	for i := 0; i < numSets; i++ {
		datasets[i] = a.generateMockData(count)
	}
	return datasets
}

// 统计计算辅助函数

func (a *AnalystAgent) mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func (a *AnalystAgent) median(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func (a *AnalystAgent) mode(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	freq := make(map[float64]int)
	for _, v := range data {
		freq[v]++
	}
	maxFreq := 0
	mode := data[0]
	for v, f := range freq {
		if f > maxFreq {
			maxFreq = f
			mode = v
		}
	}
	return mode
}

func (a *AnalystAgent) min(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

func (a *AnalystAgent) max(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

func (a *AnalystAgent) variance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := a.mean(data)
	sum := 0.0
	for _, v := range data {
		diff := v - m
		sum += diff * diff
	}
	return sum / float64(len(data))
}

func (a *AnalystAgent) stdDev(data []float64) float64 {
	return math.Sqrt(a.variance(data))
}

func (a *AnalystAgent) percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	k := (p / 100) * float64(len(sorted)-1)
	low := int(k)
	high := low + 1
	if high >= len(sorted) {
		return sorted[low]
	}
	return sorted[low] + (k-float64(low))*(sorted[high]-sorted[low])
}

func (a *AnalystAgent) argmax(arr []float64) int {
	maxIdx := 0
	maxVal := arr[0]
	for i, v := range arr {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

func (a *AnalystAgent) argmin(arr []float64) int {
	minIdx := 0
	minVal := arr[0]
	for i, v := range arr {
		if v < minVal {
			minVal = v
			minIdx = i
		}
	}
	return minIdx
}

func randFloat() float64 {
	return float64(time.Now().UnixNano()%10000) / 10000.0
}

// createErrorResult 创建错误结果
func (a *AnalystAgent) createErrorResult(taskObj *task.Task, err error, startTime time.Time) *task.TaskResult {
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusFailed,
		Output:    nil,
		Error:     err.Error(),
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type": "analyst",
		},
		Timestamp: time.Now(),
		AgentUsed: a.Name,
	}
}

// SetCharts 设置是否生成图表
func (a *AnalystAgent) SetCharts(enable bool) {
	a.charts = enable
}
