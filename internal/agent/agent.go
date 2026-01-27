package agent

import (
	"context"
	"fmt"
	"strings"

	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/internal/llm"
	aiagenthttp "ai-agent-assistant/pkg/http"
	"ai-agent-assistant/internal/memory"
	"ai-agent-assistant/internal/rag"
	"ai-agent-assistant/internal/tools"
	"ai-agent-assistant/pkg/models"
)

// Agent 智能体
type Agent struct {
	config      *config.Config
	llmManager  *llm.LLMManager
	memoryMgr   *memory.MemoryManager
	toolManager *tools.ToolManager
	rag         *rag.RAG
}

// NewAgent 创建智能体
func NewAgent(cfg *config.Config) (*Agent, error) {
	// 设置HTTP代理配置
	aiagenthttp.SetConfig(&cfg.Proxy)

	// 初始化LLM管理器
	llmMgr, err := llm.NewLLMManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM manager: %w", err)
	}

	// 初始化记忆管理器
	memoryMgr := memory.NewMemoryManager(cfg.Memory.MaxHistory, cfg.Memory.StoreType)

	// 初始化工具管理器
	toolMgr := tools.NewToolManager(cfg.Tools.Enabled)

	// 初始化RAG系统
	ragSystem, err := rag.NewRAG(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init RAG: %w", err)
	}

	return &Agent{
		config:      cfg,
		llmManager:  llmMgr,
		memoryMgr:   memoryMgr,
		toolManager: toolMgr,
		rag:         ragSystem,
	}, nil
}

// Chat 对话处理
func (a *Agent) Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 确定使用的模型
	modelName := req.Model
	if modelName == "" {
		modelName = a.config.Agent.DefaultModel
	}

	// 获取或创建会话
	_, err := a.memoryMgr.GetOrCreateSession(req.SessionID, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 添加用户消息
	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	if err := a.memoryMgr.AddMessage(req.SessionID, userMsg); err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	// 获取历史消息
	history, _ := a.memoryMgr.GetHistory(req.SessionID)

	// RAG增强：如果启用RAG，检索相关知识并增强用户消息
	if req.UseRAG {
		ragContext, err := a.rag.BuildContext(ctx, req.Message, 3) // 检索top-3
		if err == nil && ragContext != "" {
			// 创建增强的消息内容
			enhancedMsg := models.Message{
				Role:    "system",
				Content: ragContext,
			}
			// 将RAG上下文插入到历史记录的开头
			history = append([]models.Message{enhancedMsg}, history...)
		}
	}

	// 获取LLM提供商
	provider, err := a.llmManager.GetProvider(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// 调用LLM
	var response string
	if req.Stream {
		streamCh, err := provider.ChatStream(ctx, history)
		if err != nil {
			return nil, fmt.Errorf("stream chat failed: %w", err)
		}

		// 收集流式响应
		for chunk := range streamCh {
			response += chunk
		}
	} else {
		response, err = provider.Chat(ctx, history)
		if err != nil {
			return nil, fmt.Errorf("chat failed: %w", err)
		}
	}

	// 添加助手消息
	assistantMsg := models.Message{
		Role:    "assistant",
		Content: response,
	}
	if err := a.memoryMgr.AddMessage(req.SessionID, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to add assistant message: %w", err)
	}

	// 构建响应
	resp := &models.ChatResponse{
		SessionID: req.SessionID,
		Message:   response,
		Model:     modelName,
		Metadata:  req.Metadata,
	}

	return resp, nil
}

// ChatWithTools 带工具调用的对话处理
func (a *Agent) ChatWithTools(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 确定使用的模型
	modelName := req.Model
	if modelName == "" {
		modelName = a.config.Agent.DefaultModel
	}

	// 获取或创建会话
	_, err := a.memoryMgr.GetOrCreateSession(req.SessionID, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 添加用户消息
	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	if err := a.memoryMgr.AddMessage(req.SessionID, userMsg); err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	// 获取历史消息
	history, _ := a.memoryMgr.GetHistory(req.SessionID)

	// 获取LLM提供商
	provider, err := a.llmManager.GetProvider(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// 检测是否需要使用工具
	toolName, toolArgs := a.detectToolNeeds(req.Message)

	var toolCalls []models.ToolCall
	var response string

	if toolName != "" {
		// 执行工具
		tool, ok := a.toolManager.GetTool(toolName)
		if ok {
			result, err := tool.Execute(ctx, toolArgs)
			if err != nil {
				result = fmt.Sprintf("工具执行失败: %v", err)
			}

			toolCalls = append(toolCalls, models.ToolCall{
				ID:        toolName,
				Name:      toolName,
				Arguments: toolArgs,
				Result:    result,
			})

			// 添加工具调用消息到历史
			toolMsg := models.Message{
				Role:    "system",
				Content: fmt.Sprintf("工具[%s]执行结果: %s", toolName, result),
			}
			a.memoryMgr.AddMessage(req.SessionID, toolMsg)

			// 重新获取历史并调用LLM
			history, _ = a.memoryMgr.GetHistory(req.SessionID)
			response, err = provider.Chat(ctx, history)
			if err != nil {
				return nil, fmt.Errorf("chat after tool failed: %w", err)
			}
		}
	} else {
		// 直接调用LLM
		response, err = provider.Chat(ctx, history)
		if err != nil {
			return nil, fmt.Errorf("chat failed: %w", err)
		}
	}

	// 添加助手消息
	assistantMsg := models.Message{
		Role:    "assistant",
		Content: response,
	}
	if err := a.memoryMgr.AddMessage(req.SessionID, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to add assistant message: %w", err)
	}

	return &models.ChatResponse{
		SessionID: req.SessionID,
		Message:   response,
		Model:     modelName,
		ToolCalls: toolCalls,
		Metadata:  req.Metadata,
	}, nil
}

// detectToolNeeds 检测是否需要使用工具
func (a *Agent) detectToolNeeds(message string) (string, map[string]interface{}) {
	// 1. 股票查询检测（优先级最高，因为通常比较明确）
	if containsAny(message, []string{"股票", "股价", "报价", "stock", "AAPL", "股市"}) {
		symbol := extractStockSymbol(message)
		return "stock_quote", map[string]interface{}{"symbol": symbol}
	}

	// 2. 天气查询检测
	if containsAny(message, []string{"天气", "气温", "温度", "weather"}) {
		// 提取城市名（简化版）
		city := extractCity(message)
		return "weather", map[string]interface{}{"city": city}
	}

	// 3. 计算器检测
	if containsAny(message, []string{"计算", "加", "减", "乘", "除", "+", "-", "*", "/"}) {
		return "calculator", map[string]interface{}{"expression": message}
	}

	// 4. 搜索检测
	if containsAny(message, []string{"搜索", "查一下", "百度", "search"}) {
		// 提取搜索关键词（简化版）
		query := extractSearchQuery(message)
		return "search", map[string]interface{}{"query": query}
	}

	return "", nil
}

// containsAny 检查消息是否包含任何关键词
func containsAny(message string, keywords []string) bool {
	for _, keyword := range keywords {
		if contains(message, keyword) {
			return true
		}
	}
	return false
}

// contains 简单的字符串包含检查
func contains(message, keyword string) bool {
	return len(message) >= len(keyword) &&
		(message == keyword ||
		 len(message) > len(keyword) && indexOf(message, keyword) >= 0)
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// extractCity 提取城市名
func extractCity(message string) string {
	cities := []string{"北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "西安", "南京", "New York", "London", "Tokyo"}
	for _, city := range cities {
		if contains(message, city) {
			return city
		}
	}
	return "北京" // 默认城市
}

// extractSearchQuery 提取搜索关键词
func extractSearchQuery(message string) string {
	// 移除常见的搜索动词
	keywords := []string{"搜索", "查一下", "百度", "search", "帮我搜"}
	query := message
	for _, keyword := range keywords {
		idx := indexOf(query, keyword)
		if idx >= 0 {
			query = query[idx+len(keyword):]
			break
		}
	}
	// 去除首尾空格
	query = strings.TrimSpace(query)
	// 去除中文标点
	query = strings.TrimRight(query, "，。！？、；：")

	if query == "" {
		query = "最新AI新闻"
	}
	return query
}

// extractStockSymbol 提取股票代码
func extractStockSymbol(message string) string {
	// 常见的股票代码模式
	message = strings.ToUpper(message)

	// 检查是否包含括号中的股票代码，如 (AAPL) 或（AAPL）
	if strings.Contains(message, "(") && strings.Contains(message, ")") {
		start := strings.Index(message, "(") + 1
		end := strings.Index(message, ")")
		if end > start {
			symbol := message[start:end]
			if isValidStockSymbol(symbol) {
				return symbol
			}
		}
	}

	// 检查中文括号（支持中文输入）
	if strings.Contains(message, "（") && strings.Contains(message, "）") {
		start := strings.Index(message, "（") + 3
		end := strings.Index(message, "）")
		if end > start {
			symbol := message[start:end]
			if isValidStockSymbol(symbol) {
				return symbol
			}
		}
	}

	// 常见股票代码映射
	stockMap := map[string]string{
		"苹果":   "AAPL",
		"微软":   "MSFT",
		"谷歌":   "GOOGL",
		"亚马逊":  "AMZN",
		"特斯拉":  "TSLA",
		"英伟达":  "NVDA",
		"腾讯":   "0700.HK",
		"阿里巴巴": "BABA",
		"百度":   "BIDU",
		"京东":   "JD",
		"美团":   "MPNGF",
	}

	for name, symbol := range stockMap {
		if strings.Contains(message, name) {
			return symbol
		}
	}

	// 默认返回苹果
	return "AAPL"
}

// isValidStockSymbol 检查是否是有效的股票代码
func isValidStockSymbol(symbol string) bool {
	// 股票代码通常是1-5个大写字母或数字
	if len(symbol) < 1 || len(symbol) > 5 {
		return false
	}
	for _, ch := range symbol {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '.') {
			return false
		}
	}
	return true
}

// ClearSession 清除会话
func (a *Agent) ClearSession(sessionID string) error {
	return a.memoryMgr.ClearSession(sessionID)
}

// GetSession 获取会话信息
func (a *Agent) GetSession(sessionID string) (*models.Session, error) {
	return a.memoryMgr.GetSession(sessionID)
}
