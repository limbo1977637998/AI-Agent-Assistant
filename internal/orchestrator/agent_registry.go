package orchestrator

import (
	"fmt"
	"sync"
	"time"
)

// AgentInfo Agent信息
type AgentInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`         // expert, general, custom
	Capabilities []string          `json:"capabilities"` // Agent能力列表
	Endpoint     string            `json:"endpoint"`     // Agent访问端点
	Status       string            `json:"status"`       // active, inactive, busy
	Metadata     map[string]string `json:"metadata"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	CreatedAt    time.Time         `json:"created_at"`
}

// AgentRegistry Agent注册表
type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*AgentInfo // key: agent_name
}

// NewAgentRegistry 创建Agent注册表
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentInfo),
	}
}

// Register 注册Agent
func (r *AgentRegistry) Register(agent *AgentInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agent.Name]; exists {
		return fmt.Errorf("agent %s already registered", agent.Name)
	}

	agent.CreatedAt = time.Now()
	agent.LastHeartbeat = time.Now()
	agent.Status = "active"

	r.agents[agent.Name] = agent
	return nil
}

// Unregister 注销Agent
func (r *AgentRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	delete(r.agents, name)
	return nil
}

// Get 获取Agent信息
func (r *AgentRegistry) Get(name string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", name)
	}

	return agent, nil
}

// List 列出所有Agent
func (r *AgentRegistry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// ListByType 按类型列出Agent
func (r *AgentRegistry) ListByType(agentType string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0)
	for _, agent := range r.agents {
		if agent.Type == agentType {
			agents = append(agents, agent)
		}
	}
	return agents
}

// ListByCapability 按能力列出Agent
func (r *AgentRegistry) ListByCapability(capability string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0)
	for _, agent := range r.agents {
		for _, cap := range agent.Capabilities {
			if cap == capability {
				agents = append(agents, agent)
				break
			}
		}
	}
	return agents
}

// UpdateHeartbeat 更新心跳
func (r *AgentRegistry) UpdateHeartbeat(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	agent.LastHeartbeat = time.Now()
	return nil
}

// UpdateStatus 更新Agent状态
func (r *AgentRegistry) UpdateStatus(name, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	agent.Status = status
	return nil
}

// GetActiveAgents 获取所有活跃的Agent
func (r *AgentRegistry) GetActiveAgents() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0)
	for _, agent := range r.agents {
		if agent.Status == "active" {
			agents = append(agents, agent)
		}
	}
	return agents
}

// CheckHealth 检查Agent健康状态
func (r *AgentRegistry) CheckHealth(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return false
	}

	// 如果超过30秒没有心跳，认为不健康
	return time.Since(agent.LastHeartbeat) < 30*time.Second
}

// GetIdleAgent 获取一个空闲的Agent
func (r *AgentRegistry) GetIdleAgent() (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, agent := range r.agents {
		if agent.Status == "active" {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no idle agent available")
}

// FindBestAgent 根据能力找到最匹配的Agent
func (r *AgentRegistry) FindBestAgent(requiredCapabilities []string) (*AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var bestAgent *AgentInfo
	maxMatch := 0

	for _, agent := range r.agents {
		if agent.Status != "active" {
			continue
		}

		matchCount := 0
		for _, reqCap := range requiredCapabilities {
			for _, agentCap := range agent.Capabilities {
				if agentCap == reqCap {
					matchCount++
					break
				}
			}
		}

		if matchCount > maxMatch {
			maxMatch = matchCount
			bestAgent = agent
		}
	}

	if bestAgent == nil {
		return nil, fmt.Errorf("no agent found with required capabilities")
	}

	return bestAgent, nil
}

// Count 统计Agent数量
func (r *AgentRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// CountByStatus 按状态统计Agent数量
func (r *AgentRegistry) CountByStatus(status string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, agent := range r.agents {
		if agent.Status == status {
			count++
		}
	}
	return count
}
