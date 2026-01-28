package task

import (
	"context"
	"fmt"
	"strings"
)

// Decomposer 任务分解器接口
type Decomposer interface {
	Decompose(ctx context.Context, task *Task) ([]*Task, error)
}

// AIDecomposer AI驱动的任务分解器
type AIDecomposer struct {
	model LLMModel
}

// LLMModel LLM模型接口（简化版）
type LLMModel interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewAIDecomposer 创建AI分解器
func NewAIDecomposer(model LLMModel) *AIDecomposer {
	return &AIDecomposer{
		model: model,
	}
}

// Decompose 使用AI分解任务
func (d *AIDecomposer) Decompose(ctx context.Context, task *Task) ([]*Task, error) {
	// 构建分解提示
	prompt := d.buildDecompositionPrompt(task)

	messages := []Message{
		{Role: "system", Content: "你是一个任务规划专家，擅长将复杂任务分解为可执行的子任务。"},
		{Role: "user", Content: prompt},
	}

	// 调用LLM进行分解
	response, err := d.model.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM decomposition failed: %w", err)
	}

	// 解析LLM响应
	subTasks, err := d.parseDecomposition(response, task)
	if err != nil {
		return nil, fmt.Errorf("failed to parse decomposition: %w", err)
	}

	return subTasks, nil
}

// buildDecompositionPrompt 构建分解提示
func (d *AIDecomposer) buildDecompositionPrompt(task *Task) string {
	return fmt.Sprintf(`请将以下任务分解为3-5个可执行的子任务：

任务目标：%s

要求：
1. 每个子任务应该具体、可执行
2. 子任务之间应该有逻辑顺序
3. 明确每个子任务的输入和输出
4. 考虑任务之间的依赖关系

请以JSON格式返回，格式如下：
[
  {
    "id": "subtask-1",
    "goal": "子任务1的目标",
    "type": "single",
    "depends_on": []
  },
  {
    "id": "subtask-2",
    "goal": "子任务2的目标",
    "type": "single",
    "depends_on": ["subtask-1"]
  }
]`, task.Goal)
}

// parseDecomposition 解析分解结果
func (d *AIDecomposer) parseDecomposition(response string, parentTask *Task) ([]*Task, error) {
	// 简化实现：假设LLM返回的是纯JSON
	// 实际应用中需要更复杂的解析逻辑

	// 这里先用简单的规则来演示
	subTasks := make([]*Task, 0)

	// 提取任务中的关键词
	keywords := extractKeywords(parentTask.Goal)

	// 根据关键词生成子任务
	for i, keyword := range keywords {
		subTask := &Task{
			ID:       fmt.Sprintf("%s-subtask-%d", parentTask.ID, i+1),
			Type:     "single",
			Goal:     fmt.Sprintf("%s：%s", keyword, parentTask.Goal),
			Priority: parentTask.Priority,
			Status:   TaskStatusPending,
			Metadata: make(map[string]interface{}),
		}

		// 设置依赖关系
		if i > 0 {
			subTask.DependsOn = []string{fmt.Sprintf("%s-subtask-%d", parentTask.ID, i)}
		}

		subTasks = append(subTasks, subTask)
	}

	return subTasks, nil
}

// extractKeywords 从任务描述中提取关键词
func extractKeywords(goal string) []string {
	// 简单的关键词提取逻辑
	keywords := []string{}

	// 搜索类任务
	if strings.Contains(goal, "搜索") || strings.Contains(goal, "查找") {
		keywords = append(keywords, "搜索")
	}

	// 分析类任务
	if strings.Contains(goal, "分析") || strings.Contains(goal, "研究") {
		if len(keywords) == 0 {
			keywords = append(keywords, "收集信息")
		}
		keywords = append(keywords, "数据分析")
	}

	// 写作类任务
	if strings.Contains(goal, "写") || strings.Contains(goal, "生成") || strings.Contains(goal, "报告") {
		if len(keywords) == 0 {
			keywords = append(keywords, "收集素材")
		}
		keywords = append(keywords, "撰写内容")
	}

	// 编程类任务
	if strings.Contains(goal, "开发") || strings.Contains(goal, "写代码") {
		if len(keywords) == 0 {
			keywords = append(keywords, "设计方案")
		}
		keywords = append(keywords, "编写代码")
		keywords = append(keywords, "测试验证")
	}

	// 默认分解
	if len(keywords) == 0 {
		keywords = append(keywords, "任务规划")
		keywords = append(keywords, "执行任务")
		keywords = append(keywords, "验证结果")
	}

	return keywords
}

// TemplateDecomposer 模板化分解器
type TemplateDecomposer struct {
	templates map[string][]string // task_type -> subtasks template
}

// NewTemplateDecomposer 创建模板分解器
func NewTemplateDecomposer() *TemplateDecomposer {
	return &TemplateDecomposer{
		templates: map[string][]string{
			"research": {
				"确定研究范围和目标",
				"收集相关信息",
				"分析整理数据",
				"撰写研究报告",
			},
			"development": {
				"需求分析",
				"系统设计",
				"编码实现",
				"测试验证",
				"部署上线",
			},
			"analysis": {
				"数据收集",
				"数据清洗",
				"统计分析",
				"生成报告",
			},
			"writing": {
				"确定主题和结构",
				"收集素材",
				"撰写初稿",
				"修改润色",
			},
		},
	}
}

// Decompose 使用模板分解任务
func (d *TemplateDecomposer) Decompose(ctx context.Context, task *Task) ([]*Task, error) {
	// 确定任务类型
	taskType := d.determineTaskType(task.Goal)

	// 获取对应的模板
	template, exists := d.templates[taskType]
	if !exists {
		template = d.templates["research"] // 默认使用研究模板
	}

	// 根据模板生成子任务
	subTasks := make([]*Task, 0)
	for i, goal := range template {
		subTask := &Task{
			ID:       fmt.Sprintf("%s-%d", task.ID, i+1),
			Type:     "single",
			Goal:     goal,
			Priority: task.Priority,
			Status:   TaskStatusPending,
			Metadata: make(map[string]interface{}),
		}

		// 设置依赖关系（每个任务依赖前一个）
		if i > 0 {
			subTask.DependsOn = []string{fmt.Sprintf("%s-%d", task.ID, i)}
		}

		subTasks = append(subTasks, subTask)
	}

	return subTasks, nil
}

// determineTaskType 确定任务类型
func (d *TemplateDecomposer) determineTaskType(goal string) string {
	goal = strings.ToLower(goal)

	if strings.Contains(goal, "开发") || strings.Contains(goal, "编程") || strings.Contains(goal, "写代码") {
		return "development"
	}

	if strings.Contains(goal, "分析") || strings.Contains(goal, "统计") {
		return "analysis"
	}

	if strings.Contains(goal, "写") || strings.Contains(goal, "文章") || strings.Contains(goal, "报告") {
		return "writing"
	}

	if strings.Contains(goal, "研究") || strings.Contains(goal, "调查") || strings.Contains(goal, "搜索") {
		return "research"
	}

	return "research" // 默认为研究类型
}

// HierarchicalDecomposer 层级分解器
type HierarchicalDecomposer struct {
	maxDepth int
	decomposer Decomposer
}

// NewHierarchicalDecomposer 创建层级分解器
func NewHierarchicalDecomposer(decomposer Decomposer, maxDepth int) *HierarchicalDecomposer {
	return &HierarchicalDecomposer{
		maxDepth:  maxDepth,
		decomposer: decomposer,
	}
}

// Decompose 递归分解任务
func (d *HierarchicalDecomposer) Decompose(ctx context.Context, task *Task) ([]*Task, error) {
	return d.decomposeRecursive(ctx, task, 0)
}

// decomposeRecursive 递归分解
func (d *HierarchicalDecomposer) decomposeRecursive(ctx context.Context, task *Task, depth int) ([]*Task, error) {
	// 达到最大深度，停止分解
	if depth >= d.maxDepth {
		return []*Task{task}, nil
	}

	// 分解任务
	subTasks, err := d.decomposer.Decompose(ctx, task)
	if err != nil {
		return nil, err
	}

	// 如果子任务仍然很复杂，继续分解
	result := make([]*Task, 0)
	for _, subTask := range subTasks {
		// 检查子任务是否需要进一步分解
		if d.needsFurtherDecomposition(subTask) && depth < d.maxDepth-1 {
			furtherTasks, err := d.decomposeRecursive(ctx, subTask, depth+1)
			if err != nil {
				// 如果分解失败，保留原任务
				result = append(result, subTask)
			} else {
				result = append(result, furtherTasks...)
			}
		} else {
			result = append(result, subTask)
		}
	}

	return result, nil
}

// needsFurtherDecomposition 判断是否需要进一步分解
func (d *HierarchicalDecomposer) needsFurtherDecomposition(task *Task) bool {
	// 简单的启发式规则
	goal := task.Goal

	// 如果任务描述很长（超过100字符），可能需要分解
	if len(goal) > 100 {
		return true
	}

	// 如果任务包含多个动词，可能需要分解
	verbs := []string{"和", "并且", "然后", "之后", "同时"}
	for _, verb := range verbs {
		if strings.Contains(goal, verb) {
			return true
		}
	}

	return false
}
