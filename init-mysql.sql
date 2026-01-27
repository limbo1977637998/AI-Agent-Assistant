-- AI Agent Assistant 数据库初始化脚本
-- 使用方法: mysql -uroot -p1977637998 < init-mysql.sql

-- 1. 创建数据库
CREATE DATABASE IF NOT EXISTS agent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE agent_db;

-- 2. 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL COMMENT '会话ID',
    user_id VARCHAR(255) DEFAULT '' COMMENT '用户ID',
    model VARCHAR(50) DEFAULT 'glm' COMMENT '使用的模型',
    metadata JSON COMMENT '会话元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

-- 3. 消息历史表
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL COMMENT '会话ID',
    role VARCHAR(20) NOT NULL COMMENT '角色: user/assistant/system/tool',
    content TEXT NOT NULL COMMENT '消息内容',
    tool_calls JSON COMMENT '工具调用信息',
    tokens_used INT DEFAULT 0 COMMENT 'Token使用量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_session_id (session_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息历史表';

-- 4. 用户记忆表
CREATE TABLE IF NOT EXISTS user_memories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL COMMENT '用户ID',
    memory TEXT NOT NULL COMMENT '记忆内容',
    topics VARCHAR(500) DEFAULT '' COMMENT '主题标签，逗号分隔',
    importance FLOAT DEFAULT 1.0 COMMENT '重要性评分 0-1',
    memory_type VARCHAR(50) DEFAULT 'preference' COMMENT '记忆类型: preference/background/history',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_importance (importance),
    INDEX idx_memory_type (memory_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户记忆表';

-- 5. 工具调用记录表
CREATE TABLE IF NOT EXISTS tool_calls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) COMMENT '会话ID',
    user_id VARCHAR(255) COMMENT '用户ID',
    tool_name VARCHAR(100) NOT NULL COMMENT '工具名称',
    arguments JSON COMMENT '工具参数',
    result TEXT COMMENT '工具执行结果',
    success BOOLEAN DEFAULT TRUE COMMENT '是否成功',
    error_msg TEXT COMMENT '错误信息',
    duration INT DEFAULT 0 COMMENT '执行时长(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_session_id (session_id),
    INDEX idx_tool_name (tool_name),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具调用记录表';

-- 6. Agent运行记录表
CREATE TABLE IF NOT EXISTS agent_runs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    run_id VARCHAR(255) UNIQUE NOT NULL COMMENT '运行ID',
    session_id VARCHAR(255) COMMENT '会话ID',
    user_id VARCHAR(255) COMMENT '用户ID',
    input TEXT COMMENT '用户输入',
    output TEXT COMMENT 'Agent输出',
    model VARCHAR(50) COMMENT '使用的模型',
    input_tokens INT DEFAULT 0 COMMENT '输入Token数',
    output_tokens INT DEFAULT 0 COMMENT '输出Token数',
    total_tokens INT DEFAULT 0 COMMENT '总Token数',
    estimated_cost FLOAT DEFAULT 0 COMMENT '预估成本(元)',
    latency INT DEFAULT 0 COMMENT '响应时长(毫秒)',
    success BOOLEAN DEFAULT TRUE COMMENT '是否成功',
    error_msg TEXT COMMENT '错误信息',
    rag_used BOOLEAN DEFAULT FALSE COMMENT '是否使用RAG',
    tools_used JSON COMMENT '使用的工具列表',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_run_id (run_id),
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent运行记录表';

-- 7. 知识库表
CREATE TABLE IF NOT EXISTS knowledge_base (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    content_hash VARCHAR(64) UNIQUE NOT NULL COMMENT '内容哈希(SHA256)',
    source VARCHAR(500) DEFAULT '' COMMENT '来源',
    content_type VARCHAR(50) DEFAULT 'text' COMMENT '内容类型: text/file/url',
    chunk_count INT DEFAULT 0 COMMENT '分块数量',
    total_chars INT DEFAULT 0 COMMENT '总字符数',
    metadata JSON COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_content_hash (content_hash),
    INDEX idx_source (source)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='知识库表';

-- 8. 知识分块表
CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    knowledge_id BIGINT NOT NULL COMMENT '知识库ID',
    chunk_index INT NOT NULL COMMENT '分块索引',
    content TEXT NOT NULL COMMENT '分块内容',
    vector_id VARCHAR(255) COMMENT 'Milvus向量ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_knowledge_id (knowledge_id),
    INDEX idx_chunk_index (chunk_index),
    FOREIGN KEY (knowledge_id) REFERENCES knowledge_base(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='知识分块表';

-- 9. 向量集合配置表
CREATE TABLE IF NOT EXISTS vector_collections (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    collection_name VARCHAR(100) UNIQUE NOT NULL COMMENT '集合名称',
    dimension INT DEFAULT 1024 COMMENT '向量维度',
    index_type VARCHAR(50) DEFAULT 'HNSW' COMMENT '索引类型',
    metric_type VARCHAR(50) DEFAULT 'COSINE' COMMENT '距离度量',
    description VARCHAR(500) DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_collection_name (collection_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='向量集合配置表';

-- 10. 系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    description VARCHAR(500) DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 插入默认配置
INSERT INTO system_config (config_key, config_value, description) VALUES
('milvus_collection_name', 'agent_knowledge', 'Milvus知识库集合名称'),
('embedding_model', 'embedding-2', 'Embedding模型名称'),
('embedding_dimension', '1024', 'Embedding向量维度'),
('rag_top_k', '3', 'RAG检索TopK数量'),
('rag_threshold', '0.3', 'RAG检索相似度阈值')
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- 创建测试数据（可选）
-- INSERT INTO sessions (session_id, user_id, model) VALUES ('test_session', 'test_user', 'glm');
-- INSERT INTO messages (session_id, role, content) VALUES ('test_session', 'user', '你好');

-- 显示表信息
SELECT 'Database initialized successfully!' AS status;
SHOW TABLES;
SELECT COUNT(*) AS table_count FROM information_schema.tables WHERE table_schema = 'agent_db';
