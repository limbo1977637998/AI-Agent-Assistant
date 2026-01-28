package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// DataProcessingResult 数据处理结果
type DataProcessingResult struct {
	Success   bool                   `json:"success"`             // 处理是否成功
	Message   string                 `json:"message"`             // 结果消息
	Data      interface{}            `json:"data,omitempty"`      // 返回数据
	Error     string                 `json:"error,omitempty"`     // 错误信息
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // 元数据（行数、列数等）
}

// DataProcessorTool 数据处理工具
// 提供CSV/JSON处理、数据清洗、统计分析等功能
type DataProcessorTool struct {
	name        string
	description string
	version     string
}

// NewDataProcessorTool 创建数据处理工具实例
func NewDataProcessorTool() *DataProcessorTool {
	return &DataProcessorTool{
		name:        "data_processor",
		description: "数据处理工具 - CSV/JSON处理、数据清洗、统计分析",
		version:     "1.0.0",
	}
}

// Name 返回工具名称
func (t *DataProcessorTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *DataProcessorTool) Description() string {
	return t.description
}

// Version 返回工具版本
func (t *DataProcessorTool) Version() string {
	return t.version
}

// Execute 执行数据处理操作
// 支持的操作类型：parse_csv, parse_json, clean, filter, aggregate, transform, merge
func (t *DataProcessorTool) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "parse_csv":
		return t.parseCSV(params)
	case "parse_json":
		return t.parseJSON(params)
	case "clean":
		return t.cleanData(params)
	case "filter":
		return t.filterData(params)
	case "aggregate":
		return t.aggregateData(params)
	case "transform":
		return t.transformData(params)
	case "merge":
		return t.mergeData(params)
	case "sort":
		return t.sortData(params)
	case "deduplicate":
		return t.deduplicateData(params)
	case "fill_missing":
		return t.fillMissingValues(params)
	default:
		return &DataProcessingResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的操作类型: %s", operation),
		}, nil
	}
}

// parseCSV 解析CSV数据
// 参数：
//   - content: CSV内容字符串（必填）
//   - has_header: 是否有表头（可选，默认true）
//   - delimiter: 分隔符（可选，默认","）
func (t *DataProcessorTool) parseCSV(params map[string]interface{}) (*DataProcessingResult, error) {
	content, ok := params["content"].(string)
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: content",
		}, nil
	}

	hasHeader := true
	if hh, ok := params["has_header"].(bool); ok {
		hasHeader = hh
	}

	delimiter := ","
	if d, ok := params["delimiter"].(string); ok {
		delimiter = d
	}

	// 解析CSV
	reader := csv.NewReader(strings.NewReader(content))
	if delimiter != "," {
		reader.Comma = rune(delimiter[0])
	}

	records, err := reader.ReadAll()
	if err != nil {
		return &DataProcessingResult{
			Success: false,
			Error:   fmt.Sprintf("CSV解析失败: %v", err),
		}, nil
	}

	if len(records) == 0 {
		return &DataProcessingResult{
			Success: false,
			Error:   "CSV内容为空",
		}, nil
	}

	var headers []string
	var data []map[string]interface{}

	if hasHeader {
		headers = records[0]
		dataRows := records[1:]

		for _, row := range dataRows {
			rowMap := make(map[string]interface{})
			for i, value := range row {
				if i < len(headers) {
					rowMap[headers[i]] = value
				}
			}
			data = append(data, rowMap)
		}
	} else {
		// 无表头，使用列索引作为键
		for _, row := range records {
			rowMap := make(map[string]interface{})
			for i, value := range row {
				rowMap[fmt.Sprintf("column_%d", i)] = value
			}
			data = append(data, rowMap)
		}
	}

	return &DataProcessingResult{
		Success: true,
		Message: "CSV解析成功",
		Data: map[string]interface{}{
			"headers": headers,
			"data":    data,
		},
		Metadata: map[string]interface{}{
			"row_count":    len(data),
			"column_count": len(headers),
			"has_header":   hasHeader,
		},
	}, nil
}

// parseJSON 解析JSON数据
// 参数：
//   - content: JSON内容字符串（必填）
func (t *DataProcessorTool) parseJSON(params map[string]interface{}) (*DataProcessingResult, error) {
	content, ok := params["content"].(string)
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: content",
		}, nil
	}

	var data interface{}
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		return &DataProcessingResult{
			Success: false,
			Error:   fmt.Sprintf("JSON解析失败: %v", err),
		}, nil
	}

	// 分析数据结构
	metadata := t.analyzeJSONStructure(data)

	return &DataProcessingResult{
		Success: true,
		Message: "JSON解析成功",
		Data:    data,
		Metadata: metadata,
	}, nil
}

// analyzeJSONStructure 分析JSON数据结构
func (t *DataProcessorTool) analyzeJSONStructure(data interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	switch v := data.(type) {
	case []interface{}:
		metadata["type"] = "array"
		metadata["length"] = len(v)
		if len(v) > 0 {
			if _, ok := v[0].(map[string]interface{}); ok {
				// 对象数组
				keys := make([]string, 0)
				for key := range v[0].(map[string]interface{}) {
					keys = append(keys, key)
				}
				metadata["fields"] = keys
			}
		}
	case map[string]interface{}:
		metadata["type"] = "object"
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		metadata["fields"] = keys
	default:
		metadata["type"] = "primitive"
	}

	return metadata
}

// cleanData 数据清洗
// 参数：
//   - data: 待清洗的数据（数组，必填）
//   - operations: 清洗操作列表（可选）
// 支持的操作：remove_empty, trim_whitespace, normalize_case, remove_duplicates
func (t *DataProcessorTool) cleanData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	var operations []string
	if ops, ok := params["operations"].([]interface{}); ok {
		for _, op := range ops {
			if opStr, ok := op.(string); ok {
				operations = append(operations, opStr)
			}
		}
	}

	// 默认执行所有清洗操作
	if len(operations) == 0 {
		operations = []string{"remove_empty", "trim_whitespace", "normalize_case"}
	}

	result := make([]interface{}, 0)
	removedCount := 0

	for _, row := range dataParam {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		// 移除空行
		if contains(operations, "remove_empty") && t.isEmptyRow(rowMap) {
			removedCount++
			continue
		}

		// 清洗每个字段
		cleanedRow := make(map[string]interface{})
		for key, value := range rowMap {
			cleanedValue := value

			// 去除首尾空格
			if contains(operations, "trim_whitespace") {
				if str, ok := value.(string); ok {
					cleanedValue = strings.TrimSpace(str)
				}
			}

			// 标准化大小写
			if contains(operations, "normalize_case") {
				if str, ok := cleanedValue.(string); ok {
					cleanedValue = strings.ToLower(strings.TrimSpace(str))
				}
			}

			cleanedRow[key] = cleanedValue
		}

		result = append(result, cleanedRow)
	}

	// 去重
	if contains(operations, "remove_duplicates") {
		originalLen := len(result)
		result = t.deduplicate(result)
		removedCount += (originalLen - len(result))
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("数据清洗完成：保留 %d 行，移除 %d 行", len(result), removedCount),
		Data:    result,
		Metadata: map[string]interface{}{
			"original_count":  len(dataParam),
			"cleaned_count":   len(result),
			"removed_count":   removedCount,
			"operations":      operations,
		},
	}, nil
}

// isEmptyRow 检查是否为空行
func (t *DataProcessorTool) isEmptyRow(row map[string]interface{}) bool {
	for _, value := range row {
		if str, ok := value.(string); ok {
			if strings.TrimSpace(str) != "" {
				return false
			}
		} else if value != nil {
			return false
		}
	}
	return true
}

// deduplicate 数据去重
func (t *DataProcessorTool) deduplicate(data []interface{}) []interface{} {
	seen := make(map[string]bool)
	result := make([]interface{}, 0)

	for _, row := range data {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			result = append(result, row)
			continue
		}

		// 生成行的哈希作为唯一标识
		rowJSON, _ := json.Marshal(rowMap)
		rowHash := string(rowJSON)

		if !seen[rowHash] {
			seen[rowHash] = true
			result = append(result, row)
		}
	}

	return result
}

// filterData 数据过滤
// 参数：
//   - data: 待过滤的数据（必填）
//   - conditions: 过滤条件（必填）
//     格式：[{"field": "age", "operator": ">", "value": 18}]
// 支持的操作符：>, >=, <, <=, ==, !=, contains, starts_with, ends_with
func (t *DataProcessorTool) filterData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	conditionsParam, ok := params["conditions"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: conditions",
		}, nil
	}

	var conditions []FilterCondition
	for _, c := range conditionsParam {
		condMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condition := FilterCondition{
			Field:    condMap["field"].(string),
			Operator: condMap["operator"].(string),
			Value:    condMap["value"],
		}
		conditions = append(conditions, condition)
	}

	result := make([]interface{}, 0)
	for _, row := range dataParam {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		// 检查是否满足所有条件
		if t.matchConditions(rowMap, conditions) {
			result = append(result, row)
		}
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("过滤完成：从 %d 行中筛选出 %d 行", len(dataParam), len(result)),
		Data:    result,
		Metadata: map[string]interface{}{
			"original_count": len(dataParam),
			"filtered_count": len(result),
			"conditions":     conditions,
		},
	}, nil
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符
	Value    interface{} `json:"value"`    // 比较值
}

// matchConditions 检查行是否匹配所有条件
func (t *DataProcessorTool) matchConditions(row map[string]interface{}, conditions []FilterCondition) bool {
	for _, condition := range conditions {
		fieldValue, exists := row[condition.Field]
		if !exists {
			return false
		}

		if !t.evaluateCondition(fieldValue, condition.Operator, condition.Value) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估单个条件
func (t *DataProcessorTool) evaluateCondition(fieldValue interface{}, operator string, conditionValue interface{}) bool {
	switch operator {
	case "==":
		return t.compareValues(fieldValue, conditionValue) == 0
	case "!=":
		return t.compareValues(fieldValue, conditionValue) != 0
	case ">":
		return t.compareValues(fieldValue, conditionValue) > 0
	case ">=":
		return t.compareValues(fieldValue, conditionValue) >= 0
	case "<":
		return t.compareValues(fieldValue, conditionValue) < 0
	case "<=":
		return t.compareValues(fieldValue, conditionValue) <= 0
	case "contains":
		return t.containsValue(fieldValue, conditionValue)
	case "starts_with":
		return t.startsWithValue(fieldValue, conditionValue)
	case "ends_with":
		return t.endsWithValue(fieldValue, conditionValue)
	default:
		return false
	}
}

// compareValues 比较两个值
func (t *DataProcessorTool) compareValues(a, b interface{}) int {
	aFloat, aErr := toFloat64(a)
	bFloat, bErr := toFloat64(b)

	if aErr == nil && bErr == nil {
		switch {
		case aFloat < bFloat:
			return -1
		case aFloat > bFloat:
			return 1
		default:
			return 0
		}
	}

	// 字符串比较
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

// containsValue 检查是否包含
func (t *DataProcessorTool) containsValue(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(aStr, bStr)
}

// startsWithValue 检查是否以...开头
func (t *DataProcessorTool) startsWithValue(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasPrefix(aStr, bStr)
}

// endsWithValue 检查是否以...结尾
func (t *DataProcessorTool) endsWithValue(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasSuffix(aStr, bStr)
}

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return math.NaN(), fmt.Errorf("无法转换为float64")
	}
}

// aggregateData 数据聚合
// 参数：
//   - data: 待聚合的数据（必填）
//   - group_by: 分组字段（可选）
//   - aggregations: 聚合操作（必填）
//     格式：[{"field": "age", "operation": "avg"}]
// 支持的操作：count, sum, avg, min, max, first, last
func (t *DataProcessorTool) aggregateData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	var groupByField string
	if gb, ok := params["group_by"].(string); ok {
		groupByField = gb
	}

	aggregationsParam, ok := params["aggregations"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: aggregations",
		}, nil
	}

	var aggregations []Aggregation
	for _, a := range aggregationsParam {
		aggMap, ok := a.(map[string]interface{})
		if !ok {
			continue
		}

		aggregation := Aggregation{
			Field:     aggMap["field"].(string),
			Operation: aggMap["operation"].(string),
		}
		if alias, ok := aggMap["alias"].(string); ok {
			aggregation.Alias = alias
		}
		aggregations = append(aggregations, aggregation)
	}

	// 如果有分组字段，执行分组聚合
	if groupByField != "" {
		return t.groupByAndAggregate(dataParam, groupByField, aggregations)
	}

	// 无分组，直接聚合
	result := t.performAggregations(dataParam, aggregations)

	return &DataProcessingResult{
		Success: true,
		Message: "数据聚合完成",
		Data:    result,
		Metadata: map[string]interface{}{
			"row_count":     len(dataParam),
			"aggregations":  aggregations,
		},
	}, nil
}

// Aggregation 聚合操作
type Aggregation struct {
	Field     string `json:"field"`     // 字段名
	Operation string `json:"operation"` // 操作：count, sum, avg, min, max
	Alias     string `json:"alias"`     // 别名
}

// groupByAndAggregate 分组并聚合
func (t *DataProcessorTool) groupByAndAggregate(data []interface{}, groupByField string, aggregations []Aggregation) (*DataProcessingResult, error) {
	// 按分组字段组织数据
	groups := make(map[string][]interface{})

	for _, row := range data {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		groupValue, exists := rowMap[groupByField]
		if !exists {
			continue
		}

		groupKey := fmt.Sprintf("%v", groupValue)
		groups[groupKey] = append(groups[groupKey], row)
	}

	// 对每个分组执行聚合
	result := make([]map[string]interface{}, 0)
	for groupKey, groupData := range groups {
		groupResult := t.performAggregations(groupData, aggregations)
		groupResult[groupByField] = groupKey
		result = append(result, groupResult)
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("分组聚合完成：%d 个分组", len(result)),
		Data:    result,
		Metadata: map[string]interface{}{
			"group_count":  len(result),
			"group_by":     groupByField,
			"aggregations": aggregations,
		},
	}, nil
}

// performAggregations 执行聚合操作
func (t *DataProcessorTool) performAggregations(data []interface{}, aggregations []Aggregation) map[string]interface{} {
	result := make(map[string]interface{})

	for _, agg := range aggregations {
		var alias string
		if agg.Alias != "" {
			alias = agg.Alias
		} else {
			alias = fmt.Sprintf("%s_%s", agg.Field, agg.Operation)
		}

		value := t.performAggregation(data, agg.Field, agg.Operation)
		result[alias] = value
	}

	return result
}

// performAggregation 执行单个聚合操作
func (t *DataProcessorTool) performAggregation(data []interface{}, field, operation string) interface{} {
	switch operation {
	case "count":
		return len(data)

	case "sum":
		sum := 0.0
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[field]; exists {
				if f, err := toFloat64(value); err == nil {
					sum += f
				}
			}
		}
		return sum

	case "avg":
		sum := 0.0
		count := 0
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[field]; exists {
				if f, err := toFloat64(value); err == nil {
					sum += f
					count++
				}
			}
		}
		if count > 0 {
			return sum / float64(count)
		}
		return 0.0

	case "min":
		min := math.MaxFloat64
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[field]; exists {
				if f, err := toFloat64(value); err == nil && f < min {
					min = f
				}
			}
		}
		return min

	case "max":
		max := -math.MaxFloat64
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[field]; exists {
				if f, err := toFloat64(value); err == nil && f > max {
					max = f
				}
			}
		}
		return max

	case "first":
		if len(data) > 0 {
			rowMap, _ := data[0].(map[string]interface{})
			return rowMap[field]
		}
		return nil

	case "last":
		if len(data) > 0 {
			rowMap, _ := data[len(data)-1].(map[string]interface{})
			return rowMap[field]
		}
		return nil

	default:
		return nil
	}
}

// transformData 数据转换
// 参数：
//   - data: 待转换的数据（必填）
//   - transformations: 转换规则（必填）
//     格式：[{"field": "age", "operation": "multiply", "value": 2}]
// 支持的操作：add, subtract, multiply, divide, replace, regex_replace, format
func (t *DataProcessorTool) transformData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	transformationsParam, ok := params["transformations"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: transformations",
		}, nil
	}

	var transformations []Transformation
	for _, tr := range transformationsParam {
		transMap, ok := tr.(map[string]interface{})
		if !ok {
			continue
		}

		transformation := Transformation{
			Field:     transMap["field"].(string),
			Operation: transMap["operation"].(string),
			Value:     transMap["value"],
		}
		transformations = append(transformations, transformation)
	}

	result := make([]interface{}, 0)
	for _, row := range dataParam {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			result = append(result, row)
			continue
		}

		// 复制原始行
		newRow := make(map[string]interface{})
		for k, v := range rowMap {
			newRow[k] = v
		}

		// 应用转换
		for _, transform := range transformations {
			if _, exists := newRow[transform.Field]; exists {
				newRow[transform.Field] = t.applyTransformation(newRow[transform.Field], transform)
			}
		}

		result = append(result, newRow)
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("数据转换完成：%d 行", len(result)),
		Data:    result,
		Metadata: map[string]interface{}{
			"row_count":       len(result),
			"transformations": transformations,
		},
	}, nil
}

// Transformation 转换操作
type Transformation struct {
	Field     string      `json:"field"`     // 字段名
	Operation string      `json:"operation"` // 操作：add, subtract, multiply, divide, replace等
	Value     interface{} `json:"value"`     // 转换值
}

// applyTransformation 应用转换
func (t *DataProcessorTool) applyTransformation(fieldValue interface{}, transform Transformation) interface{} {
	switch transform.Operation {
	case "add":
		if f, err := toFloat64(fieldValue); err == nil {
			if v, err := toFloat64(transform.Value); err == nil {
				return f + v
			}
		}
	case "subtract":
		if f, err := toFloat64(fieldValue); err == nil {
			if v, err := toFloat64(transform.Value); err == nil {
				return f - v
			}
		}
	case "multiply":
		if f, err := toFloat64(fieldValue); err == nil {
			if v, err := toFloat64(transform.Value); err == nil {
				return f * v
			}
		}
	case "divide":
		if f, err := toFloat64(fieldValue); err == nil {
			if v, err := toFloat64(transform.Value); err == nil && v != 0 {
				return f / v
			}
		}
	case "replace":
		if str, ok := fieldValue.(string); ok {
			if oldValue, ok := transform.Value.(map[string]interface{})["old"].(string); ok {
				if newValue, ok := transform.Value.(map[string]interface{})["new"].(string); ok {
					return strings.ReplaceAll(str, oldValue, newValue)
				}
			}
		}
	case "regex_replace":
		if str, ok := fieldValue.(string); ok {
			if pattern, ok := transform.Value.(map[string]interface{})["pattern"].(string); ok {
				if replacement, ok := transform.Value.(map[string]interface{})["replacement"].(string); ok {
					re := regexp.MustCompile(pattern)
					return re.ReplaceAllString(str, replacement)
				}
			}
		}
	case "uppercase":
		if str, ok := fieldValue.(string); ok {
			return strings.ToUpper(str)
		}
	case "lowercase":
		if str, ok := fieldValue.(string); ok {
			return strings.ToLower(str)
		}
	case "round":
		if f, err := toFloat64(fieldValue); err == nil {
			if precision, ok := transform.Value.(float64); ok {
				multiplier := math.Pow(10, precision)
				return math.Round(f*multiplier) / multiplier
			}
		}
	}

	return fieldValue
}

// mergeData 合并数据
// 参数：
//   - data1: 第一个数据集（必填）
//   - data2: 第二个数据集（必填）
//   - join_type: 连接类型（inner, left, right, full，默认inner）
//   - on: 连接字段（必填）
func (t *DataProcessorTool) mergeData(params map[string]interface{}) (*DataProcessingResult, error) {
	data1Param, ok := params["data1"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data1",
		}, nil
	}

	data2Param, ok := params["data2"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data2",
		}, nil
	}

	joinOn, ok := params["on"].(string)
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: on（连接字段）",
		}, nil
	}

	joinType := "inner"
	if jt, ok := params["join_type"].(string); ok {
		joinType = jt
	}

	// 构建索引
	index2 := t.buildIndex(data2Param, joinOn)

	result := make([]interface{}, 0)

	switch joinType {
	case "inner":
		// 内连接：只保留两边都有的
		for _, row := range data1Param {
			rowMap, _ := row.(map[string]interface{})
			key := fmt.Sprintf("%v", rowMap[joinOn])
			if matchingRows, exists := index2[key]; exists {
				for _, matchRow := range matchingRows {
					merged := t.mergeRows(rowMap, matchRow)
					result = append(result, merged)
				}
			}
		}

	case "left":
		// 左连接：保留左边所有，右边没有的填null
		for _, row := range data1Param {
			rowMap, _ := row.(map[string]interface{})
			key := fmt.Sprintf("%v", rowMap[joinOn])
			if matchingRows, exists := index2[key]; exists {
				for _, matchRow := range matchingRows {
					merged := t.mergeRows(rowMap, matchRow)
					result = append(result, merged)
				}
			} else {
				result = append(result, rowMap)
			}
		}

	case "right":
		// 右连接：保留右边所有，左边没有的填null
		index1 := t.buildIndex(data1Param, joinOn)
		for _, row := range data2Param {
			rowMap, _ := row.(map[string]interface{})
			key := fmt.Sprintf("%v", rowMap[joinOn])
			if matchingRows, exists := index1[key]; exists {
				for _, matchRow := range matchingRows {
					merged := t.mergeRows(matchRow, rowMap)
					result = append(result, merged)
				}
			} else {
				result = append(result, rowMap)
			}
		}

	case "full":
		// 全连接：保留所有
		processedKeys := make(map[string]bool)
		for _, row := range data1Param {
			rowMap, _ := row.(map[string]interface{})
			key := fmt.Sprintf("%v", rowMap[joinOn])
			processedKeys[key] = true

			if matchingRows, exists := index2[key]; exists {
				for _, matchRow := range matchingRows {
					merged := t.mergeRows(rowMap, matchRow)
					result = append(result, merged)
				}
			} else {
				result = append(result, rowMap)
			}
		}

		// 添加右边独有的行
		for _, row := range data2Param {
			rowMap, _ := row.(map[string]interface{})
			key := fmt.Sprintf("%v", rowMap[joinOn])
			if !processedKeys[key] {
				result = append(result, rowMap)
			}
		}
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("数据合并完成：%d 行（%s join）", len(result), joinType),
		Data:    result,
		Metadata: map[string]interface{}{
			"result_count": len(result),
			"data1_count":  len(data1Param),
			"data2_count":  len(data2Param),
			"join_type":    joinType,
			"join_on":      joinOn,
		},
	}, nil
}

// buildIndex 构建索引
func (t *DataProcessorTool) buildIndex(data []interface{}, field string) map[string][]map[string]interface{} {
	index := make(map[string][]map[string]interface{})

	for _, row := range data {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		key := fmt.Sprintf("%v", rowMap[field])
		index[key] = append(index[key], rowMap)
	}

	return index
}

// mergeRows 合并两行
func (t *DataProcessorTool) mergeRows(row1, row2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// 复制第一行
	for k, v := range row1 {
		merged[k] = v
	}

	// 复制第二行（冲突字段添加后缀）
	for k, v := range row2 {
		if _, exists := merged[k]; exists {
			merged[k+"_2"] = v
		} else {
			merged[k] = v
		}
	}

	return merged
}

// sortData 数据排序
// 参数：
//   - data: 待排序的数据（必填）
//   - sort_by: 排序字段（必填）
//   - order: 排序方向（asc, desc，默认asc）
func (t *DataProcessorTool) sortData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	sortBy, ok := params["sort_by"].(string)
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: sort_by",
		}, nil
	}

	order := "asc"
	if o, ok := params["order"].(string); ok {
		order = o
	}

	// 复制数据避免修改原始数据
	sorted := make([]interface{}, len(dataParam))
	copy(sorted, dataParam)

	// 排序
	sort.Slice(sorted, func(i, j int) bool {
		rowI, _ := sorted[i].(map[string]interface{})
		rowJ, _ := sorted[j].(map[string]interface{})

		cmp := t.compareValues(rowI[sortBy], rowJ[sortBy])

		if order == "desc" {
			return cmp > 0
		}
		return cmp < 0
	})

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("数据排序完成：%d 行（按 %s %s）", len(sorted), sortBy, order),
		Data:    sorted,
		Metadata: map[string]interface{}{
			"row_count": len(sorted),
			"sort_by":   sortBy,
			"order":     order,
		},
	}, nil
}

// deduplicateData 数据去重
// 参数：
//   - data: 待去重的数据（必填）
//   - deduplicate_by: 去重字段（可选，不指定则整行去重）
func (t *DataProcessorTool) deduplicateData(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	duplicateBy, hasBy := params["deduplicate_by"].(string)

	result := make([]interface{}, 0)
	seen := make(map[string]bool)

	for _, row := range dataParam {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		var key string
		if hasBy {
			key = fmt.Sprintf("%v", rowMap[duplicateBy])
		} else {
			rowJSON, _ := json.Marshal(rowMap)
			key = string(rowJSON)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, row)
		}
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("去重完成：从 %d 行中保留 %d 行唯一数据", len(dataParam), len(result)),
		Data:    result,
		Metadata: map[string]interface{}{
			"original_count": len(dataParam),
			"unique_count":   len(result),
			"removed_count":  len(dataParam) - len(result),
			"deduplicate_by": duplicateBy,
		},
	}, nil
}

// fillMissingValues 填充缺失值
// 参数：
//   - data: 待处理的数据（必填）
//   - fill_rules: 填充规则（必填）
//     格式：[{"field": "age", "strategy": "mean", "value": 0}]
// 支持的策略：mean, median, mode, forward_fill, backward_fill, value
func (t *DataProcessorTool) fillMissingValues(params map[string]interface{}) (*DataProcessingResult, error) {
	dataParam, ok := params["data"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: data（必须是数组）",
		}, nil
	}

	fillRulesParam, ok := params["fill_rules"].([]interface{})
	if !ok {
		return &DataProcessingResult{
			Success: false,
			Error:   "缺少必填参数: fill_rules",
		}, nil
	}

	var fillRules []FillRule
	for _, fr := range fillRulesParam {
		ruleMap, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}

		rule := FillRule{
			Field:    ruleMap["field"].(string),
			Strategy: ruleMap["strategy"].(string),
			Value:    ruleMap["value"],
		}
		fillRules = append(fillRules, rule)
	}

	result := make([]interface{}, 0)
	fillCount := 0

	for _, row := range dataParam {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			result = append(result, row)
			continue
		}

		// 复制原始行
		newRow := make(map[string]interface{})
		for k, v := range rowMap {
			newRow[k] = v
		}

		// 应用填充规则
		for _, rule := range fillRules {
			if value, exists := newRow[rule.Field]; exists && t.isMissing(value) {
				fillValue := t.calculateFillValue(dataParam, rule)
				newRow[rule.Field] = fillValue
				fillCount++
			}
		}

		result = append(result, newRow)
	}

	return &DataProcessingResult{
		Success: true,
		Message: fmt.Sprintf("缺失值填充完成：填充了 %d 个值", fillCount),
		Data:    result,
		Metadata: map[string]interface{}{
			"row_count":   len(result),
			"fill_count":  fillCount,
			"fill_rules":  fillRules,
		},
	}, nil
}

// FillRule 填充规则
type FillRule struct {
	Field    string      `json:"field"`    // 字段名
	Strategy string      `json:"strategy"` // 策略：mean, median, mode, value等
	Value    interface{} `json:"value"`    // 填充值（当strategy为value时使用）
}

// isMissing 检查是否为缺失值
func (t *DataProcessorTool) isMissing(value interface{}) bool {
	if value == nil {
		return true
	}
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == "" || str == "null" || str == "NULL" || str == "NA"
	}
	return false
}

// calculateFillValue 计算填充值
func (t *DataProcessorTool) calculateFillValue(data []interface{}, rule FillRule) interface{} {
	switch rule.Strategy {
	case "mean":
		sum := 0.0
		count := 0
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[rule.Field]; exists && !t.isMissing(value) {
				if f, err := toFloat64(value); err == nil {
					sum += f
					count++
				}
			}
		}
		if count > 0 {
			return sum / float64(count)
		}

	case "median":
		values := make([]float64, 0)
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[rule.Field]; exists && !t.isMissing(value) {
				if f, err := toFloat64(value); err == nil {
					values = append(values, f)
				}
			}
		}
		if len(values) > 0 {
			sort.Float64s(values)
			mid := len(values) / 2
			if len(values)%2 == 0 {
				return (values[mid-1] + values[mid]) / 2
			}
			return values[mid]
		}

	case "mode":
		frequency := make(map[string]int)
		for _, row := range data {
			rowMap, _ := row.(map[string]interface{})
			if value, exists := rowMap[rule.Field]; exists && !t.isMissing(value) {
				key := fmt.Sprintf("%v", value)
				frequency[key]++
			}
		}
		if len(frequency) > 0 {
			maxCount := 0
			var modeValue string
			for key, count := range frequency {
				if count > maxCount {
					maxCount = count
					modeValue = key
				}
			}
			return modeValue
		}

	case "forward_fill":
		// 使用前一个非缺失值
		for i := len(data) - 1; i > 0; i-- {
			rowMap, _ := data[i-1].(map[string]interface{})
			if value, exists := rowMap[rule.Field]; exists && !t.isMissing(value) {
				return value
			}
		}

	case "backward_fill":
		// 使用后一个非缺失值
		for i := 0; i < len(data)-1; i++ {
			rowMap, _ := data[i+1].(map[string]interface{})
			if value, exists := rowMap[rule.Field]; exists && !t.isMissing(value) {
				return value
			}
		}

	case "value":
		return rule.Value
	}

	return nil
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
