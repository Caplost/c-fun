package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// OutlineItem 表示大纲的一个知识点条目
type OutlineItem struct {
	Section    string   // 小节编号，如 "2.1.2"
	Title      string   // 小节标题，如 "C++程序设计"
	Knowledge  string   // 具体知识点，如 "整数型：int、long long"
	Difficulty int      // 难度等级 1-10
	Tags       []string // 相关标签
	Path       string   // 完整路径，如 "2.1.2.基本数据类型.整数型"
}

// OutlineSection 表示大纲的一个章节
type OutlineSection struct {
	ID       string            // 章节ID，例如 "2.1"
	Title    string            // 章节标题
	Children []*OutlineSection // 子章节
	Items    []OutlineItem     // 本章节的知识点条目
	Level    int               // 章节级别
}

// ParseOutlineFile 解析大纲文件并返回知识点条目
func ParseOutlineFile(filePath string) ([]OutlineItem, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("警告: 大纲文件不存在: %s", filePath)
		// 返回空结果和错误
		return nil, fmt.Errorf("大纲文件不存在: %s", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("无法读取大纲文件: %v", err)
		return nil, err
	}

	log.Printf("成功读取大纲文件，大小: %d 字节", len(data))

	// 按行分割内容
	lines := strings.Split(string(data), "\n")
	var items []OutlineItem

	// 正则表达式
	numListRegex := regexp.MustCompile(`^(\d+)\. `)
	difficultyRegex := regexp.MustCompile(`【(\d+)】`)
	bulletListRegex := regexp.MustCompile(`^- `)
	indentedListRegex := regexp.MustCompile(`^   - `)

	// 解析行
	var currentMainSection string // 如 "2.1"
	var currentMainTitle string   // 如 "入门级"
	var currentSubSection string  // 如 "2.1.1"
	var currentSubTitle string    // 如 "基础知识与编程环境"
	var currentParentItem string  // 子列表的父条目，如 "程序基本概念"
	var currentDifficulty int = 1 // 当前知识点难度

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析主章节标题 (## 开头)
		if strings.HasPrefix(line, "## ") {
			section := strings.TrimPrefix(line, "## ")
			parts := strings.SplitN(section, " ", 2)
			if len(parts) == 2 {
				currentMainSection = parts[0]
				currentMainTitle = parts[1]
				log.Printf("解析到主章节: %s %s", currentMainSection, currentMainTitle)
			}
			continue
		}

		// 解析子章节标题 (### 开头)
		if strings.HasPrefix(line, "### ") {
			section := strings.TrimPrefix(line, "### ")
			parts := strings.SplitN(section, " ", 2)
			if len(parts) == 2 {
				currentSubSection = parts[0]
				currentSubTitle = parts[1]
				log.Printf("解析到子章节: %s %s", currentSubSection, currentSubTitle)
			}
			continue
		}

		// 解析编号条目 (数字+点开头)
		if numListRegex.MatchString(line) {
			parts := numListRegex.FindStringSubmatch(line)
			if len(parts) > 1 {
				// 提取编号后的内容
				content := numListRegex.ReplaceAllString(line, "")

				// 提取难度级别 【x】
				var difficultyStr string
				if matches := difficultyRegex.FindStringSubmatch(content); len(matches) > 1 {
					difficultyStr = matches[1]
					currentDifficulty, _ = strconv.Atoi(difficultyStr)
					// 去掉难度标记，只保留内容
					content = difficultyRegex.ReplaceAllString(content, "")
					content = strings.TrimSpace(content)
				}

				// 记录当前父条目
				currentParentItem = content

				// 只有当条目不是子列表的父条目时才添加为知识点
				if !strings.HasSuffix(content, ":") && i+1 < len(lines) && !indentedListRegex.MatchString(strings.TrimSpace(lines[i+1])) {
					item := OutlineItem{
						Section:    currentSubSection,
						Title:      currentSubTitle,
						Knowledge:  content,
						Difficulty: currentDifficulty,
						Tags:       []string{currentMainTitle},
						Path:       fmt.Sprintf("%s.%s.%s", currentMainSection, currentSubTitle, content),
					}
					items = append(items, item)
					log.Printf("解析到编号知识点: %s (难度: %d)", content, currentDifficulty)
				}
			}
			continue
		}

		// 解析一级列表 (- 开头)
		if bulletListRegex.MatchString(line) && !indentedListRegex.MatchString(line) {
			content := bulletListRegex.ReplaceAllString(line, "")

			// 提取难度级别 【x】
			itemDifficulty := currentDifficulty
			if matches := difficultyRegex.FindStringSubmatch(content); len(matches) > 1 {
				difficultyStr := matches[1]
				itemDifficulty, _ = strconv.Atoi(difficultyStr)
				// 去掉难度标记，只保留内容
				content = difficultyRegex.ReplaceAllString(content, "")
				content = strings.TrimSpace(content)
			}

			tags := []string{currentMainTitle}
			if currentParentItem != "" {
				tags = append(tags, currentParentItem)
			}

			item := OutlineItem{
				Section:    currentSubSection,
				Title:      currentSubTitle,
				Knowledge:  content,
				Difficulty: itemDifficulty,
				Tags:       tags,
				Path:       fmt.Sprintf("%s.%s.%s", currentMainSection, currentSubTitle, content),
			}
			items = append(items, item)
			log.Printf("解析到列表知识点: %s (难度: %d, 标签: %v)", content, itemDifficulty, tags)
			continue
		}

		// 解析二级列表 (缩进的 - 开头)
		if indentedListRegex.MatchString(line) {
			content := indentedListRegex.ReplaceAllString(line, "")

			// 提取难度级别 【x】
			itemDifficulty := currentDifficulty
			if matches := difficultyRegex.FindStringSubmatch(content); len(matches) > 1 {
				difficultyStr := matches[1]
				itemDifficulty, _ = strconv.Atoi(difficultyStr)
				// 去掉难度标记，只保留内容
				content = difficultyRegex.ReplaceAllString(content, "")
				content = strings.TrimSpace(content)
			}

			tags := []string{currentMainTitle}
			if currentParentItem != "" {
				tags = append(tags, currentParentItem)
			}

			item := OutlineItem{
				Section:    currentSubSection,
				Title:      currentSubTitle,
				Knowledge:  content,
				Difficulty: itemDifficulty,
				Tags:       tags,
				Path:       fmt.Sprintf("%s.%s.%s.%s", currentMainSection, currentSubTitle, currentParentItem, content),
			}
			items = append(items, item)
			log.Printf("解析到二级列表知识点: %s (难度: %d, 标签: %v)", content, itemDifficulty, tags)
			continue
		}
	}

	log.Printf("成功解析大纲文件，找到 %d 个知识点", len(items))
	return items, nil
}

// 根据章节计算难度
func calculateDifficulty(section string) int {
	// 简单的难度计算逻辑：基于章节编号
	if strings.HasPrefix(section, "2.1") {
		return 3 // 入门级难度中等
	} else if strings.HasPrefix(section, "2.2") {
		return 5 // 提高级难度稍高
	} else if strings.HasPrefix(section, "2.3") {
		return 8 // NOI级难度很高
	}
	return 3 // 默认中等难度
}

// BuildOutlineTree 从大纲条目构建树形结构
func BuildOutlineTree(items []OutlineItem) []*OutlineSection {
	sectionMap := make(map[string]*OutlineSection)
	var rootSections []*OutlineSection

	// 第一遍：创建所有章节
	for _, item := range items {
		sectionID := item.Section

		// 检查章节是否已存在
		if _, ok := sectionMap[sectionID]; !ok {
			level := 1
			title := ""

			// 提取章节标题
			for _, i := range items {
				if i.Section == sectionID {
					title = i.Title
					break
				}
			}

			newSection := &OutlineSection{
				ID:       sectionID,
				Title:    title,
				Children: []*OutlineSection{},
				Items:    []OutlineItem{},
				Level:    level,
			}

			sectionMap[sectionID] = newSection
			rootSections = append(rootSections, newSection)
		}

		// 将条目添加到对应章节
		if section, ok := sectionMap[sectionID]; ok {
			section.Items = append(section.Items, item)
		}
	}

	// 排序根章节
	// 这里可以根据章节ID进行排序，但简单起见暂时不实现

	return rootSections
}

// GetKnowledgePointsBySection 根据章节ID获取知识点
func GetKnowledgePointsBySection(filePath string, sectionID string) ([]OutlineItem, error) {
	// 检查文件路径是否有效
	if filePath == "" {
		filePath = GetDefaultOutlinePath()
		log.Printf("使用默认大纲文件路径: %s", filePath)
	}

	items, err := ParseOutlineFile(filePath)
	if err != nil {
		log.Printf("解析大纲文件失败: %v", err)
		return nil, err
	}

	log.Printf("为章节 %s 检索知识点，共有 %d 个知识点待筛选", sectionID, len(items))

	var result []OutlineItem

	// 根据章节ID筛选
	for _, item := range items {
		// 章节ID完全匹配
		if item.Section == sectionID {
			result = append(result, item)
			continue
		}

		// 章节ID前缀匹配（如2.1匹配2.1.1, 2.1.2等所有子章节）
		if strings.HasPrefix(item.Section, sectionID+".") {
			result = append(result, item)
			continue
		}

		// 路径前缀匹配
		if strings.HasPrefix(item.Path, sectionID+".") {
			result = append(result, item)
			continue
		}

		// 特殊情况：获取所有知识点
		if sectionID == "all" {
			result = append(result, item)
		}
	}

	log.Printf("为章节 %s 找到 %d 个相关知识点", sectionID, len(result))
	return result, nil
}

// GetDefaultOutlinePath 获取默认的大纲文件路径
func GetDefaultOutlinePath() string {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return "data/2024年信息赛大纲.md"
	}
	return filepath.Join(workDir, "data", "2024年信息赛大纲.md")
}
