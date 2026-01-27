package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Agent  AgentConfig  `mapstructure:"agent"`
	Models ModelsConfig `mapstructure:"models"`
	Memory MemoryConfig `mapstructure:"memory"`
	Tools  ToolsConfig  `mapstructure:"tools"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type AgentConfig struct {
	DefaultModel string `mapstructure:"default_model"`
	MaxTokens    int    `mapstructure:"max_tokens"`
	Temperature  float64 `mapstructure:"temperature"`
	EnableStream bool   `mapstructure:"enable_stream"`
}

type ModelsConfig struct {
	GLM  ModelConfig `mapstructure:"glm"`
	Qwen ModelConfig `mapstructure:"qwen"`
}

type ModelConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

type MemoryConfig struct {
	MaxHistory int    `mapstructure:"max_history"`
	StoreType  string `mapstructure:"store_type"`
}

type ToolsConfig struct {
	Enabled []string `mapstructure:"enabled"`
}

var GlobalConfig *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

func GetModelConfig(modelName string) (ModelConfig, error) {
	switch modelName {
	case "glm":
		return GlobalConfig.Models.GLM, nil
	case "qwen":
		return GlobalConfig.Models.Qwen, nil
	default:
		return ModelConfig{}, fmt.Errorf("unknown model: %s", modelName)
	}
}
