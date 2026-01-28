package handler

import (
	"context"
	"strconv"

	aiagentconfig "ai-agent-assistant/internal/config"
	aiagenteval "ai-agent-assistant/internal/eval"
	aiagentllm "ai-agent-assistant/internal/llm"
	aiagentmemory "ai-agent-assistant/internal/memory"
	aiagentrag "ai-agent-assistant/internal/rag"
	aigentreasoning "ai-agent-assistant/internal/reasoning"
	"ai-agent-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

// EnhancedHandler 增强版Handler
type EnhancedHandler struct {
	config          *aiagentconfig.Config
	modelManager    *aiagentllm.ModelManager
	ragSystem       *aiagentrag.RAGEnhanced
	sessionManager  *aiagentmemory.EnhancedSessionManager
	memoryManager   *aiagentmemory.EnhancedMemoryManager
}

// NewEnhancedHandler 创建增强版Handler
func NewEnhancedHandler(
	cfg *aiagentconfig.Config,
	modelManager *aiagentllm.ModelManager,
	ragSystem *aiagentrag.RAGEnhanced,
	sessionManager *aiagentmemory.EnhancedSessionManager,
	memoryManager *aiagentmemory.EnhancedMemoryManager,
) *EnhancedHandler {
	return &EnhancedHandler{
		config:         cfg,
		modelManager:   modelManager,
		ragSystem:      ragSystem,
		sessionManager: sessionManager,
		memoryManager:  memoryManager,
	}
}

// handleChat 处理聊天请求
func HandleChat(c *gin.Context, cfg *aiagentconfig.Config, modelManager *aiagentllm.ModelManager, sessionManager *aiagentmemory.EnhancedSessionManager) {
	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
		Model     string `json:"model,omitempty"`
		WithTools bool   `json:"with_tools,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取模型
	modelName := req.Model
	if modelName == "" {
		modelName = cfg.Agent.DefaultModel
	}

	model, err := modelManager.GetModel(modelName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Model not available"})
		return
	}

	// 获取或创建会话
	_, _ = sessionManager.GetOrCreateSession(req.SessionID, modelName)

	// 添加用户消息
	sessionManager.AddMessage(req.SessionID, models.Message{
		Role:    "user",
		Content: req.Message,
	})

	// 获取历史
	history, _ := sessionManager.GetHistory(req.SessionID)

	// 调用模型
	ctx := context.Background()
	response, err := model.Chat(ctx, history)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 添加助手消息
	sessionManager.AddMessage(req.SessionID, models.Message{
		Role:    "assistant",
		Content: response,
	})

	c.JSON(200, gin.H{
		"response":  response,
		"model":     modelName,
		"session_id": req.SessionID,
	})
}

// handleChatWithRAG 处理RAG增强对话
func HandleChatWithRAG(c *gin.Context, cfg *aiagentconfig.Config, modelManager *aiagentllm.ModelManager, ragSystem *aiagentrag.RAGEnhanced, sessionManager *aiagentmemory.EnhancedSessionManager) {
	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
		TopK      int    `json:"top_k,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 3
	}

	// RAG检索
	ctx := context.Background()
	ragContext, err := ragSystem.BuildContext(ctx, req.Message, topK)
	if err != nil {
		c.JSON(500, gin.H{"error": "RAG retrieval failed"})
		return
	}

	// 构建增强消息
	messages := []models.Message{
		{Role: "system", Content: ragContext},
		{Role: "user", Content: req.Message},
	}

	// 调用模型
	model, _ := modelManager.GetModel(cfg.Agent.DefaultModel)
	response, err := model.Chat(ctx, messages)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"response":   response,
		"rag_used":   true,
		"session_id": req.SessionID,
	})
}

// handleChainOfThought 处理思维链推理
func HandleChainOfThought(c *gin.Context, modelManager *aiagentllm.ModelManager) {
	var req struct {
		Task string `json:"task"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取推理模型
	model, _ := modelManager.GetModel("deepseek-r1")
	if model == nil {
		// 回退到默认模型
		model, _ = modelManager.GetModel("qwen")
	}

	if model == nil {
		c.JSON(500, gin.H{"error": "No reasoning model available"})
		return
	}

	// 创建思维链推理器
	cot := aigentreasoning.NewChainOfThought(model, true)

	// 执行推理
	ctx := context.Background()
	reasoning, answer, err := cot.Reason(ctx, req.Task)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"reasoning": reasoning,
		"answer":    answer,
	})
}

// handleReflection 处理自我反思
func HandleReflection(c *gin.Context, modelManager *aiagentllm.ModelManager) {
	var req struct {
		Task              string   `json:"task"`
		PreviousAttempts []string `json:"previous_attempts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取模型
	model, _ := modelManager.GetModel("qwen")
	if model == nil {
		c.JSON(500, gin.H{"error": "No model available"})
		return
	}

	// 创建反思器
	reflection := aigentreasoning.NewReflection(model, 1)

	// 执行反思
	ctx := context.Background()
	reflectionText, improvedAnswer, err := reflection.Reflect(ctx, req.Task, req.PreviousAttempts)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"reflection":      reflectionText,
		"improved_answer": improvedAnswer,
	})
}

// handleGetSession 获取会话
func HandleGetSession(c *gin.Context, sessionManager *aiagentmemory.EnhancedSessionManager) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session_id is required"})
		return
	}

	session, err := sessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(200, gin.H{
		"session_id": session.ID,
		"model":      session.Model,
		"summary":    session.Summary,
		"state":      session.State,
		"created_at": session.CreatedAt,
		"updated_at": session.UpdatedAt,
	})
}

// handleClearSession 清除会话
func HandleClearSession(c *gin.Context, sessionManager *aiagentmemory.EnhancedSessionManager) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session_id is required"})
		return
	}

	if err := sessionManager.Clear(sessionID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Session cleared"})
}

// handleUpdateState 更新会话状态
func HandleUpdateState(c *gin.Context, sessionManager *aiagentmemory.EnhancedSessionManager) {
	var req struct {
		SessionID string                 `json:"session_id"`
		Updates   map[string]interface{} `json:"updates"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	version, err := sessionManager.UpdateState(req.SessionID, req.Updates)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "State updated",
		"version": version,
	})
}

// handleExtractMemory 提取记忆
func HandleExtractMemory(c *gin.Context, memoryManager *aiagentmemory.EnhancedMemoryManager) {
	var req struct {
		UserID      string `json:"user_id"`
		Conversation string `json:"conversation"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	memories, err := memoryManager.ExtractMemories(ctx, req.UserID, req.Conversation)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 添加到记忆存储
	for _, memory := range memories {
		memoryManager.AddMemory(ctx, memory)
	}

	c.JSON(200, gin.H{
		"message":  "Memories extracted",
		"count":     len(memories),
		"memories": memories,
	})
}

// handleSearchMemory 搜索记忆
func HandleSearchMemory(c *gin.Context, memoryManager *aiagentmemory.EnhancedMemoryManager) {
	userID := c.Query("user_id")
	query := c.Query("query")
	limit := c.DefaultQuery("limit", "10")

	limitInt, _ := strconv.Atoi(limit)

	if userID == "" || query == "" {
		c.JSON(400, gin.H{"error": "user_id and query are required"})
		return
	}

	ctx := context.Background()
	memories, err := memoryManager.SemanticSearch(ctx, userID, query, limitInt)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"query":    query,
		"count":    len(memories),
		"memories": memories,
	})
}

// handleAddKnowledge 添加知识
func HandleAddKnowledge(c *gin.Context, ragSystem *aiagentrag.RAGEnhanced) {
	var req struct {
		Text   string `json:"text"`
		Source string `json:"source"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := ragSystem.AddText(ctx, req.Text, req.Source); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Knowledge added successfully"})
}

// handleAddKnowledgeFromDoc 从文档添加知识
func HandleAddKnowledgeFromDoc(c *gin.Context, ragSystem *aiagentrag.RAGEnhanced) {
	var req struct {
		DocPath string `json:"doc_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := ragSystem.AddDocument(ctx, req.DocPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Document added successfully"})
}

// handleGetKnowledgeStats 获取知识库统计
func HandleGetKnowledgeStats(c *gin.Context, ragSystem *aiagentrag.RAGEnhanced) {
	stats := ragSystem.GetStats()

	c.JSON(200, gin.H{
		"stats": stats,
	})
}

// handleSearchKnowledge 搜索知识库
func HandleSearchKnowledge(c *gin.Context, ragSystem *aiagentrag.RAGEnhanced) {
	var req struct {
		Query string `json:"query"`
		TopK  int    `json:"top_k,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 3
	}

	ctx := context.Background()
	results, err := ragSystem.RetrieveEnhanced(ctx, req.Query, topK)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"query":    req.Query,
		"count":    len(results),
		"results":  results,
	})
}

// handleEvaluation 执行评估
func HandleEvaluation(c *gin.Context, modelManager *aiagentllm.ModelManager) {
	var req struct {
		TestCases []aiagenteval.TestCase `json:"test_cases"`
		Accuracy bool                       `json:"accuracy,omitempty"`
		Performance bool                   `json:"performance,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	model, _ := modelManager.GetModel("qwen")
	if model == nil {
		c.JSON(500, gin.H{"error": "No model available"})
		return
	}

	builder := aiagenteval.NewEvaluatorBuilder()

	if req.Accuracy || (!req.Accuracy && !req.Performance) {
		builder.WithAccuracy("similarity", model, 0.7)
	}

	if req.Performance || (!req.Accuracy && !req.Performance) {
		builder.WithPerformance(5)
	}

	manager := builder.Build()

	ctx := context.Background()
	results, err := manager.RunEvaluations(ctx, model, req.TestCases)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	report := manager.GenerateReport(results)

	c.JSON(200, gin.H{
		"results": results,
		"report":  report,
		"overall_score": manager.GetOverallScore(results),
	})
}

// handleListModels 列出可用模型
func HandleListModels(c *gin.Context, modelManager *aiagentllm.ModelManager) {
	factory := aiagentllm.NewModelFactory()

	supportedModels := factory.GetSupportedModels()
	loadedModels := modelManager.ListModels()

	c.JSON(200, gin.H{
		"supported_models": supportedModels,
		"loaded_models":    loadedModels,
	})
}

// handleGetModelInfo 获取模型信息
func HandleGetModelInfo(c *gin.Context, modelManager *aiagentllm.ModelManager) {
	modelName := c.Param("name")

	info := modelManager.GetModelInfo(modelName)

	if info == nil {
		c.JSON(404, gin.H{"error": "Model not found"})
		return
	}

	c.JSON(200, gin.H{
		"model": modelName,
		"info":  info,
	})
}
