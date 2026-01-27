# AI Agent Assistant

智能对话Agent应用后端，支持多模型切换、工具调用、对话记忆和RAG知识库。

## 特性

- **多模型支持**: 支持智谱GLM-4和阿里云千问模型
- **工具调用**: 内置计算器、天气查询、网络搜索等工具
- **对话记忆**: 自动管理对话历史，支持多会话
- **RAG知识库**: 检索增强生成，支持自定义知识库
- **流式输出**: 支持流式和非流式响应
- **RESTful API**: 简洁的HTTP接口设计

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **配置管理**: Viper
- **HTTP客户端**: 标准库 net/http
- **向量化**: GLM Embedding-2
- **向量存储**: 内存向量数据库

## 项目结构

```
ai-agent-assistant/
├── cmd/
│   └── server/
│       └── main.go          # 主程序入口
├── internal/
│   ├── agent/               # Agent核心逻辑
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP处理器
│   ├── llm/                 # 大模型集成
│   ├── memory/              # 记忆管理
│   ├── rag/                 # RAG知识库
│   │   ├── parser/          # 文档解析
│   │   ├── chunker/         # 文本分块
│   │   ├── embedding/       # 向量化
│   │   ├── retriever/       # 检索器
│   │   └── store/           # 向量存储
│   └── tools/               # 工具集
├── pkg/
│   └── models/              # 数据模型
├── config.yaml              # 配置文件
├── go.mod
├── RAG_GUIDE.md             # RAG功能测试指南
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

编辑 `config.yaml` 文件，配置你的API密钥：

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
```

### 3. 运行

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 聊天接口

**请求**

```bash
POST /api/v1/chat
Content-Type: application/json

{
  "session_id": "user-123",
  "message": "你好，请介绍一下自己",
  "model": "glm",           // 可选，glm或qwen
  "stream": false,          // 可选，是否流式输出
  "with_tools": false       // 可选，是否启用工具调用
}
```

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "session_id": "user-123",
    "message": "你好！我是一个智能助手...",
    "model": "glm",
    "tool_calls": []
  }
}
```

### 获取会话信息

```bash
GET /api/v1/session?session_id=user-123
```

### 清除会话

```bash
DELETE /api/v1/session?session_id=user-123
```

### 健康检查

```bash
GET /health
```

## 使用示例

### 基础对话

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-001",
    "message": "什么是人工智能？"
  }'
```

### 使用工具调用

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-002",
    "message": "北京今天天气怎么样？",
    "with_tools": true
  }'
```

### 切换模型

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-003",
    "message": "你好",
    "model": "qwen"
  }'
```

### RAG增强对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-rag",
    "message": "请介绍Go语言的特点",
    "model": "glm"
  }'
```

**注意**: 使用RAG功能前，需要先通过知识库管理API添加知识内容。

### 知识库管理

#### 添加知识

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Go语言是Google开发的静态类型编程语言",
    "source": "Go语言介绍"
  }'
```

#### 查看知识库统计

```bash
curl http://localhost:8080/api/v1/knowledge/stats
```

#### 搜索知识库

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Go语言",
    "top_k": 3
  }'
```

**详细说明**: 查看 [RAG_GUIDE.md](RAG_GUIDE.md)

## 工具说明

当前内置以下工具：

1. **calculator**: 数学计算
   - 用法: "计算 1+2*3"

2. **weather**: 天气查询
   - 用法: "查询北京的天气"

3. **search**: 网络搜索
   - 用法: "搜索最新的AI新闻"

## 测试

项目提供了两种测试方式：

### 方式1: 使用Python测试脚本（推荐）

```bash
# 需要先安装requests库
pip install requests

# 运行测试脚本
python3 test_api.py
```

### 方式2: 使用Bash测试脚本

```bash
# 需要先安装json_pp或jq
./test.sh
```

### 方式3: 手动测试

使用curl命令手动测试各个接口，参考上面的API接口示例。

## 配置说明

### server
- `port`: 服务端口
- `mode`: 运行模式 (debug/release/test)

### agent
- `default_model`: 默认模型 (glm/qwen)
- `max_tokens`: 最大token数
- `temperature`: 温度参数
- `enable_stream`: 是否启用流式输出

### memory
- `max_history`: 最大历史记录数
- `store_type`: 存储类型 (memory/redis)

### tools
- `enabled`: 启用的工具列表

## 开发计划

- [ ] 支持更多大模型（Claude、GPT等）
- [ ] 实现Redis持久化存储
- [ ] 添加更多工具（文件处理、代码执行等）
- [ ] 支持多模态输入输出
- [ ] 添加流式API的Server-Sent Events支持
- [ ] 实现Agent编排能力

## License

MIT
