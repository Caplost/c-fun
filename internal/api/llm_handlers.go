package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/user/cppjudge/internal/llm"
	"github.com/user/cppjudge/internal/models"
	"github.com/user/cppjudge/internal/utils"
)

// GenerateAIProblem 处理AI生成题目请求
func (h *Handler) GenerateAIProblem(w http.ResponseWriter, r *http.Request) {
	// 为请求添加超时上下文，设置为3分钟
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()

	// 使用上下文更新请求
	r = r.WithContext(ctx)

	// 设置长连接和保持活动
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Keep-Alive", "timeout=180")

	// 解析请求
	var req struct {
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
		ModelType                string   `json:"model_type"`                 // 模型类型
		AutoGenerate             bool     `json:"auto_generate"`              // 是否自动生成标题和内容
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求格式: "+err.Error())
		return
	}

	// 验证必要字段
	if req.OutlineSection == "" {
		respondError(w, http.StatusBadRequest, "大纲章节不能为空")
		return
	}

	if len(req.KnowledgePoints) == 0 {
		respondError(w, http.StatusBadRequest, "至少需要选择一个知识点")
		return
	}

	// 如果是自动生成模式，为标题生成一个占位符
	if req.AutoGenerate || req.Title == "" {
		log.Printf("使用自动生成模式，将根据知识点自动生成题目")
		// 使用知识点作为标题的一部分
		if len(req.KnowledgePoints) > 0 {
			req.Title = fmt.Sprintf("基于 %s 的编程题目", strings.Join(req.KnowledgePoints, ", "))
		} else {
			req.Title = "自动生成的编程题目"
		}
	}

	if req.Difficulty == "" {
		req.Difficulty = "Medium" // 默认中等难度
	}

	if req.TestCaseCount <= 0 {
		req.TestCaseCount = 3 // 默认3个测试用例
	}

	// 获取配置文件路径
	var configPath string
	if req.ModelType == "" {
		// 使用默认配置
		configPath = llm.GetDefaultProblemConfigPath()
	} else {
		// 根据指定模型类型获取配置
		configPath = llm.GetLLMConfigPath(req.ModelType)
	}
	log.Printf("使用模型配置: %s (%s)", req.ModelType, configPath)

	// 构建生成请求
	generationReq := llm.ProblemGenerationRequest{
		Title:                    req.Title,
		OutlineSection:           req.OutlineSection,
		KnowledgePoints:          req.KnowledgePoints,
		Difficulty:               req.Difficulty,
		ProblemType:              req.ProblemType,
		TimeComplexity:           req.TimeComplexity,
		SpaceComplexity:          req.SpaceComplexity,
		AdditionalReqs:           req.AdditionalReqs,
		TestCaseCount:            req.TestCaseCount,
		IncludeReferenceSolution: req.IncludeReferenceSolution,
		IncludeAnalysis:          req.IncludeAnalysis,
	}

	// 创建问题生成器
	generator, err := llm.NewProblemGenerator(configPath)
	if err != nil {
		log.Printf("创建问题生成器失败: %v", err)
		respondError(w, http.StatusInternalServerError, "创建问题生成器失败: "+err.Error())
		return
	}

	// 使用上下文生成问题
	log.Printf("开始生成问题: %s", req.Title)

	// 添加错误恢复
	defer func() {
		if r := recover(); r != nil {
			log.Printf("生成问题时发生崩溃: %v", r)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("生成问题时发生内部错误: %v", r))
		}
	}()

	problem, err := generator.GenerateProblemWithContext(ctx, generationReq)
	if err != nil {
		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("生成问题超时: %v", err)
			respondError(w, http.StatusGatewayTimeout, "生成问题超时，请稍后再试")
			return
		}

		// 检查是否是网络连接错误
		if strings.Contains(err.Error(), "connection") ||
			strings.Contains(err.Error(), "network") ||
			strings.Contains(err.Error(), "reset by peer") {
			log.Printf("生成问题时发生网络错误: %v", err)
			respondError(w, http.StatusBadGateway, "无法连接到AI服务，请稍后再试: "+err.Error())
			return
		}

		log.Printf("生成问题失败: %v", err)
		respondError(w, http.StatusInternalServerError, "生成问题失败: "+err.Error())
		return
	}

	// 返回生成的问题
	log.Printf("成功生成问题: %s", problem.Title)
	respondJSON(w, http.StatusOK, problem)
}

// GetOutlineKnowledgePoints 获取大纲知识点
func (h *Handler) GetOutlineKnowledgePoints(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	sectionID := r.URL.Query().Get("section")
	if sectionID == "" {
		respondError(w, http.StatusBadRequest, "缺少必要的查询参数: section")
		return
	}

	// 解析大纲文件
	items, err := utils.GetKnowledgePointsBySection(utils.GetDefaultOutlinePath(), sectionID)
	if err != nil {
		log.Printf("解析大纲文件失败: %v", err)
		respondError(w, http.StatusInternalServerError, "解析大纲文件失败: "+err.Error())
		return
	}

	// 返回结果
	respondJSON(w, http.StatusOK, items)
}

// SaveGeneratedProblem 保存生成的题目
func (h *Handler) SaveGeneratedProblem(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var genProblem llm.GeneratedProblem
	if err := json.NewDecoder(r.Body).Decode(&genProblem); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求格式: "+err.Error())
		return
	}

	// 创建新问题
	problem := models.Problem{
		Title:             genProblem.Title,
		Description:       genProblem.Description,
		Difficulty:        genProblem.Difficulty,
		TimeLimit:         genProblem.TimeLimit,
		MemoryLimit:       genProblem.MemoryLimit,
		KnowledgeTag:      genProblem.KnowledgeTag,
		ReferenceSolution: genProblem.ReferenceSolution, // 保存参考解答
		ThinkingAnalysis:  genProblem.ThinkingAnalysis,  // 保存思维分析
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 添加到数据存储
	savedProblem, err := h.store.AddProblem(problem)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "保存问题失败: "+err.Error())
		return
	}

	// 添加测试用例
	for i, tc := range genProblem.TestCases {
		testCase := models.TestCase{
			ProblemID: savedProblem.ID,
			Input:     tc.Input,
			Output:    tc.Output,
			IsExample: true, // 默认为示例测试用例
		}

		// 只将前两个测试用例设为示例
		if i >= 2 {
			testCase.IsExample = false
		}

		// 添加到数据存储
		_, err := h.store.AddTestCase(testCase)
		if err != nil {
			log.Printf("添加测试用例失败: %v", err)
			// 继续添加其他测试用例，不中断流程
		}
	}

	// 返回保存的问题
	response := map[string]interface{}{
		"problem": savedProblem,
		"message": "问题保存成功",
	}
	respondJSON(w, http.StatusCreated, response)
}
