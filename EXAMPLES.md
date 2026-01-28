# AI Agent Assistant v0.4 使用示例

本目录包含v0.4新功能的使用示例和最佳实践。

## 📁 示例列表

1. [基础对话](#1-基础对话)
2. [多模型切换](#2-多模型切换)
3. [RAG增强功能](#3-rag增强功能)
4. [推理能力](#4-推理能力)
5. [智能记忆管理](#5-智能记忆管理)
6. [会话状态管理](#6-会话状态管理)
7. [评估系统](#7-评估系统)

---

## 1. 基础对话

### 示例1.1：简单对话

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "demo-001",
    "message": "你好，请介绍一下你自己"
  }'
```

### 示例1.2：多轮对话

```bash
# 第一轮
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "demo-002",
    "message": "我叫张三"
  }'

# 第二轮（会记住之前的对话）
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "demo-002",
    "message": "我叫什么名字？"
  }'
```

---

## 2. 多模型切换

### 示例2.1：使用不同模型

```bash
# 使用GLM模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "model-test",
    "message": "用三句话解释什么是微服务",
    "model": "glm"
  }'

# 切换到千问模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "model-test",
    "message": "用三句话解释什么是微服务",
    "model": "qwen"
  }'

# 使用OpenAI GPT-4（需要配置API Key）
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "model-test",
    "message": "用三句话解释什么是微服务",
    "model": "gpt-4"
  }'
```

### 示例2.2：查看模型信息

```bash
# 列出所有支持模型
curl http://localhost:8080/api/v1/models

# 查看特定模型信息
curl http://localhost:8080/api/v1/models/qwen
```

---

## 3. RAG增强功能

### 示例3.1：添加知识

```bash
# 添加文本知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "AI Agent Assistant v0.4是一个功能强大的智能体框架",
    "source": "产品介绍"
  }'

# 从PDF文档添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -H 'Content-Type: application/json' \
  -d '{
    "doc_path": "/path/to/document.pdf"
  }'
```

### 示例3.2：RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "rag-demo",
    "message": "v0.4有哪些新特性？"
  }'
```

### 示例3.3：搜索知识库

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "语义分块",
    "top_k": 5
  }'
```

### 示例3.4：查看知识库统计

```bash
curl http://localhost:8080/api/v1/knowledge/stats
```

---

## 4. 推理能力

### 示例4.1：思维链推理

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "请逐步推理：一个农场有鸡和兔共50只，有140条腿，鸡和兔各多少只？"
  }'
```

**预期响应**:
```json
{
  "reasoning": "【思考过程】\n设鸡有x只，兔有y只\nx+y=50\n2x+4y=140\n解方程得：x=30, y=20",
  "answer": "农场有30只鸡和20只兔子。"
}
```

### 示例4.2：自我反思

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/reflect \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "解释什么是RESTful API",
    "previous_attempts": [
      "RESTful API是一种API风格",
      "REST使用HTTP方法（GET、POST等）"
    ]
  }'
```

---

## 5. 智能记忆管理

### 示例5.1：自动提取记忆

```bash
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "user-alice",
    "conversation": "用户：我叫Alice，来自北京。\\n助手：你好Alice！北京是个美丽的城市。\\n用户：是的，我是个软件工程师，主要用Python工作。"
  }'
```

**响应**:
```json
{
  "message": "Memories extracted",
  "count": 2,
  "memories": [
    {
      "id": "mem_001",
      "user_id": "user-alice",
      "content": "用户名叫Alice，来自北京，是个软件工程师，主要使用Python工作",
      "topics": ["个人信息", "职业", "位置"],
      "importance": 0.85
    }
  ]
}
```

### 示例5.2：语义搜索记忆

```bash
curl "http://localhost:8080/api/v1/memory/search?user_id=user-alice&query=职业&limit=3"
```

### 示例5.3：记忆优化

```go
// 代码示例：配置记忆优化策略
memoryManager.SetOptimizationStrategy("importance") // 重要性优化
memoryManager.SetOptimizationStrategy("time_decay")  // 时间衰减优化
memoryManager.SetOptimizationStrategy("summarization") // 摘要优化
```

---

## 6. 会话状态管理

### 示例6.1：更新会话状态

```bash
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-session",
    "updates": {
      "username": "johndoe",
      "theme": "dark",
      "language": "zh",
      "preferences": {
        "notifications": true,
        "auto_save": true
      }
    }
  }'
```

### 示例6.2：获取会话信息

```bash
curl "http://localhost:8080/api/v1/session?session_id=user-session"
```

**响应**:
```json
{
  "session_id": "user-session",
  "model": "qwen",
  "summary": "讨论了AI技术问题...",
  "state": {
    "data": {
      "username": "johndoe",
      "theme": "dark",
      "language": "zh"
    },
    "version": 2
  },
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:05:00Z"
}
```

### 示例6.3：清除会话

```bash
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=user-session"
```

---

## 7. 评估系统

### 示例7.1：准确性评估

```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{
    "test_cases": [
      {
        "input": "1+1=?",
        "expected": "2"
      },
      {
        "input": "中国首都是哪里？",
        "expected": "北京"
      },
      {
        "input": "Go语言的作者是谁？",
        "expected": "Robert Griesemer、Rob Pike和Ken Thompson"
      }
    ],
    "accuracy": true,
    "performance": false
  }'
```

**评估报告示例**:
```
==================================================
评估报告
==================================================

评估器: AccuracyEval
总用例数: 3
通过数: 2
失败数: 1
准确率: 66.67%
得分: 0.83
耗时: 5.2s

详细指标:
  avg_score: 0.83
  pass_rate: 0.67
  threshold: 0.80

--------------------------------------------------

评估器: PerformanceEval
总用例数: 15 (5次运行 × 3个用例)
通过数: 15
失败数: 0
得分: 28.50 rps
耗时: 526ms

详细指标:
  avg_latency_ms: 52
  p50_latency_ms: 48
  p95_latency_ms: 89
  p99_latency_ms: 120
  throughput_rps: 28.50
  avg_tokens: 45.2
  tokens_per_second: 1285.3
```

---

## 高级场景示例

### 场景1：构建客服机器人

```bash
# 1. 提取用户信息
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "customer-001",
    "conversation": "用户：我想退货。\\n助手：请告诉我您的订单号。\\n用户：订单号是ORD-2024-001。"
  }'

# 2. 更新订单状态到会话
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "customer-001",
    "updates": {
      "order_id": "ORD-2024-001",
      "issue": "退货请求",
      "status": "处理中"
    }
  }'

# 3. 使用RAG查询退货政策
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "customer-001",
    "message": "退货政策是什么？"
  }'
```

### 场景2：知识问答系统

```bash
# 1. 添加产品文档
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -H 'Content-Type: application/json' \
  -d '{
    "doc_path": "/docs/product-manual.pdf"
  }'

# 2. 语义搜索相关知识
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "如何重置密码？",
    "top_k": 3
  }'

# 3. RAG增强回答
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content: application/json' \
  -d '{
    "session_id": "qa-session",
    "message": "用户忘记密码怎么办？"
  }'
```

### 场景3：代码分析助手

```bash
# 使用思维链分析代码
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "分析以下Go代码的问题：\\n\\nfunc add(a, b int) int {\\n    return a + b\\n}\\n\\nfunc main() {\\n    fmt.Println(add(1, 2))\\n}\\n"
  }'
```

---

## 最佳实践

### 1. 模型选择策略

**简单任务**（如打招呼、闲聊）：
- 使用 `qwen-turbo` 或 `glm-4-flash`（快速便宜）

**复杂任务**（如推理、分析）：
- 使用 `deepseek-r1` 或 `qwen-max`
- 或使用思维链功能

**平衡性能和成本**：
- 使用 `qwen-plus`（性价比高）

### 2. RAG使用建议

**文档类型**：
- 短文档：直接固定分块
- 长文档：语义分块
- 结构化文档：递归分块

**检索方式**：
- 精确匹配：向量检索
- 模糊搜索：混合检索（向量+关键词）
- 高准确度：混合检索+重排序

### 3. 记忆管理建议

**提取频率**：
- 重要对话结束后立即提取
- 批量处理历史对话

**优化策略**：
- 重视程度：`importance`
- 旧数据：`time_decay`
- 数据去重：`summarization`

### 4. 会话管理建议

**自动摘要阈值**：
- 短对话：10条消息
- 长对话：5条消息

**状态管理**：
- 定期保存重要状态
- 使用版本控制避免冲突

---

## 常见问题

### Q1：如何添加新的模型？

创建新的模型实现，然后注册到工厂：

```go
// 1. 实现Model接口
type MyModel struct { ... }

// 2. 在factory.go中注册
func (f *ModelFactory) CreateModel(modelName string, cfg *config.Config) (Model, error) {
    switch modelName {
    case "my-model":
        return NewMyModel(cfg.Models.MyModel)
    ...
    }
}
```

### Q2：如何自定义分块策略？

实现`Chunker`接口：

```go
type MyChunker struct {
    ...
}

func (c *MyChunker) Split(text string) []string {
    // 自定义分块逻辑
}
```

### Q3：如何添加新的评估指标？

实现`Evaluator`接口：

```go
type MyEvaluator struct {
    ...
}

func (e *MyEvaluator) Evaluate(ctx context.Context, model Model, dataset []TestCase) (*EvalResult, error) {
    // 自定义评估逻辑
}
```

---

## 相关文档

- [使用指南](USAGE_GUIDE.md) - 完整使用指南
- [API文档](API_CURL_DOCS.md) - API参考
- [配置说明](config.yaml.example) - 配置文件模板

---

**最后更新**: 2026-01-27
**版本**: v0.4
