package chunking

import (
	"context"
	"testing"
)

// TestRecursiveCharacterChunker 测试递归字符分块器
func TestRecursiveCharacterChunker(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		chunkSize      int
		chunkOverlap   int
		minChunks      int
		maxChunks      int
		shouldPreserve bool
	}{
		{
			name:         "简单中文文本",
			text:         "这是一个测试。这是一个测试。这是一个测试。这是一个测试。",
			chunkSize:    20,
			chunkOverlap: 5,
			minChunks:    1,
			maxChunks:    5,
			shouldPreserve: true,
		},
		{
			name:         "长文本段落",
			text:         "第一段内容。第二段内容。\n\n第三段内容。第四段内容。\n\n第五段内容。第六段内容。",
			chunkSize:    30,
			chunkOverlap: 5,
			minChunks:    2,
			maxChunks:    10,
			shouldPreserve: true,
		},
		{
			name:         "短文本不需要分块",
			text:         "短文本",
			chunkSize:    100,
			chunkOverlap: 10,
			minChunks:    1,
			maxChunks:    1,
			shouldPreserve: true,
		},
		{
			name:         "英文文本",
			text:         "This is a test. This is only a test. This is a test. This is only a test.",
			chunkSize:    30,
			chunkOverlap: 5,
			minChunks:    1,
			maxChunks:    5,
			shouldPreserve: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ChunkerConfig{
				ChunkSize:     tt.chunkSize,
				ChunkOverlap:  tt.chunkOverlap,
				MinChunkSize:  tt.chunkSize / 10,
				Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?", " ", ""},
				KeepSeparator: false,
			}

			chunker, err := NewRecursiveCharacterChunker(config)
			if err != nil {
				t.Fatalf("创建分块器失败: %v", err)
			}

			chunks, err := chunker.Split(context.Background(), tt.text)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			// 验证分块数量
			if len(chunks) < tt.minChunks || len(chunks) > tt.maxChunks {
				t.Errorf("分块数量 %d 不在范围 [%d, %d] 内", len(chunks), tt.minChunks, tt.maxChunks)
			}

			// 验证内容完整性
			if tt.shouldPreserve {
				totalLength := 0
				for _, chunk := range chunks {
					totalLength += len(chunk.Content)
				}
				// 由于 overlap，总长度可能小于原文，但不应小于原文的 50%
				minExpectedLength := len(tt.text) * 50 / 100
				if totalLength < minExpectedLength {
					t.Errorf("内容长度不匹配: 原文 %d, 合并后 %d, 期望至少 %d", len(tt.text), totalLength, minExpectedLength)
				}
			}

			// 验证元数据
			for i, chunk := range chunks {
				if chunk.Metadata.Index != i {
					t.Errorf("分块 %d 的索引不正确: 期望 %d, 实际 %d", i, i, chunk.Metadata.Index)
				}
				if chunk.Metadata.ChunkType != "recursive_character" {
					t.Errorf("分块类型不正确: %s", chunk.Metadata.ChunkType)
				}
				if chunk.Content == "" {
					t.Errorf("分块 %d 内容为空", i)
				}
			}

			t.Logf("分块结果: 共 %d 个块", len(chunks))
			for i, chunk := range chunks {
				t.Logf("  块 %d: %d 字符, %d tokens, 内容预览: %.50s...",
					i, len(chunk.Content), chunk.Metadata.TokenCount, chunk.Content)
			}
		})
	}
}

// TestSmallToBigChunker 测试小到大分块器
func TestSmallToBigChunker(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		smallSize  int
		bigSize    int
		overlap    int
		parentMerge int
	}{
		{
			name:        "中等长度文本",
			text:        "第一句。第二句。第三句。第四句。第五句。第六句。第七句。第八句。第九句。第十句。",
			smallSize:   20,
			bigSize:     60,
			overlap:     5,
			parentMerge: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smallConfig := ChunkerConfig{
				ChunkSize:     tt.smallSize,
				ChunkOverlap:  tt.overlap,
				Separators:    []string{"。", "!", "?", ".", " ", ""},
				KeepSeparator: false,
			}

			bigConfig := ChunkerConfig{
				ChunkSize:     tt.bigSize,
				ChunkOverlap:  tt.overlap,
				Separators:    []string{"。", "!", "?", ".", " ", ""},
				KeepSeparator: false,
			}

			chunker, err := NewSmallToBigChunker(smallConfig, bigConfig, tt.parentMerge)
			if err != nil {
				t.Fatalf("创建分块器失败: %v", err)
			}

			chunks, err := chunker.Split(context.Background(), tt.text)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			t.Logf("小到大分块结果: 共 %d 个小块", len(chunks))
			for i, chunk := range chunks {
				parentIndex, ok := chunk.Metadata.AdditionalMetadata["parent_content"]
				if !ok {
					t.Errorf("分块 %d 缺少父块内容", i)
				} else {
					parentContent := parentIndex.(string)
					t.Logf("  块 %d: %d 字符, 父块 %d, 父块内容: %.50s...",
						i, len(chunk.Content), chunk.Metadata.ParentChunkIndex, parentContent)
				}
			}

			// 验证所有小块都有父块引用
			for i, chunk := range chunks {
				if _, ok := chunk.Metadata.AdditionalMetadata["parent_content"]; !ok {
					t.Errorf("分块 %d 缺少父块内容", i)
				}
				if chunk.Metadata.ParentChunkIndex < 0 {
					t.Errorf("分块 %d 的父块索引无效: %d", i, chunk.Metadata.ParentChunkIndex)
				}
			}
		})
	}
}

// TestParentDocumentChunker 测试父文档分块器
func TestParentDocumentChunker(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		parentSize      int
		childSize       int
		overlap         int
		childPerParent  int
	}{
		{
			name:           "长文档",
			text:           "第一句。第二句。第三句。第四句。第五句。第六句。第七句。第八句。第九句。第十句。" +
				"第十一句。第十二句。第十三句。第十四句。第十五句。第十六句。第十七句。第十八句。",
			parentSize:     80,
			childSize:      20,
			overlap:        5,
			childPerParent: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentConfig := ChunkerConfig{
				ChunkSize:     tt.parentSize,
				ChunkOverlap:  tt.overlap,
				Separators:    []string{"。", "!", "?", ".", " ", ""},
				KeepSeparator: false,
			}

			childConfig := ChunkerConfig{
				ChunkSize:     tt.childSize,
				ChunkOverlap:  tt.overlap,
				Separators:    []string{"。", "!", "?", ".", " ", ""},
				KeepSeparator: false,
			}

			chunker, err := NewParentDocumentChunker(parentConfig, childConfig, tt.childPerParent)
			if err != nil {
				t.Fatalf("创建分块器失败: %v", err)
			}

			chunks, err := chunker.Split(context.Background(), tt.text)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			t.Logf("父文档分块结果: 共 %d 个子块", len(chunks))
			for i, chunk := range chunks {
				parentContent, ok := chunk.Metadata.AdditionalMetadata["parent_content"]
				if !ok {
					t.Errorf("分块 %d 缺少父块内容", i)
				} else {
					parentStr := parentContent.(string)
					t.Logf("  块 %d: %d 字符, 父块 %d, 父块内容: %.50s...",
						i, len(chunk.Content), chunk.Metadata.ParentChunkIndex, parentStr)
				}
			}

			// 验证所有子块都有父块引用
			for i, chunk := range chunks {
				if _, ok := chunk.Metadata.AdditionalMetadata["parent_content"]; !ok {
					t.Errorf("子块 %d 缺少父块内容", i)
				}
				if _, ok := chunk.Metadata.AdditionalMetadata["parent_start_pos"]; !ok {
					t.Errorf("子块 %d 缺少父块起始位置", i)
				}
			}

			// 测试获取父块
			parentChunks, err := chunker.GetParentChunks(context.Background(), tt.text)
			if err != nil {
				t.Fatalf("获取父块失败: %v", err)
			}

			t.Logf("共有 %d 个父块", len(parentChunks))
			for i, parent := range parentChunks {
				t.Logf("  父块 %d: %d 字符, %d tokens, 内容预览: %.50s...",
					i, len(parent.Content), parent.Metadata.TokenCount, parent.Content)
			}
		})
	}
}

// TestChunkerFactory 测试分块器工厂
func TestChunkerFactory(t *testing.T) {
	factory := NewChunkerFactory()

	// 测试创建递归分块器
	t.Run("创建递归分块器", func(t *testing.T) {
		config := ChunkerConfig{
			ChunkSize:     100,
			ChunkOverlap:  20,
			Separators:    []string{"\n\n", "\n", "。", " ", ""},
			KeepSeparator: false,
		}

		chunker, err := factory.CreateChunker("recursive", config)
		if err != nil {
			t.Fatalf("创建递归分块器失败: %v", err)
		}

		if chunker.Name() != "recursive_character" {
			t.Errorf("分块器名称不正确: %s", chunker.Name())
		}
	})

	// 测试列出的分块器类型
	t.Run("列出分块器类型", func(t *testing.T) {
		types := factory.ListChunkerTypes()
		if len(types) < 4 {
			t.Errorf("分块器类型数量不足: %d", len(types))
		}

		t.Logf("支持的分块器类型:")
		for _, chunkerType := range types {
			info := factory.GetChunkerInfo(chunkerType)
			t.Logf("  - %s: %s", chunkerType, info["description"])
		}
	})

	// 测试创建小到大分块器
	t.Run("创建小到大分块器", func(t *testing.T) {
		config := map[string]interface{}{
			"small": ChunkerConfig{
				ChunkSize:    20,
				ChunkOverlap: 5,
			},
			"big": ChunkerConfig{
				ChunkSize:    60,
				ChunkOverlap: 5,
			},
			"parent_merge": 3,
		}

		chunker, err := factory.CreateChunker("small_to_big", config)
		if err != nil {
			t.Fatalf("创建小到大分块器失败: %v", err)
		}

		if chunker.Name() != "small_to_big" {
			t.Errorf("分块器名称不正确: %s", chunker.Name())
		}
	})

	// 测试创建父文档分块器
	t.Run("创建父文档分块器", func(t *testing.T) {
		config := map[string]interface{}{
			"parent": ChunkerConfig{
				ChunkSize:    100,
				ChunkOverlap: 10,
			},
			"child": ChunkerConfig{
				ChunkSize:    20,
				ChunkOverlap: 5,
			},
			"child_per_parent": 5,
		}

		chunker, err := factory.CreateChunker("parent_document", config)
		if err != nil {
			t.Fatalf("创建父文档分块器失败: %v", err)
		}

		if chunker.Name() != "parent_document" {
			t.Errorf("分块器名称不正确: %s", chunker.Name())
		}
	})
}

// TestChunkerManager 测试分块器管理器
func TestChunkerManager(t *testing.T) {
	manager := NewChunkerManager()

	// 测试创建递归分块器
	t.Run("通过管理器创建递归分块器", func(t *testing.T) {
		chunker, err := manager.CreateRecursiveChunker(100, 20)
		if err != nil {
			t.Fatalf("创建递归分块器失败: %v", err)
		}

		text := "这是一个测试。这是一个测试。这是一个测试。这是一个测试。"
		chunks, err := chunker.Split(context.Background(), text)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("分块结果为空")
		}

		t.Logf("通过管理器创建的分块器工作正常，共 %d 个块", len(chunks))
	})

	// 测试创建小到大分块器
	t.Run("通过管理器创建小到大分块器", func(t *testing.T) {
		chunker, err := manager.CreateSmallToBigChunker(20, 60, 5)
		if err != nil {
			t.Fatalf("创建小到大分块器失败: %v", err)
		}

		if chunker.Name() != "small_to_big" {
			t.Errorf("分块器名称不正确: %s", chunker.Name())
		}
	})

	// 测试列出的分块器
	t.Run("列出可用分块器", func(t *testing.T) {
		chunkers := manager.ListAvailableChunkers()
		if len(chunkers) < 4 {
			t.Errorf("可用分块器数量不足: %d", len(chunkers))
		}

		t.Logf("可用分块器:")
		for _, chunkerType := range chunkers {
			desc := manager.GetChunkerDescription(chunkerType)
			t.Logf("  - %s: %s", chunkerType, desc)
		}
	})
}

// TestChunkerConfig 测试分块器配置
func TestChunkerConfig(t *testing.T) {
	// 测试默认配置
	t.Run("默认配置", func(t *testing.T) {
		config := DefaultChunkerConfig()

		if config.ChunkSize != 500 {
			t.Errorf("默认 ChunkSize 不正确: %d", config.ChunkSize)
		}
		if config.ChunkOverlap != 50 {
			t.Errorf("默认 ChunkOverlap 不正确: %d", config.ChunkOverlap)
		}
		if len(config.Separators) == 0 {
			t.Error("默认分隔符为空")
		}

		t.Logf("默认配置: ChunkSize=%d, Overlap=%d, Separators=%d",
			config.ChunkSize, config.ChunkOverlap, len(config.Separators))
	})

	// 测试配置验证
	t.Run("配置验证", func(t *testing.T) {
		tests := []struct {
			name    string
			config  ChunkerConfig
			wantErr bool
		}{
			{
				name: "有效配置",
				config: ChunkerConfig{
					ChunkSize:    100,
					ChunkOverlap: 20,
				},
				wantErr: false,
			},
			{
				name: "ChunkSize 为 0",
				config: ChunkerConfig{
					ChunkSize:    0,
					ChunkOverlap: 20,
				},
				wantErr: true,
			},
			{
				name: "Overlap 大于 ChunkSize",
				config: ChunkerConfig{
					ChunkSize:    50,
					ChunkOverlap: 100,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				chunker, err := NewRecursiveCharacterChunker(tt.config)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewRecursiveCharacterChunker() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err == nil {
					if err := chunker.Validate(); (err != nil) != tt.wantErr {
						t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
					}
				}
			})
		}
	})
}
