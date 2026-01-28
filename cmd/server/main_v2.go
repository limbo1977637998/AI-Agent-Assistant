package main

import (
	"fmt"
	"log"
	"strings"

	aiagentconfig "ai-agent-assistant/internal/config"
	aiagentexpert "ai-agent-assistant/internal/agent/expert"
	aiagentllm "ai-agent-assistant/internal/llm"
	aiagentmemory "ai-agent-assistant/internal/memory"
	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	aiagentrag "ai-agent-assistant/internal/rag"
	"ai-agent-assistant/internal/handler"

	"github.com/gin-gonic/gin"
)

// main 主函数
// 初始化所有组件并启动HTTP服务器
func main() {
	// ============================================================
	// 第一步：加载配置文件
	// ============================================================
	log.Println("📋 加载配置文件...")
	cfg, err := aiagentconfig.Load("config.yaml")
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}
	log.Printf("✅ 配置加载成功 - 模式: %s, 端口: %d", cfg.Server.Mode, cfg.Server.Port)

	// ============================================================
	// 第二步：初始化模型管理器
	// ============================================================
	log.Println("🤖 初始化模型管理器...")
	modelManager := aiagentllm.NewModelManager(cfg)

	// 加载配置的模型
	for _, modelConfig := range cfg.Models {
		if err := modelManager.LoadModel(modelConfig.Name, modelConfig.Provider, modelConfig.Model); err != nil {
			log.Printf("⚠️  警告: 无法加载模型 %s: %v", modelConfig.Name, err)
		} else {
			log.Printf("✅ 模型加载成功: %s (%s)", modelConfig.Name, modelConfig.Provider)
		}
	}

	// ============================================================
	// 第三步：初始化内存管理器
	// ============================================================
	log.Println("💾 初始化内存管理器...")
	memoryManager := aiagentmemory.NewEnhancedMemoryManager(
		cfg.Memory.MaxMessages,
		cfg.Memory.RetentionDays,
	)

	// ============================================================
	// 第四步：初始化会话管理器
	// ============================================================
	log.Println("🔄 初始化会话管理器...")
	sessionManager := aiagentmemory.NewEnhancedSessionManager(
		memoryManager,
		cfg.Session.Timeout,
	)

	// ============================================================
	// 第五步：初始化RAG系统
	// ============================================================
	log.Println("📚 初始化RAG系统...")
	ragSystem, err := aiagentrag.NewRAGEnhanced(cfg, modelManager)
	if err != nil {
		log.Printf("⚠️  警告: RAG系统初始化失败: %v", err)
		log.Println("💡 提示: RAG功能将不可用，但其他功能正常运行")
	} else {
		log.Println("✅ RAG系统初始化成功")
	}

	// ============================================================
	// 第六步：初始化Agent编排器
	// ============================================================
	log.Println("🎭 初始化Agent编排器...")

	// 创建Agent注册表
	agentRegistry := aiagentorchestrator.NewAgentRegistry()
	log.Println("✅ Agent注册表创建成功")

	// 创建任务调度器
	taskScheduler := aiagentorchestrator.NewTaskScheduler(agentRegistry, 10) // 10个并发worker
	log.Println("✅ 任务调度器创建成功")

	// 创建专家Agent工厂
	expertFactory := aiagentexpert.NewFactory()
	log.Println("✅ 专家Agent工厂创建成功")

	// 注册所有专家Agent到注册表
	err = expertFactory.RegisterAllAgents(agentRegistry)
	if err != nil {
		log.Fatalf("❌ Agent注册失败: %v", err)
	}
	log.Println("✅ 所有专家Agent注册成功")

	// 列出已注册的Agent
	agents := agentRegistry.GetAll()
	for _, agent := range agents {
		log.Printf("   📌 %s (%s) - %d项能力", agent.Name, agent.Type, len(agent.Capabilities))
	}

	// ============================================================
	// 第七步：创建HTTP处理器
	// ============================================================
	log.Println("🌐 创建HTTP处理器...")

	// 创建增强版Handler（兼容原有功能）
	enhancedHandler := handler.NewEnhancedHandler(
		cfg,
		modelManager,
		ragSystem,
		sessionManager,
		memoryManager,
	)

	// 创建Agent Handler（新增功能）
	agentHandler := handler.NewAgentHandler(
		cfg,
		expertFactory,
		agentRegistry,
		taskScheduler,
	)
	log.Println("✅ HTTP处理器创建成功")

	// ============================================================
	// 第八步：配置路由
	// ============================================================
	log.Println("🛣️  配置API路由...")

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由器
	router := gin.Default()

	// 添加恢复中间件（防止panic导致服务器崩溃）
	router.Use(gin.Recovery())

	// API v1 路由组
	api := router.Group("/api/v1")
	{
		// ========================================================
		// 原有功能：聊天和会话管理
		// ========================================================
		api.POST("/chat", func(c *gin.Context) {
			handler.HandleChat(c, cfg, modelManager, sessionManager)
		})
		api.POST("/chat/rag", func(c *gin.Context) {
			handler.HandleChatWithRAG(c, cfg, modelManager, ragSystem, sessionManager)
		})

		// 会话管理
		api.GET("/session", func(c *gin.Context) {
			handler.HandleGetSession(c, sessionManager)
		})
		api.DELETE("/session", func(c *gin.Context) {
			handler.HandleClearSession(c, sessionManager)
		})

		// ========================================================
		// 原有功能：知识库管理
		// ========================================================
		knowledge := api.Group("/knowledge")
		{
			knowledge.POST("/add", func(c *gin.Context) {
				handler.HandleAddKnowledge(c, cfg, ragSystem)
			})
			knowledge.POST("/add/doc", func(c *gin.Context) {
				handler.HandleAddKnowledgeFromDoc(c, cfg, ragSystem)
			})
			knowledge.GET("/stats", func(c *gin.Context) {
				handler.HandleGetKnowledgeStats(c, ragSystem)
			})
			knowledge.POST("/search", func(c *gin.Context) {
				handler.HandleSearchKnowledge(c, ragSystem)
			})
		}

		// ========================================================
		// 新增功能：Agent管理
		// ========================================================
		agentHandler.RegisterRoutes(api.Group("/")) // AgentHandler会自己创建子路由组

		// ========================================================
		// 新增功能：分析和研究（简化路由）
		// ========================================================
		analysis := api.Group("/analysis")
		{
			analysis.POST("/search", agentHandler.PerformSearch)      // 网络搜索
			analysis.POST("/analyze", agentHandler.PerformAnalysis)   // 数据分析
			analysis.POST("/write", agentHandler.PerformWriting)      // 内容生成
			analysis.POST("/report", agentHandler.GenerateReport)     // 生成报告
		}
	}

	// ============================================================
	// 第九步：健康检查端点
	// ============================================================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"version": "v0.5",
			"features": gin.H{
				"chat": true,
				"rag": true,
				"agents": true,
				"workflow": true,
			},
		})
	})

	// ============================================================
	// 第十步：启动HTTP服务器
	// ============================================================
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// 打印启动信息
	separator := strings.Repeat("=", 60)
	log.Println("\n" + separator)
	log.Println("🚀 AI Agent Assistant v0.5")
	log.Println(separator)
	log.Printf("🌐 服务器地址: http://localhost%s", addr)
	log.Printf("📖 API文档: http://localhost%s/api/v1", addr)
	log.Println("\n📋 可用功能:")
	log.Println("   • 聊天对话: POST /api/v1/chat")
	log.Println("   • RAG增强对话: POST /api/v1/chat/rag")
	log.Println("   • 会话管理: GET/DELETE /api/v1/session")
	log.Println("   • 知识库管理: /api/v1/knowledge/*")
	log.Println("\n🤖 Agent功能:")
	log.Println("   • Agent列表: GET /api/v1/agents")
	log.Println("   • Agent详情: GET /api/v1/agents/:id")
	log.Println("   • 执行任务: POST /api/v1/tasks")
	log.Println("\n📊 分析功能:")
	log.Println("   • 网络搜索: POST /api/v1/analysis/search")
	log.Println("   • 数据分析: POST /api/v1/analysis/analyze")
	log.Println("   • 内容生成: POST /api/v1/analysis/write")
	log.Println("   • 生成报告: POST /api/v1/analysis/report")
	log.Println(separator + "\n")

	// 启动服务器
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
