# 高级 RAG 模式实现文档

**创建时间**: 2026-01-28 21:05
**阶段**: 第五阶段 - 高级 RAG 模式开发
**状态**: ✅ 已完成

---

## 📋 目录

1. [概述](#概述)
2. [Graph RAG](#graph-rag)
3. [Self-RAG](#self-rag)
4. [Corrective RAG](#corrective-rag)
5. [Agentic RAG](#agentic-rag)
6. [RAG 编排器](#rag-编排器)
7. [单元测试](#单元测试)
8. [API 接口](#api-接口)
9. [测试清单](#测试清单)

---

## 概述

本阶段实现了四种高级 RAG 模式，并通过编排器统一管理：

### 实现的高级模式

| 模式 | 文件位置 | 核心特性 |
|------|---------|---------|
| **Enhanced Graph RAG** | `internal/rag/graph/enhanced_graph_rag.go` | 动态社区摘要、层次化检索、实体评分、路径查找 |
| **Enhanced Self-RAG** | `internal/rag/adaptive/enhanced_self_rag.go` | 动态阈值调整、多维度评估、自适应策略 |
| **Corrective RAG** | `internal/rag/adaptive/corrective_rag.go` | 事实核查、逻辑验证、一致性检查 |
| **Agentic RAG** | `internal/rag/adaptive/agentic_rag.go` | ReAct 模式、计划执行、反思机制 |
| **Orchestrator** | `internal/rag/advanced/orchestrator.go` | 模式选择、模式切换、结果融合 |

---

## Graph RAG

### 核心特性

Enhanced Graph RAG 基于 Microsoft Research 2024 年的论文，实现了以下增强功能：

#### 1. 动态社区摘要
- 根据查询上下文生成针对性摘要
- 社区权重计算（关键词匹配度）
- 层次化社区遍历

#### 2. 实体重要性评分
- 名称匹配度
- 类型匹配度
- 描述相关性

#### 3. 多跳关系检索
- BFS 遍历实体关系
- 相关性过滤（阈值 0.5）
- 最多支持 N 跳检索

#### 4. 路径查找
- 最短路径算法（BFS）
- 实体间关系路径构建
- 路径上下文生成

### 主要方法

```go
// 增强版全局检索
func (egr *EnhancedGraphRAG) EnhancedGlobalSearch(
    ctx context.Context,
    graph *KnowledgeGraph,
    query string,
    topK int,
) ([]string, error)

// 增强版局部检索
func (egr *EnhancedGraphRAG) EnhancedLocalSearch(
    ctx context.Context,
    graph *KnowledgeGraph,
    query string,
    topK int,
) ([]string, error)

// 基于路径的检索
func (egr *EnhancedGraphRAG) PathBasedSearch(
    ctx context.Context,
    graph *KnowledgeGraph,
    query string,
    topK int,
) ([]string, error)
```

### 使用示例

```go
// 创建 Enhanced Graph RAG
config := graph.DefaultGraphRAGConfig()
enhancedGraphRAG, err := graph.NewEnhancedGraphRAG(llm, config)

// 全局检索（适合概览性查询）
contexts, err := enhancedGraphRAG.EnhancedGlobalSearch(ctx, knowledgeGraph, "系统架构概览", 10)

// 局部检索（适合实体关系查询）
contexts, err := enhancedGraphRAG.EnhancedLocalSearch(ctx, knowledgeGraph, "用户和订单的关系", 10)

// 路径检索（适合实体间路径查询）
contexts, err := enhancedGraphRAG.PathBasedSearch(ctx, knowledgeGraph, "从用户到产品的路径", 10)
```

### 配置选项

```go
type GraphRAGConfig struct {
    CommunityDetectionAlgo string // "louvain", "leiden"
    MinCommunitySize       int
    MaxHierarchyDepth      int
}
```

---

## Self-RAG

### 核心特性

Enhanced Self-RAG 实现了自我反思机制，支持动态调整检索策略：

#### 1. 动态阈值调整
- 基于历史性能动态调整质量阈值
- 最小改进率配置（默认 10%）
- 阈值范围限制（0.5 - 0.95）

#### 2. 多维度质量评估
- **相关性**: 查询词匹配度
- **覆盖率**: 文档数量和多样性
- **准确性**: 关键词包含度
- **完整性**: 答案长度评估

#### 3. 自适应策略
- 语义检索（相关性低时）
- 全局检索（覆盖率低时）
- 混合检索（准确性低时）

#### 4. 性能追踪
- 查询历史记录
- 分数趋势分析
- 延迟监控

### 主要方法

```go
// 增强版检索（带自我反思）
func (esr *EnhancedSelfRAG) EnhancedRetrieve(
    ctx context.Context,
    query string,
    initialTopK int,
) ([]string, *QualityMetrics, error)

// 评估质量
func (esr *EnhancedSelfRAG) evaluateQuality(
    ctx context.Context,
    query string,
    docs []string,
    answer string,
) (*QualityMetrics, error)

// 获取性能历史
func (esr *EnhancedSelfRAG) GetPerformanceHistory(
    query string,
) []QueryPerformance
```

### 使用示例

```go
// 创建 Enhanced Self-RAG
config := adaptive.DefaultEnhancedSelfRAGConfig()
config.DynamicThresholding = true
config.MultiDimensionalEval = true
config.AdaptiveStrategy = true

enhancedSelfRAG, err := adaptive.NewEnhancedSelfRAG(llm, config)

// 执行检索（自动反思和优化）
docs, metrics, err := enhancedSelfRAG.EnhancedRetrieve(ctx, "如何优化数据库性能", 5)

// 查看质量指标
fmt.Printf("相关性: %.2f\n", metrics.RelevanceScore)
fmt.Printf("覆盖率: %.2f\n", metrics.CoverageScore)
fmt.Printf("准确性: %.2f\n", metrics.AccuracyScore)
fmt.Printf("完整性: %.2f\n", metrics.CompletenessScore)
fmt.Printf("综合得分: %.2f\n", metrics.OverallScore)

// 查看改进建议
for _, issue := range metrics.Issues {
    fmt.Printf("问题: %s\n", issue)
}
for _, suggestion := range metrics.Suggestions {
    fmt.Printf("建议: %s\n", suggestion)
}
```

### 质量指标说明

| 指标 | 计算方式 | 阈值 |
|------|---------|------|
| RelevanceScore | 查询词匹配度 / 总词数 | ≥ 0.6 |
| CoverageScore | 文档数 / 期望文档数 (5) | ≥ 0.6 |
| AccuracyScore | 关键词匹配数 / 总关键词 | ≥ 0.6 |
| CompletenessScore | 答案长度 (50-500 字) | ≥ 0.6 |
| OverallScore | 加权平均 (0.3, 0.2, 0.3, 0.2) | ≥ 0.6 |

---

## Corrective RAG

### 核心特性

Corrective RAG 实现了三种纠错机制，主动检测和修正错误：

#### 1. 事实核查 (Fact Checking)
- 检测幻觉和虚假信息
- 上下文事实验证
- 纠正建议生成

#### 2. 逻辑验证 (Logic Validation)
- 推理链连贯性检查
- 识别逻辑跳跃和谬误
- 推理步骤评估

#### 3. 一致性检查 (Consistency Checking)
- 上下文一致性验证
- 引用准确性检查
- 冲突信息识别

#### 4. 迭代纠错
- 最多 3 轮纠错（可配置）
- 置信度阈值（默认 0.8）
- 纠错历史记录

### 主要方法

```go
// 检索并纠错
func (crag *CorrectiveRAG) RetrieveAndCorrect(
    ctx context.Context,
    query string,
    topK int,
) (*CorrectedResult, error)

// 事实核查
func (c *FactChecker) CheckFact(
    ctx context.Context,
    statement string,
    contexts []string,
) (*FactCheckResult, error)

// 逻辑验证
func (v *LogicValidator) ValidateLogic(
    ctx context.Context,
    reasoning string,
    contexts []string,
) (*LogicValidationResult, error)

// 一致性检查
func (c *ConsistencyChecker) CheckConsistency(
    ctx context.Context,
    answer string,
    contexts []string,
) (*ConsistencyResult, error)
```

### 使用示例

```go
// 创建 Corrective RAG
config := adaptive.DefaultCorrectiveRAGConfig()
config.EnableFactCheck = true
config.EnableLogicValidation = true
config.EnableConsistencyCheck = true
config.MaxCorrectionRounds = 3
config.ConfidenceThreshold = 0.8

correctiveRAG, err := adaptive.NewCorrectiveRAG(llm, config)

// 执行检索并自动纠错
result, err := correctiveRAG.RetrieveAndCorrect(ctx, "什么是深度学习", 10)

// 查看纠错结果
fmt.Printf("初始答案: %s\n", result.InitialAnswer)
fmt.Printf("纠错后答案: %s\n", result.CorrectedAnswer)
fmt.Printf("纠错轮数: %d\n", result.TotalCorrections)

// 查看纠错历史
for _, round := range result.CorrectionHistory {
    fmt.Printf("\n第 %d 轮纠错:\n", round.Round)
    for _, correction := range round.Corrections {
        fmt.Printf("  [%s] %s\n", correction.Type, correction.Description)
        fmt.Printf("  建议: %s\n", correction.Suggestion)
    }
}

// 查看最终验证结果
fmt.Printf("\n最终验证:\n")
fmt.Printf("  综合置信度: %.2f\n", result.FinalValidation.OverallConfidence)
fmt.Printf("  事实核查: %.2f\n", result.FinalValidation.FactCheckConfidence)
fmt.Printf("  逻辑验证: %.2f\n", result.FinalValidation.LogicCheckConfidence)
fmt.Printf("  一致性检查: %.2f\n", result.FinalValidation.ConsistencyConfidence)
fmt.Printf("  是否通过: %v\n", result.FinalValidation.Passed)
```

### 纠错类型

| 类型 | 检查内容 | 触发条件 |
|------|---------|---------|
| Fact | 事实准确性、幻觉检测 | 上下文不支持 |
| Logic | 推理逻辑、步骤连贯性 | 逻辑跳跃/谬误 |
| Consistency | 上下文一致性、引用准确性 | 信息冲突/矛盾 |
| Completeness | 信息完整性、充分性 | 关键信息缺失 |

---

## Agentic RAG

### 核心特性

Agentic RAG 实现了智能代理系统，支持自主决策和工具使用：

#### 1. ReAct 模式 (Reasoning + Acting)
- Thought → Action → Observation 循环
- 推理引导行动
- 观察反馈调整

#### 2. Plan-and-Execute 模式
- 前期规划步骤
- 依赖关系管理
- 顺序执行计划

#### 3. 反思机制 (Reflexion)
- 执行后反思
- 错误分析和改进
- 迭代优化

#### 4. 工具使用
- VectorSearchTool: 向量搜索
- GraphSearchTool: 图谱搜索
- HybridSearchTool: 混合搜索
- KnowledgeQueryTool: 知识库查询

### 主要方法

```go
// 执行代理式查询
func (ar *AgenticRAG) Query(
    ctx context.Context,
    query string,
) (*AgentResult, error)

// ReAct 模式
func (ar *AgenticRAG) reactMode(
    ctx context.Context,
    query string,
) (*AgentResult, error)

// Plan-and-Execute 模式
func (ar *AgenticRAG) planAndExecuteMode(
    ctx context.Context,
    query string,
) (*AgentResult, error)

// 反思模式
func (ar *AgenticRAG) reflexionMode(
    ctx context.Context,
    query string,
) (*AgentResult, error)
```

### 使用示例

```go
// 创建 Agentic RAG
config := adaptive.DefaultAgenticRAGConfig()
config.Mode = "react" // "react", "plan_execute", "reflexion"
config.MaxIterations = 10
config.EnableReflection = true

agenticRAG, err := adaptive.NewAgenticRAG(llm, config)

// 执行查询
result, err := agenticRAG.Query(ctx, "分析用户行为数据并给出优化建议")

// 查看思考过程
for i, thought := range result.Thoughts {
    fmt.Printf("[Thought %d] %s\n", i+1, thought.Content)
}

// 查看行动过程
for i, action := range result.Actions {
    fmt.Printf("[Action %d] 工具: %s, 输入: %s\n", i+1, action.Tool, action.Input)
}

// 查看观察结果
for i, obs := range result.Observations {
    fmt.Printf("[Observation %d] %s\n", i+1, obs.Content)
}

// 查看最终答案
fmt.Printf("\n答案: %s\n", result.Answer)
fmt.Printf("迭代次数: %d\n", result.Iterations)
fmt.Printf("置信度: %.2f\n", result.Confidence)
```

### 代理状态

```go
type AgentState struct {
    Query          string      // 查询内容
    CurrentStep    int         // 当前步骤
    Iterations     int         // 迭代次数
    Completed      bool        // 是否完成
    Observations   []Observation // 观察记录
    Thoughts       []Thought   // 思考记录
    Actions        []Action    // 行动记录
    Answer         string      // 最终答案
    Confidence     float64     // 置信度
}
```

### 可用工具

| 工具 | 描述 | 使用场景 |
|------|------|---------|
| VectorSearchTool | 向量搜索 | 语义相似度检索 |
| GraphSearchTool | 图谱搜索 | 实体关系检索 |
| HybridSearchTool | 混合搜索 | 综合检索 |
| KnowledgeQueryTool | 知识库查询 | 直接查询 |

---

## RAG 编排器

### 核心功能

Advanced RAG Orchestrator 统一管理所有高级 RAG 模式：

#### 1. 模式选择
- 基于查询特征自动选择最佳模式
- 支持手动指定模式
- 查询分析（类型、复杂度、领域）

#### 2. 模式切换
- 动态模式切换
- 性能监控
- 降级策略

#### 3. 结果融合
- 多模式并行执行
- 结果质量评分
- 最佳结果选择或融合

### 查询分析

```go
type QueryAnalysis struct {
    QueryType          string  // definition, procedure, reasoning, global, specific
    Complexity         string  // simple, medium, complex
    Domain             string  // technical, general, specific
    RequiresGraph      bool
    RequiresReasoning  bool
    Keywords           []string
}
```

### 模式选择策略

| 查询类型 | 推荐模式 | 理由 |
|---------|---------|------|
| Global/概览 | Enhanced Graph RAG | 社区摘要，层次检索 |
| Reasoning/推理 | Enhanced Graph RAG / Agentic RAG | 实体关系，多跳检索 |
| Complex + Technical | Corrective RAG | 高准确性要求 |
| Default | Enhanced Self-RAG | 平衡性能和质量 |

### 使用示例

```go
// 创建编排器
config := advanced.DefaultOrchestratorConfig()
config.DefaultMode = "auto"
config.EnableAutoModeSelection = true
config.EnableModeSwitching = true
config.EnableResultFusion = true

orchestrator, err := advanced.NewAdvancedRAGOrchestrator(config)

// 注入各个 RAG 实现
orchestrator.SetEnhancedGraphRAG(enhancedGraphRAG)
orchestrator.SetEnhancedSelfRAG(enhancedSelfRAG)
orchestrator.SetCorrectiveRAG(correctiveRAG)
orchestrator.SetAgenticRAG(agenticRAG)

// 自动模式选择查询
result, err := orchestrator.Query(ctx, "系统整体架构是什么", "auto")

// 指定模式查询
result, err := orchestrator.Query(ctx, "如何优化性能", "agentic")

// 融合模式查询（执行所有模式并融合结果）
result, err := orchestrator.Query(ctx, "复杂查询", "fused")

// 查看结果
fmt.Printf("使用模式: %s\n", result.ModeUsed)
fmt.Printf("查询类型: %s\n", result.QueryAnalysis.QueryType)
fmt.Printf("复杂度: %s\n", result.QueryAnalysis.Complexity)
fmt.Printf("答案: %s\n", result.Answer)
fmt.Printf("延迟: %v\n", result.Latency)

// 查看模式特定信息
switch result.Mode {
case "enhanced_graph_rag":
    fmt.Printf("图谱层次: %v\n", result.GraphHierarchy)
case "enhanced_self_rag":
    fmt.Printf("质量得分: %.2f\n", result.QualityMetrics.OverallScore)
case "corrective_rag":
    fmt.Printf("纠错次数: %d\n", result.TotalCorrections)
    fmt.Printf("最终验证: %.2f\n", result.FinalValidation.OverallConfidence)
case "agentic_rag":
    fmt.Printf("迭代次数: %d\n", result.Iterations)
    fmt.Printf("思考步骤: %d\n", len(result.Thoughts))
}
```

### 配置选项

```go
type OrchestratorConfig struct {
    DefaultMode             string  // 默认模式: "auto", "graph_rag", "self_rag", "corrective", "agentic"
    EnableAutoModeSelection  bool    // 启用自动模式选择
    EnableModeSwitching      bool    // 启用模式切换
    EnableResultFusion       bool    // 启用结果融合
    ModeTimeout              int64   // 单个模式超时（毫秒）
}
```

---

## 单元测试

### 测试文件结构

```
internal/rag/
├── graph/
│   ├── enhanced_graph_rag.go
│   └── enhanced_graph_rag_test.go
├── adaptive/
│   ├── enhanced_self_rag.go
│   ├── enhanced_self_rag_test.go
│   ├── corrective_rag.go
│   ├── corrective_rag_test.go
│   ├── agentic_rag.go
│   └── agentic_rag_test.go
└── advanced/
    ├── orchestrator.go
    └── orchestrator_test.go
```

### Graph RAG 测试

```go
// internal/rag/graph/enhanced_graph_rag_test.go

package graph

import (
    "context"
    "testing"
)

func TestEnhancedGraphRAG_GlobalSearch(t *testing.T) {
    // 创建 mock LLM
    llm := &MockLLM{}

    // 创建 Enhanced Graph RAG
    config := DefaultGraphRAGConfig()
    rag, err := NewEnhancedGraphRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Graph RAG: %v", err)
    }

    // 创建测试图谱
    graph := &KnowledgeGraph{
        Entities: []*Entity{
            {ID: "1", Name: "User", Type: "Entity"},
            {ID: "2", Name: "Order", Type: "Entity"},
        },
        Relations: []*Relation{
            {From: "1", To: "2", Type: "places"},
        },
    }

    // 测试全局检索
    ctx := context.Background()
    contexts, err := rag.EnhancedGlobalSearch(ctx, graph, "系统架构", 10)
    if err != nil {
        t.Fatalf("Global search failed: %v", err)
    }

    if len(contexts) == 0 {
        t.Error("Expected at least one context")
    }
}

func TestEnhancedGraphRAG_LocalSearch(t *testing.T) {
    // 测试局部检索
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()
    rag, _ := NewEnhancedGraphRAG(llm, config)

    graph := createTestGraph()
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
    // 测试路径检索
    llm := &MockLLM{}
    config := DefaultGraphRAGConfig()
    rag, _ := NewEnhancedGraphRAG(llm, config)

    graph := createTestGraph()
    ctx := context.Background()

    contexts, err := rag.PathBasedSearch(ctx, graph, "从User到Product的路径", 10)
    if err != nil {
        t.Fatalf("Path search failed: %v", err)
    }

    if len(contexts) == 0 {
        t.Error("Expected at least one path context")
    }
}

// Mock LLM 实现
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    return "Mock response", nil
}

func createTestGraph() *KnowledgeGraph {
    return &KnowledgeGraph{
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
}
```

### Self-RAG 测试

```go
// internal/rag/adaptive/enhanced_self_rag_test.go

package adaptive

import (
    "context"
    "testing"
)

func TestEnhancedSelfRAG_EnhancedRetrieve(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    rag, err := NewEnhancedSelfRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Enhanced Self-RAG: %v", err)
    }

    ctx := context.Background()
    docs, metrics, err := rag.EnhancedRetrieve(ctx, "如何优化数据库", 5)
    if err != nil {
        t.Fatalf("Enhanced retrieve failed: %v", err)
    }

    if len(docs) == 0 {
        t.Error("Expected at least one document")
    }

    if metrics == nil {
        t.Error("Expected metrics")
    }

    if metrics.OverallScore < 0 || metrics.OverallScore > 1 {
        t.Errorf("OverallScore out of range: %.2f", metrics.OverallScore)
    }
}

func TestEnhancedSelfRAG_QualityEvaluation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    config.MultiDimensionalEval = true
    rag, _ := NewEnhancedSelfRAG(llm, config)

    ctx := context.Background()
    docs := []string{"文档1", "文档2", "文档3"}
    answer := "这是一个答案"

    metrics, err := rag.evaluateQuality(ctx, "测试查询", docs, answer)
    if err != nil {
        t.Fatalf("Quality evaluation failed: %v", err)
    }

    if metrics.RelevanceScore < 0 || metrics.RelevanceScore > 1 {
        t.Errorf("RelevanceScore out of range: %.2f", metrics.RelevanceScore)
    }

    if metrics.CoverageScore < 0 || metrics.CoverageScore > 1 {
        t.Errorf("CoverageScore out of range: %.2f", metrics.CoverageScore)
    }
}

func TestEnhancedSelfRAG_DynamicThreshold(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultEnhancedSelfRAGConfig()
    config.DynamicThresholding = true
    rag, _ := NewEnhancedSelfRAG(llm, config)

    // 记录一些性能历史
    rag.recordPerformance("测试查询", 0.7, 5, 100*time.Millisecond)
    rag.recordPerformance("测试查询", 0.75, 5, 120*time.Millisecond)

    // 计算动态阈值
    threshold := rag.calculateDynamicThreshold("测试查询", 1)
    if threshold < 0.5 || threshold > 0.95 {
        t.Errorf("Dynamic threshold out of range: %.2f", threshold)
    }
}
```

### Corrective RAG 测试

```go
// internal/rag/adaptive/corrective_rag_test.go

package adaptive

import (
    "context"
    "testing"
)

func TestCorrectiveRAG_RetrieveAndCorrect(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, err := NewCorrectiveRAG(llm, config)
    if err != nil {
        t.Fatalf("Failed to create Corrective RAG: %v", err)
    }

    ctx := context.Background()
    result, err := rag.RetrieveAndCorrect(ctx, "什么是机器学习", 10)
    if err != nil {
        t.Fatalf("RetrieveAndCorrect failed: %v", err)
    }

    if result.InitialAnswer == "" {
        t.Error("Expected initial answer")
    }

    if result.CorrectedAnswer == "" {
        t.Error("Expected corrected answer")
    }

    if result.TotalCorrections < 0 || result.TotalCorrections > 3 {
        t.Errorf("Invalid correction count: %d", result.TotalCorrections)
    }

    if result.FinalValidation == nil {
        t.Error("Expected final validation")
    }
}

func TestCorrectiveRAG_FactCheck(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, _ := NewCorrectiveRAG(llm, config)

    ctx := context.Background()
    statement := "机器学习是人工智能的一个分支"
    contexts := []string{"机器学习是AI的子集", "深度学习是机器学习的子集"}

    result, err := rag.factChecker.CheckFact(ctx, statement, contexts)
    if err != nil {
        t.Fatalf("Fact check failed: %v", err)
    }

    if result.Confidence < 0 || result.Confidence > 1 {
        t.Errorf("Confidence out of range: %.2f", result.Confidence)
    }
}

func TestCorrectiveRAG_LogicValidation(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultCorrectiveRAGConfig()
    rag, _ := NewCorrectiveRAG(llm, config)

    ctx := context.Background()
    reasoning := "首先，我们需要分析问题。然后，我们找到解决方案。"
    contexts := []string{"问题分析", "解决方案"}

    result, err := rag.logicValidator.ValidateLogic(ctx, reasoning, contexts)
    if err != nil {
        t.Fatalf("Logic validation failed: %v", err)
    }

    if result.Confidence < 0 || result.Confidence > 1 {
        t.Errorf("Confidence out of range: %.2f", result.Confidence)
    }
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

    if result.Confidence < 0 || result.Confidence > 1 {
        t.Errorf("Confidence out of range: %.2f", result.Confidence)
    }
}
```

### Agentic RAG 测试

```go
// internal/rag/adaptive/agentic_rag_test.go

package adaptive

import (
    "context"
    "testing"
)

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

    if result.Iterations == 0 {
        t.Error("Expected at least one iteration")
    }
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
}

func TestAgenticRAG_Query_ReflexionMode(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultAgenticRAGConfig()
    config.Mode = "reflexion"
    config.EnableReflection = true
    config.MaxIterations = 5

    rag, _ := NewAgenticRAG(llm, config)

    ctx := context.Background()
    result, err := rag.Query(ctx, "需要反思的复杂查询")
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.Answer == "" {
        t.Error("Expected answer")
    }
}

func TestAgenticRAG_ToolExecution(t *testing.T) {
    tool := &VectorSearchTool{}

    ctx := context.Background()
    result, err := tool.Execute(ctx, "测试查询")
    if err != nil {
        t.Fatalf("Tool execution failed: %v", err)
    }

    if result == "" {
        t.Error("Expected tool result")
    }

    if !tool.ValidateInput("测试查询") {
        t.Error("Expected valid input")
    }

    if tool.ValidateInput("") {
        t.Error("Expected invalid input for empty string")
    }
}
```

### Orchestrator 测试

```go
// internal/rag/advanced/orchestrator_test.go

package advanced

import (
    "context"
    "testing"

    "ai-agent-assistant/internal/rag/adaptive"
    "ai-agent-assistant/internal/rag/graph"
)

func TestOrchestrator_AutoModeSelection(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultOrchestratorConfig()
    config.EnableAutoModeSelection = true

    orchestrator, err := NewAdvancedRAGOrchestrator(config)
    if err != nil {
        t.Fatalf("Failed to create orchestrator: %v", err)
    }

    // 注入 RAG 实现
    graphRAG, _ := graph.NewEnhancedGraphRAG(llm, graph.DefaultGraphRAGConfig())
    selfRAG, _ := adaptive.NewEnhancedSelfRAG(llm, adaptive.DefaultEnhancedSelfRAGConfig())
    correctiveRAG, _ := adaptive.NewCorrectiveRAG(llm, adaptive.DefaultCorrectiveRAGConfig())
    agenticRAG, _ := adaptive.NewAgenticRAG(llm, adaptive.DefaultAgenticRAGConfig())

    orchestrator.SetEnhancedGraphRAG(graphRAG)
    orchestrator.SetEnhancedSelfRAG(selfRAG)
    orchestrator.SetCorrectiveRAG(correctiveRAG)
    orchestrator.SetAgenticRAG(agenticRAG)

    // 测试全局查询 → Graph RAG
    ctx := context.Background()
    result, err := orchestrator.Query(ctx, "系统整体架构概览", "auto")
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if result.Answer == "" {
        t.Error("Expected answer")
    }

    if result.QueryAnalysis == nil {
        t.Error("Expected query analysis")
    }

    if result.ModeUsed == "" {
        t.Error("Expected mode used")
    }
}

func TestOrchestrator_SpecificMode(t *testing.T) {
    llm := &MockLLM{}
    config := DefaultOrchestratorConfig()

    orchestrator, _ := NewAdvancedRAGOrchestrator(config)
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
}

func TestOrchestrator_QueryAnalysis(t *testing.T) {
    config := DefaultOrchestratorConfig()
    orchestrator, _ := NewAdvancedRAGOrchestrator(config)

    ctx := context.Background()

    // 测试全局查询
    analysis := orchestrator.analyzeQuery(ctx, "系统整体架构")
    if analysis.QueryType != "global" {
        t.Errorf("Expected query type 'global', got '%s'", analysis.QueryType)
    }

    // 测试推理查询
    analysis = orchestrator.analyzeQuery(ctx, "为什么会出现这个问题")
    if analysis.QueryType != "reasoning" {
        t.Errorf("Expected query type 'reasoning', got '%s'", analysis.QueryType)
    }

    // 测试复杂度
    analysis = orchestrator.analyzeQuery(ctx, "这是一个非常长的查询，包含了大量的详细信息")
    if analysis.Complexity != "complex" {
        t.Errorf("Expected complexity 'complex', got '%s'", analysis.Complexity)
    }
}

func TestOrchestrator_ModeSelector(t *testing.T) {
    selector := &DefaultModeSelector{}
    ctx := context.Background()

    // 测试全局查询
    analysis := &QueryAnalysis{
        QueryType:     "global",
        Complexity:    "medium",
        RequiresGraph: true,
    }

    mode, err := selector.SelectMode(ctx, "全局查询", analysis)
    if err != nil {
        t.Fatalf("Mode selection failed: %v", err)
    }

    if mode != "enhanced_graph" {
        t.Errorf("Expected mode 'enhanced_graph', got '%s'", mode)
    }

    // 测试推理查询
    analysis = &QueryAnalysis{
        QueryType:         "reasoning",
        Complexity:        "complex",
        RequiresReasoning: true,
    }

    mode, _ = selector.SelectMode(ctx, "推理查询", analysis)
    if mode != "agentic" {
        t.Errorf("Expected mode 'agentic', got '%s'", mode)
    }
}
```

### 运行测试

```bash
# 运行所有测试
go test ./internal/rag/graph/... ./internal/rag/adaptive/... ./internal/rag/advanced/... -v

# 运行特定包的测试
go test ./internal/rag/graph/... -v
go test ./internal/rag/adaptive/... -v
go test ./internal/rag/advanced/... -v

# 运行特定测试
go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_GlobalSearch -v

# 查看测试覆盖率
go test ./internal/rag/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## API 接口

### REST API 端点

#### 1. Enhanced Graph RAG 查询

**请求**
```bash
curl -X POST http://localhost:8080/api/v1/rag/graph/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "系统整体架构概览",
    "mode": "global",
    "top_k": 10
  }'
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "系统整体架构概览",
    "answer": "系统采用微服务架构...",
    "contexts": [
      "社区摘要1...",
      "社区摘要2..."
    ],
    "mode": "enhanced_graph_rag",
    "context_type": "graph",
    "graph_hierarchy": {...},
    "latency": "150ms"
  }
}
```

#### 2. Enhanced Self-RAG 查询

**请求**
```bash
curl -X POST http://localhost:8080/api/v1/rag/self/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何优化数据库性能",
    "top_k": 5
  }'
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "如何优化数据库性能",
    "answer": "数据库性能优化可以从以下几个方面入手...",
    "contexts": [...],
    "mode": "enhanced_self_rag",
    "context_type": "self_reflective",
    "quality_metrics": {
      "relevance_score": 0.85,
      "coverage_score": 0.90,
      "accuracy_score": 0.88,
      "completeness_score": 0.82,
      "overall_score": 0.86,
      "issues": [],
      "suggestions": []
    },
    "latency": "200ms"
  }
}
```

#### 3. Corrective RAG 查询

**请求**
```bash
curl -X POST http://localhost:8080/api/v1/rag/corrective/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是深度学习",
    "top_k": 10
  }'
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "什么是深度学习",
    "initial_answer": "深度学习是机器学习的一个分支...",
    "corrected_answer": "深度学习是机器学习的一个分支，它使用多层神经网络...",
    "contexts": [...],
    "mode": "corrective_rag",
    "context_type": "corrective",
    "correction_history": [
      {
        "round": 1,
        "corrections": [
          {
            "type": "fact",
            "confidence": 0.3,
            "description": "描述不够详细",
            "suggestion": "补充神经网络的细节"
          }
        ],
        "before": "初始答案",
        "after": "纠正后答案"
      }
    ],
    "total_corrections": 1,
    "final_validation": {
      "overall_confidence": 0.92,
      "fact_check_confidence": 0.95,
      "logic_check_confidence": 0.90,
      "consistency_confidence": 0.91,
      "passed": true
    },
    "latency": "350ms"
  }
}
```

#### 4. Agentic RAG 查询

**请求**
```bash
curl -X POST http://localhost:8080/api/v1/rag/agentic/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析用户行为数据并给出优化建议",
    "mode": "react",
    "max_iterations": 10
  }'
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "分析用户行为数据并给出优化建议",
    "answer": "基于用户行为数据分析，建议从以下几个方面优化...",
    "contexts": [...],
    "mode": "agentic_rag",
    "context_type": "agentic",
    "thoughts": [
      {
        "content": "需要分析用户行为数据",
        "reasoning": "这是一个数据分析任务"
      },
      {
        "content": "使用向量搜索工具",
        "reasoning": "需要找到相关的用户数据"
      }
    ],
    "actions": [
      {
        "tool": "vector_search",
        "input": "用户行为数据"
      }
    ],
    "observations": [
      {
        "content": "找到相关用户数据",
        "type": "search_result"
      }
    ],
    "iterations": 5,
    "confidence": 0.88,
    "latency": "500ms"
  }
}
```

#### 5. Orchestrator 统一查询

**请求**
```bash
# 自动模式选择
curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "系统架构概览",
    "mode": "auto"
  }'

# 指定模式
curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "优化建议",
    "mode": "agentic"
  }'

# 融合模式
curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "复杂查询",
    "mode": "fused"
  }'
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "系统架构概览",
    "answer": "系统采用微服务架构...",
    "mode": "enhanced_graph_rag",
    "mode_used": "enhanced_graph",
    "query_analysis": {
      "query_type": "global",
      "complexity": "medium",
      "domain": "technical",
      "requires_graph": true,
      "requires_reasoning": false,
      "keywords": ["系统", "架构", "概览"]
    },
    "latency": "180ms",
    "graph_hierarchy": {...}
  }
}
```

### 请求参数说明

#### Graph RAG 参数
| 参数 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| query | string | 是 | 查询内容 | - |
| mode | string | 否 | 检索模式 | global, local, path |
| top_k | int | 否 | 返回数量 | 默认 10 |

#### Self-RAG 参数
| 参数 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| query | string | 是 | 查询内容 | - |
| top_k | int | 否 | 初始检索数量 | 默认 5 |
| enable_dynamic_threshold | bool | 否 | 启用动态阈值 | 默认 true |

#### Corrective RAG 参数
| 参数 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| query | string | 是 | 查询内容 | - |
| top_k | int | 否 | 检索数量 | 默认 10 |
| max_correction_rounds | int | 否 | 最大纠错轮数 | 默认 3 |

#### Agentic RAG 参数
| 参数 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| query | string | 是 | 查询内容 | - |
| mode | string | 否 | 代理模式 | react, plan_execute, reflexion |
| max_iterations | int | 否 | 最大迭代次数 | 默认 10 |
| enable_reflection | bool | 否 | 启用反思 | 默认 true |

#### Orchestrator 参数
| 参数 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| query | string | 是 | 查询内容 | - |
| mode | string | 否 | RAG 模式 | auto, enhanced_graph, enhanced_self, corrective, agentic, fused |

---

## 测试清单

### 编译验证 ✅

```bash
# 编译所有高级 RAG 包
go build ./internal/rag/graph/... ./internal/rag/adaptive/... ./internal/rag/advanced/...

# 预期结果：编译成功，无错误
# 实际结果：✅ 通过
```

### 单元测试

#### Graph RAG 测试

- [ ] TestEnhancedGraphRAG_GlobalSearch
  ```bash
  go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_GlobalSearch -v
  ```

- [ ] TestEnhancedGraphRAG_LocalSearch
  ```bash
  go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_LocalSearch -v
  ```

- [ ] TestEnhancedGraphRAG_PathBasedSearch
  ```bash
  go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_PathBasedSearch -v
  ```

- [ ] TestEnhancedGraphRAG_EntityScoring
  ```bash
  go test ./internal/rag/graph/... -run TestEnhancedGraphRAG_EntityScoring -v
  ```

#### Self-RAG 测试

- [ ] TestEnhancedSelfRAG_EnhancedRetrieve
  ```bash
  go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG_EnhancedRetrieve -v
  ```

- [ ] TestEnhancedSelfRAG_QualityEvaluation
  ```bash
  go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG_QualityEvaluation -v
  ```

- [ ] TestEnhancedSelfRAG_DynamicThreshold
  ```bash
  go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG_DynamicThreshold -v
  ```

- [ ] TestEnhancedSelfRAG_PerformanceTracking
  ```bash
  go test ./internal/rag/adaptive/... -run TestEnhancedSelfRAG_PerformanceTracking -v
  ```

#### Corrective RAG 测试

- [ ] TestCorrectiveRAG_RetrieveAndCorrect
  ```bash
  go test ./internal/rag/adaptive/... -run TestCorrectiveRAG_RetrieveAndCorrect -v
  ```

- [ ] TestCorrectiveRAG_FactCheck
  ```bash
  go test ./internal/rag/adaptive/... -run TestCorrectiveRAG_FactCheck -v
  ```

- [ ] TestCorrectiveRAG_LogicValidation
  ```bash
  go test ./internal/rag/adaptive/... -run TestCorrectiveRAG_LogicValidation -v
  ```

- [ ] TestCorrectiveRAG_ConsistencyCheck
  ```bash
  go test ./internal/rag/adaptive/... -run TestCorrectiveRAG_ConsistencyCheck -v
  ```

- [ ] TestCorrectiveRAG_CorrectionLoop
  ```bash
  go test ./internal/rag/adaptive/... -run TestCorrectiveRAG_CorrectionLoop -v
  ```

#### Agentic RAG 测试

- [ ] TestAgenticRAG_Query_ReactMode
  ```bash
  go test ./internal/rag/adaptive/... -run TestAgenticRAG_Query_ReactMode -v
  ```

- [ ] TestAgenticRAG_Query_PlanAndExecuteMode
  ```bash
  go test ./internal/rag/adaptive/... -run TestAgenticRAG_Query_PlanAndExecuteMode -v
  ```

- [ ] TestAgenticRAG_Query_ReflexionMode
  ```bash
  go test ./internal/rag/adaptive/... -run TestAgenticRAG_Query_ReflexionMode -v
  ```

- [ ] TestAgenticRAG_ToolExecution
  ```bash
  go test ./internal/rag/adaptive/... -run TestAgenticRAG_ToolExecution -v
  ```

#### Orchestrator 测试

- [ ] TestOrchestrator_AutoModeSelection
  ```bash
  go test ./internal/rag/advanced/... -run TestOrchestrator_AutoModeSelection -v
  ```

- [ ] TestOrchestrator_SpecificMode
  ```bash
  go test ./internal/rag/advanced/... -run TestOrchestrator_SpecificMode -v
  ```

- [ ] TestOrchestrator_QueryAnalysis
  ```bash
  go test ./internal/rag/advanced/... -run TestOrchestrator_QueryAnalysis -v
  ```

- [ ] TestOrchestrator_ResultFusion
  ```bash
  go test ./internal/rag/advanced/... -run TestOrchestrator_ResultFusion -v
  ```

- [ ] TestOrchestrator_ModeSelector
  ```bash
  go test ./internal/rag/advanced/... -run TestOrchestrator_ModeSelector -v
  ```

### 集成测试

#### 端到端流程测试

- [ ] 测试完整的查询流程
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "测试查询", "mode": "auto"}'
  ```

- [ ] 测试模式切换
  ```bash
  # 测试从 auto 切换到特定模式
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "测试查询", "mode": "auto"}'

  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "测试查询", "mode": "enhanced_graph"}'
  ```

- [ ] 测试融合模式
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "复杂查询", "mode": "fused"}'
  ```

### 性能测试

- [ ] Graph RAG 全局检索性能
  ```bash
  ab -n 100 -c 10 -p graph_payload.json -T application/json \
    http://localhost:8080/api/v1/rag/graph/query
  ```

- [ ] Self-RAG 增强检索性能
  ```bash
  ab -n 100 -c 10 -p self_payload.json -T application/json \
    http://localhost:8080/api/v1/rag/self/query
  ```

- [ ] Corrective RAG 纠错检索性能
  ```bash
  ab -n 100 -c 10 -p corrective_payload.json -T application/json \
    http://localhost:8080/api/v1/rag/corrective/query
  ```

- [ ] Agentic RAG 代理检索性能
  ```bash
  ab -n 100 -c 10 -p agentic_payload.json -T application/json \
    http://localhost:8080/api/v1/rag/agentic/query
  ```

### 功能验证

#### 查询类型识别

- [ ] 全局查询识别
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "系统整体架构", "mode": "auto"}'
  # 预期：选择 Enhanced Graph RAG
  ```

- [ ] 推理查询识别
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "为什么性能会下降", "mode": "auto"}'
  # 预期：选择 Agentic RAG 或 Enhanced Graph RAG
  ```

- [ ] 技术查询识别
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "复杂的算法优化问题", "mode": "auto"}'
  # 预期：选择 Corrective RAG
  ```

#### 质量指标验证

- [ ] Self-RAG 质量评估
  ```bash
  # 检查返回的 quality_metrics 字段
  # 验证所有分数在 0-1 范围内
  ```

- [ ] Corrective RAG 验证
  ```bash
  # 检查返回的 final_validation 字段
  # 验证所有置信度在 0-1 范围内
  ```

- [ ] Agentic RAG 置信度
  ```bash
  # 检查返回的 confidence 字段
  # 验证置信度在 0-1 范围内
  ```

### 错误处理

- [ ] 无效查询处理
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "", "mode": "auto"}'
  # 预期：返回错误信息
  ```

- [ ] 不支持的模式处理
  ```bash
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "测试", "mode": "invalid_mode"}'
  # 预期：返回错误或使用默认模式
  ```

- [ ] 超时处理
  ```bash
  # 配置较短的 ModeTimeout
  # 发送复杂查询，验证超时处理
  ```

### 边界条件

- [ ] 空结果处理
  ```bash
  # 查询不存在的主题
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "完全不相关的内容xyz123", "mode": "auto"}'
  # 预期：返回"未找到相关信息"消息
  ```

- [ ] 极长查询处理
  ```bash
  # 发送超长查询（>1000 字）
  # 验证正确处理
  ```

- [ ] 特殊字符处理
  ```bash
  # 发送包含特殊字符的查询
  curl -X POST http://localhost:8080/api/v1/rag/orchestrator/query \
    -H "Content-Type: application/json" \
    -d '{"query": "测试 <script>alert(\"xss\")</script>", "mode": "auto"}'
  # 预期：正确转义和处理
  ```

---

## 总结

### 已完成功能

✅ **Enhanced Graph RAG**
- 动态社区摘要生成
- 层次化检索
- 实体重要性评分
- 多跳关系检索
- 路径查找

✅ **Enhanced Self-RAG**
- 动态阈值调整
- 多维度质量评估
- 自适应策略选择
- 性能追踪和优化

✅ **Corrective RAG**
- 事实核查
- 逻辑验证
- 一致性检查
- 迭代纠错

✅ **Agentic RAG**
- ReAct 模式
- Plan-and-Execute 模式
- 反思机制
- 工具使用

✅ **RAG Orchestrator**
- 自动模式选择
- 模式切换
- 结果融合
- 查询分析

### 文件清单

```
internal/rag/
├── graph/
│   ├── graph_rag.go
│   ├── enhanced_graph_rag.go           # ✅ 新增
│   └── enhanced_graph_rag_test.go      # 📝 待创建
├── adaptive/
│   ├── self_reflective_rag.go
│   ├── enhanced_self_rag.go            # ✅ 新增
│   ├── enhanced_self_rag_test.go       # 📝 待创建
│   ├── corrective_rag.go               # ✅ 新增
│   ├── corrective_rag_test.go          # 📝 待创建
│   ├── agentic_rag.go                  # ✅ 新增
│   └── agentic_rag_test.go             # 📝 待创建
└── advanced/
    ├── orchestrator.go                 # ✅ 新增
    └── orchestrator_test.go            # 📝 待创建
```

### 下一步工作

1. **单元测试实现**: 根据文档中的测试示例创建完整的单元测试
2. **集成测试**: 创建端到端的集成测试
3. **性能优化**: 根据性能测试结果进行优化
4. **文档完善**: 补充更多使用示例和最佳实践
5. **生产部署**: 准备生产环境配置和监控

---

**文档结束**

如有问题，请联系开发团队或查看项目文档。
