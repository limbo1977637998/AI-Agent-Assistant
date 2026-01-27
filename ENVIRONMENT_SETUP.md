# AI Agent Assistant 环境初始化指南

> **版本**: v0.2.0 | **日期**: 2026-01-27
> 本指南详细说明了AI Agent Assistant项目的完整环境配置步骤。

---

## 📋 目录

1. [系统要求](#系统要求)
2. [数据库环境](#数据库环境)
3. [外部服务](#外部服务)
4. [初始化步骤](#初始化步骤)
5. [验证测试](#验证测试)
6. [故障排查](#故障排查)

---

## 系统要求

### 必需环境

- **Go**: >= 1.21
- **Node.js**: >= 18.0.0 (用于MCP工具服务)
- **MySQL**: >= 8.0
- **Docker**: >= 20.10 (用于Milvus和Redis)

### 可选环境

- **Git**: 版本控制
- **Make**: 构建工具

---

## 数据库环境

### 1. MySQL数据库

**用途**: 会话历史、用户记忆、工具调用记录、Agent运行日志

**连接信息**:
```yaml
host: localhost
port: 3306
database: agent_db
user: root
password: 1977637998
```

**数据表设计**:
- `sessions` - 会话表
- `messages` - 消息历史表
- `user_memories` - 用户记忆表
- `tool_calls` - 工具调用记录表
- `agent_runs` - Agent运行记录表
- `knowledge_base` - 知识库表
- `knowledge_chunks` - 知识分块表
- `vector_collections` - 向量集合配置表
- `system_config` - 系统配置表

**初始化命令**:
```bash
# 方式1：使用初始化脚本
mysql -uroot -p1977637998 < init-mysql.sql

# 方式2：手动执行
mysql -uroot -p1977637998 -e "CREATE DATABASE IF NOT EXISTS agent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 检查数据库
mysql -uroot -p1977637998 -e "USE agent_db; SHOW TABLES;"
```

---

### 2. Milvus向量数据库

**用途**: 知识库向量持久化、语义检索

**部署方式**: Docker容器

**启动命令**:
```bash
# 使用Docker Compose启动
docker-compose up -d etcd minio milvus

# 或单独启动
docker-compose up -d milvus
```

**连接信息**:
```yaml
address: localhost:19530
collection_name: agent_knowledge
dimension: 1024  # GLM embedding-2
index_type: HNSW
metric_type: COSINE
```

**健康检查**:
```bash
# 检查Milvus状态
curl http://localhost:9091/healthz

# 查看日志
docker-compose logs -f milvus
```

**数据卷映射**:
```yaml
./volumes/milvus:/var/lib/milvus  # 向量数据持久化
./volumes/etcd:/etcd             # etcd数据
./volumes/minio:/minio_data      # MinIO对象存储
```

---

### 3. Redis缓存

**用途**: 工具结果缓存、会话缓存、LLM响应缓存

**部署方式**: Docker容器

**启动命令**:
```bash
docker-compose up -d redis
```

**连接信息**:
```yaml
addr: localhost:6379
password: redis_pass_1977637998
db: 0
```

**缓存策略**:
```yaml
tool_result_ttl: 1h       # 工具结果缓存1小时
llm_response_ttl: 5m      # LLM响应缓存5分钟
session_ttl: 24h          # 会话缓存24小时
knowledge_cache_ttl: 30m  # 知识检索缓存30分钟
```

**健康检查**:
```bash
# 检查Redis状态
docker exec agent_redis redis-cli -a redis_pass_1977637998 ping

# 查看日志
docker-compose logs -f redis
```

**数据卷映射**:
```yaml
./volumes/redis:/data  # Redis数据持久化
```

---

## 外部服务

### LLM API (已配置)

| 服务 | API Key | 模型 | 用途 |
|------|---------|------|------|
| **智谱GLM** | `678c6ae94fad47679a52f07054c6bc8e.9Kt6eBgeVZedDYGZ` | glm-4-flash | LLM对话、Embedding |
| **千问Qwen** | `sk-1b6d69e06af7493f8018a4bf9fa394d2` | qwen-plus | LLM对话 |

### MCP工具服务 (新增18种工具)

**项目位置**: `/Users/gongpengfei/Desktop/ClaudeCode/my-mcp-server`

**启动命令**:
```bash
cd /Users/gongpengfei/Desktop/ClaudeCode/my-mcp-server
npm start
```

**可用工具列表** (共18种):

#### 🔍 搜索类 (3种)
1. **web_search** - DuckDuckGo互联网搜索
2. **web_read** - 网页内容读取
3. **github_search** - GitHub仓库搜索

#### 💻 开发类 (2种)
4. **github_repo_info** - GitHub仓库详情
5. **file_read** - 本地文件读取

#### 💰 金融类 (2种)
6. **stock_quote** - 股票实时报价
7. **stock_info** - 股票详细信息

#### 🧮 计算类 (1种)
8. **calculate** - 数学计算（支持统计学、三角函数等）

#### 🔧 工具类 (5种)
9. **json_validate** - JSON验证与美化
10. **csv_to_json** - CSV转JSON
11. **timestamp_convert** - 时间戳转换
12. **uuid_generate** - UUID生成
13. **hash_generate** - 哈希计算（MD5/SHA1/SHA256/SHA512）

#### 🌐 实用类 (3种)
14. **weather** - 天气查询
15. **text_similarity** - 文本相似度计算
16. **keyword_extract** - 关键词提取

#### 📊 数据类 (2种)
17. **csv_to_json** - CSV数据转换
18. **keyword_extract** - TF-IDF关键词提取

**所有工具均为免费，无需API Key！**

---

## 初始化步骤

### Step 1: 启动Docker服务

```bash
# 进入项目目录
cd /Users/gongpengfei/Desktop/ClaudeCode/ai-agent-assistant

# 启动所有Docker服务（Milvus + Redis）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 等待服务就绪（约1-2分钟）
sleep 30
```

**预期输出**:
```
NAME              IMAGE                      STATUS
agent_etcd        quay.io/coreos/etcd        Up
agent_minio       minio/minio:latest         Up
agent_milvus      milvusdb/milvus:latest     Up (healthy)
agent_redis       redis:7-alpine             Up (healthy)
```

---

### Step 2: 初始化MySQL数据库

```bash
# 方式1：使用初始化脚本（推荐）
mysql -uroot -p1977637998 < init-mysql.sql

# 方式2：手动创建
mysql -uroot -p1977637998 -e "CREATE DATABASE IF NOT EXISTS agent_db CHARACTER SET utf8mb4;"

# 验证数据库创建
mysql -uroot -p1977637998 -e "USE agent_db; SHOW TABLES;"
```

**预期输出**:
```
+---------------------------+
| Tables_in_agent_db        |
+---------------------------+
| sessions                  |
| messages                  |
| user_memories             |
| tool_calls                |
| agent_runs                |
| knowledge_base            |
| knowledge_chunks          |
| vector_collections        |
| system_config             |
+---------------------------+
```

---

### Step 3: 验证服务连接

```bash
# 1. 测试MySQL连接
mysql -uroot -p1977637998 -e "SELECT 1 AS test;"

# 2. 测试Milvus连接
curl http://localhost:9091/healthz

# 3. 测试Redis连接
docker exec agent_redis redis-cli -a redis_pass_1977637998 ping

# 4. 测试MCP工具服务
cd /Users/gongpengfei/Desktop/ClaudeCode/my-mcp-server
npm start
```

---

### Step 4: 编译并启动Agent服务

```bash
# 返回Agent项目目录
cd /Users/gongpengfei/Desktop/ClaudeCode/ai-agent-assistant

# 编译项目
go build -o bin/server cmd/server/main.go

# 启动服务
./bin/server

# 或直接运行
go run cmd/server/main.go
```

**预期输出**:
```
[GIN-debug] [WARNING] Creating an Engine instance...
2026/01/27 xx:xx:xx Starting AI Agent Assistant on :8080
2026/01/27 xx:xx:xx Model: glm
2026/01/27 xx:xx:xx Enabled tools: [calculator weather search]
2026/01/27 xx:xx:xx RAG enabled: true (Knowledge Base Support)
2026/01/27 xx:xx:xx Knowledge API: /api/v1/knowledge/*
2026/01/27 xx:xx:xx RAG Chat: /api/v1/chat/rag
[GIN-debug] Listening and serving HTTP on :8080
```

---

## 验证测试

### 测试1: 健康检查

```bash
curl http://localhost:8080/health
```

**预期输出**:
```json
{
  "status": "ok",
  "version": "v0.2.0"
}
```

---

### 测试2: 基础对话

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_env",
    "message": "你好",
    "model": "glm"
  }'
```

---

### 测试3: RAG功能

```bash
# 1. 添加知识
curl -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "测试知识：AI Agent Assistant是一个强大的AI助手框架",
    "source": "环境测试"
  }'

# 2. 查看知识库
curl http://localhost:8080/api/v1/knowledge/stats

# 3. RAG对话
curl -X POST http://localhost:8080/api/v1/chat/rag \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_env",
    "message": "AI Agent Assistant是什么？",
    "model": "glm"
  }'
```

---

### 测试4: MCP工具服务

```bash
# 在MCP项目目录测试工具
cd /Users/gongpengfei/Desktop/ClaudeCode/my-mcp-server

# 启动MCP服务
npm start

# 测试股票查询
# 在MCP客户端调用 stock_quote 工具
{
  "name": "stock_quote",
  "arguments": {
    "symbol": "AAPL"
  }
}

# 测试数学计算
{
  "name": "calculate",
  "arguments": {
    "expression": "sqrt(16) + sin(30)"
  }
}
```

---

## 故障排查

### 问题1: Docker服务启动失败

**症状**: `docker-compose up -d` 失败

**解决方案**:
```bash
# 1. 检查Docker是否运行
docker info

# 2. 检查端口占用
lsof -i :19530  # Milvus
lsof -i :6379   # Redis
lsof -i :9091   # Milvus REST API

# 3. 清理并重启
docker-compose down
docker system prune -f
docker-compose up -d

# 4. 查看日志
docker-compose logs milvus
docker-compose logs redis
```

---

### 问题2: MySQL连接失败

**症状**: `Error 2002 (HY000): Can't connect to local MySQL server`

**解决方案**:
```bash
# 1. 检查MySQL是否运行
brew services list | grep mysql
# 或
ps aux | grep mysql

# 2. 启动MySQL
brew services start mysql
# 或
sudo systemctl start mysql

# 3. 测试连接
mysql -uroot -p1977637998 -e "SELECT 1;"

# 4. 检查密码
mysql -uroot -p
```

---

### 问题3: Milvus健康检查失败

**症状**: `curl http://localhost:9091/healthz` 返回错误

**解决方案**:
```bash
# 1. 等待Milvus完全启动（可能需要1-2分钟）
docker-compose logs -f milvus

# 2. 检查依赖服务
docker-compose ps

# 3. 重启Milvus
docker-compose restart milvus

# 4. 清理数据并重启
docker-compose down
rm -rf volumes/milvus/*
docker-compose up -d
```

---

### 问题4: Agent服务启动失败

**症状**: 启动时报错 `panic: failed to connect database`

**解决方案**:
```bash
# 1. 检查配置文件
cat config.yaml | grep -A 10 database

# 2. 验证MySQL连接
mysql -uroot -p1977637998 agent_db -e "SHOW TABLES;"

# 3. 检查表是否存在
mysql -uroot -p1977637998 agent_db -e "DESCRIBE sessions;"

# 4. 重新初始化数据库
mysql -uroot -p1977637998 < init-mysql.sql

# 5. 查看详细错误
go run cmd/server/main.go 2>&1 | tee server.log
```

---

## 常用命令

### Docker服务管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart [milvus|redis]

# 查看日志
docker-compose logs -f [milvus|redis]

# 查看服务状态
docker-compose ps

# 清理数据（谨慎！）
docker-compose down
rm -rf volumes/*
```

---

### MySQL数据库管理

```bash
# 连接数据库
mysql -uroot -p1977637998 agent_db

# 查看所有表
SHOW TABLES;

# 查看表结构
DESCRIBE sessions;

# 查看数据
SELECT * FROM sessions LIMIT 10;

# 清空表（保留结构）
TRUNCATE TABLE messages;

# 删除表
DROP TABLE IF EXISTS test_table;

# 备份数据库
mysqldump -uroot -p1977637998 agent_db > backup.sql

# 恢复数据库
mysql -uroot -p1977637998 agent_db < backup.sql
```

---

### Redis管理

```bash
# 连接Redis
docker exec -it agent_redis redis-cli -a redis_pass_1977637998

# 查看所有键
KEYS *

# 查看键值
GET key_name

# 删除键
DEL key_name

# 清空所有数据
FLUSHALL

# 查看Redis信息
INFO
```

---

### 日志查看

```bash
# Agent服务日志
tail -f server.log

# Docker服务日志
docker-compose logs -f milvus
docker-compose logs -f redis

# 系统日志
tail -f /var/log/system.log
```

---

## 快速启动脚本

项目提供了自动化初始化脚本：

```bash
# 执行环境初始化
./setup-env.sh
```

脚本会自动：
1. ✅ 创建数据卷目录
2. ✅ 启动Docker服务（Milvus + Redis）
3. ✅ 检查MySQL连接
4. ✅ 显示服务状态

---

## 环境配置文件

### config.yaml (主配置)

```yaml
# 数据库配置
database:
  provider: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    database: "agent_db"
    user: "root"
    password: "1977637998"

# 向量数据库配置
vectordb:
  provider: "milvus"
  milvus:
    address: "localhost:19530"
    collection_name: "agent_knowledge"

# Redis缓存配置
cache:
  enabled: true
  redis:
    addr: "localhost:6379"
    password: "redis_pass_1977637998"
```

---

## 总结

完成以上步骤后，您的环境应该包括：

✅ **MySQL数据库** - 9张表，用于数据持久化
✅ **Milvus向量库** - 知识库向量存储
✅ **Redis缓存** - 性能优化
✅ **MCP工具服务** - 18种免费工具
✅ **Agent服务** - 核心应用

**下一步**: 开始使用Agent进行对话、添加知识、测试RAG功能！

---

**文档版本**: v1.0
**最后更新**: 2026-01-27
**维护者**: Limbo
