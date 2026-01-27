package memory

import (
	"fmt"
	"sync"
	"time"

	"ai-agent-assistant/pkg/models"
)

// Memory 记忆存储接口
type Memory interface {
	AddMessage(sessionID string, message models.Message) error
	GetHistory(sessionID string) ([]models.Message, error)
	Clear(sessionID string) error
	GetSession(sessionID string) (*models.Session, error)
	CreateSession(sessionID string, model string) (*models.Session, error)
}

// InMemoryMemory 内存存储实现
type InMemoryMemory struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
	maxHistory int
}

// NewInMemoryMemory 创建内存存储
func NewInMemoryMemory(maxHistory int) *InMemoryMemory {
	return &InMemoryMemory{
		sessions:   make(map[string]*models.Session),
		maxHistory: maxHistory,
	}
}

// AddMessage 添加消息到会话
func (m *InMemoryMemory) AddMessage(sessionID string, message models.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		session = &models.Session{
			ID:       sessionID,
			Messages: make([]models.Message, 0, m.maxHistory),
			Metadata: make(map[string]interface{}),
		}
		m.sessions[sessionID] = session
	}

	session.Messages = append(session.Messages, message)

	// 限制历史记录数量
	if len(session.Messages) > m.maxHistory {
		session.Messages = session.Messages[len(session.Messages)-m.maxHistory:]
	}

	return nil
}

// GetHistory 获取会话历史
func (m *InMemoryMemory) GetHistory(sessionID string) ([]models.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return []models.Message{}, nil
	}

	return session.Messages, nil
}

// Clear 清除会话
func (m *InMemoryMemory) Clear(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// GetSession 获取会话
func (m *InMemoryMemory) GetSession(sessionID string) (*models.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// CreateSession 创建新会话
func (m *InMemoryMemory) CreateSession(sessionID string, model string) (*models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	session := &models.Session{
		ID:       sessionID,
		Messages: make([]models.Message, 0, m.maxHistory),
		Model:    model,
		Metadata: map[string]interface{}{
			"created_at": time.Now().Unix(),
		},
	}

	m.sessions[sessionID] = session
	return session, nil
}

// MemoryManager 记忆管理器
type MemoryManager struct {
	store Memory
}

// NewMemoryManager 创建记忆管理器
func NewMemoryManager(maxHistory int, storeType string) *MemoryManager {
	var store Memory

	switch storeType {
	case "memory":
		store = NewInMemoryMemory(maxHistory)
	default:
		store = NewInMemoryMemory(maxHistory)
	}

	return &MemoryManager{
		store: store,
	}
}

// AddMessage 添加消息
func (m *MemoryManager) AddMessage(sessionID string, message models.Message) error {
	return m.store.AddMessage(sessionID, message)
}

// GetHistory 获取历史
func (m *MemoryManager) GetHistory(sessionID string) ([]models.Message, error) {
	return m.store.GetHistory(sessionID)
}

// ClearSession 清除会话
func (m *MemoryManager) ClearSession(sessionID string) error {
	return m.store.Clear(sessionID)
}

// GetOrCreateSession 获取或创建会话
func (m *MemoryManager) GetOrCreateSession(sessionID string, model string) (*models.Session, error) {
	session, err := m.store.GetSession(sessionID)
	if err == nil {
		return session, nil
	}

	return m.store.CreateSession(sessionID, model)
}

// GetSession 获取会话
func (m *MemoryManager) GetSession(sessionID string) (*models.Session, error) {
	return m.store.GetSession(sessionID)
}
