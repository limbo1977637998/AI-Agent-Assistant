#!/bin/bash

echo "=== RAG功能调试测试 ==="
echo ""

# 1. 测试embedding API
echo "1. 测试Embedding API..."
curl -s -X POST https://open.bigmodel.cn/api/paas/v4/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 678c6ae94fad47679a52f07054c6bc8e.9Kt6eBgeVZedDYGZ" \
  -d '{"model":"embedding-2","input":["测试"]}' | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'data' in data and len(data['data']) > 0:
        vec_len = len(data['data'][0]['embedding'])
        print(f'✓ Embedding正常，向量维度: {vec_len}')
    else:
        print('✗ Embedding失败')
except Exception as e:
    print(f'✗ 解析错误: {e}')
"
echo ""

# 2. 添加知识
echo "2. 添加知识到库..."
curl -s -X POST http://localhost:8080/api/v1/knowledge/add \
  -H "Content-Type: application/json" \
  -d '{
    "text": "调试信息：这里是测试内容123456789",
    "source": "调试测试"
  }' | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data['code'] == 200:
        print('✓ 接口调用成功')
    else:
        print(f'✗ 接口返回错误: {data}')
except Exception as e:
    print(f'✗ 解析错误: {e}')
"
echo ""

# 3. 检查知识库统计
echo "3. 检查知识库统计..."
curl -s http://localhost:8080/api/v1/knowledge/stats | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data['code'] == 200:
        count = data['data']['vector_count']
        print(f'知识库向量数量: {count}')
        if count > 0:
            print('✓ 向量已创建')
        else:
            print('✗ 向量未创建！知识添加失败')
    else:
        print(f'✗ 查询失败: {data}')
except Exception as e:
    print(f'✗ 解析错误: {e}')
"
echo ""

# 4. 搜索测试
echo "4. 搜索知识库..."
curl -s -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "调试",
    "top_k": 1
  }' | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data['code'] == 200:
        results = data['data']
        if len(results) > 0:
            print(f'✓ 找到 {len(results)} 条相关内容')
            print(f'内容预览: {results[0][:100]}...')
        else:
            print('✗ 未找到相关内容')
    else:
        print(f'✗ 搜索失败: {data}')
except Exception as e:
    print(f'✗ 解析错误: {e}')
"
echo ""

echo "=== 诊断总结 ==="
echo "如果embedding正常但vector_count为0，说明："
echo "  - AddText方法中Embedding调用失败"
echo "  - 或者向量存储失败"
echo ""
echo "建议：检查server.log获取详细错误信息"
