package mcp

import (
	"context"
	"fmt"
	"strings"
)

// MCPTool MCP工具适配器
type MCPTool struct {
	client *Client
	tool   *Tool
}

// NewMCPTool 创建MCP工具适配器
func NewMCPTool(client *Client, tool *Tool) *MCPTool {
	return &MCPTool{
		client: client,
		tool:   tool,
	}
}

// Name 返回工具名称
func (t *MCPTool) Name() string {
	return t.tool.Name
}

// Description 返回工具描述
func (t *MCPTool) Description() string {
	return t.tool.Description
}

// Execute 执行工具
func (t *MCPTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	resp, err := t.client.CallTool(ctx, t.tool.Name, args)
	if err != nil {
		return "", fmt.Errorf("failed to call tool %s: %w", t.tool.Name, err)
	}

	if resp.IsError {
		return "", fmt.Errorf("tool execution failed: %v", resp.Content)
	}

	// 将响应内容转换为字符串
	return formatContent(resp.Content), nil
}

// formatContent 格式化工具响应内容
func formatContent(content []interface{}) string {
	var result strings.Builder

	for _, item := range content {
		switch v := item.(type) {
		case string:
			result.WriteString(v)
		case map[string]interface{}:
			if text, ok := v["text"].(string); ok {
				result.WriteString(text)
			} else {
				// 将map转换为JSON字符串
				result.WriteString(formatMap(v))
			}
		case []interface{}:
			result.WriteString(formatContent(v))
		default:
			result.WriteString(fmt.Sprintf("%v", v))
		}
	}

	return result.String()
}

// formatMap 格式化map为字符串
func formatMap(m map[string]interface{}) string {
	var result strings.Builder
	result.WriteString("{")

	first := true
	for k, v := range m {
		if !first {
			result.WriteString(", ")
		}
		first = false

		result.WriteString(fmt.Sprintf("%s: ", k))

		switch val := v.(type) {
		case string:
			result.WriteString(fmt.Sprintf("%q", val))
		case map[string]interface{}:
			result.WriteString(formatMap(val))
		case []interface{}:
			result.WriteString("[")
			for i, item := range val {
				if i > 0 {
					result.WriteString(", ")
				}
				result.WriteString(fmt.Sprintf("%v", item))
			}
			result.WriteString("]")
		default:
			result.WriteString(fmt.Sprintf("%v", val))
		}
	}

	result.WriteString("}")
	return result.String()
}

// GetInputSchema 获取输入schema
func (t *MCPTool) GetInputSchema() map[string]interface{} {
	return t.tool.InputSchema
}
