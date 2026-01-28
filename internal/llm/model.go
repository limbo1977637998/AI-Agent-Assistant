package llm

import (
	"context"

	"ai-agent-assistant/pkg/models"
)

// Model 统一的模型接口
type Model interface {
	// Chat 对话接口
	Chat(ctx context.Context, messages []models.Message) (string, error)

	// ChatStream 流式对话接口
	ChatStream(ctx context.Context, messages []models.Message) (<-chan string, error)

	// SupportsToolCalling 是否支持工具调用
	SupportsToolCalling() bool

	// SupportsEmbedding 是否支持向量化
	SupportsEmbedding() bool

	// Embed 文本向量化（如果支持）
	Embed(ctx context.Context, text string) ([]float64, error)

	// GetModelName 获取模型名称
	GetModelName() string

	// GetProviderName 获取提供商名称
	GetProviderName() string
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments string                 `json:"arguments"`
}

// ChatResponse 聊天响应（包含工具调用）
type ChatResponse struct {
	Content      string             `json:"content"`
	ToolCalls    []ToolCall         `json:"tool_calls,omitempty"`
	FinishReason string             `json:"finish_reason"`
	Usage       *Usage             `json:"usage,omitempty"`
}

// Usage Token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingResponse 向量化响应
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// ModelConfig 模型配置（扩展现有的config.ModelConfig）
type ModelConfig struct {
	APIKey             string  `json:"api_key"`
	BaseURL            string  `json:"base_url"`
	Model              string  `json:"model"`
	Temperature        float64 `json:"temperature,omitempty"`
	MaxTokens          int     `json:"max_tokens,omitempty"`
	TopP               float64 `json:"top_p,omitempty"`
	TimeoutSeconds     int     `json:"timeout,omitempty"`
	EnableToolCalling  bool    `json:"enable_tool_calling,omitempty"`
}

// ModelWithOptions 带选项的模型接口
type ModelWithOptions interface {
	Model

	// ChatWithOptions 带选项的对话
	ChatWithOptions(ctx context.Context, messages []models.Message, options map[string]interface{}) (*ChatResponse, error)

	// SetTemperature 设置温度参数
	SetTemperature(temp float64)

	// SetMaxTokens 设置最大token数
	SetMaxTokens(tokens int)
}

// StreamingModel 流式模型接口
type StreamingModel interface {
	// ChatStreamWithCallback 带回调的流式对话
	ChatStreamWithCallback(ctx context.Context, messages []models.Message, callback func(chunk string)) error
}

// ModelWithReasoning 支持推理的模型接口
type ModelWithReasoning interface {
	Model

	// ReasonWithChainOfThought 思维链推理
	ReasonWithChainOfThought(ctx context.Context, task string, showReasoning bool) (reasoning string, answer string, err error)

	// Reflect 自我反思
	Reflect(ctx context.Context, previousRuns []string) (reflection string, err error)
}

// ========== 通用API请求/响应类型 ==========

// APIChatRequest 通用聊天API请求
type APIChatRequest struct {
	Model       string                 `json:"model"`
	Messages    []APIChatMessage       `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Tools       []Tool                 `json:"tools,omitempty"`
	ToolChoice  string                 `json:"tool_choice,omitempty"`
}

// APIChatMessage 通用聊天消息
type APIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIChatResponse 通用聊天API响应
type APIChatResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []APIChoice         `json:"choices"`
	Usage   *Usage              `json:"usage,omitempty"`
}

// APIChoice 通用选择
type APIChoice struct {
	Index        int             `json:"index"`
	Message      APIChatMessage  `json:"message"`
	Delta        *APIChatMessage `json:"delta,omitempty"`
	FinishReason string          `json:"finish_reason"`
}

// APIEmbeddingRequest 通用向量化API请求
type APIEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// APIEmbeddingResponse 通用向量化API响应
type APIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage *struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

