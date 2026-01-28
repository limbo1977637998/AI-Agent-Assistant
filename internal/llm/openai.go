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

// OpenAIModel OpenAI模型（GPT-4, GPT-3.5等）
type OpenAIModel struct {
	config ModelConfig
	client *http.Client
}

// NewOpenAIModel 创建OpenAI模型
func NewOpenAIModel(config ModelConfig) (*OpenAIModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	return &OpenAIModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Chat 实现Chat接口
func (m *OpenAIModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	response, err := m.ChatWithOptions(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// ChatStream 实现流式Chat接口
func (m *OpenAIModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
	reqBody := m.buildChatRequest(messages, true)

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
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					ch <- content
				}
				if streamResp.Choices[0].FinishReason != "" {
					return
				}
			}
		}
	}()

	return ch, nil
}

// ChatWithOptions 带选项的对话
func (m *OpenAIModel) ChatWithOptions(ctx context.Context, messages []models.Message, options map[string]interface{}) (*ChatResponse, error) {
	reqBody := m.buildChatRequest(messages, false)

	// 应用选项
	if options != nil {
		if temp, ok := options["temperature"].(float64); ok {
			reqBody.Temperature = temp
		}
		if maxTokens, ok := options["max_tokens"].(int); ok {
			reqBody.MaxTokens = maxTokens
		}
	}

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content   string       `json:"content"`
				ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}

	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	return &ChatResponse{
		Content:      choice.Message.Content,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
		Usage:        &openaiResp.Usage,
	}, nil
}

// SupportsToolCalling OpenAI支持工具调用
func (m *OpenAIModel) SupportsToolCalling() bool {
	return m.config.EnableToolCalling
}

// SupportsEmbedding OpenAI支持向量化（通过单独的API）
func (m *OpenAIModel) SupportsEmbedding() bool {
	return true
}

// Embed 文本向量化
func (m *OpenAIModel) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.BaseURL+"/embeddings", bytes.NewBuffer(jsonData))
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
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embedResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// GetModelName 获取模型名称
func (m *OpenAIModel) GetModelName() string {
	return m.config.Model
}

// GetProviderName 获取提供商名称
func (m *OpenAIModel) GetProviderName() string {
	return "openai"
}

// SetTemperature 设置温度
func (m *OpenAIModel) SetTemperature(temp float64) {
	m.config.Temperature = temp
}

// SetMaxTokens 设置最大token数
func (m *OpenAIModel) SetMaxTokens(tokens int) {
	m.config.MaxTokens = tokens
}

// openAIChatRequest OpenAI聊天请求结构
type openAIChatRequest struct {
	Model       string                 `json:"model"`
	Messages    []openAIChatMessage    `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Tools       []map[string]interface{} `json:"tools,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// buildChatRequest 构建聊天请求
func (m *OpenAIModel) buildChatRequest(messages []models.Message, stream bool) openAIChatRequest {
	chatMessages := make([]openAIChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = openAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return openAIChatRequest{
		Model:       m.config.Model,
		Messages:    chatMessages,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		Stream:      stream,
	}
}
