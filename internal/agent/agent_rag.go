package agent

import (
	"context"

	"ai-agent-assistant/pkg/models"
)

// AddKnowledge 添加知识到知识库
func (a *Agent) AddKnowledge(ctx context.Context, text string, source string) error {
	return a.rag.AddText(ctx, text, source)
}

// AddKnowledgeFromDoc 从文档添加知识
func (a *Agent) AddKnowledgeFromDoc(ctx context.Context, docPath string) error {
	return a.rag.AddDocument(ctx, docPath)
}

// GetKnowledgeStats 获取知识库统计
func (a *Agent) GetKnowledgeStats() map[string]interface{} {
	return a.rag.GetStats()
}

// SearchKnowledge 搜索知识库
func (a *Agent) SearchKnowledge(ctx context.Context, query string, topK int) ([]string, error) {
	return a.rag.Retrieve(ctx, query, topK)
}

// ChatWithRAG 使用RAG增强的对话
func (a *Agent) ChatWithRAG(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 确保启用RAG
	req.UseRAG = true
	return a.Chat(ctx, req)
}
