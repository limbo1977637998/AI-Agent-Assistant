# 阶段五测试清单 - 高级 RAG 模式

**测试阶段**: 第五阶段 - 高级 RAG 模式
**创建时间**: 2026-01-28 21:05
**Tag**: v0.5.0

---

## 📋 测试概述

本测试清单涵盖了高级 RAG 模式的所有核心功能，包括：
- Enhanced Graph RAG（增强版图谱检索）
- Enhanced Self-RAG（增强版自我反思检索）
- Corrective RAG（纠错式检索）
- Agentic RAG（代理式检索）
- RAG Orchestrator（统一编排器）

---

## 🔧 环境准备

### 1. 依赖检查

```bash
# 检查 Go 版本
go version
# 要求: >= 1.21

# 检查项目依赖
go mod verify
go mod tidy

# 编译检查
go build ./...
```

### 2. 配置文件

创建测试配置文件 `config_test.yaml`:

```yaml
llm:
  provider: "openai" # 或其他支持的 provider
  model: "gpt-4"
  api_key: "your-api-key"
  temperature: 0.7
  max_tokens: 2000

rag:
  vector_store:
    type: "chroma"
    host: "localhost"
    port: 8000

  graph_store:
    type: "neo4j"
    uri: "bolt://localhost:7687"
    username: "neo4j"
    password: "password"

server:
  port: 8080
  log_level: "debug"
```

---

## ✅ 编译测试

### 1. 全量编译

```bash
# 编译所有包
go build ./...

# 预期结果：编译成功，无错误
```

### 2. 特定包编译

```bash
# Graph RAG
go build ./internal/rag/graph/...

# Adaptive RAG
go build ./internal/rag/adaptive/...

# Advanced RAG
go build ./internal/rag/advanced/...
```

### 3. 编译输出验证

```bash
# 查看编译产物
ls -lh ./ai-agent-assistant

# 预期：生成可执行文件
```

---

## 🧪 单元测试

### Graph RAG 测试

创建测试文件 `internal/rag/graph/enhanced_graph_rag_test.go`:

```go
package graph

import (
    "context"
    "testing"

    "ai-agent-assistant/internal/rag/adaptive"
)

// MockLLM 模拟 LLM
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    return "Mock response for testing", nil
}

func TestEnhancedGraphRAG_Creation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()

    rag, err := NewEnhancedGraphRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Graph RAG: %v", err)
    }

    if rag == nil {
        t.Fatal("Expected non-nil EnhancedGraphRAG")
    }
}

func TestEnhancedGraphRAG_GlobalSearch(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()
    rag, err := NewEnhancedGraphRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Graph RAG: %v", err)
    }

    // 创建测试图谱
    graph := &KnowledgeGraph{
        Entities: []*Entity{
            {ID: "1", Name: "User", Type: "Entity", Description: "用户实体"},
            {ID: "2", Name: "Order", Type: "Entity", Description: "订单实体"},
        },
        Relations: []*Relation{
            {From: "1", To: "2", Type: "places"},
        },
    }

    ctx := context.Background()
    contexts, err := rag.EnhancedGlobalSearch(ctx, graph, "系统架构", 10)

    if err != nil {
        t.Fatalf("Global search failed: %v", err)
    }

    if len(contexts) == 0 {
        t.Error("Expected at least one context")
    }

    t.Logf("Retrieved %d contexts", len(contexts))
    for i, ctx := range contexts {
        t.Logf("Context %d: %s", i+1, ctx)
    }
}

func TestEnhancedGraphRAG_LocalSearch(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()
    rag, err := NewEnhancedGraphRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Graph RAG: %v", err)
    }

    graph := &KnowledgeGraph{
        Entities: []*Entity{
            {ID: "1", Name: "User", Type: "Entity"},
            {ID: "2", Name: "Order", Type: "Entity"},
            {ID: "3", Name: "Product", Type: "Entity"},
        },
        Relations: []*Relation{
            {From: "1", To: "2", Type: "places"},
            {From: "2", To: "3", Type: "contains"},
        },
    }

    ctx := context.Background()
    contexts, err := rag.EnhancedLocalSearch(ctx, graph, "User和Order的关系", 10)

    if err != nil {
        t.Fatalf("Local search failed: %v", err)
    }

    if len(contexts) == 0 {
        t.Error("Expected at least one context")
    }
}

func TestEnhancedGraphRAG_PathBasedSearch(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()
    rag, err := NewEnhancedGraphRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Graph RAG: %v", err)
    }

    graph := &KnowledgeGraph{
        Entities: []*Entity{
            {ID: "1", Name: "User", Type: "Entity"},
            {ID: "2", Name: "Order", Type: "Entity"},
            {ID: "3", Name: "Product", Type: "Entity"},
        },
        Relations: []*Relation{
            {From: "1", To: "2", Type: "places"},
            {From: "2", To: "3", Type: "contains"},
        },
    }

    ctx := context.Background()
    contexts, err := rag.PathBasedSearch(ctx, graph, "从User到Product的路径", 10)

    if err != nil {
        t.Fatalf("Path search failed: %v", err)
    }

    t.Logf("Path search returned %d contexts", len(contexts))
}
```

**运行测试**:

```bash
# 运行 Graph RAG 测试
go test ./internal/rag/graph/... -v

# 运行特定测试
go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_GlobalSearch -v

# 查看覆盖率
go test ./internal/rag/graph/... -cover -coverprofile=graph_coverage.out
go tool cover -html=graph_coverage.out -o graph_coverage.html
```

### Self-RAG 测试

创建测试文件 `internal/rag/adaptive/enhanced_self_rag_test.go`:

```go
package adaptive

import (
    "context"
    "testing"
    "time"
)

func TestEnhancedSelfRAG_Creation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()

    rag, err := NewEnhancedSelfRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Self-RAG: %v", err)
    }

    if rag == nil {
        t.Fatal("Expected non-nil EnhancedSelfRAG")
    }
}

func TestEnhancedSelfRAG_EnhancedRetrieve(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    rag, err := NewEnhancedSelfRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Self-RAG: %v", err)
    }

    ctx := context.Background()
    docs, metrics, err := rag.EnhancedRetrieve(ctx, "如何优化数据库性能", 5)

    if err != nil {
        t.Fatalf("Enhanced retrieve failed: %v", err)
    }

    if len(docs) == 0 {
        t.Error("Expected at least one document")
    }

    if metrics == nil {
        t.Fatal("Expected metrics")
    }

    if metrics.OverallScore < 0 || metrics.OverallScore > 1 {
        t.Errorf("OverallScore out of range: %.2f", metrics.OverallScore)
    }

    t.Logf("Quality Metrics:")
    t.Logf("  Relevance: %.2f", metrics.RelevanceScore)
    t.Logf("  Coverage: %.2f", metrics.CoverageScore)
    t.Logf("  Accuracy: %.2f", metrics.AccuracyScore)
    t.Logf("  Completeness: %.2f", metrics.CompletenessScore)
    t.Logf("  Overall: %.2f", metrics.OverallScore)
}

func TestEnhancedSelfRAG_DynamicThreshold(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    config.DynamicThresholding = true
    rag, _ := NewEnhancedSelfRAG(llm, config)

    // 记录性能历史
    rag.recordPerformance("测试查询", 0.7, 5, 100*time.Millisecond)
    rag.recordPerformance("测试查询", 0.75, 5, 120*time.Millisecond)

    threshold := rag.calculateDynamicThreshold("测试查询", 1)

    if threshold < 0.5 || threshold > 0.95 {
        t.Errorf("Dynamic threshold out of range: %.2f", threshold)
    }

    t.Logf("Calculated dynamic threshold: %.2f", threshold)
}

func TestEnhancedSelfRAG_PerformanceHistory(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    rag, _ := NewEnhancedSelfRAG(llm, config)

    query := "性能测试查询"
    rag.recordPerformance(query, 0.8, 10, 150*time.Millisecond)

    history := rag.GetPerformanceHistory(query)
    if history == nil {
        t.Error("Expected non-nil history")
    } else {
        t.Logf("Performance history length: %d", len(history))
    }
}
```

**运行测试**:

```bash
# 运行 Self-RAG 测试
go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG -v

# 查看覆盖率
go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG -cover -coverprofile=selfrag_coverage.out
```

### Corrective RAG 测试

```go
package adaptive

import (
    "context"
    "testing"
)

func TestCorrectiveRAG_Creation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()

    rag, err := NewCorrectiveRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Corrective RAG: %v", err)
    }

    if rag == nil {
        t.Fatal("Expected non-nil CorrectiveRAG")
    }
}

func TestCorrectiveRAG_RetrieveAndCorrect(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, err := NewCorrectiveRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Corrective RAG: %v", err)
    }

    ctx := context.Background()
    result, err := rag.RetrieveAndCorrect(ctx, "什么是深度学习", 10)

    if err != nil {
        t.Fatalf("RetrieveAndCorrect failed: %v", err)
    }

    if result.InitialAnswer == "" {
        t.Error("Expected initial answer")
    }

    if result.CorrectedAnswer == "" {
        t.Error("Expected corrected answer")
    }

    if result.FinalValidation == nil {
        t.Error("Expected final validation")
    }

    t.Logf("Initial Answer: %s", result.InitialAnswer)
    t.Logf("Corrected Answer: %s", result.CorrectedAnswer)
    t.Logf("Total Corrections: %d", result.TotalCorrections)
    t.Logf("Final Validation: %.2f", result.FinalValidation.OverallConfidence)
}

func TestCorrectiveRAG_FactCheck(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, _ := NewCorrectiveRAG(llm, config)

    ctx := context.Background()
    statement := "机器学习是人工智能的一个分支"
    contexts := []string{
        "机器学习是AI的重要分支",
        "深度学习是机器学习的子集",
    }

    result, err := rag.factChecker.CheckFact(ctx, statement, contexts)
    if err != nil {
        t.Fatalf("Fact check failed: %v", err)
    }

    if result.Confidence < 0 || result.Confidence > 1 {
        t.Errorf("Confidence out of range: %.2f", result.Confidence)
    }

    t.Logf("Fact Check Result: IsFactual=%v, Confidence=%.2f",
        result.IsFactual, result.Confidence)
}

func TestCorrectiveRAG_LogicValidation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, _ := NewCorrectiveRAG(llm, config)

    ctx := context.Background()
    reasoning := "首先分析问题，然后找到解决方案，最后验证结果"
    contexts := []string{"问题分析", "解决方案"}

    result, err := rag.logicValidator.ValidateLogic(ctx, reasoning, contexts)
    if err != nil {
        t.Fatalf("Logic validation failed: %v", err)
    }

    t.Logf("Logic Validation: IsValid=%v, Confidence=%.2f",
        result.IsValid, result.Confidence)
}

func TestCorrectiveRAG_ConsistencyCheck(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, _ := NewCorrectiveRAG(llm, config)

    ctx := context.Background()
    answer := "根据上下文，答案是这样"
    contexts := []string{"上下文1", "上下文2"}

    result, err := rag.consistencyChecker.CheckConsistency(ctx, answer, contexts)
    if err != nil {
        t.Fatalf("Consistency check failed: %v", err)
    }

    t.Logf("Consistency Check: IsConsistent=%v, Confidence=%.2f",
        result.IsConsistent, result.Confidence)
}
```

**运行测试**:

```bash
go test ./internal/rag/adaptive/... -run TestCorrectiveRAG -v
```

### Agentic RAG 测试

```go
package adaptive

import (
    "context"
    "testing"
)

func TestAgenticRAG_Creation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultAgenticRAGConfig()

    rag, err := NewAgenticRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Agentic RAG: %v", err)
    }

    if rag == nil {
        t.Fatal("Expected non-nil AgenticRAG")
    }
}

func TestAgenticRAG_Query_ReactMode(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultAgenticRAGConfig()
    config.Mode = "react"
    config.MaxIterations = 5

    rag, err := NewAgenticRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Agentic RAG: %v", err)
    }

    ctx := context.Background()
    result, err := rag.Query(ctx, "分析用户行为数据")

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.Answer == "" {
        t.Error("Expected answer")
    }

    if len(result.Thoughts) == 0 {
        t.Error("Expected at least one thought")
    }

    if len(result.Actions) == 0 {
        t.Error("Expected at least one action")
    }

    if len(result.Observations) == 0 {
        t.Error("Expected at least one observation")
    }

    t.Logf("Query completed in %d iterations", result.Iterations)
    t.Logf("Answer: %s", result.Answer)
}

func TestAgenticRAG_Query_PlanAndExecuteMode(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultAgenticRAGConfig()
    config.Mode = "plan_execute"
    config.MaxIterations = 10

    rag, _ := NewAgenticRAG(llm, config)

    ctx := context.Background()
    result, err := rag.Query(ctx, "多步骤分析任务")

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.Answer == "" {
        t.Error("Expected answer")
    }

    t.Logf("Plan-and-Execute completed: %s", result.Answer)
}

func TestAgenticRAG_Tools(t *testing.T) {
    tools := []AgentTool{
        &VectorSearchTool{},
        &GraphSearchTool{},
        &HybridSearchTool{},
        &KnowledgeQueryTool{},
    }

    ctx := context.Background()
    for _, tool := range tools {
        t.Logf("Testing tool: %s", tool.Name())

        result, err := tool.Execute(ctx, "测试查询")
        if err != nil {
            t.Errorf("Tool %s execution failed: %v", tool.Name(), err)
        }

        if result == "" {
            t.Errorf("Tool %s returned empty result", tool.Name())
        }

        if !tool.ValidateInput("测试查询") {
            t.Errorf("Tool %s rejected valid input", tool.Name())
        }

        if tool.ValidateInput("") {
            t.Errorf("Tool %s accepted invalid input", tool.Name())
        }
    }
}
```

**运行测试**:

```bash
go test ./internal/rag/adaptive/... -run TestAgenticRAG -v
```

### Orchestrator 测试

创建测试文件 `internal/rag/advanced/orchestrator_test.go`:

```go
package advanced

import (
    "context"
    "testing"

    "ai-agent-assistant/internal/rag/adaptive"
    "ai-agent-assistant/internal/rag/graph"
)

type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    return "Mock response", nil
}

func TestOrchestrator_Creation(t *testing.T) {
    config := DefaultOrchestratorConfig()

    orchestrator, err := NewAdvancedRAGOrchestrator(config)
    if err != nil {
        t.Fatalf("Failed to create orchestrator: %v", err)
    }

    if orchestrator == nil {
        t.Fatal("Expected non-nil orchestrator")
    }
}

func TestOrchestrator_QueryAnalysis(t *testing.T) {
    config := DefaultOrchestratorConfig()
    orchestrator, _ := NewAdvancedRAGOrchestrator(config)

    ctx := context.Background()

    testCases := []struct {
        query          string
        expectedType   string
        expectedComplexity string
    }{
        {"系统整体架构", "global", "medium"},
        {"为什么会出现这个问题", "reasoning", "medium"},
        {"这是一个非常长的查询，包含了大量的详细信息", "general", "complex"},
    }

    for _, tc := range testCases {
        analysis := orchestrator.analyzeQuery(ctx, tc.query)
        t.Logf("Query: %s", tc.query)
        t.Logf("  Type: %s (expected: %s)", analysis.QueryType, tc.expectedType)
        t.Logf("  Complexity: %s (expected: %s)", analysis.Complexity, tc.expectedComplexity)
    }
}

func TestOrchestrator_ModeSelector(t *testing.T) {
    selector := &DefaultModeSelector{}
    ctx := context.Background()

    testCases := []struct {
        name     string
        analysis *QueryAnalysis
        expected string
    }{
        {
            name: "Global query",
            analysis: &QueryAnalysis{
                QueryType:     "global",
                Complexity:    "medium",
                RequiresGraph: true,
            },
            expected: "enhanced_graph",
        },
        {
            name: "Complex reasoning query",
            analysis: &QueryAnalysis{
                QueryType:          "reasoning",
                Complexity:         "complex",
                RequiresReasoning:  true,
            },
            expected: "agentic",
        },
        {
            name: "Simple query",
            analysis: &QueryAnalysis{
                QueryType:  "general",
                Complexity: "simple",
            },
            expected: "enhanced_self",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mode, err := selector.SelectMode(ctx, "测试查询", tc.analysis)
            if err != nil {
                t.Fatalf("Mode selection failed: %v", err)
            }

            if mode != tc.expected {
                t.Errorf("Expected mode %s, got %s", tc.expected, mode)
            }

            t.Logf("Selected mode: %s", mode)
        })
    }
}

func TestOrchestrator_SpecificMode(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultOrchestratorConfig()

    orchestrator, _ := NewAdvancedRAGOrchestrator(config)

    // 注入 RAG 实现
    graphRAG, _ := graph.NewEnhancedGraphRAG(llm, graph.DefaultGraphRAGConfig())
    orchestrator.SetEnhancedGraphRAG(graphRAG)

    ctx := context.Background()
    result, err := orchestrator.Query(ctx, "测试查询", "enhanced_graph")

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.Mode != "enhanced_graph_rag" {
        t.Errorf("Expected mode 'enhanced_graph_rag', got '%s'", result.Mode)
    }

    t.Logf("Query result: %s", result.Answer)
}
```

**运行测试**:

```bash
go test ./internal/rag/advanced/... -v
```

---

## 🌐 集成测试

### 1. 启动服务

```bash
# 启动 API 服务器
go run cmd/server/main.go --config config_test.yaml

# 预期输出：
# Server started on port 8080
# RAG modules initialized
```

### 2. Graph RAG API 测试

```bash
# 全局检索
curl -X POST http://localhost:8080/api/v1/rag/graph/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "系统整体架构概览",
    "mode": "global",
    "top_k": 10
  }'

# 预期响应：
# {
#   "code": 0,
#   "data": {
#     "query": "系统整体架构概览",
#     "answer": "...",
#     "mode": "enhanced_graph_rag"
#   }
# }
```

### 3. Self-RAG API 测试

```bash
curl -X POST http://localhost:8080/api/v1/rag/self/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何优化数据库性能",
    "top_k": 5
  }'

# 检查响应中的 quality_metrics 字段
```

### 4. Corrective RAG API 测试

```bash
curl -X POST http://localhost:8080/api/v1/rag/corrective/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是深度学习",
    "top_k": 10
  }'

# 检查响应中的 correction_history 和 final_validation 字段
```

### 5. Agentic RAG API 测试

```bash
curl -X POST http://localhost:8080/api/v1/rag/agentic/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析用户行为数据并给出优化建议",
    "mode": "react",
    "max_iterations": 10
  }'

# 检查响应中的 thoughts, actions, observations 字段
```

### 6. Orchestrator API 测试

```bash
# 自动模式选择
curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "系统整体架构",
    "mode": "auto"
  }'

# 检查 mode_used 和 query_analysis 字段
```

---

## 📊 性能测试

### 1. 基准测试

```bash
# Graph RAG 基准测试
go test ./internal/rag/graph/... -bench=. -benchmem

# Self-RAG 基准测试
go test ./internal/rag/adaptive/... -bench=EnhancedSelfRAG -benchmem

# Agentic RAG 基准测试
go test ./internal/rag/adaptive/... -bench=AgenticRAG -benchmem
```

### 2. 负载测试

```bash
# 安装 Apache Bench
# macOS: brew install httpd
# Ubuntu: apt-get install apache2-utils

# Graph RAG 负载测试
ab -n 1000 -c 10 -p graph_payload.json -T application/json \
   http://localhost:8080/api/v1/rag/graph/query

# Self-RAG 负载测试
ab -n 1000 -c 10 -p self_payload.json -T application/json \
   http://localhost:8080/api/v1/rag/self/query

# Agentic RAG 负载测试
ab -n 100 -c 5 -p agentic_payload.json -T application/json \
   http://localhost:8080/api/v1/rag/agentic/query
```

### 3. 性能指标验证

| 模式 | 目标 P50 延迟 | 目标 P99 延迟 | 目标 QPS |
|------|--------------|--------------|----------|
| Graph RAG | < 200ms | < 500ms | > 50 |
| Self-RAG | < 300ms | < 800ms | > 30 |
| Corrective RAG | < 400ms | < 1000ms | > 20 |
| Agentic RAG | < 500ms | < 1500ms | > 10 |

---

## ✅ 验收标准

### 功能完整性

- [ ] 所有四种 RAG 模式均能正常工作
- [ ] Orchestrator 能正确选择模式
- [ ] 各模式返回结果格式正确
- [ ] 质量指标在合理范围内（0-1）

### 测试覆盖率

```bash
# 查看总体覆盖率
go test ./internal/rag/... -cover -coverprofile=coverage.out
go tool cover -func=coverage.out

# 预期：总体覆盖率 > 60%
```

### 性能要求

- [ ] Graph RAG: P99 < 500ms
- [ ] Self-RAG: P99 < 800ms
- [ ] Corrective RAG: P99 < 1000ms
- [ ] Agentic RAG: P99 < 1500ms

### 稳定性

- [ ] 连续运行 1000 次查询无崩溃
- [ ] 并发 10 个请求无错误
- [ ] 内存无泄漏

---

## 🐛 已知问题

### 1. 限制和注意事项

- Graph RAG 需要预构建知识图谱
- Agentic RAG 在复杂查询时可能需要多次迭代
- Corrective RAG 依赖 LLM 质量，可能产生不一致的纠错结果

### 2. 待优化项

- [ ] 优化 Graph RAG 的社区检测算法
- [ ] 增强 Self-RAG 的动态阈值调整策略
- [ ] 改进 Agentic RAG 的工具选择逻辑
- [ ] 优化 Orchestrator 的模式选择算法

---

## 📝 测试报告模板

### 测试执行记录

**测试人员**: ___________
**测试时间**: ___________
**环境**: ___________

| 测试项 | 状态 | 备注 |
|--------|------|------|
| 编译测试 | ☐ 通过 / ☐ 失败 | |
| Graph RAG 单元测试 | ☐ 通过 / ☐ 失败 | |
| Self-RAG 单元测试 | ☐ 通过 / ☐ 失败 | |
| Corrective RAG 单元测试 | ☐ 通过 / ☐ 失败 | |
| Agentic RAG 单元测试 | ☐ 通过 / ☐ 失败 | |
| Orchestrator 单元测试 | ☐ 通过 / ☐ 失败 | |
| 集成测试 | ☐ 通过 / ☐ 失败 | |
| 性能测试 | ☐ 通过 / ☐ 失败 | |

### 问题记录

| 问题ID | 描述 | 严重程度 | 状态 |
|--------|------|---------|------|
| 1 | | ☐ 严重 / ☐ 一般 / ☐ 轻微 | ☐ 待修复 / ☐ 已修复 |
| 2 | | ☐ 严重 / ☐ 一般 / ☐ 轻微 | ☐ 待修复 / ☐ 已修复 |

### 测试结论

☐ **通过** - 所有关键功能正常，性能达标
☐ **有条件通过** - 存在次要问题，不影响核心功能
☐ **不通过** - 存在严重问题，需要修复后重新测试

---

## 🚀 下一步

1. **完成单元测试**: 根据上述测试用例实现完整的单元测试
2. **执行集成测试**: 启动服务并执行 API 测试
3. **性能优化**: 根据性能测试结果进行优化
4. **生产准备**: 准备生产环境配置和监控

---

**测试清单结束**

如有问题，请参考详细文档：`PHASE5_ADVANCED_RAG_20260128_2105.md`
