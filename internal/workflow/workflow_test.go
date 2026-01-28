package workflow

import (
	"testing"
)

// TestWorkflowDefinition 测试工作流定义
func TestWorkflowDefinition(t *testing.T) {
	// 创建工作流
	workflow := NewWorkflow("test-workflow", "测试工作流")

	// 验证基本属性
	if workflow.Name != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got '%s'", workflow.Name)
	}

	if workflow.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", workflow.Version)
	}

	// 添加步骤
	step1 := &Step{
		ID:          "step-1",
		Name:        "第一步",
		Type:        "task",
		Agent:       "worker-1",
		Description: "执行第一个任务",
	}

	step2 := &Step{
		ID:          "step-2",
		Name:        "第二步",
		Type:        "task",
		Agent:       "worker-2",
		Description: "执行第二个任务",
		DependsOn:   []string{"step-1"},
	}

	workflow.AddStep(step1)
	workflow.AddStep(step2)

	// 验证步骤数量
	if len(workflow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(workflow.Steps))
	}

	// 测试获取步骤
	retrievedStep := workflow.GetStep("step-1")
	if retrievedStep == nil {
		t.Error("Failed to retrieve step-1")
	}

	if retrievedStep.Name != "第一步" {
		t.Errorf("Expected step name '第一步', got '%s'", retrievedStep.Name)
	}
}

// TestWorkflowParser 测试工作流解析器
func TestWorkflowParser(t *testing.T) {
	parser := NewParser(".")

	// YAML定义
	yamlDef := `
name: research_workflow
description: 研究工作流
version: 1.0

agents:
  - name: researcher
    type: expert
    role: 搜索信息
    capabilities:
      - search
      - analyze

steps:
  - id: search
    name: 搜索信息
    type: task
    agent: researcher
    tool: web_search
    config:
      query: "Golang latest version"

  - id: analyze
    name: 分析结果
    type: task
    agent: researcher
    depends_on:
      - search
`

	// 解析YAML
	workflow, err := parser.ParseFromString(yamlDef, "yaml")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// 验证解析结果
	if workflow.Name != "research_workflow" {
		t.Errorf("Expected workflow name 'research_workflow', got '%s'", workflow.Name)
	}

	if len(workflow.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(workflow.Agents))
	}

	if len(workflow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(workflow.Steps))
	}

	// 验证步骤依赖关系
	step2 := workflow.GetStep("analyze")
	if step2 == nil {
		t.Fatal("Failed to get analyze step")
	}

	if len(step2.DependsOn) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(step2.DependsOn))
	}

	if step2.DependsOn[0] != "search" {
		t.Errorf("Expected dependency 'search', got '%s'", step2.DependsOn[0])
	}
}

// TestDAG 测试DAG构建和拓扑排序
func TestDAG(t *testing.T) {
	// 创建工作流
	workflow := NewWorkflow("dag-test", "DAG测试")

	// 添加步骤（形成DAG）
	workflow.AddStep(&Step{ID: "A", Name: "任务A", Type: "task"})
	workflow.AddStep(&Step{ID: "B", Name: "任务B", Type: "task", DependsOn: []string{"A"}})
	workflow.AddStep(&Step{ID: "C", Name: "任务C", Type: "task", DependsOn: []string{"A"}})
	workflow.AddStep(&Step{ID: "D", Name: "任务D", Type: "task", DependsOn: []string{"B", "C"}})

	// 构建DAG
	dag, err := BuildDAGFromWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to build DAG: %v", err)
	}

	// 测试拓扑排序
	order, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}

	// 验证排序结果
	if len(order) != 4 {
		t.Errorf("Expected 4 nodes in order, got %d", len(order))
	}

	// A必须在B和C之前
	aIndex := indexOf(order, "A")
	bIndex := indexOf(order, "B")
	cIndex := indexOf(order, "C")

	if aIndex >= bIndex {
		t.Error("A should come before B")
	}

	if aIndex >= cIndex {
		t.Error("A should come before C")
	}

	// D必须在B和C之后
	dIndex := indexOf(order, "D")

	if bIndex >= dIndex {
		t.Error("B should come before D")
	}

	if cIndex >= dIndex {
		t.Error("C should come before D")
	}

	// 测试层级结构
	levels := dag.GetLevels()
	if len(levels) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(levels))
	}

	// 第一层应该只有A
	if len(levels[0]) != 1 || levels[0][0] != "A" {
		t.Error("First level should contain only A")
	}

	// 测试可视化
	viz := dag.Visualize()
	t.Logf("DAG Visualization:\n%s", viz)
}

// TestDAGCycleDetection 测试循环检测
func TestDAGCycleDetection(t *testing.T) {
	// 创建有循环的工作流
	workflow := NewWorkflow("cycle-test", "循环测试")

	workflow.AddStep(&Step{ID: "A", Name: "任务A", Type: "task"})
	workflow.AddStep(&Step{ID: "B", Name: "任务B", Type: "task", DependsOn: []string{"A"}})
	workflow.AddStep(&Step{ID: "C", Name: "任务C", Type: "task", DependsOn: []string{"B"}})
	// A依赖C，形成循环
	workflow.AddStep(&Step{ID: "A2", Name: "任务A2", Type: "task", DependsOn: []string{"C"}})

	dag := NewDAG()

	// 添加节点
	for _, step := range workflow.Steps {
		dag.AddNode(step)
	}

	// 添加边（除了会形成循环的最后一条）
	for _, step := range workflow.Steps {
		if step.ID == "A2" {
			continue // 跳过会造成循环的边
		}
		for _, depID := range step.DependsOn {
			dag.AddEdge(step.ID, depID)
		}
	}

	// 现在尝试添加循环的边
	err := dag.AddEdge("A", "C")
	if err == nil {
		t.Error("Expected error when adding cycle, got nil")
	}

	// 验证错误消息
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestWorkflowExecution 测试工作流执行
func TestWorkflowExecution(t *testing.T) {
	workflow := NewWorkflow("exec-test", "执行测试")
	workflow.AddStep(&Step{ID: "step-1", Name: "步骤1", Type: "task"})

	// 创建执行实例
	inputs := map[string]interface{}{
		"topic": "Golang",
	}

	execution := NewWorkflowExecution(workflow, inputs)

	// 验证执行实例
	if execution.WorkflowID != workflow.ID {
		t.Errorf("Expected workflow ID %s, got %s", workflow.ID, execution.WorkflowID)
	}

	if execution.Status != WorkflowStatusPending {
		t.Errorf("Expected status pending, got %s", execution.Status)
	}

	// 标记为完成
	execution.MarkCompleted()

	if execution.Status != WorkflowStatusCompleted {
		t.Errorf("Expected status completed, got %s", execution.Status)
	}

	if execution.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

// TestDAGParallelExecution 测试并行执行
func TestDAGParallelExecution(t *testing.T) {
	workflow := NewWorkflow("parallel-test", "并行执行测试")

	// A -> B -> D
	//      -> C
	workflow.AddStep(&Step{ID: "A", Name: "任务A", Type: "task"})
	workflow.AddStep(&Step{ID: "B", Name: "任务B", Type: "task", DependsOn: []string{"A"}})
	workflow.AddStep(&Step{ID: "C", Name: "任务C", Type: "task", DependsOn: []string{"A"}})
	workflow.AddStep(&Step{ID: "D", Name: "任务D", Type: "task", DependsOn: []string{"B", "C"}})

	dag, _ := BuildDAGFromWorkflow(workflow)

	// 获取可执行的步骤组
	groups := dag.GetExecutableSteps(map[string]bool{})

	t.Logf("Executable groups: %v", groups)

	// 验证分组
	if len(groups) != 3 {
		t.Errorf("Expected 3 execution groups, got %d", len(groups))
	}

	// 第一组应该只有A
	if len(groups[0]) != 1 || groups[0][0] != "A" {
		t.Error("First group should contain only A")
	}

	// 第二组应该有B和C（可以并行）
	if len(groups[1]) != 2 {
		t.Errorf("Second group should have 2 parallel tasks, got %d", len(groups[1]))
	}

	// 第三组应该只有D
	if len(groups[2]) != 1 || groups[2][0] != "D" {
		t.Error("Third group should contain only D")
	}
}

// Helper function
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
