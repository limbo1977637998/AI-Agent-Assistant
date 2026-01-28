package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// EntityExtractor 实体关系提取器
//
// 功能: 从文本中提取实体和关系
//
// 提取的实体类型:
//   - 人物 (Person)
//   - 组织 (Organization)
//   - 地点 (Location)
//   - 时间 (Time)
//   - 概念/技术 (Concept)
//   - 数字/数量 (Quantity)
//
// 提取的关系类型:
//   - is_a (是)
//   - part_of (属于)
//   - located_in (位于)
//   - created_by (创建于)
//   - used_for (用于)
//   - related_to (相关)
type EntityExtractor struct {
	llm LLMProvider
	config ExtractorConfig
}

// ExtractorConfig 提取器配置
type ExtractorConfig struct {
	// MaxEntities 最大实体数量
	MaxEntities int

	// MaxRelations 最大关系数量
	MaxRelations int

	// Language 语言 (zh, en)
	Language string

	// EntityTypes 要提取的实体类型
	EntityTypes []string

	// RelationTypes 要提取的关系类型
	RelationTypes []string
}

// DefaultExtractorConfig 返回默认配置
func DefaultExtractorConfig() ExtractorConfig {
	return ExtractorConfig{
		MaxEntities:  50,
		MaxRelations: 100,
		Language:     "zh",
		EntityTypes:  []string{"人物", "组织", "地点", "时间", "概念", "数量"},
		RelationTypes: []string{"是", "属于", "位于", "创建于", "用于", "相关"},
	}
}

// NewEntityExtractor 创建实体提取器
func NewEntityExtractor(llm LLMProvider, config ExtractorConfig) (*EntityExtractor, error) {
	if llm == nil {
		return nil, fmt.Errorf("LLM provider is required")
	}

	// 设置默认值
	if config.MaxEntities <= 0 {
		config.MaxEntities = 50
	}
	if config.MaxRelations <= 0 {
		config.MaxRelations = 100
	}
	if config.Language == "" {
		config.Language = "zh"
	}
	if len(config.EntityTypes) == 0 {
		config.EntityTypes = []string{"人物", "组织", "地点", "时间", "概念", "数量"}
	}

	return &EntityExtractor{
		llm:    llm,
		config: config,
	}, nil
}

// Extract 从文本中提取实体和关系
func (ee *EntityExtractor) Extract(ctx context.Context, text string) (*GraphData, error) {
	// 1. 提取实体
	entities, err := ee.extractEntities(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %w", err)
	}

	// 2. 提取关系
	relations, err := ee.extractRelations(ctx, text, entities)
	if err != nil {
		return nil, fmt.Errorf("failed to extract relations: %w", err)
	}

	return &GraphData{
		Entities:  entities,
		Relations: relations,
	}, nil
}

// extractEntities 提取实体
func (ee *EntityExtractor) extractEntities(ctx context.Context, text string) ([]*Entity, error) {
	prompt := ee.buildEntityExtractionPrompt(text)

	response, err := ee.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析实体
	entities, err := ee.parseEntities(response)
	if err != nil {
		// 如果解析失败，尝试简单提取
		return ee.extractSimpleEntities(text), nil
	}

	// 限制数量
	if len(entities) > ee.config.MaxEntities {
		entities = entities[:ee.config.MaxEntities]
	}

	return entities, nil
}

// extractRelations 提取关系
func (ee *EntityExtractor) extractRelations(ctx context.Context, text string, entities []*Entity) ([]*Relation, error) {
	prompt := ee.buildRelationExtractionPrompt(text, entities)

	response, err := ee.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析关系
	relations, err := ee.parseRelations(response, entities)
	if err != nil {
		// 如果解析失败，返回空关系
		return []*Relation{}, nil
	}

	// 限制数量
	if len(relations) > ee.config.MaxRelations {
		relations = relations[:ee.config.MaxRelations]
	}

	return relations, nil
}

// buildEntityExtractionPrompt 构建实体提取提示
func (ee *EntityExtractor) buildEntityExtractionPrompt(text string) string {
	if ee.config.Language == "zh" {
		return ee.buildChineseEntityPrompt(text)
	}
	return ee.buildEnglishEntityPrompt(text)
}

// buildChineseEntityPrompt 构建中文实体提取提示
func (ee *EntityExtractor) buildChineseEntityPrompt(text string) string {
	entityTypes := strings.Join(ee.config.EntityTypes, "、")

	return fmt.Sprintf(`你是一个专业的实体识别专家。请从以下文本中提取实体。

要求:
1. 实体类型包括: %s
2. 提取的实体应该是重要的、有意义的
3. 每个实体应该有明确的类型
4. 以 JSON 格式输出: {"entities": [{"name": "实体名", "type": "类型", "description": "描述"}]}

文本:
%s

提取的实体 (JSON):`, entityTypes, text)
}

// buildEnglishEntityPrompt 构建英文实体提取提示
func (ee *EntityExtractor) buildEnglishEntityPrompt(text string) string {
	entityTypes := strings.Join(ee.config.EntityTypes, ", ")

	return fmt.Sprintf(`You are a professional entity extraction expert. Please extract entities from the following text.

Requirements:
1. Entity types include: %s
2. Extract important and meaningful entities
3. Each entity should have a clear type
4. Output in JSON format: {"entities": [{"name": "Entity Name", "type": "Type", "description": "Description"}]}

Text:
%s

Extracted entities (JSON):`, entityTypes, text)
}

// parseEntities 解析实体
func (ee *EntityExtractor) parseEntities(response string) ([]*Entity, error) {
	response = strings.TrimSpace(response)

	// 尝试解析 JSON
	var result struct {
		Entities []*Entity `json:"entities"`
	}

	if err := json.Unmarshal([]byte(response), &result); err == nil {
		return result.Entities, nil
	}

	// JSON 解析失败，返回空
	return nil, fmt.Errorf("failed to parse entities")
}

// extractSimpleEntities 简单的实体提取（基于规则）
func (ee *EntityExtractor) extractSimpleEntities(text string) []*Entity {
	entities := make([]*Entity, 0)

	// 简单规则：提取大写开头的词（英文）或引号包裹的内容（中文）
	words := strings.Fields(text)
	for i, word := range words {
		// 跳过太短的词
		if len(word) < 2 {
			continue
		}

		// 英文实体：大写开头
		if word[0] >= 'A' && word[0] <= 'Z' {
			entities = append(entities, &Entity{
				ID:          fmt.Sprintf("entity_%d", i),
				Name:        word,
				Type:        "概念",
				Description: "",
			})
		}
	}

	return entities
}

// buildRelationExtractionPrompt 构建关系提取提示
func (ee *EntityExtractor) buildRelationExtractionPrompt(text string, entities []*Entity) string {
	if ee.config.Language == "zh" {
		return ee.buildChineseRelationPrompt(text, entities)
	}
	return ee.buildEnglishRelationPrompt(text, entities)
}

// buildChineseRelationPrompt 构建中文关系提取提示
func (ee *EntityExtractor) buildChineseRelationPrompt(text string, entities []*Entity) string {
	// 构建实体列表
	entityList := ""
	for _, e := range entities {
		entityList += fmt.Sprintf("- %s (%s)\n", e.Name, e.Type)
	}

	return fmt.Sprintf(`你是一个专业的关系抽取专家。请从文本中提取实体之间的关系。

实体列表:
%s

要求:
1. 识别实体之间的关系
2. 关系应该是明确的、有意义的
3. 以 JSON 格式输出: {"relations": [{"from": "实体1", "to": "实体2", "type": "关系类型", "description": "描述"}]}

文本:
%s

提取的关系 (JSON):`, entityList, text)
}

// buildEnglishRelationPrompt 构建英文关系提取提示
func (ee *EntityExtractor) buildEnglishRelationPrompt(text string, entities []*Entity) string {
	entityList := ""
	for _, e := range entities {
		entityList += fmt.Sprintf("- %s (%s)\n", e.Name, e.Type)
	}

	return fmt.Sprintf(`You are a professional relation extraction expert. Please extract relationships between entities from the text.

Entity list:
%s

Requirements:
1. Identify relationships between entities
2. Relationships should be clear and meaningful
3. Output in JSON format: {"relations": [{"from": "Entity1", "to": "Entity2", "type": "Relation Type", "description": "Description"}]}

Text:
%s

Extracted relations (JSON):`, entityList, text)
}

// parseRelations 解析关系
func (ee *EntityExtractor) parseRelations(response string, entities []*Entity) ([]*Relation, error) {
	response = strings.TrimSpace(response)

	// 尝试解析 JSON
	var result struct {
		Relations []*Relation `json:"relations"`
	}

	if err := json.Unmarshal([]byte(response), &result); err == nil {
		// 验证实体是否存在
		entityMap := make(map[string]bool)
		for _, e := range entities {
			entityMap[e.Name] = true
		}

		// 过滤无效的关系
		validRelations := make([]*Relation, 0)
		for _, r := range result.Relations {
			if entityMap[r.From] && entityMap[r.To] {
				r.ID = fmt.Sprintf("rel_%s_%s", r.From, r.To)
				validRelations = append(validRelations, r)
			}
		}

		return validRelations, nil
	}

	// JSON 解析失败
	return nil, fmt.Errorf("failed to parse relations")
}

// LLMProvider LLM 提供者接口
type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Entity 实体
type Entity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Relation 关系
type Relation struct {
	ID          string `json:"id"`
	From        string `json:"from"`        // 源实体
	To          string `json:"to"`          // 目标实体
	Type        string `json:"type"`        // 关系类型
	Description string `json:"description"` // 描述
}

// GraphData 图数据
type GraphData struct {
	Entities  []*Entity   `json:"entities"`
	Relations []*Relation `json:"relations"`
}
