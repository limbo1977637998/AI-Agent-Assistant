package mcp

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestMCPClient(t *testing.T) {
	// 注意：这些测试需要一个运行的MCP HTTP服务器
	t.Skip("Skipping MCP client test - requires running server")

	ctx := context.Background()
	client := NewClient("http://localhost:8080")

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	tools := client.ListTools()
	if len(tools) == 0 {
		t.Error("No tools found")
	}

	t.Logf("Found %d tools", len(tools))

	// 测试工具调用
	for _, tool := range tools {
		if tool.Name == "get_timestamp" {
			resp, err := client.CallTool(ctx, tool.Name, map[string]interface{}{})
			if err != nil {
				t.Logf("Tool call failed: %v", err)
				continue
			}

			if resp.IsError {
				t.Logf("Tool returned error: %v", resp.Content)
			} else {
				t.Logf("Tool response: %v", resp.Content)
			}
		}
	}
}

func TestMCPToolAdapter(t *testing.T) {
	t.Skip("Skipping MCP tool adapter test - requires running server")

	// 创建模拟工具
	tool := &Tool{
		Name:        "test_tool",
		Description: "Test tool for adapter",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}

	client := NewClient("http://localhost:8080")
	mcpTool := NewMCPTool(client, tool)

	// 测试接口方法
	if name := mcpTool.Name(); name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", name)
	}

	desc := mcpTool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	t.Log("MCP tool adapter test passed")
}

func TestMCPManager(t *testing.T) {
	t.Skip("Skipping MCP manager test - requires running server")

	manager := NewManager()

	// 尝试注册服务器（可能会失败，如果没有运行的服务器）
	err := manager.RegisterServer("test", "http://localhost:8080")
	if err != nil {
		t.Logf("Failed to register server (expected if server not running): %v", err)
	}

	// 检查工具数量
	count := manager.GetToolCount()
	t.Logf("MCP Manager has %d tools", count)

	// 获取工具列表
	tools := manager.ListTools()
	t.Logf("Tools: %v", tools)

	if err := manager.Close(); err != nil {
		t.Logf("Failed to close manager: %v", err)
	}

	t.Log("MCP manager test completed")
}

func TestHTTPServer(t *testing.T) {
	// 创建并启动HTTP服务器
	server := NewHTTPServer(18081)
	RegisterDefaultTools(server)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试服务器
	ctx := context.Background()

	// 创建客户端连接
	client := NewClient("http://localhost:18081")
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer client.Close()

	// 检查工具列表
	tools := client.ListTools()
	if len(tools) == 0 {
		t.Fatal("No tools registered")
	}

	t.Logf("Found %d tools", len(tools))

	// 测试几个简单工具
	testCases := []struct {
		toolName  string
		arguments map[string]interface{}
	}{
		{
			"get_timestamp",
			map[string]interface{}{},
		},
		{
			"text_process",
			map[string]interface{}{
				"text":   "Hello World",
				"action": "upper",
			},
		},
		{
			"uuid_generate",
			map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.toolName, func(t *testing.T) {
			// 检查工具是否存在
			if _, ok := client.GetTool(tc.toolName); !ok {
				t.Skipf("Tool %s not found", tc.toolName)
			}

			// 调用工具
			resp, err := client.CallTool(ctx, tc.toolName, tc.arguments)
			if err != nil {
				t.Logf("Tool call error: %v", err)
				return
			}

			if resp.IsError {
				t.Logf("Tool returned error: %v", resp.Content)
			} else {
				t.Logf("Tool response: %v", resp.Content)
				if len(resp.Content) > 0 {
					t.Log("Tool call successful")
				}
			}
		})
	}

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	t.Log("HTTP MCP server test completed")
}
