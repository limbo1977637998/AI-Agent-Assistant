package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// HTTPServer HTTP MCP服务器
type HTTPServer struct {
	port    int
	tools   map[string]*HTTPTool
	server  *http.Server
	mu      sync.RWMutex
}

// HTTPTool HTTP工具定义
type HTTPTool struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// NewHTTPServer 创建HTTP MCP服务器
func NewHTTPServer(port int) *HTTPServer {
	return &HTTPServer{
		port:  port,
		tools: make(map[string]*HTTPTool),
	}
}

// RegisterTool 注册工具
func (s *HTTPServer) RegisterTool(tool *HTTPTool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools[tool.Name] = tool
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	fmt.Printf("MCP HTTP Server listening on port %d\n", s.port)
	return s.server.ListenAndServe()
}

// Shutdown 关闭服务器
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleRequest 处理HTTP请求
func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	method, _ := payload["method"].(string)

	switch method {
	case "tools/list":
		s.handleListTools(w, r)
	case "tools/call":
		s.handleCallTool(w, r, payload)
	default:
		http.Error(w, fmt.Sprintf("unknown method: %s", method), http.StatusBadRequest)
	}
}

// handleListTools 处理工具列表请求
func (s *HTTPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]*Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, &Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		})
	}

	response := map[string]interface{}{
		"tools": tools,
	}

	json.NewEncoder(w).Encode(response)
}

// handleCallTool 处理工具调用请求
func (s *HTTPServer) handleCallTool(w http.ResponseWriter, r *http.Request, payload map[string]interface{}) {
	params, _ := payload["params"].(map[string]interface{})
	name, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]interface{})

	s.mu.RLock()
	tool, ok := s.tools[name]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, fmt.Sprintf("tool not found: %s", name), http.StatusNotFound)
		return
	}

	// 执行工具
	result, err := tool.Handler(context.Background(), arguments)

	var content []interface{}
	var isError bool

	if err != nil {
		isError = true
		content = []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": err.Error(),
			},
		}
	} else {
		content = []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("%v", result),
			},
		}
	}

	response := map[string]interface{}{
		"content": content,
		"isError": isError,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTools 获取所有工具
func (s *HTTPServer) GetTools() []*HTTPTool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]*HTTPTool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return tools
}
