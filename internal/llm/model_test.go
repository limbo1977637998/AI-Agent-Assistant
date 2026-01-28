package llm

import (
	"testing"
)

// TestModelFactory 测试模型工厂
func TestModelFactory(t *testing.T) {
	factory := NewModelFactory()

	// 测试支持的模型列表
	supportedModels := factory.GetSupportedModels()
	if len(supportedModels) == 0 {
		t.Error("No supported models found")
	}

	t.Logf("Supported models: %v", supportedModels)

	// 测试支持的提供商
	supportedProviders := factory.GetSupportedProviders()
	if len(supportedProviders) == 0 {
		t.Error("No supported providers found")
	}

	t.Logf("Supported providers: %v", supportedProviders)
}

// TestGLMModelBasic 测试GLM模型基础功能
func TestGLMModelBasic(t *testing.T) {
	// 创建配置
	config := ModelConfig{
		APIKey:  "test-key",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		Model:   "glm-4-flash",
	}

	// 创建模型
	model, err := NewGLMModel(config)
	if err != nil {
		t.Fatalf("Failed to create GLM model: %v", err)
	}

	// 测试模型接口实现
	if model.GetModelName() != "glm-4-flash" {
		t.Errorf("Expected model name 'glm-4-flash', got '%s'", model.GetModelName())
	}

	if model.GetProviderName() != "zhipu" {
		t.Errorf("Expected provider 'zhipu', got '%s'", model.GetProviderName())
	}

	// 测试方法签名
	if !model.SupportsToolCalling() {
		t.Error("GLM should support tool calling")
	}

	if model.SupportsEmbedding() {
		t.Error("GLM should not support native embedding")
	}
}

// TestQwenModelBasic 测试千问模型基础功能
func TestQwenModelBasic(t *testing.T) {
	config := ModelConfig{
		APIKey:  "test-key",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Model:   "qwen-plus",
	}

	model, err := NewQwenModel(config)
	if err != nil {
		t.Fatalf("Failed to create Qwen model: %v", err)
	}

	if model.GetModelName() != "qwen-plus" {
		t.Errorf("Expected model name 'qwen-plus', got '%s'", model.GetModelName())
	}

	if model.GetProviderName() != "qwen" {
		t.Errorf("Expected provider 'qwen', got '%s'", model.GetProviderName())
	}

	// 千问应该支持Embedding
	if !model.SupportsEmbedding() {
		t.Error("Qwen should support embedding")
	}
}

// TestOpenAIModelBasic 测试OpenAI模型基础功能
func TestOpenAIModelBasic(t *testing.T) {
	config := ModelConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-3.5-turbo",
	}

	model, err := NewOpenAIModel(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI model: %v", err)
	}

	if model.GetModelName() != "gpt-3.5-turbo" {
		t.Errorf("Expected model name 'gpt-3.5-turbo', got '%s'", model.GetModelName())
	}

	if model.GetProviderName() != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", model.GetProviderName())
	}
}

// TestModelInterfaceConsistency 测试模型接口一致性
func TestModelInterfaceConsistency(t *testing.T) {
	config := ModelConfig{
		APIKey:  "test-key",
		BaseURL: "https://test.com/v1",
		Model:   "test-model",
	}

	// 测试各个模型的接口一致性
	models := []struct {
		name  string
		model Model
	}{
		{"GLM", &GLMModel{config: config, client: nil}},
		{"Qwen", &QwenModel{config: config, client: nil}},
		{"OpenAI", &OpenAIModel{config: config, client: nil}},
		{"Claude", &ClaudeModel{config: config, client: nil}},
		{"DeepSeek", &DeepSeekModel{config: config, client: nil}},
	}

	for _, tc := range models {
		t.Run(tc.name, func(t *testing.T) {
			// 测试必需方法存在
			if tc.model.GetModelName() == "" {
				t.Error("GetModelName should not return empty string")
			}

			if tc.model.GetProviderName() == "" {
				t.Error("GetProviderName should not return empty string")
			}

			// 测试方法签名
			tc.model.SupportsToolCalling()
			tc.model.SupportsEmbedding()
		})
	}
}

// TestModelManager 测试模型管理器
func TestModelManager(t *testing.T) {
	// 这个测试需要真实的配置，作为示例展示结构
	t.Skip("Requires actual configuration")

	/*
	cfg := &config.Config{
		Models: config.ModelsConfig{
			GLM: config.ModelConfig{
				APIKey:  os.Getenv("GLM_API_KEY"),
				BaseURL: "https://open.bigmodel.cn/api/paas/v4",
				Model:   "glm-4-flash",
			},
			Qwen: config.ModelConfig{
				APIKey:  os.Getenv("QWEN_API_KEY"),
				BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				Model:   "qwen-plus",
			},
		},
	}

	manager, err := NewModelManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create model manager: %v", err)
	}

	// 测试获取模型
	glmModel, err := manager.GetModel("glm")
	if err != nil {
		t.Errorf("Failed to get GLM model: %v", err)
	}

	if glmModel == nil {
		t.Error("GLM model should not be nil")
	}

	// 测试列出模型
	models := manager.ListModels()
	if len(models) == 0 {
		t.Error("Should have at least one model loaded")
	}

	// 测试模型信息
	info := manager.GetModelInfo("glm")
	if info == nil {
		t.Error("Model info should not be nil")
	} else {
		t.Logf("GLM Model Info: %+v", info)
	}
	*/
}
