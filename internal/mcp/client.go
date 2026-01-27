package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Client MCP客户端
type Client struct {
	serverURL string
	httpClient *http.Client
	tools      map[string]*Tool
	mu        sync.RWMutex
}

// Tool MCP工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCallRequest 工具调用请求
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResponse 工具调用响应
type ToolCallResponse struct {
	Content []interface{} `json:"content"`
	IsError bool           `json:"isError"`
}

// NewClient 创建MCP客户端
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{},
		tools:      make(map[string]*Tool),
	}
}

// Connect 连接到MCP服务器
func (c *Client) Connect(ctx context.Context) error {
	// 列出可用的工具
	tools, err := c.listTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 保存工具列表
	for _, tool := range tools {
		c.tools[tool.Name] = tool
	}

	return nil
}

// listTools 列出所有可用工具
func (c *Client) listTools(ctx context.Context) ([]*Tool, error) {
	req := map[string]interface{}{
		"method": "tools/list",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []*Tool `json:"tools"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %w", err)
	}

	return result.Tools, nil
}

// ListTools 获取已注册的工具列表
func (c *Client) ListTools() []*Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]*Tool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetTool 获取指定工具
func (c *Client) GetTool(name string) (*Tool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tool, ok := c.tools[name]
	return tool, ok
}

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolCallResponse, error) {
	// 检查工具是否存在
	if _, ok := c.GetTool(name); !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	req := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var result ToolCallResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool call response: %w", err)
	}

	return &result, nil
}

// sendRequest 发送HTTP请求
func (c *Client) sendRequest(ctx context.Context, payload map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return responseBody, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tools = make(map[string]*Tool)
	return nil
}
