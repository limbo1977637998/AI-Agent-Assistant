# AI Agent Assistant - Curl API 文档

服务器地址: `http://localhost:8080`

## 目录
- [健康检查](#健康检查)
- [聊天接口](#聊天接口)
- [会话管理](#会话管理)
- [知识库管理](#知识库管理)
- [MCP工具列表](#mcp工具列表)

---

## 健康检查

### 检查服务健康状态

```bash
curl -X GET http://localhost:8080/health
```

**响应示例:**
```json
{
  "status": "healthy",
  "model": "glm",
  "rag_enabled": true,
  "tools_enabled": ["calculator", "weather", "search"]
}
```

---

## 聊天接口

### 1. 普通对话

使用大模型进行对话，支持工具调用。

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请帮我计算 25 * 4 + 10",
    "session_id": "user_123"
  }'
```

**请求参数:**
- `message` (string, 必需): 用户消息
- `session_id` (string, 可选): 会话ID，用于保持上下文

**响应示例:**
```json
{
  "response": "根据计算，25 * 4 + 10 = 110",
  "session_id": "user_123",
  "tool_calls": [
    {
      "tool": "calculator",
      "arguments": {"expression": "25 * 4 + 10"},
      "result": "110"
    }
  ],
  "cached": false
}
```

### 2. RAG增强对话

使用知识库增强的对话，适合需要查询知识库的场景。

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "message": "什么是RAG？",
    "session_id": "user_123",
    "top_k": 3
  }'
```

**请求参数:**
- `message` (string, 必需): 用户消息
- `session_id` (string, 可选): 会话ID
- `top_k` (int, 可选): 检索知识数量，默认3

**响应示例:**
```json
{
  "response": "RAG（Retrieval-Augmented Generation）是一种结合检索和生成的AI技术...",
  "session_id": "user_123",
  "retrieved_knowledge": [
    {
      "content": "RAG是将向量检索与大模型结合的技术",
      "score": 0.85
    }
  ],
  "tool_calls": [],
  "cached": false
}
```

---

## 会话管理

### 1. 获取会话历史

获取指定会话的消息历史。

```bash
curl -X GET "http://localhost:8080/api/v1/session?session_id=user_123"
```

**请求参数:**
- `session_id` (string, 必需): 会话ID

**响应示例:**
```json
{
  "session_id": "user_123",
  "messages": [
    {
      "role": "user",
      "content": "你好",
      "timestamp": "2026-01-27T17:00:00+08:00"
    },
    {
      "role": "assistant",
      "content": "你好！有什么可以帮助你的吗？",
      "timestamp": "2026-01-27T17:00:01+08:00"
    }
  ],
  "total_messages": 2
}
```

### 2. 清除会话

清除指定会话的历史记录。

```bash
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=user_123"
```

**请求参数:**
- `session_id` (string, 必需): 会话ID

**响应示例:**
```json
{
  "success": true,
  "message": "Session cleared successfully"
}
```

---

## 知识库管理

### 1. 添加文本知识

向知识库添加文本内容。

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "人工智能是计算机科学的一个分支，致力于创建能够执行通常需要人类智能的任务的系统。",
    "source": "wiki"
  }'
```

**请求参数:**
- `text` (string, 必需): 知识内容
- `source` (string, 可选): 知识来源，默认为 "manual"

**响应示例:**
```json
{
  "code": 200,
  "message": "success"
}
```

### 2. 从文档添加知识

上传文档并添加到知识库（支持txt, md, pdf等格式）。

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add/doc \
  -F "file=@/path/to/document.pdf" \
  -F "metadata={\"source\": \"company_doc\", \"category\": \"technical\"}"
```

**请求参数:**
- `file` (file, 必需): 文档文件
- `metadata` (string, 可选): 元数据JSON字符串

**响应示例:**
```json
{
  "success": true,
  "chunks_added": 15,
  "message": "Document processed and added to knowledge base"
}
```

### 3. 知识库统计

获取知识库统计信息。

```bash
curl -X GET http://localhost:8080/api/v1/knowledge/stats
```

**响应示例:**
```json
{
  "total_vectors": 1250,
  "collection_name": "agent_knowledge",
  "dimension": 1024,
  "index_type": "HNSW",
  "metric_type": "COSINE"
}
```

### 4. 搜索知识库

在知识库中搜索相关内容。

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是机器学习？",
    "top_k": 5
  }'
```

**请求参数:**
- `query` (string, 必需): 搜索查询
- `top_k` (int, 可选): 返回结果数量，默认5

**响应示例:**
```json
{
  "results": [
    {
      "id": 12345,
      "content": "机器学习是人工智能的一个子集...",
      "score": 0.92,
      "metadata": {
        "source": "wiki",
        "category": "AI"
      }
    },
    {
      "id": 12346,
      "content": "深度学习是机器学习的一种方法...",
      "score": 0.88,
      "metadata": {
        "source": "textbook",
        "category": "ML"
      }
    }
  ],
  "total_results": 2
}
```

---

## MCP工具列表

系统集成了18种MCP工具，可以在对话中自动调用：

### 工具列表

1. **web_search** - 在互联网上搜索信息（使用DuckDuckGo）
2. **github_search** - 搜索GitHub仓库和代码
3. **stock_quote** - 获取股票实时报价（使用Yahoo Finance）
4. **stock_info** - 获取股票基本信息
5. **weather** - 查询城市天气信息
6. **calculate** - 执行数学计算
7. **file_read** - 读取本地文件内容
8. **uuid_generate** - 生成UUID（唯一标识符）
9. **hash_generate** - 生成字符串的哈希值（支持MD5、SHA1、SHA256、SHA512）
10. **get_timestamp** - 获取当前时间戳和格式化时间
11. **url_encode_decode** - URL编码或解码字符串
12. **base64_encode_decode** - Base64编码或解码字符串
13. **json_format** - 格式化JSON字符串
14. **ip_lookup** - 查询IP地址的地理位置信息
15. **whois** - 查询域名WHOIS信息
16. **http_request** - 发送HTTP GET请求获取网页内容
17. **text_process** - 文本处理（大小写转换、倒序、字数统计等）
18. **unit_convert** - 单位转换（长度、重量、温度等）

### 工具使用示例

#### 使用工具调用股票查询

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "帮我查询苹果公司(AAPL)的股价",
    "session_id": "user_123"
  }'
```

#### 使用工具查询天气

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "北京今天天气怎么样？",
    "session_id": "user_123"
  }'
```

#### 使用工具搜索网络

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "帮我搜索一下Go语言最新版本",
    "session_id": "user_123"
  }'
```

---

## 错误处理

所有API在发生错误时都会返回统一的错误格式：

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": {}
}
```

常见错误码：
- `INVALID_REQUEST`: 请求参数无效
- `SESSION_NOT_FOUND`: 会话不存在
- `TOOL_EXECUTION_FAILED`: 工具执行失败
- `KNOWLEDGE_NOT_FOUND`: 知识库未找到
- `INTERNAL_ERROR`: 服务器内部错误

---

## 性能优化

### 缓存机制

系统实现了两级缓存：
1. **工具结果缓存** (TTL: 1小时) - 相同的工具调用参数会返回缓存结果
2. **LLM响应缓存** (TTL: 5分钟) - 相同的查询会返回缓存的AI响应

缓存命中时响应会包含 `"cached": true` 标记。

### 批量操作

对于批量添加知识，建议：
- 单次添加内容不超过1000字
- 批量添加时控制请求频率，避免过载

---

## 配置说明

当前配置（config.yaml）：
- **服务端口**: 8080
- **运行模式**: debug
- **默认模型**: glm-4-flash
- **向量数据库**: memory (内存模式)
- **缓存**: Redis (localhost:6379)
- **持久化**: MySQL (localhost:3306)

修改配置后需重启服务生效。

---

## 注意事项

1. **首次启动**前请确保以下服务已启动：
   - MySQL (端口3306)
   - Redis (端口6379)
   - （可选）Milvus (端口19530)

2. **会话管理**：
   - 建议为每个用户使用唯一的session_id
   - 会话历史默认保存在MySQL中
   - 可以通过DELETE /session清除历史

3. **知识库**：
   - 使用内存向量数据库时，知识在服务重启后丢失
   - 生产环境建议配置Milvus进行持久化
   - 支持txt、md、pdf等格式的文档

4. **工具调用**：
   - 工具会根据用户输入自动选择和调用
   - 某些工具（如天气、股票）需要网络连接
   - 工具调用失败不会中断对话流程

---

## 快速开始

1. 启动服务：
```bash
./bin/server
```

2. 测试健康检查：
```bash
curl http://localhost:8080/health
```

3. 发起第一次对话：
```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好"}'
```

4. 添加知识到知识库：
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{"text": "我的第一个知识条目", "source": "manual"}'
```

5. 使用RAG对话：
```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{"message": "我添加了什么知识？"}'
```

---

生成时间: 2026-01-27
文档版本: v1.0
