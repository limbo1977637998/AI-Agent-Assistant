package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Proxy     ProxyConfig     `mapstructure:"proxy"`
	Agent     AgentConfig     `mapstructure:"agent"`
	Models    ModelsConfig    `mapstructure:"models"`
	Memory    MemoryConfig    `mapstructure:"memory"`
	Tools     ToolsConfig     `mapstructure:"tools"`
	Database  DatabaseConfig  `mapstructure:"database"`
	VectorDB  VectorDBConfig  `mapstructure:"vectordb"`
	Cache     CacheConfig     `mapstructure:"cache"`
	RAG       RAGConfig       `mapstructure:"rag"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type ProxyConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	HTTPProxy  string `mapstructure:"http_proxy"`
	HTTPSProxy string `mapstructure:"https_proxy"`
	NoProxy    string `mapstructure:"no_proxy"`
}

type AgentConfig struct {
	DefaultModel   string `mapstructure:"default_model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	MaxTokens      int    `mapstructure:"max_tokens"`
	Temperature    float64 `mapstructure:"temperature"`
	EnableStream   bool   `mapstructure:"enable_stream"`
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

type DatabaseConfig struct {
	Provider string                `mapstructure:"provider"`
	MySQL    MySQLDatabaseConfig    `mapstructure:"mysql"`
}

type MySQLDatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

type VectorDBConfig struct {
	Provider string     `mapstructure:"provider"`
	Milvus   MilvusConfig `mapstructure:"milvus"`
}

type MilvusConfig struct {
	Address        string `mapstructure:"address"`
	CollectionName string `mapstructure:"collection_name"`
	Dimension      int    `mapstructure:"dimension"`
	IndexType      string `mapstructure:"index_type"`
	MetricType     string `mapstructure:"metric_type"`
	EmbeddingModel string `mapstructure:"embedding_model"`
}

type CacheConfig struct {
	Enabled bool        `mapstructure:"enabled"`
	Provider string     `mapstructure:"provider"`
	Redis   RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Addr            string `mapstructure:"addr"`
	Password        string `mapstructure:"password"`
	DB              int    `mapstructure:"db"`
	PoolSize        int    `mapstructure:"pool_size"`
	ToolResultTTL   string `mapstructure:"tool_result_ttl"`
	LLMResponseTTL  string `mapstructure:"llm_response_ttl"`
	SessionTTL      string `mapstructure:"session_ttl"`
	KnowledgeCacheTTL string `mapstructure:"knowledge_cache_ttl"`
}

type RAGConfig struct {
	Enabled            bool    `mapstructure:"enabled"`
	TopK               int     `mapstructure:"top_k"`
	Threshold          float64 `mapstructure:"threshold"`
	ChunkSize          int     `mapstructure:"chunk_size"`
	ChunkOverlap       int     `mapstructure:"chunk_overlap"`
	EnableHybridSearch bool    `mapstructure:"enable_hybrid_search"`
}

type MonitoringConfig struct {
	Enabled    bool             `mapstructure:"enabled"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Tracing    TracingConfig    `mapstructure:"tracing"`
}

type PrometheusConfig struct {
	Port int    `mapstructure:"port"`
	Path string `mapstructure:"path"`
}

type TracingConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	JaegerEndpoint string  `mapstructure:"jaeger_endpoint"`
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
