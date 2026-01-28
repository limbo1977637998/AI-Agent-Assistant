# AI Agent Assistant v0.5 - 工具 API 完整测试命令

**服务器地址**: http://localhost:8080
**API 基础路径**: /api/v1

---

## 📋 目录

1. [工具管理 API](#1-工具管理-api)
2. [文件操作工具](#2-文件操作工具)
3. [数据处理工具](#3-数据处理工具)
4. [批量操作工具](#4-批量操作工具)
5. [工具链 API](#5-工具链-api)
6. [批量工具调用](#6-批量工具调用)

---

## 1. 工具管理 API

### 1.1 获取所有工具列表

```bash
curl -X GET http://localhost:8080/api/v1/tools
```

**预期响应**:
```json
{
  "success": true,
  "message": "获取工具列表成功",
  "data": {
    "tools": [
      {
        "name": "file_ops",
        "description": "文件操作工具 - 批量文件处理、格式转换、压缩解压",
        "version": "1.0.0"
      },
      {
        "name": "data_processor",
        "description": "数据处理工具 - CSV/JSON处理、数据清洗、统计分析",
        "version": "1.0.0"
      },
      {
        "name": "batch_ops",
        "description": "批量操作工具 - 批量HTTP请求、并发控制、批量处理",
        "version": "1.0.0"
      }
    ],
    "count": 3
  }
}
```

---

### 1.2 获取指定工具信息

```bash
curl -X GET http://localhost:8080/api/v1/tools/file_ops
```

**其他工具名称**: `data_processor`, `batch_ops`

---

### 1.3 获取工具能力描述

```bash
curl -X GET http://localhost:8080/api/v1/tools/file_ops/capabilities
```

**预期响应**:
```json
{
  "success": true,
  "message": "获取工具能力成功",
  "data": {
    "name": "file_ops",
    "description": "文件操作工具 - 批量文件处理、格式转换、压缩解压",
    "version": "1.0.0",
    "operations": [
      "read", "write", "batch_read", "convert",
      "compress", "decompress", "list", "delete"
    ]
  }
}
```

---

## 2. 文件操作工具

### 2.1 写入文件

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "write",
    "params": {
      "path": "/tmp/test_ai_agent.txt",
      "content": "Hello from AI Agent Assistant!\n这是测试文件内容。\n第二行内容。",
      "overwrite": true
    }
  }'
```

**参数说明**:
- `path`: 文件路径（必填）
- `content`: 文件内容（必填）
- `overwrite`: 是否覆盖（可选，默认 false）

---

### 2.2 读取文件

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "read",
    "params": {
      "path": "/tmp/test_ai_agent.txt"
    }
  }'
```

---

### 2.3 列出目录文件

```bash
# 列出 /tmp 目录下的所有 .txt 文件
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "list",
    "params": {
      "path": "/tmp",
      "pattern": "*.txt",
      "recursive": false
    }
  }'
```

**参数说明**:
- `path`: 目录路径（必填）
- `pattern`: 文件匹配模式（可选，默认 *）
- `recursive`: 是否递归（可选，默认 false）

---

### 2.4 批量读取文件

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "batch_read",
    "params": {
      "paths": [
        "/tmp/file1.txt",
        "/tmp/file2.txt",
        "/tmp/file3.txt"
      ]
    }
  }'
```

---

### 2.5 文件格式转换 (JSON ↔ CSV)

```bash
# JSON 转 CSV
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "convert",
    "params": {
      "path": "/tmp/data.json",
      "target_format": "csv",
      "output_path": "/tmp/data.csv"
    }
  }'
```

---

### 2.6 压缩文件

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "compress",
    "params": {
      "files": [
        "/tmp/file1.txt",
        "/tmp/file2.txt",
        "/tmp/file3.txt"
      ],
      "output": "/tmp/archive.zip"
    }
  }'
```

---

### 2.7 解压文件

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "decompress",
    "params": {
      "source": "/tmp/archive.zip",
      "destination": "/tmp/extracted"
    }
  }'
```

---

## 3. 数据处理工具

### 3.1 解析 CSV 数据

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "parse_csv",
    "params": {
      "content": "name,age,city\n张三,25,北京\n李四,30,上海\n王五,28,深圳",
      "has_header": true,
      "delimiter": ","
    }
  }'
```

**预期响应**:
```json
{
  "success": true,
  "message": "CSV解析成功",
  "data": {
    "headers": ["name", "age", "city"],
    "data": [
      {"name": "张三", "age": "25", "city": "北京"},
      {"name": "李四", "age": "30", "city": "上海"},
      {"name": "王五", "age": "28", "city": "深圳"}
    ]
  },
  "metadata": {
    "row_count": 3,
    "column_count": 3,
    "has_header": true
  }
}
```

---

### 3.2 解析 JSON 数据

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "parse_json",
    "params": {
      "content": "[{\"name\":\"张三\",\"age\":25},{\"name\":\"李四\",\"age\":30}]"
    }
  }'
```

---

### 3.3 数据清洗

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "clean",
    "params": {
      "data": [
        {"name": "  张三  ", "age": 25, "city": "北京"},
        {"name": "", "age": 30, "city": "上海"},
        {"name": "李四", "age": 0, "city": "  上海  "},
        {"name": "王五", "age": 28, "city": "深圳"}
      ],
      "operations": ["remove_empty", "trim_whitespace", "normalize_case"]
    }
  }'
```

**清洗操作说明**:
- `remove_empty`: 移除空行
- `trim_whitespace`: 去除首尾空格
- `normalize_case`: 标准化大小写（转为小写）
- `remove_duplicates`: 去重

---

### 3.4 数据过滤

```bash
# 多条件过滤：status=active 且 age>=25
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "filter",
    "params": {
      "data": [
        {"name": "张三", "age": 25, "status": "active"},
        {"name": "李四", "age": 30, "status": "inactive"},
        {"name": "王五", "age": 28, "status": "active"},
        {"name": "赵六", "age": 22, "status": "active"}
      ],
      "conditions": [
        {
          "field": "status",
          "operator": "==",
          "value": "active"
        },
        {
          "field": "age",
          "operator": ">=",
          "value": 25
        }
      ]
    }
  }'
```

**支持的操作符**:
- `==`, `!=`, `>`, `>=`, `<`, `<=`
- `contains`, `starts_with`, `ends_with`

---

### 3.5 数据聚合

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "aggregate",
    "params": {
      "data": [
        {"category": "A", "amount": 100},
        {"category": "B", "amount": 200},
        {"category": "A", "amount": 150},
        {"category": "B", "amount": 250},
        {"category": "A", "amount": 120}
      ],
      "group_by": "category",
      "aggregations": [
        {
          "field": "amount",
          "operation": "sum",
          "alias": "total_amount"
        },
        {
          "field": "amount",
          "operation": "avg",
          "alias": "average_amount"
        },
        {
          "field": "amount",
          "operation": "count",
          "alias": "count"
        }
      ]
    }
  }'
```

**聚合操作**:
- `count`: 计数
- `sum`: 求和
- `avg`: 平均值
- `min`: 最小值
- `max`: 最大值
- `first`: 第一个值
- `last`: 最后一个值

---

### 3.6 数据排序

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "sort",
    "params": {
      "data": [
        {"name": "张三", "score": 85},
        {"name": "李四", "score": 92},
        {"name": "王五", "score": 78},
        {"name": "赵六", "score": 88}
      ],
      "sort_by": "score",
      "order": "desc"
    }
  }'
```

**排序方向**: `asc` (升序), `desc` (降序)

---

### 3.7 数据去重

```bash
# 按指定字段去重
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "deduplicate",
    "params": {
      "data": [
        {"name": "张三", "city": "北京"},
        {"name": "李四", "city": "上海"},
        {"name": "张三", "city": "北京"},
        {"name": "王五", "city": "深圳"}
      ],
      "deduplicate_by": "name"
    }
  }'
```

---

### 3.8 缺失值填充

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "fill_missing",
    "params": {
      "data": [
        {"name": "张三", "age": 25, "score": null},
        {"name": "李四", "age": null, "score": 85},
        {"name": "王五", "age": 28, "score": 90}
      ],
      "fill_rules": [
        {
          "field": "age",
          "strategy": "mean"
        },
        {
          "field": "score",
          "strategy": "value",
          "value": 0
        }
      ]
    }
  }'
```

**填充策略**:
- `mean`: 平均值
- `median`: 中位数
- `mode`: 众数
- `forward_fill`: 前向填充
- `backward_fill`: 后向填充
- `value`: 固定值

---

### 3.9 数据转换

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "transform",
    "params": {
      "data": [
        {"name": "hello"},
        {"name": "world"}
      ],
      "transformations": [
        {
          "field": "name",
          "operation": "uppercase"
        }
      ]
    }
  }'
```

**转换操作**:
- `uppercase`: 转大写
- `lowercase`: 转小写
- `add`: 加法
- `subtract`: 减法
- `multiply`: 乘法
- `divide`: 除法
- `replace`: 替换
- `regex_replace`: 正则替换
- `round`: 四舍五入

---

### 3.10 数据合并

```bash
# Inner Join
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "merge",
    "params": {
      "data1": [
        {"id": 1, "name": "张三", "age": 25},
        {"id": 2, "name": "李四", "age": 30}
      ],
      "data2": [
        {"id": 1, "city": "北京"},
        {"id": 2, "city": "上海"},
        {"id": 3, "city": "深圳"}
      ],
      "join_type": "inner",
      "on": "id"
    }
  }'
```

**连接类型**:
- `inner`: 内连接（只保留两边都有的）
- `left`: 左连接（保留左边所有）
- `right`: 右连接（保留右边所有）
- `full`: 全连接（保留所有）

---

## 4. 批量操作工具

### 4.1 批量 HTTP 请求

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "batch_ops",
    "operation": "batch_http",
    "params": {
      "requests": [
        {
          "url": "https://httpbin.org/get",
          "method": "GET"
        },
        {
          "url": "https://httpbin.org/post",
          "method": "POST",
          "body": "{\"test\": \"data\"}"
        }
      ],
      "concurrency": 2,
      "timeout": 10
    }
  }'
```

---

### 4.2 批量数据处理

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "batch_ops",
    "operation": "batch_process",
    "params": {
      "items": ["hello", "world", "golang", "ai", "agent"],
      "processor": "uppercase",
      "concurrency": 3
    }
  }'
```

**内置处理器**:
- `uppercase`: 转大写
- `lowercase`: 转小写
- `reverse`: 反转字符串
- `double`: 数值翻倍
- `square`: 数值平方

---

### 4.3 并行执行任务

```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "batch_ops",
    "operation": "parallel_execute",
    "params": {
      "tasks": [
        {
          "name": "task1",
          "operation": "uppercase",
          "params": {"input": "hello"}
        },
        {
          "name": "task2",
          "operation": "reverse",
          "params": {"input": "world"}
        }
      ],
      "stop_on_error": false
    }
  }'
```

---

## 5. 工具链 API

### 5.1 获取所有工具链

```bash
curl -X GET http://localhost:8080/api/v1/tools/chains
```

**预期响应**:
```json
{
  "success": true,
  "message": "获取工具链列表成功",
  "data": {
    "chains": [
      {"name": "data_processing", "steps": 4},
      {"name": "batch_download_process", "steps": 3},
      {"name": "data_analysis", "steps": 4}
    ],
    "count": 3
  }
}
```

---

### 5.2 执行数据处理工具链

```bash
# 读取CSV -> 清洗数据 -> 转换格式 -> 保存
curl -X POST http://localhost:8080/api/v1/tools/chains/data_processing/execute \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

### 5.3 执行数据分析工具链

```bash
# 读取数据 -> 过滤 -> 聚合 -> 生成报告
curl -X POST http://localhost:8080/api/v1/tools/chains/data_analysis/execute \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## 6. 批量工具调用

### 6.1 批量执行多个工具

```bash
curl -X POST http://localhost:8080/api/v1/tools/batch \
  -H "Content-Type: application/json" \
  -d '{
    "calls": [
      {
        "tool_name": "file_ops",
        "operation": "write",
        "params": {
          "path": "/tmp/file1.txt",
          "content": "Content 1",
          "overwrite": true
        }
      },
      {
        "tool_name": "file_ops",
        "operation": "write",
        "params": {
          "path": "/tmp/file2.txt",
          "content": "Content 2",
          "overwrite": true
        }
      },
      {
        "tool_name": "data_processor",
        "operation": "parse_csv",
        "params": {
          "content": "name,age\nAlice,30\nBob,25",
          "has_header": true
        }
      }
    ]
  }'
```

---

## 📝 快速测试序列

### 场景1: 完整的数据处理流程

```bash
# 1. 写入测试数据文件
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "write",
    "params": {
      "path": "/tmp/sales_data.csv",
      "content": "product,category,amount\niPhone,Electronics,999\nMacBook,Electronics,1999\nHeadphones,Electronics,199\nT-Shirt,Clothing,29\nJeans,Clothing,79",
      "overwrite": true
    }
  }'

# 2. 读取并解析CSV
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "read",
    "params": {"path": "/tmp/sales_data.csv"}
  }'

# 3. 按类别聚合数据
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "aggregate",
    "params": {
      "data": [
        {"product": "iPhone", "category": "Electronics", "amount": 999},
        {"product": "MacBook", "category": "Electronics", "amount": 1999},
        {"product": "Headphones", "category": "Electronics", "amount": 199},
        {"product": "T-Shirt", "category": "Clothing", "amount": 29},
        {"product": "Jeans", "category": "Clothing", "amount": 79}
      ],
      "group_by": "category",
      "aggregations": [
        {"field": "amount", "operation": "sum", "alias": "total_sales"},
        {"field": "amount", "operation": "avg", "alias": "avg_price"},
        {"field": "product", "operation": "count", "alias": "product_count"}
      ]
    }
  }'
```

### 场景2: 数据清洗和分析

```bash
# 1. 清洗脏数据
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "clean",
    "params": {
      "data": [
        {"name": "  Alice  ", "age": 25, "score": 85},
        {"name": "", "age": 30, "score": 90},
        {"name": "Bob", "age": 0, "score": 75},
        {"name": "  Charlie  ", "age": 28, "score": 95}
      ],
      "operations": ["remove_empty", "trim_whitespace"]
    }
  }'

# 2. 过滤高分学生
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "filter",
    "params": {
      "data": [
        {"name": "Alice", "age": 25, "score": 85},
        {"name": "Bob", "age": 30, "score": 90},
        {"name": "Charlie", "age": 28, "score": 95}
      ],
      "conditions": [
        {"field": "score", "operator": ">=", "value": 90}
      ]
    }
  }'

# 3. 按分数排序
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "sort",
    "params": {
      "data": [
        {"name": "Alice", "score": 85},
        {"name": "Bob", "score": 90},
        {"name": "Charlie", "score": 95}
      ],
      "sort_by": "score",
      "order": "desc"
    }
  }'
```

---

## 🔍 故障排查

### 检查服务器状态

```bash
# 健康检查
curl http://localhost:8080/health

# 查看Agent列表
curl http://localhost:8080/api/v1/agents

# 查看工具列表
curl http://localhost:8080/api/v1/tools
```

### 常见错误

1. **404 Not Found**: 路由未正确注册，检查服务器启动日志
2. **500 Internal Server Error**: 查看服务器日志 `/tmp/server_v05.log`
3. **Invalid request body**: 检查 JSON 格式是否正确

---

## 📊 性能测试

### 批量操作性能测试

```bash
# 测试并发处理能力
time curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "batch_ops",
    "operation": "batch_process",
    "params": {
      "items": ["item1","item2","item3","item4","item5","item6","item7","item8","item9","item10"],
      "processor": "uppercase",
      "concurrency": 5
    }
  }'
```

---

**文档版本**: v1.0
**更新时间**: 2026-01-28
**适用版本**: AI Agent Assistant v0.5
