package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai-agent-assistant/internal/config"
	"ai-agent-assistant/pkg/models"
)

// LLMProvider 大模型提供商接口
type LLMProvider interface {
	Chat(ctx context.Context, messages []models.Message) (string, error)
	ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error)
}

// GLMProvider GLM模型提供商
type GLMProvider struct {
	config config.ModelConfig
	client *http.Client
}

// QwenProvider 千问模型提供商
type QwenProvider struct {
	config config.ModelConfig
	client *http.Client
}

// NewGLMProvider 创建GLM提供商
func NewGLMProvider(cfg config.ModelConfig) (*GLMProvider, error) {
	return &GLMProvider{
		config: cfg,
		client: &http.Client{},
	}, nil
}

// NewQwenProvider 创建千问提供商
func NewQwenProvider(cfg config.ModelConfig) (*QwenProvider, error) {
	return &QwenProvider{
		config: cfg,
		client: &http.Client{},
	}, nil
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice 选择
type Choice struct {
	Message      ChatMessage `json:"message"`
	Delta        *ChatMessage `json:"delta,omitempty"` // 用于流式响应
	FinishReason string       `json:"finish_reason"`
}

// Chat 实现GLM的Chat方法
func (p *GLMProvider) Chat(ctx context.Context, messages []models.Message) (string, error) {
	chatMessages := convertMessages(messages)

	reqBody := ChatRequest{
		Model:    p.config.Model,
		Messages: chatMessages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatStream 实现GLM的流式Chat
func (p *GLMProvider) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	chatMessages := convertMessages(messages)

	reqBody := ChatRequest{
		Model:    p.config.Model,
		Messages: chatMessages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status=%d", resp.StatusCode)
	}

	ch := make(chan string)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp struct {
				Choices []Choice `json:"choices"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta != nil {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					ch <- content
				}
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].FinishReason != "" {
				return
			}
		}
	}()

	return ch, nil
}

// Chat 实现千问的Chat方法
func (p *QwenProvider) Chat(ctx context.Context, messages []models.Message) (string, error) {
	chatMessages := convertMessages(messages)

	reqBody := ChatRequest{
		Model:    p.config.Model,
		Messages: chatMessages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatStream 实现千问的流式Chat
func (p *QwenProvider) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	chatMessages := convertMessages(messages)

	reqBody := ChatRequest{
		Model:    p.config.Model,
		Messages: chatMessages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status=%d", resp.StatusCode)
	}

	ch := make(chan string)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp struct {
				Choices []Choice `json:"choices"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta != nil {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					ch <- content
				}
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].FinishReason != "" {
				return
			}
		}
	}()

	return ch, nil
}

// LLMManager LLM管理器
type LLMManager struct {
	providers map[string]LLMProvider
}

// NewLLMManager 创建LLM管理器
func NewLLMManager(cfg *config.Config) (*LLMManager, error) {
	manager := &LLMManager{
		providers: make(map[string]LLMProvider),
	}

	// 初始化GLM
	glmProvider, err := NewGLMProvider(cfg.Models.GLM)
	if err != nil {
		return nil, fmt.Errorf("failed to init GLM provider: %w", err)
	}
	manager.providers["glm"] = glmProvider

	// 初始化千问
	qwenProvider, err := NewQwenProvider(cfg.Models.Qwen)
	if err != nil {
		return nil, fmt.Errorf("failed to init Qwen provider: %w", err)
	}
	manager.providers["qwen"] = qwenProvider

	return manager, nil
}

// GetProvider 获取指定的模型提供商
func (m *LLMManager) GetProvider(modelName string) (LLMProvider, error) {
	provider, ok := m.providers[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelName)
	}
	return provider, nil
}

// convertMessages 转换消息格式
func convertMessages(messages []models.Message) []ChatMessage {
	chatMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return chatMessages
}
