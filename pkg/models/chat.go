package models

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string                 `json:"session_id"`
	Message   string                 `json:"message"`
	Model     string                 `json:"model,omitempty"`     // glm or qwen
	Stream    bool                   `json:"stream,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID string                 `json:"session_id"`
	Message   string                 `json:"message"`
	Model     string                 `json:"model"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Result   string                 `json:"result,omitempty"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`    // user, assistant, system, tool
	Content string `json:"content"`
	ToolID  string `json:"tool_id,omitempty"`
}

// Session 会话
type Session struct {
	ID       string     `json:"id"`
	Messages []Message  `json:"messages"`
	Model    string     `json:"model"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
