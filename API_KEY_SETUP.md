# API Key 配置说明

## GLM-4 (智谱AI) ✅ 已验证可用
- API Key: `678c6ae94fad47679a52f07054c6bc8e.9Kt6eBgeVZedDYGZ`
- 状态: 正常工作
- Base URL: `https://open.bigmodel.cn/api/paas/v4`

## 千问 (阿里云) ❌ 需要修复
- API Key: `sk-a33ede10769340e88bcd55fc3b67fc28`
- 状态: 401 认证失败
- 错误: `invalid_api_key`

### 解决方案

#### 方案1: 更新API Key
1. 访问阿里云百炼平台: https://bailian.console.aliyun.com/
2. 获取正确的API Key
3. 更新 `config.yaml` 中的 `qwen.api_key`

#### 方案2: 暂时使用GLM模型
目前GLM模型工作正常，可以继续使用它：
```bash
# 在请求中指定GLM模型
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test","message":"你好","model":"glm"}'
```

#### 方案3: 禁用千问模型
如果暂时不需要千问，可以在代码中注释掉相关配置。

## 工具调用状态

✅ **天气查询**: 已修复，支持中文城市名
- 测试: 北京天气查询成功
- 支持: 北京、上海、广州、深圳等主要城市

✅ **搜索工具**: 可用
✅ **计算器工具**: 可用（简化实现）

## 推荐配置

当前推荐配置：
- 默认模型: GLM-4 (已验证可用)
- 工具: 全部启用
- 记忆: 内存存储（10条历史）
