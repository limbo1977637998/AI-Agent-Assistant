#!/bin/bash

echo "🚀 AI Agent Assistant 环境初始化"
echo "=================================="
echo ""

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

echo "✅ Docker运行正常"
echo ""

# 创建数据卷目录
echo "📁 创建数据卷目录..."
mkdir -p volumes/{etcd,minio,milvus,redis}
echo "✅ 数据卷目录创建完成"
echo ""

# 检查MySQL连接
echo "🔍 检查MySQL连接..."
if mysql -uroot -p1977637998 -e "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ MySQL连接成功"
else
    echo "❌ MySQL连接失败，请检查MySQL是否运行"
    echo "   提示: 可以使用 'brew services start mysql' 启动MySQL"
    exit 1
fi
echo ""

# 启动Docker服务
echo "🐳 启动Docker服务（Milvus + Redis）..."
docker-compose up -d

echo ""
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo ""
echo "📊 服务状态检查："
echo "=================================="

# 检查Milvus
if curl -s http://localhost:9091/healthz > /dev/null 2>&1; then
    echo "✅ Milvus: 运行正常 (REST API: http://localhost:9091)"
    echo "   gRPC端口: 19530"
else
    echo "⏳ Milvus: 正在启动中... (可能需要1-2分钟)"
    echo "   查看日志: docker-compose logs -f milvus"
fi

# 检查Redis
if docker exec agent_redis redis-cli -a redis_pass_1977637998 ping > /dev/null 2>&1; then
    echo "✅ Redis: 运行正常 (端口: 6379)"
else
    echo "⚠️  Redis: 启动中..."
    echo "   查看日志: docker-compose logs -f redis"
fi

echo ""
echo "=================================="
echo "🎉 环境初始化完成！"
echo ""
echo "📝 后续步骤："
echo "1. 创建MySQL数据库: mysql -uroot -p1977637998 -e 'CREATE DATABASE IF NOT EXISTS agent_db;'"
echo "2. 运行数据库迁移: go run cmd/migrate/main.go"
echo "3. 启动Agent服务: go run cmd/server/main.go"
echo ""
echo "🔧 常用命令："
echo "- 查看服务状态: docker-compose ps"
echo "- 查看日志: docker-compose logs -f [milvus|redis]"
echo "- 停止服务: docker-compose down"
echo "- 重启服务: docker-compose restart"
echo ""
