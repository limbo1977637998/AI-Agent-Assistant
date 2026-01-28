package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	aiagenttask "ai-agent-assistant/internal/task"
)

// Monitor 工作流监控器
// 负责收集工作流执行指标、跟踪状态、记录性能数据
type Monitor struct {
	mu                sync.RWMutex
	executions        map[string]*WorkflowExecutionMetrics // 工作流执行记录
	agentMetrics      map[string]*AgentMetrics            // Agent性能指标
	enabled           bool                                 // 是否启用监控
	metricsRetention  time.Duration                        // 指标保留时间
	eventChannel      chan *MonitorEvent                   // 事件通道
	eventBufferSize   int                                  // 事件缓冲区大小
	collectInterval   time.Duration                        // 指标收集间隔
	stopChan          chan struct{}                        // 停止信号
	listeners         []MonitorListener                    // 监听器列表
}

// WorkflowExecutionMetrics 工作流执行指标
type WorkflowExecutionMetrics struct {
	ExecutionID       string                    `json:"execution_id"`        // 执行ID
	WorkflowID        string                    `json:"workflow_id"`         // 工作流ID
	Status            string                    `json:"status"`              // 状态
	StartTime         time.Time                 `json:"start_time"`          // 开始时间
	EndTime           time.Time                 `json:"end_time,omitempty"`  // 结束时间
	Duration          time.Duration             `json:"duration"`            // 执行时长
	StepMetrics       map[string]*StepMetrics   `json:"step_metrics"`        // 步骤指标
	AgentUsage        map[string]int            `json:"agent_usage"`         // Agent使用统计
	ErrorCount        int                       `json:"error_count"`         // 错误数量
	WarningCount      int                       `json:"warning_count"`       // 警告数量
	ResourceUsage     *ResourceUsage            `json:"resource_usage"`      // 资源使用情况
	CustomMetrics     map[string]interface{}    `json:"custom_metrics"`      // 自定义指标
}

// StepMetrics 步骤执行指标
type StepMetrics struct {
	StepID            string                 `json:"step_id"`            // 步骤ID
	Agent             string                 `json:"agent"`              // 使用的Agent
	StartTime         time.Time              `json:"start_time"`         // 开始时间
	EndTime           time.Time              `json:"end_time"`           // 结束时间
	Duration          time.Duration          `json:"duration"`           // 执行时长
	Status            string                 `json:"status"`             // 状态
	InputSize         int                    `json:"input_size"`         // 输入大小
	OutputSize        int                    `json:"output_size"`        // 输出大小
	RetryCount        int                    `json:"retry_count"`        // 重试次数
	Error             error                  `json:"error,omitempty"`    // 错误信息
	PerformanceScore  float64                `json:"performance_score"`  // 性能评分
}

// AgentMetrics Agent性能指标
type AgentMetrics struct {
	AgentID           string                 `json:"agent_id"`            // Agent ID
	TotalExecutions   int                    `json:"total_executions"`    // 总执行次数
	SuccessExecutions int                    `json:"success_executions"`  // 成功执行次数
	FailedExecutions  int                    `json:"failed_executions"`   // 失败执行次数
	AverageDuration   time.Duration          `json:"average_duration"`    // 平均执行时长
	MinDuration       time.Duration          `json:"min_duration"`        // 最小执行时长
	MaxDuration       time.Duration          `json:"max_duration"`        // 最大执行时长
	TotalDataProcessed int64                 `json:"total_data_processed"` // 总处理数据量
	LastUsed          time.Time              `json:"last_used"`           // 最后使用时间
	CapabilityUsage   map[string]int         `json:"capability_usage"`    // 能力使用统计
	PerformanceScore  float64                `json:"performance_score"`   // 性能评分
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsage       int64         `json:"memory_usage"`        // 内存使用（字节）
	CPUUsage          float64       `json:"cpu_usage"`           // CPU使用率
	NetworkIO         int64         `json:"network_io"`          // 网络IO（字节）
	DiskIO            int64         `json:"disk_io"`             // 磁盘IO（字节）
	ConcurrentTasks   int           `json:"concurrent_tasks"`    // 并发任务数
}

// MonitorEvent 监控事件
type MonitorEvent struct {
	Type        string                 `json:"type"`                  // 事件类型
	Timestamp   time.Time              `json:"timestamp"`             // 时间戳
	ExecutionID string                 `json:"execution_id"`          // 执行ID
	WorkflowID  string                 `json:"workflow_id"`           // 工作流ID
	StepID      string                 `json:"step_id,omitempty"`     // 步骤ID
	Agent       string                 `json:"agent,omitempty"`       // Agent
	Data        map[string]interface{} `json:"data"`                  // 事件数据
}

// MonitorListener 监听器接口
type MonitorListener interface {
	// OnEvent 处理监控事件
	OnEvent(event *MonitorEvent) error
	// OnMetricsUpdate 指标更新通知
	OnMetricsUpdate(metrics *WorkflowExecutionMetrics) error
}

// NewMonitor 创建新的监控器
func NewMonitor() *Monitor {
	return &Monitor{
		executions:       make(map[string]*WorkflowExecutionMetrics),
		agentMetrics:     make(map[string]*AgentMetrics),
		enabled:          true,
		metricsRetention: 24 * time.Hour, // 默认保留24小时
		eventChannel:     make(chan *MonitorEvent, 1000),
		eventBufferSize:  1000,
		collectInterval:  1 * time.Minute,
		stopChan:         make(chan struct{}),
		listeners:        make([]MonitorListener, 0),
	}
}

// Start 启动监控器
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	m.enabled = true
	m.mu.Unlock()

	// 启动事件处理goroutine
	go m.processEvents(ctx)

	// 启动指标收集goroutine
	go m.collectMetricsPeriodically(ctx)

	// 启动指标清理goroutine
	go m.cleanupOldMetrics(ctx)

	return nil
}

// Stop 停止监控器
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enabled = false
	close(m.stopChan)
	close(m.eventChannel)
}

// RecordWorkflowStart 记录工作流开始
func (m *Monitor) RecordWorkflowStart(executionID, workflowID string) *WorkflowExecutionMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return nil
	}

	metrics := &WorkflowExecutionMetrics{
		ExecutionID:    executionID,
		WorkflowID:     workflowID,
		Status:         "running",
		StartTime:      time.Now(),
		StepMetrics:    make(map[string]*StepMetrics),
		AgentUsage:     make(map[string]int),
		CustomMetrics:  make(map[string]interface{}),
		ResourceUsage:  &ResourceUsage{},
	}

	m.executions[executionID] = metrics

	// 发送事件
	m.publishEvent(&MonitorEvent{
		Type:        "workflow_started",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		Data: map[string]interface{}{
			"start_time": metrics.StartTime,
		},
	})

	return metrics
}

// RecordWorkflowEnd 记录工作流结束
func (m *Monitor) RecordWorkflowEnd(executionID string, status string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	metrics.Status = status
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	if err != nil {
		metrics.ErrorCount++
	}

	// 发送事件
	eventData := map[string]interface{}{
		"end_time":  metrics.EndTime,
		"duration":  metrics.Duration,
		"status":    status,
	}

	if err != nil {
		eventData["error"] = err.Error()
	}

	m.publishEvent(&MonitorEvent{
		Type:        "workflow_completed",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  metrics.WorkflowID,
		Data:        eventData,
	})

	// 通知监听器
	m.notifyListenersMetricsUpdate(metrics)
}

// RecordStepStart 记录步骤开始
func (m *Monitor) RecordStepStart(executionID, stepID, agent string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	stepMetrics := &StepMetrics{
		StepID:    stepID,
		Agent:     agent,
		StartTime: time.Now(),
		Status:    "running",
	}

	metrics.StepMetrics[stepID] = stepMetrics

	// 更新Agent使用统计
	metrics.AgentUsage[agent]++

	// 发送事件
	m.publishEvent(&MonitorEvent{
		Type:        "step_started",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  metrics.WorkflowID,
		StepID:      stepID,
		Agent:       agent,
		Data: map[string]interface{}{
			"start_time": stepMetrics.StartTime,
		},
	})
}

// RecordStepEnd 记录步骤结束
func (m *Monitor) RecordStepEnd(executionID, stepID string, status string, result *aiagenttask.TaskResult, inputSize, outputSize int, retries int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	stepMetrics, exists := metrics.StepMetrics[stepID]
	if !exists {
		return
	}

	stepMetrics.EndTime = time.Now()
	stepMetrics.Duration = stepMetrics.EndTime.Sub(stepMetrics.StartTime)
	stepMetrics.Status = status
	stepMetrics.InputSize = inputSize
	stepMetrics.OutputSize = outputSize
	stepMetrics.RetryCount = retries

	if result != nil && result.Error != "" {
		// 将 string 类型的 error 转换为 error 类型
		stepMetrics.Error = fmt.Errorf(result.Error)
	}

	// 计算性能评分（基于执行时间和成功率）
	stepMetrics.PerformanceScore = m.calculateStepPerformanceScore(stepMetrics)

	// 更新Agent指标
	m.updateAgentMetrics(stepMetrics.Agent, stepMetrics)

	// 发送事件
	eventData := map[string]interface{}{
		"end_time":          stepMetrics.EndTime,
		"duration":          stepMetrics.Duration,
		"status":            status,
		"input_size":        inputSize,
		"output_size":       outputSize,
		"retry_count":       retries,
		"performance_score": stepMetrics.PerformanceScore,
	}

	if result != nil && result.Error != "" {
		eventData["error"] = result.Error
	}

	m.publishEvent(&MonitorEvent{
		Type:        "step_completed",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  metrics.WorkflowID,
		StepID:      stepID,
		Agent:       stepMetrics.Agent,
		Data:        eventData,
	})
}

// RecordError 记录错误
func (m *Monitor) RecordError(executionID, stepID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	metrics.ErrorCount++

	// 发送事件
	m.publishEvent(&MonitorEvent{
		Type:        "error",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  metrics.WorkflowID,
		StepID:      stepID,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

// RecordWarning 记录警告
func (m *Monitor) RecordWarning(executionID, stepID string, warning string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	metrics.WarningCount++

	// 发送事件
	m.publishEvent(&MonitorEvent{
		Type:        "warning",
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		WorkflowID:  metrics.WorkflowID,
		StepID:      stepID,
		Data: map[string]interface{}{
			"warning": warning,
		},
	})
}

// UpdateResourceUsage 更新资源使用情况
func (m *Monitor) UpdateResourceUsage(executionID string, usage *ResourceUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	metrics.ResourceUsage = usage
}

// AddCustomMetric 添加自定义指标
func (m *Monitor) AddCustomMetric(executionID string, key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return
	}

	metrics, exists := m.executions[executionID]
	if !exists {
		return
	}

	metrics.CustomMetrics[key] = value
}

// GetExecutionMetrics 获取工作流执行指标
func (m *Monitor) GetExecutionMetrics(executionID string) (*WorkflowExecutionMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("执行指标不存在: %s", executionID)
	}

	return metrics, nil
}

// GetAllExecutions 获取所有执行记录
func (m *Monitor) GetAllExecutions() []*WorkflowExecutionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executions := make([]*WorkflowExecutionMetrics, 0, len(m.executions))
	for _, metrics := range m.executions {
		executions = append(executions, metrics)
	}

	return executions
}

// GetAgentMetrics 获取Agent性能指标
func (m *Monitor) GetAgentMetrics(agentID string) (*AgentMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.agentMetrics[agentID]
	if !exists {
		return nil, fmt.Errorf("Agent指标不存在: %s", agentID)
	}

	return metrics, nil
}

// GetAllAgentMetrics 获取所有Agent指标
func (m *Monitor) GetAllAgentMetrics() map[string]*AgentMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	result := make(map[string]*AgentMetrics, len(m.agentMetrics))
	for k, v := range m.agentMetrics {
		result[k] = v
	}

	return result
}

// GetPerformanceReport 获取性能报告
func (m *Monitor) GetPerformanceReport(workflowID string) *PerformanceReport {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := &PerformanceReport{
		WorkflowID:    workflowID,
		GeneratedAt:   time.Now(),
		Executions:    make([]*WorkflowExecutionMetrics, 0),
		AgentSummary:  make(map[string]*AgentMetrics),
	}

	// 收集工作流执行记录
	for _, metrics := range m.executions {
		if metrics.WorkflowID == workflowID {
			report.Executions = append(report.Executions, metrics)
		}
	}

	// 计算统计信息
	if len(report.Executions) > 0 {
		var totalDuration time.Duration
		successCount := 0

		for _, exec := range report.Executions {
			totalDuration += exec.Duration
			if exec.Status == "completed" {
				successCount++
			}
		}

		report.TotalExecutions = len(report.Executions)
		report.SuccessCount = successCount
		report.AverageDuration = totalDuration / time.Duration(len(report.Executions))
		report.SuccessRate = float64(successCount) / float64(len(report.Executions)) * 100
	}

	// 收集Agent摘要
	for agentID, metrics := range m.agentMetrics {
		report.AgentSummary[agentID] = metrics
	}

	return report
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	WorkflowID       string                      `json:"workflow_id"`         // 工作流ID
	GeneratedAt      time.Time                   `json:"generated_at"`        // 生成时间
	TotalExecutions  int                         `json:"total_executions"`    // 总执行次数
	SuccessCount     int                         `json:"success_count"`       // 成功次数
	AverageDuration  time.Duration               `json:"average_duration"`    // 平均执行时长
	SuccessRate      float64                     `json:"success_rate"`        // 成功率
	Executions       []*WorkflowExecutionMetrics `json:"executions"`          // 执行记录
	AgentSummary     map[string]*AgentMetrics    `json:"agent_summary"`       // Agent摘要
	Recommendations  []string                    `json:"recommendations"`      // 优化建议
}

// calculateStepPerformanceScore 计算步骤性能评分
func (m *Monitor) calculateStepPerformanceScore(step *StepMetrics) float64 {
	// 基础评分100分
	score := 100.0

	// 根据执行时间扣分（超过1秒每秒扣1分）
	if step.Duration.Seconds() > 1 {
		score -= (step.Duration.Seconds() - 1) * 1
	}

	// 根据重试次数扣分（每次重试扣5分）
	score -= float64(step.RetryCount) * 5

	// 根据状态扣分
	if step.Status == "failed" {
		score -= 50
	} else if step.Status == "cancelled" {
		score -= 30
	}

	// 确保评分不低于0
	if score < 0 {
		score = 0
	}

	return score
}

// updateAgentMetrics 更新Agent指标
func (m *Monitor) updateAgentMetrics(agentID string, step *StepMetrics) {
	agentMetrics, exists := m.agentMetrics[agentID]
	if !exists {
		agentMetrics = &AgentMetrics{
			AgentID:          agentID,
			CapabilityUsage:  make(map[string]int),
			MinDuration:      step.Duration,
			MaxDuration:      step.Duration,
		}
		m.agentMetrics[agentID] = agentMetrics
	}

	agentMetrics.TotalExecutions++
	agentMetrics.LastUsed = time.Now()

	if step.Status == "completed" {
		agentMetrics.SuccessExecutions++
	} else {
		agentMetrics.FailedExecutions++
	}

	// 更新执行时长统计
	if step.Duration < agentMetrics.MinDuration {
		agentMetrics.MinDuration = step.Duration
	}
	if step.Duration > agentMetrics.MaxDuration {
		agentMetrics.MaxDuration = step.Duration
	}

	// 计算平均时长
	if agentMetrics.TotalExecutions > 0 {
		totalDuration := agentMetrics.AverageDuration * time.Duration(agentMetrics.TotalExecutions-1)
		agentMetrics.AverageDuration = (totalDuration + step.Duration) / time.Duration(agentMetrics.TotalExecutions)
	}

	// 计算性能评分
	agentMetrics.PerformanceScore = m.calculateAgentPerformanceScore(agentMetrics)
}

// calculateAgentPerformanceScore 计算Agent性能评分
func (m *Monitor) calculateAgentPerformanceScore(metrics *AgentMetrics) float64 {
	if metrics.TotalExecutions == 0 {
		return 0
	}

	// 成功率权重50%
	successRate := float64(metrics.SuccessExecutions) / float64(metrics.TotalExecutions)
	score := successRate * 50

	// 平均执行时长权重30%（基于5秒为基准）
	if metrics.AverageDuration.Seconds() > 0 {
		durationScore := 30.0 / (1 + metrics.AverageDuration.Seconds()/5.0)
		score += durationScore
	} else {
		score += 30
	}

	// 执行次数权重20%（最多20分）
	executionScore := float64(metrics.TotalExecutions) * 2
	if executionScore > 20 {
		executionScore = 20
	}
	score += executionScore

	return score
}

// AddListener 添加监听器
func (m *Monitor) AddListener(listener MonitorListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, listener)
}

// RemoveListener 移除监听器
func (m *Monitor) RemoveListener(listener MonitorListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, l := range m.listeners {
		if l == listener {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			break
		}
	}
}

// publishEvent 发布事件
func (m *Monitor) publishEvent(event *MonitorEvent) {
	select {
	case m.eventChannel <- event:
	default:
		// 事件通道已满，丢弃事件
		fmt.Printf("警告: 监控事件通道已满，丢弃事件: %s\n", event.Type)
	}
}

// processEvents 处理事件
func (m *Monitor) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case event := <-m.eventChannel:
			m.handleEvent(event)
		}
	}
}

// handleEvent 处理单个事件
func (m *Monitor) handleEvent(event *MonitorEvent) {
	// 通知所有监听器
	for _, listener := range m.listeners {
		if err := listener.OnEvent(event); err != nil {
			fmt.Printf("监听器处理事件失败: %v\n", err)
		}
	}
}

// notifyListenersMetricsUpdate 通知监听器指标更新
func (m *Monitor) notifyListenersMetricsUpdate(metrics *WorkflowExecutionMetrics) {
	for _, listener := range m.listeners {
		if err := listener.OnMetricsUpdate(metrics); err != nil {
			fmt.Printf("监听器处理指标更新失败: %v\n", err)
		}
	}
}

// collectMetricsPeriodically 定期收集指标
func (m *Monitor) collectMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(m.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics 收集系统指标
func (m *Monitor) collectSystemMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 这里可以收集系统级别的指标
	// 例如：内存使用、CPU使用率等
	// 简化实现，实际可以使用 gopsutil 等库
}

// cleanupOldMetrics 清理旧指标
func (m *Monitor) cleanupOldMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.removeExpiredMetrics()
		}
	}
}

// removeExpiredMetrics 移除过期指标
func (m *Monitor) removeExpiredMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, metrics := range m.executions {
		if now.Sub(metrics.StartTime) > m.metricsRetention {
			delete(m.executions, id)
		}
	}
}

// IsEnabled 检查监控是否启用
func (m *Monitor) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.enabled
}

// SetEnabled 设置监控状态
func (m *Monitor) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enabled = enabled
}

// GetStats 获取监控统计信息
func (m *Monitor) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"enabled":           m.enabled,
		"executions_count":  len(m.executions),
		"agents_count":      len(m.agentMetrics),
		"event_buffer_size": len(m.eventChannel),
		"listeners_count":   len(m.listeners),
	}
}
