package expert

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-agent-assistant/internal/task"
)

// ResearcherAgent 研究专家Agent
type ResearcherAgent struct {
	*BaseAgent
	searchAPIKey string
	searchEngine string // google, bing, duckduckgo
	maxResults   int
	timeout      time.Duration
}

// NewResearcherAgent 创建研究Agent
func NewResearcherAgent() *ResearcherAgent {
	base := NewBaseAgent(
		"researcher-001",
		"Researcher",
		"researcher",
		"信息收集和研究专家，擅长网络搜索、数据收集和信息整理",
		[]string{
			"web_search",
			"information_collection",
			"data_gathering",
			"fact_checking",
			"source_analysis",
			"literature_review",
		},
	)

	return &ResearcherAgent{
		BaseAgent:    base,
		searchEngine: "duckduckgo", // 默认使用DuckDuckGo（无需API key）
		maxResults:   10,
		timeout:      30 * time.Second,
	}
}

// Execute 执行研究任务
func (r *ResearcherAgent) Execute(ctx context.Context, taskObj *task.Task) (*task.TaskResult, error) {
	startTime := time.Now()
	r.UpdateStatus("running")

	// 验证任务
	if err := r.ValidateTask(taskObj); err != nil {
		return r.createErrorResult(taskObj, err, startTime), err
	}

	// 解析任务目标
	researchGoal := taskObj.Goal

	// 根据任务类型选择研究方法
	var output interface{}
	var err error

	if strings.Contains(researchGoal, "搜索") || strings.Contains(researchGoal, "查找") {
		output, err = r.performWebSearch(ctx, researchGoal, taskObj.Requirements)
	} else if strings.Contains(researchGoal, "分析") || strings.Contains(researchGoal, "研究") {
		output, err = r.performResearch(ctx, researchGoal, taskObj.Requirements)
	} else if strings.Contains(researchGoal, "验证") || strings.Contains(researchGoal, "核查") {
		output, err = r.performFactCheck(ctx, researchGoal, taskObj.Requirements)
	} else {
		// 默认执行搜索
		output, err = r.performWebSearch(ctx, researchGoal, taskObj.Requirements)
	}

	if err != nil {
		r.UpdateStatus("failed")
		return r.createErrorResult(taskObj, err, startTime), err
	}

	r.UpdateStatus("idle")
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusCompleted,
		Output:    output,
		Error:     "",
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type":    "researcher",
			"search_engine": r.searchEngine,
			"result_count":  r.getResultCount(output),
		},
		Timestamp: time.Now(),
		AgentUsed: r.Name,
	}, nil
}

// performWebSearch 执行网络搜索
func (r *ResearcherAgent) performWebSearch(ctx context.Context, query string, requirements interface{}) (interface{}, error) {
	// 构建搜索查询
	searchQuery := r.buildSearchQuery(query, requirements)

	// 执行搜索
	results, err := r.search(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// 整理结果
	summarizedResults := r.summarizeSearchResults(results)

	return map[string]interface{}{
		"query":   searchQuery,
		"results": summarizedResults,
		"count":   len(summarizedResults),
		"source":  r.searchEngine,
	}, nil
}

// performResearch 执行深度研究
func (r *ResearcherAgent) performResearch(ctx context.Context, topic string, requirements interface{}) (interface{}, error) {
	// 多角度搜索
	queries := []string{
		topic + " 概述",
		topic + " 最新研究",
		topic + " 分析报告",
		topic + " 应用案例",
	}

	allResults := make([]map[string]interface{}, 0)
	sources := make([]string, 0)

	for _, query := range queries {
		results, err := r.search(ctx, query)
		if err != nil {
			continue // 某个搜索失败不影响其他搜索
		}

		for _, result := range results {
			allResults = append(allResults, result)
			sources = append(sources, result["url"].(string))
		}
	}

	// 去重和排序
	uniqueResults := r.deduplicateResults(allResults)

	// 生成研究报告
	report := r.generateResearchReport(topic, uniqueResults)

	return map[string]interface{}{
		"topic":      topic,
		"report":     report,
		"sources":    sources,
		"result_count": len(uniqueResults),
	}, nil
}

// performFactCheck 执行事实核查
func (r *ResearcherAgent) performFactCheck(ctx context.Context, claim string, requirements interface{}) (interface{}, error) {
	// 搜索相关证据
	queries := []string{
		claim + " 真假",
		claim + " 验证",
		claim + " 事实核查",
	}

	evidence := make([]map[string]interface{}, 0)

	for _, query := range queries {
		results, err := r.search(ctx, query)
		if err != nil {
			continue
		}
		evidence = append(evidence, results...)
	}

	// 分析证据
	analysis := r.analyzeEvidence(claim, evidence)

	return map[string]interface{}{
		"claim":    claim,
		"analysis": analysis,
		"evidence": evidence,
		"verdict":  r.getVerdict(analysis),
	}, nil
}

// search 执行搜索（简化实现）
func (r *ResearcherAgent) search(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// 根据搜索引擎类型选择实现
	switch r.searchEngine {
	case "duckduckgo":
		return r.searchDuckDuckGo(ctx, query)
	case "google":
		return r.searchGoogle(ctx, query)
	default:
		return r.searchDuckDuckGo(ctx, query)
	}
}

// searchDuckDuckGo 使用DuckDuckGo搜索（无需API key）
func (r *ResearcherAgent) searchDuckDuckGo(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// DuckDuckGo Instant Answer API
	apiURL := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json"

	// 创建HTTP请求
	client := &http.Client{Timeout: r.timeout}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ResearchAgent/1.0)")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 提取相关结果
	results := make([]map[string]interface{}, 0)

	// 添加即时答案
	if abstract, ok := result["Abstract"].(string); ok && abstract != "" {
		results = append(results, map[string]interface{}{
			"title":   result["Heading"].(string),
			"url":     result["AbstractURL"].(string),
			"snippet": abstract,
			"source":  "DuckDuckGo",
		})
	}

	// 添加相关主题
	if relatedTopics, ok := result["RelatedTopics"].([]interface{}); ok {
		count := 0
		for _, topic := range relatedTopics {
			if count >= r.maxResults {
				break
			}

			if topicMap, ok := topic.(map[string]interface{}); ok {
				if text, ok := topicMap["Text"].(string); ok {
					if firstURL, ok := topicMap["FirstURL"].(string); ok {
						results = append(results, map[string]interface{}{
							"title":   r.extractTitle(text),
							"url":     firstURL,
							"snippet": text,
							"source":  "DuckDuckGo",
						})
						count++
					}
				}
			}
		}
	}

	// 如果没有结果，返回模拟数据用于演示
	if len(results) == 0 {
		results = r.getMockResults(query, r.maxResults)
	}

	return results, nil
}

// searchGoogle 使用Google搜索
func (r *ResearcherAgent) searchGoogle(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// Google Custom Search API需要API key
	// 这里简化实现，返回模拟结果
	return r.getMockResults(query, r.maxResults), nil
}

// getMockResults 获取模拟搜索结果（用于演示）
func (r *ResearcherAgent) getMockResults(query string, count int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	for i := 1; i <= count; i++ {
		results = append(results, map[string]interface{}{
			"title":   fmt.Sprintf("搜索结果 #%d - %s", i, query),
			"url":     fmt.Sprintf("https://example.com/result%d?q=%s", i, url.QueryEscape(query)),
			"snippet": fmt.Sprintf("这是关于'%s'的第%d条搜索结果摘要。", query, i),
			"source":  "Mock Search Engine",
		})
	}

	return results
}

// buildSearchQuery 构建搜索查询
func (r *ResearcherAgent) buildSearchQuery(goal string, requirements interface{}) string {
	query := goal

	// 如果有额外要求，添加到查询中
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if keywords, ok := reqMap["keywords"].([]string); ok {
			query += " " + strings.Join(keywords, " ")
		}
		if timeRange, ok := reqMap["time_range"].(string); ok {
			query += " " + timeRange
		}
	}

	return query
}

// summarizeSearchResults 总结搜索结果
func (r *ResearcherAgent) summarizeSearchResults(results []map[string]interface{}) []map[string]interface{} {
	// 返回前N个结果
	if len(results) > r.maxResults {
		return results[:r.maxResults]
	}
	return results
}

// deduplicateResults 去重搜索结果
func (r *ResearcherAgent) deduplicateResults(results []map[string]interface{}) []map[string]interface{} {
	seen := make(map[string]bool)
	unique := make([]map[string]interface{}, 0)

	for _, result := range results {
		url := result["url"].(string)
		if !seen[url] {
			seen[url] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// generateResearchReport 生成研究报告
func (r *ResearcherAgent) generateResearchReport(topic string, results []map[string]interface{}) string {
	report := fmt.Sprintf("# %s 研究报告\n\n", topic)
	report += fmt.Sprintf("## 概述\n\n本次研究共收集到%d条相关信息。\n\n", len(results))
	report += "## 主要发现\n\n"

	for i, result := range results {
		if i >= 5 { // 只列出前5条
			break
		}
		title := result["title"].(string)
		snippet := result["snippet"].(string)
		report += fmt.Sprintf("%d. %s\n   %s\n\n", i+1, title, snippet)
	}

	report += fmt.Sprintf("\n## 信息源\n\n共收集了%d个信息源。", len(results))

	return report
}

// analyzeEvidence 分析证据
func (r *ResearcherAgent) analyzeEvidence(claim string, evidence []map[string]interface{}) string {
	if len(evidence) == 0 {
		return "未找到相关证据"
	}

	supporting := 0
	refuting := 0

	for _, ev := range evidence {
		snippet := strings.ToLower(ev["snippet"].(string))
		if strings.Contains(snippet, "真实") || strings.Contains(snippet, "正确") {
			supporting++
		}
		if strings.Contains(snippet, "虚假") || strings.Contains(snippet, "错误") {
			refuting++
		}
	}

	return fmt.Sprintf("找到%d条证据，其中%d条支持，%条反驳", len(evidence), supporting, refuting)
}

// getVerdict 获取判定结果
func (r *ResearcherAgent) getVerdict(analysis string) string {
	if strings.Contains(analysis, "0条反驳") {
		return "可能为真"
	}
	if strings.Contains(analysis, "0条支持") {
		return "可能为假"
	}
	return "需要更多证据"
}

// extractTitle 从文本中提取标题
func (r *ResearcherAgent) extractTitle(text string) string {
	// 简单提取：取前100个字符
	if len(text) > 100 {
		return text[:100] + "..."
	}
	return text
}

// getResultCount 获取结果数量
func (r *ResearcherAgent) getResultCount(output interface{}) int {
	if outputMap, ok := output.(map[string]interface{}); ok {
		if count, ok := outputMap["count"].(int); ok {
			return count
		}
	}
	return 0
}

// createErrorResult 创建错误结果
func (r *ResearcherAgent) createErrorResult(taskObj *task.Task, err error, startTime time.Time) *task.TaskResult {
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusFailed,
		Output:    nil,
		Error:     err.Error(),
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type": "researcher",
		},
		Timestamp: time.Now(),
		AgentUsed: r.Name,
	}
}

// SetSearchEngine 设置搜索引擎
func (r *ResearcherAgent) SetSearchEngine(engine string) {
	r.searchEngine = engine
}

// SetMaxResults 设置最大结果数
func (r *ResearcherAgent) SetMaxResults(max int) {
	r.maxResults = max
}

// SetSearchAPIKey 设置搜索API密钥
func (r *ResearcherAgent) SetSearchAPIKey(apiKey string) {
	r.searchAPIKey = apiKey
}
