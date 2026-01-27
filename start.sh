#!/bin/bash

echo "Starting AI Agent Assistant..."

# 检查二进制文件是否存在
if [ ! -f "bin/ai-agent-assistant" ]; then
    echo "Binary not found. Building..."
    GOPATH=/Users/gongpengfei/go go build -o bin/ai-agent-assistant cmd/server/main.go
fi

# 启动服务
./bin/ai-agent-assistant
