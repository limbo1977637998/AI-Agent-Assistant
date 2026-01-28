#!/bin/bash

# AI Agent Assistant v0.5 API测试脚本
# 使用现有的API端点进行测试

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 AI Agent Assistant v0.5 API测试脚本"
echo "================================"
echo ""

# 测试1: 健康检查
echo "📋 测试1: 健康检查"
echo "端点: GET /health"
curl -s -X GET http://localhost:8080/health | jq '.' || echo "服务器未启动或无法连接"
echo ""
echo ""

# 测试2: 获取所有Agent（v0.5新功能）
echo "📋 测试2: 获取所有Agent"
echo "端点: GET /api/v1/agents"
curl -s -X GET $BASE_URL/agents | jq '.' || echo "端点可能未实现"
echo ""
echo ""

# 测试3: 网络搜索（v0.5新功能）
echo "📋 测试3: 网络搜索"
echo "端点: POST /api/v1/analysis/search"
curl -s -X POST $BASE_URL/analysis/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "人工智能最新发展",
    "max_results": 5
  }' | jq '.' || echo "端点可能未实现"
echo ""
echo ""

# 测试4: 数据分析（v0.5新功能）
echo "📋 测试4: 数据分析"
echo "端点: POST /api/v1/analysis/analyze"
curl -s -X POST $BASE_URL/analysis/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_type": "statistical",
    "data": [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]
  }' | jq '.' || echo "端点可能未实现"
echo ""
echo ""

# 测试5: 内容生成（v0.5新功能）
echo "📋 测试5: 内容生成"
echo "端点: POST /api/v1/analysis/write"
curl -s -X POST $BASE_URL/analysis/write \
  -H "Content-Type: application/json" \
  -d '{
    "content_type": "article",
    "topic": "AI技术发展趋势",
    "style": "formal",
    "length": 500
  }' | jq '.' || echo "端点可能未实现"
echo ""
echo ""

echo "================================"
echo "✅ 测试完成！"
echo ""
echo "📖 说明："
echo "如果看到'connection refused'错误，说明服务器未启动"
echo "如果看到'404 not found'，说明该API端点尚未实现"
echo "正常的v0.4 API端点应该可以正常工作"
echo ""
echo "🔧 启动服务器命令："
echo "  cd /Users/gongpengfei/Desktop/ClaudeCode/ai-agent-assistant"
echo "  GOPATH=/tmp/go GOCACHE=/tmp/go-cache go run cmd/server/main.go"
