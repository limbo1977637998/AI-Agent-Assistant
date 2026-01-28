package main

import (
	"fmt"
	"log"

	aiagentconfig "ai-agent-assistant/internal/config"
	aiagentexpert "ai-agent-assistant/internal/agent/expert"
	aiagentorchestrator "ai-agent-assistant/internal/orchestrator"
	"ai-agent-assistant/internal/handler"
	aitools "ai-agent-assistant/internal/tools"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := aiagentconfig.Load("config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	fmt.Println("🚀 AI Agent Assistant v0.5")
	fmt.Println("========================================")

	// 初始化Agent系统
	agentRegistry := aiagentorchestrator.NewAgentRegistry()
	expertFactory := aiagentexpert.NewFactory()

	// 创建工具管理器并设置到工厂
	toolManager := aitools.NewToolManager(&aitools.ToolManagerConfig{
		AutoRegister: true,
	})
	expertFactory.SetToolManager(toolManager)

	expertFactory.RegisterAllAgents(agentRegistry)

	// 列出Agent
	agents := agentRegistry.List()
	fmt.Printf("✅ 已注册 %d 个Agent:\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("   - %s (%s): %d项能力\n", agent.Name, agent.Type, len(agent.Capabilities))
	}

	// 创建Agent Handler
	agentHandler := handler.NewAgentHandler(
		cfg,
		expertFactory,
		agentRegistry,
		nil, // scheduler
	)

	// 创建路由
	router := gin.Default()
	gin.SetMode(cfg.Server.Mode)

	// 注册路由
	api := router.Group("/api/v1")
	{
		// v0.5 新增API
		agentHandler.RegisterRoutes(api)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "v0.5",
			"agents":   len(agents),
			"message": "AI Agent Assistant v0.5 - Agent编排和工作流系统",
		})
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("\n🌐 服务器启动成功！\n")
	fmt.Printf("   地址: http://localhost%s\n", addr)
	fmt.Printf("   健康检查: http://localhost%s/health\n", addr)
	fmt.Printf("   Agent列表: http://localhost%s/api/v1/agents\n", addr)
	fmt.Printf("   网络搜索: http://localhost%s/api/v1/analysis/search\n", addr)
	fmt.Printf("   数据分析: http://localhost%s/api/v1/analysis/analyze\n", addr)
	fmt.Printf("   内容生成: http://localhost%s/api/v1/analysis/write\n", addr)
	fmt.Println("\n按 Ctrl+C 停止服务器")
	fmt.Println("========================================")

	if err := router.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
