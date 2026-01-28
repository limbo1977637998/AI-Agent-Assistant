# AI Agent Assistant v0.4 使用指南

> **版本**: v0.4 | **更新日期**: 2026-01-27

---

## 📚 目录

- [新功能概览](#新功能概览)
- [快速开始](#快速开始)
- [多模型使用](#多模型使用)
- [RAG增强功能](#rag增强功能)
- [推理能力使用](#推理能力使用)
- [评估与监控](#评估与监控)
- [高级配置](#高级配置)

---

## 新功能概览

v0.4版本相比v0.3，新增以下核心功能：

### ✨ 多模型支持
- 支持5+主流模型提供商（GLM、千问、OpenAI、Claude、DeepSeek）
- 统一的Model接口，无缝切换
- 模型工厂模式，易于扩展

### 🧠 RAG增强
- **语义分块** - 基于千问Embedding的智能分块
- **混合检索** - 向量+BM25关键词检索
- **重排序** - CrossEncoder API结果重排

### 🤔 推理能力
- **思维链推理** - 展示思考过程
- **自我反思** - 检查错误并改进答案
- **多步推理** - 复杂任务分步执行

### 📊 评估与监控
- **准确性评估** - 精确匹配/相似度/LLM判断
- **性能评估** - 延迟/吞吐量/Token使用
- **OpenTelemetry追踪** - 分布式链路追踪
- **Prometheus监控** - 实时指标收集

### 💾 智能记忆管理
- **自动提取** - LLM自动从对话提取重要信息
- **语义检索** - 基于向量相似度的智能检索
- **记忆优化** - 时间衰减、重要性评分、去重合并

### 🔄 会话管理增强
- **自动摘要** - 超过阈值自动生成会话摘要
- **并发控制** - 读写锁保证线程安全
- **状态管理** - 版本控制的结构化存储

---

## 快速开始

### 1. 配置模型

编辑 `config.yaml.example`：

```yaml
models:
  glm:
    api_key: "your-glm-api-key"
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    model: "glm-4-flash"

  qwen:
    api_key: "your-qwen-api-key"
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    model: "qwen-plus"

agent:
  default_model: qwen              # 默认对话模型
  embedding_model: qwen           # Embedding模型（推荐千问）
```

### 2. 启动服务

```bash
# 编译
go build -o bin/server cmd/server/main.go

# 运行
./bin/server
```

服务将在 `http://localhost:8080` 启动。

### 3. 健康检查

```bash
curl http://localhost:8080/health
```

**响应**:
```json
{
  "status": "healthy",
  "version": "v0.4",
  "features": [
    "Multi-Model Support",
    "Enhanced RAG",
    "Reasoning Capability",
    "Auto Memory Extraction",
    "Auto Session Summary",
    "Evaluation & Monitoring"
  ]
}
```

---

## 多模型使用

### 查看可用模型

```bash
curl http://localhost:8080/api/v1/models
```

**响应**:
```json
{
  "supported_models": [
    "glm-4-flash",
    "qwen-plus",
    "qwen-turbo",
    "gpt-3.5-turbo",
    "gpt-4",
    "claude-3-5-sonnet",
    "deepseek-chat",
    "deepseek-r1"
  ],
  "loaded_models": ["glm", "qwen"]
}
```

### 切换模型对话

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "message": "你好",
    "model": "gpt-4"
  }'
```

---

## RAG增强功能

### 添加文档（语义分块）

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -H "Content-Type: application/json" \
  -d '{
    "doc_path": "/path/to/document.pdf"
  }'
```

### 混合检索+重排序

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是RAG？",
    "top_k": 5
  }'
```

**响应**:
```json
{
  "query": "什么是RAG？",
  "count": 5,
  "results": [
    "RAG是检索增强生成...",
    "RAG系统包含三个核心组件...",
    ...
  ]
}
```

---

## 推理能力使用

### 思维链推理

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H "Content-Type: application/json" \
  -d '{
    "task": "25 * 4 = ?"
  }'
```

**响应**:
```json
{
  "reasoning": "【思考过程】\n首先分析问题核心是数学计算。\n然后计算 25 * 4 = 100。",
  "answer": "25乘以4等于100。"
}
```

### 自我反思

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/reflect \
  -H "Content-Type: application/json" \
  -d '{
    "task": "解释什么是机器学习",
    "previous_attempts": [
      "机器学习是AI的一个分支",
      "ML让计算机从数据中学习"
    ]
  }'
```

---

## 评估与监控

### 准确性评估

```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H "Content-Type: application/json" \
  -d '{
    "test_cases": [
      {
        "input": "1+1=?",
        "expected": "2"
      },
      {
        "input": "2+2=?",
        "expected": "4"
      }
    ],
    "accuracy": true,
    "performance": true
  }'
```

**响应**:
```json
{
  "results": [...],
  "report": "评估报告...",
  "overall_score": 0.85
}
```

### Prometheus监控

访问指标端点：

```bash
curl http://localhost:9090/metrics
```

**可用指标**:
- `agent_requests_total` - 请求总数
- `agent_response_time_seconds` - 响应时间
- `agent_tokens_total` - Token使用量
- `agent_tool_calls_total` - 工具调用次数
- `agent_cache_hits_total` - 缓存命中数
- `agent_rag_retrievals_total` - RAG检索次数

---

## 智能记忆管理

### 自动提取记忆

```bash
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "conversation": "用户：我叫张三，喜欢Go语言。\\n助手：你好张三，很高兴认识你。"
  }'
```

**响应**:
```json
{
  "message": "Memories extracted",
  "count": 1,
  "memories": [
    {
      "id": "mem_123",
      "user_id": "user-123",
      "content": "用户名叫张三，喜欢Go语言编程",
      "topics": ["个人信息", "偏好"],
      "importance": 0.8
    }
  ]
}
```

### 语义检索记忆

```bash
curl "http://localhost:8080/api/v1/memory/search?user_id=user-123&query=编程&limit=5"
```

---

## 会话状态管理

### 更新会话状态

```bash
curl -X POST http://localhost:8080/api/v1/session/state \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "updates": {
      "username": "zhangsan",
      "theme": "dark",
      "language": "zh"
    }
  }'
```

**响应**:
```json
{
  "message": "State updated",
  "version": 1
}
```

### 获取会话信息

```bash
curl "http://localhost:8080/api/v1/session?session_id=user-123"
```

**响应**:
```json
{
  "session_id": "user-123",
  "model": "qwen",
  "summary": "用户讨论了技术问题...",
  "state": {
    "data": {...},
    "version": 1
  }
}
```

---

## 高级配置

### 启用OpenTelemetry追踪

编辑 `config.yaml`:

```yaml
monitoring:
  enabled: true
  tracing:
    enabled: true
    jaeger_endpoint: "http://localhost:4318"
```

### 启用Prometheus监控

```yaml
monitoring:
  enabled: true
  prometheus:
    port: 9090
    path: "/metrics"
```

### 配置RAG增强

```yaml
rag:
  enabled: true
  top_k: 3
  threshold: 0.3
  chunk_size: 500
  chunk_overlap: 50
  enable_hybrid_search: true  # 启用混合检索
```

### 配置记忆管理

```yaml
memory:
  max_history: 10
  store_type: "mysql"
  enable_user_memory: true
  enable_state_memory: true
  memory_optimization: "importance"
```

---

## 使用示例

### 示例1：多模型对话

```bash
# 使用GLM对话
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test", "message": "你好", "model": "glm"}'

# 切换到千问
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test", "message": "你好", "model": "qwen"}'
```

### 示例2：RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "message": "v0.4有哪些新特性？"
  }'
```

### 示例3：思维链推理

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H "Content-Type: application/json" \
  -d '{
    "task": "分析一下Go语言的优势和应用场景"
  }'
```

### 示例4：提取和管理记忆

```bash
# 1. 提取记忆
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-456",
    "conversation": "用户：我最近在学习Rust。\\n助手：Rust是一门系统编程语言..."
  }'

# 2. 搜索记忆
curl "http://localhost:8080/api/v1/memory/search?user_id=user-456&query=Rust&limit=3"
```

---

## API参考

### 完整API列表

| 端点 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/v1/chat` | POST | 基础对话 |
| `/api/v1/chat/rag` | POST | RAG增强对话 |
| `/api/v1/reasoning/cot` | POST | 思维链推理 |
| `/api/v1/reasoning/reflect` | POST | 自我反思 |
| `/api/v1/session` | GET | 获取会话 |
| `/api/v1/session` | DELETE | 清除会话 |
| `/api/v1/session/state` | POST | 更新会话状态 |
| `/api/v1/memory/extract` | POST | 提取记忆 |
| `/api/v1/memory/search` | GET | 搜索记忆 |
| `/api/v1/knowledge/add` | POST | 添加知识 |
| `/api/v1/knowledge/add/doc` | POST | 从文档添加知识 |
| `/api/v1/knowledge/search` | POST | 搜索知识 |
| `/api/v1/knowledge/stats` | GET | 知识库统计 |
| `/api/v1/eval/accuracy` | POST | 准确性评估 |
| `/api/v1/models` | GET | 列出模型 |
| `/api/v1/models/:name` | GET | 获取模型信息 |

---

## 故障排查

### 问题1：模型未加载

**症状**：调用API返回"Model not available"

**解决**：
1. 检查 `config.yaml.example` 中的API Key是否正确
2. 查看日志输出确认模型初始化状态
3. 使用 `GET /api/v1/models` 查看已加载模型

### 问题2：RAG检索无结果

**症状**：RAG对话返回空上下文

**解决**：
1. 确认已添加知识文档
2. 检查Embedding模型是否可用
3. 尝试降低 `threshold` 值
4. 增加 `top_k` 值

### 问题3：推理功能不可用

**症状**：推理API返回错误

**解决**：
1. 确认配置了DeepSeek-R1或回退到千问
2. 检查模型是否支持推理功能
3. 查看日志中的错误详情

---

## 性能优化建议

### 1. 缓存策略

- 启用Redis缓存LLM响应和工具结果
- 调整缓存TTL以平衡性能和实时性

### 2. 并发控制

- 使用连接池管理数据库连接
- 限制并发API请求数

### 3. 资源管理

- 设置合理的Token限制
- 定期清理过期会话
- 优化向量数据库索引

---

## 下一步

- 查看 [API_CURL_DOCS.md](API_CURL_DOCS.md) 获取完整API文档
- 运行测试验证功能：`./run_tests.sh`
- 查看示例代码：`examples/`

---

**文档生成时间**: 2026-01-27
**版本**: v0.4
