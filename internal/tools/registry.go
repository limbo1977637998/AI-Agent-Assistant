package tools

import (
	"context"
	"fmt"
	"sync"
)

// Tool 工具接口
// 所有工具都需要实现这个接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// Version 返回工具版本
	Version() string
}

// ExecutionContext 工具执行上下文
type ExecutionContext struct {
	context.Context
	AgentID    string                 `json:"agent_id"`             // Agent ID
	TaskID     string                 `json:"task_id"`              // 任务ID
	Parameters map[string]interface{} `json:"parameters"`           // 参数
	Metadata   map[string]interface{} `json:"metadata,omitempty"`   // 元数据
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success   bool                   `json:"success"`              // 是否成功
	Message   string                 `json:"message"`              // 结果消息
	Data      interface{}            `json:"data,omitempty"`       // 返回数据
	Error     string                 `json:"error,omitempty"`      // 错误信息
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // 元数据
}

// ToolExecutor 工具执行器接口
type ToolExecutor interface {
	Tool
	// Execute 执行工具操作
	Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
}

// Registry 工具注册表
// 管理所有可用的工具
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]ToolExecutor
}

// NewRegistry 创建新的工具注册表
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]ToolExecutor),
	}
}

// Register 注册工具
func (r *Registry) Register(tool ToolExecutor) error {
	if tool == nil {
		return fmt.Errorf("工具不能为空")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("工具已存在: %s", name)
	}

	r.tools[name] = tool
	return nil
}

// Unregister 注销工具
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("工具不存在: %s", name)
	}

	delete(r.tools, name)
	return nil
}

// Get 获取工具
func (r *Registry) Get(name string) (ToolExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("工具不存在: %s", name)
	}

	return tool, nil
}

// List 列出所有工具
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// ListByName 按名称列出工具
func (r *Registry) ListByName() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}

// Execute 执行工具操作
func (r *Registry) Execute(ctx context.Context, toolName, operation string, params map[string]interface{}) (interface{}, error) {
	tool, err := r.Get(toolName)
	if err != nil {
		return nil, err
	}

	return tool.Execute(ctx, operation, params)
}

// GetToolInfo 获取工具信息
func (r *Registry) GetToolInfo(name string) (map[string]interface{}, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":        tool.Name(),
		"description": tool.Description(),
		"version":     tool.Version(),
	}, nil
}

// GetAllToolsInfo 获取所有工具信息
func (r *Registry) GetAllToolsInfo() []map[string]interface{} {
	tools := r.List()
	infos := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		info := map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"version":     tool.Version(),
		}
		infos = append(infos, info)
	}

	return infos
}

// HasTool 检查工具是否存在
func (r *Registry) HasTool(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.tools[name]
	return exists
}

// Count 返回工具数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Clear 清空所有工具
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]ToolExecutor)
}

// ToolManager 工具管理器
// 提供更高级的工具管理功能
type ToolManager struct {
	registry *Registry
	config   *ToolManagerConfig
}

// ToolManagerConfig 工具管理器配置
type ToolManagerConfig struct {
	AutoRegister bool     `json:"auto_register"` // 是否自动注册内置工具
	EnabledTools []string `json:"enabled_tools"` // 启用的工具列表
}

// NewToolManager 创建工具管理器
func NewToolManager(config *ToolManagerConfig) *ToolManager {
	if config == nil {
		config = &ToolManagerConfig{
			AutoRegister: true,
		}
	}

	manager := &ToolManager{
		registry: NewRegistry(),
		config:   config,
	}

	// 自动注册内置工具
	if config.AutoRegister {
		manager.registerBuiltinTools()
	}

	return manager
}

// registerBuiltinTools 注册内置工具
func (m *ToolManager) registerBuiltinTools() {
	// 注册文件操作工具
	m.registry.Register(NewFileOpsTool())

	// 注册数据处理工具
	m.registry.Register(NewDataProcessorTool())

	// 注册批量操作工具
	m.registry.Register(NewBatchOpsTool())
}

// GetRegistry 获取工具注册表
func (m *ToolManager) GetRegistry() *Registry {
	return m.registry
}

// ExecuteTool 执行工具操作
func (m *ToolManager) ExecuteTool(ctx context.Context, toolName, operation string, params map[string]interface{}) (interface{}, error) {
	// 检查工具是否启用
	if !m.isToolEnabled(toolName) {
		return nil, fmt.Errorf("工具未启用: %s", toolName)
	}

	return m.registry.Execute(ctx, toolName, operation, params)
}

// isToolEnabled 检查工具是否启用
func (m *ToolManager) isToolEnabled(toolName string) bool {
	// 如果没有启用列表，则所有工具都启用
	if len(m.config.EnabledTools) == 0 {
		return true
	}

	for _, enabled := range m.config.EnabledTools {
		if enabled == toolName {
			return true
		}
	}

	return false
}

// GetAvailableTools 获取可用工具列表
func (m *ToolManager) GetAvailableTools() []map[string]interface{} {
	allTools := m.registry.GetAllToolsInfo()

	// 如果有启用列表，过滤
	if len(m.config.EnabledTools) > 0 {
		available := make([]map[string]interface{}, 0)
		for _, tool := range allTools {
			name := tool["name"].(string)
			if m.isToolEnabled(name) {
				available = append(available, tool)
			}
		}
		return available
	}

	return allTools
}

// EnableTool 启用工具
func (m *ToolManager) EnableTool(toolName string) {
	m.config.EnabledTools = append(m.config.EnabledTools, toolName)
}

// DisableTool 禁用工具
func (m *ToolManager) DisableTool(toolName string) {
	for i, name := range m.config.EnabledTools {
		if name == toolName {
			m.config.EnabledTools = append(m.config.EnabledTools[:i], m.config.EnabledTools[i+1:]...)
			break
		}
	}
}

// GetToolCapabilities 获取工具能力描述
func (m *ToolManager) GetToolCapabilities(toolName string) (map[string]interface{}, error) {
	tool, err := m.registry.Get(toolName)
	if err != nil {
		return nil, err
	}

	capabilities := map[string]interface{}{
		"name":        tool.Name(),
		"description": tool.Description(),
		"version":     tool.Version(),
	}

	// 根据工具类型添加特定能力
	switch toolName {
	case "file_ops":
		capabilities["operations"] = []string{
			"read", "write", "batch_read", "convert",
			"compress", "decompress", "list", "delete",
		}
	case "data_processor":
		capabilities["operations"] = []string{
			"parse_csv", "parse_json", "clean", "filter",
			"aggregate", "transform", "merge", "sort",
			"deduplicate", "fill_missing",
		}
	case "batch_ops":
		capabilities["operations"] = []string{
			"batch_http", "batch_process", "parallel_execute",
			"concurrent_limit",
		}
	}

	return capabilities, nil
}

// GetAllCapabilities 获取所有工具的能力
func (m *ToolManager) GetAllCapabilities() map[string]interface{} {
	tools := m.GetAvailableTools()
	capabilities := make(map[string]interface{})

	for _, tool := range tools {
		name := tool["name"].(string)
		toolCaps, _ := m.GetToolCapabilities(name)
		capabilities[name] = toolCaps
	}

	return capabilities
}

// ToolWithExecutor 带执行器的工具包装器
// 用于包装简单的函数为工具
type ToolWithExecutor struct {
	name        string
	description string
	version     string
	executor    func(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
}

// NewToolWithExecutor 创建带执行器的工具
func NewToolWithExecutor(name, description, version string,
	executor func(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)) *ToolWithExecutor {
	return &ToolWithExecutor{
		name:        name,
		description: description,
		version:     version,
		executor:    executor,
	}
}

// Name 返回工具名称
func (t *ToolWithExecutor) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *ToolWithExecutor) Description() string {
	return t.description
}

// Version 返回工具版本
func (t *ToolWithExecutor) Version() string {
	return t.version
}

// Execute 执行工具操作
func (t *ToolWithExecutor) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	return t.executor(ctx, operation, params)
}
