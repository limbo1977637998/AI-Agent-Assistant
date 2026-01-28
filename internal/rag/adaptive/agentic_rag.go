package adaptive

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// AgenticRAG 代理式 RAG
//
// 核心思想:
//   使用智能代理（Agent）自主决策检索策略
//   代理可以观察、思考、行动和反思
//
// 代理能力:
//   1. 观察环境（查询分析、上下文评估）
//   2. 思考规划（策略选择、推理）
//   3. 执行行动（检索、生成）
//   4. 反思改进（自我纠错）
//
// 架构模式:
//   - ReAct Pattern (Reasoning + Acting)
//   - Plan-and-Execute
//   - Reflexion Pattern
//
// 论文基础:
//   "ReAct: Synergizing Reasoning and Acting in Language Models"
//   "Plan-and-Solve: Breaking Down Complex Questions"
type AgenticRAG struct {
	llm          LLMProvider
	tools        []AgentTool
	memory       *AgentMemory
	planner      Planner
	executor     Executor
	reflector    Reflector
	config       AgenticRAGConfig
	state        *AgentState
	mu           sync.RWMutex
}

// AgenticRAGConfig Agentic RAG 配置
type AgenticRAGConfig struct {
	// MaxIterations 最大迭代次数
	MaxIterations int

	// EnableReAct 是否启用 ReAct 模式
	EnableReAct bool

	// EnablePlanning 是否启用规划模式
	EnablePlanning bool

	// EnableReflection 是否启用反思模式
	EnableReflection bool

	// ToolTimeout 工具执行超时（毫秒）
	ToolTimeout int64
}

// AgentTool 代理工具接口
type AgentTool interface {
	// Name 工具名称
	Name() string

	// Description 工具描述
	Description() string

	// Execute 执行工具
	Execute(ctx context.Context, input string) (string, error)

	// ValidateInput 验证输入
	ValidateInput(input string) bool
}

// AgentMemory 代理记忆
type AgentMemory struct {
	observations  []Observation
	thoughts      []Thought
	actions       []Action
	steps         []string
	mu            sync.RWMutex
}

// Observation 观察
type Observation struct {
	Content string
	Type    string
}

// Thought 思考
type Thought struct {
	Content string
	Reasoning string
}

// Action 行动
type Action struct {
	Tool   string
	Input  string
	Output string
}

// Planner 规划器接口
type Planner interface {
	Plan(ctx context.Context, query string) (*Plan, error)
}

// Executor 执行器接口
type Executor interface {
	Execute(ctx context.Context, action *Action) (string, error)
}

// Reflector 反思器接口
type Reflector interface {
	Reflect(ctx context.Context, state *AgentState) (*Reflection, error)
}

// Plan 计划
type Plan struct {
	Goal      string
	Steps     []PlanStep
	Reasoning string
}

// PlanStep 计划步骤
type PlanStep struct {
	Step        int
	Description string
	Tool        string
	Input       string
	DependsOn   []int
}

// Reflection 反思
type Reflection struct {
	Content      string
	NeedAdjust   bool
	Adjustments  []string
	Confidence   float64
}

// AgentState 代理状态
type AgentState struct {
	Query          string
	CurrentStep    int
	Iterations     int
	Completed      bool
	Observations   []Observation
	Thoughts       []Thought
	Actions        []Action
	Answer         string
	Confidence     float64
}

// DefaultAgenticRAGConfig 返回默认配置
func DefaultAgenticRAGConfig() AgenticRAGConfig {
	return AgenticRAGConfig{
		MaxIterations:   10,
		EnableReAct:     true,
		EnablePlanning:  true,
		EnableReflection: true,
		ToolTimeout:     30000, // 30 秒
	}
}

// NewAgenticRAG 创建代理式 RAG
func NewAgenticRAG(llm LLMProvider, config AgenticRAGConfig) (*AgenticRAG, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	agent := &AgenticRAG{
		llm:      llm,
		tools:    make([]AgentTool, 0),
		memory:   &AgentMemory{},
		planner:  &DefaultPlanner{llm: llm},
		executor: &DefaultExecutor{llm: llm},
		reflector: &DefaultReflector{llm: llm},
		config:   config,
		state:    &AgentState{},
	}

	// 注册默认工具
	agent.registerDefaultTools()

	return agent, nil
}

// registerDefaultTools 注册默认工具
func (ar *AgenticRAG) registerDefaultTools() {
	ar.tools = append(ar.tools,
		&VectorSearchTool{},
		&GraphSearchTool{},
		&HybridSearchTool{},
		&KnowledgeQueryTool{},
	)
}

// Query 执行代理式查询
func (ar *AgenticRAG) Query(ctx context.Context, query string) (*AgentResult, error) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// 初始化状态
	ar.state = &AgentState{
		Query:        query,
		CurrentStep: 0,
		Iterations:   0,
		Completed:    false,
		Observations: make([]Observation, 0),
		Thoughts:     make([]Thought, 0),
		Actions:      make([]Action, 0),
		Confidence:   0.0,
	}

	// ReAct 模式
	if ar.config.EnableReAct {
		return ar.reactMode(ctx, query)
	}

	// Plan-and-Execute 模式
	if ar.config.EnablePlanning {
		return ar.planAndExecuteMode(ctx, query)
	}

	// 默认使用 ReAct
	return ar.reactMode(ctx, query)
}

// reactMode ReAct 模式（推理 + 行动）
// 论文: "ReAct: Synergizing Reasoning and Acting in Language Models"
func (ar *AgenticRAG) reactMode(ctx context.Context, query string) (*AgentResult, error) {
	fmt.Printf("[ReAct] 开始处理查询: %s\n", query)

	for ar.state.Iterations < ar.config.MaxIterations && !ar.state.Completed {
		ar.state.Iterations++

		// Step 1: Thought (思考)
		thought := ar.generateThought(ctx, query)
		ar.state.Thoughts = append(ar.state.Thoughts, *thought)
		fmt.Printf("[Thought %d] %s\n", ar.state.Iterations, thought.Content)

		// Step 2: Action (行动)
		action := ar.decideAction(ctx, thought)
		ar.state.Actions = append(ar.state.Actions, *action)
		fmt.Printf("[Action %d] 使用工具: %s, 输入: %s\n", ar.state.Iterations, action.Tool, action.Input)

		// Step 3: Observation (观察)
		observation := ar.executeAction(ctx, action)
		ar.state.Observations = append(ar.state.Observations, *observation)
		fmt.Printf("[Observation %d] %s\n", ar.state.Iterations, observation.Content)

		// Step 4: Check if complete (检查是否完成)
		if ar.checkCompletion(ctx) {
			ar.state.Completed = true
			break
		}
	}

	// Step 5: Generate Answer (生成答案)
	answer := ar.generateAnswer(ctx)
	ar.state.Answer = answer

	// Step 6: Reflection (反思)
	if ar.config.EnableReflection {
		reflection, _ := ar.reflector.Reflect(ctx, ar.state)
		ar.state.Confidence = reflection.Confidence

		if reflection.NeedAdjust {
			// 可以根据反思结果调整
			fmt.Printf("[Reflection] 需要调整: %v\n", reflection.Adjustments)
		}
	}

	return &AgentResult{
		Query:         query,
		Answer:        ar.state.Answer,
		Thoughts:      ar.state.Thoughts,
		Actions:       ar.state.Actions,
		Observations:  ar.state.Observations,
		Iterations:    ar.state.Iterations,
		Confidence:    ar.state.Confidence,
	}, nil
}

// planAndExecuteMode Plan-and-Execute 模式
func (ar *AgenticRAG) planAndExecuteMode(ctx context.Context, query string) (*AgentResult, error) {
	fmt.Printf("[Plan-and-Execute] 开始处理查询: %s\n", query)

	// Step 1: Plan (规划)
	plan, err := ar.planner.Plan(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	fmt.Printf("[Plan] 目标: %s\n", plan.Goal)
	fmt.Printf("[Plan] 步骤数: %d\n", len(plan.Steps))

	// Step 2: Execute (执行计划)
	for i, step := range plan.Steps {
		ar.state.CurrentStep = i + 1
		ar.state.Iterations++

		// 检查依赖
		if !ar.checkDependencies(step, ar.state) {
			return nil, fmt.Errorf("dependencies not met for step %d", i)
		}

		// 执行步骤
		action := &Action{
			Tool:  step.Tool,
			Input: step.Input,
		}

		observation := ar.executeAction(ctx, action)
		ar.state.Actions = append(ar.state.Actions, *action)
		ar.state.Observations = append(ar.state.Observations, *observation)

		fmt.Printf("[Step %d] %s: %s\n", i+1, step.Description, observation.Content)
	}

	// Step 3: Generate Answer (生成答案)
	answer := ar.generateAnswer(ctx)
	ar.state.Answer = answer

	return &AgentResult{
		Query:         query,
		Answer:        answer,
		Actions:       ar.state.Actions,
		Observations:  ar.state.Observations,
		Iterations:    ar.state.Iterations,
		Confidence:    0.8, // 默认置信度
	}, nil
}

// generateThought 生成思考
func (ar *AgenticRAG) generateThought(ctx context.Context, query string) *Thought {
	// 构建思考提示
	prompt := ar.buildThoughtPrompt(query, ar.state)

	response, err := ar.llm.Generate(ctx, prompt)
	if err != nil {
		return &Thought{
			Content: "思考生成失败",
			Reasoning: err.Error(),
		}
	}

	return &Thought{
		Content:  strings.TrimSpace(response),
		Reasoning: "基于查询和观察进行推理",
	}
}

// buildThoughtPrompt 构建思考提示
func (ar *AgenticRAG) buildThoughtPrompt(query string, state *AgentState) string {
	var sb strings.Builder

	sb.WriteString("你是智能检索助手，请分析当前查询并决定下一步行动。\n\n")
	sb.WriteString(fmt.Sprintf("查询: %s\n\n", query))

	if len(state.Observations) > 0 {
		sb.WriteString("历史观察:\n")
		for i, obs := range state.Observations {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, obs.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("可用工具:\n")
	for _, tool := range ar.tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}
	sb.WriteString("\n")

	sb.WriteString("请思考:\n")
	sb.WriteString("1. 分析查询的意图和需求\n")
	sb.WriteString("2. 评估当前信息是否足够\n")
	sb.WriteString("3. 决定下一步行动\n")
	sb.WriteString("4. 说明理由\n\n")
	sb.WriteString("输出格式:\n")
	sb.WriteString("Thought: [你的思考过程]\n")

	return sb.String()
}

// decideAction 决定行动
func (ar *AgenticRAG) decideAction(ctx context.Context, thought *Thought) *Action {
	// 从思考中提取行动
	content := strings.ToLower(thought.Content)

	// 简化实现：基于关键词匹配
	for _, tool := range ar.tools {
		toolName := strings.ToLower(tool.Name())
		if strings.Contains(content, toolName) || strings.Contains(content, toolName+"_search") {
			// 提取输入
			input := ar.extractActionInput(content, tool.Name())
			return &Action{
				Tool:  tool.Name(),
				Input: input,
			}
		}
	}

	// 默认：使用向量搜索
	return &Action{
		Tool:  "vector_search",
		Input: ar.state.Query,
	}
}

// extractActionInput 从思考中提取行动输入
func (ar *AgenticRAG) extractActionInput(content, toolName string) string {
	// 简化实现：查找引号内容或使用查询
	quotes := regexp.MustCompile(`"([^"]+)"`)
	matches := quotes.FindAllStringSubmatch(content, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1]
	}

	// 默认使用原始查询
	return ar.state.Query
}

// executeAction 执行行动
func (ar *AgenticRAG) executeAction(ctx context.Context, action *Action) *Observation {
	// 查找工具
	var tool AgentTool
	for _, t := range ar.tools {
		if t.Name() == action.Tool {
			tool = t
			break
		}
	}

	if tool == nil {
		return &Observation{
			Content: fmt.Sprintf("错误: 未找到工具 %s", action.Tool),
			Type:    "error",
		}
	}

	// 验证输入
	if !tool.ValidateInput(action.Input) {
		return &Observation{
			Content: fmt.Sprintf("错误: 输入验证失败"),
			Type:    "error",
		}
	}

	// 执行工具
	output, err := tool.Execute(ctx, action.Input)
	action.Output = output

	if err != nil {
		return &Observation{
			Content: fmt.Sprintf("错误: %v", err),
			Type:    "error",
		}
	}

	return &Observation{
		Content: output,
		Type:    "success",
	}
}

// checkCompletion 检查是否完成
func (ar *AgenticRAG) checkCompletion(ctx context.Context) bool {
	// 检查条件:
	// 1. 已有足够的观察
	if len(ar.state.Observations) < 2 {
		return false
	}

	// 2. 最后一个观察是成功的
	lastObs := ar.state.Observations[len(ar.state.Observations)-1]
	if lastObs.Type == "error" {
		return false
	}

	// 3. 达到最大迭代次数
	if ar.state.Iterations >= ar.config.MaxIterations {
		return true
	}

	// 4. 使用 LLM 判断
	return ar.askCompletion(ctx)
}

// askCompletion 询问是否完成
func (ar *AgenticRAG) askCompletion(ctx context.Context) bool {
	prompt := fmt.Sprintf(`基于以下信息，判断是否已经收集到足够的信息来回答查询。

查询: %s

观察数量: %d
最后观察: %s

请回答 "yes" 或 "no"。`, ar.state.Query, len(ar.state.Observations), ar.state.Observations[len(ar.state.Observations)-1].Content)

	response, err := ar.llm.Generate(ctx, prompt)
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return strings.Contains(response, "yes") || strings.Contains(response, "是")
}

// generateAnswer 生成答案
func (ar *AgenticRAG) generateAnswer(ctx context.Context) string {
	// 构建答案提示
	var sb strings.Builder

	sb.WriteString("基于以下信息回答查询。\n\n")
	sb.WriteString(fmt.Sprintf("查询: %s\n\n", ar.state.Query))

	if len(ar.state.Observations) > 0 {
		sb.WriteString("相关信息:\n")
		for i, obs := range ar.state.Observations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, obs.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("要求:\n")
	sb.WriteString("1. 基于观察信息回答\n")
	sb.WriteString("2. 如果信息不足，诚实说明\n")
	sb.WriteString("3. 答案要清晰、准确、有条理\n\n")
	sb.WriteString("回答:")

	prompt := sb.String()

	response, err := ar.llm.Generate(ctx, prompt)
	if err != nil {
		return "答案生成失败"
	}

	return strings.TrimSpace(response)
}

// checkDependencies 检查步骤依赖
func (ar *AgenticRAG) checkDependencies(step PlanStep, state *AgentState) bool {
	if len(step.DependsOn) == 0 {
		return true
	}

	// 检查所有依赖步骤是否已完成
	for _, depIndex := range step.DependsOn {
		if depIndex >= state.CurrentStep {
			return false
		}
	}

	return true
}

// AgentResult 代理结果
type AgentResult struct {
	Query         string
	Answer        string
	Thoughts      []Thought
	Actions       []Action
	Observations  []Observation
	Iterations    int
	Confidence    float64
}

// ===== 默认工具实现 =====

// VectorSearchTool 向量搜索工具
type VectorSearchTool struct {
	// 简化实现，移除接口依赖
}

func (t *VectorSearchTool) Name() string {
	return "vector_search"
}

func (t *VectorSearchTool) Description() string {
	return "使用向量相似度搜索文档"
}

func (t *VectorSearchTool) ValidateInput(input string) bool {
	return len(input) > 0
}

func (t *VectorSearchTool) Execute(ctx context.Context, input string) (string, error) {
	// 简化实现：返回模拟结果
	return fmt.Sprintf("向量搜索 '%s' 完成", input), nil
}

// GraphSearchTool 图搜索工具
type GraphSearchTool struct {
	// 简化实现，移除接口依赖
}

func (t *GraphSearchTool) Name() string {
	return "graph_search"
}

func (t *GraphSearchTool) Description() string {
	return "使用知识图谱搜索实体关系"
}

func (t *GraphSearchTool) ValidateInput(input string) bool {
	return len(input) > 0
}

func (t *GraphSearchTool) Execute(ctx context.Context, input string) (string, error) {
	// 简化实现：返回模拟结果
	return fmt.Sprintf("图谱搜索 '%s' 完成", input), nil
}

// HybridSearchTool 混合搜索工具
type HybridSearchTool struct {
	// 简化实现，移除接口依赖
}

func (t *HybridSearchTool) Name() string {
	return "hybrid_search"
}

func (t *HybridSearchTool) Description() string {
	return "使用混合检索（向量+关键词）"
}

func (t *HybridSearchTool) ValidateInput(input string) bool {
	return len(input) > 0
}

func (t *HybridSearchTool) Execute(ctx context.Context, input string) (string, error) {
	// 简化实现：返回模拟结果
	return fmt.Sprintf("混合搜索 '%s' 完成", input), nil
}

// KnowledgeQueryTool 知识库查询工具
type KnowledgeQueryTool struct {
	rag interface{}
}

func (t *KnowledgeQueryTool) Name() string {
	return "knowledge_query"
}

func (t *KnowledgeQueryTool) Description() string {
	return "查询知识库中的特定信息"
}

func (t *KnowledgeQueryTool) ValidateInput(input string) bool {
	return len(input) > 0
}

func (t *KnowledgeQueryTool) Execute(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("知识库查询: %s", input), nil
}

// ===== 默认规划器和执行器 =====

// DefaultPlanner 默认规划器
type DefaultPlanner struct {
	llm LLMProvider
}

func (p *DefaultPlanner) Plan(ctx context.Context, query string) (*Plan, error) {
	prompt := fmt.Sprintf(`请为以下查询制定一个执行计划。

查询: %s

可用工具:
- vector_search: 向量搜索
- graph_search: 图谱搜索
- hybrid_search: 混合搜索
- knowledge_query: 知识库查询

要求:
1. 分解查询为多个步骤
2. 为每个步骤选择合适的工具
3. 考虑步骤间的依赖关系
4. 说明整体目标

回答格式:
Goal: [整体目标]
Steps:
1. [描述] - Tool: [工具] - Input: [输入] - Depends: [依赖步骤]
2. ...

Plan:`, query)

	_, err := p.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化解析
	plan := &Plan{
		Goal:      "回答查询: " + query,
		Steps:     make([]PlanStep, 0),
		Reasoning: "基于查询类型制定计划",
	}

	// 默认单步计划
	plan.Steps = append(plan.Steps, PlanStep{
		Step:        1,
		Description: "检索相关信息",
		Tool:        "vector_search",
		Input:       query,
		DependsOn:   []int{},
	})

	return plan, nil
}

// DefaultExecutor 默认执行器
type DefaultExecutor struct {
	llm LLMProvider
}

func (e *DefaultExecutor) Execute(ctx context.Context, action *Action) (string, error) {
	return fmt.Sprintf("执行 %s: %s", action.Tool, action.Input), nil
}

// DefaultReflector 默认反思器
type DefaultReflector struct {
	llm LLMProvider
}

func (r *DefaultReflector) Reflect(ctx context.Context, state *AgentState) (*Reflection, error) {
	prompt := fmt.Sprintf(`请反思以下查询处理过程，并提供改进建议。

查询: %s
迭代次数: %d
观察数量: %d

反思:
1. 答案是否充分？
2. 是否需要更多信息？
3. 查询策略是否合适？

回答格式:
{
  "content": "反思内容",
  "need_adjust": true/false,
  "adjustments": ["建议1", "建议2"],
  "confidence": 0.0-1.0
}`, state.Query, state.Iterations, len(state.Observations))

	response, err := r.llm.Generate(ctx, prompt)
	if err != nil {
		return &Reflection{
			Content:    "反思失败",
			NeedAdjust: false,
			Confidence: 0.5,
		}, nil
	}

	return &Reflection{
		Content:     response,
		NeedAdjust:  false,
		Adjustments:  make([]string, 0),
		Confidence:   0.8,
	}, nil
}

// AddTool 添加工具
func (ar *AgenticRAG) AddTool(tool AgentTool) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.tools = append(ar.tools, tool)
}

// GetState 获取当前状态
func (ar *AgenticRAG) GetState() *AgentState {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return ar.state
}
