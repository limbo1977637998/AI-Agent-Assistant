package workflow

import (
	"fmt"
	"sort"
)

// DAG 有向无环图
type DAG struct {
	nodes map[string]*Node  // node_id -> node
	edges map[string][]string // node_id -> dependent_nodes
}

// Node DAG节点
type Node struct {
	ID       string
	Step     *Step
	InDegree int  // 入度（依赖数量）
	Visited  bool // 是否已访问
}

// NewDAG 创建DAG
func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
	}
}

// AddNode 添加节点
func (d *DAG) AddNode(step *Step) error {
	if _, exists := d.nodes[step.ID]; exists {
		return fmt.Errorf("node %s already exists", step.ID)
	}

	d.nodes[step.ID] = &Node{
		ID:   step.ID,
		Step: step,
	}
	d.edges[step.ID] = make([]string, 0)

	return nil
}

// AddEdge 添加边（依赖关系）
func (d *DAG) AddEdge(from, to string) error {
	// from 依赖 to
	// 也就是说必须先完成to，才能执行from
	// 边的方向：to -> from
	if _, exists := d.nodes[from]; !exists {
		return fmt.Errorf("node %s does not exist", from)
	}
	if _, exists := d.nodes[to]; !exists {
		return fmt.Errorf("node %s does not exist", to)
	}

	// 检查是否会造成循环
	if d.wouldCreateCycle(from, to) {
		return fmt.Errorf("adding edge %s -> %s would create a cycle", to, from)
	}

	// edges[from] 表示from依赖的所有节点
	d.edges[from] = append(d.edges[from], to)
	d.nodes[from].InDegree++

	return nil
}

// wouldCreateCycle 检查是否会创建循环
func (d *DAG) wouldCreateCycle(from, to string) bool {
	visited := make(map[string]bool)
	return d.hasCycleDFS(to, from, visited)
}

// hasCycleDFS DFS检测循环
func (d *DAG) hasCycleDFS(current, target string, visited map[string]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true

	for _, neighbor := range d.edges[current] {
		if !visited[neighbor] {
			if d.hasCycleDFS(neighbor, target, visited) {
				return true
			}
		}
	}

	return false
}

// TopologicalSort 拓扑排序
func (d *DAG) TopologicalSort() ([]string, error) {
	// Kahn算法
	result := make([]string, 0)
	queue := make([]string, 0)

	// 初始化入度为0的节点
	for _, node := range d.nodes {
		node.InDegree = len(d.GetDepends(node.ID))
		if node.InDegree == 0 {
			queue = append(queue, node.ID)
		}
	}

	// 处理队列
	for len(queue) > 0 {
		// 排序保证确定性
		sort.Strings(queue)

		// 取出第一个节点
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)

		// 找到所有依赖当前节点的节点，减少它们的入度
		for nodeID, dependencies := range d.edges {
			if contains(dependencies, current) {
				d.nodes[nodeID].InDegree--
				if d.nodes[nodeID].InDegree == 0 {
					queue = append(queue, nodeID)
				}
			}
		}
	}

	// 检查是否有环
	if len(result) != len(d.nodes) {
		return nil, fmt.Errorf("cycle detected in workflow")
	}

	return result, nil
}

// contains 检查slice是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetDependencies 获取节点的所有依赖
func (d *DAG) GetDependencies(nodeID string) []string {
	if deps, exists := d.edges[nodeID]; exists {
		return deps
	}
	return []string{}
}

// GetDepends 获取节点依赖的节点（别名）
func (d *DAG) GetDepends(nodeID string) []string {
	return d.GetDependencies(nodeID)
}

// GetExecutableSteps 获取可并行执行的步骤组
func (d *DAG) GetExecutableSteps(completedSteps map[string]bool) [][]string {
	groups := make([][]string, 0)

	// 创建已完成步骤的副本
	completed := make(map[string]bool)
	for k, v := range completedSteps {
		completed[k] = v
	}

	for {
		// 找出所有依赖都已完成的节点
		ready := make([]string, 0)
		for nodeID := range d.nodes {
			if completed[nodeID] {
				continue
			}

			// 检查所有依赖是否完成
			allDepsCompleted := true
			for _, depID := range d.GetDependencies(nodeID) {
				if !completed[depID] {
					allDepsCompleted = false
					break
				}
			}

			if allDepsCompleted {
				ready = append(ready, nodeID)
			}
		}

		if len(ready) == 0 {
			break
		}

		// 这组节点可以并行执行
		groups = append(groups, ready)

		// 标记为已完成，继续下一轮
		for _, nodeID := range ready {
			completed[nodeID] = true
		}
	}

	return groups
}

// Validate 验证DAG
func (d *DAG) Validate() error {
	// 检查是否有孤立节点
	hasEdges := make(map[string]bool)
	for _, neighbors := range d.edges {
		for _, neighbor := range neighbors {
			hasEdges[neighbor] = true
		}
	}

	// 检查是否有环
	_, err := d.TopologicalSort()
	if err != nil {
		return fmt.Errorf("DAG validation failed: %w", err)
	}

	// 检查是否有未定义的依赖
	for nodeID, node := range d.nodes {
		for _, depID := range node.Step.DependsOn {
			if _, exists := d.nodes[depID]; !exists {
				return fmt.Errorf("node %s depends on undefined node %s", nodeID, depID)
			}
		}
	}

	return nil
}

// BuildDAGFromWorkflow 从工作流构建DAG
func BuildDAGFromWorkflow(workflow *Workflow) (*DAG, error) {
	dag := NewDAG()

	// 添加所有节点
	for _, step := range workflow.Steps {
		if err := dag.AddNode(step); err != nil {
			return nil, fmt.Errorf("failed to add node %s: %w", step.ID, err)
		}
	}

	// 添加所有边（依赖关系）
	for _, step := range workflow.Steps {
		for _, depID := range step.DependsOn {
			if err := dag.AddEdge(step.ID, depID); err != nil {
				return nil, fmt.Errorf("failed to add edge %s -> %s: %w", depID, step.ID, err)
			}
		}
	}

	// 验证DAG
	if err := dag.Validate(); err != nil {
		return nil, fmt.Errorf("DAG validation failed: %w", err)
	}

	return dag, nil
}

// GetLevels 获取DAG的层级结构（用于并行执行）
func (d *DAG) GetLevels() [][]string {
	levels := make([][]string, 0)
	completed := make(map[string]bool)

	for {
		ready := make([]string, 0)

		for nodeID := range d.nodes {
			if completed[nodeID] {
				continue
			}

			// 检查所有依赖是否已完成
			allDepsCompleted := true
			for _, depID := range d.GetDependencies(nodeID) {
				if !completed[depID] {
					allDepsCompleted = false
					break
				}
			}

			if allDepsCompleted {
				ready = append(ready, nodeID)
			}
		}

		if len(ready) == 0 {
			break
		}

		// 排序保证确定性
		sort.Strings(ready)

		levels = append(levels, ready)

		// 标记为已完成
		for _, nodeID := range ready {
			completed[nodeID] = true
		}
	}

	return levels
}

// Visualize 可视化DAG（返回文本格式）
func (d *DAG) Visualize() string {
	result := "DAG Structure:\n"

	// 按拓扑排序输出
	order, err := d.TopologicalSort()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	for _, nodeID := range order {
		deps := d.GetDependencies(nodeID)

		result += fmt.Sprintf("  %s", nodeID)
		if len(deps) > 0 {
			result += fmt.Sprintf(" (depends on: %v)", deps)
		}
		result += "\n"
	}

	return result
}
