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

// DeepSeekModel DeepSeek模型（包括推理模型DeepSeek-R1）
type DeepSeekModel struct {
	config ModelConfig
	client *http.Client
}

// NewDeepSeekModel 创建DeepSeek模型
func NewDeepSeekModel(config ModelConfig) (*DeepSeekModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("DeepSeek API key is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.deepseek.com/v1"
	}
	if config.Model == "" {
		config.Model = "deepseek-chat"
	}

	return &DeepSeekModel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Chat 实现Chat接口
func (m *DeepSeekModel) Chat(ctx context.Context, messages []models.Message) (string, error) {
	response, err := m.ChatWithOptions(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// ChatStream 实现流式Chat接口
func (m *DeepSeekModel) ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error) {
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
						Reasoning string `json:"reasoning_content"` // 推理内容（R1模型）
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
				delta := streamResp.Choices[0].Delta
				// 优先输出推理内容
				if delta.Reasoning != "" {
					ch <- "[思考] " + delta.Reasoning
				}
				if delta.Content != "" {
					ch <- delta.Content
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
func (m *DeepSeekModel) ChatWithOptions(ctx context.Context, messages []models.Message, options map[string]interface{}) (*ChatResponse, error) {
	reqBody := m.buildChatRequest(messages, false)

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

	var deepseekResp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				Reasoning string `json:"reasoning_content"` // 推理内容（R1模型）
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}

	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := deepseekResp.Choices[0]
	content := choice.Message.Content

	// 如果有推理内容，添加到响应中
	if choice.Message.Reasoning != "" {
		content = "[思考过程]\n" + choice.Message.Reasoning + "\n\n[答案]\n" + content
	}

	return &ChatResponse{
		Content:      content,
		FinishReason: choice.FinishReason,
		Usage:        &deepseekResp.Usage,
	}, nil
}

// SupportsToolCalling DeepSeek支持工具调用
func (m *DeepSeekModel) SupportsToolCalling() bool {
	return true
}

// SupportsEmbedding DeepSeek暂不支持原生Embedding
func (m *DeepSeekModel) SupportsEmbedding() bool {
	return false
}

// Embed DeepSeek暂不支持Embedding
func (m *DeepSeekModel) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("DeepSeek does not support native embedding API")
}

// GetModelName 获取模型名称
func (m *DeepSeekModel) GetModelName() string {
	return m.config.Model
}

// GetProviderName 获取提供商名称
func (m *DeepSeekModel) GetProviderName() string {
	return "deepseek"
}

// SetTemperature 设置温度
func (m *DeepSeekModel) SetTemperature(temp float64) {
	m.config.Temperature = temp
}

// SetMaxTokens 设置最大token数
func (m *DeepSeekModel) SetMaxTokens(tokens int) {
	m.config.MaxTokens = tokens
}

// ReasonWithChainOfThought 思维链推理（使用DeepSeek-R1）
func (m *DeepSeekModel) ReasonWithChainOfThought(ctx context.Context, task string, showReasoning bool) (string, string, error) {
	// 构造推理提示
	prompt := fmt.Sprintf("请逐步思考以下问题，展示你的推理过程：\n\n%s", task)

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	response, err := m.ChatWithOptions(ctx, messages, nil)
	if err != nil {
		return "", "", err
	}

	// 解析推理内容和答案
	content := response.Content

	// DeepSeek-R1会在reasoning_content字段返回推理过程
	// 这里简单分离（实际应用中可以更复杂的解析）
	reasoning := content
	answer := content

	if showReasoning {
		return reasoning, answer, nil
	}

	return "", answer, nil
}

// Reflect 自我反思
func (m *DeepSeekModel) Reflect(ctx context.Context, previousRuns []string) (string, string, error) {
	prompt := "请检查之前的回答，识别可能的错误或改进点：\n\n"
	for i, run := range previousRuns {
		prompt += fmt.Sprintf("尝试%d：%s\n\n", i+1, run)
	}
	prompt += "请提供反思和改进后的答案。"

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	response, err := m.Chat(ctx, messages)
	if err != nil {
		return "", "", err
	}

	// 简单分离反思和改进后的答案
	// 实际应用中可以用更复杂的解析
	return response, response, nil
}

// deepseekChatRequest DeepSeek聊天请求结构
type deepseekChatRequest struct {
	Model       string                 `json:"model"`
	Messages    []deepseekChatMessage  `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
}

type deepseekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// buildChatRequest 构建聊天请求
func (m *DeepSeekModel) buildChatRequest(messages []models.Message, stream bool) deepseekChatRequest {
	chatMessages := make([]deepseekChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = deepseekChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return deepseekChatRequest{
		Model:       m.config.Model,
		Messages:    chatMessages,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		Stream:      stream,
	}
}
