-- AI Agent Assistant v0.4 Database Schema
-- 创建时间: 2026-01-28

-- 如果不存在则创建数据库
CREATE DATABASE IF NOT EXISTS agent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE agent_db;

-- ============================================
-- 会话和消息表
-- ============================================

-- 会话表
DROP TABLE IF EXISTS sessions;
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY COMMENT '会话ID',
    model VARCHAR(50) NOT NULL COMMENT '使用的模型',
    summary TEXT COMMENT '会话摘要',
    state JSON COMMENT '会话状态（版本控制）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

-- 消息表
DROP TABLE IF EXISTS messages;
CREATE TABLE messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '消息ID',
    session_id VARCHAR(255) NOT NULL COMMENT '会话ID',
    role VARCHAR(20) NOT NULL COMMENT '角色（user/assistant/system）',
    content TEXT NOT NULL COMMENT '消息内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_session_id (session_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

-- ============================================
-- 用户记忆表
-- ============================================

DROP TABLE IF EXISTS user_memories;
CREATE TABLE user_memories (
    id VARCHAR(255) PRIMARY KEY COMMENT '记忆ID',
    user_id VARCHAR(255) NOT NULL COMMENT '用户ID',
    content TEXT NOT NULL COMMENT '记忆内容',
    topics JSON COMMENT '主题标签',
    importance DECIMAL(3,2) DEFAULT 0.5 COMMENT '重要性评分',
    access_count INT DEFAULT 0 COMMENT '访问次数',
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后访问时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_importance (importance DESC),
    INDEX idx_last_accessed (last_accessed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户记忆表';

-- ============================================
-- 知识库表
-- ============================================

DROP TABLE IF EXISTS knowledge_chunks;
CREATE TABLE knowledge_chunks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '知识块ID',
    content TEXT NOT NULL COMMENT '知识块内容',
    source VARCHAR(255) COMMENT '来源',
    chunk_index INT COMMENT '分块索引',
    embedding JSON COMMENT '向量embedding',
    metadata JSON COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_source (source),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='知识块表';

-- ============================================
-- 工具调用记录表
-- ============================================

DROP TABLE IF EXISTS tool_calls;
CREATE TABLE tool_calls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '工具调用ID',
    session_id VARCHAR(255) NOT NULL COMMENT '会话ID',
    tool_name VARCHAR(100) NOT NULL COMMENT '工具名称',
    parameters JSON COMMENT '调用参数',
    result TEXT COMMENT '调用结果',
    status VARCHAR(20) COMMENT '状态（success/failed）',
    error_message TEXT COMMENT '错误信息',
    duration_ms INT COMMENT '执行时长（毫秒）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_session_id (session_id),
    INDEX idx_tool_name (tool_name),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具调用记录表';

-- ============================================
-- Agent运行记录表
-- ============================================

DROP TABLE IF EXISTS agent_runs;
CREATE TABLE agent_runs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '运行ID',
    session_id VARCHAR(255) COMMENT '会话ID',
    task_description TEXT COMMENT '任务描述',
    model VARCHAR(50) COMMENT '使用的模型',
    tools_used JSON COMMENT '使用的工具',
    reasoning TEXT COMMENT '推理过程',
    result TEXT COMMENT '最终结果',
    status VARCHAR(20) COMMENT '状态',
    duration_ms INT COMMENT '执行时长',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_session_id (session_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent运行记录表';

-- ============================================
-- 评估结果表
-- ============================================

DROP TABLE IF EXISTS evaluation_results;
CREATE TABLE evaluation_results (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '评估ID',
    evaluator_name VARCHAR(100) NOT NULL COMMENT '评估器名称',
    test_case_id VARCHAR(255) COMMENT '测试用例ID',
    input TEXT COMMENT '输入',
    expected TEXT COMMENT '期望输出',
    actual TEXT COMMENT '实际输出',
    passed BOOLEAN COMMENT '是否通过',
    score DECIMAL(5,2) COMMENT '得分',
    metrics JSON COMMENT '指标',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_evaluator_name (evaluator_name),
    INDEX idx_passed (passed),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评估结果表';

-- ============================================
-- 索引优化
-- ============================================

-- 为常用查询创建复合索引
CREATE INDEX idx_sessions_updated ON sessions(updated_at DESC);
CREATE INDEX idx_messages_session_created ON messages(session_id, created_at);
CREATE INDEX idx_user_memories_user_importance ON user_memories(user_id, importance DESC);
CREATE INDEX idx_tool_calls_session_created ON tool_calls(session_id, created_at);

-- ============================================
-- 初始化数据（可选）
-- ============================================

-- 插入默认会话示例（可选）
-- INSERT INTO sessions (id, model, summary) VALUES
-- ('demo-session', 'glm', '示例会话');

-- ============================================
-- 完成提示
-- ============================================

SELECT 'Database schema created successfully!' AS Status;
SELECT COUNT(*) AS total_tables FROM information_schema.tables WHERE table_schema = 'agent_db';
