package graph

import (
	"fmt"
	"strings"
)

// CommunityDetector 社区检测器接口
type CommunityDetector interface {
	// DetectCommunities 检测社区
	DetectCommunities(graph *KnowledgeGraph) ([]*Community, error)

	// HierarchicalCluster 层次化聚类
	HierarchicalCluster(graph *KnowledgeGraph, levels int) *CommunityHierarchy
}

// LouvainDetector Louvain 社区检测器
//
// 算法说明:
//   Louvain 算法是一种基于模块度优化的社区发现算法
//   通过迭代优化节点分配来最大化模块度
//
// 优点:
//   - 快速高效
//   - 可以处理大规模图
//   - 自动发现社区数量
//
// 参考:
//   "Fast unfolding of communities in large networks" (Blondel et al., 2008)
type LouvainDetector struct {
	resolution float64 // 分辨率参数 (影响社区大小)
}

// NewLouvainDetector 创建 Louvain 检测器
func NewLouvainDetector(resolution float64) *LouvainDetector {
	if resolution <= 0 {
		resolution = 1.0 // 默认分辨率
	}

	return &LouvainDetector{
		resolution: resolution,
	}
}

// DetectCommunities 检测社区
func (ld *LouvainDetector) DetectCommunities(graph *KnowledgeGraph) ([]*Community, error) {
	if graph == nil || len(graph.Entities) == 0 {
		return []*Community{}, nil
	}

	// 初始化：每个节点一个社区
	communities := make(map[string]*Community)
	for _, entity := range graph.Entities {
		comm := &Community{
			ID:        entity.ID,
			Entities:  []string{entity.ID},
			EntityIDs: map[string]bool{entity.ID: true},
		}
		communities[entity.ID] = comm
	}

	// 迭代优化
	improved := true
	iteration := 0
	maxIterations := 100

	for improved && iteration < maxIterations {
		improved = false
		iteration++

		// 第一阶段：节点移动
		for _, entity := range graph.Entities {
			oldCommID := ld.getCommunityID(communities, entity.ID)

			// 计算将节点移动到每个邻居社区带来的模块度增益
			bestCommID := oldCommID
			maxDeltaMod := 0.0

			neighbors := ld.getNeighbors(graph, entity.ID)
			for _, neighborID := range neighbors {
				newCommID := ld.getCommunityID(communities, neighborID)
				deltaMod := ld.calculateDeltaModularity(graph, communities, entity.ID, oldCommID, newCommID)

				if deltaMod > maxDeltaMod {
					maxDeltaMod = deltaMod
					bestCommID = newCommID
				}
			}

			// 如果有改进，移动节点
			if bestCommID != oldCommID && maxDeltaMod > 0 {
				ld.moveNode(communities, entity.ID, oldCommID, bestCommID)
				improved = true
			}
		}

		// 第二阶段：社区合并
		if improved {
			ld.mergeCommunities(communities)
		}
	}

	// 转换为结果
	result := make([]*Community, 0, len(communities))
	for _, comm := range communities {
		if len(comm.Entities) > 0 {
			comm.Level = 0
			result = append(result, comm)
		}
	}

	return result, nil
}

// HierarchicalCluster 层次化聚类
func (ld *LouvainDetector) HierarchicalCluster(graph *KnowledgeGraph, levels int) *CommunityHierarchy {
	hierarchy := &CommunityHierarchy{
		Levels: make([]*HierarchyLevel, 0, levels),
	}

	// 第 0 层：原始节点
	communities, _ := ld.DetectCommunities(graph)
	hierarchy.Levels = append(hierarchy.Levels, &HierarchyLevel{
		Level:       0,
		Communities: communities,
	})

	// 上层层次：聚合社区
	for level := 1; level < levels; level++ {
		// 合并小社区
		mergedCommunities := ld.mergeSmallCommunities(communities)

		hierarchy.Levels = append(hierarchy.Levels, &HierarchyLevel{
			Level:       level,
			Communities: mergedCommunities,
		})

		communities = mergedCommunities

		// 如果只剩一个社区，停止
		if len(communities) <= 1 {
			break
		}
	}

	return hierarchy
}

// getCommunityID 获取节点所属的社区 ID
func (ld *LouvainDetector) getCommunityID(communities map[string]*Community, entityID string) string {
	for commID, comm := range communities {
		if comm.EntityIDs[entityID] {
			return commID
		}
	}
	return ""
}

// getNeighbors 获取节点的邻居
func (ld *LouvainDetector) getNeighbors(graph *KnowledgeGraph, entityID string) []string {
	neighbors := make([]string, 0)

	for _, rel := range graph.Relations {
		if rel.From == entityID {
			neighbors = append(neighbors, rel.To)
		} else if rel.To == entityID {
			neighbors = append(neighbors, rel.From)
		}
	}

	return neighbors
}

// calculateDeltaModularity 计算模块度增益
func (ld *LouvainDetector) calculateDeltaModularity(
	graph *KnowledgeGraph,
	communities map[string]*Community,
	entityID string,
	oldCommID string,
	newCommID string,
) float64 {
	// 简化的模块度计算
	// 实际实现应该更复杂

	// 计算节点在旧社区中的度
	oldCommInternalDegree := ld.calculateInternalDegree(graph, communities, entityID, oldCommID)

	// 计算节点在新社区中的度
	newCommInternalDegree := ld.calculateInternalDegree(graph, communities, entityID, newCommID)

	// 节点的总度
	totalDegree := ld.calculateTotalDegree(graph, entityID)

	// 模块度增益（简化公式）
	delta := float64(newCommInternalDegree-oldCommInternalDegree) / float64(totalDegree)

	return delta
}

// calculateInternalDegree 计算社区内部度
func (ld *LouvainDetector) calculateInternalDegree(
	graph *KnowledgeGraph,
	communities map[string]*Community,
	entityID string,
	commID string,
) int {
	degree := 0
	comm := communities[commID]
	if comm == nil {
		return 0
	}

	for _, rel := range graph.Relations {
		if rel.From == entityID && comm.EntityIDs[rel.To] {
			degree++
		} else if rel.To == entityID && comm.EntityIDs[rel.From] {
			degree++
		}
	}

	return degree
}

// calculateTotalDegree 计算节点的总度
func (ld *LouvainDetector) calculateTotalDegree(graph *KnowledgeGraph, entityID string) int {
	degree := 0

	for _, rel := range graph.Relations {
		if rel.From == entityID || rel.To == entityID {
			degree++
		}
	}

	return degree
}

// moveNode 移动节点到新社区
func (ld *LouvainDetector) moveNode(
	communities map[string]*Community,
	entityID string,
	oldCommID string,
	newCommID string,
) {
	// 从旧社区移除
	oldComm := communities[oldCommID]
	if oldComm != nil {
		delete(oldComm.EntityIDs, entityID)
		entities := make([]string, 0, len(oldComm.Entities)-1)
		for _, e := range oldComm.Entities {
			if e != entityID {
				entities = append(entities, e)
			}
		}
		oldComm.Entities = entities
	}

	// 添加到新社区
	newComm := communities[newCommID]
	if newComm != nil {
		newComm.EntityIDs[entityID] = true
		newComm.Entities = append(newComm.Entities, entityID)
	}
}

// mergeCommunities 合并社区
func (ld *LouvainDetector) mergeCommunities(communities map[string]*Community) {
	// 简化实现：不做实际合并
	// 实际 Louvain 算法会将社区压缩成超节点
}

// mergeSmallCommunities 合并小社区
func (ld *LouvainDetector) mergeSmallCommunities(communities []*Community) []*Community {
	// 找出小社区（少于 3 个节点）
	smallComms := make([]*Community, 0)
	largeComms := make([]*Community, 0)

	for _, comm := range communities {
		if len(comm.Entities) < 3 {
			smallComms = append(smallComms, comm)
		} else {
			largeComms = append(largeComms, comm)
		}
	}

	// 将小社区合并到最近的大社区
	for _, smallComm := range smallComms {
		bestLargeComm := ld.findClosestCommunity(smallComm, largeComms)
		if bestLargeComm != nil {
			// 合并
			bestLargeComm.Entities = append(bestLargeComm.Entities, smallComm.Entities...)
			for _, entityID := range smallComm.Entities {
				bestLargeComm.EntityIDs[entityID] = true
			}
		} else {
			// 没有找到大社区，保留
			largeComms = append(largeComms, smallComm)
		}
	}

	return largeComms
}

// findClosestCommunity 找到最近的社区
func (ld *LouvainDetector) findClosestCommunity(smallComm *Community, largeComms []*Community) *Community {
	var bestComm *Community
	maxOverlap := 0

	for _, largeComm := range largeComms {
		overlap := ld.calculateCommunityOverlap(smallComm, largeComm)
		if overlap > maxOverlap {
			maxOverlap = overlap
			bestComm = largeComm
		}
	}

	if maxOverlap > 0 {
		return bestComm
	}

	return nil
}

// calculateCommunityOverlap 计算社区重叠度
func (ld *LouvainDetector) calculateCommunityOverlap(comm1, comm2 *Community) int {
	overlap := 0

	for _, entityID := range comm1.Entities {
		if comm2.EntityIDs[entityID] {
			overlap++
		}
	}

	return overlap
}

// KnowledgeGraph 知识图谱
type KnowledgeGraph struct {
	Entities  []*Entity   `json:"entities"`
	Relations []*Relation `json:"relations"`
}

// Community 社区
type Community struct {
	ID        string          `json:"id"`
	Entities  []string        `json:"entities"`
	EntityIDs map[string]bool `json:"entity_ids"`
	Level     int             `json:"level"`      // 层次级别
	Summary   string          `json:"summary"`    // 社区摘要
	Metadata  map[string]interface{} `json:"metadata"`
}

// CommunityHierarchy 社区层次结构
type CommunityHierarchy struct {
	Levels []*HierarchyLevel `json:"levels"`
}

// HierarchyLevel 层次级别
type HierarchyLevel struct {
	Level       int         `json:"level"`
	Communities []*Community `json:"communities"`
}

// calculateModularity 计算模块度（用于评估社区质量）
func (ld *LouvainDetector) calculateModularity(graph *KnowledgeGraph, communities map[string]*Community) float64 {
	m := 0.0 // 模块度

	// 总边数
	totalEdges := len(graph.Relations)
	if totalEdges == 0 {
		return 0.0
	}

	// 计算每个社区内部的边
	internalEdges := 0
	for _, comm := range communities {
		for _, rel := range graph.Relations {
			if comm.EntityIDs[rel.From] && comm.EntityIDs[rel.To] {
				internalEdges++
			}
		}
	}

	// 简化的模块度计算
	m = float64(internalEdges) / float64(totalEdges)

	return m
}

// optimizeCommunities 优化社区（使用模拟退火）
func (ld *LouvainDetector) optimizeCommunities(graph *KnowledgeGraph, communities []*Community) []*Community {
	// 简化实现：直接返回
	return communities
}

// summarizeCommunities 为社区生成摘要
func (ld *LouvainDetector) summarizeCommunities(graph *KnowledgeGraph, communities []*Community) error {
	// 为每个社区生成摘要
	for _, comm := range communities {
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

		// 生成摘要（简单实现）
		comm.Summary = ld.generateSummary(entities)
	}

	return nil
}

// generateSummary 生成社区摘要
func (ld *LouvainDetector) generateSummary(entities []*Entity) string {
	if len(entities) == 0 {
		return ""
	}

	// 简单实现：列出实体名称
	entityNames := make([]string, 0, len(entities))
	for _, e := range entities {
		entityNames = append(entityNames, e.Name)
	}

	return fmt.Sprintf("社区包含 %d 个实体: %s", len(entities), strings.Join(entityNames, "、"))
}
