package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai-agent-assistant/pkg/models"
)

// ClaudeModel Anthropic Claude模型
type ClaudeModel struct {
	config ModelConfig
	client *http.Client
}

// NewClaudeModel 创建Claude模型
func NewClaudeModel(config ModelConfig) (*ClaudeModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}
	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}

	return &ClaudeModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Chat 实现Chat接口
func (m *ClaudeModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	response, err := m.ChatWithOptions(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// ChatStream 实现流式Chat接口
func (m *ClaudeModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	reqBody := m.buildChatRequest(messages, true)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", m.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := m.client.Do(req)
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
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if streamResp.Type == "content_block_delta" && streamResp.Delta.Text != "" {
				ch <- streamResp.Delta.Text
			}

			if streamResp.Type == "message_stop" {
				return
			}
		}
	}()

	return ch, nil
}

// ChatWithOptions 带选项的对话
func (m *ClaudeModel) ChatWithOptions(ctx context.Context, messages []models.Message, options map[string]interface{}) (*ChatResponse, error) {
	reqBody := m.buildChatRequest(messages, false)

	// 应用选项
	if options != nil {
		if maxTokens, ok := options["max_tokens"].(int); ok {
			reqBody.MaxTokens = maxTokens
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", m.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var claudeResp struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Content      []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取文本内容
	var content string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &ChatResponse{
		Content: content,
		FinishReason: claudeResp.StopReason,
		Usage: &Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

// SupportsToolCalling Claude支持工具调用
func (m *ClaudeModel) SupportsToolCalling() bool {
	return true // Claude 3+ 都支持工具调用
}

// SupportsEmbedding Claude不支持原生Embedding
func (m *ClaudeModel) SupportsEmbedding() bool {
	return false
}

// Embed Claude不支持Embedding
func (m *ClaudeModel) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("Claude does not support native embedding API")
}

// GetModelName 获取模型名称
func (m *ClaudeModel) GetModelName() string {
	return m.config.Model
}

// GetProviderName 获取提供商名称
func (m *ClaudeModel) GetProviderName() string {
	return "anthropic"
}

// SetTemperature 设置温度
func (m *ClaudeModel) SetTemperature(temp float64) {
	m.config.Temperature = temp
}

// SetMaxTokens 设置最大token数
func (m *ClaudeModel) SetMaxTokens(tokens int) {
	m.config.MaxTokens = tokens
}

// claudeChatRequest Claude聊天请求结构
type claudeChatRequest struct {
	Model     string                `json:"model"`
	MaxTokens int                   `json:"max_tokens"`
	Messages  []claudeChatMessage   `json:"messages"`
	System    string                `json:"system,omitempty"`
	Stream    bool                  `json:"stream,omitempty"`
}

type claudeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// buildChatRequest 构建聊天请求
func (m *ClaudeModel) buildChatRequest(messages []models.Message, stream bool) claudeChatRequest {
	// Claude需要分离系统消息
	var systemMsg string
	chatMessages := make([]claudeChatMessage, 0)

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
		} else {
			chatMessages = append(chatMessages, claudeChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	return claudeChatRequest{
		Model:     m.config.Model,
		MaxTokens: m.config.MaxTokens,
		Messages:  chatMessages,
		System:    systemMsg,
		Stream:    stream,
	}
}
