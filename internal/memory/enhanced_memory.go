package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// UserMemory 用户记忆
type UserMemory struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Content   string                 `json:"content"`
	Topics    []string               `json:"topics"`
	Importance float64                `json:"importance"` // 0-1，重要性评分
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	AccessedAt time.Time              `json:"accessed_at"` // 最后访问时间
	AccessCount int                   `json:"access_count"` // 访问次数
	Vector    []float64              `json:"vector"`       // 用于语义检索
	mu        sync.RWMutex
}

// EnhancedMemoryManager 增强版记忆管理器
type EnhancedMemoryManager struct {
	mu              sync.RWMutex
	memories        map[string][]*UserMemory // userID -> memories
	embeddingModel  llm.Model
	enableAutoExtract bool
	enableSemanticSearch bool
	optimizationStrategy string // "summarization", "time_decay", "importance"
}

// NewEnhancedMemoryManager 创建增强版记忆管理器
func NewEnhancedMemoryManager(embeddingModel llm.Model) *EnhancedMemoryManager {
	return &EnhancedMemoryManager{
		memories:            make(map[string][]*UserMemory),
		embeddingModel:      embeddingModel,
		enableAutoExtract:   embeddingModel != nil,
		enableSemanticSearch: embeddingModel != nil,
		optimizationStrategy: "importance", // 默认重要性优化
	}
}

// ExtractMemories 自动从对话中提取记忆
func (m *EnhancedMemoryManager) ExtractMemories(ctx context.Context, userID string, conversation string) ([]*UserMemory, error) {
	if !m.enableAutoExtract || m.embeddingModel == nil {
		return nil, fmt.Errorf("auto extract not enabled or no embedding model")
	}

	// 构建提取提示
	prompt := fmt.Sprintf(`请从以下对话中提取需要记住的重要信息。

对话内容：
%s

请提取以下类型的信息：
1. 用户偏好（喜欢的、不喜欢的）
2. 个人信息（姓名、职业、兴趣等）
3. 重要决定或计划
4. 其他值得记住的信息

请以JSON格式返回，格式如下：
[
  {
    "content": "用户喜欢Go语言编程",
    "topics": ["编程", "Go", "偏好"],
    "importance": 0.8
  }
]

只返回JSON数组，不要其他内容。`, conversation)

	// 调用LLM提取
	response, err := m.embeddingModel.Chat(ctx, []models.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return nil, fmt.Errorf("failed to extract memories: %w", err)
	}

	// 解析响应（简化版，实际应用中应该有更健壮的JSON解析）
	memories := m.parseExtractedMemories(response, userID)

	// 为每条记忆计算embedding
	for _, memory := range memories {
		if m.embeddingModel.SupportsEmbedding() {
			vector, err := m.embeddingModel.Embed(ctx, memory.Content)
			if err == nil {
				memory.Vector = vector
			}
		}
	}

	return memories, nil
}

// parseExtractedMemories 解析提取的记忆
func (m *EnhancedMemoryManager) parseExtractedMemories(response, userID string) []*UserMemory {
	// 简化版：按行分割
	lines := strings.Split(response, "\n")
	memories := make([]*UserMemory, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// 简单解析（实际应该用JSON解析）
		if strings.Contains(line, "content:") {
			content := extractValue(line, "content")
			topics := extractTopics(line)
			importance := extractImportance(line)

			memory := &UserMemory{
				ID:        generateID(),
				UserID:    userID,
				Content:   content,
				Topics:    topics,
				Importance: importance,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				AccessedAt: time.Now(),
			}
			memories = append(memories, memory)
		}
	}

	return memories
}

// extractValue 提取字段值
func extractValue(line, field string) string {
	start := strings.Index(line, field+": \"")
	if start == -1 {
		return ""
	}
	start += len(field) + 3 // 跳过 "field: \""
	end := strings.Index(line[start:], "\"")
	if end == -1 {
		return line[start:]
	}
	return line[start : start+end]
}

// extractTopics 提取主题
func extractTopics(line string) []string {
	start := strings.Index(line, "topics:")
	if start == -1 {
		return []string{}
	}

	// 简化版：查找方括号内容
	BracketStart := strings.Index(line[start:], "[")
	if BracketStart == -1 {
		return []string{}
	}
	BracketStart += start

	BracketEnd := strings.Index(line[BracketStart:], "]")
	if BracketEnd == -1 {
		return []string{}
	}

	topicsStr := line[BracketStart+1 : BracketStart+BracketEnd]
	topics := strings.Split(topicsStr, ",")

	result := make([]string, 0, len(topics))
	for _, topic := range topics {
		topic = strings.TrimSpace(strings.Trim(topic, "\""))
		if topic != "" {
			result = append(result, topic)
		}
	}

	return result
}

// extractImportance 提取重要性
func extractImportance(line string) float64 {
	start := strings.Index(line, "importance:")
	if start == -1 {
		return 0.5 // 默认中等重要性
	}

	start += len("importance:")
	end := strings.Index(line[start:], ",")
	if end == -1 {
		end = strings.Index(line[start:], "}")
	}

	valueStr := strings.TrimSpace(line[start : start+end])
	valueStr = strings.TrimSuffix(valueStr, ",")

	var importance float64
	fmt.Sscanf(valueStr, "%f", &importance)

	if importance <= 0 {
		importance = 0.5
	}
	if importance > 1 {
		importance = 1.0
	}

	return importance
}

// AddMemory 添加记忆
func (m *EnhancedMemoryManager) AddMemory(ctx context.Context, memory *UserMemory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否需要去重
	memories := m.memories[memory.UserID]
	for _, existing := range memories {
		if similarity := cosineSimilarity(memory.Vector, existing.Vector); similarity > 0.9 {
			// 相似度过高，合并记忆
			m.mergeMemories(existing, memory)
			return nil
		}
	}

	// 添加新记忆
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()
	memory.AccessedAt = time.Now()

	m.memories[memory.UserID] = append(memories, memory)

	return nil
}

// mergeMemories 合并相似记忆
func (m *EnhancedMemoryManager) mergeMemories(existing, new *UserMemory) {
	// 合并内容
	existing.Content = fmt.Sprintf("%s | %s", existing.Content, new.Content)

	// 合并主题
	topicMap := make(map[string]bool)
	for _, topic := range existing.Topics {
		topicMap[topic] = true
	}
	for _, topic := range new.Topics {
		topicMap[topic] = true
	}

	topics := make([]string, 0, len(topicMap))
	for topic := range topicMap {
		topics = append(topics, topic)
	}
	existing.Topics = topics

	// 更新重要性（取最大值）
	if new.Importance > existing.Importance {
		existing.Importance = new.Importance
	}

	existing.UpdatedAt = time.Now()
}

// SemanticSearch 语义检索记忆
func (m *EnhancedMemoryManager) SemanticSearch(ctx context.Context, userID string, query string, limit int) ([]*UserMemory, error) {
	if !m.enableSemanticSearch || m.embeddingModel == nil {
		return m.GetMemories(userID, limit), nil
	}

	// 向量化查询
	queryVector, err := m.embeddingModel.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	memories := m.memories[userID]

	// 计算相似度并排序
	type memoryScore struct {
		memory *UserMemory
		score  float64
	}

	scores := make([]memoryScore, 0, len(memories))
	for _, memory := range memories {
		similarity := cosineSimilarity(queryVector, memory.Vector)

		// 更新访问统计
		memory.mu.Lock()
		memory.AccessedAt = time.Now()
		memory.AccessCount++
		memory.mu.Unlock()

		scores = append(scores, memoryScore{
			memory: memory,
			score:  similarity,
		})
	}

	// 按相似度排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回topK
	if limit > len(scores) {
		limit = len(scores)
	}

	result := make([]*UserMemory, limit)
	for i := 0; i < limit; i++ {
		result[i] = scores[i].memory
	}

	return result, nil
}

// GetMemories 获取用户记忆（带优化）
func (m *EnhancedMemoryManager) GetMemories(userID string, limit int) []*UserMemory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memories := m.memories[userID]

	// 根据优化策略排序和过滤
	optimized := m.optimizeMemories(memories)

	if limit > 0 && limit < len(optimized) {
		return optimized[:limit]
	}

	return optimized
}

// optimizeMemories 优化记忆
func (m *EnhancedMemoryManager) optimizeMemories(memories []*UserMemory) []*UserMemory {
	switch m.optimizationStrategy {
	case "time_decay":
		return m.timeDecayOptimization(memories)
	case "importance":
		return m.importanceOptimization(memories)
	case "summarization":
		return m.summarizationOptimization(memories)
	default:
		return memories
	}
}

// timeDecayOptimization 时间衰减优化
func (m *EnhancedMemoryManager) timeDecayOptimization(memories []*UserMemory) []*UserMemory {
	now := time.Now()

	type memoryScore struct {
		memory *UserMemory
		score  float64
	}

	scores := make([]memoryScore, len(memories))
	for i, memory := range memories {
		// 计算时间衰减得分
		hoursPassed := now.Sub(memory.AccessedAt).Hours()
		decayScore := memory.Importance * math.Exp(-hoursPassed/720) // 30天半衰期

		scores[i] = memoryScore{
			memory: memory,
			score:  decayScore,
		}
	}

	// 按得分排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]*UserMemory, len(scores))
	for i := range scores {
		result[i] = scores[i].memory
	}

	return result
}

// importanceOptimization 重要性优化
func (m *EnhancedMemoryManager) importanceOptimization(memories []*UserMemory) []*UserMemory {
	// 按重要性排序
	sorted := make([]*UserMemory, len(memories))
	copy(sorted, memories)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Importance > sorted[j].Importance
	})

	return sorted
}

// summarizationOptimization 摘要优化（合并相似记忆）
func (m *EnhancedMemoryManager) summarizationOptimization(memories []*UserMemory) []*UserMemory {
	// 简化版：按主题分组
	groups := make(map[string][]*UserMemory)

	for _, memory := range memories {
		for _, topic := range memory.Topics {
			groups[topic] = append(groups[topic], memory)
		}
	}

	// 为每个主题选择最重要的记忆
	result := make([]*UserMemory, 0)
	for _, group := range groups {
		if len(group) == 1 {
			result = append(result, group[0])
		} else {
			// 选择重要性最高的
			best := group[0]
			for _, mem := range group[1:] {
				if mem.Importance > best.Importance {
					best = mem
				}
			}
			result = append(result, best)
		}
	}

	return result
}

// OptimizeMemories 手动触发优化
func (m *EnhancedMemoryManager) OptimizeMemories(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	memories := m.memories[userID]
	optimized := m.optimizeMemories(memories)

	// 删除重复或低质量的记忆
	m.memories[userID] = optimized

	return nil
}

// SetOptimizationStrategy 设置优化策略
func (m *EnhancedMemoryManager) SetOptimizationStrategy(strategy string) {
	switch strategy {
	case "summarization", "time_decay", "importance":
		m.optimizationStrategy = strategy
	default:
		m.optimizationStrategy = "importance"
	}
}

// EnableAutoExtract 启用自动提取
func (m *EnhancedMemoryManager) EnableAutoExtract(enable bool) {
	m.enableAutoExtract = enable && m.embeddingModel != nil
}

// EnableSemanticSearch 启用语义检索
func (m *EnhancedMemoryManager) EnableSemanticSearch(enable bool) {
	m.enableSemanticSearch = enable && m.embeddingModel != nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// generateID 生成ID
func generateID() string {
	return fmt.Sprintf("mem_%d", time.Now().UnixNano())
}
