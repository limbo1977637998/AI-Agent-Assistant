#!/bin/bash

# AI Agent Assistant v0.5 - 工具API测试脚本
# 用于测试所有工具相关的API端点

BASE_URL="http://localhost:8080/api/v1"

echo "=================================================="
echo "🔧 AI Agent Assistant v0.5 - 工具API测试"
echo "=================================================="

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
test_api() {
    local test_name="$1"
    local method="$2"
    local url="$3"
    local data="$4"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${BLUE}测试 ${TOTAL_TESTS}: ${test_name}${NC}"
    echo "请求: ${method} ${url}"

    if [ -n "$data" ]; then
        echo "数据: ${data}"
        response=$(curl -s -X "${method}" "${url}" \
            -H "Content-Type: application/json" \
            -d "${data}")
    else
        response=$(curl -s -X "${method}" "${url}")
    fi

    # 检查响应
    if echo "$response" | grep -q '"success":true'; then
        echo -e "${GREEN}✅ 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ 失败${NC}"
        echo "响应: ${response}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    # 显示部分响应
    if [ "$method" = "GET" ]; then
        echo "$response" | head -c 200
        echo "..."
    fi
}

# ============================================================
# 1. 工具列表测试
# ============================================================
echo -e "\n${YELLOW}===== 1. 工具管理 =====${NC}"

test_api "获取所有工具列表" \
    "GET" \
    "${BASE_URL}/tools"

test_api "获取 file_ops 工具信息" \
    "GET" \
    "${BASE_URL}/tools/file_ops"

test_api "获取 data_processor 工具能力" \
    "GET" \
    "${BASE_URL}/tools/data_processor/capabilities"

# ============================================================
# 2. 文件操作工具测试
# ============================================================
echo -e "\n${YELLOW}===== 2. 文件操作工具 =====${NC}"

# 创建测试文件
test_api "写入测试文件" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "file_ops",
      "operation": "write",
      "params": {
        "path": "/tmp/ai_agent_test.txt",
        "content": "Hello from AI Agent Assistant!\n这是测试文件内容。",
        "overwrite": true
      }
    }'

# 读取文件
test_api "读取测试文件" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "file_ops",
      "operation": "read",
      "params": {
        "path": "/tmp/ai_agent_test.txt"
      }
    }'

# 列出文件
test_api "列出目录文件" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "file_ops",
      "operation": "list",
      "params": {
        "path": "/tmp",
        "pattern": "*.txt"
      }
    }'

# ============================================================
# 3. 数据处理工具测试
# ============================================================
echo -e "\n${YELLOW}===== 3. 数据处理工具 =====${NC}"

# CSV解析
test_api "解析CSV数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "data_processor",
      "operation": "parse_csv",
      "params": {
        "content": "name,age,city\n张三,25,北京\n李四,30,上海\n王五,28,深圳",
        "has_header": true
      }
    }'

# 数据清洗
test_api "清洗数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "data_processor",
      "operation": "clean",
      "params": {
        "data": [
          {"name": "  张三  ", "age": 25, "city": "北京"},
          {"name": "", "age": 30, "city": "上海"},
          {"name": "李四", "age": 0, "city": "  上海  "}
        ],
        "operations": ["trim_whitespace", "remove_empty"]
      }
    }'

# 数据过滤
test_api "过滤数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "data_processor",
      "operation": "filter",
      "params": {
        "data": [
          {"name": "张三", "age": 25, "status": "active"},
          {"name": "李四", "age": 30, "status": "inactive"},
          {"name": "王五", "age": 28, "status": "active"}
        ],
        "conditions": [
          {"field": "status", "operator": "==", "value": "active"},
          {"field": "age", "operator": ">=", "value": 28}
        ]
      }
    }'

# 数据聚合
test_api "聚合数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "data_processor",
      "operation": "aggregate",
      "params": {
        "data": [
          {"category": "A", "amount": 100},
          {"category": "B", "amount": 200},
          {"category": "A", "amount": 150},
          {"category": "B", "amount": 250}
        ],
        "group_by": "category",
        "aggregations": [
          {"field": "amount", "operation": "sum", "alias": "total"},
          {"field": "amount", "operation": "avg", "alias": "average"}
        ]
      }
    }'

# 数据排序
test_api "排序数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "data_processor",
      "operation": "sort",
      "params": {
        "data": [
          {"name": "张三", "score": 85},
          {"name": "李四", "score": 92},
          {"name": "王五", "score": 78}
        ],
        "sort_by": "score",
        "order": "desc"
      }
    }'

# ============================================================
# 4. 批量操作工具测试
# ============================================================
echo -e "\n${YELLOW}===== 4. 批量操作工具 =====${NC}"

test_api "批量处理数据" \
    "POST" \
    "${BASE_URL}/tools/execute" \
    '{
      "tool_name": "batch_ops",
      "operation": "batch_process",
      "params": {
        "items": ["hello", "WORLD", "GoLang", "AI", "AGENT"],
        "processor": "uppercase",
        "concurrency": 3
      }
    }'

# ============================================================
# 5. 工具链测试
# ============================================================
echo -e "\n${YELLOW}===== 5. 工具链 =====${NC}"

test_api "获取工具链列表" \
    "GET" \
    "${BASE_URL}/tools/chains"

test_api "执行数据分析工具链" \
    "POST" \
    "${BASE_URL}/tools/chains/data_analysis/execute" \
    '{
      "input": null
    }'

# ============================================================
# 6. 批量工具调用测试
# ============================================================
echo -e "\n${YELLOW}===== 6. 批量工具调用 =====${NC}"

test_api "批量执行多个工具" \
    "POST" \
    "${BASE_URL}/tools/batch" \
    '{
      "calls": [
        {
          "tool_name": "data_processor",
          "operation": "parse_csv",
          "params": {
            "content": "name,age\nAlice,30\nBob,25",
            "has_header": true
          }
        },
        {
          "tool_name": "data_processor",
          "operation": "filter",
          "params": {
            "data": [{"name": "Alice", "age": 30}],
            "conditions": [{"field": "age", "operator": ">", "value": 25}]
          }
        }
      ]
    }'

# ============================================================
# 测试总结
# ============================================================
echo ""
echo "=================================================="
echo "📊 测试总结"
echo "=================================================="
echo -e "总测试数: ${TOTAL_TESTS}"
echo -e "${GREEN}通过: ${PASSED_TESTS}${NC}"
echo -e "${RED}失败: ${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✨ 所有测试通过！${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  部分测试失败${NC}"
    exit 1
fi
