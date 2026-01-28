package main

import (
	"fmt"
	"log"

	aiagentconfig "ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/internal/rag"
	"ai-agent-assistant/internal/memory"
	"ai-agent-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := aiagentconfig.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("\n🚀 AI Agent Assistant v0.4 - 测试服务器")
	fmt.Println("=====================================\n")

	// 2. 创建模型管理器
	modelManager, err := llm.NewModelManager(cfg)
	if err != nil {
		log.Printf("Warning: Failed to create model manager: %v", err)
	} else {
		fmt.Printf("✅ Model Manager created\n")
		fmt.Printf("   Supported models: %d\n", len(modelManager.ListModels()))
	}

	// 3. 创建RAG系统
	ragSystem, err := rag.NewRAG(cfg)
	if err != nil {
		log.Printf("Warning: Failed to create RAG: %v", err)
	} else {
		fmt.Printf("✅ RAG System created\n")
	}

	// 4. 创建会话管理器
	embeddingModel, _ := modelManager.GetModel(cfg.Agent.EmbeddingModel)
	sessionManager := memory.NewEnhancedSessionManager(
		cfg.Memory.MaxHistory,
		cfg.Memory.StoreType,
		embeddingModel,
	)
	fmt.Printf("✅ Session Manager created\n")

	// 5. 设置Gin路由
	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

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

	// 模型管理API
	router.GET("/api/v1/models", func(c *gin.Context) {
		factory := llm.NewModelFactory()
		supportedModels := factory.GetSupportedModels()
		loadedModels := modelManager.ListModels()

		c.JSON(200, gin.H{
			"supported_models": supportedModels,
			"loaded_models":    loadedModels,
		})
	})

	// 基础对话API（简化版）
	router.POST("/api/v1/chat", func(c *gin.Context) {
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
		sessionManager.AddMessage(req.SessionID, models.Message{
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
		sessionManager.AddMessage(req.SessionID, models.Message{
			Role:    "assistant",
			Content: response,
		})

		c.JSON(200, gin.H{
			"response":  response,
			"model":     modelName,
			"session_id": req.SessionID,
		})
	})

	// RAG对话API
	router.POST("/api/v1/chat/rag", func(c *gin.Context) {
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
		results, err := ragSystem.Retrieve(ctx, req.Message, topK)
		if err != nil {
			c.JSON(500, gin.H{"error": "RAG retrieval failed"})
			return
		}

		// 构建增强消息
		context := "知识库内容：\n" + fmt.Sprintf("%v", results)
		messages := []models.Message{
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
	})

	// 获取会话
	router.GET("/api/v1/session", func(c *gin.Context) {
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
			"created_at": session.CreatedAt,
		})
	})

	// 清除会话
	router.DELETE("/api/v1/session", func(c *gin.Context) {
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
	})

	// 启动信息
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("\n✅ Server ready!\n")
	fmt.Printf("📍 Address: http://0.0.0.0%s\n", addr)
	fmt.Printf("🏥 Health Check: http://0.0.0.0%s/health\n", addr)
	fmt.Printf("🤖 Models API: http://0.0.0.0%s/api/v1/models\n\n", addr)
	fmt.Println("=====================================")

	// 启动服务器
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
