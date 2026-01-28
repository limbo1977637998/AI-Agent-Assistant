package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server 监控服务器
type Server struct {
	metrics         *Metrics
	port            int
	httpServer      *http.Server
	enabled         bool
}

// NewServer 创建监控服务器
func NewServer(metrics *Metrics, port int) *Server {
	return &Server{
		metrics: metrics,
		port:    port,
		enabled: true,
	}
}

// Start 启动监控服务器
func (s *Server) Start() error {
	if !s.enabled {
		return nil
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Metrics info endpoint
	mux.HandleFunc("/metrics/info", s.metricsInfoHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// 启动服务器
	go func() {
		fmt.Printf("Monitoring server starting on :%d\n", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Monitoring server error: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止监控服务器
func (s *Server) Stop(ctx context.Context) error {
	if !s.enabled || s.httpServer == nil {
		return nil
	}

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}

// healthHandler 健康检查处理器
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"agent-monitoring"}`)
}

// metricsInfoHandler 指标信息处理器
func (s *Server) metricsInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"service": "agent-monitoring",
		"version": "1.0.0",
		"endpoints": {
			"/metrics": "Prometheus metrics",
			"/health": "Health check"
		}
	}`)
}

// Enable 启用监控
func (s *Server) Enable() {
	s.enabled = true
}

// Disable 禁用监控
func (s *Server) Disable() {
	s.enabled = false
}

// IsEnabled 检查是否启用
func (s *Server) IsEnabled() bool {
	return s.enabled
}

// GetMetrics 获取指标收集器
func (s *Server) GetMetrics() *Metrics {
	return s.metrics
}
