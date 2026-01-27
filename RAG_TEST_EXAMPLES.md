# RAG功能测试案例

## 🚀 快速测试（5个简单案例）

### 案例1: 添加知识并查询

**步骤1: 添加知识**
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Vue.js是一个渐进式JavaScript框架，用于构建用户界面。特点：响应式数据绑定、组件化开发、虚拟DOM。由尤雨溪开发。",
    "source": "Vue框架介绍"
  }'
```

**步骤2: 使用RAG对话**
```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-vue",
    "message": "Vue.js有什么特点？",
    "model": "glm"
  }'
```

**预期结果**: AI会基于你添加的知识回答，提到"渐进式"、"响应式"、"尤雨溪"等关键词。

---

### 案例2: 多知识测试

**步骤1: 添加多条知识**
```bash
# 添加React知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "React是Facebook开发的JavaScript库，用于构建用户界面。特点：组件化、虚拟DOM、单向数据流、JSX语法。",
    "source": "React介绍"
  }'

# 添加Angular知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Angular是Google开发的完整前端框架。特点：TypeScript、依赖注入、双向数据绑定、完整解决方案。",
    "source": "Angular介绍"
  }'
```

**步骤2: 对比测试**
```bash
# RAG对话
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-frontend",
    "message": "React和Angular有什么区别？",
    "model": "glm"
  }'
```

---

### 案例3: 搜索知识库

**步骤1: 搜索知识**
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "前端框架",
    "top_k": 5
  }'
```

**预期结果**: 返回包含"React"、"Angular"、"Vue"等相关知识的文本片段。

---

### 案例4: 查看知识库统计

```bash
curl http://localhost:8080/api/v1/knowledge/stats
```

**预期输出**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "type": "memory",
    "vector_count": 5,
    "dimension": 1024
  }
}
```

---

### 案例5: 创建文档并添加

**步骤1: 创建测试文档**
```bash
cat > /tmp/test_kb.txt << 'EOF'
微服务架构是一种将应用程序拆分为一组小型服务的方法。
每个服务运行在自己的进程中，通过轻量级机制（通常是HTTP API）通信。
微服务的优势包括：技术栈灵活、独立部署、易于扩展、团队自治。
挑战包括：分布式复杂性、数据一致性、服务发现、监控调试。
EOF
```

**步骤2: 从文档添加知识**
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -H "Content-Type: application/json" \
  -d '{
    "doc_path": "/tmp/test_kb.txt"
  }'
```

**步骤3: 使用知识库**
```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-microservice",
    "message": "微服务架构有哪些优势和挑战？",
    "model": "glm"
  }'
```

---

## 🔍 观察RAG效果

### 对比测试

**不使用RAG（普通对话）**:
```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "no-rag",
    "message": "Vue.js是什么？",
    "model": "glm"
  }'
```

**使用RAG**:
```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "with-rag",
    "message": "Vue.js是什么？",
    "model": "glm"
  }'
```

**差异**:
- 不使用RAG: 使用模型自身知识，回答可能通用但不具体
- 使用RAG: 基于你添加的知识库，回答更详细、更贴合你的需求

---

## 📊 完整测试流程

### 一键测试所有功能

```bash
# 运行完整测试脚本
./test_rag.sh
```

### 手动逐步测试

```bash
# 1. 健康检查
curl http://localhost:8080/health

# 2. 添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{"text":"测试内容","source":"测试"}'

# 3. 查看统计
curl http://localhost:8080/api/v1/knowledge/stats

# 4. 搜索知识
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{"query":"测试","top_k":3}'

# 5. RAG对话
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test","message":"测试问题","model":"glm"}'
```

---

## 💡 测试建议

### 推荐测试顺序

1. **先测试健康检查** - 确保服务正常
2. **添加1-2条知识** - 建立基础知识库
3. **查看统计** - 确认知识已添加
4. **测试RAG对话** - 验证检索功能
5. **对比测试** - 体验RAG的效果差异

### 常见问题

**Q: 知识库重启后会清空吗？**
A: 是的，当前使用内存存储，重启后清空

**Q: 如何持久化知识库？**
A: 需要实现Redis持久化（待开发）

**Q: 支持哪些文档格式？**
A: 当前支持TXT、MD，PDF待实现

**Q: 可以添加多少知识？**
A: 当前版本适合中小规模（< 1000条）

---

## 🎯 快速验证

想快速验证RAG是否工作？执行这三条命令：

```bash
# 1. 添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{"text":"测试：地球是圆的","source":"地理知识"}'

# 2. 查看统计（应该显示vector_count: 1）
curl http://localhost:8080/api/v1/knowledge/stats

# 3. RAG对话
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{"session_id":"quick-test","message":"地球是什么形状？","model":"glm"}'
```

如果第3步的回答中提到了"圆的"，说明RAG工作正常！✅
