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

// QwenModel 阿里云千问模型
type QwenModel struct {
	config ModelConfig
	client *http.Client
}

// NewQwenModel 创建千问模型
func NewQwenModel(config ModelConfig) (*QwenModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Qwen API key is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if config.Model == "" {
		config.Model = "qwen-plus"
	}

	return &QwenModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Chat 实现Chat接口
func (m *QwenModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	reqBody := m.buildAPIChatRequest(messages, false)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)

	resp, err := m.client.Do(req)
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

	var chatResp APIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatStream 实现流式Chat接口
func (m *QwenModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	reqBody := m.buildAPIChatRequest(messages, true)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)

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
				Choices []APIChoice `json:"choices"`
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

// SupportsToolCalling 千问支持工具调用
func (m *QwenModel) SupportsToolCalling() bool {
	return true // 千问Plus/Max支持工具调用
}

// SupportsEmbedding 千问支持Embedding（需要单独的API调用）
func (m *QwenModel) SupportsEmbedding() bool {
	return true
}

// Embed 文本向量化（使用千问text-embedding-v3）
func (m *QwenModel) Embed(ctx context.Context, text string) ([]float64, error) {
	if len(text) > 8000 {
		text = text[:8000] // 千问Embedding有长度限制
	}

	reqBody := map[string]interface{}{
		"model": "text-embedding-v3",
		"input": map[string]string{
			"text": text,
		},
		"parameters": map[string]int{
			"text_type": 1, // 1表示文档
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)

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

	var embedResp struct {
		Output struct {
			Embeddings []struct {
				Embedding []float64 `json:"embedding"`
			} `json:"embeddings"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embedResp.Output.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return embedResp.Output.Embeddings[0].Embedding, nil
}

// GetModelName 获取模型名称
func (m *QwenModel) GetModelName() string {
	return m.config.Model
}

// GetProviderName 获取提供商名称
func (m *QwenModel) GetProviderName() string {
	return "qwen"
}

// SetTemperature 设置温度
func (m *QwenModel) SetTemperature(temp float64) {
	m.config.Temperature = temp
}

// SetMaxTokens 设置最大token数
func (m *QwenModel) SetMaxTokens(tokens int) {
	m.config.MaxTokens = tokens
}

// buildAPIChatRequest 构建聊天请求
func (m *QwenModel) buildAPIChatRequest(messages []models.Message, stream bool) APIChatRequest {
	chatMessages := make([]APIChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = APIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return APIChatRequest{
		Model:       m.config.Model,
		Messages:    chatMessages,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		Stream:      stream,
	}
}
