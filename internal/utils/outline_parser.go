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
	Section    string   `json:"section"`    // 小节编号，如 "2.1.2"
	Title      string   `json:"title"`      // 小节标题，如 "C++程序设计"
	Knowledge  string   `json:"knowledge"`  // 具体知识点，如 "整数型：int、long long"
	Difficulty int      `json:"difficulty"` // 难度等级 1-10
	Tags       []string `json:"tags"`       // 相关标签
	Path       string   `json:"path"`       // 完整路径，如 "2.1.2.基本数据类型.整数型"
}

// OutlineSection 表示大纲的一个章节
type OutlineSection struct {
	ID       string            `json:"id"`       // 章节ID，例如 "2.1"
	Title    string            `json:"title"`    // 章节标题
	Children []*OutlineSection `json:"children"` // 子章节
	Items    []OutlineItem     `json:"items"`    // 本章节的知识点条目
	Level    int               `json:"level"`    // 章节级别
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
	var currentGroupTitle string  // 当前组标题，例如 "程序基本概念"

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
				// 新章节开始，重置组相关变量
				currentParentItem = ""
				currentGroupTitle = ""
			}
			continue
		}

		// 解析编号条目 (数字+点开头)
		if numListRegex.MatchString(line) {
			match := numListRegex.FindStringSubmatch(line)
			if len(match) > 1 {
				// 提取难度
				diffMatch := difficultyRegex.FindStringSubmatch(line)
				if len(diffMatch) > 1 {
					currentDifficulty, _ = strconv.Atoi(diffMatch[1])
				}

				// 去除难度标记，获取纯文本内容
				content := difficultyRegex.ReplaceAllString(line, "")
				content = numListRegex.ReplaceAllString(content, "")
				content = strings.TrimSpace(content)

				// 检查下一行是否是缩进的列表项
				var nextLine string
				if i+1 < len(lines) {
					nextLine = strings.TrimSpace(lines[i+1])
				}

				if nextLine != "" && indentedListRegex.MatchString(nextLine) {
					// 这是一个组标题
					currentParentItem = content
					currentGroupTitle = content
					log.Printf("解析到组标题: %s (在%s下)", currentParentItem, currentSubSection)
				} else {
					// 这是一个独立的知识点项目
					// 如果在2.1.2章节下，添加更多调试信息
					if currentSubSection == "2.1.2" {
						log.Printf("在2.1.2章节下找到独立知识点: %s, 难度: %d", content, currentDifficulty)
					}

					// 构建完整路径
					path := currentSubSection
					if currentGroupTitle != "" {
						path += "." + currentGroupTitle
					}

					// 创建并添加知识点
					item := OutlineItem{
						Section:    currentSubSection,
						Title:      currentSubTitle,
						Knowledge:  content,
						Difficulty: currentDifficulty,
						Tags:       []string{currentSubTitle, currentGroupTitle},
						Path:       path,
					}
					items = append(items, item)
				}
			}
			continue
		}

		// 处理缩进列表项 (3-4个空格后跟 - )
		if indentedListRegex.MatchString(line) {
			if currentParentItem != "" {
				// 提取难度
				diffMatch := difficultyRegex.FindStringSubmatch(line)
				if len(diffMatch) > 1 {
					currentDifficulty, _ = strconv.Atoi(diffMatch[1])
				}

				// 去除难度标记，获取纯文本内容
				content := difficultyRegex.ReplaceAllString(line, "")
				content = indentedListRegex.ReplaceAllString(content, "")
				content = strings.TrimSpace(content)

				// 如果在2.1.2章节下，添加更多调试信息
				if currentSubSection == "2.1.2" {
					log.Printf("在2.1.2章节的 %s 分组下找到知识点: %s, 难度: %d",
						currentParentItem, content, currentDifficulty)
				}

				// 构建完整路径
				path := currentSubSection + "." + currentParentItem

				// 创建并添加知识点
				item := OutlineItem{
					Section:    currentSubSection,
					Title:      currentSubTitle,
					Knowledge:  content,
					Difficulty: currentDifficulty,
					Tags:       []string{currentSubTitle, currentParentItem},
					Path:       path,
				}
				items = append(items, item)
			}
			continue
		}

		// 处理非缩进的列表项 (- 开头)
		if bulletListRegex.MatchString(line) {
			// 提取难度
			diffMatch := difficultyRegex.FindStringSubmatch(line)
			if len(diffMatch) > 1 {
				currentDifficulty, _ = strconv.Atoi(diffMatch[1])
			}

			// 去除难度标记，获取纯文本内容
			content := difficultyRegex.ReplaceAllString(line, "")
			content = bulletListRegex.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)

			// 如果在2.1.2章节下，添加更多调试信息
			if currentSubSection == "2.1.2" {
				log.Printf("在2.1.2章节下找到非缩进知识点: %s, 难度: %d", content, currentDifficulty)
			}

			// 构建完整路径
			path := currentSubSection

			// 创建并添加知识点
			item := OutlineItem{
				Section:    currentSubSection,
				Title:      currentSubTitle,
				Knowledge:  content,
				Difficulty: currentDifficulty,
				Tags:       []string{currentSubTitle},
				Path:       path,
			}
			items = append(items, item)
		}
	}

	log.Printf("成功解析大纲文件，共找到 %d 个知识点", len(items))

	// 特殊验证2.1.2章节的知识点是否存在
	section212Count := 0
	for _, item := range items {
		if item.Section == "2.1.2" {
			section212Count++
		}
	}
	log.Printf("2.1.2 章节知识点数量: %d", section212Count)

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

	// 添加特别的调试信息
	log.Printf("请求获取章节 %s 的知识点", sectionID)

	items, err := ParseOutlineFile(filePath)
	if err != nil {
		log.Printf("解析大纲文件失败: %v", err)
		return nil, err
	}

	log.Printf("为章节 %s 检索知识点，共有 %d 个知识点待筛选", sectionID, len(items))

	// 特殊处理章节2.1.2
	if sectionID == "2.1.2" {
		log.Printf("特殊处理章节 2.1.2 C++程序设计")
		var section212Items []OutlineItem
		for _, item := range items {
			if item.Section == "2.1.2" {
				section212Items = append(section212Items, item)
			}
		}

		log.Printf("直接筛选出 %d 个属于章节 2.1.2 的知识点", len(section212Items))
		if len(section212Items) > 0 {
			// 返回特别筛选的结果
			return section212Items, nil
		}
	}

	var result []OutlineItem

	// 判断章节是否相关的辅助函数
	isRelatedSection := func(itemSection, targetSection string) bool {
		// 完全匹配
		if itemSection == targetSection {
			return true
		}

		// 子章节匹配（如 targetSection=2.1 匹配 itemSection=2.1.1）
		if strings.HasPrefix(itemSection, targetSection+".") {
			return true
		}

		// 父章节匹配（如 targetSection=2.1.1 匹配 itemSection=2.1）
		if strings.HasPrefix(targetSection, itemSection+".") {
			return true
		}

		// 同级章节匹配（例如 targetSection=2.1 应该同时匹配 2.2, 2.3 等）
		targetParts := strings.Split(targetSection, ".")
		itemParts := strings.Split(itemSection, ".")

		// 特殊情况：获取所有知识点
		if targetSection == "all" {
			return true
		}

		// 对于二级以上章节（如 2.1），检查是否同属一个大章节（如同属于 2）
		if len(targetParts) >= 2 && len(itemParts) >= 2 {
			// 检查第一部分是否相同（例如都属于 "2"）
			if targetParts[0] == itemParts[0] {
				return true
			}
		}

		return false
	}

	// 根据章节ID筛选
	for _, item := range items {
		// 使用辅助函数检查章节是否相关
		if isRelatedSection(item.Section, sectionID) {
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

	// 记录章节统计信息
	sectionCounts := make(map[string]int)
	for _, item := range result {
		sectionCounts[item.Section]++
	}

	for section, count := range sectionCounts {
		log.Printf("章节 %s: %d 个知识点", section, count)
	}

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
