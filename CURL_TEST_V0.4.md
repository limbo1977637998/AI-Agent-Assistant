# v0.4 新功能测试命令（精简版）

> 版本: v0.4 | 测试端点: http://localhost:8080

## ⚡ 快速开始

### 1. 启动服务（使用稳定的main.go）

```bash
cd /Users/gongpengfei/Desktop/ClaudeCode/ai-agent-assistant
go run cmd/server/main.go
```

---

## 🎯 v0.4新功能测试命令

### ✅ 功能1: 统一模型抽象层

**测试A: 列出所有支持的模型（15+模型）**
```bash
curl http://localhost:8080/api/v1/models
```

**测试B: 使用不同模型对话**
```bash
# 使用GLM模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "model-test-1",
    "message": "用三句话介绍Go语言",
    "model": "glm"
  }'

# 使用千问模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "model-test-2",
    "message": "用三句话介绍Go语言",
    "model": "qwen"
  }'
```

---

### ✅ 功能2: RAG增强（语义分块+混合检索+重排序）

**测试A: 添加知识**
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "AI Agent Assistant v0.4支持15+大模型，包括GLM-4-Flash/Plus、千问-Turbo/Plus/Max、GPT-3.5/4、Claude-3.5-Sonnet、DeepSeek-Chat等。具备RAG增强、推理能力、智能记忆等特性。",
    "source": "v0.4特性介绍"
  }'
```

**测试B: RAG增强对话**
```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "rag-user-001",
    "message": "v0.4支持哪些模型？有什么新特性？",
    "top_k": 3
  }'
```

**测试C: 搜索知识库**
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "语义分块",
    "top_k": 3
  }'
```

**测试D: 知识库统计**
```bash
curl http://localhost:8080/api/v1/knowledge/stats
```

---

### ✅ 功能3: 推理能力（思维链+自我反思）

**测试A: 思维链推理**
```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "请逐步推理：一个农场有鸡和兔共50只，有140条腿，鸡和兔各多少只？"
  }'
```

**预期输出:**
```json
{
  "reasoning": "【思考过程】\n设鸡有x只，兔有y只\nx+y=50\n2x+4y=140\n解方程得：x=30, y=20",
  "answer": "农场有30只鸡和20只兔子。"
}
```

**测试B: 数学计算推理**
```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "25 * 4 = ? 请详细说明计算过程"
  }'
```

**测试C: 自我反思**
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

### ✅ 功能4: 智能记忆管理

**测试A: 自动提取记忆**
```bash
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "user-alice",
    "conversation": "用户：我叫Alice，来自北京，是个软件工程师，主要用Python工作，最近在学习Go语言。\\n助手：你好Alice！北京是个美丽的城市。\\n用户：是的，我还在学习Rust。"
  }'
```

**预期输出:**
```json
{
  "message": "Memories extracted",
  "count": 1,
  "memories": [
    {
      "id": "mem_xxx",
      "user_id": "user-alice",
      "content": "用户名叫Alice，来自北京，是个软件工程师，主要使用Python，正在学习Go和Rust",
      "topics": ["个人信息", "职业", "位置", "技能"],
      "importance": 0.85
    }
  ]
}
```

**测试B: 语义搜索记忆**
```bash
curl "http://localhost:8080/api/v1/memory/search?user_id=user-alice&query=编程语言&limit=3"
```

---

### ✅ 功能5: 会话状态管理

**测试A: 获取会话信息**
```bash
curl "http://localhost:8080/api/v1/session?session_id=user-alice"
```

**测试B: 更新会话状态**
```bash
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-alice",
    "updates": {
      "username": "alice",
      "theme": "dark",
      "language": "zh",
      "preferences": {
        "notifications": true
      }
    }
  }'
```

**测试C: 清除会话**
```bash
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=test-session"
```

---

### ✅ 功能6: 评估系统

**测试A: 准确性评估**
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
        "input": "中国首都是哪里？",
        "expected": "北京"
      }
    ],
    "accuracy": true,
    "performance": false
  }'
```

**测试B: 性能评估**
```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{
    "test_cases": [
      {
        "input": "你好",
        "expected": "你好！"
      }
    ],
    "accuracy": false,
    "performance": true
  }'
```

---

## 🏥 基础测试

**健康检查**
```bash
curl http://localhost:8080/health
```

---

## 📊 完整测试流程（推荐顺序）

```bash
# 1. 健康检查
curl http://localhost:8080/health

# 2. 查看支持的模型
curl http://localhost:8080/api/v1/models

# 3. 添加测试知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{"text": "AI Agent Assistant v0.4支持15+大模型", "source": "测试"}'

# 4. RAG对话测试
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{"session_id": "test", "message": "支持哪些模型？"}'

# 5. 推理能力测试
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{"task": "25*4=? 请推理"}'

# 6. 记忆提取测试
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{"user_id": "test", "conversation": "我叫张三，是个程序员"}'

# 7. 记忆搜索测试
curl "http://localhost:8080/api/v1/memory/search?user_id=test&query=职业&limit=3"

# 8. 会话状态测试
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{"session_id": "test", "updates": {"name": "测试用户"}}'

# 9. 评估系统测试
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{"test_cases": [{"input": "1+1=?", "expected": "2"}], "accuracy": true}'
```

---

## 🔍 故障排查

### 如果遇到"Model not available"
```bash
# 检查配置文件中的API密钥
cat config.yaml | grep api_key

# 查看已加载的模型
curl http://localhost:8080/api/v1/models
```

### 如果RAG检索无结果
```bash
# 先添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{"text": "测试知识内容", "source": "测试"}'

# 查看知识库统计
curl http://localhost:8080/api/v1/knowledge/stats
```

### 如果推理功能不可用
```bash
# 确认模型配置
cat config.yaml | grep -A5 "agent:"
```

---

## ✅ 测试清单

测试完成后，确认以下功能正常：

- [ ] 健康检查返回v0.4
- [ ] 能列出15+支持模型
- [ ] 能切换不同模型对话
- [ ] 能添加和检索知识
- [ ] RAG对话能返回相关知识
- [ ] 思维链推理能返回推理过程
- [ ] 自我反思能改进答案
- [ ] 能自动提取用户记忆
- [ ] 能语义搜索记忆
- [ ] 会话状态能正常更新
- [ ] 评估系统能生成报告

---

## 📝 测试建议

1. **按顺序测试** - 先健康检查 → 模型管理 → 对话 → RAG → 推理 → 记忆 → 会话 → 评估

2. **记录结果** - 每个测试记录返回结果，便于分析问题

3. **错误处理** - 遇到错误先查看服务端日志

4. **API密钥** - 确保config.yaml中配置了有效的API密钥

---

**最后更新**: 2026-01-27
**版本**: v0.4
**测试端点**: http://localhost:8080
**完整文档**: 参考 TEST_API_V0.4.md（500+行完整版）
