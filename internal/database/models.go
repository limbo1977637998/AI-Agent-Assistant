package database

import (
	"encoding/json"
	"time"
)

// Session 会话模型
type Session struct {
	ID        int64     `db:"id" json:"id"`
	SessionID string    `db:"session_id" json:"session_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Model     string    `db:"model" json:"model"`
	Metadata  []byte    `db:"metadata" json:"metadata,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Message 消息模型
type Message struct {
	ID         int64     `db:"id" json:"id"`
	SessionID  string    `db:"session_id" json:"session_id"`
	Role       string    `db:"role" json:"role"` // user, assistant, system, tool
	Content    string    `db:"content" json:"content"`
	ToolCalls  []byte    `db:"tool_calls" json:"tool_calls,omitempty"`
	TokensUsed int       `db:"tokens_used" json:"tokens_used"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// UserMemory 用户记忆模型
type UserMemory struct {
	ID           int64     `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	Memory       string    `db:"memory" json:"memory"`
	Topics       string    `db:"topics" json:"topics,omitempty"`
	Importance   float64   `db:"importance" json:"importance"`
	MemoryType   string    `db:"memory_type" json:"memory_type"` // preference, background, history
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// ToolCall 工具调用记录
type ToolCall struct {
	ID         int64     `db:"id" json:"id"`
	SessionID  string    `db:"session_id" json:"session_id,omitempty"`
	UserID     string    `db:"user_id" json:"user_id,omitempty"`
	ToolName   string    `db:"tool_name" json:"tool_name"`
	Arguments  []byte    `db:"arguments" json:"arguments,omitempty"`
	Result     string    `db:"result" json:"result,omitempty"`
	Success    bool      `db:"success" json:"success"`
	ErrorMsg   string    `db:"error_msg" json:"error_msg,omitempty"`
	Duration   int       `db:"duration" json:"duration"` // 毫秒
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// AgentRun Agent运行记录
type AgentRun struct {
	ID            int64     `db:"id" json:"id"`
	RunID         string    `db:"run_id" json:"run_id"`
	SessionID     string    `db:"session_id" json:"session_id,omitempty"`
	UserID        string    `db:"user_id" json:"user_id,omitempty"`
	Input         string    `db:"input" json:"input"`
	Output        string    `db:"output" json:"output,omitempty"`
	Model         string    `db:"model" json:"model,omitempty"`
	InputTokens   int       `db:"input_tokens" json:"input_tokens"`
	OutputTokens  int       `db:"output_tokens" json:"output_tokens"`
	TotalTokens   int       `db:"total_tokens" json:"total_tokens"`
	EstimatedCost float64   `db:"estimated_cost" json:"estimated_cost"`
	Latency       int       `db:"latency" json:"latency"` // 毫秒
	Success       bool      `db:"success" json:"success"`
	ErrorMsg      string    `db:"error_msg" json:"error_msg,omitempty"`
	RAGUsed       bool      `db:"rag_used" json:"rag_used"`
	ToolsUsed     []byte    `db:"tools_used" json:"tools_used,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// Helper function to convert metadata to/from JSON
func (s *Session) SetMetadata(metadata map[string]interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	s.Metadata = data
	return nil
}

func (s *Session) GetMetadata() (map[string]interface{}, error) {
	if len(s.Metadata) == 0 {
		return nil, nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(s.Metadata, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// Helper function to convert tool_calls to/from JSON
func (m *Message) SetToolCalls(toolCalls []map[string]interface{}) error {
	data, err := json.Marshal(toolCalls)
	if err != nil {
		return err
	}
	m.ToolCalls = data
	return nil
}

func (m *Message) GetToolCalls() ([]map[string]interface{}, error) {
	if len(m.ToolCalls) == 0 {
		return nil, nil
	}
	var toolCalls []map[string]interface{}
	if err := json.Unmarshal(m.ToolCalls, &toolCalls); err != nil {
		return nil, err
	}
	return toolCalls, nil
}

// Helper function to convert arguments to/from JSON
func (t *ToolCall) SetArguments(args map[string]interface{}) error {
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	t.Arguments = data
	return nil
}

func (t *ToolCall) GetArguments() (map[string]interface{}, error) {
	if len(t.Arguments) == 0 {
		return nil, nil
	}
	var args map[string]interface{}
	if err := json.Unmarshal(t.Arguments, &args); err != nil {
		return nil, err
	}
	return args, nil
}

// Helper function to convert tools_used to/from JSON
func (a *AgentRun) SetToolsUsed(tools []string) error {
	data, err := json.Marshal(tools)
	if err != nil {
		return err
	}
	a.ToolsUsed = data
	return nil
}

func (a *AgentRun) GetToolsUsed() ([]string, error) {
	if len(a.ToolsUsed) == 0 {
		return nil, nil
	}
	var tools []string
	if err := json.Unmarshal(a.ToolsUsed, &tools); err != nil {
		return nil, err
	}
	return tools, nil
}
