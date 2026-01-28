package workflow

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Parser 工作流解析器
type Parser struct {
	workflowDir string // 工作流定义文件目录
}

// NewParser 创建解析器
func NewParser(workflowDir string) *Parser {
	return &Parser{
		workflowDir: workflowDir,
	}
}

// ParseFromString 从字符串解析工作流定义（支持JSON和YAML）
func (p *Parser) ParseFromString(data string, format string) (*Workflow, error) {
	var yamlDef WorkflowDefinitionYAML

	var err error
	if format == "yaml" || format == "" {
		err = yaml.Unmarshal([]byte(data), &yamlDef)
	} else if format == "json" {
		err = yaml.Unmarshal([]byte(data), &yamlDef) // YAML是JSON的超集
	} else {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	return p.convertFromYAML(&yamlDef)
}

// ParseFromYAML 从YAML结构解析
func (p *Parser) ParseFromYAML(yamlDef *WorkflowDefinitionYAML) (*Workflow, error) {
	return p.convertFromYAML(yamlDef)
}

// convertFromYAML 从YAML定义转换为工作流
func (p *Parser) convertFromYAML(yamlDef *WorkflowDefinitionYAML) (*Workflow, error) {
	workflow := NewWorkflow(yamlDef.Name, yamlDef.Description)
	workflow.Version = yamlDef.Version

	if workflow.Version == "" {
		workflow.Version = "1.0"
	}

	// 转换Agents
	for _, agentRef := range yamlDef.Agents {
		workflow.AddAgent(&AgentRef{
			Name:         agentRef.Name,
			Type:         agentRef.Type,
			Role:         agentRef.Role,
			Capabilities: agentRef.Capabilities,
		})
	}

	// 转换Variables
	for _, variable := range yamlDef.Variables {
		workflow.AddVariable(&Variable{
			Name:         variable.Name,
			Type:         variable.Type,
			DefaultValue: variable.DefaultValue,
			Required:     variable.Required,
			Description:  variable.Description,
		})
	}

	// 转换Steps
	for _, step := range yamlDef.Steps {
		workflowStep, err := p.convertStep(&step)
		if err != nil {
			return nil, fmt.Errorf("failed to convert step %s: %w", step.ID, err)
		}
		workflow.AddStep(workflowStep)
	}

	// 转换Config
	if yamlDef.Config != nil {
		workflow.Config = &WorkflowConfig{
			MaxRetries:        getIntValue(yamlDef.Config, "max_retries", 3),
			Timeout:           getDurationValue(yamlDef.Config, "timeout"),
			ParallelExecution: getBoolValue(yamlDef.Config, "parallel_execution", false),
			ContinueOnError:   getBoolValue(yamlDef.Config, "continue_on_error", false),
		}
	}

	// 转换Metadata
	if yamlDef.Metadata != nil {
		workflow.Metadata = yamlDef.Metadata
	}

	return workflow, nil
}

// convertStep 转换步骤
func (p *Parser) convertStep(yamlStep *YAMLStep) (*Step, error) {
	step := &Step{
		ID:          yamlStep.ID,
		Name:        yamlStep.Name,
		Description: yamlStep.Description,
		Type:        yamlStep.Type,
		Agent:       yamlStep.Agent,
		Tool:        yamlStep.Tool,
		DependsOn:   yamlStep.DependsOn,
		Config:      yamlStep.Config,
		Inputs:      yamlStep.Inputs,
		Outputs:     yamlStep.Outputs,
		Metadata:    yamlStep.Metadata,
	}

	// 设置默认类型
	if step.Type == "" {
		step.Type = "task"
	}

	// 转换Conditions
	if len(yamlStep.Conditions) > 0 {
		step.Conditions = make([]*Condition, len(yamlStep.Conditions))
		for i, cond := range yamlStep.Conditions {
			step.Conditions[i] = &Condition{
				Variable: cond.Variable,
				Operator: cond.Operator,
				Value:    cond.Value,
				Then:     cond.Then,
				Else:     cond.Else,
			}
		}
	}

	// 转换Retry
	if yamlStep.Retry != nil {
		step.Retry = &RetryConfig{
			MaxRetries: getIntValue(yamlStep.Retry, "max_retries", 3),
			Delay:      getDurationValue(yamlStep.Retry, "delay"),
			Backoff:    getFloatValue(yamlStep.Retry, "backoff", 2.0),
		}
	}

	// 转换Timeout
	if yamlStep.Timeout != "" {
		duration, err := time.ParseDuration(yamlStep.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		step.Timeout = duration
	}

	return step, nil
}

// ToYAML 将工作流转换为YAML
func (p *Parser) ToYAML(workflow *Workflow) (string, error) {
	yamlDef := &WorkflowDefinitionYAML{
		Name:        workflow.Name,
		Description: workflow.Description,
		Version:     workflow.Version,
		Agents:      make([]YAMLLAgentRef, 0),
		Variables:   make([]YAMLVariable, 0),
		Steps:       make([]YAMLStep, 0),
		Config:      make(map[string]interface{}),
		Metadata:    workflow.Metadata,
	}

	// 转换Agents
	for _, agent := range workflow.Agents {
		yamlDef.Agents = append(yamlDef.Agents, YAMLLAgentRef{
			Name:         agent.Name,
			Type:         agent.Type,
			Role:         agent.Role,
			Capabilities: agent.Capabilities,
		})
	}

	// 转换Variables
	for _, variable := range workflow.Variables {
		yamlDef.Variables = append(yamlDef.Variables, YAMLVariable{
			Name:         variable.Name,
			Type:         variable.Type,
			DefaultValue: variable.DefaultValue,
			Required:     variable.Required,
			Description:  variable.Description,
		})
	}

	// 转换Steps
	for _, step := range workflow.Steps {
		yamlDef.Steps = append(yamlDef.Steps, p.convertStepToYAML(step))
	}

	// 转换Config
	if workflow.Config != nil {
		if workflow.Config.MaxRetries > 0 {
			yamlDef.Config["max_retries"] = workflow.Config.MaxRetries
		}
		if workflow.Config.Timeout > 0 {
			yamlDef.Config["timeout"] = workflow.Config.Timeout.String()
		}
		yamlDef.Config["parallel_execution"] = workflow.Config.ParallelExecution
		yamlDef.Config["continue_on_error"] = workflow.Config.ContinueOnError
	}

	// 序列化为YAML
	data, err := yaml.Marshal(yamlDef)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow to YAML: %w", err)
	}

	return string(data), nil
}

// convertStepToYAML 转换步骤为YAML格式
func (p *Parser) convertStepToYAML(step *Step) YAMLStep {
	yamlStep := YAMLStep{
		ID:          step.ID,
		Name:        step.Name,
		Description: step.Description,
		Type:        step.Type,
		Agent:       step.Agent,
		Tool:        step.Tool,
		DependsOn:   step.DependsOn,
		Config:      step.Config,
		Inputs:      step.Inputs,
		Outputs:     step.Outputs,
		Metadata:    step.Metadata,
	}

	// 转换Conditions
	if len(step.Conditions) > 0 {
		yamlStep.Conditions = make([]YAMLCondition, len(step.Conditions))
		for i, cond := range step.Conditions {
			yamlStep.Conditions[i] = YAMLCondition{
				Variable: cond.Variable,
				Operator: cond.Operator,
				Value:    cond.Value,
				Then:     cond.Then,
				Else:     cond.Else,
			}
		}
	}

	// 转换Retry
	if step.Retry != nil {
		yamlStep.Retry = make(map[string]interface{})
		if step.Retry.MaxRetries > 0 {
			yamlStep.Retry["max_retries"] = step.Retry.MaxRetries
		}
		if step.Retry.Delay > 0 {
			yamlStep.Retry["delay"] = step.Retry.Delay.String()
		}
		if step.Retry.Backoff > 0 {
			yamlStep.Retry["backoff"] = step.Retry.Backoff
		}
	}

	// 转换Timeout
	if step.Timeout > 0 {
		yamlStep.Timeout = step.Timeout.String()
	}

	return yamlStep
}

// Helper functions for getting values from maps

func getIntValue(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// Try to parse string to int
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return defaultValue
}

func getFloatValue(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultValue
}

func getBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "yes"
		}
	}
	return defaultValue
}

func getDurationValue(m map[string]interface{}, key string) time.Duration {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case time.Duration:
			return v
		case string:
			d, err := time.ParseDuration(v)
			if err == nil {
				return d
			}
		case int:
			return time.Duration(v) * time.Second
		case float64:
			return time.Duration(v) * time.Second
		}
	}
	return 0
}
