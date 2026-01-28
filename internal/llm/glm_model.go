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

// GLMModel 智谱GLM模型
type GLMModel struct {
	config ModelConfig
	client *http.Client
}

// NewGLMModel 创建GLM模型
func NewGLMModel(config ModelConfig) (*GLMModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("GLM API key is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
	}
	if config.Model == "" {
		config.Model = "glm-4-flash"
	}

	return &GLMModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Chat 实现Chat接口
func (m *GLMModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
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
func (m *GLMModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
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

// SupportsToolCalling GLM支持工具调用
func (m *GLMModel) SupportsToolCalling() bool {
	return true // GLM-4系列支持工具调用
}

// SupportsEmbedding GLM暂不支持原生Embedding
func (m *GLMModel) SupportsEmbedding() bool {
	return false
}

// Embed GLM暂不支持Embedding（需要使用千问或其他Embedding服务）
func (m *GLMModel) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("GLM does not support native embedding API, please use Qwen embedding instead")
}

// GetModelName 获取模型名称
func (m *GLMModel) GetModelName() string {
	return m.config.Model
}

// GetProviderName 获取提供商名称
func (m *GLMModel) GetProviderName() string {
	return "zhipu"
}

// SetTemperature 设置温度
func (m *GLMModel) SetTemperature(temp float64) {
	m.config.Temperature = temp
}

// SetMaxTokens 设置最大token数
func (m *GLMModel) SetMaxTokens(tokens int) {
	m.config.MaxTokens = tokens
}

// buildAPIChatRequest 构建聊天请求
func (m *GLMModel) buildAPIChatRequest(messages []models.Message, stream bool) APIChatRequest {
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
