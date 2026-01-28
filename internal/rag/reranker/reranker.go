package reranker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Reranker 重排序器接口
type Reranker interface {
	Rerank(ctx context.Context, query string, documents []Document) ([]Document, error)
}

// Document 文档
type Document struct {
	ID      string
	Content string
	Score   float64 // 原始得分
}

// CrossEncoderReranker CrossEncoder重排序器
type CrossEncoderReranker struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewCrossEncoderReranker 创建CrossEncoder重排序器
func NewCrossEncoderReranker(apiKey, baseURL, model string) (*CrossEncoderReranker, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if baseURL == "" {
		// 默认使用Cohere Rerank API（也可以配置为其他服务）
		baseURL = "https://api.cohere.ai/v1"
	}

	if model == "" {
		model = "rerank-english-v2.0"
	}

	return &CrossEncoderReranker{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

// Rerank 重排序文档
func (r *CrossEncoderReranker) Rerank(ctx context.Context, query string, documents []Document) ([]Document, error) {
	if len(documents) == 0 {
		return documents, nil
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"model":             r.model,
		"query":             query,
		"documents":         extractContents(documents),
		"top_n":             len(documents),
		"return_documents":  true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.baseURL+"/rerank", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cohereResp struct {
		Results []struct {
			Index          int     `json:"index"`
			RelevanceScore float64 `json:"relevance_score"`
			Document       struct {
				Text string `json:"text"`
			} `json:"document"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &cohereResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 根据重排序结果重新组织文档
	reranked := make([]Document, len(cohereResp.Results))
	for i, result := range cohereResp.Results {
		originalDoc := documents[result.Index]
		reranked[i] = Document{
			ID:      originalDoc.ID,
			Content: originalDoc.Content,
			Score:   result.RelevanceScore, // 使用重排序后的得分
		}
	}

	return reranked, nil
}

// SimpleReranker 简单的基于规则的重排序器
type SimpleReranker struct {
	keywordWeight float64 // 关键词权重
	vectorWeight  float64 // 向量相似度权重
}

// NewSimpleReranker 创建简单重排序器
func NewSimpleReranker(keywordWeight, vectorWeight float64) *SimpleReranker {
	return &SimpleReranker{
		keywordWeight: keywordWeight,
		vectorWeight:  vectorWeight,
	}
}

// Rerank 重排序文档（基于关键词匹配和向量相似度的加权）
func (sr *SimpleReranker) Rerank(ctx context.Context, query string, documents []Document) ([]Document, error) {
	// 计算每个文档的重排序得分
	for i := range documents {
		// 关键词匹配得分
		keywordScore := sr.calculateKeywordScore(query, documents[i].Content)

		// 组合得分（假设原始Score是向量相似度）
		documents[i].Score = sr.keywordWeight*keywordScore + sr.vectorWeight*documents[i].Score
	}

	// 按新得分降序排序
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Score > documents[j].Score
	})

	return documents, nil
}

// calculateKeywordScore 计算关键词匹配得分
func (sr *SimpleReranker) calculateKeywordScore(query, content string) float64 {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	score := 0.0
	queryWords := strings.Fields(queryLower)

	for _, word := range queryWords {
		if strings.Contains(contentLower, word) {
			score += 1.0
		}
	}

	// 归一化
	if len(queryWords) > 0 {
		score = score / float64(len(queryWords))
	}

	return score
}

// extractContents 提取文档内容
func extractContents(documents []Document) []string {
	contents := make([]string, len(documents))
	for i, doc := range documents {
		contents[i] = doc.Content
	}
	return contents
}
