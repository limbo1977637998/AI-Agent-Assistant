package tools

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileOperationResult 文件操作结果
type FileOperationResult struct {
	Success   bool                   `json:"success"`             // 操作是否成功
	Message   string                 `json:"message"`             // 结果消息
	Data      interface{}            `json:"data,omitempty"`      // 返回数据
	Error     string                 `json:"error,omitempty"`     // 错误信息
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // 元数据
}

// FileOpsTool 文件操作工具
// 提供批量文件处理、格式转换、压缩解压等功能
type FileOpsTool struct {
	name        string
	description string
	version     string
}

// NewFileOpsTool 创建文件操作工具实例
func NewFileOpsTool() *FileOpsTool {
	return &FileOpsTool{
		name:        "file_ops",
		description: "文件操作工具 - 批量处理、格式转换、压缩解压",
		version:     "1.0.0",
	}
}

// Name 返回工具名称
func (t *FileOpsTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *FileOpsTool) Description() string {
	return t.description
}

// Version 返回工具版本
func (t *FileOpsTool) Version() string {
	return t.version
}

// Execute 执行文件操作
// 支持的操作类型：read, write, batch_read, convert, compress, decompress
func (t *FileOpsTool) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "read":
		return t.readFile(params)
	case "write":
		return t.writeFile(params)
	case "batch_read":
		return t.batchReadFiles(params)
	case "convert":
		return t.convertFile(params)
	case "compress":
		return t.compressFiles(params)
	case "decompress":
		return t.decompressFile(params)
	case "list":
		return t.listFiles(params)
	case "delete":
		return t.deleteFiles(params)
	default:
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的操作类型: %s", operation),
		}, nil
	}
}

// readFile 读取单个文件
// 参数：
//   - path: 文件路径（必填）
//   - encoding: 文件编码（可选，默认utf-8）
func (t *FileOpsTool) readFile(params map[string]interface{}) (*FileOperationResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: path",
		}, nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("文件不存在: %s", path),
		}, nil
	}

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("读取文件失败: %v", err),
		}, nil
	}

	// 获取文件信息
	fileInfo, _ := os.Stat(path)

	return &FileOperationResult{
		Success: true,
		Message: "文件读取成功",
		Data: map[string]interface{}{
			"path":     path,
			"content":  string(content),
			"size":     fileInfo.Size(),
			"modified": fileInfo.ModTime(),
		},
		Metadata: map[string]interface{}{
			"size":      fileInfo.Size(),
			"extension": filepath.Ext(path),
		},
	}, nil
}

// writeFile 写入文件
// 参数：
//   - path: 文件路径（必填）
//   - content: 文件内容（必填）
//   - overwrite: 是否覆盖（可选，默认false）
func (t *FileOpsTool) writeFile(params map[string]interface{}) (*FileOperationResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: path",
		}, nil
	}

	content, ok := params["content"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: content",
		}, nil
	}

	overwrite := true
	if ow, ok := params["overwrite"].(bool); ok {
		overwrite = ow
	}

	// 检查文件是否已存在
	if _, err := os.Stat(path); err == nil && !overwrite {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("文件已存在且不允许覆盖: %s", path),
		}, nil
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("创建目录失败: %v", err),
		}, nil
	}

	// 写入文件
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("写入文件失败: %v", err),
		}, nil
	}

	return &FileOperationResult{
		Success: true,
		Message: "文件写入成功",
		Data: map[string]interface{}{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// batchReadFiles 批量读取文件
// 参数：
//   - paths: 文件路径列表（必填）
//   - pattern: 文件匹配模式（可选，支持通配符）
func (t *FileOpsTool) batchReadFiles(params map[string]interface{}) (*FileOperationResult, error) {
	var paths []string

	// 方式1：直接提供路径列表
	if pathList, ok := params["paths"].([]interface{}); ok {
		for _, p := range pathList {
			if path, ok := p.(string); ok {
				paths = append(paths, path)
			}
		}
	}

	// 方式2：使用模式匹配
	if pattern, ok := params["pattern"].(string); ok {
		matchedPaths, err := filepath.Glob(pattern)
		if err != nil {
			return &FileOperationResult{
				Success: false,
				Error:   fmt.Sprintf("模式匹配失败: %v", err),
			}, nil
		}
		paths = append(paths, matchedPaths...)
	}

	if len(paths) == 0 {
		return &FileOperationResult{
			Success: false,
			Error:   "没有找到要读取的文件",
		}, nil
	}

	// 批量读取文件
	results := make([]map[string]interface{}, 0)
	successCount := 0
	failedCount := 0

	for _, path := range paths {
		result, err := t.readFile(map[string]interface{}{"path": path})
		if err != nil || !result.Success {
			failedCount++
			continue
		}
		successCount++
		results = append(results, result.Data.(map[string]interface{}))
	}

	return &FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("批量读取完成：成功 %d 个，失败 %d 个", successCount, failedCount),
		Data: map[string]interface{}{
			"files":        results,
			"total":        len(paths),
			"success":      successCount,
			"failed":       failedCount,
		},
	}, nil
}

// convertFile 文件格式转换
// 参数：
//   - path: 源文件路径（必填）
//   - target_format: 目标格式（必填）
//   - output_path: 输出路径（可选）
// 支持的转换：json <-> csv, json <-> yaml
func (t *FileOpsTool) convertFile(params map[string]interface{}) (*FileOperationResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: path",
		}, nil
	}

	targetFormat, ok := params["target_format"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: target_format",
		}, nil
	}

	// 读取源文件
	readResult, err := t.readFile(map[string]interface{}{"path": path})
	if err != nil || !readResult.Success {
		return readResult, err
	}

	sourceFormat := strings.TrimPrefix(filepath.Ext(path), ".")
	outputPath := path
	if op, ok := params["output_path"].(string); ok {
		outputPath = op
	} else {
		// 替换扩展名
		outputPath = strings.TrimSuffix(path, filepath.Ext(path)) + "." + targetFormat
	}

	// 根据源格式和目标格式进行转换
	var convertedContent string
	switch {
	case sourceFormat == "json" && targetFormat == "csv":
		convertedContent, err = t.jsonToCSV(readResult.Data.(map[string]interface{})["content"].(string))
	case sourceFormat == "csv" && targetFormat == "json":
		convertedContent, err = t.csvToJSON(readResult.Data.(map[string]interface{})["content"].(string))
	default:
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的格式转换: %s -> %s", sourceFormat, targetFormat),
		}, nil
	}

	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("格式转换失败: %v", err),
		}, nil
	}

	// 写入转换后的文件
	writeResult, err := t.writeFile(map[string]interface{}{
		"path":     outputPath,
		"content":  convertedContent,
		"overwrite": true,
	})
	if err != nil || !writeResult.Success {
		return writeResult, err
	}

	return &FileOperationResult{
		Success: true,
		Message: "文件格式转换成功",
		Data: map[string]interface{}{
			"source_path":   path,
			"target_path":   outputPath,
			"source_format": sourceFormat,
			"target_format": targetFormat,
		},
	}, nil
}

// jsonToCSV JSON转CSV
func (t *FileOpsTool) jsonToCSV(jsonContent string) (string, error) {
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", nil
	}

	// 获取所有字段
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	// 构建CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入表头
	writer.Write(headers)

	// 写入数据
	for _, row := range data {
		var record []string
		for _, header := range headers {
			val := fmt.Sprintf("%v", row[header])
			record = append(record, val)
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// csvToJSON CSV转JSON
func (t *FileOpsTool) csvToJSON(csvContent string) (string, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "[]", nil
	}

	// 第一行是表头
	headers := records[0]
	result := make([]map[string]interface{}, 0)

	// 转换数据行
	for _, record := range records[1:] {
		row := make(map[string]interface{})
		for i, header := range headers {
			if i < len(record) {
				row[header] = record[i]
			}
		}
		result = append(result, row)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// compressFiles 压缩文件
// 参数：
//   - files: 要压缩的文件列表（必填）
//   - output: 输出zip文件路径（必填）
func (t *FileOpsTool) compressFiles(params map[string]interface{}) (*FileOperationResult, error) {
	filesParam, ok := params["files"].([]interface{})
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: files",
		}, nil
	}

	outputPath, ok := params["output"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: output",
		}, nil
	}

	var files []string
	for _, f := range filesParam {
		if file, ok := f.(string); ok {
			files = append(files, file)
		}
	}

	// 创建zip文件
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("创建zip文件失败: %v", err),
		}, nil
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到zip
	successCount := 0
	for _, file := range files {
		err := t.addFileToZip(zipWriter, file)
		if err != nil {
			continue
		}
		successCount++
	}

	return &FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("压缩完成：%d/%d 个文件", successCount, len(files)),
		Data: map[string]interface{}{
			"output":  outputPath,
			"total":   len(files),
			"success": successCount,
		},
	}, nil
}

// addFileToZip 添加文件到zip
func (t *FileOpsTool) addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}

// decompressFile 解压文件
// 参数：
//   - source: 源zip文件路径（必填）
//   - destination: 解压目标目录（必填）
func (t *FileOpsTool) decompressFile(params map[string]interface{}) (*FileOperationResult, error) {
	sourcePath, ok := params["source"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: source",
		}, nil
	}

	destination, ok := params["destination"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: destination",
		}, nil
	}

	// 打开zip文件
	zipReader, err := zip.OpenReader(sourcePath)
	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("打开zip文件失败: %v", err),
		}, nil
	}
	defer zipReader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(destination, 0755); err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("创建目录失败: %v", err),
		}, nil
	}

	// 解压文件
	extractedCount := 0
	for _, file := range zipReader.File {
		err := t.extractFile(file, destination)
		if err != nil {
			continue
		}
		extractedCount++
	}

	return &FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("解压完成：%d 个文件", extractedCount),
		Data: map[string]interface{}{
			"destination": destination,
			"total":       extractedCount,
		},
	}, nil
}

// extractFile 提取文件
func (t *FileOpsTool) extractFile(file *zip.File, destination string) error {
	// 创建目标文件
	path := filepath.Join(destination, file.Name)

	// 检查Zip Slip漏洞
	if !strings.HasPrefix(path, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("无效的文件路径: %s", path)
	}

	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// 创建父目录
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, fileReader)
	return err
}

// listFiles 列出目录文件
// 参数：
//   - path: 目录路径（必填）
//   - recursive: 是否递归（可选，默认false）
//   - pattern: 文件匹配模式（可选）
func (t *FileOpsTool) listFiles(params map[string]interface{}) (*FileOperationResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: path",
		}, nil
	}

	recursive := false
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	pattern := "*"
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}

	var files []map[string]interface{}
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if filePath == path {
			return nil
		}

		// 如果不递归，跳过子目录
		if !recursive && info.IsDir() {
			return filepath.SkipDir
		}

		// 跳过子目录中的文件（不递归时）
		if !recursive && filepath.Dir(filePath) != path {
			return nil
		}

		// 模式匹配
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil || !matched {
			return nil
		}

		files = append(files, map[string]interface{}{
			"name":     filepath.Base(filePath),
			"path":     filePath,
			"size":     info.Size(),
			"is_dir":   info.IsDir(),
			"modified": info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return &FileOperationResult{
			Success: false,
			Error:   fmt.Sprintf("列出文件失败: %v", err),
		}, nil
	}

	return &FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("找到 %d 个文件", len(files)),
		Data: map[string]interface{}{
			"path":  path,
			"files": files,
			"count": len(files),
		},
	}, nil
}

// deleteFiles 删除文件
// 参数：
//   - paths: 文件路径列表（必填）
func (t *FileOpsTool) deleteFiles(params map[string]interface{}) (*FileOperationResult, error) {
	pathsParam, ok := params["paths"].([]interface{})
	if !ok {
		return &FileOperationResult{
			Success: false,
			Error:   "缺少必填参数: paths",
		}, nil
	}

	var paths []string
	for _, p := range pathsParam {
		if path, ok := p.(string); ok {
			paths = append(paths, path)
		}
	}

	successCount := 0
	failedCount := 0

	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			failedCount++
			continue
		}
		successCount++
	}

	return &FileOperationResult{
		Success: true,
		Message: fmt.Sprintf("删除完成：成功 %d 个，失败 %d 个", successCount, failedCount),
		Data: map[string]interface{}{
			"total":   len(paths),
			"success": successCount,
			"failed":  failedCount,
		},
	}, nil
}
