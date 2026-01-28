package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	aiagentconfig "ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/handler"
	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/internal/memory"
	"ai-agent-assistant/internal/monitoring"
	"ai-agent-assistant/internal/tracing"
	aiagentrag "ai-agent-assistant/internal/rag"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := aiagentconfig.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化追踪系统（如果启用）
	if cfg.Monitoring.Tracing.Enabled {
		_, err = tracing.InitTracer(
			"ai-agent-assistant",
			cfg.Monitoring.Tracing.JaegerEndpoint,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize tracing: %v", err)
		}
	}

	// 3. 创建模型管理器
	modelManager, err := llm.NewModelManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create model manager: %v", err)
	}

	// 4. 创建监控服务器
	var monitoringServer *monitoring.Server
	if cfg.Monitoring.Enabled {
		metrics := monitoring.NewMetrics()
		monitoringServer = monitoring.NewServer(metrics, cfg.Monitoring.Prometheus.Port)

		if err := monitoringServer.Start(); err != nil {
			log.Printf("Warning: Failed to start monitoring server: %v", err)
		} else {
			log.Printf("Monitoring server started on :%d", cfg.Monitoring.Prometheus.Port)
		}
	}

	// 5. 创建增强版RAG系统
	ragSystem, err := aiagentrag.NewRAGEnhanced(cfg, modelManager)
	if err != nil {
		log.Fatalf("Failed to create enhanced RAG system: %v", err)
	}


	// 6. 创建增强版会话管理器
	// 获取embedding模型
	embeddingModel, _ := modelManager.GetModel(cfg.Agent.EmbeddingModel)
	sessionManager := memory.NewEnhancedSessionManager(
		cfg.Memory.MaxHistory,
		cfg.Memory.StoreType,
		embeddingModel,
	)

	sessionManager.EnableAutoSummary(true)
	sessionManager.SetSummaryThreshold(cfg.Memory.MaxHistory)

	// 7. 创建增强版记忆管理器
	memoryManager := memory.NewEnhancedMemoryManager(embeddingModel)
	memoryManager.EnableAutoExtract(true)
	memoryManager.EnableSemanticSearch(true)
	memoryManager.SetOptimizationStrategy("importance")

	// 8. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 9. 创建路由
	router := setupRouter(cfg, modelManager, ragSystem, sessionManager, memoryManager)

	// 10. 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// 打印启动信息
	printStartupInfo(cfg)

	// 优雅关闭
	setupGracefulShutdown(monitoringServer)

	// 启动HTTP服务器
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter 设置路由
func setupRouter(
	cfg *aiagentconfig.Config,
	modelManager *llm.ModelManager,
	ragSystem *aiagentrag.RAGEnhanced,
	sessionManager *memory.EnhancedSessionManager,
	memoryManager *memory.EnhancedMemoryManager,
) *gin.Engine {
	router := gin.Default()

	// API v1 路由
	api := router.Group("/api/v1")
	{
		// === 对话接口 ===
		api.POST("/chat", func(c *gin.Context) {
			handler.HandleChat(c, cfg, modelManager, sessionManager)
		})

		api.POST("/chat/rag", func(c *gin.Context) {
			handleChatWithRAG(c, cfg, modelManager, ragSystem, sessionManager)
		})

		// === 推理接口 ===
		api.POST("/reasoning/cot", func(c *gin.Context) {
			handleChainOfThought(c, modelManager)
		})

		api.POST("/reasoning/reflect", func(c *gin.Context) {
			handleReflection(c, modelManager)
		})

		// === 会话管理 ===
		api.GET("/session", func(c *gin.Context) {
			handleGetSession(c, sessionManager)
		})

		api.DELETE("/session", func(c *gin.Context) {
			handleClearSession(c, sessionManager)
		})

		api.POST("/session/state", func(c *gin.Context) {
			handleUpdateState(c, sessionManager)
		})

		// === 记忆管理 ===
		api.POST("/memory/extract", func(c *gin.Context) {
			handleExtractMemory(c, memoryManager)
		})

		api.GET("/memory/search", func(c *gin.Context) {
			handleSearchMemory(c, memoryManager)
		})

		// === 知识库管理 ===
		knowledge := api.Group("/knowledge")
		{
			knowledge.POST("/add", func(c *gin.Context) {
				handleAddKnowledge(c, ragSystem)
			})

			knowledge.POST("/add/doc", func(c *gin.Context) {
				handleAddKnowledgeFromDoc(c, ragSystem)
			})

			knowledge.GET("/stats", func(c *gin.Context) {
				handleGetKnowledgeStats(c, ragSystem)
			})

			knowledge.POST("/search", func(c *gin.Context) {
				handleSearchKnowledge(c, ragSystem)
			})
		}

		// === 评估接口 ===
		api.POST("/eval/accuracy", func(c *gin.Context) {
			handleEvaluation(c, modelManager)
		})

		// === 模型管理接口 ===
		api.GET("/models", func(c *gin.Context) {
			handleListModels(c, modelManager)
		})

		api.GET("/models/:name", func(c *gin.Context) {
			handleGetModelInfo(c, modelManager)
		})
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
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

// 打印启动信息
func printStartupInfo(cfg *aiagentconfig.Config) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println(" 🚀 AI Agent Assistant v0.4")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf(" Server: http://0.0.0.0:%d\n", cfg.Server.Port)
	fmt.Printf(" Mode: %s\n", cfg.Server.Mode)
	fmt.Printf(" Default Model: %s\n", cfg.Agent.DefaultModel)
	fmt.Printf(" Embedding Model: %s\n", cfg.Agent.EmbeddingModel)
	fmt.Printf(" RAG: %s\n", getBoolStatus(cfg.RAG.Enabled))
	fmt.Printf(" Memory: %s\n", cfg.Memory.StoreType)
	fmt.Printf(" Monitoring: %s\n", getBoolStatus(cfg.Monitoring.Enabled))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(" 🎯 New Features:")
	fmt.Println("   ✅ Multi-Model Support (GLM, Qwen, OpenAI, Claude, DeepSeek)")
	fmt.Println("   ✅ Enhanced RAG (Semantic Chunking, Hybrid Search, Rerank)")
	fmt.Println("   ✅ Reasoning Capability (Chain-of-Thought, Self-Reflection)")
	fmt.Println("   ✅ Auto Memory Extraction & Semantic Search")
	fmt.Println("   ✅ Auto Session Summary & State Management")
	fmt.Println("   ✅ Evaluation & Monitoring System")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
}

// 优雅关闭
func setupGracefulShutdown(monitoringServer *monitoring.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		if monitoringServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = monitoringServer.Stop(ctx)
			cancel()
		}

		os.Exit(0)
	}()
}

func getBoolStatus(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}
