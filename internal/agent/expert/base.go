package expert

import (
	"context"
	"fmt"
	"time"

	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	"ai-agent-assistant/internal/task"
	aitools "ai-agent-assistant/internal/tools"
)

// BaseAgent 专家Agent基类
type BaseAgent struct {
	ID           string
	Name         string
	Type         string
	Description  string
	Capabilities []string
	Config       map[string]interface{}
	Status       string
	StartTime    time.Time
	ToolIntegration *aitools.AgentToolIntegration // 工具集成
}

// NewBaseAgent 创建基础Agent
func NewBaseAgent(id, name, agentType, description string, capabilities []string) *BaseAgent {
	return &BaseAgent{
		ID:           id,
		Name:         name,
		Type:         agentType,
		Description:  description,
		Capabilities: capabilities,
		Config:       make(map[string]interface{}),
		Status:       "idle",
		StartTime:    time.Now(),
	}
}

// Execute 执行任务（由子类实现具体逻辑）
func (a *BaseAgent) Execute(ctx context.Context, taskObj *task.Task) (*task.TaskResult, error) {
	// 更新状态
	a.Status = "running"

	// 子类实现具体逻辑
	result := &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusPending,
		Output:    nil,
		Duration:  0,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		AgentUsed: a.Name,
	}

	return result, nil
}

// GetInfo 获取Agent信息
func (a *BaseAgent) GetInfo() *aiagentorchestrator.AgentInfo {
	metadata := make(map[string]string)
	metadata["start_time"] = a.StartTime.Format(time.RFC3339)

	return &aiagentorchestrator.AgentInfo{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		Capabilities:  a.Capabilities,
		Status:        a.Status,
		LastHeartbeat: time.Now(),
		Metadata:      metadata,
	}
}

// UpdateStatus 更新状态
func (a *BaseAgent) UpdateStatus(status string) {
	a.Status = status
}

// SetConfig 设置配置
func (a *BaseAgent) SetConfig(key string, value interface{}) {
	if a.Config == nil {
		a.Config = make(map[string]interface{})
	}
	a.Config[key] = value
}

// GetConfig 获取配置
func (a *BaseAgent) GetConfig(key string) (interface{}, bool) {
	if a.Config == nil {
		return nil, false
	}
	value, exists := a.Config[key]
	return value, exists
}

// ValidateTask 验证任务是否适合此Agent
func (a *BaseAgent) ValidateTask(task *task.Task) error {
	// 检查任务类型是否匹配
	if task.Type != "" && task.Type != a.Type && task.Type != "single" {
		return fmt.Errorf("task type %s doesn't match agent type %s", task.Type, a.Type)
	}

	// 检查任务需求是否在Agent能力范围内
	if len(task.RequiredCapabilities) > 0 {
		for _, cap := range task.RequiredCapabilities {
			if !a.HasCapability(cap) {
				return fmt.Errorf("agent missing required capability: %s", cap)
			}
		}
	}

	return nil
}

// HasCapability 检查是否具有某能力
func (a *BaseAgent) HasCapability(capability string) bool {
	for _, cap := range a.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetCapabilities 获取所有能力
func (a *BaseAgent) GetCapabilities() []string {
	return a.Capabilities
}

// GetStatus 获取状态
func (a *BaseAgent) GetStatus() string {
	return a.Status
}

// GetID 获取ID
func (a *BaseAgent) GetID() string {
	return a.ID
}

// GetName 获取名称
func (a *BaseAgent) GetName() string {
	return a.Name
}

// GetType 获取类型
func (a *BaseAgent) GetType() string {
	return a.Type
}

// GetDescription 获取描述
func (a *BaseAgent) GetDescription() string {
	return a.Description
}

// SetToolIntegration 设置工具集成
func (a *BaseAgent) SetToolIntegration(toolIntegration *aitools.AgentToolIntegration) {
	a.ToolIntegration = toolIntegration
}

// GetToolIntegration 获取工具集成
func (a *BaseAgent) GetToolIntegration() *aitools.AgentToolIntegration {
	return a.ToolIntegration
}

// HasTool 检查是否有指定工具
func (a *BaseAgent) HasTool(toolName string) bool {
	if a.ToolIntegration == nil {
		return false
	}
	return a.ToolIntegration.HasTool(toolName)
}

// CallTool 调用工具
// 参数：
//   - ctx: 上下文
//   - toolName: 工具名称
//   - operation: 操作类型
//   - params: 操作参数
// 返回：
//   - result: 工具执行结果
//   - err: 错误信息
func (a *BaseAgent) CallTool(ctx context.Context, toolName, operation string, params map[string]interface{}) (interface{}, error) {
	if a.ToolIntegration == nil {
		return nil, fmt.Errorf("工具集成未初始化")
	}

	return a.ToolIntegration.CallTool(ctx, toolName, operation, params)
}

// GetAvailableTools 获取可用工具列表
func (a *BaseAgent) GetAvailableTools() []map[string]interface{} {
	if a.ToolIntegration == nil {
		return []map[string]interface{}{}
	}

	return a.ToolIntegration.GetAvailableTools()
}

// GetToolCapabilities 获取工具能力
func (a *BaseAgent) GetToolCapabilities(toolName string) (map[string]interface{}, error) {
	if a.ToolIntegration == nil {
		return nil, fmt.Errorf("工具集成未初始化")
	}

	return a.ToolIntegration.GetToolCapabilities(toolName)
}

// BatchCallTools 批量调用工具
func (a *BaseAgent) BatchCallTools(ctx context.Context, calls []aitools.ToolCall) ([]aitools.ToolResult, error) {
	if a.ToolIntegration == nil {
		return nil, fmt.Errorf("工具集成未初始化")
	}

	return a.ToolIntegration.BatchCallTools(ctx, calls)
}
