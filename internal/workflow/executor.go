package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	"ai-agent-assistant/internal/task"
)

// Executor 工作流执行器
type Executor struct {
	registry       *aiagentorchestrator.AgentRegistry
	scheduler      *aiagentorchestrator.TaskScheduler
	lifecycleMgr   *task.LifecycleManager
	decomposer     task.Decomposer
	aggregator     task.Aggregator
	stateMgr       *StateManager
}

// NewExecutor 创建执行器
func NewExecutor(
	registry *aiagentorchestrator.AgentRegistry,
	scheduler *aiagentorchestrator.TaskScheduler,
) *Executor {
	return &Executor{
		registry:     registry,
		scheduler:    scheduler,
		lifecycleMgr: task.NewLifecycleManager(),
		decomposer:   task.NewTemplateDecomposer(),
		aggregator:   task.NewSimpleAggregator(),
		stateMgr:     NewStateManager(),
	}
}

// Execute 执行工作流
func (e *Executor) Execute(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (*WorkflowExecution, error) {
	// 创建执行实例
	execution := NewWorkflowExecution(workflow, inputs)

	// 初始化状态
	e.stateMgr.SetExecution(execution.ID, execution)

	// 更新执行状态
	execution.Status = WorkflowStatusRunning

	// 构建DAG
	dag, err := BuildDAGFromWorkflow(workflow)
	if err != nil {
		execution.MarkFailed(fmt.Errorf("failed to build DAG: %w", err))
		return execution, err
	}

	// 获取执行层级
	levels := dag.GetLevels()

	// 逐层执行
	for levelIndex, levelSteps := range levels {
		fmt.Printf("  📍 执行第%d层，共%d个步骤\n", levelIndex+1, len(levelSteps))

		// 执行这一层的所有步骤
		results := e.executeLevel(ctx, execution, dag, levelSteps)

		// 检查是否有步骤失败
		for _, result := range results {
			if !result.Success {
				// 如果配置了continue_on_error，继续执行
				if execution.Workflow.Config != nil && execution.Workflow.Config.ContinueOnError {
					fmt.Printf("  ⚠️  步骤 %s 失败，但继续执行\n", result.StepID)
				} else {
					execution.MarkFailed(fmt.Errorf("step %s failed", result.StepID))
					return execution, fmt.Errorf("workflow execution failed at step %s", result.StepID)
				}
			}
		}
	}

	// 标记完成
	execution.MarkCompleted()
	e.stateMgr.UpdateExecution(execution.ID, execution)

	return execution, nil
}

// executeLevel 执行某一层的步骤
func (e *Executor) executeLevel(ctx context.Context, execution *WorkflowExecution, dag *DAG, stepIDs []string) []*StepResult {
	results := make([]*StepResult, len(stepIDs))

	// 如果配置了并行执行，使用并发
	if execution.Workflow != nil && execution.Workflow.Config != nil && execution.Workflow.Config.ParallelExecution {
		results = e.executeParallel(ctx, execution, stepIDs)
	} else {
		results = e.executeSequential(ctx, execution, stepIDs)
	}

	return results
}

// executeSequential 顺序执行步骤
func (e *Executor) executeSequential(ctx context.Context, execution *WorkflowExecution, stepIDs []string) []*StepResult {
	results := make([]*StepResult, len(stepIDs))

	for i, stepID := range stepIDs {
		step := execution.Workflow.GetStep(stepID)
		if step == nil {
			results[i] = &StepResult{
				StepID:  stepID,
				Success: false,
				Error:   "step not found",
			}
			continue
		}

		fmt.Printf("    → 执行步骤: %s (%s)\n", stepID, step.Name)
		results[i] = e.executeStep(ctx, execution, step)
	}

	return results
}

// executeParallel 并行执行步骤
func (e *Executor) executeParallel(ctx context.Context, execution *WorkflowExecution, stepIDs []string) []*StepResult {
	var wg sync.WaitGroup
	results := make([]*StepResult, len(stepIDs))
	resultChan := make(chan *StepResult, len(stepIDs))

	for i, stepID := range stepIDs {
		wg.Add(1)
		go func(index int, stepID string) {
			defer wg.Done()

			step := execution.Workflow.GetStep(stepID)
			if step == nil {
				resultChan <- &StepResult{
					StepID:  stepID,
					Success: false,
					Error:   "step not found",
				}
				return
			}

			fmt.Printf("    → 并行执行步骤: %s (%s)\n", stepID, step.Name)
			resultChan <- e.executeStep(ctx, execution, step)
		}(i, stepID)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	i := 0
	for result := range resultChan {
		results[i] = result
		i++
	}

	return results
}

// executeStep 执行单个步骤
func (e *Executor) executeStep(ctx context.Context, execution *WorkflowExecution, step *Step) *StepResult {
	result := &StepResult{
		StepID:  step.ID,
		Success: true,
	}

	// 创建步骤状态
	now := time.Now()
	stepState := &task.TaskState{
		TaskID:    step.ID,
		Status:    task.TaskStatusPending,
		Stage:     "pending",
		Input:     execution.Inputs,
		StartedAt: &now,
		Metadata:  make(map[string]interface{}),
	}

	// 创建一个临时的task.Task用于生命周期管理
	tempTask := &task.Task{
		ID:         step.ID,
		Type:       step.Type,
		Goal:       step.Name,
		Status:     task.TaskStatusPending,
		Priority:   task.PriorityNormal,
		Requirements: execution.Inputs,
		CreatedAt: now,
	}
	e.lifecycleMgr.Create(tempTask)

	// 更新为运行中
	e.lifecycleMgr.UpdateStatus(step.ID, task.TaskStatusRunning, "step execution started")
	stepState.Status = task.TaskStatusRunning
	stepState.Stage = "executing"

	// 根据步骤类型执行
	var output interface{}
	var err error

	switch step.Type {
	case "task":
		output, err = e.executeTaskStep(ctx, execution, step)
	case "condition":
		output, err = e.executeConditionStep(ctx, execution, step)
	case "parallel":
		output, err = e.executeParallelStep(ctx, execution, step)
	case "sequential":
		output, err = e.executeSequentialStep(ctx, execution, step)
	default:
		output, err = e.executeTaskStep(ctx, execution, step)
	}

	// 更新结果
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		e.lifecycleMgr.SetError(step.ID, err)
		e.lifecycleMgr.UpdateStatus(step.ID, task.TaskStatusFailed, "execution failed")
	} else {
		result.Output = output
		e.lifecycleMgr.SetOutput(step.ID, output)
		e.lifecycleMgr.UpdateStatus(step.ID, task.TaskStatusCompleted, "execution completed")
	}

	// 保存到执行状态
	status := StepStatusCompleted
	if !result.Success {
		status = StepStatusFailed
	}

	duration := time.Duration(0)
	if stepState.StartedAt != nil {
		duration = time.Since(*stepState.StartedAt)
	}

	execution.SetStepState(step.ID, &StepState{
		StepID:      step.ID,
		Status:      status,
		Input:       stepState.Input,
		Output:      result.Output,
		Error:       result.Error,
		Duration:    duration,
		AgentUsed:   step.Agent,
		RetryCount:  0,
	})

	return result
}

// executeTaskStep 执行任务步骤
func (e *Executor) executeTaskStep(ctx context.Context, execution *WorkflowExecution, step *Step) (interface{}, error) {
	// 查找合适的Agent
	var agent *aiagentorchestrator.AgentInfo
	var err error

	if step.Agent != "" {
		// 指定了Agent
		agent, err = e.registry.Get(step.Agent)
		if err != nil {
			return nil, fmt.Errorf("agent %s not found: %w", step.Agent, err)
		}
	} else {
		// 自动选择Agent
		// 根据工具能力选择
		capabilities := []string{}
		if step.Tool != "" {
			capabilities = append(capabilities, step.Tool)
		}
		agent, err = e.registry.FindBestAgent(capabilities)
		if err != nil {
			return nil, fmt.Errorf("no suitable agent found: %w", err)
		}
	}

	// 执行任务（这里简化实现，实际应该调用Agent）
	// TODO: 实际调用Agent的Execute方法
	output := fmt.Sprintf("Task '%s' executed by %s", step.Name, agent.Name)

	// 如果有工具，调用工具
	if step.Tool != "" {
		output = fmt.Sprintf("%s (using tool: %s)", output, step.Tool)
	}

	// 处理输入映射
	if len(step.Inputs) > 0 {
		for key, inputExpr := range step.Inputs {
			// 简化实现：直接使用输入表达式作为值
			if value, exists := execution.Inputs[inputExpr]; exists {
				// 替换输出中的占位符（如果有的话）
				output = fmt.Sprintf("%s (input: %s=%v)", output, key, value)
			}
		}
	}

	return output, nil
}

// executeConditionStep 执行条件步骤
func (e *Executor) executeConditionStep(ctx context.Context, execution *WorkflowExecution, step *Step) (interface{}, error) {
	if len(step.Conditions) == 0 {
		return nil, fmt.Errorf("no conditions defined for step %s", step.ID)
	}

	// 评估条件
	for _, condition := range step.Conditions {
		matched, err := e.evaluateCondition(ctx, execution, condition)
		if err != nil {
			return nil, err
		}

		if matched {
			// 返回Then分支
			return fmt.Sprintf("Condition matched, executing: %s", condition.Then), nil
		}
	}

	// 所有条件都不匹配，返回Else分支
	if step.Conditions[0].Else != "" {
		return fmt.Sprintf("No condition matched, executing: %s", step.Conditions[0].Else), nil
	}

	return nil, fmt.Errorf("no condition matched and no else branch")
}

// executeParallelStep 执行并行步骤
func (e *Executor) executeParallelStep(ctx context.Context, execution *WorkflowExecution, step *Step) (interface{}, error) {
	// 并行步骤实际上是一个容器，包含多个子步骤
	// 这里简化实现，实际应该递归调用executeLevel
	return fmt.Sprintf("Parallel step executed with %d sub-steps", len(step.DependsOn)), nil
}

// executeSequentialStep 执行顺序步骤
func (e *Executor) executeSequentialStep(ctx context.Context, execution *WorkflowExecution, step *Step) (interface{}, error) {
	// 顺序步骤实际上是一个容器，包含多个子步骤
	// 这里简化实现
	return fmt.Sprintf("Sequential step executed with %d sub-steps", len(step.DependsOn)), nil
}

// evaluateCondition 评估条件
func (e *Executor) evaluateCondition(ctx context.Context, execution *WorkflowExecution, condition *Condition) (bool, error) {
	// 获取变量值
	var varValue interface{}
	if value, exists := execution.Inputs[condition.Variable]; exists {
		varValue = value
	} else {
		return false, fmt.Errorf("variable %s not found in inputs", condition.Variable)
	}

	// 简单类型转换
	var typedValue interface{}
	switch v := varValue.(type) {
	case string:
		typedValue = v
	case float64:
		typedValue = int(v)
	case int:
		typedValue = v
	case bool:
		typedValue = v
	default:
		typedValue = fmt.Sprintf("%v", v)
	}

	// 根据操作符比较
	switch condition.Operator {
	case "eq":
		return fmt.Sprintf("%v", typedValue) == fmt.Sprintf("%v", condition.Value), nil
	case "ne":
		return fmt.Sprintf("%v", typedValue) != fmt.Sprintf("%v", condition.Value), nil
	case "gt":
		return compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), ">"), nil
	case "lt":
		return compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), "<"), nil
	case "gte":
		return compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), ">="), nil
	case "lte":
		return compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), "<="), nil
	case "in":
		return checkIn(typedValue, condition.Value)
	case "not_in":
		result, _ := checkIn(typedValue, condition.Value)
		return !result, nil
	case "contains":
		return checkContains(typedValue, condition.Value)
	default:
		return false, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

// compareNumbers 比较数字
func compareNumbers(a, b, operator string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch operator {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 转换为float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case string:
		f, err := parseFloat(val)
		return f, err == nil
	default:
		return 0, false
	}
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// checkIn 检查值是否在列表中
func checkIn(value interface{}, list interface{}) (bool, error) {
	listSlice, ok := list.([]interface{})
	if !ok {
		return false, nil
	}

	for _, item := range listSlice {
		if fmt.Sprintf("%v", item) == fmt.Sprintf("%v", value) {
			return true, nil
		}
	}

	return false, nil
}

// checkContains 检查是否包含字符串
func checkContains(value interface{}, substr interface{}) (bool, error) {
	valueStr, ok := value.(string)
	if !ok {
		return false, nil
	}

	substrStr, ok := substr.(string)
	if !ok {
		return false, nil
	}

	// 简单的字符串包含检查
	return strings.Contains(valueStr, substrStr), nil
}

// StepResult 步骤执行结果
type StepResult struct {
	StepID  string      `json:"step_id"`
	Success bool        `json:"success"`
	Output  interface{} `json:"output"`
	Error   string      `json:"error,omitempty"`
}
