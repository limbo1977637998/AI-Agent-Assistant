package query

import (
	"fmt"
)

// QueryOptimizerFactory 查询优化器工厂
type QueryOptimizerFactory struct {
	optimizers map[string]QueryOptimizerStrategy
}

// NewQueryOptimizerFactory 创建查询优化器工厂
func NewQueryOptimizerFactory() *QueryOptimizerFactory {
	return &QueryOptimizerFactory{
		optimizers: make(map[string]QueryOptimizerStrategy),
	}
}

// CreateOptimizer 创建查询优化器
//
// 参数:
//   optimizerType: 优化器类型 (rewrite, decompose, expand, hyde)
//   llm: LLM 提供者
//   embedding: 向量化提供者（HyDE 需要）
//   config: 配置
//
// 返回:
//   QueryOptimizerStrategy: 优化器实例
//   error: 创建错误
func (f *QueryOptimizerFactory) CreateOptimizer(
	optimizerType string,
	llm LLMProvider,
	embedding EmbeddingProvider,
	config QueryOptimizerConfig,
) (QueryOptimizerStrategy, error) {
	switch optimizerType {
	case "rewrite", "rewriting":
		return NewQueryRewriter(llm, config)

	case "decompose", "decomposition":
		return NewQueryDecomposer(llm, config)

	case "expand", "expansion":
		return NewQueryExpander(llm, config)

	case "hyde":
		if embedding == nil {
			return nil, fmt.Errorf("embedding provider is required for HyDE")
		}
		return NewHyDERetriever(llm, embedding, config)

	default:
		return nil, fmt.Errorf("unknown optimizer type: %s", optimizerType)
	}
}

// RegisterOptimizer 注册优化器
func (f *QueryOptimizerFactory) RegisterOptimizer(name string, optimizer QueryOptimizerStrategy) error {
	if name == "" {
		return fmt.Errorf("optimizer name cannot be empty")
	}

	if optimizer == nil {
		return fmt.Errorf("optimizer cannot be nil")
	}

	if err := optimizer.Validate(); err != nil {
		return fmt.Errorf("invalid optimizer: %w", err)
	}

	f.optimizers[name] = optimizer
	return nil
}

// GetOptimizer 获取已注册的优化器
func (f *QueryOptimizerFactory) GetOptimizer(name string) (QueryOptimizerStrategy, error) {
	optimizer, ok := f.optimizers[name]
	if !ok {
		return nil, fmt.Errorf("optimizer '%s' not found", name)
	}

	return optimizer, nil
}

// ListOptimizers 列出所有优化器类型
func (f *QueryOptimizerFactory) ListOptimizers() []string {
	return []string{
		"rewrite",      // 查询重写
		"decompose",    // 查询分解
		"expand",       // 查询扩展
		"hyde",         // 假设性文档嵌入
	}
}

// GetOptimizerInfo 获取优化器信息
func (f *QueryOptimizerFactory) GetOptimizerInfo(optimizerType string) map[string]interface{} {
	info := map[string]interface{}{
		"type": optimizerType,
	}

	switch optimizerType {
	case "rewrite", "rewriting":
		info["name"] = "Query Rewriter"
		info["description"] = "查询重写器，使用 LLM 改写查询使其更清晰、更具体"
		info["use_case"] = "模糊查询、不完整查询、需要澄清的查询"
		info["requires_llm"] = true
		info["requires_embedding"] = false

	case "decompose", "decomposition":
		info["name"] = "Query Decomposer"
		info["description"] = "查询分解器，将复杂查询分解为多个简单子查询"
		info["use_case"] = "复杂问题、多个信息需求、多角度检索"
		info["requires_llm"] = true
		info["requires_embedding"] = false

	case "expand", "expansion":
		info["name"] = "Query Expander"
		info["description"] = "查询扩展器，使用同义词和相关词扩展查询"
		info["use_case"] = "专业术语检索、同义词多的查询、高召回率场景"
		info["requires_llm"] = true
		info["requires_embedding"] = false

	case "hyde":
		info["name"] = "HyDE (Hypothetical Document Embeddings)"
		info["description"] = "假设性文档嵌入，生成假设答案并用其向量检索"
		info["use_case"] = "问答系统、事实查询、高召回率场景"
		info["requires_llm"] = true
		info["requires_embedding"] = true
		info["paper"] = "Precise Zero-Shot Dense Retrieval without Relevance Labels (2022)"

	default:
		info["error"] = "Unknown optimizer type"
	}

	return info
}
