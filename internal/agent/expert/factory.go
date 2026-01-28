package expert

import (
	"context"
	"fmt"

	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	"ai-agent-assistant/internal/task"
	aitools "ai-agent-assistant/internal/tools"
)

// Factory 专家Agent工厂
type Factory struct {
	researcher *ResearcherAgent
	analyst    *AnalystAgent
	writer     *WriterAgent
	toolManager *aitools.ToolManager // 工具管理器
}

// NewFactory 创建工厂
func NewFactory() *Factory {
	return &Factory{
		researcher: NewResearcherAgent(),
		analyst:    NewAnalystAgent(),
		writer:     NewWriterAgent(),
		toolManager: nil, // 延迟初始化
	}
}

// SetToolManager 设置工具管理器
func (f *Factory) SetToolManager(toolManager *aitools.ToolManager) {
	f.toolManager = toolManager

	// 为所有 Agent 设置工具集成
	agents := []ExpertAgent{f.researcher, f.analyst, f.writer}
	for _, agent := range agents {
		if baseAgent, ok := agent.(*BaseAgent); ok {
			toolIntegration := aitools.NewAgentToolIntegration(baseAgent.ID, toolManager)
			baseAgent.SetToolIntegration(toolIntegration)
		}
	}
}

// GetToolManager 获取工具管理器
func (f *Factory) GetToolManager() *aitools.ToolManager {
	return f.toolManager
}

// CreateAgent 创建指定类型的Agent
func (f *Factory) CreateAgent(agentType string) (ExpertAgent, error) {
	switch agentType {
	case "researcher":
		return f.researcher, nil
	case "analyst":
		return f.analyst, nil
	case "writer":
		return f.writer, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
}

// GetAllAgents 获取所有Agent
func (f *Factory) GetAllAgents() map[string]ExpertAgent {
	return map[string]ExpertAgent{
		"researcher": f.researcher,
		"analyst":    f.analyst,
		"writer":     f.writer,
	}
}

// RegisterAllAgents 注册所有Agent到注册表
func (f *Factory) RegisterAllAgents(registry *aiagentorchestrator.AgentRegistry) error {
	agents := f.GetAllAgents()

	for _, agent := range agents {
		info := agent.GetInfo()
		if err := registry.Register(info); err != nil {
			return fmt.Errorf("failed to register agent: %w", err)
		}
	}

	return nil
}

// ExpertAgent 专家Agent接口
type ExpertAgent interface {
	Execute(ctx context.Context, taskObj *task.Task) (*task.TaskResult, error)
	GetInfo() *aiagentorchestrator.AgentInfo
	UpdateStatus(status string)
	GetCapabilities() []string
	HasCapability(capability string) bool
}

// GetAgentByCapability 根据能力获取Agent
func (f *Factory) GetAgentByCapability(capability string) (ExpertAgent, error) {
	agents := f.GetAllAgents()

	for _, agent := range agents {
		if agent.HasCapability(capability) {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no agent found with capability: %s", capability)
}

// GetAgentsByCapabilities 根据多个能力获取Agent
func (f *Factory) GetAgentsByCapabilities(capabilities []string) []ExpertAgent {
	matchedAgents := make([]ExpertAgent, 0)
	agents := f.GetAllAgents()

	for _, agent := range agents {
		matchesAll := true
		for _, cap := range capabilities {
			if !agent.HasCapability(cap) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			matchedAgents = append(matchedAgents, agent)
		}
	}

	return matchedAgents
}

// GetAgentInfo 获取所有Agent信息
func (f *Factory) GetAgentInfo() []map[string]interface{} {
	agents := f.GetAllAgents()
	info := make([]map[string]interface{}, 0)

	for _, agent := range agents {
		agentInfo := agent.GetInfo()
		info = append(info, map[string]interface{}{
			"name":         agentInfo.Name,
			"type":         agentInfo.Type,
			"capabilities": agentInfo.Capabilities,
			"status":       agentInfo.Status,
		})
	}

	return info
}
