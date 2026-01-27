#!/bin/bash

# AI Agent Assistant 测试脚本

BASE_URL="http://localhost:8080"

echo "========================================"
echo "AI Agent Assistant API 测试"
echo "========================================"
echo ""

# 测试健康检查
echo "1. 测试健康检查接口"
curl -s "${BASE_URL}/health" | json_pp || curl -s "${BASE_URL}/health"
echo -e "\n"

# 基础对话测试
echo "2. 测试基础对话 (GLM模型)"
curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-001",
    "message": "你好，请简单介绍一下你自己"
  }' | json_pp || echo "响应格式化失败，原始响应已显示"
echo -e "\n"

# 测试工具调用 - 天气查询
echo "3. 测试工具调用 - 天气查询"
curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-002",
    "message": "北京今天天气怎么样？",
    "with_tools": true
  }' | json_pp || echo "响应格式化失败"
echo -e "\n"

# 测试多轮对话
echo "4. 测试多轮对话"
curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-003",
    "message": "我叫小明"
  }' | json_pp || echo "响应格式化失败"
echo -e "\n"

curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-003",
    "message": "我叫什么名字？"
  }' | json_pp || echo "响应格式化失败"
echo -e "\n"

# 测试切换模型
echo "5. 测试切换到千问模型"
curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-004",
    "message": "你好",
    "model": "qwen"
  }' | json_pp || echo "响应格式化失败"
echo -e "\n"

# 测试获取会话信息
echo "6. 测试获取会话信息"
curl -s "${BASE_URL}/api/v1/session?session_id=test-003" | json_pp || echo "响应格式化失败"
echo -e "\n"

# 测试清除会话
echo "7. 测试清除会话"
curl -s -X DELETE "${BASE_URL}/api/v1/session?session_id=test-003" | json_pp || echo "响应格式化失败"
echo -e "\n"

echo "========================================"
echo "测试完成"
echo "========================================"
