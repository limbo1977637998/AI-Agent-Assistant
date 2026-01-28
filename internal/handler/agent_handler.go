package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	aiagentconfig "ai-agent-assistant/internal/config"
	aiagentexpert "ai-agent-assistant/internal/agent/expert"
	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	aiagenttask "ai-agent-assistant/internal/task"
	aitools "ai-agent-assistant/internal/tools"
	"ai-agent-assistant/internal/workflow"

	"github.com/gin-gonic/gin"
)

// AgentHandler Agent处理器
// 负责处理所有与Agent相关的HTTP请求
type AgentHandler struct {
	config           *aiagentconfig.Config           // 配置信息
	agentFactory     *aiagentexpert.Factory          // Agent工厂
	agentRegistry    *aiagentorchestrator.AgentRegistry // Agent注册表
	taskScheduler    *aiagentorchestrator.TaskScheduler // 任务调度器
	workflowExecutor *workflow.Executor              // 工作流执行器
	stateManager     *workflow.StateManager          // 状态管理器
	toolManager      *aitools.ToolManager            // 工具管理器
}

// NewAgentHandler 创建Agent处理器
// 参数：
//   - cfg: 应用配置
//   - factory: Agent工厂实例
//   - registry: Agent注册表
//   - scheduler: 任务调度器
// 返回：
//   - *AgentHandler: Agent处理器实例
func NewAgentHandler(
	cfg *aiagentconfig.Config,
	factory *aiagentexpert.Factory,
	registry *aiagentorchestrator.AgentRegistry,
	scheduler *aiagentorchestrator.TaskScheduler,
) *AgentHandler {
	// 创建工作流执行器
	workflowExecutor := workflow.NewExecutor(registry, scheduler)

	// 创建工具管理器
	toolManager := aitools.NewToolManager(&aitools.ToolManagerConfig{
		AutoRegister: true,
	})

	// 将工具管理器设置到工厂
	factory.SetToolManager(toolManager)

	return &AgentHandler{
		config:           cfg,
		agentFactory:     factory,
		agentRegistry:    registry,
		taskScheduler:    scheduler,
		workflowExecutor: workflowExecutor,
		stateManager:     workflow.NewStateManager(),
		toolManager:      toolManager,
	}
}

// RegisterRoutes 注册Agent相关的路由
// 将所有Agent相关的API端点注册到Gin路由器
func (h *AgentHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Agent管理相关路由
	agentGroup := router.Group("/agents")
	{
		// GET /agents - 获取所有Agent列表
		agentGroup.GET("", h.ListAgents)

		// GET /agents/:id - 获取指定Agent的详细信息
		agentGroup.GET("/:id", h.GetAgent)

		// GET /agents/:id/capabilities - 获取Agent的能力列表
		agentGroup.GET("/:id/capabilities", h.GetAgentCapabilities)

		// GET /agents/:id/status - 获取Agent的当前状态
		agentGroup.GET("/:id/status", h.GetAgentStatus)

		// POST /agents/:id/heartbeat - 更新Agent心跳
		agentGroup.POST("/:id/heartbeat", h.UpdateAgentHeartbeat)
	}

	// 任务执行相关路由
	taskGroup := router.Group("/tasks")
	{
		// POST /tasks - 创建并执行新任务
		taskGroup.POST("", h.ExecuteTask)

		// GET /tasks/:id - 获取任务执行状态
		taskGroup.GET("/:id", h.GetTaskStatus)

		// POST /tasks/batch - 批量执行任务
		taskGroup.POST("/batch", h.ExecuteBatchTasks)
	}

	// 工作流相关路由
	workflowGroup := router.Group("/workflows")
	{
		// POST /workflows - 创建新工作流
		workflowGroup.POST("", h.CreateWorkflow)

		// GET /workflows - 获取所有工作流列表
		workflowGroup.GET("", h.ListWorkflows)

		// GET /workflows/:id - 获取工作流详情
		workflowGroup.GET("/:id", h.GetWorkflow)

		// POST /workflows/:id/execute - 执行工作流
		workflowGroup.POST("/:id/execute", h.ExecuteWorkflow)

		// GET /workflows/:id/executions - 获取工作流执行历史
		workflowGroup.GET("/:id/executions", h.GetWorkflowExecutions)

		// DELETE /workflows/:id - 删除工作流
		workflowGroup.DELETE("/:id", h.DeleteWorkflow)
	}

	// 分析和研究相关路由
	analysisGroup := router.Group("/analysis")
	{
		// POST /analysis/search - 执行网络搜索
		analysisGroup.POST("/search", h.PerformSearch)

		// POST /analysis/analyze - 执行数据分析
		analysisGroup.POST("/analyze", h.PerformAnalysis)

		// POST /analysis/write - 执行内容生成
		analysisGroup.POST("/write", h.PerformWriting)

		// POST /analysis/report - 生成分析报告
		analysisGroup.POST("/report", h.GenerateReport)
	}

	// 工具相关路由
	toolsGroup := router.Group("/tools")
	{
		// GET /tools - 获取所有可用工具列表
		toolsGroup.GET("", h.ListTools)

		// GET /tools/:name - 获取指定工具的详细信息
		toolsGroup.GET("/:name", h.GetToolInfo)

		// GET /tools/:name/capabilities - 获取工具的能力描述
		toolsGroup.GET("/:name/capabilities", h.GetToolCapabilities)

		// POST /tools/execute - 执行工具操作
		toolsGroup.POST("/execute", h.ExecuteTool)

		// POST /tools/batch - 批量执行工具
		toolsGroup.POST("/batch", h.BatchExecuteTools)

		// GET /tools/chains - 获取所有工具链
		toolsGroup.GET("/chains", h.ListToolChains)

		// POST /tools/chains/:name/execute - 执行工具链
		toolsGroup.POST("/chains/:name/execute", h.ExecuteToolChain)
	}
}

// ListAgents 获取所有Agent列表
// 返回系统中所有可用的Agent及其基本信息
//
// 响应示例：
// {
//   "agents": [
//     {
//       "id": "researcher-001",
//       "name": "Researcher",
//       "type": "researcher",
//       "capabilities": ["web_search", "information_collection"],
//       "status": "idle"
//     }
//   ],
//   "total": 3
// }
func (h *AgentHandler) ListAgents(c *gin.Context) {
	// 从工厂获取所有Agent的信息
	agentsInfo := h.agentFactory.GetAgentInfo()

	// 返回Agent列表
	c.JSON(http.StatusOK, gin.H{
		"agents": agentsInfo,
		"total":  len(agentsInfo),
	})
}

// GetAgent 获取指定Agent的详细信息
// 参数：
//   - id: Agent的ID（路径参数）
//
// 响应示例：
// {
//   "agent": {
//     "id": "researcher-001",
//     "name": "Researcher",
//     "type": "researcher",
//     "capabilities": ["web_search", "information_collection"],
//     "status": "idle",
//     "metadata": {...}
//   }
// }
func (h *AgentHandler) GetAgent(c *gin.Context) {
	// 获取Agent ID
	agentID := c.Param("id")

	// 从注册表获取Agent信息
	agent, err := h.agentRegistry.Get(agentID)
	if err != nil {
		// Agent不存在
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
			"id":    agentID,
		})
		return
	}

	// 返回Agent详细信息
	c.JSON(http.StatusOK, gin.H{
		"agent": agent,
	})
}

// GetAgentCapabilities 获取Agent的能力列表
// 参数：
//   - id: Agent的ID（路径参数）
//
// 响应示例：
// {
//   "agent_id": "researcher-001",
//   "capabilities": ["web_search", "information_collection", "fact_checking"],
//   "total": 6
// }
func (h *AgentHandler) GetAgentCapabilities(c *gin.Context) {
	// 获取Agent ID
	agentID := c.Param("id")

	// 从注册表获取Agent信息
	agent, err := h.agentRegistry.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
			"id":    agentID,
		})
		return
	}

	// 返回Agent能力列表
	c.JSON(http.StatusOK, gin.H{
		"agent_id":     agentID,
		"capabilities": agent.Capabilities,
		"total":        len(agent.Capabilities),
	})
}

// GetAgentStatus 获取Agent的当前状态
// 参数：
//   - id: Agent的ID（路径参数）
//
// 响应示例：
// {
//   "agent_id": "researcher-001",
//   "status": "idle",
//   "last_heartbeat": "2024-01-28T14:30:00Z",
//   "healthy": true
// }
func (h *AgentHandler) GetAgentStatus(c *gin.Context) {
	// 获取Agent ID
	agentID := c.Param("id")

	// 从注册表获取Agent信息
	agent, err := h.agentRegistry.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
			"id":    agentID,
		})
		return
	}

	// 检查Agent健康状态
	isHealthy := h.agentRegistry.CheckHealth(agentID)

	// 返回Agent状态
	c.JSON(http.StatusOK, gin.H{
		"agent_id":       agentID,
		"status":         agent.Status,
		"last_heartbeat": agent.LastHeartbeat,
		"healthy":        isHealthy,
	})
}

// UpdateAgentHeartbeat 更新Agent心跳
// 用于Agent保活，防止被判定为不活跃
// 参数：
//   - id: Agent的ID（路径参数）
//
// 响应示例：
// {
//   "agent_id": "researcher-001",
//   "status": "success",
//   "timestamp": "2024-01-28T14:30:00Z"
// }
func (h *AgentHandler) UpdateAgentHeartbeat(c *gin.Context) {
	// 获取Agent ID
	agentID := c.Param("id")

	// 更新心跳时间
	err := h.agentRegistry.UpdateHeartbeat(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to update heartbeat",
			"id":    agentID,
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"agent_id":  agentID,
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ExecuteTask 创建并执行新任务
// 请求体示例：
// {
//   "type": "researcher",
//   "goal": "搜索关于AI的最新发展",
//   "priority": 1,
//   "requirements": {
//     "keywords": ["AI", "人工智能"],
//     "max_results": 10
//   }
// }
//
// 响应示例：
// {
//   "task_id": "task-001",
//   "status": "running",
//   "agent": "Researcher",
//   "started_at": "2024-01-28T14:30:00Z"
// }
func (h *AgentHandler) ExecuteTask(c *gin.Context) {
	// 解析请求体
	var req struct {
		Type         string                 `json:"type" binding:"required"`         // Agent类型
		Goal         string                 `json:"goal" binding:"required"`         // 任务目标
		Priority     int                    `json:"priority"`                        // 任务优先级（0-3）
		Requirements map[string]interface{} `json:"requirements"`                    // 任务要求
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 根据类型创建Agent
	agent, err := h.agentFactory.CreateAgent(req.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid agent type",
			"type":  req.Type,
		})
		return
	}

	// 创建任务对象
	task := &aiagenttask.Task{
		ID:           generateTaskID(),     // 生成唯一任务ID
		Type:         req.Type,
		Goal:         req.Goal,
		Requirements: req.Requirements,
		Priority:     aiagenttask.TaskPriority(req.Priority),
		Status:       aiagenttask.TaskStatusPending,
		CreatedAt:    time.Now(),
	}

	// 在后台执行任务
	go func() {
		ctx := context.Background()
		_, _ = agent.Execute(ctx, task)
	}()

	// 返回任务信息
	c.JSON(http.StatusAccepted, gin.H{
		"task_id":    task.ID,
		"status":     task.Status,
		"agent":      agent.GetInfo().Name,
		"started_at": time.Now().Format(time.RFC3339),
	})
}

// GetTaskStatus 获取任务执行状态
// 参数：
//   - id: 任务ID（路径参数）
//
// 响应示例：
// {
//   "task_id": "task-001",
//   "status": "completed",
//   "result": {...},
//   "duration": "2.5s"
// }
func (h *AgentHandler) GetTaskStatus(c *gin.Context) {
	// 获取任务ID
	taskID := c.Param("id")

	// TODO: 从状态管理器获取任务状态
	// 当前版本简化实现，返回任务ID
	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "Task status tracking will be implemented in next version",
	})
}

// ExecuteBatchTasks 批量执行任务
// 请求体示例：
// {
//   "tasks": [
//     {"type": "researcher", "goal": "搜索AI信息"},
//     {"type": "analyst", "goal": "分析数据"}
//   ]
// }
//
// 响应示例：
// {
//   "batch_id": "batch-001",
//   "tasks": [
//     {"task_id": "task-001", "status": "running"},
//     {"task_id": "task-002", "status": "running"}
//   ],
//   "total": 2
// }
func (h *AgentHandler) ExecuteBatchTasks(c *gin.Context) {
	// 解析请求体
	var req struct {
		Tasks []struct {
			Type         string                 `json:"type" binding:"required"`
			Goal         string                 `json:"goal" binding:"required"`
			Priority     int                    `json:"priority"`
			Requirements map[string]interface{} `json:"requirements"`
		} `json:"tasks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 生成批次ID
	batchID := generateBatchID()

	// 处理每个任务
	taskResponses := make([]gin.H, 0, len(req.Tasks))
	for _, taskReq := range req.Tasks {
		// 创建Agent
		agent, err := h.agentFactory.CreateAgent(taskReq.Type)
		if err != nil {
			taskResponses = append(taskResponses, gin.H{
				"error": "Invalid agent type",
				"type":  taskReq.Type,
			})
			continue
		}

		// 创建任务
		task := &aiagenttask.Task{
			ID:           generateTaskID(),
			Type:         taskReq.Type,
			Goal:         taskReq.Goal,
			Requirements: taskReq.Requirements,
			Priority:     aiagenttask.TaskPriority(taskReq.Priority),
			Status:       aiagenttask.TaskStatusPending,
			CreatedAt:    time.Now(),
		}

		// 在后台执行任务
		go func(t *aiagenttask.Task) {
			ctx := context.Background()
			_, _ = agent.Execute(ctx, t)
		}(task)

		taskResponses = append(taskResponses, gin.H{
			"task_id": task.ID,
			"status":  task.Status,
		})
	}

	// 返回批次信息
	c.JSON(http.StatusAccepted, gin.H{
		"batch_id": batchID,
		"tasks":    taskResponses,
		"total":    len(req.Tasks),
	})
}

// CreateWorkflow 创建新工作流
// 请求体示例：
// {
//   "name": "研究工作流",
//   "definition": {...}
// }
func (h *AgentHandler) CreateWorkflow(c *gin.Context) {
	// 解析请求体
	var req struct {
		Name       string                 `json:"name" binding:"required"`
		Definition map[string]interface{} `json:"definition" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// TODO: 实现工作流创建逻辑
	c.JSON(http.StatusCreated, gin.H{
		"workflow_id": generateWorkflowID(),
		"name":        req.Name,
		"status":      "created",
	})
}

// ListWorkflows 获取所有工作流列表
func (h *AgentHandler) ListWorkflows(c *gin.Context) {
	// TODO: 从状态管理器获取工作流列表
	c.JSON(http.StatusOK, gin.H{
		"workflows": []gin.H{},
		"total":     0,
	})
}

// GetWorkflow 获取工作流详情
func (h *AgentHandler) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	// TODO: 从状态管理器获取工作流详情
	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"message":     "Workflow details will be implemented",
	})
}

// ExecuteWorkflow 执行工作流
// 请求体示例：
// {
//   "inputs": {
//     "topic": "AI技术",
//     "max_results": 10
//   }
// }
func (h *AgentHandler) ExecuteWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	// 解析输入参数
	var req struct {
		Inputs map[string]interface{} `json:"inputs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// TODO: 执行工作流逻辑
	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": generateExecutionID(),
		"workflow_id":  workflowID,
		"status":       "running",
	})
}

// GetWorkflowExecutions 获取工作流执行历史
func (h *AgentHandler) GetWorkflowExecutions(c *gin.Context) {
	workflowID := c.Param("id")

	// TODO: 获取执行历史
	c.JSON(http.StatusOK, gin.H{
		"workflow_id":  workflowID,
		"executions":   []gin.H{},
		"total":        0,
	})
}

// DeleteWorkflow 删除工作流
func (h *AgentHandler) DeleteWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	// TODO: 删除工作流逻辑
	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"status":      "deleted",
	})
}

// PerformSearch 执行网络搜索
// 使用Researcher Agent进行信息搜索
// 请求体示例：
// {
//   "query": "人工智能最新发展",
//   "max_results": 10,
//   "time_range": "最近一周"
// }
func (h *AgentHandler) PerformSearch(c *gin.Context) {
	// 解析请求体
	var req struct {
		Query      string                 `json:"query" binding:"required"`      // 搜索查询
		MaxResults int                    `json:"max_results"`                  // 最大结果数
		TimeRange  string                 `json:"time_range"`                   // 时间范围
		Keywords   []string               `json:"keywords"`                     // 关键词
		Options    map[string]interface{} `json:"options"`                      // 额外选项
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 创建Researcher Agent
	researcher, err := h.agentFactory.CreateAgent("researcher")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create researcher agent",
		})
		return
	}

	// 准备任务要求
	requirements := map[string]interface{}{
		"query":       req.Query,
		"max_results": req.MaxResults,
		"time_range":  req.TimeRange,
		"keywords":    req.Keywords,
	}

	// 创建任务
	task := &aiagenttask.Task{
		ID:           generateTaskID(),
		Type:         "researcher",
		Goal:         req.Query,
		Requirements: requirements,
		Priority:     aiagenttask.PriorityNormal,
		Status:       aiagenttask.TaskStatusPending,
		CreatedAt:    time.Now(),
	}

	// 执行搜索
	ctx := context.Background()
	result, err := researcher.Execute(ctx, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	// 返回搜索结果
	c.JSON(http.StatusOK, gin.H{
		"task_id": result.TaskID,
		"query":   req.Query,
		"result":  result.Output,
		"status":  result.Status,
		"agent":   result.AgentUsed,
	})
}

// PerformAnalysis 执行数据分析
// 使用Analyst Agent进行数据分析
// 请求体示例：
// {
//   "analysis_type": "statistical",
//   "data": [10, 20, 30, 40, 50],
//   "options": {
//     "generate_charts": true
//   }
// }
func (h *AgentHandler) PerformAnalysis(c *gin.Context) {
	// 解析请求体
	var req struct {
		AnalysisType string                 `json:"analysis_type" binding:"required"` // 分析类型
		Data         interface{}            `json:"data" binding:"required"`          // 待分析数据
		Options      map[string]interface{} `json:"options"`                          // 分析选项
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 创建Analyst Agent
	analyst, err := h.agentFactory.CreateAgent("analyst")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create analyst agent",
		})
		return
	}

	// 准备任务要求
	requirements := map[string]interface{}{
		"data":         req.Data,
		"analysis_type": req.AnalysisType,
	}
	if req.Options != nil {
		for k, v := range req.Options {
			requirements[k] = v
		}
	}

	// 创建任务
	goal := "执行数据分析"
	if req.AnalysisType == "statistical" {
		goal = "分析数据的统计特征"
	} else if req.AnalysisType == "trend" {
		goal = "分析数据趋势"
	}

	task := &aiagenttask.Task{
		ID:           generateTaskID(),
		Type:         "analyst",
		Goal:         goal,
		Requirements: requirements,
		Priority:     aiagenttask.PriorityNormal,
		Status:       aiagenttask.TaskStatusPending,
		CreatedAt:    time.Now(),
	}

	// 执行分析
	ctx := context.Background()
	result, err := analyst.Execute(ctx, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"details": err.Error(),
		})
		return
	}

	// 返回分析结果
	c.JSON(http.StatusOK, gin.H{
		"task_id":       result.TaskID,
		"analysis_type": req.AnalysisType,
		"result":        result.Output,
		"status":        result.Status,
		"agent":         result.AgentUsed,
	})
}

// PerformWriting 执行内容生成
// 使用Writer Agent进行内容创作
// 请求体示例：
// {
//   "content_type": "article",
//   "topic": "人工智能技术",
//   "style": "formal",
//   "length": 1000
// }
func (h *AgentHandler) PerformWriting(c *gin.Context) {
	// 解析请求体
	var req struct {
		ContentType  string                 `json:"content_type" binding:"required"` // 内容类型
		Topic        string                 `json:"topic" binding:"required"`         // 主题
		Style        string                 `json:"style"`                            // 写作风格
		Length       int                    `json:"length"`                           // 内容长度
		Keywords     []string               `json:"keywords"`                         // 关键词
		Options      map[string]interface{} `json:"options"`                          // 额外选项
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 创建Writer Agent
	writer, err := h.agentFactory.CreateAgent("writer")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create writer agent",
		})
		return
	}

	// 准备任务要求
	requirements := map[string]interface{}{
		"style":  req.Style,
		"length": req.Length,
		"keywords": req.Keywords,
	}
	if req.Options != nil {
		for k, v := range req.Options {
			requirements[k] = v
		}
	}

	// 根据内容类型生成目标描述
	goal := "撰写文章"
	if req.ContentType == "report" {
		goal = "撰写报告"
	} else if req.ContentType == "summary" {
		goal = "生成摘要"
	}

	// 创建任务
	task := &aiagenttask.Task{
		ID:           generateTaskID(),
		Type:         "writer",
		Goal:         goal + "：" + req.Topic,
		Requirements: requirements,
		Priority:     aiagenttask.PriorityNormal,
		Status:       aiagenttask.TaskStatusPending,
		CreatedAt:    time.Now(),
	}

	// 执行写作
	ctx := context.Background()
	result, err := writer.Execute(ctx, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Writing failed",
			"details": err.Error(),
		})
		return
	}

	// 返回生成的内容
	c.JSON(http.StatusOK, gin.H{
		"task_id":      result.TaskID,
		"content_type": req.ContentType,
		"topic":        req.Topic,
		"result":       result.Output,
		"status":       result.Status,
		"agent":        result.AgentUsed,
		"word_count":   result.Metadata["word_count"],
	})
}

// GenerateReport 生成综合报告
// 协调多个Agent生成综合分析报告
// 请求体示例：
// {
//   "topic": "AI技术发展",
//   "sections": ["研究", "分析", "总结"]
// }
func (h *AgentHandler) GenerateReport(c *gin.Context) {
	// 解析请求体
	var req struct {
		Topic    string                 `json:"topic" binding:"required"`    // 报告主题
		Sections []string               `json:"sections"`                    // 报告章节
		Options  map[string]interface{} `json:"options"`                     // 报告选项
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 生成报告ID
	reportID := generateReportID()

	// 在后台生成报告（耗时操作）
	go func() {
		// TODO: 实现多Agent协作生成报告
		// 1. 使用Researcher收集信息
		// 2. 使用Analyst分析数据
		// 3. 使用Writer生成最终报告
	}()

	// 返回报告生成任务
	c.JSON(http.StatusAccepted, gin.H{
		"report_id": reportID,
		"topic":     req.Topic,
		"status":    "generating",
		"message":   "Report is being generated in background",
	})
}

// 辅助函数：生成唯一ID

// generateTaskID 生成唯一的任务ID
// 格式：task-时间戳-随机数
func generateTaskID() string {
	return fmt.Sprintf("task-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%1000)
}

// generateBatchID 生成唯一的批次ID
// 格式：batch-时间戳-随机数
func generateBatchID() string {
	return fmt.Sprintf("batch-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%1000)
}

// generateWorkflowID 生成唯一的工作流ID
// 格式：workflow-时间戳-随机数
func generateWorkflowID() string {
	return fmt.Sprintf("workflow-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%1000)
}

// generateExecutionID 生成唯一的执行ID
// 格式：exec-时间戳-随机数
func generateExecutionID() string {
	return fmt.Sprintf("exec-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%1000)
}

// generateReportID 生成唯一的报告ID
// 格式：report-时间戳-随机数
func generateReportID() string {
	return fmt.Sprintf("report-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%1000)
}

// ============================================================
// 工具相关API处理函数
// ============================================================

// ListTools 获取所有可用工具列表
// GET /api/v1/tools
func (h *AgentHandler) ListTools(c *gin.Context) {
	tools := h.toolManager.GetAvailableTools()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取工具列表成功",
		"data": gin.H{
			"tools": tools,
			"count": len(tools),
		},
	})
}

// GetToolInfo 获取指定工具的详细信息
// GET /api/v1/tools/:name
func (h *AgentHandler) GetToolInfo(c *gin.Context) {
	toolName := c.Param("name")

	info, err := h.toolManager.GetToolCapabilities(toolName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("工具不存在: %s", toolName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取工具信息成功",
		"data":    info,
	})
}

// GetToolCapabilities 获取工具的能力描述
// GET /api/v1/tools/:name/capabilities
func (h *AgentHandler) GetToolCapabilities(c *gin.Context) {
	toolName := c.Param("name")

	capabilities, err := h.toolManager.GetToolCapabilities(toolName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("工具不存在: %s", toolName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取工具能力成功",
		"data":    capabilities,
	})
}

// ExecuteTool 执行工具操作
// POST /api/v1/tools/execute
//
// 请求体示例：
// {
//   "tool_name": "file_ops",
//   "operation": "read",
//   "params": {
//     "path": "/tmp/test.txt"
//   }
// }
func (h *AgentHandler) ExecuteTool(c *gin.Context) {
	var req struct {
		ToolName  string                 `json:"tool_name" binding:"required"`  // 工具名称
		Operation  string                 `json:"operation" binding:"required"`  // 操作类型
		Params     map[string]interface{} `json:"params"`                        // 操作参数
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 执行工具
	ctx := context.Background()
	result, err := h.toolManager.ExecuteTool(ctx, req.ToolName, req.Operation, req.Params)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "工具执行失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工具执行成功",
		"data":    result,
	})
}

// BatchExecuteTools 批量执行工具
// POST /api/v1/tools/batch
//
// 请求体示例：
// {
//   "calls": [
//     {"tool_name": "file_ops", "operation": "read", "params": {"path": "/tmp/file1.txt"}},
//     {"tool_name": "file_ops", "operation": "read", "params": {"path": "/tmp/file2.txt"}}
//   ],
//   "concurrency": 3
// }
func (h *AgentHandler) BatchExecuteTools(c *gin.Context) {
	var req struct {
		Calls      []aitools.ToolCall `json:"calls" binding:"required"`     // 工具调用列表
		Concurrency int               `json:"concurrency"`                  // 并发数（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 创建工具集成
	toolIntegration := aitools.NewAgentToolIntegration("batch_handler", h.toolManager)

	// 批量执行
	ctx := context.Background()
	results, err := toolIntegration.BatchCallTools(ctx, req.Calls)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量工具执行失败",
			"details": err.Error(),
		})
		return
	}

	// 统计结果
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("批量执行完成：%d 成功，%d 失败", successCount, failureCount),
		"data": gin.H{
			"results":      results,
			"total":        len(results),
			"success":      successCount,
			"failed":       failureCount,
		},
	})
}

// ListToolChains 获取所有工具链
// GET /api/v1/tools/chains
func (h *AgentHandler) ListToolChains(c *gin.Context) {
	// 创建预定义的工具链
	chains := aitools.CreateToolChains(h.toolManager)

	chainList := make([]gin.H, 0)
	for name, chain := range chains {
		chainList = append(chainList, gin.H{
			"name":  name,
			"steps": len(chain.GetSteps()),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取工具链列表成功",
		"data": gin.H{
			"chains": chainList,
			"count":  len(chainList),
		},
	})
}

// ExecuteToolChain 执行工具链
// POST /api/v1/tools/chains/:name/execute
//
// 请求体示例：
// {
//   "input": {...}  // 初始输入数据（可选）
// }
func (h *AgentHandler) ExecuteToolChain(c *gin.Context) {
	chainName := c.Param("name")

	var req struct {
		Input interface{} `json:"input"` // 初始输入数据
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用空输入
		req.Input = nil
	}

	// 创建工具链执行器
	executor := aitools.NewToolChainExecutor(h.toolManager)

	// 注册预定义的工具链
	chains := aitools.CreateToolChains(h.toolManager)
	for _, chain := range chains {
		executor.RegisterChain(chain)
	}

	// 执行工具链
	ctx := context.Background()
	result, err := executor.ExecuteChain(ctx, chainName, req.Input)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "工具链执行失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工具链执行成功",
		"data": gin.H{
			"chain_name": chainName,
			"result":     result,
		},
	})
}
