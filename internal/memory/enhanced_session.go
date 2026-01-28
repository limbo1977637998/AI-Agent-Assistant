package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"ai-agent-assistant/internal/llm"
	"ai-agent-assistant/pkg/models"
)

// EnhancedSessionManager 增强版会话管理器
type EnhancedSessionManager struct {
	mu              sync.RWMutex
	sessions        map[string]*EnhancedSession
	maxHistory      int
	enableAutoSummary bool
	summaryModel    llm.Model
	summaryThreshold int // 超过此消息数时自动摘要
	storeType       string // "memory", "mysql", "redis"
}

// EnhancedSession 增强版会话
type EnhancedSession struct {
	ID              string
	Model           string
	Messages        []models.Message
	Summary         string            // 会话摘要
	State           SessionState      // 结构化状态
	Metadata        map[string]interface{}
	CreatedAt       time.Time
	UpdatedAt       time.Time
	mu              sync.RWMutex
}

// SessionState 会话状态
type SessionState struct {
	Data     map[string]interface{} `json:"data"`
	Version  int                    `json:"version"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewEnhancedSessionManager 创建增强版会话管理器
func NewEnhancedSessionManager(maxHistory int, storeType string, summaryModel llm.Model) *EnhancedSessionManager {
	return &EnhancedSessionManager{
		sessions:          make(map[string]*EnhancedSession),
		maxHistory:        maxHistory,
		enableAutoSummary: summaryModel != nil,
		summaryModel:      summaryModel,
		summaryThreshold:  10, // 默认10条消息后摘要
		storeType:         storeType,
	}
}

// GetOrCreateSession 获取或创建会话（并发安全）
func (m *EnhancedSessionManager) GetOrCreateSession(sessionID, modelName string) (*EnhancedSession, error) {
	// 先尝试读锁获取
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		return session, nil
	}

	// 不存在，用写锁创建
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	session, exists = m.sessions[sessionID]
	if exists {
		return session, nil
	}

	// 创建新会话
	session = &EnhancedSession{
		ID:        sessionID,
		Model:     modelName,
		Messages:  make([]models.Message, 0, m.maxHistory),
		State: SessionState{
			Data:     make(map[string]interface{}),
			Version:  0,
			UpdatedAt: time.Now(),
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.sessions[sessionID] = session
	return session, nil
}

// AddMessage 添加消息（带自动摘要和并发控制）
func (m *EnhancedSessionManager) AddMessage(sessionID string, message models.Message) error {
	session, err := m.GetOrCreateSession(sessionID, "")
	if err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// 添加消息
	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()

	// 检查是否需要自动摘要
	if m.enableAutoSummary && len(session.Messages) > m.summaryThreshold {
		go m.autoSummary(sessionID) // 异步生成摘要
	}

	// 限制历史记录数量
	if len(session.Messages) > m.maxHistory {
		// 保留最新的消息，旧消息会被摘要替代
		oldMessages := session.Messages[:len(session.Messages)-m.maxHistory]
		session.Messages = session.Messages[len(session.Messages)-m.maxHistory:]

		// 触发摘要生成
		if m.enableAutoSummary {
			go m.generateSummaryFromMessages(sessionID, oldMessages)
		}
	}

	return nil
}

// autoSummary 自动摘要
func (m *EnhancedSessionManager) autoSummary(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := m.GetSession(sessionID)
	if err != nil {
		return
	}

	summary, err := m.generateSummary(ctx, session.Messages)
	if err != nil {
		return // 摘要生成失败，不影响主流程
	}

	session.mu.Lock()
	session.Summary = summary
	session.mu.Unlock()
}

// generateSummary 生成摘要
func (m *EnhancedSessionManager) generateSummary(ctx context.Context, messages []models.Message) (string, error) {
	if m.summaryModel == nil {
		return "", fmt.Errorf("summary model not configured")
	}

	// 构建摘要提示
	prompt := m.buildSummaryPrompt(messages)

	llmMessages := []models.Message{
		{Role: "user", Content: prompt},
	}

	return m.summaryModel.Chat(ctx, llmMessages)
}

// buildSummaryPrompt 构建摘要提示
func (m *EnhancedSessionManager) buildSummaryPrompt(messages []models.Message) string {
	var sb strings.Builder

	sb.WriteString("请将以下对话历史总结为简洁的摘要，保留关键信息：\n\n")

	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		if i >= 10 {
			// 只摘要最近10条
			sb.WriteString("... (更多对话)\n")
			break
		}
	}

	sb.WriteString("\n摘要要求：\n")
	sb.WriteString("1. 简洁明了\n")
	sb.WriteString("2. 保留关键信息（用户需求、重要决定等）\n")
	sb.WriteString("3. 控制在200字以内")

	return sb.String()
}

// generateSummaryFromMessages 从消息列表生成摘要
func (m *EnhancedSessionManager) generateSummaryFromMessages(sessionID string, messages []models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := m.generateSummary(ctx, messages)
	if err != nil {
		return
	}

	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		session.mu.Lock()
		session.Summary = summary
		session.mu.Unlock()
	}
}

// GetHistory 获取历史（带摘要）
func (m *EnhancedSessionManager) GetHistory(sessionID string) ([]models.Message, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// 如果有摘要，添加到开头
	messages := make([]models.Message, 0, len(session.Messages)+1)
	if session.Summary != "" {
		messages = append(messages, models.Message{
			Role:    "system",
			Content: "[会话摘要]\n" + session.Summary,
		})
	}

	messages = append(messages, session.Messages...)
	return messages, nil
}

// GetSession 获取会话（并发安全）
func (m *EnhancedSessionManager) GetSession(sessionID string) (*EnhancedSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// Clear 清除会话
func (m *EnhancedSessionManager) Clear(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// UpdateState 更新会话状态（并发安全，版本控制）
func (m *EnhancedSessionManager) UpdateState(sessionID string, updates map[string]interface{}) (int, error) {
	session, err := m.GetOrCreateSession(sessionID, "")
	if err != nil {
		return 0, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// 更新状态
	for key, value := range updates {
		session.State.Data[key] = value
	}

	// 增加版本号
	session.State.Version++
	session.State.UpdatedAt = time.Now()
	session.UpdatedAt = time.Now()

	return session.State.Version, nil
}

// GetState 获取会话状态
func (m *EnhancedSessionManager) GetState(sessionID string) (SessionState, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return SessionState{}, err
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// 返回状态的副本
	return SessionState{
		Data:     copyMap(session.State.Data),
		Version:  session.State.Version,
		UpdatedAt: session.State.UpdatedAt,
	}, nil
}

// SetMetadata 设置元数据
func (m *EnhancedSessionManager) SetMetadata(sessionID string, metadata map[string]interface{}) error {
	session, err := m.GetOrCreateSession(sessionID, "")
	if err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	for key, value := range metadata {
		session.Metadata[key] = value
	}

	session.UpdatedAt = time.Now()

	return nil
}

// GetMetadata 获取元数据
func (m *EnhancedSessionManager) GetMetadata(sessionID string) (map[string]interface{}, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// 返回副本
	return copyMap(session.Metadata), nil
}

// ListSessions 列出所有会话
func (m *EnhancedSessionManager) ListSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}

// GetSessionCount 获取会话数量
func (m *EnhancedSessionManager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// SetSummaryModel 设置摘要模型
func (m *EnhancedSessionManager) SetSummaryModel(model llm.Model) {
	m.summaryModel = model
	m.enableAutoSummary = model != nil
}

// SetSummaryThreshold 设置摘要阈值
func (m *EnhancedSessionManager) SetSummaryThreshold(threshold int) {
	m.summaryThreshold = threshold
}

// EnableAutoSummary 启用自动摘要
func (m *EnhancedSessionManager) EnableAutoSummary(enable bool) {
	m.enableAutoSummary = enable && m.summaryModel != nil
}

// copyMap 复制map
func copyMap(m map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{})
	for k, v := range m {
		copied[k] = v
	}
	return copied
}
