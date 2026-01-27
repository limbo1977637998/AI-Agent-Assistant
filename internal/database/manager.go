package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Manager 数据库管理器
type Manager struct {
	client  *MySQLClient
	Sessions SessionRepository
	Messages MessageRepository
	ToolCalls ToolCallRepository
	UserMemories UserMemoryRepository
	AgentRuns AgentRunRepository
}

// NewManager 创建数据库管理器
func NewManager(config *MySQLConfig) (*Manager, error) {
	// 创建MySQL客户端
	client, err := NewMySQLClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mysql client: %w", err)
	}

	// 创建各个Repository
	sessionsRepo := NewSessionRepository(client.GetDB())
	messagesRepo := NewMessageRepository(client.GetDB())
	toolCallsRepo := NewToolCallRepository(client.GetDB())
	userMemoriesRepo := NewUserMemoryRepository(client.GetDB())
	agentRunsRepo := NewAgentRunRepository(client.GetDB())

	manager := &Manager{
		client:       client,
		Sessions:     sessionsRepo,
		Messages:     messagesRepo,
		ToolCalls:    toolCallsRepo,
		UserMemories: userMemoriesRepo,
		AgentRuns:    agentRunsRepo,
	}

	return manager, nil
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	return m.client.Close()
}

// GetClient 获取MySQL客户端
func (m *Manager) GetClient() *MySQLClient {
	return m.client
}

// BeginTxC 开始事务
func (m *Manager) BeginTxC() (*Transaction, error) {
	tx, err := m.client.BeginTxC()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx:      tx,
		manager: m,
	}, nil
}

// Transaction 事务
type Transaction struct {
	tx      *sqlx.Tx
	manager *Manager
}

// Commit 提交事务
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback 回滚事务
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// GetTx 获取事务对象
func (t *Transaction) GetTx() *sqlx.Tx {
	return t.tx
}

// Ping 测试数据库连接
func (m *Manager) Ping(ctx context.Context) error {
	return m.client.GetDB().PingContext(ctx)
}
