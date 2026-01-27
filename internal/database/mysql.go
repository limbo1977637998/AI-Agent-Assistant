package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	Charset         string
	ParseTime       bool
	Loc             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime string
}

// DefaultMySQLConfig 返回默认配置
func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Database:        "agent_db",
		User:            "root",
		Password:        "1977637998",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: "5m",
	}
}

// MySQLClient MySQL客户端
type MySQLClient struct {
	DB     *sqlx.DB
	config *MySQLConfig
}

// NewMySQLClient 创建MySQL客户端
func NewMySQLClient(config *MySQLConfig) (*MySQLClient, error) {
	if config == nil {
		config = DefaultMySQLConfig()
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		boolToStr(config.ParseTime),
		config.Loc,
	)

	// 连接数据库
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	// 解析连接最大生命周期
	if config.ConnMaxLifetime != "" {
		lifetime, err := time.ParseDuration(config.ConnMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conn_max_lifetime: %w", err)
		}
		db.SetConnMaxLifetime(lifetime)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	client := &MySQLClient{
		DB:     db,
		config: config,
	}

	return client, nil
}

// Close 关闭数据库连接
func (c *MySQLClient) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// GetDB 获取数据库连接
func (c *MySQLClient) GetDB() *sqlx.DB {
	return c.DB
}

// BeginTxC 开始事务
func (c *MySQLClient) BeginTxC() (*sqlx.Tx, error) {
	return c.DB.Beginx()
}

// Helper function
func boolToStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
