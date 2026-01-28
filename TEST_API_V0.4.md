# AI Agent Assistant v0.4 API 测试命令集合

> 版本: v0.4 | 更新日期: 2026-01-27

本文档包含所有v0.4新开发的API接口的curl测试命令。

## 📋 前置准备

### 1. 启动服务

```bash
# 编译
go build -o bin/server cmd/server/main_enhanced.go

# 运行
./bin/server
```

服务将在 `http://localhost:8080` 启动。

### 2. 配置文件

确保 `config.yaml` 已配置好以下内容：
- 模型API密钥（GLM、千问、OpenAI等）
- MySQL数据库连接
- Redis连接
- Embedding模型配置

---

## 🏥 健康检查

### 1. 基础健康检查

```bash
curl http://localhost:8080/health
```

**预期响应**:
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

## 🤖 模型管理接口

### 2. 列出所有支持的模型

```bash
curl http://localhost:8080/api/v1/models
```

**预期响应**:
```json
{
  "supported_models": [
    "glm-4-flash",
    "glm-4-plus",
    "glm-4-alltools",
    "qwen-turbo",
    "qwen-plus",
    "qwen-max",
    "gpt-3.5-turbo",
    "gpt-4",
    "gpt-4-turbo",
    "gpt-4o",
    "claude-3-5-sonnet",
    "claude-3-opus",
    "claude-3-haiku",
    "deepseek-chat",
    "deepseek-coder",
    "deepseek-r1"
  ],
  "loaded_models": ["glm", "qwen"]
}
```

### 3. 获取特定模型信息

```bash
curl http://localhost:8080/api/v1/models/qwen
```

---

## 💬 对话接口

### 4. 基础对话（使用默认模型）

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-001",
    "message": "你好，请介绍一下你自己"
  }'
```

### 5. 指定模型对话（GLM）

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-002",
    "message": "用三句话解释什么是Go语言",
    "model": "glm"
  }'
```

### 6. 指定模型对话（千问）

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-003",
    "message": "用三句话解释什么是Go语言",
    "model": "qwen"
  }'
```

### 7. 多轮对话（会话记忆）

```bash
# 第一轮
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-004",
    "message": "我叫张三"
  }'

# 第二轮（会记住之前的对话）
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-004",
    "message": "我叫什么名字？"
  }'
```

---

## 🧠 RAG增强对话

### 8. RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "rag-test-001",
    "message": "v0.4有哪些新特性？",
    "top_k": 3
  }'
```

**预期响应**:
```json
{
  "response": "根据知识库，v0.4包含以下新特性...",
  "rag_used": true,
  "session_id": "rag-test-001"
}
```

### 9. 添加文本知识

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "AI Agent Assistant v0.4是一个功能强大的智能体框架，支持多模型、RAG增强、推理能力等特性",
    "source": "测试文档"
  }'
```

### 10. 添加文档知识（PDF）

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -H 'Content-Type: application/json' \
  -d '{
    "doc_path": "/path/to/document.pdf"
  }'
```

### 11. 搜索知识库

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "语义分块",
    "top_k": 5
  }'
```

**预期响应**:
```json
{
  "query": "语义分块",
  "count": 5,
  "results": [
    "语义分块是基于embedding相似度的智能文本分割方法...",
    ...
  ]
}
```

### 12. 获取知识库统计

```bash
curl http://localhost:8080/api/v1/knowledge/stats
```

**预期响应**:
```json
{
  "stats": {
    "total_chunks": 150,
    "total_documents": 10,
    "total_embeddings": 150
  }
}
```

---

## 🤔 推理能力接口

### 13. 思维链推理（Chain-of-Thought）

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

### 14. 数学计算推理

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "25 * 4 = ? 请详细说明计算过程"
  }'
```

### 15. 代码分析推理

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "分析以下Go代码的特点：\n\nfunc add(a, b int) int {\n    return a + b\n}"
  }'
```

### 16. 自我反思（Self-Reflection）

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

**预期响应**:
```json
{
  "reflection": "根据之前的尝试，需要补充以下内容：RESTful API是基于REST架构风格的API...",
  "improved_answer": "RESTful API是遵循REST架构风格的网络API，它使用HTTP协议的标准方法（GET、POST、PUT、DELETE）来操作资源..."
}
```

---

## 💾 会话管理接口

### 17. 获取会话信息

```bash
curl "http://localhost:8080/api/v1/session?session_id=test-001"
```

**预期响应**:
```json
{
  "session_id": "test-001",
  "model": "glm",
  "summary": "用户询问了自我介绍...",
  "state": {
    "data": {
      "username": "zhangsan",
      "theme": "dark"
    },
    "version": 2
  },
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:05:00Z"
}
```

### 18. 更新会话状态

```bash
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "test-001",
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

**预期响应**:
```json
{
  "message": "State updated",
  "version": 1
}
```

### 19. 清除会话

```bash
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=test-001"
```

**预期响应**:
```json
{
  "message": "Session cleared"
}
```

---

## 🧠 记忆管理接口

### 20. 自动提取记忆

```bash
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "user-alice",
    "conversation": "用户：我叫Alice，来自北京，是个软件工程师，主要用Python工作。\n助手：你好Alice！北京是个美丽的城市。\n用户：是的，我还在学习Go语言。"
  }'
```

**预期响应**:
```json
{
  "message": "Memories extracted",
  "count": 2,
  "memories": [
    {
      "id": "mem_001",
      "user_id": "user-alice",
      "content": "用户名叫Alice，来自北京，是个软件工程师，主要使用Python和Go语言",
      "topics": ["个人信息", "职业", "位置", "技能"],
      "importance": 0.85
    }
  ]
}
```

### 21. 语义搜索记忆

```bash
curl "http://localhost:8080/api/v1/memory/search?user_id=user-alice&query=职业&limit=3"
```

**预期响应**:
```json
{
  "query": "职业",
  "count": 2,
  "memories": [
    {
      "id": "mem_001",
      "content": "Alice是个软件工程师",
      "similarity": 0.92
    }
  ]
}
```

### 22. 搜索技术偏好

```bash
curl "http://localhost:8080/api/v1/memory/search?user_id=user-alice&query=编程语言&limit=5"
```

---

## 📊 评估系统接口

### 23. 准确性评估（精确匹配）

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
        "input": "2+2=?",
        "expected": "4"
      },
      {
        "input": "3+3=?",
        "expected": "6"
      }
    ],
    "accuracy": true,
    "performance": false
  }'
```

**预期响应**:
```json
{
  "results": [
    {
      "input": "1+1=?",
      "expected": "2",
      "actual": "2",
      "passed": true,
      "score": 1.0
    },
    ...
  ],
  "report": "评估报告...",
  "overall_score": 0.95
}
```

### 24. 准确性评估（相似度匹配）

```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{
    "test_cases": [
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

### 25. 性能评估

```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{
    "test_cases": [
      {
        "input": "你好",
        "expected": "你好！"
      },
      {
        "input": "介绍一下Go语言",
        "expected": "Go语言是..."
      }
    ],
    "accuracy": false,
    "performance": true
  }'
```

### 26. 综合评估（准确性+性能）

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
        "input": "什么是AI？",
        "expected": "人工智能"
      }
    ],
    "accuracy": true,
    "performance": true
  }'
```

---

## 🧪 高级测试场景

### 场景1：客服机器人完整流程

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
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "AI Agent Assistant支持语义分块、混合检索和重排序功能",
    "source": "产品手册"
  }'

# 2. 语义搜索相关知识
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "RAG检索方式",
    "top_k": 3
  }'

# 3. RAG增强回答
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "qa-session",
    "message": "系统支持哪些检索方式？"
  }'
```

### 场景3：推理能力测试

```bash
# 1. 思维链推理数学问题
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "如果一个三角形的三边分别是3、4、5，它是什么三角形？请推理"
  }'

# 2. 自我反思改进答案
curl -X POST http://localhost:8080/api/v1/reasoning/reflect \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "解释什么是微服务架构",
    "previous_attempts": [
      "微服务是一种架构",
      "微服务将应用拆分成小服务"
    ]
  }'
```

---

## 📈 监控端点

### 27. Prometheus监控指标

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
- `agent_reasoning_count` - 推理调用次数

---

## 🔍 故障排查

### 检查服务状态

```bash
# 检查服务是否启动
curl http://localhost:8080/health

# 检查已加载的模型
curl http://localhost:8080/api/v1/models

# 查看知识库统计
curl http://localhost:8080/api/v1/knowledge/stats
```

### 常见错误处理

**错误1: Model not available**
```bash
# 检查配置文件中的API密钥
# 确认模型是否已加载
curl http://localhost:8080/api/v1/models
```

**错误2: RAG retrieval failed**
```bash
# 检查是否已添加知识
curl http://localhost:8080/api/v1/knowledge/stats

# 尝试添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{"text": "测试知识", "source": "test"}'
```

---

## 📝 测试建议

### 1. 按顺序测试

建议按照以下顺序进行测试：
1. 健康检查 → 2. 模型管理 → 3. 基础对话 → 4. RAG功能 → 5. 推理能力 → 6. 会话管理 → 7. 记忆管理 → 8. 评估系统

### 2. 重点测试功能

**v0.4核心新功能**:
- ✅ 多模型切换（测试4-6）
- ✅ RAG增强（测试8-12）
- ✅ 推理能力（测试13-16）
- ✅ 会话状态管理（测试17-19）
- ✅ 智能记忆管理（测试20-22）
- ✅ 评估系统（测试23-26）

### 3. 性能测试

使用Apache Bench进行压力测试：
```bash
# 安装ab工具
brew install httpd

# 并发测试
ab -n 1000 -c 10 -T 'application/json' -p test_payload.json http://localhost:8080/api/v1/chat
```

---

## ✅ 测试清单

完成以下测试后，可以确认v0.4功能正常：

- [ ] 健康检查返回v0.4
- [ ] 能列出所有支持的模型（15+模型）
- [ ] 能在不同模型间切换（GLM、千问等）
- [ ] RAG对话能检索到知识
- [ ] 能添加和搜索知识库
- [ ] 思维链推理能返回推理过程
- [ ] 自我反思能改进答案
- [ ] 会话状态能正常更新
- [ ] 记忆提取和搜索功能正常
- [ ] 评估系统能生成报告
- [ ] Prometheus指标正常收集

---

**文档生成时间**: 2026-01-27
**版本**: v0.4
**测试端点**: http://localhost:8080
**监控端点**: http://localhost:9090
