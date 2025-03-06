package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/user/cppjudge/internal/models"
)

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// QuestionTemplate 表示问题模板
type QuestionTemplate struct {
	Type        string
	Format      string
	Placeholder string
}

// 问题模板集合
var multipleChoiceTemplates = []QuestionTemplate{
	{
		Type:   models.QuestionTypeMultipleChoice,
		Format: "以下关于%s的描述，哪一项是正确的？",
	},
	{
		Type:   models.QuestionTypeMultipleChoice,
		Format: "在学习%s时，下列哪一项说法是正确的？",
	},
	{
		Type:   models.QuestionTypeMultipleChoice,
		Format: "对于%s，以下哪个选项是正确的？",
	},
}

var fillBlankTemplates = []QuestionTemplate{
	{
		Type:        models.QuestionTypeFillBlank,
		Format:      "在%s中，_____是关键的概念或操作。",
		Placeholder: "_____",
	},
	{
		Type:        models.QuestionTypeFillBlank,
		Format:      "%s的主要特点是_____。",
		Placeholder: "_____",
	},
	{
		Type:        models.QuestionTypeFillBlank,
		Format:      "在处理%s时，我们通常使用_____方法。",
		Placeholder: "_____",
	},
}

var shortAnswerTemplates = []QuestionTemplate{
	{
		Type:   models.QuestionTypeShortAnswer,
		Format: "简述%s的基本概念和应用场景。",
	},
	{
		Type:   models.QuestionTypeShortAnswer,
		Format: "描述%s的主要特点以及实现方法。",
	},
	{
		Type:   models.QuestionTypeShortAnswer,
		Format: "请解释%s的工作原理，并给出一个简单的例子。",
	},
}

// GenerateQuestionFromOutlineItem 根据大纲条目生成题目
func GenerateQuestionFromOutlineItem(item OutlineItem) models.OutlineQuestion {
	// 根据难度选择问题类型
	questionType := selectQuestionType(item.Difficulty)

	// 生成问题内容
	content, options, answer, explanation := "", []string{}, "", ""

	switch questionType {
	case models.QuestionTypeMultipleChoice:
		content, options, answer, explanation = generateMultipleChoiceQuestion(item)
	case models.QuestionTypeFillBlank:
		content, answer, explanation = generateFillBlankQuestion(item)
	case models.QuestionTypeShortAnswer:
		content, answer, explanation = generateShortAnswerQuestion(item)
	}

	// 创建问题
	question := models.OutlineQuestion{
		Type:         questionType,
		Difficulty:   item.Difficulty,
		Content:      content,
		Options:      options,
		Answer:       answer,
		Explanation:  explanation,
		KnowledgeTag: item.Tags,
		OutlineRef:   item.Section,
		CreatedAt:    time.Now(),
	}

	return question
}

// GenerateQuestionsFromOutline 根据大纲生成多个题目
func GenerateQuestionsFromOutline(items []OutlineItem, count int) []models.OutlineQuestion {
	// 如果需要的题目数量大于条目数量，则每个条目生成一个题目
	if count >= len(items) {
		questions := make([]models.OutlineQuestion, len(items))
		for i, item := range items {
			questions[i] = GenerateQuestionFromOutlineItem(item)
		}
		return questions
	}

	// 随机选择条目生成题目
	selectedIndices := rand.Perm(len(items))[:count]
	questions := make([]models.OutlineQuestion, count)

	for i, idx := range selectedIndices {
		questions[i] = GenerateQuestionFromOutlineItem(items[idx])
	}

	return questions
}

// GenerateQuizFromOutlineSection 根据大纲章节生成测试
func GenerateQuizFromOutlineSection(items []OutlineItem, section string, questionCount int) models.Quiz {
	// 过滤出指定章节的条目
	var sectionItems []OutlineItem
	for _, item := range items {
		if strings.HasPrefix(item.Section, section) {
			sectionItems = append(sectionItems, item)
		}
	}

	// 生成问题
	questions := GenerateQuestionsFromOutline(sectionItems, questionCount)

	// 提取问题ID
	questionIDs := make([]int, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	// 提取标签
	tagMap := make(map[string]bool)
	for _, q := range questions {
		for _, tag := range q.KnowledgeTag {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	// 创建测试
	quiz := models.Quiz{
		Title:        fmt.Sprintf("%s 知识点测试", section),
		Description:  fmt.Sprintf("这是关于%s章节知识点的测试，共包含%d道题目", section, len(questions)),
		QuestionIDs:  questionIDs,
		KnowledgeTag: tags,
		CreatedAt:    time.Now(),
	}

	return quiz
}

// 根据难度选择问题类型
func selectQuestionType(difficulty int) string {
	if difficulty <= 3 {
		// 低难度更多选择题
		types := []string{
			models.QuestionTypeMultipleChoice,
			models.QuestionTypeMultipleChoice,
			models.QuestionTypeFillBlank,
		}
		return types[rand.Intn(len(types))]
	} else if difficulty <= 6 {
		// 中等难度平衡分布
		types := []string{
			models.QuestionTypeMultipleChoice,
			models.QuestionTypeFillBlank,
			models.QuestionTypeFillBlank,
			models.QuestionTypeShortAnswer,
		}
		return types[rand.Intn(len(types))]
	} else {
		// 高难度更多简答题
		types := []string{
			models.QuestionTypeMultipleChoice,
			models.QuestionTypeFillBlank,
			models.QuestionTypeShortAnswer,
			models.QuestionTypeShortAnswer,
		}
		return types[rand.Intn(len(types))]
	}
}

// 生成选择题
func generateMultipleChoiceQuestion(item OutlineItem) (content string, options []string, answer string, explanation string) {
	// 选择一个模板
	template := multipleChoiceTemplates[rand.Intn(len(multipleChoiceTemplates))]

	// 生成问题内容
	content = fmt.Sprintf(template.Format, item.Knowledge)

	// 生成选项和答案
	correctIndex := rand.Intn(4) // 0-3之间随机选择正确答案的位置
	options = make([]string, 4)

	// 根据知识点构造选项
	options[correctIndex] = generateCorrectOption(item)

	// 生成错误选项
	for i := 0; i < 4; i++ {
		if i != correctIndex {
			options[i] = generateIncorrectOption(item, options[:i])
		}
	}

	// 设置答案 (A, B, C, D)
	answer = string('A' + correctIndex)

	// 生成解释
	explanation = generateExplanation(item, options[correctIndex])

	return content, options, answer, explanation
}

// 生成填空题
func generateFillBlankQuestion(item OutlineItem) (content string, answer string, explanation string) {
	// 选择一个模板
	template := fillBlankTemplates[rand.Intn(len(fillBlankTemplates))]

	// 生成问题内容和答案
	content = fmt.Sprintf(template.Format, item.Knowledge)
	answer = extractKeyword(item.Knowledge)

	// 生成解释
	explanation = generateExplanation(item, answer)

	return content, answer, explanation
}

// 生成简答题
func generateShortAnswerQuestion(item OutlineItem) (content string, answer string, explanation string) {
	// 选择一个模板
	template := shortAnswerTemplates[rand.Intn(len(shortAnswerTemplates))]

	// 生成问题内容
	content = fmt.Sprintf(template.Format, item.Knowledge)

	// 生成参考答案
	answer = generateReferenceAnswer(item)

	// 生成解释，简答题通常不需要额外解释
	explanation = "参考上述答案，根据实际情况可能有多种正确表述。"

	return content, answer, explanation
}

// 生成正确选项
func generateCorrectOption(item OutlineItem) string {
	// 这里应该根据知识点生成正确的描述
	// 为了演示，我们使用一些通用的模板
	templates := []string{
		"%s是%s的核心概念。",
		"%s可以用于解决%s相关的问题。",
		"在%s中，%s是一种常用的方法。",
		"%s是实现%s的重要基础。",
	}

	template := templates[rand.Intn(len(templates))]
	keywords := extractKeywords(item.Knowledge)

	if len(keywords) >= 2 {
		return fmt.Sprintf(template, keywords[0], keywords[1])
	} else if len(keywords) == 1 {
		return fmt.Sprintf(template, keywords[0], item.Knowledge)
	} else {
		return fmt.Sprintf(template, item.Knowledge, "相关算法")
	}
}

// 生成错误选项
func generateIncorrectOption(item OutlineItem, existingOptions []string) string {
	// 这里应该根据知识点生成错误的描述，且与其他选项不重复
	// 为了演示，我们使用一些通用的错误描述
	templates := []string{
		"%s与%s没有直接关系。",
		"%s不能用于解决%s问题。",
		"在%s中，%s是一个错误的概念。",
		"%s不是%s的必要条件。",
		"%s通常与%s相反。",
	}

	// 尝试生成直到找到一个不重复的选项
	for attempts := 0; attempts < 10; attempts++ {
		template := templates[rand.Intn(len(templates))]
		keywords := extractKeywords(item.Knowledge)

		var option string
		if len(keywords) >= 2 {
			option = fmt.Sprintf(template, keywords[0], keywords[1])
		} else if len(keywords) == 1 {
			option = fmt.Sprintf(template, keywords[0], item.Knowledge)
		} else {
			option = fmt.Sprintf(template, item.Knowledge, "常见算法")
		}

		// 检查是否与已有选项重复
		duplicate := false
		for _, existing := range existingOptions {
			if existing == option {
				duplicate = true
				break
			}
		}

		if !duplicate {
			return option
		}
	}

	// 如果多次尝试都无法生成不重复的选项，返回一个带随机数的选项
	return fmt.Sprintf("选项 %d：这不是一个正确的描述。", rand.Intn(100))
}

// 生成解释
func generateExplanation(item OutlineItem, correctOption string) string {
	// 根据难度和知识点生成解释
	switch item.Difficulty {
	case 1, 2, 3:
		return fmt.Sprintf("正确答案是：%s 根据大纲中关于%s的描述，%s", correctOption, item.Knowledge, correctOption)
	case 4, 5, 6:
		return fmt.Sprintf("正确答案是：%s 在%s中，%s 这是因为%s的特性决定了它的使用方式。", correctOption, item.Title, correctOption, item.Knowledge)
	default:
		return fmt.Sprintf("正确答案是：%s 在高级的%s应用中，%s 深入理解这一点对掌握%s至关重要。", correctOption, item.Title, correctOption, item.Knowledge)
	}
}

// 生成参考答案
func generateReferenceAnswer(item OutlineItem) string {
	// 根据难度和知识点生成参考答案
	switch item.Difficulty {
	case 1, 2, 3:
		return fmt.Sprintf("%s是指%s。它的主要特点是简单易用，常见于基础编程中。", item.Knowledge, extractDefinition(item.Knowledge))
	case 4, 5, 6:
		return fmt.Sprintf("%s是%s的一个重要概念。它的主要特点包括：\n1. 高效性\n2. 灵活性\n3. 适用于复杂场景\n\n在实际应用中，通常需要结合具体问题进行分析和使用。",
			item.Knowledge, item.Title)
	default:
		return fmt.Sprintf("%s是%s中的高级概念。\n\n它的核心原理是：通过特定的数据结构和算法组合，解决复杂的计算问题。\n\n应用场景包括：\n1. 大规模数据处理\n2. 高性能计算\n3. 复杂系统优化\n\n实现时需要注意效率和正确性的平衡。",
			item.Knowledge, item.Title)
	}
}

// 提取关键词
func extractKeywords(text string) []string {
	// 简单地按照标点符号和空格分割文本
	separators := []string{"：", "、", "，", ",", " ", "（", "）", "(", ")"}

	result := []string{text}
	for _, sep := range separators {
		var newResult []string
		for _, part := range result {
			splits := strings.Split(part, sep)
			for _, s := range splits {
				if s != "" {
					newResult = append(newResult, s)
				}
			}
		}
		result = newResult
	}

	// 过滤掉太短的词
	var filtered []string
	for _, word := range result {
		if len(word) > 1 {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// 提取单个关键词
func extractKeyword(text string) string {
	keywords := extractKeywords(text)
	if len(keywords) > 0 {
		return keywords[rand.Intn(len(keywords))]
	}
	return text
}

// 提取定义（简化版，实际应用中可能需要更复杂的逻辑）
func extractDefinition(text string) string {
	if strings.Contains(text, "：") {
		parts := strings.Split(text, "：")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	return fmt.Sprintf("与%s相关的一种概念或技术", text)
}
