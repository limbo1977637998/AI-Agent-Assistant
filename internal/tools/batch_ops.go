package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	Success    bool                   `json:"success"`              // 操作是否成功
	Message    string                 `json:"message"`              // 结果消息
	Data       interface{}            `json:"data,omitempty"`       // 返回数据
	Error      string                 `json:"error,omitempty"`      // 错误信息
	Statistics map[string]interface{} `json:"statistics,omitempty"` // 统计信息
}

// BatchOpsTool 批量操作工具
// 提供批量HTTP请求、并发控制、批量处理等功能
type BatchOpsTool struct {
	name        string
	description string
	version     string
	httpClient  *http.Client
}

// NewBatchOpsTool 创建批量操作工具实例
func NewBatchOpsTool() *BatchOpsTool {
	return &BatchOpsTool{
		name:        "batch_ops",
		description: "批量操作工具 - 批量HTTP请求、并发控制、批量处理",
		version:     "1.0.0",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Name 返回工具名称
func (t *BatchOpsTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *BatchOpsTool) Description() string {
	return t.description
}

// Version 返回工具版本
func (t *BatchOpsTool) Version() string {
	return t.version
}

// Execute 执行批量操作
// 支持的操作类型：batch_http, batch_process, parallel_execute
func (t *BatchOpsTool) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "batch_http":
		return t.batchHTTPRequests(ctx, params)
	case "batch_process":
		return t.batchProcess(ctx, params)
	case "parallel_execute":
		return t.parallelExecute(ctx, params)
	case "concurrent_limit":
		return t.concurrentLimitProcess(ctx, params)
	default:
		return &BatchOperationResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的操作类型: %s", operation),
		}, nil
	}
}

// batchHTTPRequests 批量HTTP请求
// 参数：
//   - requests: 请求列表（必填）
//     格式：[{"url": "http://example.com", "method": "GET", "headers": {...}, "body": "..."}]
//   - concurrency: 并发数（可选，默认10）
//   - timeout: 超时时间（可选，默认30秒）
func (t *BatchOpsTool) batchHTTPRequests(ctx context.Context, params map[string]interface{}) (*BatchOperationResult, error) {
	requestsParam, ok := params["requests"].([]interface{})
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: requests",
		}, nil
	}

	concurrency := 10
	if c, ok := params["concurrency"].(float64); ok {
		concurrency = int(c)
	}

	timeout := 30
	if to, ok := params["timeout"].(float64); ok {
		timeout = int(to)
	}

	var requests []HTTPRequest
	for _, r := range requestsParam {
		reqMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		request := HTTPRequest{
			URL:    reqMap["url"].(string),
			Method: "GET",
		}

		if method, ok := reqMap["method"].(string); ok {
			request.Method = method
		}

		if headers, ok := reqMap["headers"].(map[string]interface{}); ok {
			request.Headers = make(map[string]string)
			for k, v := range headers {
				request.Headers[k] = fmt.Sprintf("%v", v)
			}
		}

		if body, ok := reqMap["body"].(string); ok {
			request.Body = body
		}

		requests = append(requests, request)
	}

	if len(requests) == 0 {
		return &BatchOperationResult{
			Success: false,
			Error:   "没有有效的请求",
		}, nil
	}

	// 执行批量请求
	results := t.executeRequestsConcurrent(requests, concurrency, time.Duration(timeout)*time.Second)

	// 统计结果
	successCount := 0
	failureCount := 0
	totalTime := time.Duration(0)

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalTime += result.Duration
		} else {
			failureCount++
		}
	}

	avgTime := time.Duration(0)
	if successCount > 0 {
		avgTime = totalTime / time.Duration(successCount)
	}

	return &BatchOperationResult{
		Success: true,
		Message: fmt.Sprintf("批量请求完成：%d 成功，%d 失败", successCount, failureCount),
		Data: map[string]interface{}{
			"results": results,
		},
		Statistics: map[string]interface{}{
			"total":        len(requests),
			"success":      successCount,
			"failed":       failureCount,
			"avg_duration": avgTime.String(),
		},
	}, nil
}

// HTTPRequest HTTP请求定义
type HTTPRequest struct {
	URL     string            `json:"url"`              // 请求URL
	Method  string            `json:"method"`           // HTTP方法
	Headers map[string]string `json:"headers,omitempty"` // 请求头
	Body    string            `json:"body,omitempty"`   // 请求体
}

// HTTPResponse HTTP响应结果
type HTTPResponse struct {
	URL        string        `json:"url"`         // 请求URL
	StatusCode int           `json:"status_code"` // 状态码
	Body       string        `json:"body"`        // 响应体
	Headers    map[string]string `json:"headers"` // 响应头
	Duration   time.Duration `json:"duration"`    // 请求耗时
	Error      error         `json:"error,omitempty"` // 错误信息
}

// executeRequestsConcurrent 并发执行HTTP请求
func (t *BatchOpsTool) executeRequestsConcurrent(requests []HTTPRequest, concurrency int, timeout time.Duration) []HTTPResponse {
	results := make([]HTTPResponse, len(requests))

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, request := range requests {
		wg.Add(1)
		go func(index int, req HTTPRequest) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行请求
			results[index] = t.executeHTTPRequest(req, timeout)
		}(i, request)
	}

	wg.Wait()
	return results
}

// executeHTTPRequest 执行单个HTTP请求
func (t *BatchOpsTool) executeHTTPRequest(request HTTPRequest, timeout time.Duration) HTTPResponse {
	startTime := time.Now()

	// 创建请求
	req, err := http.NewRequest(request.Method, request.URL, nil)
	if err != nil {
		return HTTPResponse{
			URL:      request.URL,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}

	// 设置请求头
	for k, v := range request.Headers {
		req.Header.Set(k, v)
	}

	// 设置请求体
	if request.Body != "" {
		req.Body = io.NopCloser(strings.NewReader(request.Body))
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// 发送请求
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return HTTPResponse{
			URL:      request.URL,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{
			URL:        request.URL,
			StatusCode: resp.StatusCode,
			Error:      err,
			Duration:   time.Since(startTime),
		}
	}

	// 收集响应头
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return HTTPResponse{
		URL:        request.URL,
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Headers:    headers,
		Duration:   time.Since(startTime),
	}
}

// batchProcess 批量处理
// 参数：
//   - items: 待处理的项目列表（必填）
//   - processor: 处理函数名称（必填）
//   - params: 处理函数参数（可选）
//   - concurrency: 并发数（可选，默认5）
func (t *BatchOpsTool) batchProcess(ctx context.Context, params map[string]interface{}) (*BatchOperationResult, error) {
	itemsParam, ok := params["items"].([]interface{})
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: items",
		}, nil
	}

	processor, ok := params["processor"].(string)
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: processor",
		}, nil
	}

	concurrency := 5
	if c, ok := params["concurrency"].(float64); ok {
		concurrency = int(c)
	}

	processorParams := make(map[string]interface{})
	if p, ok := params["params"].(map[string]interface{}); ok {
		processorParams = p
	}

	// 执行批量处理
	results := t.processItemsConcurrent(itemsParam, processor, processorParams, concurrency)

	// 统计结果
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	return &BatchOperationResult{
		Success: true,
		Message: fmt.Sprintf("批量处理完成：%d 成功，%d 失败", successCount, failureCount),
		Data: map[string]interface{}{
			"results": results,
		},
		Statistics: map[string]interface{}{
			"total":   len(itemsParam),
			"success": successCount,
			"failed":  failureCount,
		},
	}, nil
}

// ProcessResult 处理结果
type ProcessResult struct {
	Item    interface{} `json:"item"`             // 原始项目
	Result  interface{} `json:"result,omitempty"` // 处理结果
	Error   error       `json:"error,omitempty"`  // 错误信息
	Index   int         `json:"index"`            // 索引
}

// processItemsConcurrent 并发处理项目
func (t *BatchOpsTool) processItemsConcurrent(items []interface{}, processor string, params map[string]interface{}, concurrency int) []ProcessResult {
	results := make([]ProcessResult, len(items))

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(index int, item interface{}) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行处理
			results[index] = t.processItem(item, processor, params, index)
		}(i, item)
	}

	wg.Wait()
	return results
}

// processItem 处理单个项目
func (t *BatchOpsTool) processItem(item interface{}, processor string, params map[string]interface{}, index int) ProcessResult {
	// 根据处理器类型执行不同的处理逻辑
	switch processor {
	case "uppercase":
		if str, ok := item.(string); ok {
			return ProcessResult{
				Item:   item,
				Result: strings.ToUpper(str),
				Index:  index,
			}
		}
	case "lowercase":
		if str, ok := item.(string); ok {
			return ProcessResult{
				Item:   item,
				Result: strings.ToLower(str),
				Index:  index,
			}
		}
	case "reverse":
		if str, ok := item.(string); ok {
			return ProcessResult{
				Item:   item,
				Result: reverseString(str),
				Index:  index,
			}
		}
	case "double":
		if num, ok := item.(float64); ok {
			return ProcessResult{
				Item:   item,
				Result: num * 2,
				Index:  index,
			}
		}
	case "square":
		if num, ok := item.(float64); ok {
			return ProcessResult{
				Item:   item,
				Result: num * num,
				Index:  index,
			}
		}
	}

	return ProcessResult{
		Item:   item,
		Result: item,
		Index:  index,
	}
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// parallelExecute 并行执行多个任务
// 参数：
//   - tasks: 任务列表（必填）
//     格式：[{"name": "task1", "operation": "...", "params": {...}}]
//   - stop_on_error: 遇到错误是否停止（可选，默认false）
func (t *BatchOpsTool) parallelExecute(ctx context.Context, params map[string]interface{}) (*BatchOperationResult, error) {
	tasksParam, ok := params["tasks"].([]interface{})
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: tasks",
		}, nil
	}

	stopOnError := false
	if soe, ok := params["stop_on_error"].(bool); ok {
		stopOnError = soe
	}

	var tasks []Task
	for _, task := range tasksParam {
		taskMap, ok := task.(map[string]interface{})
		if !ok {
			continue
		}

		t := Task{
			Name:      taskMap["name"].(string),
			Operation: taskMap["operation"].(string),
		}

		if p, ok := taskMap["params"].(map[string]interface{}); ok {
			t.Params = p
		}

		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		return &BatchOperationResult{
			Success: false,
			Error:   "没有有效的任务",
		}, nil
	}

	// 并行执行任务
	results := t.executeTasksParallel(tasks, stopOnError)

	// 统计结果
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	return &BatchOperationResult{
		Success: true,
		Message: fmt.Sprintf("并行执行完成：%d 成功，%d 失败", successCount, failureCount),
		Data: map[string]interface{}{
			"results": results,
		},
		Statistics: map[string]interface{}{
			"total":   len(tasks),
			"success": successCount,
			"failed":  failureCount,
		},
	}, nil
}

// Task 任务定义
type Task struct {
	Name      string                 `json:"name"`                // 任务名称
	Operation string                 `json:"operation"`           // 操作类型
	Params    map[string]interface{} `json:"params,omitempty"`    // 操作参数
}

// TaskResult 任务执行结果
type TaskResult struct {
	Name    string      `json:"name"`               // 任务名称
	Result  interface{} `json:"result,omitempty"`   // 执行结果
	Error   error       `json:"error,omitempty"`    // 错误信息
	Elapsed time.Duration `json:"elapsed"`          // 执行耗时
}

// executeTasksParallel 并行执行任务
func (t *BatchOpsTool) executeTasksParallel(tasks []Task, stopOnError bool) []TaskResult {
	results := make([]TaskResult, len(tasks))
	var wg sync.WaitGroup
	var errorCount int32

	for i, task := range tasks {
		wg.Add(1)
		go func(index int, task Task) {
			defer wg.Done()

			// 检查是否需要停止
			if stopOnError && atomic.LoadInt32(&errorCount) > 0 {
				results[index] = TaskResult{
					Name:  task.Name,
					Error: fmt.Errorf("因其他任务失败而跳过"),
				}
				return
			}

			// 执行任务
			startTime := time.Now()
			result, err := t.executeTask(task)
			elapsed := time.Since(startTime)

			results[index] = TaskResult{
				Name:    task.Name,
				Result:  result,
				Error:   err,
				Elapsed: elapsed,
			}

			// 如果出错且设置了停止标志
			if err != nil && stopOnError {
				atomic.AddInt32(&errorCount, 1)
			}
		}(i, task)
	}

	wg.Wait()
	return results
}

// executeTask 执行单个任务
func (t *BatchOpsTool) executeTask(task Task) (interface{}, error) {
	// 这里可以调用其他工具来执行任务
	// 简化实现：返回任务名称和参数
	return map[string]interface{}{
		"task":      task.Name,
		"operation": task.Operation,
		"params":    task.Params,
		"status":    "completed",
	}, nil
}

// concurrentLimitProcess 并发限制的批量处理
// 参数：
//   - items: 待处理的项目列表（必填）
//   - handler: 处理函数（必填，从预定义的处理器中选择）
//   - max_concurrency: 最大并发数（必填）
//   - rate_limit: 速率限制-每秒请求数（可选）
func (t *BatchOpsTool) concurrentLimitProcess(ctx context.Context, params map[string]interface{}) (*BatchOperationResult, error) {
	itemsParam, ok := params["items"].([]interface{})
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: items",
		}, nil
	}

	handler, ok := params["handler"].(string)
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: handler",
		}, nil
	}

	maxConcurrency, ok := params["max_concurrency"].(float64)
	if !ok {
		return &BatchOperationResult{
			Success: false,
			Error:   "缺少必填参数: max_concurrency",
		}, nil
	}

	rateLimit := float64(0)
	if rl, ok := params["rate_limit"].(float64); ok {
		rateLimit = rl
	}

	// 执行并发限制处理
	results := t.processWithRateLimit(itemsParam, handler, int(maxConcurrency), rateLimit)

	// 统计结果
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	return &BatchOperationResult{
		Success: true,
		Message: fmt.Sprintf("并发限制处理完成：%d 成功，%d 失败", successCount, failureCount),
		Data: map[string]interface{}{
			"results": results,
		},
		Statistics: map[string]interface{}{
			"total":           len(itemsParam),
			"success":         successCount,
			"failed":          failureCount,
			"max_concurrency": int(maxConcurrency),
			"rate_limit":      rateLimit,
		},
	}, nil
}

// processWithRateLimit 带速率限制的处理
func (t *BatchOpsTool) processWithRateLimit(items []interface{}, handler string, maxConcurrency int, rateLimit float64) []ProcessResult {
	results := make([]ProcessResult, len(items))

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// 速率限制器
	var rateLimiter *time.Ticker
	if rateLimit > 0 {
		interval := time.Duration(float64(time.Second) / rateLimit)
		rateLimiter = time.NewTicker(interval)
		defer rateLimiter.Stop()
	}

	for i, item := range items {
		wg.Add(1)

		// 如果有速率限制，先等待
		if rateLimiter != nil {
			<-rateLimiter.C
		}

		go func(index int, item interface{}) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行处理
			results[index] = t.processItem(item, handler, nil, index)
		}(i, item)
	}

	wg.Wait()
	return results
}

// BatchDownloadResult 批量下载结果
type BatchDownloadResult struct {
	Success    bool                   `json:"success"`              // 是否成功
	Message    string                 `json:"message"`              // 结果消息
	Data       interface{}            `json:"data,omitempty"`       // 返回数据
	Statistics map[string]interface{} `json:"statistics,omitempty"` // 统计信息
	Error      string                 `json:"error,omitempty"`      // 错误信息
}

// BatchDownload 批量下载文件
// 参数：
//   - urls: URL列表（必填）
//   - output_dir: 输出目录（必填）
//   - concurrency: 并发数（可选，默认5）
func (t *BatchOpsTool) BatchDownload(ctx context.Context, urls []string, outputDir string, concurrency int) (*BatchDownloadResult, error) {
	if len(urls) == 0 {
		return &BatchDownloadResult{
			Success: false,
			Error:   "URL列表为空",
		}, nil
	}

	if concurrency <= 0 {
		concurrency = 5
	}

	// 执行批量下载
	results := t.downloadFilesConcurrent(urls, outputDir, concurrency)

	// 统计结果
	successCount := 0
	failureCount := 0
	totalSize := int64(0)

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalSize += result.Size
		} else {
			failureCount++
		}
	}

	return &BatchDownloadResult{
		Success: true,
		Message: fmt.Sprintf("批量下载完成：%d 成功，%d 失败，总大小：%d 字节", successCount, failureCount, totalSize),
		Data: map[string]interface{}{
			"results": results,
		},
		Statistics: map[string]interface{}{
			"total":       len(urls),
			"success":     successCount,
			"failed":      failureCount,
			"total_size":  totalSize,
		},
	}, nil
}

// DownloadResult 下载结果
type DownloadResult struct {
	URL      string `json:"url"`                // 文件URL
	Path     string `json:"path"`               // 保存路径
	Size     int64  `json:"size"`               // 文件大小
	Duration time.Duration `json:"duration"`     // 下载耗时
	Error    error  `json:"error,omitempty"`    // 错误信息
}

// downloadFilesConcurrent 并发下载文件
func (t *BatchOpsTool) downloadFilesConcurrent(urls []string, outputDir string, concurrency int) []DownloadResult {
	results := make([]DownloadResult, len(urls))

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行下载
			results[index] = t.downloadFile(url, outputDir)
		}(i, url)
	}

	wg.Wait()
	return results
}

// downloadFile 下载单个文件
func (t *BatchOpsTool) downloadFile(url, outputDir string) DownloadResult {
	startTime := time.Now()

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return DownloadResult{
			URL:      url,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}

	// 发送请求
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return DownloadResult{
			URL:      url,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return DownloadResult{
			URL:      url,
			Error:    fmt.Errorf("HTTP状态码: %d", resp.StatusCode),
			Duration: time.Since(startTime),
		}
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DownloadResult{
			URL:      url,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}

	// 生成文件名
	filename := extractFilename(url)
	filePath := fmt.Sprintf("%s/%s", outputDir, filename)

	// 保存文件
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return DownloadResult{
			URL:      url,
			Error:    err,
			Duration: time.Since(startTime),
		}
	}

	return DownloadResult{
		URL:      url,
		Path:     filePath,
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
	}
}

// extractFilename 从URL中提取文件名
func extractFilename(url string) string {
	// 移除查询参数
	if idx := strings.Index(url, "?"); idx > 0 {
		url = url[:idx]
	}

	// 提取文件名
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// 如果文件名为空，使用默认名称
	if filename == "" {
		filename = fmt.Sprintf("download_%d", time.Now().Unix())
	}

	return filename
}
