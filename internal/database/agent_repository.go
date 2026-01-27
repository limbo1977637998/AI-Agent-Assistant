package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserMemoryRepository 用户记忆仓储接口
type UserMemoryRepository interface {
	Create(ctx context.Context, memory *UserMemory) error
	GetByID(ctx context.Context, id int64) (*UserMemory, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*UserMemory, error)
	Update(ctx context.Context, memory *UserMemory) error
	Delete(ctx context.Context, id int64) error
	SearchByTopic(ctx context.Context, userID string, topic string) ([]*UserMemory, error)
}

type userMemoryRepository struct {
	db *sqlx.DB
}

// NewUserMemoryRepository 创建UserMemoryRepository
func NewUserMemoryRepository(db *sqlx.DB) UserMemoryRepository {
	return &userMemoryRepository{db: db}
}

// Create 创建用户记忆
func (r *userMemoryRepository) Create(ctx context.Context, memory *UserMemory) error {
	query := `
		INSERT INTO user_memories (user_id, memory, topics, importance, memory_type)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		memory.UserID, memory.Memory, memory.Topics,
		memory.Importance, memory.MemoryType)
	if err != nil {
		return fmt.Errorf("failed to create user memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	memory.ID = id
	return nil
}

// GetByID 根据ID获取记忆
func (r *userMemoryRepository) GetByID(ctx context.Context, id int64) (*UserMemory, error) {
	query := `SELECT id, user_id, memory, topics, importance, memory_type, created_at, updated_at
			  FROM user_memories WHERE id = ?`

	var memory UserMemory
	err := r.db.GetContext(ctx, &memory, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user memory by id: %w", err)
	}

	return &memory, nil
}

// GetByUserID 获取用户的所有记忆
func (r *userMemoryRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*UserMemory, error) {
	query := `
		SELECT id, user_id, memory, topics, importance, memory_type, created_at, updated_at
		FROM user_memories
		WHERE user_id = ?
		ORDER BY importance DESC, created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var memories []*UserMemory
	err := r.db.SelectContext(ctx, &memories, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user memories: %w", err)
	}

	return memories, nil
}

// Update 更新记忆
func (r *userMemoryRepository) Update(ctx context.Context, memory *UserMemory) error {
	query := `
		UPDATE user_memories
		SET memory = ?, topics = ?, importance = ?, memory_type = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		memory.Memory, memory.Topics, memory.Importance,
		memory.MemoryType, memory.ID)
	if err != nil {
		return fmt.Errorf("failed to update user memory: %w", err)
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

// Delete 删除记忆
func (r *userMemoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_memories WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user memory: %w", err)
	}

	return nil
}

// SearchByTopic 根据主题搜索记忆
func (r *userMemoryRepository) SearchByTopic(ctx context.Context, userID string, topic string) ([]*UserMemory, error) {
	query := `
		SELECT id, user_id, memory, topics, importance, memory_type, created_at, updated_at
		FROM user_memories
		WHERE user_id = ? AND topics LIKE ?
		ORDER BY importance DESC
	`

	var memories []*UserMemory
	err := r.db.SelectContext(ctx, &memories, query, userID, "%"+topic+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search user memories: %w", err)
	}

	return memories, nil
}

// AgentRunRepository Agent运行记录仓储接口
type AgentRunRepository interface {
	Create(ctx context.Context, run *AgentRun) error
	GetByID(ctx context.Context, id int64) (*AgentRun, error)
	GetByRunID(ctx context.Context, runID string) (*AgentRun, error)
	GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*AgentRun, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*AgentRun, error)
	List(ctx context.Context, limit, offset int) ([]*AgentRun, error)
	GetStatsByDate(ctx context.Context, date time.Time) (map[string]int64, error)
}

type agentRunRepository struct {
	db *sqlx.DB
}

// NewAgentRunRepository 创建AgentRunRepository
func NewAgentRunRepository(db *sqlx.DB) AgentRunRepository {
	return &agentRunRepository{db: db}
}

// Create 创建Agent运行记录
func (r *agentRunRepository) Create(ctx context.Context, run *AgentRun) error {
	query := `
		INSERT INTO agent_runs
		(run_id, session_id, user_id, input, output, model, input_tokens, output_tokens, total_tokens,
		 estimated_cost, latency, success, error_msg, rag_used, tools_used)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		run.RunID, run.SessionID, run.UserID, run.Input, run.Output,
		run.Model, run.InputTokens, run.OutputTokens, run.TotalTokens,
		run.EstimatedCost, run.Latency, run.Success, run.ErrorMsg,
		run.RAGUsed, run.ToolsUsed)
	if err != nil {
		return fmt.Errorf("failed to create agent run: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	run.ID = id
	return nil
}

// GetByID 根据ID获取运行记录
func (r *agentRunRepository) GetByID(ctx context.Context, id int64) (*AgentRun, error) {
	query := `SELECT id, run_id, session_id, user_id, input, output, model, input_tokens, output_tokens,
			  total_tokens, estimated_cost, latency, success, error_msg, rag_used, tools_used, created_at
			  FROM agent_runs WHERE id = ?`

	var run AgentRun
	err := r.db.GetContext(ctx, &run, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent run by id: %w", err)
	}

	return &run, nil
}

// GetByRunID 根据RunID获取运行记录
func (r *agentRunRepository) GetByRunID(ctx context.Context, runID string) (*AgentRun, error) {
	query := `SELECT id, run_id, session_id, user_id, input, output, model, input_tokens, output_tokens,
			  total_tokens, estimated_cost, latency, success, error_msg, rag_used, tools_used, created_at
			  FROM agent_runs WHERE run_id = ?`

	var run AgentRun
	err := r.db.GetContext(ctx, &run, query, runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent run by run_id: %w", err)
	}

	return &run, nil
}

// GetBySessionID 获取会话的所有运行记录
func (r *agentRunRepository) GetBySessionID(ctx context.Context, sessionID string, limit int) ([]*AgentRun, error) {
	query := `
		SELECT id, run_id, session_id, user_id, input, output, model, input_tokens, output_tokens,
			   total_tokens, estimated_cost, latency, success, error_msg, rag_used, tools_used, created_at
		FROM agent_runs
		WHERE session_id = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var runs []*AgentRun
	err := r.db.SelectContext(ctx, &runs, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent runs by session_id: %w", err)
	}

	return runs, nil
}

// GetByUserID 获取用户的所有运行记录
func (r *agentRunRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*AgentRun, error) {
	query := `
		SELECT id, run_id, session_id, user_id, input, output, model, input_tokens, output_tokens,
			   total_tokens, estimated_cost, latency, success, error_msg, rag_used, tools_used, created_at
		FROM agent_runs
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var runs []*AgentRun
	err := r.db.SelectContext(ctx, &runs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent runs by user_id: %w", err)
	}

	return runs, nil
}

// List 列出运行记录（分页）
func (r *agentRunRepository) List(ctx context.Context, limit, offset int) ([]*AgentRun, error) {
	query := `
		SELECT id, run_id, session_id, user_id, input, output, model, input_tokens, output_tokens,
			   total_tokens, estimated_cost, latency, success, error_msg, rag_used, tools_used, created_at
		FROM agent_runs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var runs []*AgentRun
	err := r.db.SelectContext(ctx, &runs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent runs: %w", err)
	}

	return runs, nil
}

// GetStatsByDate 获取日期统计
func (r *agentRunRepository) GetStatsByDate(ctx context.Context, date time.Time) (map[string]int64, error) {
	query := `
		SELECT
			COUNT(*) as total_runs,
			SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_runs,
			SUM(input_tokens) as total_input_tokens,
			SUM(output_tokens) as total_output_tokens,
			SUM(total_tokens) as total_tokens,
			SUM(latency) as total_latency,
			AVG(latency) as avg_latency,
			SUM(estimated_cost) as total_cost
		FROM agent_runs
		WHERE DATE(created_at) = ?
	`

	var stats struct {
		TotalRuns      int64   `db:"total_runs"`
		SuccessRuns    int64   `db:"success_runs"`
		TotalInput     int64   `db:"total_input_tokens"`
		TotalOutput    int64   `db:"total_output_tokens"`
		TotalTokens    int64   `db:"total_tokens"`
		TotalLatency   int64   `db:"total_latency"`
		AvgLatency     float64 `db:"avg_latency"`
		TotalCost      float64 `db:"total_cost"`
	}

	err := r.db.GetContext(ctx, &stats, query, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by date: %w", err)
	}

	result := map[string]int64{
		"total_runs":       stats.TotalRuns,
		"success_runs":     stats.SuccessRuns,
		"total_input_tokens": stats.TotalInput,
		"total_output_tokens": stats.TotalOutput,
		"total_tokens":     stats.TotalTokens,
		"total_latency_ms":  stats.TotalLatency,
		"total_cost_cents":  int64(stats.TotalCost * 100),
	}

	return result, nil
}
