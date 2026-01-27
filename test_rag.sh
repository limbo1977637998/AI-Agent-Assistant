#!/bin/bash

# AI Agent Assistant - RAG功能测试脚本

BASE_URL="http://localhost:8080"

echo "========================================"
echo "AI Agent Assistant - RAG功能测试"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试1: 健康检查
echo -e "${BLUE}测试1: 健康检查${NC}"
curl -s "${BASE_URL}/health" | python3 -m json.tool 2>/dev/null || echo "服务未运行"
echo -e "\n"

# 测试2: 添加Go语言知识
echo -e "${GREEN}测试2: 添加Go语言知识${NC}"
curl -s -X POST "${BASE_URL}/api/v1/knowledge/add" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Go语言（Golang）是Google开发的静态类型、编译型编程语言。主要特点：1) goroutine实现轻量级并发 2) 垃圾回收自动管理内存 3) 快速编译 4) 简洁语法 5) 跨平台编译。适用于云计算、微服务、网络编程。",
    "source": "Go语言简介"
  }' | python3 -m json.tool
echo -e "\n"

# 测试3: 添加Python知识
echo -e "${GREEN}测试3: 添加Python知识${NC}"
curl -s -X POST "${BASE_URL}/api/v1/knowledge/add" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Python是解释型、面向对象、动态类型的高级语言。设计哲学：代码可读性优先、优雅简洁。应用领域：数据科学、机器学习、Web开发、自动化脚本、AI开发。优势：简单易学、生态丰富、社区活跃。",
    "source": "Python简介"
  }' | python3 -m json.tool
echo -e "\n"

# 测试4: 添加AI知识
echo -e "${GREEN}测试4: 添加AI相关知识${NC}"
curl -s -X POST "${BASE_URL}/api/v1/knowledge/add" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "机器学习是AI的子领域，让计算机从数据中学习。深度学习使用多层神经网络。大语言模型（LLM）是深度学习的重要应用，如GPT、GLM等。Transformer架构是现代LLM的基础。",
    "source": "AI知识"
  }' | python3 -m json.tool
echo -e "\n"

# 测试5: 查看知识库统计
echo -e "${YELLOW}测试5: 查看知识库统计${NC}"
curl -s "${BASE_URL}/api/v1/knowledge/stats" | python3 -m json.tool
echo -e "\n"

# 测试6: 搜索知识库
echo -e "${BLUE}测试6: 搜索知识库 - Go语言${NC}"
curl -s -X POST "${BASE_URL}/api/v1/knowledge/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Go语言的并发特性",
    "top_k": 2
  }' | python3 -m json.tool
echo -e "\n"

# 测试7: RAG对话 - Go语言问题
echo -e "${GREEN}测试7: RAG对话 - 关于Go语言${NC}"
echo "问题: Go语言有哪些特点？"
curl -s -X POST "${BASE_URL}/api/v1/chat/rag" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "rag-test-001",
    "message": "请介绍一下Go语言的主要特点",
    "model": "glm"
  }' | python3 -m json.tool | head -50
echo -e "\n...\n"

# 测试8: RAG对话 - Python问题
echo -e "${GREEN}测试8: RAG对话 - 关于Python${NC}"
echo "问题: Python的设计哲学是什么？"
curl -s -X POST "${BASE_URL}/api/v1/chat/rag" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "rag-test-002",
    "message": "Python的设计哲学强调什么？",
    "model": "glm"
  }' | python3 -m json.tool | head -50
echo -e "\n...\n"

# 测试9: RAG对话 - AI知识
echo -e "${GREEN}测试9: RAG对话 - 关于AI${NC}"
echo "问题: 什么是机器学习？"
curl -s -X POST "${BASE_URL}/api/v1/chat/rag" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "rag-test-003",
    "message": "请解释一下什么是机器学习",
    "model": "glm"
  }' | python3 -m json.tool | head -50
echo -e "\n...\n"

# 测试10: 对比测试 - 不使用RAG
echo -e "${YELLOW}测试10: 对比测试 - 不使用RAG${NC}"
echo "问题: Go语言有什么特点？（无知识库）"
curl -s -X POST "${BASE_URL}/api/v1/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "normal-test-001",
    "message": "Go语言有什么特点？",
    "model": "glm"
  }' | python3 -m json.tool | head -50
echo -e "\n...\n"

echo "========================================"
echo "测试完成！"
echo "========================================"
echo ""
echo "💡 提示："
echo "1. RAG对话会基于知识库回答，更加详细和准确"
echo "2. 普通对话使用模型自身知识，可能更通用"
echo "3. 知识库重启后会清空（当前使用内存存储）"
echo ""
