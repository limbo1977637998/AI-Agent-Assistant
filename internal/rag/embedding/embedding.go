package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai-agent-assistant/internal/config"
)

// EmbeddingProvider 向量化提供者接口
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	GetDimension() int
}

// GLMEmbedding GLM向量化提供者
type GLMEmbedding struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// QwenEmbedding 千问向量化提供者
type QwenEmbedding struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// EmbeddingRequest 向量化请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse 向量化响应
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingProvider 创建向量化提供者
// provider: "glm" 或 "qwen"
func NewEmbeddingProvider(provider string, cfg config.ModelConfig) (EmbeddingProvider, error) {
	switch provider {
	case "qwen":
		return &QwenEmbedding{
			apiKey:  cfg.APIKey,
			baseURL: cfg.BaseURL,
			model:   "text-embedding-v3", // 千问的embedding模型
			client:  &http.Client{},
		}, nil
	case "glm":
		return &GLMEmbedding{
			apiKey:  cfg.APIKey,
			baseURL: cfg.BaseURL,
			model:   "embedding-2", // GLM的embedding模型
			client:  &http.Client{},
		}, nil
	default:
		// 默认使用GLM
		return &GLMEmbedding{
			apiKey:  cfg.APIKey,
			baseURL: cfg.BaseURL,
			model:   "embedding-2",
			client:  &http.Client{},
		}, nil
	}
}

// Embed 将文本向量化
func (e *GLMEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	// 限制文本长度（API限制）
	if len(text) > 8000 {
		text = text[:8000]
	}

	reqBody := EmbeddingRequest{
		Model: e.model,
		Input: []string{text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
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

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// GetDimension 获取向量维度
func (e *GLMEmbedding) GetDimension() int {
	// GLM embedding-2 模型的维度是1024
	return 1024
}

// BatchEmbed 批量向量化
func (e *GLMEmbedding) BatchEmbed(ctx context.Context, texts []string) ([][]float64, error) {
	vectors := make([][]float64, len(texts))

	// 批量请求
	reqBody := EmbeddingRequest{
		Model: e.model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
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

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	for _, data := range embedResp.Data {
		if data.Index < len(vectors) {
			vectors[data.Index] = data.Embedding
		}
	}

	return vectors, nil
}

// CosineSimilarity 计算余弦相似度
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	// 简单的平方根实现
	// 生产环境应使用 math.Sqrt
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// ========== 千问 Embedding 实现 ==========

// Embed 将文本向量化（千问）
func (e *QwenEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	// 限制文本长度（千问API限制）
	if len(text) > 8000 {
		text = text[:8000]
	}

	reqBody := EmbeddingRequest{
		Model: e.model,
		Input: []string{text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
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

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// GetDimension 获取向量维度（千问）
func (e *QwenEmbedding) GetDimension() int {
	// 千问 text-embedding-v3 模型的维度默认是1024
	return 1024
}

// BatchEmbed 批量向量化（千问）
func (e *QwenEmbedding) BatchEmbed(ctx context.Context, texts []string) ([][]float64, error) {
	vectors := make([][]float64, len(texts))

	// 批量请求
	reqBody := EmbeddingRequest{
		Model: e.model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
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

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	for _, data := range embedResp.Data {
		if data.Index < len(vectors) {
			vectors[data.Index] = data.Embedding
		}
	}

	return vectors, nil
}
