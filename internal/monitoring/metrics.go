package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics Agent运行指标
type Metrics struct {
	// 请求计数
	RequestCount *prometheus.CounterVec

	// 响应时间
	ResponseTime *prometheus.HistogramVec

	// Token使用
	TokenUsage *prometheus.CounterVec

	// 工具调用
	ToolCallCount *prometheus.CounterVec
	ToolCallDuration *prometheus.HistogramVec
	ToolCallErrors *prometheus.CounterVec

	// 错误计数
	ErrorCount *prometheus.CounterVec

	// 缓存
	CacheHitCount *prometheus.CounterVec
	CacheMissCount *prometheus.CounterVec

	// RAG
	RAGRetrievalTime *prometheus.HistogramVec
	RAGRetrievalCount *prometheus.CounterVec
	RAGKnowledgeCount *prometheus.GaugeVec

	// 推理
	ReasoningCount *prometheus.CounterVec
	ReasoningTime *prometheus.HistogramVec
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		// 请求指标
		RequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_requests_total",
				Help: "Total number of agent requests",
			},
			[]string{"agent_name", "model", "status"},
		),

		ResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agent_response_time_seconds",
				Help:    "Agent response time in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
			},
			[]string{"agent_name", "model"},
		),

		TokenUsage: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_tokens_total",
				Help: "Total number of tokens consumed",
			},
			[]string{"agent_name", "model", "type"}, // type: prompt, completion
		),

		// 工具调用指标
		ToolCallCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_tool_calls_total",
				Help: "Total number of tool calls",
			},
			[]string{"agent_name", "tool_name", "status"},
		),

		ToolCallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agent_tool_call_duration_seconds",
				Help:    "Tool call duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
			},
			[]string{"agent_name", "tool_name"},
		),

		ToolCallErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_tool_call_errors_total",
				Help: "Total number of tool call errors",
			},
			[]string{"agent_name", "tool_name", "error_type"},
		),

		// 错误指标
		ErrorCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_errors_total",
				Help: "Total number of errors",
			},
			[]string{"agent_name", "model", "error_type"},
		),

		// 缓存指标
		CacheHitCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"agent_name", "cache_type"}, // cache_type: llm, tool, session
		),

		CacheMissCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"agent_name", "cache_type"},
		),

		// RAG指标
		RAGRetrievalTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agent_rag_retrieval_time_seconds",
				Help:    "RAG retrieval time in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"agent_name", "retrieval_type"}, // retrieval_type: vector, hybrid
		),

		RAGRetrievalCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_rag_retrievals_total",
				Help: "Total number of RAG retrievals",
			},
			[]string{"agent_name", "retrieval_type"},
		),

		RAGKnowledgeCount: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "agent_rag_knowledge_count",
				Help: "Number of documents in knowledge base",
			},
			[]string{"agent_name", "knowledge_base"},
		),

		// 推理指标
		ReasoningCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "agent_reasoning_total",
				Help: "Total number of reasoning operations",
			},
			[]string{"agent_name", "reasoning_type"}, // reasoning_type: cot, reflection
		),

		ReasoningTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "agent_reasoning_time_seconds",
				Help:    "Reasoning operation time in seconds",
				Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
			[]string{"agent_name", "reasoning_type"},
		),
	}
}

// RecordRequest 记录请求
func (m *Metrics) RecordRequest(agentName, modelName, status string, duration time.Duration) {
	m.RequestCount.WithLabelValues(agentName, modelName, status).Inc()
	m.ResponseTime.WithLabelValues(agentName, modelName).Observe(duration.Seconds())
}

// RecordTokenUsage 记录Token使用
func (m *Metrics) RecordTokenUsage(agentName, modelName, tokenType string, count int) {
	m.TokenUsage.WithLabelValues(agentName, modelName, tokenType).Add(float64(count))
}

// RecordToolCall 记录工具调用
func (m *Metrics) RecordToolCall(agentName, toolName, status string, duration time.Duration) {
	m.ToolCallCount.WithLabelValues(agentName, toolName, status).Inc()
	m.ToolCallDuration.WithLabelValues(agentName, toolName).Observe(duration.Seconds())
}

// RecordToolCallError 记录工具调用错误
func (m *Metrics) RecordToolCallError(agentName, toolName, errorType string) {
	m.ToolCallErrors.WithLabelValues(agentName, toolName, errorType).Inc()
}

// RecordError 记录错误
func (m *Metrics) RecordError(agentName, modelName, errorType string) {
	m.ErrorCount.WithLabelValues(agentName, modelName, errorType).Inc()
}

// RecordCacheHit 记录缓存命中
func (m *Metrics) RecordCacheHit(agentName, cacheType string) {
	m.CacheHitCount.WithLabelValues(agentName, cacheType).Inc()
}

// RecordCacheMiss 记录缓存未命中
func (m *Metrics) RecordCacheMiss(agentName, cacheType string) {
	m.CacheMissCount.WithLabelValues(agentName, cacheType).Inc()
}

// RecordRAGRetrieval 记录RAG检索
func (m *Metrics) RecordRAGRetrieval(agentName, retrievalType string, duration time.Duration) {
	m.RAGRetrievalCount.WithLabelValues(agentName, retrievalType).Inc()
	m.RAGRetrievalTime.WithLabelValues(agentName, retrievalType).Observe(duration.Seconds())
}

// SetKnowledgeCount 设置知识库文档数量
func (m *Metrics) SetKnowledgeCount(agentName, knowledgeBase string, count float64) {
	m.RAGKnowledgeCount.WithLabelValues(agentName, knowledgeBase).Set(count)
}

// RecordReasoning 记录推理操作
func (m *Metrics) RecordReasoning(agentName, reasoningType string, duration time.Duration) {
	m.ReasoningCount.WithLabelValues(agentName, reasoningType).Inc()
	m.ReasoningTime.WithLabelValues(agentName, reasoningType).Observe(duration.Seconds())
}

// HelperFunctions 辅助函数
func GetStatusFromError(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func GetCacheStatus(hit bool) string {
	if hit {
		return "hit"
	}
	return "miss"
}
