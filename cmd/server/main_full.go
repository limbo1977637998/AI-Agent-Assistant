package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	aiagentconfig "ai-agent-assistant/internal/config"
	aiagenteval "ai-agent-assistant/internal/eval"
	llm "ai-agent-assistant/internal/llm"
	memory "ai-agent-assistant/internal/memory"
	aiagentrag "ai-agent-assistant/internal/rag"
	aigentreasoning "ai-agent-assistant/internal/reasoning"
	pkgmodels "ai-agent-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := aiagentconfig.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("\n🚀 AI Agent Assistant v0.4 - 完整版服务器")
	fmt.Println("========================================\n")

	// 2. 创建模型管理器
	modelManager, err := llm.NewModelManager(cfg)
	if err != nil {
		log.Printf("Warning: Failed to create model manager: %v", err)
	} else {
		fmt.Printf("✅ Model Manager created\n")
		fmt.Printf("   Loaded models: %v\n", modelManager.ListModels())
	}

	// 3. 创建RAG系统
	ragSystem, err := aiagentrag.NewRAG(cfg)
	if err != nil {
		log.Printf("Warning: Failed to create RAG: %v", err)
	} else {
		fmt.Printf("✅ RAG System created\n")
	}

	// 4. 创建会话管理器
	embeddingModel, _ := modelManager.GetModel(cfg.Agent.EmbeddingModel)
	sessionManager := memory.NewEnhancedSessionManager(
		cfg.Memory.MaxHistory,
		"memory", // 使用内存存储以避免数据库依赖
		embeddingModel,
	)
	sessionManager.EnableAutoSummary(true)
	sessionManager.SetSummaryThreshold(cfg.Memory.MaxHistory)
	fmt.Printf("✅ Session Manager created\n")

	// 5. 创建记忆管理器
	memoryManager := memory.NewEnhancedMemoryManager(embeddingModel)
	memoryManager.EnableAutoExtract(true)
	memoryManager.EnableSemanticSearch(true)
	memoryManager.SetOptimizationStrategy("importance")
	fmt.Printf("✅ Memory Manager created\n")

	// 6. 创建推理管理器
	var reasoningManager *aigentreasoning.ReasoningManager
	if cfg.Agent.DefaultModel != "" {
		reasoningModel, _ := modelManager.GetModel(cfg.Agent.DefaultModel)
		if reasoningModel != nil {
			reasoningManager = aigentreasoning.NewReasoningManager(reasoningModel, true, 3)
			fmt.Printf("✅ Reasoning Manager created\n")
		}
	}

	// 7. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 8. 创建路由
	router := setupRouter(cfg, modelManager, ragSystem, sessionManager, memoryManager, reasoningManager)

	// 9. 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// 打印启动信息
	printStartupInfo(cfg)

	// 优雅关闭
	setupGracefulShutdown()

	// 启动HTTP服务器
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter 设置路由
func setupRouter(
	cfg *aiagentconfig.Config,
	modelManager *llm.ModelManager,
	ragSystem *aiagentrag.RAG,
	sessionManager *memory.EnhancedSessionManager,
	memoryManager *memory.EnhancedMemoryManager,
	reasoningManager *aigentreasoning.ReasoningManager,
) *gin.Engine {
	router := gin.Default()

	// API v1 路由
	api := router.Group("/api/v1")
	{
		// === 对话接口 ===
		api.POST("/chat", handleChat(cfg, modelManager, sessionManager))
		api.POST("/chat/rag", handleChatWithRAG(cfg, modelManager, ragSystem, sessionManager))

		// === 推理接口 ===
		if reasoningManager != nil {
			api.POST("/reasoning/cot", handleChainOfThought(reasoningManager))
			api.POST("/reasoning/reflect", handleReflection(reasoningManager))
		}

		// === 会话管理 ===
		api.GET("/session", handleGetSession(sessionManager))
		api.DELETE("/session", handleClearSession(sessionManager))
		api.POST("/session/state", handleUpdateState(sessionManager))

		// === 记忆管理 ===
		api.POST("/memory/extract", handleExtractMemory(memoryManager))
		api.GET("/memory/search", handleSearchMemory(memoryManager))

		// === 知识库管理 ===
		api.POST("/knowledge/add", handleAddKnowledge(ragSystem))
		api.GET("/knowledge/stats", handleGetKnowledgeStats(ragSystem))
		api.POST("/knowledge/search", handleSearchKnowledge(ragSystem))

		// === 评估接口 ===
		api.POST("/eval/accuracy", handleEvaluation(modelManager))

		// === 模型管理接口 ===
		api.GET("/models", handleListModels(modelManager))
		api.GET("/models/:name", handleGetModelInfo(modelManager))
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"version": "v0.4",
			"features": []string{
				"Multi-Model Support",
				"Enhanced RAG",
				"Reasoning Capability",
				"Auto Memory Extraction",
				"Auto Session Summary",
				"Evaluation & Monitoring",
			},
		})
	})

	return router
}

// Handler函数

func handleChat(cfg *aiagentconfig.Config, modelManager *llm.ModelManager, sessionManager *memory.EnhancedSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
			Message   string `json:"message"`
			Model     string `json:"model,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

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
		sessionManager.AddMessage(req.SessionID, pkgmodels.Message{
			Role:    "user",
			Content: req.Message,
		})

		// 获取历史
		history, _ := sessionManager.GetHistory(req.SessionID)

		// 调用模型
		ctx := c.Request.Context()
		response, err := model.Chat(ctx, history)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 添加助手消息
		sessionManager.AddMessage(req.SessionID, pkgmodels.Message{
			Role:    "assistant",
			Content: response,
		})

		c.JSON(200, gin.H{
			"response":  response,
			"model":     modelName,
			"session_id": req.SessionID,
		})
	}
}

func handleChatWithRAG(cfg *aiagentconfig.Config, modelManager *llm.ModelManager, ragSystem *aiagentrag.RAG, sessionManager *memory.EnhancedSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		ctx := c.Request.Context()
		context, err := ragSystem.BuildContext(ctx, req.Message, topK)
		if err != nil {
			c.JSON(500, gin.H{"error": "RAG retrieval failed"})
			return
		}

		// 构建增强消息
		messages := []pkgmodels.Message{
			{Role: "system", Content: context},
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
}

func handleChainOfThought(reasoningManager *aigentreasoning.ReasoningManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Task string `json:"task"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 执行思维链推理
		ctx := c.Request.Context()
		reasoning, answer, err := reasoningManager.ReasonWithCoTAndReflection(ctx, req.Task)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"reasoning": reasoning,
			"answer":    answer,
		})
	}
}

func handleReflection(reasoningManager *aigentreasoning.ReasoningManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Task              string   `json:"task"`
			PreviousAttempts []string `json:"previous_attempts"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 执行反思（使用CoT + Reflection）
		ctx := c.Request.Context()
		reasoning, answer, err := reasoningManager.ReasonWithCoTAndReflection(ctx, req.Task)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"reflection":      reasoning,
			"improved_answer": answer,
		})
	}
}

func handleGetSession(sessionManager *memory.EnhancedSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

func handleClearSession(sessionManager *memory.EnhancedSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

func handleUpdateState(sessionManager *memory.EnhancedSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

func handleExtractMemory(memoryManager *memory.EnhancedMemoryManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID      string `json:"user_id"`
			Conversation string `json:"conversation"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
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
}

func handleSearchMemory(memoryManager *memory.EnhancedMemoryManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		query := c.Query("query")
		limit := c.DefaultQuery("limit", "10")

		limitInt, _ := strconv.Atoi(limit)

		if userID == "" || query == "" {
			c.JSON(400, gin.H{"error": "user_id and query are required"})
			return
		}

		ctx := c.Request.Context()
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
}

func handleAddKnowledge(ragSystem *aiagentrag.RAG) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Text   string `json:"text"`
			Source string `json:"source"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
		if err := ragSystem.AddText(ctx, req.Text, req.Source); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Knowledge added successfully"})
	}
}

func handleGetKnowledgeStats(ragSystem *aiagentrag.RAG) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := ragSystem.GetStats()

		c.JSON(200, gin.H{
			"stats": stats,
		})
	}
}

func handleSearchKnowledge(ragSystem *aiagentrag.RAG) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		ctx := c.Request.Context()
		results, err := ragSystem.Retrieve(ctx, req.Query, topK)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"query":   req.Query,
			"count":   len(results),
			"results": results,
		})
	}
}

func handleEvaluation(modelManager *llm.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TestCases  []aiagenteval.TestCase `json:"test_cases"`
			Accuracy   bool             `json:"accuracy,omitempty"`
			Performance bool             `json:"performance,omitempty"`
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

		ctx := c.Request.Context()
		results, err := manager.RunEvaluations(ctx, model, req.TestCases)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		report := manager.GenerateReport(results)

		c.JSON(200, gin.H{
			"results":       results,
			"report":        report,
			"overall_score": manager.GetOverallScore(results),
		})
	}
}

func handleListModels(modelManager *llm.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		factory := llm.NewModelFactory()

		supportedModels := factory.GetSupportedModels()
		loadedModels := modelManager.ListModels()

		c.JSON(200, gin.H{
			"supported_models": supportedModels,
			"loaded_models":    loadedModels,
		})
	}
}

func handleGetModelInfo(modelManager *llm.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

// 打印启动信息
func printStartupInfo(cfg *aiagentconfig.Config) {
	fmt.Printf("\n✅ 服务器就绪！\n")
	fmt.Printf("📍 地址: http://0.0.0.0:%d\n", cfg.Server.Port)
	fmt.Printf("🏥 健康检查: http://0.0.0.0:%d/health\n", cfg.Server.Port)
	fmt.Printf("🤖 模型API: http://0.0.0.0:%d/api/v1/models\n", cfg.Server.Port)
	fmt.Printf("💬 对话API: http://0.0.0.0:%d/api/v1/chat\n", cfg.Server.Port)
	fmt.Printf("🧠 RAG对话: http://0.0.0.0:%d/api/v1/chat/rag\n", cfg.Server.Port)
	fmt.Printf("🤔 推理API: http://0.0.0.0:%d/api/v1/reasoning/cot\n", cfg.Server.Port)
	fmt.Printf("💾 记忆API: http://0.0.0.0:%d/api/v1/memory/*\n", cfg.Server.Port)
	fmt.Printf("📚 知识库: http://0.0.0.0:%d/api/v1/knowledge/*\n", cfg.Server.Port)
	fmt.Printf("📊 评估系统: http://0.0.0.0:%d/api/v1/eval/*\n\n", cfg.Server.Port)
	fmt.Println("========================================")
	fmt.Println("🎯 v0.4 完整功能已启用！")
	fmt.Println("========================================\n")
}

// 优雅关闭
func setupGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		os.Exit(0)
	}()
}
