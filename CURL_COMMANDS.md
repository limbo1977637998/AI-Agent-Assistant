# AI Agent Assistant v0.5 - 工具 API 快速命令表

## 📋 工具管理

```bash
# 1. 查看所有工具
curl http://localhost:8080/api/v1/tools

# 2. 查看工具详情
curl http://localhost:8080/api/v1/tools/file_ops

# 3. 查看工具能力
curl http://localhost:8080/api/v1/tools/file_ops/capabilities
```

---

## 📁 文件操作

```bash
# 写入文件
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"file_ops","operation":"write","params":{"path":"/tmp/test.txt","content":"Hello World","overwrite":true}}'

# 读取文件
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"file_ops","operation":"read","params":{"path":"/tmp/test.txt"}}'

# 列出文件
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"file_ops","operation":"list","params":{"path":"/tmp","pattern":"*.txt"}}'

# JSON转CSV
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"file_ops","operation":"convert","params":{"path":"/tmp/data.json","target_format":"csv"}}'
```

---

## 📊 数据处理

```bash
# 解析CSV
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"parse_csv","params":{"content":"name,age\nAlice,30\nBob,25","has_header":true}}'

# 数据清洗
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"clean","params":{"data":[{"name":"  Alice  ","age":30},{"name":"","age":25}],"operations":["trim_whitespace","remove_empty"]}}'

# 数据过滤
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"filter","params":{"data":[{"name":"Alice","age":30},{"name":"Bob","age":25}],"conditions":[{"field":"age","operator":">=","value":28}]}}'

# 数据聚合
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"aggregate","params":{"data":[{"cat":"A","val":100},{"cat":"B","val":200}],"group_by":"cat","aggregations":[{"field":"val","operation":"sum"}]}}'

# 数据排序
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"sort","params":{"data":[{"name":"A","score":85},{"name":"B","score":92}],"sort_by":"score","order":"desc"}}'

# 数据去重
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"deduplicate","params":{"data":[{"name":"A"},{"name":"B"},{"name":"A"}],"deduplicate_by":"name"}}'

# 缺失值填充
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"fill_missing","params":{"data":[{"name":"A","age":30},{"name":"B","age":null}],"fill_rules":[{"field":"age","strategy":"mean"}]}}'

# 数据转换
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"transform","params":{"data":[{"name":"hello"}],"transformations":[{"field":"name","operation":"uppercase"}]}}'

# 数据合并
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"data_processor","operation":"merge","params":{"data1":[{"id":1,"name":"A"}],"data2":[{"id":1,"city":"北京"}],"join_type":"inner","on":"id"}}'
```

---

## ⚡ 批量操作

```bash
# 批量HTTP请求
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"batch_ops","operation":"batch_http","params":{"requests":[{"url":"https://httpbin.org/get","method":"GET"}],"concurrency":2}}'

# 批量数据处理
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"batch_ops","operation":"batch_process","params":{"items":["a","b","c"],"processor":"uppercase","concurrency":3}}'

# 并行执行任务
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"tool_name":"batch_ops","operation":"parallel_execute","params":{"tasks":[{"name":"task1","operation":"uppercase","params":{"input":"hello"}}]}}'
```

---

## 🔗 工具链

```bash
# 查看工具链列表
curl http://localhost:8080/api/v1/tools/chains

# 执行数据处理链
curl -X POST http://localhost:8080/api/v1/tools/chains/data_processing/execute \
  -H "Content-Type: application/json" \
  -d '{}'

# 执行数据分析链
curl -X POST http://localhost:8080/api/v1/tools/chains/data_analysis/execute \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## 🔄 批量工具调用

```bash
# 批量执行多个工具
curl -X POST http://localhost:8080/api/v1/tools/batch \
  -H "Content-Type: application/json" \
  -d '{
    "calls": [
      {"tool_name":"file_ops","operation":"write","params":{"path":"/tmp/f1.txt","content":"data1","overwrite":true}},
      {"tool_name":"file_ops","operation":"write","params":{"path":"/tmp/f2.txt","content":"data2","overwrite":true}}
    ]
  }'
```

---

## 🎯 推荐测试顺序

### 1️⃣ 基础测试
```bash
# 健康检查
curl http://localhost:8080/health

# 工具列表
curl http://localhost:8080/api/v1/tools
```

### 2️⃣ 文件操作
```bash
# 写入 → 读取 → 列表
curl -X POST http://localhost:8080/api/v1/tools/execute -H "Content-Type: application/json" -d '{"tool_name":"file_ops","operation":"write","params":{"path":"/tmp/test.txt","content":"Hello!","overwrite":true}}'

curl -X POST http://localhost:8080/api/v1/tools/execute -H "Content-Type: application/json" -d '{"tool_name":"file_ops","operation":"read","params":{"path":"/tmp/test.txt"}}'
```

### 3️⃣ 数据处理
```bash
# CSV解析 → 过滤 → 聚合 → 排序
curl -X POST http://localhost:8080/api/v1/tools/execute -H "Content-Type: application/json" -d '{"tool_name":"data_processor","operation":"parse_csv","params":{"content":"name,age\nAlice,30\nBob,25","has_header":true}}'
```

### 4️⃣ 批量操作
```bash
# 批量处理
curl -X POST http://localhost:8080/api/v1/tools/execute -H "Content-Type: application/json" -d '{"tool_name":"batch_ops","operation":"batch_process","params":{"items":["a","b","c"],"processor":"uppercase"}}'
```

---

## 📝 格式化输出技巧

### 使用 Python 格式化 JSON
```bash
curl ... | python3 -m json.tool
```

### 使用 jq 格式化 JSON（推荐）
```bash
curl ... | jq .
```

### 只显示成功状态
```bash
curl -s ... | grep -o '"success":[^,]*'
```

---

## ⚠️ 常见问题

### 404 错误
- 检查服务器是否启动：`curl http://localhost:8080/health`
- 检查端口是否正确

### JSON 格式错误
- 使用 JSON 验证工具检查格式
- 注意转义引号：`\"`

### 权限错误
- 检查文件路径是否有读写权限
- 使用 `/tmp` 目录测试

---

**更新时间**: 2026-01-28
**适用版本**: v0.5
