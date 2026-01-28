package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// EnhancedGraphRAG 增强版 Graph RAG
//
// 新增功能:
//   1. 动态社区摘要（基于查询上下文）
//   2. 层次化检索（多层次社区遍历）
//   3. 社区权重计算
//   4. 实体重要性评分
//   5. 关系路径检索
//
// 论文基础:
//   "From Local to Global: A Graph RAG Approach to Query-Focused
//    Summarization of Repositories" (Microsoft Research, 2024)
type EnhancedGraphRAG struct {
	GraphRAG
	entityScorer    EntityScorer
	communityWeighter CommunityWeighter
	pathFinder      PathFinder
}

// EntityScorer 实体评分器接口
type EntityScorer interface {
	ScoreEntity(ctx context.Context, entity *Entity, query string) float64
}

// CommunityWeighter 社区权重计算器接口
type CommunityWeighter interface {
	WeightCommunity(ctx context.Context, community *Community, query string) float64
}

// PathFinder 路径查找器接口
type PathFinder interface {
	FindPath(ctx context.Context, graph *KnowledgeGraph, from, to string) ([]string, error)
}

// NewEnhancedGraphRAG 创建增强版 Graph RAG
func NewEnhancedGraphRAG(llm LLMProvider, config GraphRAGConfig) (*EnhancedGraphRAG, error) {
	baseRAG, err := NewGraphRAG(llm, config)
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedGraphRAG{
		GraphRAG:          *baseRAG,
		entityScorer:      &DefaultEntityScorer{},
		communityWeighter: &DefaultCommunityWeighter{},
		pathFinder:        &ShortestPathFinder{},
	}

	return enhanced, nil
}

// EnhancedGlobalSearch 增强版全局检索
// 特性:
//   1. 社区权重计算
//   2. 动态摘要生成
//   3. 层次化遍历
func (egr *EnhancedGraphRAG) EnhancedGlobalSearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 检测社区
	communities, err := egr.detector.DetectCommunities(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to detect communities: %w", err)
	}

	// 2. 计算社区权重
	type communityWeight struct {
		community *Community
		weight    float64
		summary   string
	}

	weightedCommunities := make([]*communityWeight, 0)

	for _, comm := range communities {
		// 计算社区权重
		weight := egr.communityWeighter.WeightCommunity(ctx, comm, query)

		// 生成或获取摘要
		summary := comm.Summary
		if summary == "" {
			summary, _ = egr.generateDynamicSummary(ctx, graph, comm, query)
		}

		weightedCommunities = append(weightedCommunities, &communityWeight{
			community: comm,
			weight:    weight,
			summary:   summary,
		})
	}

	// 3. 按权重排序
	sort.Slice(weightedCommunities, func(i, j int) bool {
		return weightedCommunities[i].weight > weightedCommunities[j].weight
	})

	// 4. 选择 Top-K 社区
	selectedCount := topK
	if selectedCount > len(weightedCommunities) {
		selectedCount = len(weightedCommunities)
	}

	// 5. 构建层次化上下文
	contexts := make([]string, 0)

	// 先添加最高权重的社区摘要
	for i := 0; i < selectedCount; i++ {
		contexts = append(contexts, weightedCommunities[i].summary)

		// 如果最高权重社区不够，添加其子社区
		if i == 0 && weightedCommunities[0].weight < 0.8 && selectedCount < len(communities) {
			// 添加子社区信息
			subContexts := egr.getSubCommunityContexts(ctx, graph, weightedCommunities[0].community, query)
			contexts = append(contexts, subContexts...)
		}
	}

	return contexts, nil
}

// EnhancedLocalSearch 增强版局部检索
// 特性:
//   1. 实体重要性评分
//   2. 多跳关系检索
//   3. 路径查找
func (egr *EnhancedGraphRAG) EnhancedLocalSearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 提取查询中的实体
	queryEntities := egr.extractQueryEntities(ctx, query)

	// 2. 为所有实体评分
	type entityScore struct {
		entity   *Entity
		score    float64
		relevant bool
	}

	scoredEntities := make([]*entityScore, 0)

	for _, entity := range graph.Entities {
		maxScore := 0.0
		isRelevant := false

		for _, queryEnt := range queryEntities {
			// 检查名称是否匹配
			if strings.Contains(strings.ToLower(entity.Name), strings.ToLower(queryEnt)) ||
			   strings.Contains(strings.ToLower(queryEnt), strings.ToLower(entity.Name)) {
				isRelevant = true
			}

			// 计算实体分数
			score := egr.entityScorer.ScoreEntity(ctx, entity, queryEnt)
			if score > maxScore {
				maxScore = score
			}
		}

		if isRelevant || maxScore > 0.3 {
			scoredEntities = append(scoredEntities, &entityScore{
				entity:   entity,
				score:    maxScore,
				relevant: isRelevant,
			})
		}
	}

	// 3. 按分数排序
	sort.Slice(scoredEntities, func(i, j int) bool {
		return scoredEntities[i].score > scoredEntities[j].score
	})

	// 4. 选择 Top-K 实体
	selectedCount := topK
	if selectedCount > len(scoredEntities) {
		selectedCount = len(scoredEntities)
	}

	// 5. 构建实体上下文（包括多跳关系）
	contexts := make([]string, 0)

	for i := 0; i < selectedCount; i++ {
		entity := scoredEntities[i].entity

		// 获取直接关系
		context := egr.buildEntityContext(graph, entity)

		// 如果是高相关实体，查找多跳关系
		if scoredEntities[i].relevant && scoredEntities[i].score > 0.7 {
			multiHopContexts := egr.buildMultiHopContext(ctx, graph, entity, query, 2) // 2 跳
			context += "\n相关实体: " + strings.Join(multiHopContexts, ", ")
		}

		contexts = append(contexts, context)
	}

	return contexts, nil
}

// generateDynamicSummary 生成动态社区摘要
// 特性: 基于查询上下文生成针对性摘要
func (egr *EnhancedGraphRAG) generateDynamicSummary(ctx context.Context, graph *KnowledgeGraph, comm *Community, query string) (string, error) {
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

	// 构建动态提示
	prompt := fmt.Sprintf(`请基于以下查询和实体集合，生成一个针对性的社区摘要。

查询: %s

社区实体:
%s

要求:
1. 摘要应该与查询相关
2. 突出与查询相关的实体和关系
3. 简明扼要，50-100 字
4. 只输出摘要内容

摘要:`, query, egr.formatEntityList(entities))

	// 生成摘要
	response, err := egr.llm.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// getSubCommunityContexts 获取子社区上下文
func (egr *EnhancedGraphRAG) getSubCommunityContexts(ctx context.Context, graph *KnowledgeGraph, parentComm *Community, query string) []string {
	// 检测子社区（层次化聚类）
	hierarchy := egr.detector.HierarchicalCluster(graph, 3) // 3 层

	// 找到下一级的子社区
	subContexts := make([]string, 0)

	for _, level := range hierarchy.Levels {
		if level.Level > parentComm.Level { // 假设有 Level 字段
			for _, comm := range level.Communities {
				// 检查是否是子社区（通过实体重叠判断）
				if egr.isSubCommunity(comm, parentComm) {
					summary := comm.Summary
					if summary == "" {
						summary, _ = egr.generateDynamicSummary(ctx, graph, comm, query)
					}
					subContexts = append(subContexts, summary)
				}
			}
		}
	}

	return subContexts
}

// isSubCommunity 判断是否是子社区
func (egr *EnhancedGraphRAG) isSubCommunity(child, parent *Community) bool {
	// 简化实现：检查实体重叠度
	parentEntities := make(map[string]bool)
	for _, id := range parent.Entities {
		parentEntities[id] = true
	}

	overlap := 0
	for _, id := range child.Entities {
		if parentEntities[id] {
			overlap++
		}
	}

	// 如果超过 50% 实体重叠，认为是子社区
	return float64(overlap) / float64(len(child.Entities)) > 0.5
}

// buildMultiHopContext 构建多跳关系上下文
func (egr *EnhancedGraphRAG) buildMultiHopContext(ctx context.Context, graph *KnowledgeGraph, entity *Entity, query string, maxHops int) []string {
	visited := make(map[string]bool)
	contexts := make([]string, 0)

	// BFS 遍历多跳关系
	queue := []*Entity{entity}
	visited[entity.ID] = true

	for hops := 0; hops < maxHops && len(queue) > 0; hops++ {
		levelSize := len(queue)

		for i := 0; i < levelSize; i++ {
			current := queue[0]
			queue = queue[1:]

			// 查找相关实体
			for _, rel := range graph.Relations {
				var nextEntity *Entity
				if rel.From == current.ID {
					nextEntity = egr.findEntityByID(graph, rel.To)
				} else if rel.To == current.ID {
					nextEntity = egr.findEntityByID(graph, rel.From)
				}

				if nextEntity != nil && !visited[nextEntity.ID] {
					// 评分实体
					score := egr.entityScorer.ScoreEntity(ctx, nextEntity, query)

					// 只添加高相关性的实体
					if score > 0.5 {
						contexts = append(contexts, fmt.Sprintf("%s(%s)", nextEntity.Name, rel.Type))
						visited[nextEntity.ID] = true
						queue = append(queue, nextEntity)
					}
				}
			}
		}
	}

	return contexts
}

// findEntityByID 根据 ID 查找实体
func (egr *EnhancedGraphRAG) findEntityByID(graph *KnowledgeGraph, id string) *Entity {
	for _, e := range graph.Entities {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// buildEntityContext 构建单个实体的上下文
func (egr *EnhancedGraphRAG) buildEntityContext(graph *KnowledgeGraph, entity *Entity) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("实体: %s (类型: %s)", entity.Name, entity.Type))
	if entity.Description != "" {
		sb.WriteString(fmt.Sprintf("\n描述: %s", entity.Description))
	}

	// 添加关系
	sb.WriteString("\n关系:")
	for _, rel := range graph.Relations {
		var relatedEntity *Entity
		relationType := rel.Type

		if rel.From == entity.ID {
			relatedEntity = egr.findEntityByID(graph, rel.To)
		} else if rel.To == entity.ID {
			relatedEntity = egr.findEntityByID(graph, rel.From)
		}

		if relatedEntity != nil {
			sb.WriteString(fmt.Sprintf("\n  - %s → %s", relationType, relatedEntity.Name))
		}
	}

	return sb.String()
}

// PathBasedSearch 基于路径的检索
// 特性: 查找实体间的关系路径
func (egr *EnhancedGraphRAG) PathBasedSearch(ctx context.Context, graph *KnowledgeGraph, query string, topK int) ([]string, error) {
	// 1. 提取查询中的实体对
	queryEntities := egr.extractQueryEntities(ctx, query)

	if len(queryEntities) < 2 {
		// 如果实体少于 2 个，回退到局部检索
		return egr.EnhancedLocalSearch(ctx, graph, query, topK)
	}

	// 2. 查找实体间的路径
	paths := make([][]string, 0)

	for i := 0; i < len(queryEntities); i++ {
		for j := i + 1; j < len(queryEntities); j++ {
			path, err := egr.pathFinder.FindPath(ctx, graph, queryEntities[i], queryEntities[j])
			if err == nil && len(path) > 0 {
				paths = append(paths, path)
			}
		}
	}

	// 3. 构建路径上下文
	contexts := make([]string, 0, len(paths))

	for _, path := range paths {
		context := egr.buildPathContext(graph, path)
		contexts = append(contexts, context)

		if len(contexts) >= topK {
			break
		}
	}

	return contexts, nil
}

// buildPathContext 构建路径上下文
func (egr *EnhancedGraphRAG) buildPathContext(graph *KnowledgeGraph, path []string) string {
	if len(path) == 0 {
		return ""
	}

	context := "关系路径: "
	for i := 0; i < len(path)-1; i++ {
		entity1 := egr.findEntityByID(graph, path[i])
		entity2 := egr.findEntityByID(graph, path[i+1])

		// 查找关系类型
		relationType := "相关"
		for _, rel := range graph.Relations {
			if (rel.From == path[i] && rel.To == path[i+1]) ||
			   (rel.To == path[i] && rel.From == path[i+1]) {
				relationType = rel.Type
				break
			}
		}

		if entity1 != nil && entity2 != nil {
			context += fmt.Sprintf("%s→%s→%s ", entity1.Name, relationType, entity2.Name)
		}
	}

	return context
}

// ===== 默认实现 =====

// DefaultEntityScorer 默认实体评分器
type DefaultEntityScorer struct{}

func (s *DefaultEntityScorer) ScoreEntity(ctx context.Context, entity *Entity, query string) float64 {
	queryLower := strings.ToLower(query)
	entityLower := strings.ToLower(entity.Name)
	typeLower := strings.ToLower(entity.Type)

	// 1. 名称匹配
	if strings.Contains(queryLower, entityLower) || strings.Contains(entityLower, queryLower) {
		return 1.0
	}

	// 2. 类型匹配
	if strings.Contains(queryLower, typeLower) {
		return 0.7
	}

	// 3. 描述匹配
	if entity.Description != "" {
		descLower := strings.ToLower(entity.Description)
		if strings.Contains(descLower, queryLower) {
			return 0.6
		}
	}

	return 0.3 // 默认低分
}

// DefaultCommunityWeighter 默认社区权重计算器
type DefaultCommunityWeighter struct{}

func (w *DefaultCommunityWeighter) WeightCommunity(ctx context.Context, community *Community, query string) float64 {
	queryLower := strings.ToLower(query)
	summaryLower := strings.ToLower(community.Summary)

	// 计算摘要中的关键词匹配度
	keywords := strings.Fields(queryLower)
	matchCount := 0

	for _, keyword := range keywords {
		if len(keyword) > 2 && strings.Contains(summaryLower, keyword) {
			matchCount++
		}
	}

	if len(keywords) == 0 {
		return 0.5
	}

	// 归一化权重
	weight := float64(matchCount) / float64(len(keywords))

	// 考虑社区大小（适度偏好中等大小的社区）
	size := len(community.Entities)
	if size < 3 {
		weight *= 0.8
	} else if size > 20 {
		weight *= 0.9
	}

	return weight
}

// ShortestPathFinder 最短路径查找器
type ShortestPathFinder struct{}

func (f *ShortestPathFinder) FindPath(ctx context.Context, graph *KnowledgeGraph, from, to string) ([]string, error) {
	// BFS 查找最短路径
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{from}
	visited[from] = true

	found := false
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == to {
			found = true
			break
		}

		// 查找邻居
		for _, rel := range graph.Relations {
			var neighbor string
			if rel.From == current {
				neighbor = rel.To
			} else if rel.To == current {
				neighbor = rel.From
			}

			if neighbor != "" && !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("path not found")
	}

	// 重建路径
	path := make([]string, 0)
	current := to

	for current != from {
		path = append([]string{current}, path...)
		current = parent[current]
	}
	path = append([]string{from}, path...)

	return path, nil
}
