package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	aitools "ai-agent-assistant/internal/tools"
)

func main() {
	fmt.Println("🔧 AI Agent Assistant - 工具系统使用示例")
	fmt.Println("==========================================")

	// 创建工具管理器
	toolManager := aitools.NewToolManager(&aitools.ToolManagerConfig{
		AutoRegister: true,
	})

	// 创建工具执行器
	executor := aitools.NewToolExecutor(toolManager)

	// 注册预定义的工具链
	chains := aitools.CreateToolChains(toolManager)
	for name, chain := range chains {
		executor.RegisterChain(chain)
		fmt.Printf("✅ 已注册工具链: %s (%d 个步骤)\n", name, len(chain.GetSteps()))
	}

	// 获取可用工具
	fmt.Println("\n📋 可用工具列表:")
	tools := toolManager.GetAvailableTools()
	for _, tool := range tools {
		fmt.Printf("   • %s (v%s) - %s\n", tool["name"], tool["version"], tool["description"])
	}

	ctx := context.Background()

	// 示例1: 文件操作 - 写入测试文件
	fmt.Println("\n📝 示例1: 文件操作 - 写入文件")
	fileResult, err := toolManager.ExecuteTool(ctx, "file_ops", "write", map[string]interface{}{
		"path":     "/tmp/test_ai_agent.txt",
		"content":  "Hello from AI Agent Assistant!\n这是测试文件。",
		"overwrite": true,
	})
	if err != nil {
		log.Printf("❌ 写入文件失败: %v", err)
	} else {
		fmt.Printf("✅ 文件写入成功: %v\n", fileResult)
	}

	// 示例2: 文件操作 - 读取文件
	fmt.Println("\n📖 示例2: 文件操作 - 读取文件")
	readResult, err := toolManager.ExecuteTool(ctx, "file_ops", "read", map[string]interface{}{
		"path": "/tmp/test_ai_agent.txt",
	})
	if err != nil {
		log.Printf("❌ 读取文件失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(readResult, "", "  ")
		fmt.Printf("✅ 文件读取成功:\n%s\n", string(resultJSON))
	}

	// 示例3: 数据处理 - CSV解析
	fmt.Println("\n📊 示例3: 数据处理 - 解析CSV")
	csvContent := `name,age,city
张三,25,北京
李四,30,上海
王五,28,深圳`

	parseResult, err := toolManager.ExecuteTool(ctx, "data_processor", "parse_csv", map[string]interface{}{
		"content":     csvContent,
		"has_header":  true,
	})
	if err != nil {
		log.Printf("❌ CSV解析失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(parseResult, "", "  ")
		fmt.Printf("✅ CSV解析成功:\n%s\n", string(resultJSON))
	}

	// 示例4: 数据处理 - 数据清洗
	fmt.Println("\n🧹 示例4: 数据处理 - 清洗数据")
	dirtyData := []interface{}{
		map[string]interface{}{"name": "  张三  ", "age": 25, "city": "北京"},
		map[string]interface{}{"name": "", "age": 30, "city": "上海"},
		map[string]interface{}{"name": "李四", "age": 0, "city": "  上海  "},
	}

	cleanResult, err := toolManager.ExecuteTool(ctx, "data_processor", "clean", map[string]interface{}{
		"data":       dirtyData,
		"operations": []string{"trim_whitespace", "remove_empty"},
	})
	if err != nil {
		log.Printf("❌ 数据清洗失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(cleanResult, "", "  ")
		fmt.Printf("✅ 数据清洗成功:\n%s\n", string(resultJSON))
	}

	// 示例5: 数据处理 - 数据过滤
	fmt.Println("\n🔍 示例5: 数据处理 - 过滤数据")
	filterData := []interface{}{
		map[string]interface{}{"name": "张三", "age": 25, "status": "active"},
		map[string]interface{}{"name": "李四", "age": 30, "status": "inactive"},
		map[string]interface{}{"name": "王五", "age": 28, "status": "active"},
	}

	filterResult, err := toolManager.ExecuteTool(ctx, "data_processor", "filter", map[string]interface{}{
		"data": filterData,
		"conditions": []interface{}{
			map[string]interface{}{
				"field":    "status",
				"operator": "==",
				"value":    "active",
			},
			map[string]interface{}{
				"field":    "age",
				"operator": ">=",
				"value":    28,
			},
		},
	})
	if err != nil {
		log.Printf("❌ 数据过滤失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(filterResult, "", "  ")
		fmt.Printf("✅ 数据过滤成功:\n%s\n", string(resultJSON))
	}

	// 示例6: 数据处理 - 数据聚合
	fmt.Println("\n📈 示例6: 数据处理 - 数据聚合")
	aggregateData := []interface{}{
		map[string]interface{}{"category": "A", "amount": 100},
		map[string]interface{}{"category": "B", "amount": 200},
		map[string]interface{}{"category": "A", "amount": 150},
		map[string]interface{}{"category": "B", "amount": 250},
	}

	aggregateResult, err := toolManager.ExecuteTool(ctx, "data_processor", "aggregate", map[string]interface{}{
		"data":    aggregateData,
		"group_by": "category",
		"aggregations": []interface{}{
			map[string]interface{}{
				"field":     "amount",
				"operation": "sum",
				"alias":     "total",
			},
			map[string]interface{}{
				"field":     "amount",
				"operation": "avg",
				"alias":     "average",
			},
		},
	})
	if err != nil {
		log.Printf("❌ 数据聚合失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(aggregateResult, "", "  ")
		fmt.Printf("✅ 数据聚合成功:\n%s\n", string(resultJSON))
	}

	// 示例7: 数据处理 - 数据排序
	fmt.Println("\n🔢 示例7: 数据处理 - 数据排序")
	sortData := []interface{}{
		map[string]interface{}{"name": "张三", "score": 85},
		map[string]interface{}{"name": "李四", "score": 92},
		map[string]interface{}{"name": "王五", "score": 78},
	}

	sortResult, err := toolManager.ExecuteTool(ctx, "data_processor", "sort", map[string]interface{}{
		"data":    sortData,
		"sort_by": "score",
		"order":   "desc",
	})
	if err != nil {
		log.Printf("❌ 数据排序失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(sortResult, "", "  ")
		fmt.Printf("✅ 数据排序成功:\n%s\n", string(resultJSON))
	}

	// 示例8: 批量操作 - 批量处理
	fmt.Println("\n⚡ 示例8: 批量操作 - 并发处理")
	items := []interface{}{"hello", "WORLD", "GoLang", "AI", "AGENT"}

	batchResult, err := toolManager.ExecuteTool(ctx, "batch_ops", "batch_process", map[string]interface{}{
		"items":       items,
		"processor":   "uppercase",
		"concurrency": 3,
	})
	if err != nil {
		log.Printf("❌ 批量处理失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(batchResult, "", "  ")
		fmt.Printf("✅ 批量处理成功:\n%s\n", string(resultJSON))
	}

	// 示例9: 获取工具能力
	fmt.Println("\n🔧 示例9: 获取工具能力")
	allCapabilities := toolManager.GetAllCapabilities()
	capabilitiesJSON, _ := json.MarshalIndent(allCapabilities, "", "  ")
	fmt.Printf("✅ 所有工具能力:\n%s\n", string(capabilitiesJSON))

	// 示例10: 工具链执行
	fmt.Println("\n🔗 示例10: 工具链执行")
	// 创建一个简单的工具链
	customChain := aitools.NewToolChain("custom_chain", toolManager)
	customChain.AddStep("data_processor", "parse_csv", map[string]interface{}{
		"content":    "name,age\nAlice,30\nBob,25",
		"has_header": true,
	}, "")
	customChain.AddStep("data_processor", "filter", map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"field":    "age",
				"operator": ">",
				"value":    25,
			},
		},
	}, "input")

	chainResult, err := customChain.Execute(ctx, nil)
	if err != nil {
		log.Printf("❌ 工具链执行失败: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(chainResult, "", "  ")
		fmt.Printf("✅ 工具链执行成功:\n%s\n", string(resultJSON))
	}

	// 示例11: 批量HTTP请求（使用模拟数据）
	fmt.Println("\n🌐 示例11: 批量HTTP请求（模拟）")
	httpResult, err := toolManager.ExecuteTool(ctx, "batch_ops", "batch_http", map[string]interface{}{
		"requests": []interface{}{
			map[string]interface{}{
				"url":    "https://httpbin.org/get",
				"method": "GET",
			},
		},
		"concurrency": 2,
		"timeout":     10,
	})
	if err != nil {
		log.Printf("⚠️  批量HTTP请求失败（可能网络问题）: %v", err)
	} else {
		resultJSON, _ := json.MarshalIndent(httpResult, "", "  ")
		fmt.Printf("✅ 批量HTTP请求完成:\n%s\n", string(resultJSON))
	}

	fmt.Println("\n==========================================")
	fmt.Println("✨ 所有示例执行完成！")
}
