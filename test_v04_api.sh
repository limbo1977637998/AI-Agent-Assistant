#!/bin/bash

# AI Agent Assistant v0.4 API测试脚本
# 使用方法: ./test_v04_api.sh

BASE_URL="http://localhost:8080"

echo "========================================"
echo "🚀 AI Agent Assistant v0.4 API测试"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. 健康检查
echo -e "${BLUE}1. 健康检查${NC}"
curl -s "$BASE_URL/health" | python3 -m json.tool
echo ""
echo ""

# 2. 模型列表
echo -e "${BLUE}2. 查看所有支持的模型${NC}"
curl -s "$BASE_URL/api/v1/models" | python3 -m json.tool | head -30
echo ""
echo ""

# 3. 基础对话（GLM）
echo -e "${BLUE}3. 基础对话（GLM模型）${NC}"
curl -s -X POST "$BASE_URL/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session",
    "message": "你好！请用一句话介绍你自己",
    "model": "glm"
  }' | python3 -m json.tool
echo ""

# 4. RAG知识库添加
echo -e "${BLUE}4. 添加知识到知识库${NC}"
curl -s -X POST "$BASE_URL/api/v1/knowledge/add" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "AI Agent Assistant是一个智能对话系统，支持多模型、RAG知识库、推理能力等特性。",
    "source": "system-intro"
  }' | python3 -m json.tool
echo ""

# 5. 知识库统计
echo -e "${BLUE}5. 知识库统计${NC}"
curl -s "$BASE_URL/api/v1/knowledge/stats" | python3 -m json.tool
echo ""

# 6. 推理能力测试（思维链）
echo -e "${BLUE}6. 推理能力 - 思维链推理${NC}"
curl -s -X POST "$BASE_URL/api/v1/reasoning/cot" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "小明有5个苹果，吃了2个，小红又给了他3个，现在小明有几个苹果？请详细说明计算过程"
  }' | python3 -m json.tool | head -50
echo ""

# 7. 会话状态管理
echo -e "${BLUE}7. 会话状态管理${NC}"
curl -s -X POST "$BASE_URL/api/v1/session/state" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session",
    "updates": {
      "user_name": "测试用户",
      "topic": "API测试",
      "timestamp": "2026-01-28"
    }
  }' | python3 -m json.tool
echo ""

# 8. 评估系统
echo -e "${BLUE}8. 智能评估系统${NC}"
curl -s -X POST "$BASE_URL/api/v1/eval/accuracy" \
  -H "Content-Type: application/json" \
  -d '{
    "test_cases": [
      {
        "input": "中国的首都是哪里？",
        "expected_output": "北京"
      },
      {
        "input": "3+3等于几？",
        "expected_output": "6"
      }
    ],
    "accuracy": true
  }' | python3 -m json.tool
echo ""

echo "========================================"
echo -e "${GREEN}✅ 测试完成！${NC}"
echo "========================================"
