package expert

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"ai-agent-assistant/internal/task"
)

// WriterAgent 写作专家Agent
type WriterAgent struct {
	*BaseAgent
	writingStyles []string
	maxLength     int
	templates     map[string]string
}

// NewWriterAgent 创建写作Agent
func NewWriterAgent() *WriterAgent {
	base := NewBaseAgent(
		"writer-001",
		"Writer",
		"writer",
		"内容创作专家，擅长文案撰写、报告生成和内容优化",
		[]string{
			"content_generation",
			"article_writing",
			"report_writing",
			"copywriting",
			"content_editing",
			"summarization",
			"translation",
			"proofreading",
		},
	)

	templates := map[string]string{
		"article": `# %s

## 概述
%s

## 详细内容
%s

## 结论
%s
`,
		"report": `# %s 报告

## 执行摘要
%s

## 分析
%s

## 建议
%s
`,
		"summary": `# %s - 摘要

%s
`,
	}

	return &WriterAgent{
		BaseAgent:     base,
		writingStyles: []string{"formal", "casual", "professional", "creative", "academic"},
		maxLength:     5000,
		templates:     templates,
	}
}

// Execute 执行写作任务
func (w *WriterAgent) Execute(ctx context.Context, taskObj *task.Task) (*task.TaskResult, error) {
	startTime := time.Now()
	w.UpdateStatus("running")

	// 验证任务
	if err := w.ValidateTask(taskObj); err != nil {
		return w.createErrorResult(taskObj, err, startTime), err
	}

	// 解析任务目标
	writingGoal := taskObj.Goal

	// 根据任务类型选择写作方法
	var output interface{}
	var err error

	if strings.Contains(writingGoal, "文章") || strings.Contains(writingGoal, "撰写") {
		output, err = w.writeArticle(ctx, writingGoal, taskObj.Requirements)
	} else if strings.Contains(writingGoal, "报告") || strings.Contains(writingGoal, "总结") {
		output, err = w.writeReport(ctx, writingGoal, taskObj.Requirements)
	} else if strings.Contains(writingGoal, "摘要") || strings.Contains(writingGoal, "总结") {
		output, err = w.writeSummary(ctx, taskObj.Requirements)
	} else if strings.Contains(writingGoal, "润色") || strings.Contains(writingGoal, "修改") {
		output, err = w.editContent(ctx, taskObj.Requirements)
	} else if strings.Contains(writingGoal, "翻译") {
		output, err = w.translateContent(ctx, taskObj.Requirements)
	} else {
		// 默认执行文章写作
		output, err = w.writeArticle(ctx, writingGoal, taskObj.Requirements)
	}

	if err != nil {
		w.UpdateStatus("failed")
		return w.createErrorResult(taskObj, err, startTime), err
	}

	w.UpdateStatus("idle")
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusCompleted,
		Output:    output,
		Error:     "",
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type":     "writer",
			"writing_style":  w.getStyleFromRequirements(taskObj.Requirements),
			"word_count":     w.countWords(output),
			"char_count":     w.countChars(output),
		},
		Timestamp: time.Now(),
		AgentUsed: w.Name,
	}, nil
}

// writeArticle 撰写文章
func (w *WriterAgent) writeArticle(ctx context.Context, topic string, requirements interface{}) (interface{}, error) {
	// 提取要求
	style := w.getStyleFromRequirements(requirements)
	length := w.getLengthFromRequirements(requirements)
	keywords := w.getKeywordsFromRequirements(requirements)

	// 生成大纲
	outline := w.generateOutline(topic, style)

	// 生成内容
	content := w.generateContent(topic, outline, style, length, keywords)

	// 格式化文章
	article := w.formatArticle(topic, outline, content, style)

	return map[string]interface{}{
		"content_type": "article",
		"title":        topic,
		"content":      article,
		"outline":      outline,
		"style":        style,
		"word_count":   w.countWords(article),
	}, nil
}

// writeReport 撰写报告
func (w *WriterAgent) writeReport(ctx context.Context, title string, requirements interface{}) (interface{}, error) {
	// 提取数据
	data := w.getDataFromRequirements(requirements)

	// 生成报告各部分
	executiveSummary := w.generateExecutiveSummary(data)
	analysis := w.generateAnalysis(data)
	recommendations := w.generateRecommendations(data)

	// 格式化报告
	report := fmt.Sprintf(w.templates["report"],
		title,
		executiveSummary,
		analysis,
		recommendations,
	)

	return map[string]interface{}{
		"content_type":      "report",
		"title":             title,
		"content":           report,
		"executive_summary": executiveSummary,
		"analysis":          analysis,
		"recommendations":   recommendations,
		"word_count":        w.countWords(report),
	}, nil
}

// writeSummary 撰写摘要
func (w *WriterAgent) writeSummary(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 提取原文
	content := w.getContentFromRequirements(requirements)
	title := w.getTitleFromRequirements(requirements)

	// 生成摘要
	summary := w.generateSummary(content)

	// 格式化摘要
	formattedSummary := fmt.Sprintf(w.templates["summary"],
		title,
		summary,
	)

	return map[string]interface{}{
		"content_type": "summary",
		"title":        title,
		"summary":      summary,
		"content":      formattedSummary,
		"word_count":   w.countWords(formattedSummary),
	}, nil
}

// editContent 润色内容
func (w *WriterAgent) editContent(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 提取原文
	originalContent := w.getContentFromRequirements(requirements)
	editType := w.getEditTypeFromRequirements(requirements)

	// 执行编辑
	var editedContent string
	var changes []string

	switch editType {
	case "grammar":
		editedContent, changes = w.correctGrammar(originalContent)
	case "style":
		editedContent, changes = w.improveStyle(originalContent)
	case "concise":
		editedContent, changes = w.makeConcise(originalContent)
	default:
		editedContent, changes = w.generalEdit(originalContent)
	}

	return map[string]interface{}{
		"content_type": "edited",
		"original":     originalContent,
		"edited":       editedContent,
		"changes":      changes,
		"change_count": len(changes),
	}, nil
}

// translateContent 翻译内容
func (w *WriterAgent) translateContent(ctx context.Context, requirements interface{}) (interface{}, error) {
	// 提取原文和目标语言
	content := w.getContentFromRequirements(requirements)
	targetLang := w.getTargetLanguage(requirements)

	// 简化实现：这里只是标记翻译，实际应该调用翻译API
	translation := w.mockTranslate(content, targetLang)

	return map[string]interface{}{
		"content_type":  "translation",
		"original":      content,
		"translation":   translation,
		"target_lang":   targetLang,
		"word_count":    w.countWords(translation),
	}, nil
}

// generateOutline 生成大纲
func (w *WriterAgent) generateOutline(topic string, style string) []string {
	outline := []string{
		"引言 - 介绍主题背景和重要性",
		"主体 - 详细阐述核心观点和论据",
		"分析 - 深入分析和案例说明",
		"结论 - 总结要点和展望",
	}
	return outline
}

// generateContent 生成内容
func (w *WriterAgent) generateContent(topic string, outline []string, style string, length int, keywords []string) []string {
	content := make([]string, len(outline))

	for i, section := range outline {
		// 根据风格和关键词生成段落
		paragraph := w.generateParagraph(topic, section, style, keywords)
		content[i] = paragraph
	}

	return content
}

// generateParagraph 生成段落
func (w *WriterAgent) generateParagraph(topic, section, style string, keywords []string) string {
	// 简化实现：生成模板段落
	paragraph := fmt.Sprintf("关于%s的%s部分，我们将深入探讨相关内容。", topic, section)

	// 添加关键词
	if len(keywords) > 0 {
		paragraph += fmt.Sprintf(" 本文将重点讨论%s等关键概念。", strings.Join(keywords, "、"))
	}

	// 根据风格调整
	switch style {
	case "formal":
		paragraph = "综上所述，" + paragraph
	case "casual":
		paragraph = "让我们来看看" + paragraph
	case "professional":
		paragraph = "基于专业视角，" + paragraph
	}

	return paragraph
}

// formatArticle 格式化文章
func (w *WriterAgent) formatArticle(title string, outline []string, content []string, style string) string {
	// 使用文章模板
	overview := content[0]
	details := strings.Join(content[1:len(content)-1], "\n\n")
	conclusion := content[len(content)-1]

	article := fmt.Sprintf(w.templates["article"],
		title,
		overview,
		details,
		conclusion,
	)

	return article
}

// generateExecutiveSummary 生成执行摘要
func (w *WriterAgent) generateExecutiveSummary(data map[string]interface{}) string {
	summary := "本报告基于全面的分析，提供了关键发现和建议。\n\n"
	if len(data) > 0 {
		summary += "主要发现包括数据趋势分析和重要洞察。"
	}
	return summary
}

// generateAnalysis 生成分析部分
func (w *WriterAgent) generateAnalysis(data map[string]interface{}) string {
	analysis := "## 数据分析\n\n"
	analysis += "通过深入分析，我们识别出以下关键模式：\n\n"
	analysis += "1. 趋势显示持续增长\n"
	analysis += "2. 关键指标表现稳定\n"
	analysis += "3. 异常值已得到合理解释\n"
	return analysis
}

// generateRecommendations 生成建议部分
func (w *WriterAgent) generateRecommendations(data map[string]interface{}) string {
	recommendations := "## 建议\n\n"
	recommendations += "基于上述分析，我们提出以下建议：\n\n"
	recommendations += "1. 继续监测关键指标\n"
	recommendations += "2. 优化核心流程\n"
	recommendations += "3. 加强数据收集和分析\n"
	return recommendations
}

// generateSummary 生成摘要
func (w *WriterAgent) generateSummary(content string) string {
	// 简化实现：提取前200个字符作为摘要
	if len(content) <= 200 {
		return content
	}

	// 尝试在句号处截断
	truncated := content[:200]
	lastPeriod := strings.LastIndex(truncated, "。")
	if lastPeriod > 0 {
		return truncated[:lastPeriod+1]
	}

	return truncated + "..."
}

// correctGrammar 修正语法
func (w *WriterAgent) correctGrammar(content string) (string, []string) {
	changes := make([]string, 0)

	// 简化的语法修正规则
	corrected := content

	// 修复常见错误
	corrections := map[string]string{
		"的的":   "的",
		"了了":   "了",
		"是不 是": "不是",
		"可 以":  "可以",
	}

	for wrong, right := range corrections {
		if strings.Contains(corrected, wrong) {
			corrected = strings.ReplaceAll(corrected, wrong, right)
			changes = append(changes, fmt.Sprintf("修正 '%s' -> '%s'", wrong, right))
		}
	}

	return corrected, changes
}

// improveStyle 改进风格
func (w *WriterAgent) improveStyle(content string) (string, []string) {
	changes := make([]string, 0)

	improved := content

	// 移除重复的词
	repeatedWords := []string{"非常", "特别", "十分"}
	for _, word := range repeatedWords {
		pattern := regexp.MustCompile(word + "{2,}")
		if pattern.MatchString(improved) {
			improved = pattern.ReplaceAllString(improved, word)
			changes = append(changes, fmt.Sprintf("移除重复的'%s'", word))
		}
	}

	// 改进句子结构
	improved = strings.ReplaceAll(improved, "。。", "。")
	changes = append(changes, "修正标点符号")

	return improved, changes
}

// makeConcise 使内容更简洁
func (w *WriterAgent) makeConcise(content string) (string, []string) {
	changes := make([]string, 0)

	// 移除冗余词汇
	redundant := []string{"基本上", "总的来说", "事实上"}
	concise := content

	for _, word := range redundant {
		if strings.Contains(concise, word) {
			concise = strings.ReplaceAll(concise, word, "")
			changes = append(changes, fmt.Sprintf("移除冗余词汇'%s'", word))
		}
	}

	return concise, changes
}

// generalEdit 一般编辑
func (w *WriterAgent) generalEdit(content string) (string, []string) {
	edited := content
	changes := make([]string, 0)

	// 基本清理
	edited = strings.TrimSpace(edited)
	edited = regexp.MustCompile(`\s+`).ReplaceAllString(edited, " ")
	changes = append(changes, "清理多余空格")

	// 标准化标点
	edited = strings.ReplaceAll(edited, ", ", "，")
	edited = strings.ReplaceAll(edited, ". ", "。")
	changes = append(changes, "标准化标点符号")

	return edited, changes
}

// mockTranslate 模拟翻译
func (w *WriterAgent) mockTranslate(content string, targetLang string) string {
	// 简化实现：添加翻译标记
	return fmt.Sprintf("[%s翻译] %s", targetLang, content)
}

// 辅助方法：从requirements提取信息

func (w *WriterAgent) getStyleFromRequirements(requirements interface{}) string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if style, ok := reqMap["style"].(string); ok {
			return style
		}
	}
	return "formal" // 默认风格
}

func (w *WriterAgent) getLengthFromRequirements(requirements interface{}) int {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if length, ok := reqMap["length"].(int); ok {
			return length
		}
	}
	return 1000 // 默认长度
}

func (w *WriterAgent) getKeywordsFromRequirements(requirements interface{}) []string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if keywords, ok := reqMap["keywords"].([]string); ok {
			return keywords
		}
	}
	return []string{}
}

func (w *WriterAgent) getDataFromRequirements(requirements interface{}) map[string]interface{} {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if data, ok := reqMap["data"].(map[string]interface{}); ok {
			return data
		}
	}
	return make(map[string]interface{})
}

func (w *WriterAgent) getContentFromRequirements(requirements interface{}) string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if content, ok := reqMap["content"].(string); ok {
			return content
		}
	}
	return ""
}

func (w *WriterAgent) getTitleFromRequirements(requirements interface{}) string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if title, ok := reqMap["title"].(string); ok {
			return title
		}
	}
	return "未命名文档"
}

func (w *WriterAgent) getEditTypeFromRequirements(requirements interface{}) string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if editType, ok := reqMap["edit_type"].(string); ok {
			return editType
		}
	}
	return "general"
}

func (w *WriterAgent) getTargetLanguage(requirements interface{}) string {
	if reqMap, ok := requirements.(map[string]interface{}); ok {
		if lang, ok := reqMap["target_lang"].(string); ok {
			return lang
		}
	}
	return "English"
}

// countWords 统计字数
func (w *WriterAgent) countWords(content interface{}) int {
	if contentMap, ok := content.(map[string]interface{}); ok {
		if text, ok := contentMap["content"].(string); ok {
			// 统计中文字符和英文单词
			chineseChars := regexp.MustCompile(`[\p{Han}]`).FindAllString(text, -1)
			englishWords := regexp.MustCompile(`[a-zA-Z]+`).FindAllString(text, -1)
			return len(chineseChars) + len(englishWords)
		}
	}
	return 0
}

// countChars 统计字符数
func (w *WriterAgent) countChars(content interface{}) int {
	if contentMap, ok := content.(map[string]interface{}); ok {
		if text, ok := contentMap["content"].(string); ok {
			return len([]rune(text))
		}
	}
	return 0
}

// createErrorResult 创建错误结果
func (w *WriterAgent) createErrorResult(taskObj *task.Task, err error, startTime time.Time) *task.TaskResult {
	return &task.TaskResult{
		TaskID:    taskObj.ID,
		TaskGoal:  taskObj.Goal,
		Type:      taskObj.Type,
		Status:    task.TaskStatusFailed,
		Output:    nil,
		Error:     err.Error(),
		Duration:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"agent_type": "writer",
		},
		Timestamp: time.Now(),
		AgentUsed: w.Name,
	}
}

// SetMaxLength 设置最大长度
func (w *WriterAgent) SetMaxLength(max int) {
	w.maxLength = max
}

// SetStyle 设置写作风格
func (w *WriterAgent) SetStyle(style string) {
	if w.Config == nil {
		w.Config = make(map[string]interface{})
	}
	w.Config["default_style"] = style
}

// AddTemplate 添加模板
func (w *WriterAgent) AddTemplate(name, template string) {
	w.templates[name] = template
}
