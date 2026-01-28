package workflow

import (
	"fmt"
	"time"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusPaused    WorkflowStatus = "paused"
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// Workflow 工作流定义
type Workflow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Steps       []*Step      `json:"steps"`
	Agents      []*AgentRef  `json:"agents,omitempty"`
	Variables   []*Variable  `json:"variables,omitempty"`
	Config      *WorkflowConfig `json:"config,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	MaxRetries      int           `json:"max_retries,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	ParallelExecution bool        `json:"parallel_execution,omitempty"`
	ContinueOnError bool          `json:"continue_on_error,omitempty"`
}

// AgentRef Agent引用
type AgentRef struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
}

// Variable 变量定义
type Variable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, number, boolean, object, array
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Description  string      `json:"description"`
}

// Step 工作流步骤
type Step struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // task, condition, parallel, sequential
	Agent       string            `json:"agent,omitempty"`    // 使用的Agent
	Tool        string            `json:"tool,omitempty"`     // 使用的工具
	DependsOn   []string          `json:"depends_on,omitempty"` // 依赖的步骤ID
	Config      map[string]interface{} `json:"config,omitempty"`
	Inputs      map[string]string `json:"inputs,omitempty"`   // 输入映射
	Outputs     map[string]string `json:"outputs,omitempty"`  // 输出映射
	Conditions  []*Condition      `json:"conditions,omitempty"` // 条件判断
	Retry       *RetryConfig      `json:"retry,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Condition 条件判断
type Condition struct {
	Variable string      `json:"variable"`      // 变量名
	Operator string      `json:"operator"`      // eq, ne, gt, lt, gte, lte, in, not_in, contains
	Value    interface{} `json:"value"`         // 比较值
	Then     string      `json:"then"`          // 满足条件时执行的步骤ID
	Else     string      `json:"else,omitempty"` // 不满足条件时执行的步骤ID
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int           `json:"max_retries"`
	Delay      time.Duration `json:"delay"`
	Backoff    float64       `json:"backoff"` // 指数退避系数
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	ID            string                   `json:"id"`
	WorkflowID    string                   `json:"workflow_id"`
	WorkflowName  string                   `json:"workflow_name"`
	Workflow      *Workflow                `json:"-"` // 执行的工作流定义（不序列化）
	Status        WorkflowStatus           `json:"status"`
	Inputs        map[string]interface{}  `json:"inputs"`
	Outputs       map[string]interface{}  `json:"outputs"`
	StepStates    map[string]*StepState   `json:"step_states"` // step_id -> state
	Error         string                   `json:"error,omitempty"`
	StartedAt     time.Time                `json:"started_at"`
	CompletedAt   *time.Time               `json:"completed_at,omitempty"`
	Duration      time.Duration            `json:"duration"`
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`
}

// StepState 步骤执行状态
type StepState struct {
	StepID       string       `json:"step_id"`
	Status       StepStatus   `json:"status"`
	Input        interface{}  `json:"input,omitempty"`
	Output       interface{}  `json:"output,omitempty"`
	Error        string       `json:"error,omitempty"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
	Duration     time.Duration `json:"duration"`
	RetryCount   int          `json:"retry_count"`
	AgentUsed    string       `json:"agent_used,omitempty"`
	Logs         []string     `json:"logs,omitempty"`
}

// WorkflowDefinitionYAML YAML格式的定义
type WorkflowDefinitionYAML struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version,omitempty"`
	Agents      []YAMLLAgentRef         `yaml:"agents,omitempty"`
	Variables   []YAMLVariable         `yaml:"variables,omitempty"`
	Steps       []YAMLStep             `yaml:"steps"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
	Metadata    map[string]string      `yaml:"metadata,omitempty"`
}

// YAMLLAgentRef YAML格式的Agent引用
type YAMLLAgentRef struct {
	Name         string   `yaml:"name"`
	Type         string   `yaml:"type"`
	Role         string   `yaml:"role"`
	Capabilities []string `yaml:"capabilities,omitempty"`
}

// YAMLVariable YAML格式的变量
type YAMLVariable struct {
	Name         string      `yaml:"name"`
	Type         string      `yaml:"type"`
	DefaultValue interface{} `yaml:"default_value,omitempty"`
	Required     bool        `yaml:"required"`
	Description  string      `yaml:"description,omitempty"`
}

// YAMLStep YAML格式的步骤
type YAMLStep struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Type        string                 `yaml:"type,omitempty"`
	Agent       string                 `yaml:"agent,omitempty"`
	Tool        string                 `yaml:"tool,omitempty"`
	DependsOn   []string               `yaml:"depends_on,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
	Inputs      map[string]string      `yaml:"inputs,omitempty"`
	Outputs     map[string]string      `yaml:"outputs,omitempty"`
	Conditions  []YAMLCondition        `yaml:"conditions,omitempty"`
	Retry       map[string]interface{} `yaml:"retry,omitempty"`
	Timeout     string                 `yaml:"timeout,omitempty"` // duration string
	Metadata    map[string]string      `yaml:"metadata,omitempty"`
}

// YAMLCondition YAML格式的条件
type YAMLCondition struct {
	Variable string      `yaml:"variable"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
	Then     string      `yaml:"then"`
	Else     string      `yaml:"else,omitempty"`
}

// Helper functions

// NewWorkflow 创建新工作流
func NewWorkflow(name, description string) *Workflow {
	return &Workflow{
		ID:          generateID("workflow"),
		Name:        name,
		Description: description,
		Version:     "1.0",
		Steps:       make([]*Step, 0),
		Agents:      make([]*AgentRef, 0),
		Variables:   make([]*Variable, 0),
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddStep 添加步骤
func (w *Workflow) AddStep(step *Step) {
	w.Steps = append(w.Steps, step)
	w.UpdatedAt = time.Now()
}

// AddAgent 添加Agent
func (w *Workflow) AddAgent(agent *AgentRef) {
	w.Agents = append(w.Agents, agent)
}

// AddVariable 添加变量
func (w *Workflow) AddVariable(variable *Variable) {
	w.Variables = append(w.Variables, variable)
}

// NewWorkflowExecution 创建工作流执行实例
func NewWorkflowExecution(workflow *Workflow, inputs map[string]interface{}) *WorkflowExecution {
	return &WorkflowExecution{
		ID:           generateID("exec"),
		WorkflowID:   workflow.ID,
		WorkflowName: workflow.Name,
		Workflow:     workflow, // 保存工作流定义引用
		Status:       WorkflowStatusPending,
		Inputs:       inputs,
		Outputs:      make(map[string]interface{}),
		StepStates:   make(map[string]*StepState),
		StartedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
}

// GetStep 获取步骤
func (w *Workflow) GetStep(stepID string) *Step {
	for _, step := range w.Steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

// GetStepState 获取步骤状态
func (e *WorkflowExecution) GetStepState(stepID string) *StepState {
	return e.StepStates[stepID]
}

// SetStepState 设置步骤状态
func (e *WorkflowExecution) SetStepState(stepID string, state *StepState) {
	e.StepStates[stepID] = state
}

// IsCompleted 是否完成
func (e *WorkflowExecution) IsCompleted() bool {
	return e.Status == WorkflowStatusCompleted || e.Status == WorkflowStatusFailed || e.Status == WorkflowStatusCancelled
}

// MarkCompleted 标记为完成
func (e *WorkflowExecution) MarkCompleted() {
	now := time.Now()
	e.CompletedAt = &now
	e.Duration = now.Sub(e.StartedAt)
	e.Status = WorkflowStatusCompleted
}

// MarkFailed 标记为失败
func (e *WorkflowExecution) MarkFailed(err error) {
	now := time.Now()
	e.CompletedAt = &now
	e.Duration = now.Sub(e.StartedAt)
	e.Status = WorkflowStatusFailed
	if err != nil {
		e.Error = err.Error()
	}
}

// generateID 生成ID
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
