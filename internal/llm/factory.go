package llm

import (
	"fmt"

	"ai-agent-assistant/internal/config"
)

// ModelFactory 模型工厂
type ModelFactory struct{}

// NewModelFactory 创建模型工厂
func NewModelFactory() *ModelFactory {
	return &ModelFactory{}
}

// CreateModel 创建模型（根据模型名称）
func (f *ModelFactory) CreateModel(modelName string, cfg *config.Config) (Model, error) {
	switch modelName {
	case "glm", "glm-4-flash", "glm-4-plus", "glm-4-alltools":
		modelCfg := ModelConfig{
			APIKey:  cfg.Models.GLM.APIKey,
			BaseURL: cfg.Models.GLM.BaseURL,
			Model:   cfg.Models.GLM.Model,
		}
		return NewGLMModel(modelCfg)

	case "qwen", "qwen-plus", "qwen-max", "qwen-turbo", "qwen-long":
		modelCfg := ModelConfig{
			APIKey:  cfg.Models.Qwen.APIKey,
			BaseURL: cfg.Models.Qwen.BaseURL,
			Model:   cfg.Models.Qwen.Model,
		}
		return NewQwenModel(modelCfg)

	case "openai", "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", "gpt-4o":
		// 从环境变量或配置中获取OpenAI API Key
		return NewOpenAIModel(ModelConfig{
			Model: modelName,
		})

	case "claude", "claude-3-5-sonnet", "claude-3-opus", "claude-3-haiku":
		// 从环境变量或配置中获取Claude API Key
		return NewClaudeModel(ModelConfig{
			Model: modelName,
		})

	case "deepseek", "deepseek-chat", "deepseek-coder", "deepseek-r1":
		// 从环境变量或配置中获取DeepSeek API Key
		return NewDeepSeekModel(ModelConfig{
			Model: modelName,
		})

	default:
		return nil, fmt.Errorf("unsupported model: %s", modelName)
	}
}

// CreateModelWithConfig 使用自定义配置创建模型
func (f *ModelFactory) CreateModelWithConfig(provider string, config ModelConfig) (Model, error) {
	switch provider {
	case "glm":
		return NewGLMModel(config)
	case "qwen":
		return NewQwenModel(config)
	case "openai":
		return NewOpenAIModel(config)
	case "claude":
		return NewClaudeModel(config)
	case "deepseek":
		return NewDeepSeekModel(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GetSupportedModels 获取支持的模型列表
func (f *ModelFactory) GetSupportedModels() []string {
	return []string{
		// GLM系列
		"glm-4-flash",
		"glm-4-plus",
		"glm-4-alltools",
		// 千问系列
		"qwen-turbo",
		"qwen-plus",
		"qwen-max",
		"qwen-long",
		// OpenAI系列
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		// Claude系列
		"claude-3-5-sonnet",
		"claude-3-opus",
		"claude-3-haiku",
		// DeepSeek系列
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-r1",
	}
}

// GetSupportedProviders 获取支持的提供商列表
func (f *ModelFactory) GetSupportedProviders() []string {
	return []string{
		"glm",      // 智谱GLM
		"qwen",     // 阿里云千问
		"openai",   // OpenAI
		"claude",   // Anthropic
		"deepseek", // DeepSeek
	}
}

// ModelManager 模型管理器（新版，使用Model接口）
type ModelManager struct {
	factory *ModelFactory
	models  map[string]Model
	config  *config.Config
}

// NewModelManager 创建模型管理器
func NewModelManager(cfg *config.Config) (*ModelManager, error) {
	factory := NewModelFactory()
	manager := &ModelManager{
		factory: factory,
		models:  make(map[string]Model),
		config:  cfg,
	}

	// 初始化默认模型
	if err := manager.initDefaultModels(); err != nil {
		return nil, err
	}

	return manager, nil
}

// initDefaultModels 初始化默认模型
func (m *ModelManager) initDefaultModels() error {
	// 初始化GLM
	if m.config.Models.GLM.APIKey != "" {
		glmModel, err := m.factory.CreateModel("glm", m.config)
		if err != nil {
			return err
		}
		m.models["glm"] = glmModel
	}

	// 初始化千问
	if m.config.Models.Qwen.APIKey != "" {
		qwenModel, err := m.factory.CreateModel("qwen", m.config)
		if err != nil {
			return err
		}
		m.models["qwen"] = qwenModel
	}

	return nil
}

// GetModel 获取模型
func (m *ModelManager) GetModel(modelName string) (Model, error) {
	// 如果已经初始化，直接返回
	if model, ok := m.models[modelName]; ok {
		return model, nil
	}

	// 尝试动态创建
	model, err := m.factory.CreateModel(modelName, m.config)
	if err != nil {
		return nil, err
	}

	// 缓存模型
	m.models[modelName] = model
	return model, nil
}

// RegisterModel 注册自定义模型
func (m *ModelManager) RegisterModel(name string, model Model) {
	m.models[name] = model
}

// ListModels 列出所有已加载的模型
func (m *ModelManager) ListModels() []string {
	models := make([]string, 0, len(m.models))
	for name := range m.models {
		models = append(models, name)
	}
	return models
}

// GetModelInfo 获取模型信息
func (m *ModelManager) GetModelInfo(modelName string) map[string]interface{} {
	model, err := m.GetModel(modelName)
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"name":              model.GetModelName(),
		"provider":          model.GetProviderName(),
		"supports_tools":    model.SupportsToolCalling(),
		"supports_embedding": model.SupportsEmbedding(),
	}
}
