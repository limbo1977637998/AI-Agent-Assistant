package handler

import (
	"net/http"

	"ai-agent-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

// AddKnowledgeRequest 添加知识请求
type AddKnowledgeRequest struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// AddKnowledgeFromDocRequest 从文档添加知识请求
type AddKnowledgeFromDocRequest struct {
	DocPath string `json:"doc_path"`
}

// KnowledgeStatsResponse 知识库统计响应
type KnowledgeStatsResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// SearchKnowledgeRequest 搜索知识请求
type SearchKnowledgeRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k,omitempty"`
}

// SearchKnowledgeResponse 搜索知识响应
type SearchKnowledgeResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data,omitempty"`
}

// AddKnowledge 添加知识
func (h *Handler) AddKnowledge(c *gin.Context) {
	var req AddKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    400,
			Message: "text参数不能为空",
		})
		return
	}

	if req.Source == "" {
		req.Source = "manual"
	}

	err := h.agent.AddKnowledge(c.Request.Context(), req.Text, req.Source)
	if err != nil {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    500,
			Message: "添加知识失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, KnowledgeStatsResponse{
		Code:    200,
		Message: "success",
	})
}

// AddKnowledgeFromDoc 从文档添加知识
func (h *Handler) AddKnowledgeFromDoc(c *gin.Context) {
	var req AddKnowledgeFromDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.DocPath == "" {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    400,
			Message: "doc_path参数不能为空",
		})
		return
	}

	err := h.agent.AddKnowledgeFromDoc(c.Request.Context(), req.DocPath)
	if err != nil {
		c.JSON(http.StatusOK, KnowledgeStatsResponse{
			Code:    500,
			Message: "添加文档失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, KnowledgeStatsResponse{
		Code:    200,
		Message: "success",
	})
}

// GetKnowledgeStats 获取知识库统计
func (h *Handler) GetKnowledgeStats(c *gin.Context) {
	stats := h.agent.GetKnowledgeStats()

	c.JSON(http.StatusOK, KnowledgeStatsResponse{
		Code:    200,
		Message: "success",
		Data:    stats,
	})
}

// SearchKnowledge 搜索知识库
func (h *Handler) SearchKnowledge(c *gin.Context) {
	var req SearchKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, SearchKnowledgeResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5 // 默认返回5个结果
	}

	results, err := h.agent.SearchKnowledge(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusOK, SearchKnowledgeResponse{
			Code:    500,
			Message: "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SearchKnowledgeResponse{
		Code:    200,
		Message: "success",
		Data:    results,
	})
}

// ChatWithRAG 使用RAG增强的对话
func (h *Handler) ChatWithRAG(c *gin.Context) {
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
		UseRAG:    true, // 启用RAG
		Metadata:  req.Metadata,
	}

	// 调用Agent
	resp, err := h.agent.ChatWithRAG(c.Request.Context(), chatReq)
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
