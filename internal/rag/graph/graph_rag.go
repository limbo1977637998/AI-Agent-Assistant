package graph

import (
	"context"
	"fmt"
	"strings"
)

// GraphRAG Graph RAG 检索器
//
// 策略说明:
//   使用知识图谱增强检索，支持全局和局部两种检索模式
//
// 全局检索 (Global Search):
//   - 跨社区检索
//   - 适合回答全局性问题
//   - 使用社区摘要
//
// 局部检索 (Local Search):
//   - 社区内检索
//   - 适合回答具体问题
//   - 使用实体及其关系
//
// 论文:
//   "From Local to Global: A Graph RAG Approach to Query-Focused
//    Summarization of Repositories" (Microsoft Research, 2024)
type GraphRAG struct {
	extractor   *EntityExtractor
	detector    CommunityDetector
	communitySummaries map[string]string // 社区摘要缓存
	llm         LLMProvider
	config      GraphRAGConfig
}

// GraphRAGConfig Graph RAG 配置
type GraphRAGConfig struct {
	// CommunityLevels 社区层次数量
	CommunityLevels int

	// MinCommunitySize 最小社区大小
	MinCommunitySize int

	// MaxCommunities 最大社区数量
	MaxCommunities int

	// UseSummary 是否使用摘要
	UseSummary bool
}

// DefaultGraphRAGConfig 返回默认配置
func DefaultGraphRAGConfig() GraphRAGConfig {
	return GraphRAGConfig{
		CommunityLevels:    3,
		MinCommunitySize:   3,
		MaxCommunities:     20,
		UseSummary:         true,
	}
}

// NewGraphRAG 创建 Graph RAG 检索器
func NewGraphRAG(llm LLMProvider, config GraphRAGConfig) (*GraphRAG, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	// 设置默认值
	if config.CommunityLevels <= 0 {
		config.CommunityLevels = 3
	}
	if config.MinCommunitySize <= 0 {
		config.MinCommunitySize = 3
	}
	if config.MaxCommunities <= 0 {
		config.MaxCommunities = 20
	}

	extractorConfig := DefaultExtractorConfig()
	extractor, err := NewEntityExtractor(llm, extractorConfig)
	if err != nil {
		return nil, err
	}

	detector := NewLouvainDetector(1.0)

	return &GraphRAG{
		extractor:         extractor,
		detector:          detector,
		communitySummaries: make(map[string]string),
		llm:               llm,
		config:            config,
	}, nil
}

// BuildGraph 构建知识图谱
func (gr *GraphRAG) BuildGraph(ctx context.Context, documents []string) (*KnowledgeGraph, error) {
	graph := &KnowledgeGraph{
		Entities:  make([]*Entity, 0),
		Relations: make([]*Relation, 0),
	}

	// 1. 从每个文档提取实体和关系
	entityMap := make(map[string]*Entity)
	for _, doc := range documents {
		graphData, err := gr.extractor.Extract(ctx, doc)
		if err != nil {
			continue // 跳过失败的文档
		}

		// 合并实体
		for _, entity := range graphData.Entities {
			if existing, ok := entityMap[entity.Name]; ok {
				// 实体已存在，更新描述
				if entity.Description != "" && existing.Description == "" {
					existing.Description = entity.Description
				}
			} else {
				// 新实体
				entity.ID = fmt.Sprintf("entity_%d", len(entityMap))
				entityMap[entity.Name] = entity
				graph.Entities = append(graph.Entities, entity)
			}
		}

		// 添加关系
		for _, relation := range graphData.Relations {
			// 验证实体是否存在
			if _, ok := entityMap[relation.From]; !ok {
				continue
			}
			if _, ok := entityMap[relation.To]; !ok {
				continue
			}

			// 检查是否已存在相同关系
			exists := false
			for _, r := range graph.Relations {
				if r.From == relation.From && r.To == relation.To {
					exists = true
					break
				}
			}

			if !exists {
				relation.ID = fmt.Sprintf("rel_%d", len(graph.Relations))
				graph.Relations = append(graph.Relations, relation)
			}
		}
	}

	// 2. 检测社区
	communities, err := gr.detector.DetectCommunities(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to detect communities: %w", err)
	}

	// 3. 生成社区摘要
	if gr.config.UseSummary {
		for _, comm := range communities {
			summary, err := gr.generateCommunitySummary(ctx, graph, comm)
			if err != nil {
				continue
			}
			comm.Summary = summary
			gr.communitySummaries[comm.ID] = summary
		}
	}

	return graph, nil
}

// GlobalSearch 全局检索
// 使用社区摘要进行跨社区检索
func (gr *GraphRAG) GlobalSearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 获取所有社区摘要
	summaries := gr.getAllCommunitySummaries(graph)

	if len(summaries) == 0 {
		// 如果没有摘要，回退到局部检索
		return gr.LocalSearch(ctx, graph, query, topK)
	}

	// 2. 使用查询匹配最相关的社区
	relevantSummaries := gr.rankSummaries(ctx, query, summaries, topK)

	// 3. 构建上下文
	contexts := make([]string, 0, len(relevantSummaries))
	for _, summary := range relevantSummaries {
		contexts = append(contexts, summary.Summary)
	}

	return contexts, nil
}

// LocalSearch 局部检索
// 在社区内进行检索，使用实体及其关系
func (gr *GraphRAG) LocalSearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 提取查询中的实体
	queryEntities := gr.extractQueryEntities(ctx, query)

	// 2. 找到相关的实体
	relevantEntities := gr.findRelevantEntities(graph, queryEntities, topK)

	// 3. 获取这些实体及其关系
	contexts := gr.buildEntityContexts(graph, relevantEntities)

	return contexts, nil
}

// CommunitySearch 社区检索
// 结合全局和局部检索
func (gr *GraphRAG) CommunitySearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 先尝试全局检索
	globalContexts, err := gr.GlobalSearch(ctx, graph, query, topK/2)
	if err != nil {
		return nil, err
	}

	// 2. 再进行局部检索
	localContexts, err := gr.LocalSearch(ctx, graph, query, topK/2)
	if err != nil {
		return globalContexts, nil
	}

	// 3. 合并结果
	allContexts := make([]string, 0, len(globalContexts)+len(localContexts))
	allContexts = append(allContexts, globalContexts...)
	allContexts = append(allContexts, localContexts...)

	// 4. 去重
	return gr.deduplicateContexts(allContexts), nil
}

// getAllCommunitySummaries 获取所有社区摘要
func (gr *GraphRAG) getAllCommunitySummaries(graph *KnowledgeGraph) []*CommunitySummary {
	summaries := make([]*CommunitySummary, 0)

	// 检测社区
	communities, _ := gr.detector.DetectCommunities(graph)

	for _, comm := range communities {
		if comm.Summary != "" {
			summaries = append(summaries, &CommunitySummary{
				CommunityID: comm.ID,
				Summary:     comm.Summary,
				EntityCount: len(comm.Entities),
			})
		}
	}

	return summaries
}

// CommunitySummary 社区摘要
type CommunitySummary struct {
	CommunityID string
	Summary     string
	EntityCount int
}

// rankSummaries 对社区摘要进行排序
func (gr *GraphRAG) rankSummaries(ctx context.Context, query string, summaries []*CommunitySummary, topK int) []*CommunitySummary {
	// 简单实现：基于关键词匹配
	scores := make([]float64, len(summaries))
	queryWords := strings.Fields(strings.ToLower(query))

	for i, summary := range summaries {
		summaryLower := strings.ToLower(summary.Summary)
		score := 0.0

		for _, word := range queryWords {
			if strings.Contains(summaryLower, word) {
				score += 1.0
			}
		}

		scores[i] = score
	}

	// 排序
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j] > scores[i] {
				scores[i], scores[j] = scores[j], scores[i]
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	// 返回 Top-K
	if len(summaries) > topK {
		summaries = summaries[:topK]
	}

	return summaries
}

// extractQueryEntities 从查询中提取实体
func (gr *GraphRAG) extractQueryEntities(ctx context.Context, query string) []string {
	// 简单实现：提取大写开头的词
	words := strings.Fields(query)
	entities := make([]string, 0)

	for _, word := range words {
		if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
			entities = append(entities, word)
		}
	}

	return entities
}

// findRelevantEntities 找到相关实体
func (gr *GraphRAG) findRelevantEntities(graph *KnowledgeGraph, queryEntities []string, topK int) []*Entity {
	relevant := make([]*Entity, 0)

	for _, entity := range graph.Entities {
		// 检查是否在查询中
		for _, queryEnt := range queryEntities {
			if strings.Contains(entity.Name, queryEnt) || strings.Contains(queryEnt, entity.Name) {
				relevant = append(relevant, entity)
				break
			}
		}

		if len(relevant) >= topK {
			break
		}
	}

	return relevant
}

// buildEntityContexts 构建实体上下文
func (gr *GraphRAG) buildEntityContexts(graph *KnowledgeGraph, entities []*Entity) []string {
	contexts := make([]string, 0, len(entities))

	for _, entity := range entities {
		context := fmt.Sprintf("%s (%s): ", entity.Name, entity.Type)

		// 添加相关关系
		relationCount := 0
		for _, rel := range graph.Relations {
			if rel.From == entity.ID {
				context += fmt.Sprintf("%s→%s ", rel.To, rel.Type)
				relationCount++
			} else if rel.To == entity.ID {
				context += fmt.Sprintf("%s→%s ", rel.From, rel.Type)
				relationCount++
			}

			if relationCount >= 5 {
				// 最多显示 5 个关系
				break
			}
		}

		contexts = append(contexts, context)
	}

	return contexts
}

// generateCommunitySummary 生成社区摘要
func (gr *GraphRAG) generateCommunitySummary(ctx context.Context, graph *KnowledgeGraph, comm *Community) (string, error) {
	// 获取社区中的实体
	entities := make([]*Entity, 0, len(comm.Entities))
	for _, entityID := range comm.Entities {
		for _, e := range graph.Entities {
			if e.ID == entityID {
				entities = append(entities, e)
				break
			}
		}
	}

	// 构建提示
	prompt := fmt.Sprintf(`请为以下实体集合生成一个简短的摘要，描述这些实体的共同主题或关系。

实体列表:
%s

要求:
1. 简明扼要地描述这些实体的共同主题
2. 突出关键实体和关系
3. 摘要长度: 50-100 字
4. 只输出摘要内容，不要解释

摘要:`, gr.formatEntityList(entities))

	// 生成摘要
	response, err := gr.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// formatEntityList 格式化实体列表
func (gr *GraphRAG) formatEntityList(entities []*Entity) string {
	lines := make([]string, 0, len(entities))

	for _, e := range entities {
		lines = append(lines, fmt.Sprintf("- %s (%s)", e.Name, e.Type))
	}

	return strings.Join(lines, "\n")
}

// deduplicateContexts 去重上下文
func (gr *GraphRAG) deduplicateContexts(contexts []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(contexts))

	for _, ctx := range contexts {
		if !seen[ctx] {
			seen[ctx] = true
			result = append(result, ctx)
		}
	}

	return result
}

// GetHierarchicalSummaries 获取层次化摘要
func (gr *GraphRAG) GetHierarchicalSummaries(ctx context.Context, graph *KnowledgeGraph) *CommunityHierarchy {
	hierarchy := gr.detector.HierarchicalCluster(graph, gr.config.CommunityLevels)

	// 为每个层次生成摘要
	for _, level := range hierarchy.Levels {
		for _, comm := range level.Communities {
			if comm.Summary == "" {
				summary, err := gr.generateCommunitySummary(ctx, graph, comm)
				if err == nil {
					comm.Summary = summary
				}
			}
		}
	}

	return hierarchy
}
