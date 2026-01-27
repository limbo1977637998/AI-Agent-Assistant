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
		api.POST("/chat", h.Chat)
		api.GET("/session", h.GetSession)
		api.DELETE("/session", h.ClearSession)
	}

	// 健康检查
	router.GET("/health", h.Health)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting AI Agent Assistant on %s", addr)
	log.Printf("Model: %s", cfg.Agent.DefaultModel)
	log.Printf("Enabled tools: %v", cfg.Tools.Enabled)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
