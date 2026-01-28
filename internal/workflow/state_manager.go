package workflow

import (
	"fmt"
	"sync"
	"time"
)

// StateManager 工作流状态管理器
type StateManager struct {
	mu               sync.RWMutex
	executions       map[string]*WorkflowExecution // execution_id -> execution
	workflows        map[string]*Workflow         // workflow_id -> workflow
	checkpointData    map[string][]byte            // checkpoint data for recovery
}

// NewStateManager 创建状态管理器
func NewStateManager() *StateManager {
	return &StateManager{
		executions:    make(map[string]*WorkflowExecution),
		workflows:     make(map[string]*Workflow),
		checkpointData: make(map[string][]byte),
	}
}

// SetExecution 设置工作流执行
func (m *StateManager) SetExecution(executionID string, execution *WorkflowExecution) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.executions[executionID]; exists {
		return fmt.Errorf("execution %s already exists", executionID)
	}

	m.executions[executionID] = execution
	return nil
}

// UpdateExecution 更新工作流执行状态
func (m *StateManager) UpdateExecution(executionID string, execution *WorkflowExecution) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.executions[executionID]; !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	m.executions[executionID] = execution
	return nil
}

// GetExecution 获取工作流执行
func (m *StateManager) GetExecution(executionID string) (*WorkflowExecution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	return execution, nil
}

// GetAllExecutions 获取所有执行
func (m *StateManager) GetAllExecutions() []*WorkflowExecution {
	m.mu.Lock()
	defer m.mu.Unlock()

	executions := make([]*WorkflowExecution, 0, len(m.executions))
	for _, exec := range m.executions {
		executions = append(executions, exec)
	}

	return executions
}

// GetExecutionsByStatus 按状态获取执行
func (m *StateManager) GetExecutionsByStatus(status WorkflowStatus) []*WorkflowExecution {
	m.mu.Lock()
	defer m.mu.Unlock()

	executions := make([]*WorkflowExecution, 0)
	for _, exec := range m.executions {
		if exec.Status == status {
			executions = append(executions, exec)
		}
	}

	return executions
}

// SetWorkflow 设置工作流定义
func (m *StateManager) SetWorkflow(workflow *Workflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow %s already exists", workflow.ID)
	}

	m.workflows[workflow.ID] = workflow
	return nil
}

// GetWorkflow 获取工作流
func (m *StateManager) GetWorkflow(workflowID string) (*Workflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workflow, exists := m.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return workflow, nil
}

// GetWorkflows 获取所有工作流
func (m *StateManager) GetWorkflows() []*Workflow {
	m.mu.Lock()
	defer m.mu.Unlock()

	workflows := make([]*Workflow, 0, len(m.workflows))
	for _, w := range m.workflows {
		workflows = append(workflows, w)
	}

	return workflows
}

// CreateCheckpoint 创建检查点
func (m *StateManager) CreateCheckpoint(executionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	execution, exists := m.executions[executionID]
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	// 序列化执行状态
	data, err := SerializeExecution(execution)
	if err != nil {
		return fmt.Errorf("failed to serialize execution: %w", err)
	}

	m.checkpointData[executionID] = data
	return nil
}

// LoadFromCheckpoint 从检查点恢复
func (m *StateManager) LoadFromCheckpoint(executionID string) (*WorkflowExecution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, exists := m.checkpointData[executionID]
	if !exists {
		return nil, fmt.Errorf("no checkpoint found for execution %s", executionID)
	}

	execution, err := DeserializeExecution(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize execution: %w", err)
	}

	// 恢复到状态管理器
	m.executions[executionID] = execution

	return execution, nil
}

// DeleteExecution 删除执行记录
func (m *StateManager) DeleteExecution(executionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.executions[executionID]; !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	delete(m.executions, executionID)
	delete(m.checkpointData, executionID)

	return nil
}

// DeleteWorkflow 删除工作流定义
func (m *StateManager) DeleteWorkflow(workflowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workflows[workflowID]; !exists {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	delete(m.workflows, workflowID)
	return nil
}

// GetStatistics 获取统计信息
func (m *StateManager) GetStatistics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 统计执行状态
	statusCounts := make(map[WorkflowStatus]int)
	for _, exec := range m.executions {
		statusCounts[exec.Status]++
	}

	return map[string]interface{}{
		"total_executions": len(m.executions),
		"total_workflows":   len(m.workflows),
		"status_counts":     statusCounts,
		"checkpoint_count":   len(m.checkpointData),
	}
}

// CleanupOldExecutions 清理旧的执行记录
func (m *StateManager) CleanupOldExecutions(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-olderThan)

	for executionID, execution := range m.executions {
		if execution.CompletedAt != nil && execution.CompletedAt.Before(cutoff) {
			delete(m.executions, executionID)
			delete(m.checkpointData, executionID)
			count++
		}
	}

	return count
}

// SerializeExecution 序列化执行状态
func SerializeExecution(execution *WorkflowExecution) ([]byte, error) {
	// 简化实现：使用JSON序列化
	// 实际应用中可以使用更高效的序列化方式

	// 这里返回一个简单的表示
	data := []byte(fmt.Sprintf("# WorkflowExecution\nID: %s\nWorkflowID: %s\nStatus: %s\nStartedAt: %v",
		execution.ID, execution.WorkflowID, execution.Status, execution.StartedAt))

	return data, nil
}

// DeserializeExecution 反序列化执行状态
func DeserializeExecution(data []byte) (*WorkflowExecution, error) {
	// 简化实现
	return nil, fmt.Errorf("deserialization not implemented")
}
