# AI Agent Assistant

> 智能对话Agent应用后端，支持多模型、工具调用、RAG知识库、持久化存储和HTTP代理

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-v0.4-green.svg)](https://github.com/yourusername/ai-agent-assistant)

## ✨ v0.4 新特性

- 🌟 **统一模型抽象** - 支持18+主流模型（GLM、千问、OpenAI、Claude、DeepSeek等）
- 🧠 **RAG增强** - 语义分块、混合检索（向量+BM25）、重排序
- 🤔 **推理能力** - 思维链推理、自我反思、多步推理
- 📊 **评估监控** - 智能评分（包含关系识别）、准确性/性能评估、OpenTelemetry追踪、Prometheus监控
- 💾 **智能记忆** - 自动提取、语义检索、优化策略（时间衰减、重要性评分）
- 🔄 **会话增强** - 自动摘要、并发控制、状态管理（版本控制）

---

## 🎯 核心特性

### 模型与推理
- **多模型支持** - GLM-4、千问、GPT-4、Claude、DeepSeek等18+模型
- **统一模型接口** - 无缝切换模型，统一API调用
- **推理能力增强** - 思维链推理、自我反思、多步推理
- **智能评估系统** - 包含关系识别、相似度评分、LLM判断

### 知识与记忆
- **RAG知识库** - 检索增强生成，语义分块、混合检索、重排序
- **智能记忆** - 自动提取、语义检索、记忆优化
- **会话管理** - 自动摘要、并发控制、状态版本管理
- **对话记忆** - 自动管理对话历史，支持多会话隔离

### 工具与集成
- **智能工具调用** - 自动检测意图并调用相应工具
- **MCP工具系统** - 18种内置工具（计算器、天气、搜索、股票等）
- **流式输出** - 支持实时流式响应
- **RESTful API** - 简洁易用的HTTP接口（16个端点）

### 基础设施
- **数据持久化** - MySQL存储会话、消息、记忆、知识等
- **高性能缓存** - Redis多级缓存提升响应速度
- **HTTP代理** - 支持代理访问国外API服务
- **监控追踪** - OpenTelemetry分布式追踪、Prometheus指标收集

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
| **对话模型** | GLM-4-Flash / Qwen-Plus / GPT-4 / Claude / DeepSeek | 大语言模型 |
| **监控** | OpenTelemetry + Prometheus | 分布式追踪和指标收集 |

---

## 📁 项目结构

```
ai-agent-assistant/
├── cmd/
│   └── server/
│       ├── main.go              # 主程序入口（简化版）
│       └── main_full.go         # 完整版服务器（所有v0.4功能）
├── internal/
│   ├── agent/                   # Agent核心逻辑
│   ├── cache/                   # Redis缓存系统
│   ├── config/                  # 配置管理
│   ├── database/                # MySQL数据库
│   │   └── repositories/        # 数据仓库层
│   ├── eval/                    # 评估系统
│   │   ├── evaluator.go         # 准确性评估
│   │   └── performance_eval.go  # 性能评估
│   ├── handler/                 # HTTP处理器
│   ├── llm/                     # 统一模型接口
│   │   ├── model.go             # 模型接口定义
│   │   ├── factory.go           # 模型工厂
│   │   ├── glm_model.go         # GLM实现
│   │   ├── qwen_model.go        # 千问实现
│   │   ├── openai.go            # OpenAI实现
│   │   ├── claude.go            # Claude实现
│   │   └── deepseek.go          # DeepSeek实现
│   ├── mcp/                     # MCP工具系统
│   │   ├── client.go            # MCP客户端
│   │   ├── adapter.go           # MCP工具适配器
│   │   └── tools.go             # 18种MCP工具实现
│   ├── memory/                  # 记忆管理
│   │   ├── enhanced_memory.go   # 增强记忆管理
│   │   └── enhanced_session.go  # 增强会话管理
│   ├── monitoring/              # 监控系统
│   │   ├── metrics.go           # Prometheus指标
│   │   └── server.go            # 监控服务器
│   ├── rag/                     # RAG知识库
│   │   ├── rag_enhanced.go      # 增强RAG系统
│   │   ├── chunker/             # 文本分块器
│   │   │   └── semantic_chunker.go  # 语义分块
│   │   ├── retriever/           # 检索器
│   │   │   ├── bm25.go          # BM25关键词检索
│   │   │   └── hybrid.go        # 混合检索
│   │   └── reranker/            # 重排序器
│   ├── reasoning/               # 推理能力
│   │   ├── chain_of_thought.go  # 思维链推理
│   │   ├── reflection.go        # 自我反思
│   │   └── reasoning_manager.go # 推理管理器
│   ├── tools/                   # 内置工具
│   ├── tracing/                 # OpenTelemetry追踪
│   └── vectordb/                # 向量数据库
├── pkg/
│   ├── http/                    # HTTP客户端
│   └── models/                  # 数据模型
├── database/
│   └── schema.sql               # 数据库Schema
├── config.yaml.example          # 配置文件模板
├── EXAMPLES.md                  # 使用示例
├── USAGE_GUIDE.md               # 使用指南
├── TEST_V0.4_COMPLETE.md        # API测试文档
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+ (可选，用于持久化)
- Redis 7.0+ (可选，用于缓存)
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

```bash
# 复制配置模板
cp config.yaml.example config.yaml

# 编辑配置文件
vim config.yaml
```

#### 3.1 模型API密钥（必须）

```yaml
models:
  glm:
    api_key: "YOUR_GLM_API_KEY"        # 智谱GLM API Key
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    model: "glm-4-flash"

  qwen:
    api_key: "YOUR_QWEN_API_KEY"       # 阿里云千问 API Key
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    model: "qwen-plus"
```

**获取API密钥**：
- GLM: https://open.bigmodel.cn/
- 千问: https://dashscope.aliyuncs.com/

#### 3.2 数据库配置（可选）

```yaml
database:
  provider: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    database: "agent_db"
    user: "root"
    password: "YOUR_MYSQL_PASSWORD"
```

#### 3.3 Redis配置（可选）

```yaml
cache:
  enabled: true
  provider: "redis"
  redis:
    addr: "localhost:6379"
    password: "YOUR_REDIS_PASSWORD"
```

#### 3.4 HTTP代理（可选，访问国外API）

```yaml
proxy:
  enabled: true
  http_proxy: "http://127.0.0.1:7897"
  https_proxy: "http://127.0.0.1:7897"
```

### 4. 初始化数据库（可选）

```bash
# 创建数据库
mysql -u root -p < database/schema.sql
```

### 5. 编译并运行

```bash
# 编译
go build -o bin/server cmd/server/main_full.go

# 运行完整版
./bin/server
```

服务将在 `http://localhost:8080` 启动。

---

## 📡 API接口

### 健康检查

```bash
curl http://localhost:8080/health
```

### 基础对话（支持多模型切换）

```bash
# 使用GLM模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "message": "你好",
    "model": "glm"
  }'

# 使用千问模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "user-123",
    "message": "你好",
    "model": "qwen"
  }'
```

### RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-123",
    "message": "什么是RAG？",
    "top_k": 3
  }'
```

### 推理能力（思维链）

```bash
curl -X POST http://localhost:8080/api/v1/reasoning/cot \
  -H 'Content-Type: application/json' \
  -d '{
    "task": "计算：5 + 3 * 2 = ? 并详细说明步骤"
  }'
```

### 知识库管理

```bash
# 添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "RAG是检索增强生成技术",
    "source": "RAG介绍"
  }'

# 搜索知识库
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "RAG",
    "top_k": 3
  }'

# 查看知识库统计
curl http://localhost:8080/api/v1/knowledge/stats
```

### 会话管理

```bash
# 获取会话信息
curl "http://localhost:8080/api/v1/session?session_id=user-123"

# 更新会话状态
curl -X POST http://localhost:8080/api/v1/session/state \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": "user-123",
    "updates": {
      "user_name": "Alice",
      "topic": "AI讨论"
    }
  }'

# 清除会话
curl -X DELETE "http://localhost:8080/api/v1/session?session_id=user-123"
```

### 智能记忆

```bash
# 提取记忆
curl -X POST http://localhost:8080/api/v1/memory/extract \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "alice",
    "conversation": "我是Alice，是一名软件工程师。"
  }'

# 搜索记忆
curl "http://localhost:8080/api/v1/memory/search?user_id=alice&query=工作&limit=5"
```

### 评估系统

```bash
curl -X POST http://localhost:8080/api/v1/eval/accuracy \
  -H 'Content-Type: application/json' \
  -d '{
    "test_cases": [
      {
        "input": "2+2等于几？",
        "expected_output": "4"
      }
    ],
    "accuracy": true
  }'
```

### 模型管理

```bash
# 查看所有支持的模型
curl http://localhost:8080/api/v1/models

# 查看特定模型信息
curl http://localhost:8080/api/v1/models/glm
```

---

## 🔧 内置工具

| 工具名称 | 功能 | 触发关键词 | 示例 |
|---------|------|-----------|------|
| **calculator** | 数学计算 | 计算、+、-、*、/ | "计算 25*4" |
| **weather** | 天气查询 | 天气、气温、温度 | "北京天气怎么样" |
| **search** | 网络搜索 | 搜索、查一下 | "搜索Go语言" |
| **stock_quote** | 股票查询 | 股票、股价、报价 | "查询苹果股价" |

**MCP工具**（18种）：
- web_search, github_search, stock_quote, stock_info
- weather, calculate, file_read, uuid_generate
- hash_generate, get_timestamp, url_encode_decode
- base64_encode_decode, json_format, ip_lookup
- whois, http_request, text_process, unit_convert

---

## 📊 v0.4 新功能详解

### 1. 统一模型抽象层

支持18+种主流模型，通过工厂模式统一管理：

```go
// 自动识别模型类型
modelManager, _ := llm.NewModelManager(cfg)
model, _ := modelManager.GetModel("gpt-4")  // 或 claude, deepseek等
response, _ := model.Chat(ctx, messages)
```

**支持的模型**：
- GLM系列: glm-4-flash, glm-4-plus, glm-4-alltools
- 千问系列: qwen-turbo, qwen-plus, qwen-max, qwen-long
- OpenAI系列: gpt-3.5-turbo, gpt-4, gpt-4-turbo, gpt-4o
- Claude系列: claude-3-5-sonnet, claude-3-opus, claude-3-haiku
- DeepSeek系列: deepseek-chat, deepseek-coder, deepseek-r1

### 2. RAG增强

- **语义分块**：基于Embedding相似度智能分块
- **混合检索**：向量检索 + BM25关键词检索
- **重排序**：Cross-Encoder重排序提升准确度

```bash
# 混合检索示例
POST /api/v1/knowledge/search
{
  "query": "Go语言特性",
  "top_k": 5,
  "rerank": true
}
```

### 3. 推理能力增强

- **思维链推理**：逐步展示推理过程
- **自我反思**：多轮迭代优化答案
- **多步推理**：复杂任务分解

```bash
# 思维链推理
POST /api/v1/reasoning/cot
{
  "task": "解释什么是递归，并给出例子"
}
```

### 4. 智能评估系统

- **包含关系识别**：自动识别"包含式"答案（如期望"4"，实际"2+2=4"）
- **智能评分**：支持完全匹配、包含匹配、编辑距离三层评分
- **多维度评估**：准确性、性能、可靠性

```bash
# 评估示例
POST /api/v1/eval/accuracy
{
  "test_cases": [
    {
      "input": "2+2等于几？",
      "expected_output": "4"  # 即使实际是"2 + 2 = 4"也能识别为正确
    }
  ]
}
```

### 5. 智能记忆管理

- **自动提取**：从对话中自动提取关键信息
- **语义检索**：基于向量相似度的记忆搜索
- **优化策略**：时间衰减、重要性评分

### 6. 会话状态管理

- **版本控制**：每次状态更新递增版本号
- **自动摘要**：长对话自动生成摘要
- **并发控制**：使用读写锁保证并发安全

---

## 🧪 测试

### 运行单元测试

```bash
# 测试所有功能
go test ./...

# 测试特定包
go test ./internal/llm/...
go test ./internal/eval/...
go test ./internal/reasoning/...
go test ./internal/memory/...
```

### 完整API测试

参考 [TEST_V0.4_COMPLETE.md](TEST_V0.4_COMPLETE.md) 获取完整的API测试示例，包含16个端点的详细测试命令。

---

## 📝 版本历史

### v0.4 (2026-01-28) ✅

**新增功能**：
- ✨ 统一模型抽象层，支持18+种模型
- ✨ RAG增强：语义分块、混合检索、重排序
- ✨ 推理能力：思维链、自我反思、多步推理
- ✨ 评估系统：智能评分、准确性/性能评估
- ✨ 智能记忆：自动提取、语义检索、优化策略
- ✨ 会话增强：自动摘要、状态版本管理

**优化**：
- 🎯 评分算法支持包含关系识别
- 🎯 添加config.yaml.example模板
- 🎯 完善API文档和测试用例

### v0.3 (2026-01-27)

- MySQL持久化、Redis缓存、千问Embedding、MCP工具、HTTP代理

### v0.2 (2026-01-26)

- RAG知识库和工具调用

### v0.1 (2026-01-25)

- 基础对话功能

---

## 🔮 开发路线图

- [x] v0.1 - 基础对话功能
- [x] v0.2 - RAG知识库和工具调用
- [x] v0.3 - MySQL持久化、Redis缓存、千问Embedding、MCP工具、HTTP代理
- [x] v0.4 - **统一模型抽象、RAG增强、推理能力、评估监控、智能记忆、会话增强** ✅
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

- [完整测试文档](TEST_V0.4_COMPLETE.md) - 16个API端点完整测试指南
- [使用示例](EXAMPLES.md) - 详细使用示例
- [使用指南](USAGE_GUIDE.md) - 功能使用指南
- [数据库Schema](database/schema.sql) - 数据库结构

---

## 📮 联系方式

- 项目主页: [GitHub Repository]
- Issue跟踪: [Issues]

---

**版本**: v0.4
**最后更新**: 2026-01-28
**状态**: 生产就绪 ✅
