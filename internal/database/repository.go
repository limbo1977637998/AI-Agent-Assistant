package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SessionRepository 会话仓储接口
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id int64) (*Session, error)
	GetBySessionID(ctx context.Context, sessionID string) (*Session, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
	List(ctx context.Context, limit, offset int) ([]*Session, error)
}

type sessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository 创建SessionRepository
func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, session *Session) error {
	query := `INSERT INTO sessions (session_id, user_id, model, metadata) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		session.SessionID, session.UserID, session.Model, session.Metadata)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	session.ID = id
	return nil
}

// GetByID 根据ID获取会话
func (r *sessionRepository) GetByID(ctx context.Context, id int64) (*Session, error) {
	query := `SELECT id, session_id, user_id, model, metadata, created_at, updated_at FROM sessions WHERE id = ?`

	var session Session
	err := r.db.GetContext(ctx, &session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session by id: %w", err)
	}

	return &session, nil
}

// GetBySessionID 根据SessionID获取会话
func (r *sessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*Session, error) {
	query := `SELECT id, session_id, user_id, model, metadata, created_at, updated_at FROM sessions WHERE session_id = ?`

	var session Session
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session by session_id: %w", err)
	}

	return &session, nil
}

// GetByUserID 获取用户的所有会话
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*Session, error) {
	query := `SELECT id, session_id, user_id, model, metadata, created_at, updated_at FROM sessions WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`

	if limit <= 0 {
		limit = 10 // 默认限制
	}

	var sessions []*Session
	err := r.db.SelectContext(ctx, &sessions, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user_id: %w", err)
	}

	return sessions, nil
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, session *Session) error {
	query := `UPDATE sessions SET user_id = ?, model = ?, metadata = ? WHERE session_id = ?`

	result, err := r.db.ExecContext(ctx, query,
		session.UserID, session.Model, session.Metadata, session.SessionID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete 删除会话
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = ?`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List 列出所有会话（分页）
func (r *sessionRepository) List(ctx context.Context, limit, offset int) ([]*Session, error) {
	query := `SELECT id, session_id, user_id, model, metadata, created_at, updated_at FROM sessions ORDER BY created_at DESC LIMIT ? OFFSET ?`

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var sessions []*Session
	err := r.db.SelectContext(ctx, &sessions, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}
