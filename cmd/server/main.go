package main

import (
	"fmt"
	"log"

	"ai-agent-assistant/internal/agent"
	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Agent
	ag, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 创建Handler
	h := handler.NewHandler(ag)

	// 创建路由
	router := gin.Default()

	// 注册路由
	api := router.Group("/api/v1")
	{
		// 聊天接口
		api.POST("/chat", h.Chat)
		api.POST("/chat/rag", h.ChatWithRAG) // RAG增强对话

		// 会话管理
		api.GET("/session", h.GetSession)
		api.DELETE("/session", h.ClearSession)

		// 知识库管理
		knowledge := api.Group("/knowledge")
		{
			knowledge.POST("/add", h.AddKnowledge)             // 添加文本知识
			knowledge.POST("/add/doc", h.AddKnowledgeFromDoc)  // 从文档添加知识
			knowledge.GET("/stats", h.GetKnowledgeStats)       // 知识库统计
			knowledge.POST("/search", h.SearchKnowledge)       // 搜索知识库
		}
	}

	// 健康检查
	router.GET("/health", h.Health)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting AI Agent Assistant on %s", addr)
	log.Printf("Model: %s", cfg.Agent.DefaultModel)
	log.Printf("Enabled tools: %v", cfg.Tools.Enabled)
	log.Printf("RAG enabled: true (Knowledge Base Support)")
	log.Printf("Knowledge API: /api/v1/knowledge/*")
	log.Printf("RAG Chat: /api/v1/chat/rag")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
