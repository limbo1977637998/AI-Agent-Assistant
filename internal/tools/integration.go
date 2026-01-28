package tools

import (
	"context"
	"fmt"
	"sync"
)

// AgentToolIntegration Agent工具集成
// 为Agent提供工具调用能力
type AgentToolIntegration struct {
	toolManager *ToolManager
	agentID     string
	mu          sync.RWMutex
}

// NewAgentToolIntegration 创建Agent工具集成实例
func NewAgentToolIntegration(agentID string, toolManager *ToolManager) *AgentToolIntegration {
	return &AgentToolIntegration{
		toolManager: toolManager,
		agentID:     agentID,
	}
}

// CallTool 调用工具
// 参数：
//   - toolName: 工具名称
//   - operation: 操作类型
//   - params: 操作参数
func (ati *AgentToolIntegration) CallTool(ctx context.Context, toolName, operation string, params map[string]interface{}) (interface{}, error) {
	// 添加Agent ID到上下文
	if params == nil {
		params = make(map[string]interface{})
	}
	params["agent_id"] = ati.agentID

	// 调用工具管理器执行工具
	return ati.toolManager.ExecuteTool(ctx, toolName, operation, params)
}

// GetAvailableTools 获取可用工具列表
func (ati *AgentToolIntegration) GetAvailableTools() []map[string]interface{} {
	return ati.toolManager.GetAvailableTools()
}

// GetToolCapabilities 获取工具能力
func (ati *AgentToolIntegration) GetToolCapabilities(toolName string) (map[string]interface{}, error) {
	return ati.toolManager.GetToolCapabilities(toolName)
}

// HasTool 检查工具是否可用
func (ati *AgentToolIntegration) HasTool(toolName string) bool {
	return ati.toolManager.GetRegistry().HasTool(toolName)
}

// BatchCallTools 批量调用工具
func (ati *AgentToolIntegration) BatchCallTools(ctx context.Context, calls []ToolCall) ([]ToolResult, error) {
	results := make([]ToolResult, len(calls))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, call := range calls {
		wg.Add(1)
		go func(index int, c ToolCall) {
			defer wg.Done()

			data, err := ati.CallTool(ctx, c.ToolName, c.Operation, c.Params)

			mu.Lock()
			results[index] = ToolResult{
				Success: err == nil,
				Data:    data,
			}
			if err != nil {
				results[index].Error = err.Error()
				results[index].Message = "工具调用失败"
			} else {
				results[index].Message = "工具调用成功"
			}
			mu.Unlock()
		}(i, call)
	}

	wg.Wait()
	return results, nil
}

// ToolCall 工具调用定义
type ToolCall struct {
	ToolName  string                 `json:"tool_name"`             // 工具名称
	Operation string                 `json:"operation"`             // 操作类型
	Params    map[string]interface{} `json:"params"`                // 操作参数
}

// ToolChain 工具链
// 支持多个工具串联执行
type ToolChain struct {
	name      string
	steps     []ChainStep
	toolMgr   *ToolManager
}

// ChainStep 链步骤
type ChainStep struct {
	ToolName  string                 `json:"tool_name"`             // 工具名称
	Operation string                 `json:"operation"`             // 操作类型
	Params    map[string]interface{} `json:"params"`                // 参数
	InputFrom string                 `json:"input_from,omitempty"`  // 输入来源（上一步骤的输出）
}

// NewToolChain 创建工具链
func NewToolChain(name string, toolMgr *ToolManager) *ToolChain {
	return &ToolChain{
		name:    name,
		steps:   make([]ChainStep, 0),
		toolMgr: toolMgr,
	}
}

// AddStep 添加步骤
func (tc *ToolChain) AddStep(toolName, operation string, params map[string]interface{}, inputFrom string) *ToolChain {
	step := ChainStep{
		ToolName:  toolName,
		Operation: operation,
		Params:    params,
		InputFrom: inputFrom,
	}
	tc.steps = append(tc.steps, step)
	return tc
}

// Execute 执行工具链
func (tc *ToolChain) Execute(ctx context.Context, initialInput interface{}) (interface{}, error) {
	var currentOutput interface{} = initialInput
	var lastErr error

	for i, step := range tc.steps {
		// 准备参数
		params := make(map[string]interface{})
		for k, v := range step.Params {
			params[k] = v
		}

		// 如果指定了输入来源，使用上一步的输出
		if step.InputFrom != "" && currentOutput != nil {
			params["input"] = currentOutput
		}

		// 执行工具
		result, err := tc.toolMgr.ExecuteTool(ctx, step.ToolName, step.Operation, params)
		if err != nil {
			lastErr = fmt.Errorf("步骤 %d (%s.%s) 执行失败: %w", i, step.ToolName, step.Operation, err)
			break
		}

		currentOutput = result
	}

	return currentOutput, lastErr
}

// GetSteps 获取步骤列表
func (tc *ToolChain) GetSteps() []ChainStep {
	return tc.steps
}

// GetName 获取工具链名称
func (tc *ToolChain) GetName() string {
	return tc.name
}

// ToolChainExecutor 工具链执行器
// 提供更高级的工具执行功能
type ToolChainExecutor struct {
	toolMgr      *ToolManager
	chainCache   map[string]*ToolChain
	mu           sync.RWMutex
	retryPolicy  *RetryPolicy
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`     // 最大重试次数
	RetryDelay    int           `json:"retry_delay"`     // 重试延迟（毫秒）
	BackoffFactor float64       `json:"backoff_factor"`  // 退避因子
}

// DefaultRetryPolicy 默认重试策略
var DefaultRetryPolicy = &RetryPolicy{
	MaxRetries:    3,
	RetryDelay:    1000,
	BackoffFactor: 2.0,
}

// NewToolChainExecutor 创建工具链执行器
func NewToolChainExecutor(toolMgr *ToolManager) *ToolChainExecutor {
	return &ToolChainExecutor{
		toolMgr:     toolMgr,
		chainCache:  make(map[string]*ToolChain),
		retryPolicy: DefaultRetryPolicy,
	}
}

// ExecuteWithRetry 带重试的工具执行
func (te *ToolChainExecutor) ExecuteWithRetry(ctx context.Context, toolName, operation string, params map[string]interface{}) (interface{}, error) {
	var lastErr error
	delay := te.retryPolicy.RetryDelay

	for attempt := 0; attempt <= te.retryPolicy.MaxRetries; attempt++ {
		result, err := te.toolMgr.ExecuteTool(ctx, toolName, operation, params)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 如果还有重试机会，等待
		if attempt < te.retryPolicy.MaxRetries {
			// TODO: 添加延迟等待
			_ = delay
			delay = int(float64(delay) * te.retryPolicy.BackoffFactor)
		}
	}

	return nil, fmt.Errorf("工具执行失败，已重试 %d 次: %w", te.retryPolicy.MaxRetries, lastErr)
}

// RegisterChain 注册工具链
func (te *ToolChainExecutor) RegisterChain(chain *ToolChain) {
	te.mu.Lock()
	defer te.mu.Unlock()

	te.chainCache[chain.GetName()] = chain
}

// GetChain 获取工具链
func (te *ToolChainExecutor) GetChain(name string) (*ToolChain, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	chain, exists := te.chainCache[name]
	if !exists {
		return nil, fmt.Errorf("工具链不存在: %s", name)
	}

	return chain, nil
}

// ExecuteChain 执行工具链
func (te *ToolChainExecutor) ExecuteChain(ctx context.Context, chainName string, initialInput interface{}) (interface{}, error) {
	chain, err := te.GetChain(chainName)
	if err != nil {
		return nil, err
	}

	return chain.Execute(ctx, initialInput)
}

// CreateToolChains 创建预定义的工具链
func CreateToolChains(toolMgr *ToolManager) map[string]*ToolChain {
	chains := make(map[string]*ToolChain)

	// 数据处理链：读取CSV -> 清洗数据 -> 转换格式 -> 保存
	dataProcessingChain := NewToolChain("data_processing", toolMgr).
		AddStep("file_ops", "read", map[string]interface{}{
			"path": "input.csv",
		}, "").
		AddStep("data_processor", "clean", map[string]interface{}{
			"operations": []string{"remove_empty", "trim_whitespace"},
		}, "input").
		AddStep("data_processor", "transform", map[string]interface{}{
			"transformations": []map[string]interface{}{
				{"field": "age", "operation": "round", "value": 0},
			},
		}, "input").
		AddStep("file_ops", "write", map[string]interface{}{
			"path":     "output.json",
			"overwrite": true,
		}, "input")

	chains["data_processing"] = dataProcessingChain

	// 批量下载处理链：批量下载 -> 解压 -> 处理数据
	batchDownloadChain := NewToolChain("batch_download_process", toolMgr).
		AddStep("batch_ops", "batch_http", map[string]interface{}{
			"requests": []interface{}{},
		}, "").
		AddStep("file_ops", "decompress", map[string]interface{}{
			"source":      "download.zip",
			"destination": "./extracted",
		}, "input").
		AddStep("data_processor", "parse_csv", map[string]interface{}{
			"content": "",
		}, "input")

	chains["batch_download_process"] = batchDownloadChain

	// 数据分析链：读取数据 -> 过滤 -> 聚合 -> 生成报告
	analysisChain := NewToolChain("data_analysis", toolMgr).
		AddStep("file_ops", "read", map[string]interface{}{
			"path": "data.json",
		}, "").
		AddStep("data_processor", "filter", map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"field": "status", "operator": "==", "value": "active"},
			},
		}, "input").
		AddStep("data_processor", "aggregate", map[string]interface{}{
			"group_by": "category",
			"aggregations": []map[string]interface{}{
				{"field": "amount", "operation": "sum"},
				{"field": "count", "operation": "count"},
			},
		}, "input").
		AddStep("file_ops", "write", map[string]interface{}{
			"path":     "report.json",
			"overwrite": true,
		}, "input")

	chains["data_analysis"] = analysisChain

	return chains
}

// ToolMetrics 工具使用指标
type ToolMetrics struct {
	ToolName         string                 `json:"tool_name"`             // 工具名称
	TotalCalls       int                    `json:"total_calls"`           // 总调用次数
	SuccessCalls     int                    `json:"success_calls"`         // 成功调用次数
	FailedCalls      int                    `json:"failed_calls"`          // 失败调用次数
	AverageDuration  int64                  `json:"average_duration"`      // 平均执行时长（毫秒）
	LastUsed         int64                  `json:"last_used"`             // 最后使用时间（Unix时间戳）
	PopularOperations map[string]int        `json:"popular_operations"`    // 热门操作统计
	ErrorSummary     map[string]int         `json:"error_summary"`         // 错误汇总
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*ToolMetrics
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*ToolMetrics),
	}
}

// RecordCall 记录工具调用
func (mc *MetricsCollector) RecordCall(toolName, operation string, duration int64, success bool, errorMsg string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metric, exists := mc.metrics[toolName]
	if !exists {
		metric = &ToolMetrics{
			ToolName:         toolName,
			PopularOperations: make(map[string]int),
			ErrorSummary:     make(map[string]int),
		}
		mc.metrics[toolName] = metric
	}

	metric.TotalCalls++
	metric.LastUsed = duration // 这里传入的是时间戳，需要修正

	if success {
		metric.SuccessCalls++
	} else {
		metric.FailedCalls++
		if errorMsg != "" {
			metric.ErrorSummary[errorMsg]++
		}
	}

	// 更新操作统计
	if operation != "" {
		metric.PopularOperations[operation]++
	}

	// 更新平均时长
	if metric.TotalCalls > 0 {
		totalDuration := metric.AverageDuration * int64(metric.TotalCalls-1)
		metric.AverageDuration = (totalDuration + duration) / int64(metric.TotalCalls)
	}
}

// GetMetrics 获取工具指标
func (mc *MetricsCollector) GetMetrics(toolName string) (*ToolMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metric, exists := mc.metrics[toolName]
	if !exists {
		return nil, fmt.Errorf("工具指标不存在: %s", toolName)
	}

	return metric, nil
}

// GetAllMetrics 获取所有指标
func (mc *MetricsCollector) GetAllMetrics() map[string]*ToolMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// 返回副本
	result := make(map[string]*ToolMetrics, len(mc.metrics))
	for k, v := range mc.metrics {
		result[k] = v
	}

	return result
}

// Reset 重置指标
func (mc *MetricsCollector) Reset(toolName string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if toolName == "" {
		// 重置所有
		mc.metrics = make(map[string]*ToolMetrics)
	} else {
		delete(mc.metrics, toolName)
	}
}
