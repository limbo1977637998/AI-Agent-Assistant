package retriever

import (
	"math"
	"regexp"
	"strings"
)

// BM25 BM25关键词检索算法
type BM25 struct {
	documents []Document
	idf       map[string]float64
	k1        float64 // 调节词频饱和度
	b         float64 // 调节文档长度归一化
	avgDocLen float64
}

// Document 文档
type Document struct {
	ID      string
	Content string
	Tokens  []string
}

// NewBM25 创建BM25检索器
func NewBM25(k1, b float64) *BM25 {
	return &BM25{
		documents: make([]Document, 0),
		idf:       make(map[string]float64),
		k1:        k1,
		b:         b,
	}
}

// Index 索引文档
func (bm *BM25) Index(docs []Document) {
	bm.documents = docs
	bm.calculateIDF()
	bm.calculateAvgDocLen()
}

// calculateIDF 计算IDF（逆文档频率）
func (bm *BM25) calculateIDF() {
	N := len(bm.documents)
	docFreq := make(map[string]int)

	// 统计每个词出现在多少文档中
	for _, doc := range bm.documents {
		seen := make(map[string]bool)
		for _, token := range doc.Tokens {
			if !seen[token] {
				docFreq[token]++
				seen[token] = true
			}
		}
	}

	// 计算IDF
	for term, df := range docFreq {
		dfFloat := float64(df)
		bm.idf[term] = math.Log(float64(N-df+1) / (dfFloat + 0.5))
	}
}

// calculateAvgDocLen 计算平均文档长度
func (bm *BM25) calculateAvgDocLen() {
	totalLen := 0
	for _, doc := range bm.documents {
		totalLen += len(doc.Tokens)
	}
	bm.avgDocLen = float64(totalLen) / float64(len(bm.documents))
}

// Search 搜索
func (bm *BM25) Search(query string, topK int) []SearchResult {
	queryTokens := bm.tokenize(query)
	scores := make(map[string]float64)

	// 计算每个文档的得分
	for _, doc := range bm.documents {
		score := bm.calculateScore(doc, queryTokens)
		scores[doc.ID] = score
	}

	// 排序并返回topK
	results := make([]SearchResult, 0)
	for _, doc := range bm.documents {
		results = append(results, SearchResult{
			DocID:  doc.ID,
			Score:  scores[doc.ID],
			Content: doc.Content,
		})
	}

	// 按得分降序排序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 返回topK
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

// calculateScore 计算文档得分
func (bm *BM25) calculateScore(doc Document, queryTokens []string) float64 {
	score := 0.0
	docLen := float64(len(doc.Tokens))

	for _, token := range queryTokens {
		// 计算词频
		tf := 0.0
		for _, docToken := range doc.Tokens {
			if docToken == token {
				tf++
			}
		}

		if tf == 0 {
			continue
		}

		// IDF
		idf, ok := bm.idf[token]
		if !ok {
			idf = 0
		}

		// BM25公式
		numerator := tf * (bm.k1 + 1)
		denominator := tf + bm.k1*(1-bm.b+bm.b*(docLen/bm.avgDocLen))
		score += idf * (numerator / denominator)
	}

	return score
}

// tokenize 分词
func (bm *BM25) tokenize(text string) []string {
	// 转小写
	text = strings.ToLower(text)

	// 简单的分词（支持中英文）
	// 英文按空格分，中文按字符分
	re := regexp.MustCompile(`[a-zA-Z]+|[\p{Han}]`)
	matches := re.FindAllString(text, -1)

	tokens := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 0 {
			tokens = append(tokens, match)
		}
	}

	return tokens
}

// SearchResult 搜索结果
type SearchResult struct {
	DocID   string
	Score   float64
	Content string
}
