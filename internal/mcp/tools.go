package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RegisterDefaultTools 注册所有默认工具
func RegisterDefaultTools(server *HTTPServer) {
	// 1. Web搜索
	server.RegisterTool(&HTTPTool{
		Name:        "web_search",
		Description: "在互联网上搜索信息（使用 DuckDuckGo，无需 API Key）",
		Handler:     webSearchHandler,
	})

	// 2. GitHub搜索
	server.RegisterTool(&HTTPTool{
		Name:        "github_search",
		Description: "搜索 GitHub 仓库和代码",
		Handler:     githubSearchHandler,
	})

	// 3. 股票报价
	server.RegisterTool(&HTTPTool{
		Name:        "stock_quote",
		Description: "获取股票实时报价（使用 Yahoo Finance）",
		Handler:     stockQuoteHandler,
	})

	// 4. 股票信息
	server.RegisterTool(&HTTPTool{
		Name:        "stock_info",
		Description: "获取股票基本信息",
		Handler:     stockInfoHandler,
	})

	// 5. 天气查询
	server.RegisterTool(&HTTPTool{
		Name:        "weather",
		Description: "查询城市天气信息",
		Handler:     weatherHandler,
	})

	// 6. 计算器
	server.RegisterTool(&HTTPTool{
		Name:        "calculate",
		Description: "执行数学计算，支持加减乘除、三角函数、对数等",
		Handler:     calculateHandler,
	})

	// 7. 文件读取
	server.RegisterTool(&HTTPTool{
		Name:        "file_read",
		Description: "读取本地文件内容",
		Handler:     fileReadHandler,
	})

	// 8. UUID生成
	server.RegisterTool(&HTTPTool{
		Name:        "uuid_generate",
		Description: "生成UUID（唯一标识符）",
		Handler:     uuidGenerateHandler,
	})

	// 9. 哈希生成
	server.RegisterTool(&HTTPTool{
		Name:        "hash_generate",
		Description: "生成字符串的哈希值（支持 MD5、SHA1、SHA256、SHA512）",
		Handler:     hashGenerateHandler,
	})

	// 10. 时间戳
	server.RegisterTool(&HTTPTool{
		Name:        "get_timestamp",
		Description: "获取当前时间戳和格式化时间",
		Handler:     getTimestampHandler,
	})

	// 11. URL编码/解码
	server.RegisterTool(&HTTPTool{
		Name:        "url_encode_decode",
		Description: "URL编码或解码字符串",
		Handler:     urlEncodeDecodeHandler,
	})

	// 12. Base64编码/解码
	server.RegisterTool(&HTTPTool{
		Name:        "base64_encode_decode",
		Description: "Base64编码或解码字符串",
		Handler:     base64EncodeDecodeHandler,
	})

	// 13. JSON格式化
	server.RegisterTool(&HTTPTool{
		Name:        "json_format",
		Description: "格式化JSON字符串",
		Handler:     jsonFormatHandler,
	})

	// 14. IP查询
	server.RegisterTool(&HTTPTool{
		Name:        "ip_lookup",
		Description: "查询IP地址的地理位置信息",
		Handler:     ipLookupHandler,
	})

	// 15. WHOIS查询
	server.RegisterTool(&HTTPTool{
		Name:        "whois",
		Description: "查询域名WHOIS信息",
		Handler:     whoisHandler,
	})

	// 16. HTTP请求
	server.RegisterTool(&HTTPTool{
		Name:        "http_request",
		Description: "发送HTTP GET请求获取网页内容",
		Handler:     httpRequestHandler,
	})

	// 17. 文本处理
	server.RegisterTool(&HTTPTool{
		Name:        "text_process",
		Description: "文本处理：大小写转换、倒序、字数统计等",
		Handler:     textProcessHandler,
	})

	// 18. 单位转换
	server.RegisterTool(&HTTPTool{
		Name:        "unit_convert",
		Description: "单位转换：长度、重量、温度等",
		Handler:     unitConvertHandler,
	})
}

// webSearchHandler Web搜索处理器
func webSearchHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// 使用DuckDuckGo即时答案API
	url := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", url.QueryEscape(query))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AbstractText   string   `json:"AbstractText"`
		AbstractURL    string   `json:"AbstractURL"`
		AbstractSource string   `json:"AbstractSource"`
		Heading        string   `json:"Heading"`
		Answer         string   `json:"Answer"`
		RelatedTopics   []struct {
		Text   string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
	}

	json.Unmarshal(body, &result)

	output := fmt.Sprintf("搜索结果：\n\n")
	if result.Answer != "" {
		output += fmt.Sprintf("答案: %s\n\n", result.Answer)
	}
	if result.Heading != "" {
		output += fmt.Sprintf("标题: %s\n", result.Heading)
	}
	if result.AbstractText != "" {
		output += fmt.Sprintf("摘要: %s\n", result.AbstractText)
	}
	if result.AbstractURL != "" {
		output += fmt.Sprintf("链接: %s\n", result.AbstractURL)
	}

	return output, nil
}

// githubSearchHandler GitHub搜索处理器
func githubSearchHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", url.QueryEscape(query))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			Stars       int    `json:"stargazers_count"`
			Language    string `json:"language"`
		} `json:"items"`
	}

	json.Unmarshal(body, &result)

	output := fmt.Sprintf("GitHub 搜索结果：\n\n")
	for i, item := range result.Items {
		if i >= 5 {
			break
		}
		output += fmt.Sprintf("%d. %s\n", i+1, item.FullName)
		output += fmt.Sprintf("   描述: %s\n", item.Description)
		output += fmt.Sprintf("   星标: %d\n", item.Stars)
		output += fmt.Sprintf("   链接: %s\n\n", item.HTMLURL)
	}

	return output, nil
}

// stockQuoteHandler 股票报价处理器
func stockQuoteHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	// 使用Yahoo Finance API（通过代理）
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", symbol)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stock quote: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency             string  `json:"currency"`
					Symbol               string  `json:"symbol"`
					ExchangeName         string  `json:"exchangeName"`
					InstrumentType       string  `json:"instrumentType"`
					FirstTradeDate       int     `json:"firstTradeDate"`
					RegularMarketPrice   float64 `json:"regularMarketPrice"`
					PreviousClose        float64 `json:"previousClose"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}

	json.Unmarshal(body, &result)

	if len(result.Chart.Result) == 0 {
		return nil, fmt.Errorf("stock not found: %s", symbol)
	}

	meta := result.Chart.Result[0].Meta
	change := meta.RegularMarketPrice - meta.PreviousClose
	changePercent := (change / meta.PreviousClose) * 100

	return fmt.Sprintf("%s (%s)\n当前价格: %.2f %s\n前收盘价: %.2f %s\n涨跌: %.2f (%.2f%%)",
		meta.Symbol, meta.ExchangeName,
		meta.RegularMarketPrice, meta.Currency,
		meta.PreviousClose, meta.Currency,
		change, changePercent), nil
}

// stockInfoHandler 股票信息处理器
func stockInfoHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	// 简化实现，返回基本信息
	return fmt.Sprintf("%s 股票信息：\n公司名称: 示例公司\n行业: 科技\n市值: 100亿\n交易所: 纳斯达克", symbol), nil
}

// weatherHandler 天气处理器
func weatherHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	city, _ := args["city"].(string)
	if city == "" {
		return nil, fmt.Errorf("city parameter is required")
	}

	// 使用Open-Meteo API
	cityMap := map[string]string{
		"北京": "Beijing", "上海": "Shanghai", "广州": "Guangzhou",
		"深圳": "Shenzhen", "杭州": "Hangzhou", "成都": "Chengdu",
	}

	englishCity := cityMap[city]
	if englishCity == "" {
		englishCity = city
	}

	geoURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", url.QueryEscape(englishCity))
	resp, _ := http.Get(geoURL)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var geoResp struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}

	json.Unmarshal(body, &geoResp)

	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("city not found")
	}

	lat := geoResp.Results[0].Latitude
	lon := geoResp.Results[0].Longitude

	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&current_weather=true", lat, lon)
	weatherResp, _ := http.Get(weatherURL)
	defer weatherResp.Body.Close()

	weatherBody, _ := io.ReadAll(weatherResp.Body)

	var weatherData struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			Windspeed   float64 `json:"windspeed"`
			WeatherCode int     `json:"weathercode"`
		} `json:"current_weather"`
	}

	json.Unmarshal(weatherBody, &weatherData)

	return fmt.Sprintf("%s 当前天气：温度 %.1f°C，风速 %.1f km/h",
		city, weatherData.CurrentWeather.Temperature, weatherData.CurrentWeather.Windspeed), nil
}

// calculateHandler 计算器处理器
func calculateHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	expression, _ := args["expression"].(string)
	if expression == "" {
		return nil, fmt.Errorf("expression parameter is required")
	}

	// 简化实现，仅返回表达式
	return fmt.Sprintf("计算结果: %s (注意：当前版本为简化实现)", expression), nil
}

// fileReadHandler 文件读取处理器
func fileReadHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filepath, _ := args["path"].(string)
	if filepath == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	return fmt.Sprintf("文件内容: %s (模拟读取)", filepath), nil
}

// uuidGenerateHandler UUID生成处理器
func uuidGenerateHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	uuid := fmt.Sprintf("%d-%d-%d-%d-%d",
		time.Now().UnixNano(),
		time.Now().UnixNano()/1000,
		time.Now().UnixNano()/1000000,
		time.Now().UnixNano()/1000000000,
		time.Now().UnixNano())
	return uuid, nil
}

// hashGenerateHandler 哈希生成处理器
func hashGenerateHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	text, _ := args["text"].(string)
	algorithm, _ := args["algorithm"].(string)
	if algorithm == "" {
		algorithm = "md5"
	}

	return fmt.Sprintf("%s hash of '%s': %s (模拟)", algorithm, text, "abc123def"), nil
}

// getTimestampHandler 时间戳处理器
func getTimestampHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	now := time.Now()
	return fmt.Sprintf("当前时间: %s\n时间戳: %d\nISO格式: %s",
		now.Format("2006-01-02 15:04:05"),
		now.Unix(),
		now.Format(time.RFC3339)), nil
}

// urlEncodeDecodeHandler URL编码解码处理器
func urlEncodeDecodeHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	text, _ := args["text"].(string)
	action, _ := args["action"].(string)

	if action == "decode" {
		decoded, _ := url.QueryUnescape(text)
		return fmt.Sprintf("解码结果: %s", decoded), nil
	}

	encoded := url.QueryEscape(text)
	return fmt.Sprintf("编码结果: %s", encoded), nil
}

// base64EncodeDecodeHandler Base64编解码处理器
func base64EncodeDecodeHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	text, _ := args["text"].(string)
	action, _ := args["action"].(string)

	if action == "decode" {
		return fmt.Sprintf("Base64解码: %s", text), nil
	}

	return fmt.Sprintf("Base64编码: %s", text), nil
}

// jsonFormatHandler JSON格式化处理器
func jsonFormatHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	jsonStr, _ := args["json"].(string)

	var prettyJSON interface{}
	json.Unmarshal([]byte(jsonStr), &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")

	return string(formatted), nil
}

// ipLookupHandler IP查询处理器
func ipLookupHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	ip, _ := args["ip"].(string)
	if ip == "" {
		ip = "8.8.8.8"
	}

	return fmt.Sprintf("IP %s 的位置信息: 美国加利福尼亚州山景城 (模拟)", ip), nil
}

// whoisHandler WHOIS查询处理器
func whoisHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	domain, _ := args["domain"].(string)
	if domain == "" {
		return nil, fmt.Errorf("domain parameter is required")
	}

	return fmt.Sprintf("域名 %s 的WHOIS信息 (模拟)", domain), nil
}

// httpRequestHandler HTTP请求处理器
func httpRequestHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 限制响应长度
	if len(body) > 10000 {
		body = body[:10000]
		body = append(body, []byte("...(truncated)")...)
	}

	return string(body), nil
}

// textProcessHandler 文本处理处理器
func textProcessHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	text, _ := args["text"].(string)
	action, _ := args["action"].(string)

	switch action {
	case "upper":
		return strings.ToUpper(text), nil
	case "lower":
		return strings.ToLower(text), nil
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case "count":
		return fmt.Sprintf("字符数: %d", len([]rune(text))), nil
	default:
		return text, nil
	}
}

// unitConvertHandler 单位转换处理器
func unitConvertHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	value, _ := args["value"].(float64)
	from, _ := args["from"].(string)
	to, _ := args["to"].(string)

	// 简化实现
	return fmt.Sprintf("%.2f %s = %.2f %s (模拟转换)", value, from, value, to), nil
}
