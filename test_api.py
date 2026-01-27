#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
AI Agent Assistant API 测试脚本
"""

import requests
import json

BASE_URL = "http://localhost:8080"

def print_response(title, response):
    """打印响应结果"""
    print(f"\n{'='*50}")
    print(f"{title}")
    print(f"{'='*50}")
    print(f"状态码: {response.status_code}")
    try:
        print(f"响应内容:\n{json.dumps(response.json(), ensure_ascii=False, indent=2)}")
    except:
        print(f"响应内容:\n{response.text}")

def test_health():
    """测试健康检查"""
    response = requests.get(f"{BASE_URL}/health")
    print_response("1. 健康检查", response)

def test_basic_chat():
    """测试基础对话"""
    response = requests.post(f"{BASE_URL}/api/v1/chat", json={
        "session_id": "test-001",
        "message": "你好，请简单介绍一下你自己",
        "model": "glm"
    })
    print_response("2. 基础对话 (GLM模型)", response)

def test_weather_tool():
    """测试天气工具"""
    response = requests.post(f"{BASE_URL}/api/v1/chat", json={
        "session_id": "test-002",
        "message": "北京今天天气怎么样？",
        "with_tools": True
    })
    print_response("3. 天气查询工具调用", response)

def test_multi_turn():
    """测试多轮对话"""
    print_response("4a. 第一轮对话", requests.post(f"{BASE_URL}/api/v1/chat", json={
        "session_id": "test-003",
        "message": "我叫小明"
    }))

    print_response("4b. 第二轮对话 (记住名字)", requests.post(f"{BASE_URL}/api/v1/chat", json={
        "session_id": "test-003",
        "message": "我叫什么名字？"
    }))

def test_qwen_model():
    """测试千问模型"""
    response = requests.post(f"{BASE_URL}/api/v1/chat", json={
        "session_id": "test-004",
        "message": "你好",
        "model": "qwen"
    })
    print_response("5. 千问模型对话", response)

def test_get_session():
    """测试获取会话"""
    response = requests.get(f"{BASE_URL}/api/v1/session", params={
        "session_id": "test-003"
    })
    print_response("6. 获取会话信息", response)

def test_clear_session():
    """测试清除会话"""
    response = requests.delete(f"{BASE_URL}/api/v1/session", params={
        "session_id": "test-003"
    })
    print_response("7. 清除会话", response)

def main():
    """主函数"""
    print("\n" + "="*50)
    print("AI Agent Assistant API 测试")
    print("="*50)

    try:
        test_health()
        test_basic_chat()
        test_weather_tool()
        test_multi_turn()
        test_qwen_model()
        test_get_session()
        test_clear_session()

        print("\n" + "="*50)
        print("所有测试完成！")
        print("="*50 + "\n")
    except requests.exceptions.ConnectionError:
        print("\n错误: 无法连接到服务器，请确保服务已启动")
        print("启动命令: ./bin/ai-agent-assistant")
    except Exception as e:
        print(f"\n错误: {e}")

if __name__ == "__main__":
    main()
