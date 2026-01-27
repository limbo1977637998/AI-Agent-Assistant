package mcp

import (
	"context"
	"fmt"
	"sync"

	"ai-agent-assistant/internal/tools"
)

// Manager MCP工具管理器
type Manager struct {
	clients map[string]*Client
	tools   map[string]tools.Tool
	mu      sync.RWMutex
}

// NewManager 创建MCP工具管理器
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
		tools:   make(map[string]tools.Tool),
	}
}

// RegisterServer 注册MCP服务器
func (m *Manager) RegisterServer(name string, serverURL string) error {
	ctx := context.Background()

	client := NewClient(serverURL)
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to MCP server %s: %w", name, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[name] = client

	// 注册所有工具
	for _, tool := range client.ListTools() {
		mcpTool := NewMCPTool(client, tool)
		m.tools[tool.Name] = mcpTool
	}

	return nil
}

// GetTool 获取指定工具
func (m *Manager) GetTool(name string) (tools.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, ok := m.tools[name]
	return tool, ok
}

// ListTools 列出所有工具
func (m *Manager) ListTools() []tools.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	toolList := make([]tools.Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		toolList = append(toolList, tool)
	}

	return toolList
}

// GetToolNames 获取所有工具名称
func (m *Manager) GetToolNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}

	return names
}

// GetToolCount 获取工具数量
func (m *Manager) GetToolCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.tools)
}

// RemoveServer 移除MCP服务器
func (m *Manager) RemoveServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[name]
	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	// 移除该服务器的所有工具
	for _, tool := range client.ListTools() {
		delete(m.tools, tool.Name)
	}

	// 关闭客户端
	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}

	delete(m.clients, name)
	return nil
}

// Close 关闭所有连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error

	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close server %s: %w", name, err)
		}
	}

	m.clients = make(map[string]*Client)
	m.tools = make(map[string]tools.Tool)

	return lastErr
}

// GetServers 获取所有注册的服务器名称
func (m *Manager) GetServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.clients))
	for name := range m.clients {
		servers = append(servers, name)
	}

	return servers
}
