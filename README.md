# AI Agent Assistant

> 智能对话Agent应用后端，支持多模型、工具调用、RAG知识库、持久化存储和HTTP代理

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ v0.3 新特性

- 🗄️ **MySQL持久化** - 完整的数据持久化，支持会话、消息、记忆、工具调用记录
- ⚡ **Redis缓存** - 三级缓存系统（工具结果、LLM响应、会话缓存）
- 🧠 **千问Embedding** - 支持智谱GLM和阿里云千问Embedding模型
- 🔧 **MCP工具系统** - 集成18种MCP工具（搜索、股票、天气、计算等）
- 🌐 **HTTP代理** - 支持代理配置，访问国外API（DuckDuckGo、Yahoo Finance等）
- 📊 **股票查询** - 实时股票价格查询（Yahoo Finance API）

---

## 🎯 核心特性

- **多模型支持** - 智谱GLM-4-Flash、阿里云千问Plus
- **智能工具调用** - 自动检测意图并调用相应工具
- **RAG知识库** - 检索增强生成，支持自定义知识库
- **对话记忆** - 自动管理对话历史，支持多会话隔离
- **流式输出** - 支持实时流式响应
- **RESTful API** - 简洁易用的HTTP接口
- **数据持久化** - MySQL存储所有会话和知识数据
- **高性能缓存** - Redis多级缓存提升响应速度
- **HTTP代理** - 支持代理访问国外API服务

---

## 🛠️ 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| **语言** | Go 1.21+ | 高性能、并发友好 |
| **Web框架** | Gin | 轻量级HTTP框架 |
| **配置管理** | Viper | 支持YAML配置 |
| **数据库** | MySQL 8.0+ | 数据持久化 |
| **缓存** | Redis 7.0+ | 多级缓存系统 |
| **向量存储** | 内存/Milvus | RAG向量数据库 |
| **Embedding** | GLM Embedding-2 / 千问 text-embedding-v3 | 文本向量化 |
| **对话模型** | GLM-4-Flash / Qwen-Plus | 大语言模型 |

---

## 📁 项目结构

```
ai-agent-assistant/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── internal/
│   ├── agent/                   # Agent核心逻辑
│   │   └── agent.go             # 智能体实现
│   ├── cache/                   # Redis缓存系统
│   │   ├── cache.go             # 缓存管理器
│   │   └── cache_test.go        # 缓存测试
│   ├── config/                  # 配置管理
│   │   └── config.go            # 配置加载
│   ├── database/                # MySQL数据库
│   │   ├── mysql.go             # MySQL连接池
│   │   └── repositories/        # 数据仓库层
│   │       ├── sessions.go     # 会话仓储
│   │       ├── messages.go     # 消息仓储
│   │       ├── user_memories.go # 用户记忆仓储
│   │       ├── tool_calls.go   # 工具调用仓储
│   │       └── agent_runs.go   # 运行记录仓储
│   ├── handler/                 # HTTP处理器
│   │   ├── handler.go           # 主处理器
│   │   └── knowledge_handler.go # 知识库处理器
│   ├── llm/                     # 大模型集成
│   │   ├── llm.go               # LLM接口
│   │   ├── glm.go               # 智谱GLM实现
│   │   └── qwen.go              # 阿里云千问实现
│   ├── mcp/                     # MCP工具系统
│   │   ├── client.go            # MCP客户端
│   │   ├── adapter.go           # MCP工具适配器
│   │   ├── http_server.go       # MCP HTTP服务器
│   │   ├── manager.go           # MCP管理器
│   │   └── tools.go             # 18种MCP工具实现
│   ├── memory/                  # 记忆管理
│   │   ├── memory.go            # 记忆管理器
│   │   └── user_memory.go       # 用户记忆
│   ├── rag/                     # RAG知识库
│   │   ├── rag.go               # RAG系统
│   │   ├── embedding/           # 向量化服务
│   │   │   └── embedding.go     # GLM/千问Embedding
│   │   ├── parser/              # 文档解析器
│   │   ├── chunker/             # 文本分块器
│   │   └── store/               # 向量存储
│   ├── tools/                   # 内置工具
│   │   └── tools.go             # 工具管理器
│   └── vectordb/                # 向量数据库
│       ├── milvus.go           # Milvus客户端
│       ├── memory.go           # 内存向量库
│       └── vector.go           # 向量操作
├── pkg/
│   ├── http/                    # HTTP客户端
│   │   └── client.go           # 支持代理的HTTP客户端
│   └── models/                  # 数据模型
│       └── models.go           # 通用数据模型
├── config.yaml                  # 配置文件
├── API_CURL_DOCS.md            # API文档（Curl示例）
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- Redis 7.0+
- 代理软件（可选，用于访问国外API）

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/ai-agent-assistant.git
cd ai-agent-assistant
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置

编辑 `config.yaml` 文件，配置以下内容：

#### 3.1 模型API密钥

```yaml
models:
  glm:
    api_key: "your-glm-api-key"          # 智谱GLM API Key
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    model: "glm-4-flash"

  qwen:
    api_key: "your-qwen-api-key"        # 阿里云千问 API Key
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    model: "qwen-plus"
```

#### 3.2 数据库配置

```yaml
database:
  provider: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    database: "agent_db"
    user: "root"
    password: "your_password"
```

#### 3.3 Redis配置

```yaml
cache:
  enabled: true
  provider: "redis"
  redis:
    addr: "localhost:6379"
    password: "your_redis_password"
```

#### 3.4 HTTP代理（可选，访问国外API）

```yaml
proxy:
  enabled: true
  http_proxy: "http://127.0.0.1:7897"
  https_proxy: "http://127.0.0.1:7897"
  no_proxy: "localhost,127.0.0.1"
```

#### 3.5 工具配置

```yaml
agent:
  default_model: glm              # 默认对话模型
  embedding_model: qwen           # Embedding模型（推荐千问）
```

```yaml
tools:
  enabled:
    - calculator      # 计算器
    - weather         # 天气查询
    - search          # 网络搜索
    - stock_quote     # 股票查询
```

### 4. 初始化数据库

```bash
# 创建数据库
mysql -u root -p < database/schema.sql

# 或使用MySQL命令行
CREATE DATABASE agent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5. 编译并运行

```bash
# 编译
go build -o bin/server cmd/server/main.go

# 运行
./bin/server
```

服务将在 `http://localhost:8080` 启动。

---

## 📡 API接口

### 基础对话

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "message": "你好"
  }'
```

### 启用工具调用

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-123",
    "message": "帮我查询苹果股价",
    "with_tools": true
  }'
```

### RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-123",
    "message": "什么是RAG？"
  }'
```

### 添加知识

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "RAG是检索增强生成技术",
    "source": "RAG介绍"
  }'
```

### 搜索知识库

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "RAG",
    "top_k": 3
  }'
```

### 获取会话历史

```bash
curl "http://localhost:8080/api/v1/session?session_id=user-123"
```

### 清除会话

```bash
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=user-123"
```

### 健康检查

```bash
curl http://localhost:8080/health
```

---

## 🔧 内置工具

| 工具名称 | 功能 | 触发关键词 | 示例 |
|---------|------|-----------|------|
| **calculator** | 数学计算 | 计算、+、-、*、/ | "计算 25*4" |
| **weather** | 天气查询 | 天气、气温、温度 | "北京天气怎么样" |
| **search** | 网络搜索 | 搜索、查一下 | "搜索Go语言" |
| **stock_quote** | 股票查询 | 股票、股价、报价、AAPL | "查询苹果股价" |

**MCP工具**（18种）：
- web_search - 网络搜索（DuckDuckGo）
- github_search - GitHub搜索
- stock_quote - 股票报价
- stock_info - 股票信息
- weather - 天气查询
- calculate - 计算器
- file_read - 文件读取
- uuid_generate - UUID生成
- hash_generate - 哈希生成
- get_timestamp - 时间戳
- url_encode_decode - URL编解码
- base64_encode_decode - Base64编解码
- json_format - JSON格式化
- ip_lookup - IP查询
- whois - WHOIS查询
- http_request - HTTP请求
- text_process - 文本处理
- unit_convert - 单位转换

---

## 📊 配置说明

### Server 配置

```yaml
server:
  port: 8080
  mode: debug              # debug, release, test
```

### Agent 配置

```yaml
agent:
  default_model: glm         # glm 或 qwen
  embedding_model: qwen      # glm 或 qwen（推荐千问）
  max_tokens: 2000
  temperature: 0.7
  enable_stream: true
```

### Memory 配置

```yaml
memory:
  max_history: 10            # 最大历史记录数
  store_type: mysql         # 存储类型：memory, mysql, redis
  enable_user_memory: true  # 启用用户记忆
  enable_state_memory: true # 启用状态记忆
```

### RAG 配置

```yaml
rag:
  enabled: true
  top_k: 3                  # 检索TopK数量
  threshold: 0.3            # 相似度阈值
  chunk_size: 500           # 分块大小
  chunk_overlap: 50         # 分块重叠
```

### 缓存配置

```yaml
cache:
  enabled: true
  provider: redis
  redis:
    tool_result_ttl: "1h"      # 工具结果缓存1小时
    llm_response_ttl: "5m"     # LLM响应缓存5分钟
    session_ttl: "24h"         # 会话缓存24小时
```

---

## 🧪 测试

### 运行单元测试

```bash
# 测试所有功能
go test ./...

# 测试特定包
go test ./internal/cache/...
go test ./internal/mcp/...
go test ./internal/tools/...
```

### 手动API测试

参考 [API_CURL_DOCS.md](API_CURL_DOCS.md) 获取完整的API测试示例。

---

## 📝 开发路线图

- [x] v0.1 - 基础对话功能
- [x] v0.2 - RAG知识库和工具调用
- [x] v0.3 - **MySQL持久化、Redis缓存、千问Embedding、MCP工具、HTTP代理** ✅
- [ ] v0.4 - 支持更多大模型（Claude、GPT等）
- [ ] v0.5 - Agent编排和工作流
- [ ] v0.6 - 多模态支持（图片、文件）
- [ ] v0.7 - 分布式部署支持

---

## 🤝 贡献

欢迎提交Issue和Pull Request！

---

## 📄 License

MIT License

---

## 🔗 相关文档

- [API文档（Curl示例）](API_CURL_DOCS.md)
- [RAG功能指南](RAG_GUIDE.md)
- [MCP工具文档](internal/mcp/README.md)

---

## 📮 联系方式

- 项目主页: [GitHub Repository]
- Issue跟踪: [Issues]

---

**版本**: v0.3
**最后更新**: 2026-01-27
