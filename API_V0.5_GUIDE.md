# AI Agent Assistant v0.5 API 使用指南

## 概述

v0.5版本新增了三大专家Agent（Researcher、Analyst、Writer），提供了强大的搜索、分析和内容生成能力。本文档详细说明了如何使用这些新的API功能。

---

## 目录

1. [Agent管理API](#agent管理api)
2. [任务执行API](#任务执行api)
3. [工作流API](#工作流api)
4. [分析研究API](#分析研究api)
5. [使用示例](#使用示例)
6. [错误处理](#错误处理)

---

## Agent管理API

### 1. 获取所有Agent列表

**端点**: `GET /api/v1/agents`

**描述**: 获取系统中所有可用的Agent及其基本信息

**响应示例**:
```json
{
  "agents": [
    {
      "name": "Researcher",
      "type": "researcher",
      "capabilities": [
        "web_search",
        "information_collection",
        "data_gathering",
        "fact_checking",
        "source_analysis",
        "literature_review"
      ],
      "status": "idle"
    },
    {
      "name": "Analyst",
      "type": "analyst",
      "capabilities": [
        "data_analysis",
        "statistical_analysis",
        "trend_analysis",
        "data_visualization",
        "correlation_analysis",
        "report_generation",
        "pattern_recognition"
      ],
      "status": "idle"
    },
    {
      "name": "Writer",
      "type": "writer",
      "capabilities": [
        "content_generation",
        "article_writing",
        "report_writing",
        "copywriting",
        "content_editing",
        "summarization",
        "translation",
        "proofreading"
      ],
      "status": "idle"
    }
  ],
  "total": 3
}
```

### 2. 获取Agent详细信息

**端点**: `GET /api/v1/agents/:id`

**描述**: 获取指定Agent的详细信息

**参数**:
- `id` (路径参数): Agent ID，例如 `researcher-001`

**响应示例**:
```json
{
  "agent": {
    "id": "researcher-001",
    "name": "Researcher",
    "type": "researcher",
    "capabilities": ["web_search", "information_collection"],
    "status": "idle",
    "endpoint": "",
    "metadata": {
      "start_time": "2024-01-28T14:30:00Z"
    },
    "last_heartbeat": "2024-01-28T14:30:00Z"
  }
}
```

### 3. 获取Agent能力列表

**端点**: `GET /api/v1/agents/:id/capabilities`

**描述**: 获取Agent支持的所有能力

**参数**:
- `id` (路径参数): Agent ID

**响应示例**:
```json
{
  "agent_id": "researcher-001",
  "capabilities": [
    "web_search",
    "information_collection",
    "data_gathering",
    "fact_checking",
    "source_analysis",
    "literature_review"
  ],
  "total": 6
}
```

---

## 任务执行API

### 1. 执行单个任务

**端点**: `POST /api/v1/tasks`

**描述**: 创建并执行新任务

**请求体**:
```json
{
  "type": "researcher",
  "goal": "搜索关于AI的最新发展",
  "priority": 1,
  "requirements": {
    "keywords": ["AI", "人工智能"],
    "max_results": 10,
    "time_range": "最近一周"
  }
}
```

**参数说明**:
- `type` (必需): Agent类型，可选值：`researcher`, `analyst`, `writer`
- `goal` (必需): 任务目标描述
- `priority` (可选): 任务优先级（0-3），默认1
  - 0: Low
  - 1: Normal (默认)
  - 2: High
  - 3: Urgent
- `requirements` (可选): 任务特定要求

**响应示例**:
```json
{
  "task_id": "task-1706432400-123",
  "status": "running",
  "agent": "Researcher",
  "started_at": "2024-01-28T14:30:00Z"
}
```

### 2. 批量执行任务

**端点**: `POST /api/v1/tasks/batch`

**描述**: 批量创建并执行多个任务

**请求体**:
```json
{
  "tasks": [
    {
      "type": "researcher",
      "goal": "搜索AI技术信息",
      "priority": 1
    },
    {
      "type": "analyst",
      "goal": "分析数据统计特征",
      "priority": 2,
      "requirements": {
        "data": [10, 20, 30, 40, 50]
      }
    },
    {
      "type": "writer",
      "goal": "撰写技术报告",
      "priority": 1,
      "requirements": {
        "style": "formal",
        "length": 1000
      }
    }
  ]
}
```

**响应示例**:
```json
{
  "batch_id": "batch-1706432400-456",
  "tasks": [
    {
      "task_id": "task-1706432400-001",
      "status": "running"
    },
    {
      "task_id": "task-1706432400-002",
      "status": "running"
    },
    {
      "task_id": "task-1706432400-003",
      "status": "running"
    }
  ],
  "total": 3
}
```

### 3. 获取任务状态

**端点**: `GET /api/v1/tasks/:id`

**描述**: 获取任务的执行状态和结果

**参数**:
- `id` (路径参数): 任务ID

**响应示例**:
```json
{
  "task_id": "task-1706432400-123",
  "status": "completed",
  "result": {
    "query": "搜索关于AI的最新发展",
    "results": [...],
    "count": 10
  },
  "duration": "2.5s"
}
```

---

## 分析研究API

### 1. 网络搜索

**端点**: `POST /api/v1/analysis/search`

**描述**: 使用Researcher Agent进行网络搜索和信息收集

**请求体**:
```json
{
  "query": "人工智能最新发展",
  "max_results": 10,
  "time_range": "最近一周",
  "keywords": ["AI", "深度学习", "大模型"],
  "options": {
    "include_snippets": true
  }
}
```

**参数说明**:
- `query` (必需): 搜索查询字符串
- `max_results` (可选): 返回结果最大数量，默认10
- `time_range` (可选): 时间范围限制
- `keywords` (可选): 额外的搜索关键词
- `options` (可选): 其他搜索选项

**响应示例**:
```json
{
  "task_id": "task-1706432400-789",
  "query": "人工智能最新发展",
  "result": {
    "content_type": "search_results",
    "query": "人工智能最新发展",
    "results": [
      {
        "title": "AI技术突破：GPT-5即将发布",
        "url": "https://example.com/article1",
        "snippet": "OpenAI宣布GPT-5将在今年晚些时候发布...",
        "source": "DuckDuckGo"
      },
      {
        "title": "深度学习在医疗领域的应用",
        "url": "https://example.com/article2",
        "snippet": "最新的研究表明，深度学习在医疗诊断中...",
        "source": "DuckDuckGo"
      }
    ],
    "count": 10,
    "source": "DuckDuckGo"
  },
  "status": "completed",
  "agent": "Researcher"
}
```

**使用场景**:
- 信息收集和研究
- 事实核查
- 技术趋势分析
- 竞品调研

### 2. 数据分析

**端点**: `POST /api/v1/analysis/analyze`

**描述**: 使用Analyst Agent进行数据统计分析

**请求体**:
```json
{
  "analysis_type": "statistical",
  "data": [10, 20, 30, 40, 50, 60, 70, 80, 90, 100],
  "options": {
    "generate_charts": true,
    "include_outliers": true
  }
}
```

**参数说明**:
- `analysis_type` (必需): 分析类型
  - `statistical`: 统计分析（均值、中位数、标准差等）
  - `trend`: 趋势分析和预测
  - `comparative`: 对比分析
- `data` (必需): 待分析的数据
- `options` (可选): 分析选项

**响应示例** (统计分析):
```json
{
  "task_id": "task-1706432400-999",
  "analysis_type": "statistical",
  "result": {
    "content_type": "analysis",
    "analysis_type": "statistical",
    "statistics": {
      "count": 10,
      "mean": 55.0,
      "median": 55.0,
      "mode": 0.0,
      "min": 10.0,
      "max": 100.0,
      "variance": 825.0,
      "std_dev": 28.72,
      "range": 90.0,
      "q1": 30.0,
      "q2": 55.0,
      "q3": 80.0,
      "iqr": 50.0
    },
    "charts": {
      "histogram": {
        "bins": [...],
        "bin_width": 9.0
      },
      "box_plot": {
        "min": 10.0,
        "q1": 30.0,
        "median": 55.0,
        "q3": 80.0,
        "max": 100.0,
        "outliers": []
      }
    },
    "data_points": 10
  },
  "status": "completed",
  "agent": "Analyst"
}
```

**响应示例** (趋势分析):
```json
{
  "task_id": "task-1706432401-111",
  "analysis_type": "trend",
  "result": {
    "content_type": "analysis",
    "analysis_type": "trend",
    "trend": {
      "direction": "increasing",
      "slope": 3.5,
      "intercept": 50.0,
      "strength": 3.5
    },
    "prediction": [103.5, 107.0, 110.5],
    "data_points": 30,
    "chart_data": [...]
  },
  "status": "completed",
  "agent": "Analyst"
}
```

**使用场景**:
- 业务数据分析
- 用户行为分析
- 财务报表分析
- A/B测试结果分析
- KPI指标统计

### 3. 内容生成

**端点**: `POST /api/v1/analysis/write`

**描述**: 使用Writer Agent生成各类内容

**请求体** (撰写文章):
```json
{
  "content_type": "article",
  "topic": "人工智能技术的发展",
  "style": "formal",
  "length": 1000,
  "keywords": ["AI", "机器学习", "深度学习"],
  "options": {
    "include_outline": true,
    "add_references": false
  }
}
```

**参数说明**:
- `content_type` (必需): 内容类型
  - `article`: 文章
  - `report`: 报告
  - `summary`: 摘要
- `topic` (必需): 内容主题
- `style` (可选): 写作风格
  - `formal`: 正式（默认）
  - `casual`: 休闲
  - `professional`: 专业
  - `creative`: 创意
  - `academic`: 学术
- `length` (可选): 内容长度（字数），默认1000
- `keywords` (可选): 关键词列表
- `options` (可选): 其他选项

**响应示例**:
```json
{
  "task_id": "task-1706432401-222",
  "content_type": "article",
  "topic": "人工智能技术的发展",
  "result": {
    "content_type": "article",
    "title": "人工智能技术的发展",
    "content": "# 人工智能技术的发展\n\n## 概述\n\n本文将详细阐述人工智能技术的发展历程...\n\n## 详细内容\n\n人工智能（AI）作为计算机科学的一个重要分支...\n\n## 结论\n\n人工智能技术正在快速发展，未来将...\n",
    "outline": [
      "引言 - 介绍主题背景和重要性",
      "主体 - 详细阐述核心观点和论据",
      "分析 - 深入分析和案例说明",
      "结论 - 总结要点和展望"
    ],
    "style": "formal",
    "word_count": 850
  },
  "status": "completed",
  "agent": "Writer",
  "word_count": 850
}
```

**使用场景**:
- 技术文档撰写
- 营销文案生成
- 报告自动生成
- 内容摘要
- 多语言翻译

### 4. 生成报告

**端点**: `POST /api/v1/analysis/report`

**描述**: 协调多个Agent生成综合分析报告

**请求体**:
```json
{
  "topic": "AI技术市场分析",
  "sections": [
    "市场现状",
    "技术趋势",
    "竞争分析",
    "未来展望"
  ],
  "options": {
    "include_charts": true,
    "add_recommendations": true,
    "language": "zh-CN"
  }
}
```

**参数说明**:
- `topic` (必需): 报告主题
- `sections` (可选): 报告章节列表
- `options` (可选): 报告生成选项

**响应示例**:
```json
{
  "report_id": "report-1706432401-333",
  "topic": "AI技术市场分析",
  "status": "generating",
  "message": "Report is being generated in background"
}
```

**使用场景**:
- 行业研究报告
- 技术分析报告
- 市场调研报告
- 综合评估报告

---

## 工作流API

### 1. 创建工作流

**端点**: `POST /api/v1/workflows`

**描述**: 创建新的工作流定义

**请求体**:
```json
{
  "name": "研究报告生成流程",
  "definition": {
    "steps": [
      {
        "id": "step-1",
        "type": "researcher",
        "goal": "收集信息",
        "agent": "researcher"
      },
      {
        "id": "step-2",
        "type": "analyst",
        "goal": "分析数据",
        "agent": "analyst",
        "depends_on": ["step-1"]
      },
      {
        "id": "step-3",
        "type": "writer",
        "goal": "撰写报告",
        "agent": "writer",
        "depends_on": ["step-2"]
      }
    ]
  }
}
```

### 2. 执行工作流

**端点**: `POST /api/v1/workflows/:id/execute`

**描述**: 执行指定的工作流

**请求体**:
```json
{
  "inputs": {
    "topic": "AI技术分析",
    "max_results": 10
  }
}
```

### 3. 获取工作流列表

**端点**: `GET /api/v1/workflows`

**描述**: 获取所有已定义的工作流

### 4. 获取工作流执行历史

**端点**: `GET /api/v1/workflows/:id/executions`

**描述**: 获取工作流的执行历史记录

---

## 使用示例

### 示例1: 完整的研究分析流程

```bash
# 1. 首先搜索相关信息
curl -X POST http://localhost:8080/api/v1/analysis/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "GPT-4技术特点",
    "max_results": 15
  }'

# 2. 分析收集到的数据
curl -X POST http://localhost:8080/api/v1/analysis/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_type": "statistical",
    "data": [85, 90, 88, 92, 87, 91, 89, 93],
    "options": {
      "generate_charts": true
    }
  }'

# 3. 生成分析报告
curl -X POST http://localhost:8080/api/v1/analysis/write \
  -H "Content-Type: application/json" \
  -d '{
    "content_type": "report",
    "topic": "GPT-4技术分析报告",
    "style": "professional",
    "length": 2000
  }'
```

### 示例2: 批量任务处理

```bash
curl -X POST http://localhost:8080/api/v1/tasks/batch \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {
        "type": "researcher",
        "goal": "搜索机器学习最新论文",
        "priority": 2
      },
      {
        "type": "analyst",
        "goal": "分析实验数据",
        "priority": 2,
        "requirements": {
          "data": [23, 45, 67, 89, 12, 34, 56]
        }
      },
      {
        "type": "writer",
        "goal": "撰写实验总结",
        "priority": 1
      }
    ]
  }'
```

### 示例3: 使用Python调用API

```python
import requests
import json

BASE_URL = "http://localhost:8080/api/v1"

# 网络搜索
def perform_search(query):
    response = requests.post(
        f"{BASE_URL}/analysis/search",
        json={
            "query": query,
            "max_results": 10
        }
    )
    return response.json()

# 数据分析
def analyze_data(data):
    response = requests.post(
        f"{BASE_URL}/analysis/analyze",
        json={
            "analysis_type": "statistical",
            "data": data,
            "options": {
                "generate_charts": True
            }
        }
    )
    return response.json()

# 内容生成
def generate_content(topic, style="formal"):
    response = requests.post(
        f"{BASE_URL}/analysis/write",
        json={
            "content_type": "article",
            "topic": topic,
            "style": style,
            "length": 1000
        }
    )
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 搜索
    search_result = perform_search("人工智能最新进展")
    print(f"搜索结果: {search_result['result']['count']}条")

    # 分析
    data = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]
    analysis_result = analyze_data(data)
    print(f"均值: {analysis_result['result']['statistics']['mean']}")

    # 生成内容
    content_result = generate_content("AI技术发展趋势", "professional")
    print(f"生成字数: {content_result['word_count']}")
```

---

## 错误处理

所有API在发生错误时都会返回统一的错误格式：

```json
{
  "error": "错误类型描述",
  "details": "详细错误信息",
  "code": "ERROR_CODE"
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|-----------|------|
| `INVALID_REQUEST` | 400 | 请求参数无效或缺失 |
| `AGENT_NOT_FOUND` | 404 | 指定的Agent不存在 |
| `TASK_FAILED` | 500 | 任务执行失败 |
| `UNSUPPORTED_TYPE` | 400 | 不支持的Agent或任务类型 |

### 错误处理示例

```python
import requests

try:
    response = requests.post(
        "http://localhost:8080/api/v1/analysis/search",
        json={"query": "测试查询"},
        timeout=30
    )

    if response.status_code == 200:
        result = response.json()
        print(f"成功: {result['result']}")
    else:
        error = response.json()
        print(f"错误: {error['error']}")
        print(f"详情: {error['details']}")

except requests.exceptions.Timeout:
    print("请求超时")
except requests.exceptions.ConnectionError:
    print("连接失败")
except Exception as e:
    print(f"未知错误: {str(e)}")
```

---

## 性能考虑

### 1. 超时设置

- 网络搜索: 建议30秒超时
- 数据分析: 建议60秒超时
- 内容生成: 建议90秒超时
- 报告生成: 建议300秒超时

### 2. 并发限制

- 默认支持10个并发任务
- 批量任务建议不超过20个
- 大型任务建议使用异步模式

### 3. 结果缓存

- 搜索结果缓存时间: 1小时
- 分析结果缓存时间: 6小时
- 生成内容不缓存

---

## 最佳实践

### 1. 任务设计

- ✅ 明确定义任务目标
- ✅ 提供详细的requirements
- ✅ 合理设置任务优先级
- ❌ 避免过于宽泛的查询
- ❌ 避免超大数据集一次性分析

### 2. 错误处理

- ✅ 始终检查响应状态码
- ✅ 实现重试机制
- ✅ 记录错误日志
- ❌ 不要忽略错误信息

### 3. 性能优化

- ✅ 使用批量任务提高效率
- ✅ 合理设置超时时间
- ✅ 缓存常用结果
- ❌ 避免频繁的小任务请求

---

## 附录

### A. Agent能力对照表

| Agent | 核心能力 | 典型应用场景 |
|-------|---------|------------|
| Researcher | web_search, information_collection | 信息收集、研究、事实核查 |
| Analyst | statistical_analysis, trend_analysis | 数据分析、趋势预测、报告生成 |
| Writer | content_generation, article_writing | 文章撰写、报告生成、内容编辑 |

### B. 优先级对照表

| 优先级值 | 名称 | 说明 |
|---------|------|------|
| 0 | Low | 低优先级，后台处理 |
| 1 | Normal | 普通优先级（默认） |
| 2 | High | 高优先级，优先处理 |
| 3 | Urgent | 紧急，立即处理 |

### C. 响应状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 202 | 已接受（异步处理中） |
| 400 | 请求错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

**文档版本**: v1.0
**最后更新**: 2024-01-28
**适用版本**: AI Agent Assistant v0.5+
