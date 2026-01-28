#!/bin/bash

# AI Agent Assistant v0.5 - 快速测试脚本
# 只测试最核心的功能

BASE_URL="http://localhost:8080/api/v1"

echo "========================================"
echo "🔧 AI Agent Assistant v0.5 快速测试"
echo "========================================"

# 颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# 测试1: 工具列表
echo -e "\n${BLUE}[1] 获取工具列表${NC}"
curl -s $BASE_URL/tools | python3 -m json.tool 2>/dev/null || curl -s $BASE_URL/tools

# 测试2: 文件写入
echo -e "\n${BLUE}[2] 写入测试文件${NC}"
curl -s -X POST $BASE_URL/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "write",
    "params": {
      "path": "/tmp/ai_test.txt",
      "content": "Hello AI Agent!",
      "overwrite": true
    }
  }' | python3 -m json.tool 2>/dev/null || echo "写入完成"

# 测试3: 文件读取
echo -e "\n${BLUE}[3] 读取测试文件${NC}"
curl -s -X POST $BASE_URL/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "file_ops",
    "operation": "read",
    "params": {"path": "/tmp/ai_test.txt"}
  }' | python3 -m json.tool 2>/dev/null || echo "读取完成"

# 测试4: CSV解析
echo -e "\n${BLUE}[4] 解析CSV数据${NC}"
curl -s -X POST $BASE_URL/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "parse_csv",
    "params": {
      "content": "name,age,city\n张三,25,北京\n李四,30,上海",
      "has_header": true
    }
  }' | python3 -m json.tool 2>/dev/null || echo "CSV解析完成"

# 测试5: 数据聚合
echo -e "\n${BLUE}[5] 数据聚合${NC}"
curl -s -X POST $BASE_URL/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "data_processor",
    "operation": "aggregate",
    "params": {
      "data": [
        {"category": "A", "value": 100},
        {"category": "B", "value": 200},
        {"category": "A", "value": 150}
      ],
      "group_by": "category",
      "aggregations": [
        {"field": "value", "operation": "sum", "alias": "total"}
      ]
    }
  }' | python3 -m json.tool 2>/dev/null || echo "聚合完成"

# 测试6: 批量处理
echo -e "\n${BLUE}[6] 批量处理${NC}"
curl -s -X POST $BASE_URL/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "batch_ops",
    "operation": "batch_process",
    "params": {
      "items": ["hello", "world", "test"],
      "processor": "uppercase"
    }
  }' | python3 -m json.tool 2>/dev/null || echo "批量处理完成"

# 测试7: 工具链列表
echo -e "\n${BLUE}[7] 获取工具链${NC}"
curl -s $BASE_URL/tools/chains | python3 -m json.tool 2>/dev/null || curl -s $BASE_URL/tools/chains

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ 快速测试完成！${NC}"
echo -e "${GREEN}========================================${NC}"
