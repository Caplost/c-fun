package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/cppjudge/internal/config"
	"github.com/user/cppjudge/internal/models"
)

// ProblemGenerationRequest 表示生成问题的请求
type ProblemGenerationRequest struct {
	Title                    string   `json:"title"`                      // 题目名称/主题
	OutlineSection           string   `json:"outline_section"`            // 大纲章节
	KnowledgePoints          []string `json:"knowledge_points"`           // 所需知识点
	Difficulty               string   `json:"difficulty"`                 // 难度级别
	ProblemType              string   `json:"problem_type"`               // 问题类型
	TimeComplexity           string   `json:"time_complexity"`            // 时间复杂度要求
	SpaceComplexity          string   `json:"space_complexity"`           // 空间复杂度要求
	AdditionalReqs           string   `json:"additional_reqs"`            // 额外要求
	TestCaseCount            int      `json:"test_case_count"`            // 测试用例数
	IncludeReferenceSolution bool     `json:"include_reference_solution"` // 是否包含参考解答
	IncludeAnalysis          bool     `json:"include_analysis"`           // 是否包含思维训练分析
}

// GeneratedProblem 表示生成的问题
type GeneratedProblem struct {
	models.Problem                      // 嵌入基本题目信息
	TestCases         []models.TestCase `json:"test_cases"`                   // 测试用例
	ReferenceSolution string            `json:"reference_solution,omitempty"` // 参考解答
	ThinkingAnalysis  string            `json:"thinking_analysis,omitempty"`  // 思维训练分析
}

// ParseGeneratedProblem 从LLM输出解析出生成的题目
func ParseGeneratedProblem(content string) (*GeneratedProblem, error) {
	log.Printf("开始解析生成的问题，内容长度: %d", len(content))

	// 尝试直接解析JSON格式的输出
	var problem GeneratedProblem
	err := json.Unmarshal([]byte(content), &problem)
	if err == nil && problem.Title != "" {
		log.Printf("成功直接解析JSON: 标题=%s, 描述长度=%d", problem.Title, len(problem.Description))
		return &problem, nil
	}

	log.Printf("直接JSON解析失败: %v, 尝试提取JSON部分", err)

	// 尝试查找并提取JSON部分
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := content[jsonStart : jsonEnd+1]
		log.Printf("找到JSON部分，长度: %d", len(jsonStr))

		err = json.Unmarshal([]byte(jsonStr), &problem)
		if err == nil && problem.Title != "" {
			log.Printf("成功解析提取的JSON: 标题=%s, 描述长度=%d", problem.Title, len(problem.Description))
			return &problem, nil
		}
		log.Printf("解析提取的JSON失败: %v", err)

		// 尝试不同的JSON结构
		var alternateFormat struct {
			Title             string            `json:"title"`
			Description       string            `json:"description"`
			Difficulty        string            `json:"difficulty"`
			TimeLimit         int               `json:"time_limit"`
			MemoryLimit       int               `json:"memory_limit"`
			KnowledgeTag      []string          `json:"knowledge_tag"`
			TestCases         []models.TestCase `json:"test_cases"`
			ReferenceSolution string            `json:"reference_solution"`
			ThinkingAnalysis  string            `json:"thinking_analysis"`
		}

		err = json.Unmarshal([]byte(jsonStr), &alternateFormat)
		if err == nil && alternateFormat.Title != "" {
			log.Printf("成功解析替代JSON格式: 标题=%s, 描述长度=%d", alternateFormat.Title, len(alternateFormat.Description))

			// 手动构建GeneratedProblem结构
			problem = GeneratedProblem{
				Problem: models.Problem{
					Title:        alternateFormat.Title,
					Description:  alternateFormat.Description,
					Difficulty:   alternateFormat.Difficulty,
					TimeLimit:    alternateFormat.TimeLimit,
					MemoryLimit:  alternateFormat.MemoryLimit,
					KnowledgeTag: alternateFormat.KnowledgeTag,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
				TestCases:         alternateFormat.TestCases,
				ReferenceSolution: alternateFormat.ReferenceSolution,
				ThinkingAnalysis:  alternateFormat.ThinkingAnalysis,
			}

			return &problem, nil
		}
		log.Printf("解析替代JSON格式失败: %v", err)
	}

	log.Printf("尝试解析结构化文本")
	// 如果仍无法解析，尝试解析结构化文本
	return parseStructuredText(content)
}

// parseStructuredText 从结构化文本中解析出题目信息
func parseStructuredText(content string) (*GeneratedProblem, error) {
	problem := &GeneratedProblem{
		Problem: models.Problem{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// 分割内容为各个部分
	sections := splitIntoSections(content)

	// 解析标题
	if title, ok := sections["标题"]; ok && title != "" {
		problem.Title = strings.TrimSpace(title)
	} else {
		return nil, fmt.Errorf("无法解析题目标题")
	}

	// 解析描述
	if desc, ok := sections["描述"]; ok && desc != "" {
		problem.Description = strings.TrimSpace(desc)
	} else if desc, ok := sections["题目描述"]; ok && desc != "" {
		problem.Description = strings.TrimSpace(desc)
	} else {
		return nil, fmt.Errorf("无法解析题目描述")
	}

	// 解析难度
	if diff, ok := sections["难度"]; ok {
		problem.Difficulty = normalizeDifficulty(diff)
	} else {
		problem.Difficulty = "Medium" // 默认中等难度
	}

	// 解析时间限制
	if tl, ok := sections["时间限制"]; ok {
		problem.TimeLimit = parseTimeLimit(tl)
	} else {
		problem.TimeLimit = 1000 // 默认1秒
	}

	// 解析内存限制
	if ml, ok := sections["内存限制"]; ok {
		problem.MemoryLimit = parseMemoryLimit(ml)
	} else {
		problem.MemoryLimit = 256 * 1024 // 默认256MB
	}

	// 解析知识点标签
	if tags, ok := sections["知识点"]; ok {
		problem.KnowledgeTag = parseKnowledgeTags(tags)
	}

	// 解析测试用例
	testCases := make([]models.TestCase, 0)
	if tc, ok := sections["测试用例"]; ok {
		parsedTestCases := parseTestCases(tc)
		for _, ptc := range parsedTestCases {
			testCases = append(testCases, models.TestCase{
				Input:     ptc.Input,
				Output:    ptc.Output,
				IsExample: true,
			})
		}
	}
	problem.TestCases = testCases

	// 解析参考解答
	if solution, ok := sections["参考解答"]; ok {
		problem.ReferenceSolution = solution
	} else if solution, ok := sections["解答"]; ok {
		problem.ReferenceSolution = solution
	}

	// 解析思维训练分析
	if analysis, ok := sections["思维训练分析"]; ok {
		problem.ThinkingAnalysis = analysis
	}

	return problem, nil
}

// 辅助结构和函数
type parsedTestCase struct {
	Input  string
	Output string
}

// splitIntoSections 将文本分割为各个章节
func splitIntoSections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	currentSection := ""
	currentContent := []string{}

	log.Printf("开始分析响应内容，总行数: %d", len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否为新章节
		isNewSection := false

		// 检查常见的章节标记
		if strings.HasPrefix(line, "#") {
			isNewSection = true
		} else if strings.HasPrefix(line, "标题:") || strings.HasPrefix(line, "题目标题:") {
			line = "标题"
			isNewSection = true
		} else if strings.HasPrefix(line, "描述:") || strings.HasPrefix(line, "题目描述:") {
			line = "描述"
			isNewSection = true
		} else if strings.HasPrefix(line, "难度:") || strings.HasPrefix(line, "题目难度:") {
			line = "难度"
			isNewSection = true
		} else if strings.HasPrefix(line, "时间限制:") {
			line = "时间限制"
			isNewSection = true
		} else if strings.HasPrefix(line, "内存限制:") {
			line = "内存限制"
			isNewSection = true
		} else if strings.HasPrefix(line, "知识点:") || strings.HasPrefix(line, "标签:") {
			line = "知识点"
			isNewSection = true
		} else if strings.HasPrefix(line, "测试用例:") || strings.HasPrefix(line, "样例:") {
			line = "测试用例"
			isNewSection = true
		} else if strings.HasPrefix(line, "参考解答:") || strings.HasPrefix(line, "解答:") {
			line = "参考解答"
			isNewSection = true
		} else if strings.HasPrefix(line, "思维训练:") || strings.HasPrefix(line, "思维训练分析:") {
			line = "思维训练分析"
			isNewSection = true
		}

		if isNewSection {
			// 保存之前的章节
			if currentSection != "" && len(currentContent) > 0 {
				sectionContent := strings.Join(currentContent, "\n")
				log.Printf("找到章节 '%s', 内容长度: %d", currentSection, len(sectionContent))
				sections[currentSection] = sectionContent
				currentContent = []string{}
			}

			// 提取新章节名称
			if strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) > 1 {
					currentSection = strings.TrimSpace(parts[1])
				}
			} else {
				parts := strings.SplitN(line, ":", 2)
				currentSection = strings.TrimSpace(parts[0])
				if len(parts) > 1 && parts[1] != "" {
					currentContent = append(currentContent, strings.TrimSpace(parts[1]))
				}
			}

			log.Printf("行 %d: 新章节开始: %s", i+1, currentSection)
		} else {
			// 添加到当前章节
			currentContent = append(currentContent, line)
		}
	}

	// 保存最后一个章节
	if currentSection != "" && len(currentContent) > 0 {
		sectionContent := strings.Join(currentContent, "\n")
		log.Printf("找到章节 '%s', 内容长度: %d", currentSection, len(sectionContent))
		sections[currentSection] = sectionContent
	}

	// 输出所有找到的章节
	for section := range sections {
		log.Printf("解析完成，章节: %s", section)
	}

	return sections
}

// normalizeDifficulty 规范化难度级别
func normalizeDifficulty(diff string) string {
	diff = strings.ToLower(strings.TrimSpace(diff))
	if strings.Contains(diff, "简单") || strings.Contains(diff, "easy") {
		return "Easy"
	} else if strings.Contains(diff, "困难") || strings.Contains(diff, "hard") {
		return "Hard"
	} else {
		return "Medium"
	}
}

// parseTimeLimit 解析时间限制
func parseTimeLimit(tl string) int {
	tl = strings.TrimSpace(tl)
	if strings.Contains(tl, "ms") {
		tl = strings.ReplaceAll(tl, "ms", "")
		tl = strings.TrimSpace(tl)
		if val, err := parsePositiveInt(tl); err == nil {
			return val
		}
	} else if strings.Contains(tl, "秒") || strings.Contains(tl, "s") {
		tl = strings.ReplaceAll(tl, "秒", "")
		tl = strings.ReplaceAll(tl, "s", "")
		tl = strings.TrimSpace(tl)
		if val, err := parsePositiveInt(tl); err == nil {
			return val * 1000
		}
	}
	return 1000 // 默认1秒
}

// parseMemoryLimit 解析内存限制
func parseMemoryLimit(ml string) int {
	ml = strings.TrimSpace(ml)
	if strings.Contains(ml, "KB") || strings.Contains(ml, "kb") {
		ml = strings.ReplaceAll(ml, "KB", "")
		ml = strings.ReplaceAll(ml, "kb", "")
		ml = strings.TrimSpace(ml)
		if val, err := parsePositiveInt(ml); err == nil {
			return val
		}
	} else if strings.Contains(ml, "MB") || strings.Contains(ml, "mb") {
		ml = strings.ReplaceAll(ml, "MB", "")
		ml = strings.ReplaceAll(ml, "mb", "")
		ml = strings.TrimSpace(ml)
		if val, err := parsePositiveInt(ml); err == nil {
			return val * 1024
		}
	}
	return 256 * 1024 // 默认256MB
}

// parseKnowledgeTags 解析知识点标签
func parseKnowledgeTags(tags string) []string {
	tags = strings.TrimSpace(tags)
	var result []string
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// parseTestCases 解析测试用例
func parseTestCases(tc string) []parsedTestCase {
	lines := strings.Split(tc, "\n")
	var cases []parsedTestCase
	var currentCase parsedTestCase
	isInput := true

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "输入") || strings.HasPrefix(line, "Input") {
			if currentCase.Input != "" {
				cases = append(cases, currentCase)
				currentCase = parsedTestCase{}
			}
			isInput = true
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
					currentCase.Input = strings.TrimSpace(parts[1])
				}
			}
		} else if strings.HasPrefix(line, "输出") || strings.HasPrefix(line, "Output") {
			isInput = false
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
					currentCase.Output = strings.TrimSpace(parts[1])
				}
			}
		} else {
			if isInput {
				if currentCase.Input == "" {
					currentCase.Input = line
				} else {
					currentCase.Input += "\n" + line
				}
			} else {
				if currentCase.Output == "" {
					currentCase.Output = line
				} else {
					currentCase.Output += "\n" + line
				}
			}
		}
	}

	if currentCase.Input != "" || currentCase.Output != "" {
		cases = append(cases, currentCase)
	}

	// 如果没有明确的输入/输出分隔，尝试按对分割
	if len(cases) == 0 && len(lines) >= 2 {
		for i := 0; i < len(lines); i += 2 {
			if i+1 < len(lines) {
				cases = append(cases, parsedTestCase{
					Input:  strings.TrimSpace(lines[i]),
					Output: strings.TrimSpace(lines[i+1]),
				})
			}
		}
	}

	return cases
}

// parsePositiveInt 解析正整数
func parsePositiveInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, fmt.Errorf("值必须为正数")
	}
	return val, nil
}

// ProblemGenerator 问题生成器
type ProblemGenerator struct {
	llmClient  Client
	configPath string
}

// NewProblemGenerator 创建一个新的问题生成器
func NewProblemGenerator(configPath string) (*ProblemGenerator, error) {
	// 加载LLM配置
	llmConfig, err := config.LoadLLMConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载LLM配置失败: %w", err)
	}

	// 创建LLM客户端
	llmClient, err := Factory(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	return &ProblemGenerator{
		llmClient:  llmClient,
		configPath: configPath,
	}, nil
}

// GenerateProblem 生成编程问题
func (g *ProblemGenerator) GenerateProblem(req ProblemGenerationRequest) (*GeneratedProblem, error) {
	// 加载LLM配置获取系统提示词
	llmConfig, err := config.LoadLLMConfig(g.configPath)
	if err != nil {
		return nil, fmt.Errorf("加载LLM配置失败: %w", err)
	}

	// 准备消息
	messages := []Message{
		{
			Role:    "system",
			Content: llmConfig.SystemPrompt,
		},
		{
			Role:    "user",
			Content: buildPrompt(req),
		},
	}

	// 调用LLM生成问题
	log.Printf("正在生成题目: %s, 难度: %s", req.Title, req.Difficulty)
	content, err := g.llmClient.GenerateCompletion(messages)
	if err != nil {
		return nil, fmt.Errorf("生成问题失败: %w", err)
	}

	// 解析生成的内容
	problem, err := ParseGeneratedProblem(content)
	if err != nil {
		return nil, fmt.Errorf("解析生成的问题失败: %w", err)
	}

	// 设置默认值（如果未指定）
	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000 // 1秒
	}
	if problem.MemoryLimit == 0 {
		problem.MemoryLimit = 256 * 1024 // 256MB
	}
	if problem.Difficulty == "" {
		problem.Difficulty = req.Difficulty
	}

	return problem, nil
}

// buildPrompt 构建提示词
func buildPrompt(req ProblemGenerationRequest) string {
	var sb strings.Builder

	sb.WriteString("请生成一个符合以下要求的编程问题:\n\n")

	if req.Title != "" {
		sb.WriteString("标题/主题: " + req.Title + "\n")
	}

	if req.OutlineSection != "" {
		sb.WriteString("大纲章节: " + req.OutlineSection + "\n")
	}

	if len(req.KnowledgePoints) > 0 {
		sb.WriteString("知识点: " + strings.Join(req.KnowledgePoints, ", ") + "\n")
	}

	sb.WriteString("难度级别: " + req.Difficulty + "\n")

	if req.ProblemType != "" {
		sb.WriteString("问题类型: " + req.ProblemType + "\n")
	}

	if req.TimeComplexity != "" {
		sb.WriteString("时间复杂度要求: " + req.TimeComplexity + "\n")
	}

	if req.SpaceComplexity != "" {
		sb.WriteString("空间复杂度要求: " + req.SpaceComplexity + "\n")
	}

	if req.AdditionalReqs != "" {
		sb.WriteString("额外要求: " + req.AdditionalReqs + "\n")
	}

	testCaseCount := req.TestCaseCount
	if testCaseCount <= 0 {
		testCaseCount = 3
	}
	sb.WriteString(fmt.Sprintf("生成测试用例数量: %d\n", testCaseCount))

	if req.IncludeReferenceSolution {
		sb.WriteString("请包含参考解答\n")
	}

	if req.IncludeAnalysis {
		sb.WriteString("请包含思维训练分析\n")
	}

	sb.WriteString("\n请按照以下JSON格式返回结果:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"title\": \"问题标题\",\n")
	sb.WriteString("  \"description\": \"详细的问题描述，包括题目背景、要求、输入输出格式等\",\n")
	sb.WriteString("  \"difficulty\": \"难度级别: Easy, Medium, Hard\",\n")
	sb.WriteString("  \"time_limit\": 1000,\n")
	sb.WriteString("  \"memory_limit\": 262144,\n")
	sb.WriteString("  \"knowledge_tag\": [\"相关知识点\"],\n")
	sb.WriteString("  \"test_cases\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"input\": \"测试输入\",\n")
	sb.WriteString("      \"output\": \"期望输出\",\n")
	sb.WriteString("      \"is_example\": true\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")

	if req.IncludeReferenceSolution {
		sb.WriteString("  \"reference_solution\": \"C++参考代码\",\n")
	}

	if req.IncludeAnalysis {
		sb.WriteString("  \"thinking_analysis\": \"思维训练分析\"\n")
	} else {
		sb.WriteString("  \"thinking_analysis\": \"\"\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("```\n")

	return sb.String()
}

// GetDefaultProblemConfigPath 获取默认的LLM配置文件路径
func GetDefaultProblemConfigPath() string {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return "config/deepseek_config.json"
	}

	// 默认使用普通DeepSeek配置
	return filepath.Join(workDir, "config", "deepseek_config.json")
}

// GetLLMConfigPath 根据模型类型获取配置文件路径
func GetLLMConfigPath(modelType string) string {
	workDir, err := filepath.Abs(".")
	if err != nil {
		workDir = "."
	}

	switch modelType {
	case "deepseek_silicon":
		return filepath.Join(workDir, "config", "deepseek_silicon_config.json")
	case "openai":
		return filepath.Join(workDir, "config", "llm_config.json")
	case "deepseek":
		fallthrough
	default:
		return filepath.Join(workDir, "config", "deepseek_config.json")
	}
}
