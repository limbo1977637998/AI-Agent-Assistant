package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "执行数学计算，支持加减乘除和括号。例如：'1 + 2 * 3' 或 '(10 - 5) / 2'"
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	expression, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("missing expression argument")
	}

	// 简单的示例计算（实际项目应使用eval库）
	// 这里仅返回说明
	return fmt.Sprintf("计算表达式: %s（注意：当前版本为简化实现，实际计算功能需要集成数学表达式解析库）", expression), nil
}

// WeatherTool 天气查询工具
type WeatherTool struct{}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
	return "weather"
}

func (t *WeatherTool) Description() string {
	return "查询指定城市的天气信息。参数：city（城市名称，例如：北京、上海、New York）"
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	city, ok := args["city"].(string)
	if !ok {
		return "", fmt.Errorf("missing city argument")
	}

	// 中文城市名到英文的映射
	cityMap := map[string]string{
		"北京": "Beijing",
		"上海": "Shanghai",
		"广州": "Guangzhou",
		"深圳": "Shenzhen",
		"杭州": "Hangzhou",
		"成都": "Chengdu",
		"武汉": "Wuhan",
		"西安": "Xian",
		"南京": "Nanjing",
		"重庆": "Chongqing",
		"天津": "Tianjin",
		"苏州": "Suzhou",
		"香港": "Hong Kong",
		"台北": "Taipei",
	}

	// 查找英文城市名
	englishCity, exists := cityMap[city]
	if !exists {
		// 如果不在映射表中，尝试直接使用原城市名
		englishCity = city
	}

	// 使用Open-Meteo免费API查询天气
	geocodingURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=zh&format=json", url.QueryEscape(englishCity))
	resp, err := http.Get(geocodingURL)
	if err != nil {
		return "", fmt.Errorf("failed to geocode city: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read geocoding response: %w", err)
	}

	var geoResp struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Name      string  `json:"name"`
			Country   string `json:"country"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &geoResp); err != nil {
		return "", fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return "", fmt.Errorf("city not found: %s (searched as: %s)", city, englishCity)
	}

	lat := geoResp.Results[0].Latitude
	lon := geoResp.Results[0].Longitude

	// 获取天气信息
	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&current_weather=true", lat, lon)
	weatherResp, err := http.Get(weatherURL)
	if err != nil {
		return "", fmt.Errorf("failed to get weather: %w", err)
	}
	defer weatherResp.Body.Close()

	weatherBody, err := io.ReadAll(weatherResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read weather response: %w", err)
	}

	var weatherData struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			Windspeed   float64 `json:"windspeed"`
			WeatherCode int     `json:"weathercode"`
		} `json:"current_weather"`
	}

	if err := json.Unmarshal(weatherBody, &weatherData); err != nil {
		return "", fmt.Errorf("failed to parse weather response: %w", err)
	}

	// 根据天气代码描述天气
	weatherDesc := getWeatherDescription(weatherData.CurrentWeather.WeatherCode)

	return fmt.Sprintf("%s 当前天气：温度 %.1f°C，风速 %.1f km/h，%s",
		city, weatherData.CurrentWeather.Temperature, weatherData.CurrentWeather.Windspeed, weatherDesc), nil
}

// getWeatherDescription 根据天气代码返回天气描述
func getWeatherDescription(code int) string {
	weatherCodes := map[int]string{
		0:  "晴朗",
		1:  "大部晴朗",
		2:  "多云",
		3:  "阴天",
		45: "雾",
		48: "雾凇",
		51: "毛毛雨",
		53: "毛毛雨",
		55: "毛毛雨",
		61: "小雨",
		63: "中雨",
		65: "大雨",
		71: "小雪",
		73: "中雪",
		75: "大雪",
		95: "雷雨",
	}

	if desc, ok := weatherCodes[code]; ok {
		return desc
	}
	return "天气状况未知"
}

// SearchTool 搜索工具
type SearchTool struct{}

func NewSearchTool() *SearchTool {
	return &SearchTool{}
}

func (t *SearchTool) Name() string {
	return "search"
}

func (t *SearchTool) Description() string {
	return "在互联网上搜索信息。参数：query（搜索关键词）"
}

func (t *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("missing query argument")
	}

	// 使用DuckDuckGo进行搜索（简化版）
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read search response: %w", err)
	}

	var result struct {
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		AbstractTopic string `json:"AbstractTopic"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse search response: %w", err)
	}

	var sb strings.Builder
	if result.AbstractText != "" {
		sb.WriteString(fmt.Sprintf("摘要: %s\n来源: %s\n\n", result.AbstractText, result.AbstractURL))
	}

	sb.WriteString(fmt.Sprintf("搜索关键词: %s\n", query))
	sb.WriteString("(注意：当前使用DuckDuckGo简化API，如需更全面的搜索结果，请集成专业搜索API)")

	return sb.String(), nil
}

// ToolManager 工具管理器
type ToolManager struct {
	tools map[string]Tool
}

// NewToolManager 创建工具管理器
func NewToolManager(enabledTools []string) *ToolManager {
	allTools := map[string]Tool{
		"calculator": NewCalculatorTool(),
		"weather":    NewWeatherTool(),
		"search":     NewSearchTool(),
	}

	manager := &ToolManager{
		tools: make(map[string]Tool),
	}

	for _, name := range enabledTools {
		if tool, ok := allTools[name]; ok {
			manager.tools[name] = tool
		}
	}

	return manager
}

// GetTool 获取工具
func (m *ToolManager) GetTool(name string) (Tool, bool) {
	tool, ok := m.tools[name]
	return tool, ok
}

// GetAllTools 获取所有工具
func (m *ToolManager) GetAllTools() []Tool {
	tools := make([]Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetToolDescriptions 获取工具描述
func (m *ToolManager) GetToolDescriptions() string {
	var sb strings.Builder
	sb.WriteString("可用工具:\n")
	for _, tool := range m.tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}
	return sb.String()
}
