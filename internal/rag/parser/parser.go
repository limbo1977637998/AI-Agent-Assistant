package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Parser 文档解析器接口
type Parser interface {
	Parse(filePath string) (string, error)
}

// DocumentParser 文档解析器
type DocumentParser struct{}

// NewParser 创建文档解析器
func NewParser() *DocumentParser {
	return &DocumentParser{}
}

// Parse 解析文档
func (p *DocumentParser) Parse(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// 根据文件扩展名选择解析方式
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".txt", ".md":
		return p.parseTextFile(filePath)
	case ".pdf":
		return p.parsePDF(filePath)
	case ".json", ".yaml", ".yml", ".xml", ".html", ".htm":
		return p.parseTextFile(filePath)
	default:
		// 默认尝试作为文本文件读取
		return p.parseTextFile(filePath)
	}
}

// parseTextFile 解析文本文件
func (p *DocumentParser) parseTextFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// parsePDF 解析PDF文件（简化实现）
func (p *DocumentParser) parsePDF(filePath string) (string, error) {
	// 简化实现：返回提示信息
	// 生产环境应使用专门的PDF解析库如: github.com/pdfcpu/pdfcpu
	return "", fmt.Errorf("PDF parsing not implemented yet. Please convert to text format or use txt/md files")
}

// ParseFromBytes 从字节数组解析文档
func (p *DocumentParser) ParseFromBytes(data []byte, fileType string) (string, error) {
	switch fileType {
	case "txt", "md", "json", "yaml", "yml", "xml":
		return string(data), nil
	default:
		return string(data), nil
	}
}

// ParseFromString 从字符串解析文档
func (p *DocumentParser) ParseFromString(content string) string {
	return strings.TrimSpace(content)
}
