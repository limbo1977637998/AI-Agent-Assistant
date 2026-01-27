package handler

import (
	"net/http"

	"ai-agent-assistant/internal/agent"
	"ai-agent-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

// Handler HTTP处理器
type Handler struct {
	agent *agent.Agent
}

// NewHandler 创建处理器
func NewHandler(agent *agent.Agent) *Handler {
	return &Handler{
		agent: agent,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string                 `json:"session_id" binding:"required"`
	Message   string                 `json:"message" binding:"required"`
	Model     string                 `json:"model"`
	Stream    bool                   `json:"stream"`
	WithTools bool                   `json:"with_tools"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    *models.ChatResponse   `json:"data,omitempty"`
}

// SessionResponse 会话响应
type SessionResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *models.Session    `json:"data,omitempty"`
}

// Chat 聊天接口
func (h *Handler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ChatResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 构建请求
	chatReq := &models.ChatRequest{
		SessionID: req.SessionID,
		Message:   req.Message,
		Model:     req.Model,
		Stream:    req.Stream,
		Metadata:  req.Metadata,
	}

	// 调用Agent
	var resp *models.ChatResponse
	var err error

	if req.WithTools {
		resp, err = h.agent.ChatWithTools(c.Request.Context(), chatReq)
	} else {
		resp, err = h.agent.Chat(c.Request.Context(), chatReq)
	}

	if err != nil {
		c.JSON(http.StatusOK, ChatResponse{
			Code:    500,
			Message: "处理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    200,
		Message: "success",
		Data:    resp,
	})
}

// GetSession 获取会话信息
func (h *Handler) GetSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusOK, SessionResponse{
			Code:    400,
			Message: "session_id参数缺失",
		})
		return
	}

	session, err := h.agent.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusOK, SessionResponse{
			Code:    404,
			Message: "会话不存在",
		})
		return
	}

	c.JSON(http.StatusOK, SessionResponse{
		Code:    200,
		Message: "success",
		Data:    session,
	})
}

// ClearSession 清除会话
func (h *Handler) ClearSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusOK, ChatResponse{
			Code:    400,
			Message: "session_id参数缺失",
		})
		return
	}

	if err := h.agent.ClearSession(sessionID); err != nil {
		c.JSON(http.StatusOK, ChatResponse{
			Code:    500,
			Message: "清除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    200,
		Message: "清除成功",
	})
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
