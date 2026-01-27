package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id int64) (*Message, error)
	GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*Message, error)
	CreateBatch(ctx context.Context, messages []*Message) error
	DeleteBySessionID(ctx context.Context, sessionID string) error
	CountBySessionID(ctx context.Context, sessionID string) (int64, error)
}

type messageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository 创建MessageRepository
func NewMessageRepository(db *sqlx.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (session_id, role, content, tool_calls, tokens_used)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		message.SessionID, message.Role, message.Content,
		message.ToolCalls, message.TokensUsed)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	message.ID = id
	return nil
}

// GetByID 根据ID获取消息
func (r *messageRepository) GetByID(ctx context.Context, id int64) (*Message, error) {
	query := `SELECT id, session_id, role, content, tool_calls, tokens_used, created_at
			  FROM messages WHERE id = ?`

	var message Message
	err := r.db.GetContext(ctx, &message, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message by id: %w", err)
	}

	return &message, nil
}

// GetBySessionID 获取会话的所有消息
func (r *messageRepository) GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*Message, error) {
	query := `
		SELECT id, session_id, role, content, tool_calls, tokens_used, created_at
		FROM messages
		WHERE session_id = ?
		ORDER BY created_at ASC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var messages []*Message
	err := r.db.SelectContext(ctx, &messages, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session_id: %w", err)
	}

	return messages, nil
}

// CreateBatch 批量创建消息
func (r *messageRepository) CreateBatch(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	query := `
		INSERT INTO messages (session_id, role, content, tool_calls, tokens_used)
		VALUES (?, ?, ?, ?, ?)
	`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, msg := range messages {
		_, err := stmt.ExecContext(ctx,
			msg.SessionID, msg.Role, msg.Content, msg.ToolCalls, msg.TokensUsed)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteBySessionID 删除会话的所有消息
func (r *messageRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {
	query := `DELETE FROM messages WHERE session_id = ?`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete messages by session_id: %w", err)
	}

	return nil
}

// CountBySessionID 统计会话的消息数量
func (r *messageRepository) CountBySessionID(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM messages WHERE session_id = ?`

	err := r.db.GetContext(ctx, &count, query, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// ToolCallRepository 工具调用仓储接口
type ToolCallRepository interface {
	Create(ctx context.Context, call *ToolCall) error
	GetByID(ctx context.Context, id int64) (*ToolCall, error)
	GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*ToolCall, error)
	GetByToolName(ctx context.Context, toolName string, limit int) ([]*ToolCall, error)
	List(ctx context.Context, limit, offset int) ([]*ToolCall, error)
}

type toolCallRepository struct {
	db *sqlx.DB
}

// NewToolCallRepository 创建ToolCallRepository
func NewToolCallRepository(db *sqlx.DB) ToolCallRepository {
	return &toolCallRepository{db: db}
}

// Create 创建工具调用记录
func (r *toolCallRepository) Create(ctx context.Context, call *ToolCall) error {
	query := `
		INSERT INTO tool_calls (session_id, user_id, tool_name, arguments, result, success, error_msg, duration)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		call.SessionID, call.UserID, call.ToolName, call.Arguments,
		call.Result, call.Success, call.ErrorMsg, call.Duration)
	if err != nil {
		return fmt.Errorf("failed to create tool call: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	call.ID = id
	return nil
}

// GetByID 根据ID获取工具调用
func (r *toolCallRepository) GetByID(ctx context.Context, id int64) (*ToolCall, error) {
	query := `SELECT id, session_id, user_id, tool_name, arguments, result, success, error_msg, duration, created_at
			  FROM tool_calls WHERE id = ?`

	var call ToolCall
	err := r.db.GetContext(ctx, &call, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tool call by id: %w", err)
	}

	return &call, nil
}

// GetBySessionID 获取会话的工具调用
func (r *toolCallRepository) GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*ToolCall, error) {
	query := `
		SELECT id, session_id, user_id, tool_name, arguments, result, success, error_msg, duration, created_at
		FROM tool_calls
		WHERE session_id = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var calls []*ToolCall
	err := r.db.SelectContext(ctx, &calls, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool calls by session_id: %w", err)
	}

	return calls, nil
}

// GetByToolName 根据工具名称获取调用记录
func (r *toolCallRepository) GetByToolName(ctx context.Context, toolName string, limit int) ([]*ToolCall, error) {
	query := `
		SELECT id, session_id, user_id, tool_name, arguments, result, success, error_msg, duration, created_at
		FROM tool_calls
		WHERE tool_name = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var calls []*ToolCall
	err := r.db.SelectContext(ctx, &calls, query, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool calls by tool_name: %w", err)
	}

	return calls, nil
}

// List 列出工具调用记录（分页）
func (r *toolCallRepository) List(ctx context.Context, limit, offset int) ([]*ToolCall, error) {
	query := `
		SELECT id, session_id, user_id, tool_name, arguments, result, success, error_msg, duration, created_at
		FROM tool_calls
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var calls []*ToolCall
	err := r.db.SelectContext(ctx, &calls, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tool calls: %w", err)
	}

	return calls, nil
}
