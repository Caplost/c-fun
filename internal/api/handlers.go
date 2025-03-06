package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/user/cppjudge/internal/db"
	"github.com/user/cppjudge/internal/judge"
	"github.com/user/cppjudge/internal/models"
	"github.com/user/cppjudge/internal/utils"
)

// Handler provides HTTP handlers for the API
type Handler struct {
	store        *db.MemoryStore
	judgeService *judge.Judge
}

// NewHandler creates a new handler with the given store and judge service
func NewHandler(store *db.MemoryStore, judgeService *judge.Judge) *Handler {
	return &Handler{
		store:        store,
		judgeService: judgeService,
	}
}

// respondJSON sends a JSON response with the given status code
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
	}
}

// respondError sends an error response with the given status code
func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}

// parseID extracts and validates an ID from the URL path
func parseID(r *http.Request) (int, error) {
	// Extract the ID from the URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) == 0 {
		return 0, errors.New("missing ID parameter")
	}

	// Get the last part of the path which should be the ID
	idStr := pathParts[len(pathParts)-1]

	// For paths ending with /testcases or /submissions, get the second last part
	if idStr == "testcases" || idStr == "submissions" {
		if len(pathParts) < 2 {
			return 0, errors.New("missing ID parameter")
		}
		idStr = pathParts[len(pathParts)-2]
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}

// GetProblems returns a list of all problems
func (h *Handler) GetProblems(w http.ResponseWriter, r *http.Request) {
	problems := h.store.ListProblems()
	respondJSON(w, http.StatusOK, problems)
}

// GetProblem returns details of a specific problem
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	problem, err := h.store.GetProblemByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Problem not found")
		return
	}

	// Get example test cases for this problem
	testCases, err := h.store.GetTestCasesByProblemID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve test cases")
		return
	}

	// Filter for example test cases only
	exampleTestCases := []models.TestCase{}
	for _, tc := range testCases {
		if tc.IsExample {
			exampleTestCases = append(exampleTestCases, tc)
		}
	}

	// Combine problem and example test cases in response
	response := map[string]interface{}{
		"problem":  problem,
		"examples": exampleTestCases,
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateProblem creates a new problem
func (h *Handler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var problem models.Problem
	if err := json.NewDecoder(r.Body).Decode(&problem); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if problem.Title == "" || problem.Description == "" {
		respondError(w, http.StatusBadRequest, "Title and description are required")
		return
	}

	// Set default values if not provided
	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000 // 1 second
	}
	if problem.MemoryLimit == 0 {
		problem.MemoryLimit = 128000 // 128 MB
	}

	// Create the problem
	newProblem, err := h.store.AddProblem(problem)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create problem")
		return
	}

	respondJSON(w, http.StatusCreated, newProblem)
}

// AddTestCase adds a test case to a problem
func (h *Handler) AddTestCase(w http.ResponseWriter, r *http.Request) {
	problemID, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var testCase models.TestCase
	if err := json.NewDecoder(r.Body).Decode(&testCase); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set the problem ID
	testCase.ProblemID = problemID

	// Validate input and output are provided
	if testCase.Input == "" || testCase.Output == "" {
		respondError(w, http.StatusBadRequest, "Input and output are required")
		return
	}

	// Create the test case
	newTestCase, err := h.store.AddTestCase(testCase)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create test case")
		return
	}

	respondJSON(w, http.StatusCreated, newTestCase)
}

// SubmitSolution handles a code submission
func (h *Handler) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	problemID, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Extract submission data
	var submission struct {
		UserID   int    `json:"user_id"`
		Code     string `json:"code"`
		Language string `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if submission.UserID == 0 {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	if submission.Code == "" {
		respondError(w, http.StatusBadRequest, "Code is required")
		return
	}
	if submission.Language == "" {
		submission.Language = "cpp" // Default to C++
	}

	// Create submission record
	newSubmission := models.Submission{
		UserID:      submission.UserID,
		ProblemID:   problemID,
		Code:        submission.Code,
		Language:    submission.Language,
		Status:      "Pending",
		CreatedAt:   time.Now(),
		SubmittedAt: time.Now(),
	}

	savedSubmission, err := h.store.AddSubmission(newSubmission)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save submission")
		return
	}

	// Process the submission asynchronously
	go func() {
		if err := h.judgeService.EvaluateSubmission(savedSubmission.ID); err != nil {
			fmt.Printf("Error evaluating submission %d: %v\n", savedSubmission.ID, err)
		}
	}()

	// Return the submission ID and initial status
	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"submission_id": savedSubmission.ID,
		"status":        savedSubmission.Status,
	})
}

// GetSubmission returns details of a specific submission
func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	submission, err := h.store.GetSubmissionByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Submission not found")
		return
	}

	// Get test results if available
	testResults, _ := h.store.GetTestResultsBySubmissionID(id)

	// Include test results in response
	response := map[string]interface{}{
		"submission":   submission,
		"test_results": testResults,
	}

	respondJSON(w, http.StatusOK, response)
}

// RegisterUser registers a new user
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" || user.Password == "" {
		respondError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// In a real application, you would hash the password here
	// For simplicity, we'll skip that step

	// Create the user
	newUser, err := h.store.AddUser(user)
	if err != nil {
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Don't return the password in the response
	newUser.Password = ""

	respondJSON(w, http.StatusCreated, newUser)
}

// GetUser returns details of a specific user
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.store.GetUserByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Don't return the password
	user.Password = ""

	respondJSON(w, http.StatusOK, user)
}

// ImportProblems 处理批量导入题目的请求
func (h *Handler) ImportProblems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Problems []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Difficulty  string `json:"difficulty"`
			TimeLimit   int    `json:"time_limit"`
			MemoryLimit int    `json:"memory_limit"`
			TestCases   []struct {
				Input     string `json:"input"`
				Output    string `json:"output"`
				IsExample bool   `json:"is_example"`
			} `json:"test_cases"`
		} `json:"problems"`
	}

	// 解析请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// 验证请求
	if len(req.Problems) == 0 {
		http.Error(w, "No problems provided", http.StatusBadRequest)
		return
	}

	// 转换为模型
	problems := make([]models.Problem, 0, len(req.Problems))
	testCases := make(map[int][]models.TestCase)

	for i, p := range req.Problems {
		// 设置默认值
		timeLimit := p.TimeLimit
		if timeLimit == 0 {
			timeLimit = 1000 // 默认1秒
		}

		memoryLimit := p.MemoryLimit
		if memoryLimit == 0 {
			memoryLimit = 65536 // 默认64MB
		}

		// 创建问题
		problems = append(problems, models.Problem{
			Title:       p.Title,
			Description: p.Description,
			Difficulty:  p.Difficulty,
			TimeLimit:   timeLimit,
			MemoryLimit: memoryLimit,
		})

		// 创建测试用例
		cases := make([]models.TestCase, 0, len(p.TestCases))
		for _, tc := range p.TestCases {
			cases = append(cases, models.TestCase{
				Input:     tc.Input,
				Output:    tc.Output,
				IsExample: tc.IsExample,
			})
		}
		testCases[i] = cases
	}

	// 执行批量导入
	importedIDs, err := h.store.ImportProblems(problems, testCases)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to import problems: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回结果
	resp := struct {
		IDs   []int `json:"ids"`
		Count int   `json:"count"`
	}{
		IDs:   importedIDs,
		Count: len(importedIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetAllProblemsDebug 返回系统中所有题目及其测试用例的信息，用于调试持久化
func (h *Handler) GetAllProblemsDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取所有题目
	problems := h.store.ListProblems()

	// 为每个题目获取测试用例
	type ProblemWithTestCases struct {
		models.Problem
		TestCases []models.TestCase `json:"testCases"`
	}

	result := make([]ProblemWithTestCases, 0, len(problems))
	for _, p := range problems {
		testCases, err := h.store.GetTestCasesByProblemID(p.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get test cases for problem %d: %v", p.ID, err), http.StatusInternalServerError)
			return
		}

		problemWithTC := ProblemWithTestCases{
			Problem:   p,
			TestCases: testCases,
		}
		result = append(result, problemWithTC)
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetAllProblemsWithStatus 返回所有题目并包含用户状态
func (h *Handler) GetAllProblemsWithStatus(w http.ResponseWriter, r *http.Request) {
	// 从查询参数中获取用户ID
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		respondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 获取所有题目
	problems := h.store.ListProblems()

	// 获取用户的所有题目状态
	userStatuses, err := h.store.GetUserProblemStatuses(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取用户题目状态失败")
		return
	}

	// 创建状态映射 problemID -> status
	statusMap := make(map[int]models.UserProblemStatus)
	for _, status := range userStatuses {
		statusMap[status.ProblemID] = status
	}

	// 组合题目和状态
	type ProblemWithStatus struct {
		models.Problem
		Attempted      bool      `json:"attempted"`
		Solved         bool      `json:"solved"`
		FailedAttempts int       `json:"failed_attempts"`
		LastAttemptAt  time.Time `json:"last_attempt_at,omitempty"`
		FirstSolvedAt  time.Time `json:"first_solved_at,omitempty"`
	}

	result := make([]ProblemWithStatus, 0, len(problems))
	for _, problem := range problems {
		status, exists := statusMap[problem.ID]

		problemWithStatus := ProblemWithStatus{
			Problem:        problem,
			Attempted:      false,
			Solved:         false,
			FailedAttempts: 0,
		}

		if exists {
			problemWithStatus.Attempted = status.Attempted
			problemWithStatus.Solved = status.Solved
			problemWithStatus.FailedAttempts = status.FailedAttempts
			problemWithStatus.LastAttemptAt = status.LastAttemptAt
			problemWithStatus.FirstSolvedAt = status.FirstSolvedAt
		}

		result = append(result, problemWithStatus)
	}

	respondJSON(w, http.StatusOK, result)
}

// GetProblemWithStatus 返回特定题目并包含用户状态
func (h *Handler) GetProblemWithStatus(w http.ResponseWriter, r *http.Request) {
	// 解析题目ID
	problemID, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 从查询参数中获取用户ID
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		respondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 获取题目信息
	problem, err := h.store.GetProblemByID(problemID)
	if err != nil {
		respondError(w, http.StatusNotFound, "题目不存在")
		return
	}

	// 获取用户对该题目的状态
	status, err := h.store.GetUserProblemStatus(userID, problemID)
	if err != nil && err.Error() != "user not found" {
		respondError(w, http.StatusInternalServerError, "获取用户题目状态失败")
		return
	}

	// 获取示例测试用例
	testCases, err := h.store.GetTestCasesByProblemID(problemID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取测试用例失败")
		return
	}

	// 过滤出示例测试用例
	exampleTestCases := []models.TestCase{}
	for _, tc := range testCases {
		if tc.IsExample {
			exampleTestCases = append(exampleTestCases, tc)
		}
	}

	// 组合结果
	result := struct {
		Problem        models.Problem    `json:"problem"`
		Examples       []models.TestCase `json:"examples"`
		Attempted      bool              `json:"attempted"`
		Solved         bool              `json:"solved"`
		FailedAttempts int               `json:"failed_attempts"`
		LastAttemptAt  time.Time         `json:"last_attempt_at,omitempty"`
		FirstSolvedAt  time.Time         `json:"first_solved_at,omitempty"`
	}{
		Problem:        problem,
		Examples:       exampleTestCases,
		Attempted:      status.Attempted,
		Solved:         status.Solved,
		FailedAttempts: status.FailedAttempts,
		LastAttemptAt:  status.LastAttemptAt,
		FirstSolvedAt:  status.FirstSolvedAt,
	}

	respondJSON(w, http.StatusOK, result)
}

// UpdateProblem 更新题目信息
func (h *Handler) UpdateProblem(w http.ResponseWriter, r *http.Request) {
	// 解析题目ID
	problemID, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 检查题目是否存在
	_, err = h.store.GetProblemByID(problemID)
	if err != nil {
		respondError(w, http.StatusNotFound, "题目不存在")
		return
	}

	// 解析请求体
	var updatedProblem models.Problem
	if err := json.NewDecoder(r.Body).Decode(&updatedProblem); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 确保ID匹配
	updatedProblem.ID = problemID

	// 更新题目
	problem, err := h.store.AddProblem(updatedProblem)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "更新题目失败")
		return
	}

	respondJSON(w, http.StatusOK, problem)
}

// ParseOutline 解析大纲文件并生成题目
func (h *Handler) ParseOutline(w http.ResponseWriter, r *http.Request) {
	// 解析大纲文件
	outlineFilePath := "./2024年信息赛大纲.md"
	items, err := utils.ParseOutlineFile(outlineFilePath)
	if err != nil {
		http.Error(w, "解析大纲文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回解析结果
	utils.WriteJSON(w, map[string]interface{}{
		"items_count": len(items),
		"items":       items,
	})
}

// GenerateOutlineQuestions 生成指定章节或标签的题目
func (h *Handler) GenerateOutlineQuestions(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	section := r.Form.Get("section")
	tagsStr := r.Form.Get("tags")
	countStr := r.Form.Get("count")

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	count := 10 // 默认生成10道题目
	if countStr != "" {
		parsedCount, err := strconv.Atoi(countStr)
		if err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	// 解析大纲文件
	outlineFilePath := "./2024年信息赛大纲.md"
	items, err := utils.ParseOutlineFile(outlineFilePath)
	if err != nil {
		http.Error(w, "解析大纲文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 根据章节或标签过滤条目
	var filteredItems []utils.OutlineItem
	if section != "" {
		for _, item := range items {
			if strings.HasPrefix(item.Section, section) {
				filteredItems = append(filteredItems, item)
			}
		}
	} else if len(tags) > 0 {
		for _, item := range items {
			for _, tag := range tags {
				for _, itemTag := range item.Tags {
					if itemTag == tag {
						filteredItems = append(filteredItems, item)
						break
					}
				}
			}
		}
	} else {
		filteredItems = items
	}

	// 如果没有找到匹配条目，返回错误
	if len(filteredItems) == 0 {
		http.Error(w, "未找到匹配的大纲条目", http.StatusBadRequest)
		return
	}

	// 生成问题
	questions := utils.GenerateQuestionsFromOutline(filteredItems, count)

	// 保存生成的问题
	var savedQuestions []models.OutlineQuestion
	for _, q := range questions {
		id, err := h.store.AddOutlineQuestion(q)
		if err != nil {
			http.Error(w, "保存题目失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		q.ID = id
		savedQuestions = append(savedQuestions, q)
	}

	// 返回生成的问题
	utils.WriteJSON(w, map[string]interface{}{
		"count":     len(savedQuestions),
		"questions": savedQuestions,
	})
}

// CreateQuiz 创建一个测试
func (h *Handler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	section := r.Form.Get("section")
	tagsStr := r.Form.Get("tags")
	countStr := r.Form.Get("count")
	title := r.Form.Get("title")
	description := r.Form.Get("description")

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	count := 10 // 默认生成10道题目
	if countStr != "" {
		parsedCount, err := strconv.Atoi(countStr)
		if err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	// 解析大纲文件
	outlineFilePath := "./2024年信息赛大纲.md"
	items, err := utils.ParseOutlineFile(outlineFilePath)
	if err != nil {
		http.Error(w, "解析大纲文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建测试
	var quiz models.Quiz
	if section != "" {
		// 基于章节创建测试
		quiz = utils.GenerateQuizFromOutlineSection(items, section, count)
	} else if len(tags) > 0 {
		// 基于标签过滤条目
		var filteredItems []utils.OutlineItem
		for _, item := range items {
			for _, tag := range tags {
				for _, itemTag := range item.Tags {
					if itemTag == tag {
						filteredItems = append(filteredItems, item)
						break
					}
				}
			}
		}

		// 生成问题
		questions := utils.GenerateQuestionsFromOutline(filteredItems, count)

		// 保存生成的问题
		var questionIDs []int
		for _, q := range questions {
			id, err := h.store.AddOutlineQuestion(q)
			if err != nil {
				http.Error(w, "保存题目失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			questionIDs = append(questionIDs, id)
		}

		// 创建测试
		quiz = models.Quiz{
			Title:        title,
			Description:  description,
			QuestionIDs:  questionIDs,
			KnowledgeTag: tags,
			CreatedAt:    time.Now(),
		}
	} else {
		http.Error(w, "必须指定章节或标签", http.StatusBadRequest)
		return
	}

	// 自定义标题和描述
	if title != "" {
		quiz.Title = title
	}
	if description != "" {
		quiz.Description = description
	}

	// 保存测试
	id, err := h.store.AddQuiz(quiz)
	if err != nil {
		http.Error(w, "保存测试失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	quiz.ID = id
	utils.WriteJSON(w, quiz)
}

// GetAllQuizzes 获取所有测试
func (h *Handler) GetAllQuizzes(w http.ResponseWriter, r *http.Request) {
	quizzes, err := h.store.GetAllQuizzes()
	if err != nil {
		http.Error(w, "获取测试失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, quizzes)
}

// GetQuiz 获取指定测试
func (h *Handler) GetQuiz(w http.ResponseWriter, r *http.Request) {
	// 从URL中提取ID
	path := strings.TrimPrefix(r.URL.Path, "/api/quizzes/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "无效的ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取测试
	quiz, err := h.store.GetQuiz(id)
	if err != nil {
		http.Error(w, "获取测试失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取题目详情
	var questions []models.OutlineQuestion
	for _, qid := range quiz.QuestionIDs {
		q, err := h.store.GetOutlineQuestion(qid)
		if err != nil {
			// 如果有问题获取失败，继续获取其他题目
			log.Printf("获取题目失败 ID=%d: %v", qid, err)
			continue
		}
		questions = append(questions, *q)
	}

	response := struct {
		*models.Quiz
		Questions []models.OutlineQuestion `json:"questions"`
	}{
		Quiz:      quiz,
		Questions: questions,
	}

	utils.WriteJSON(w, response)
}

// SubmitQuizAnswer 提交测试答案
func (h *Handler) SubmitQuizAnswer(w http.ResponseWriter, r *http.Request) {
	// 解析URL中的测试ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "无效的URL", http.StatusBadRequest)
		return
	}

	quizIDStr := pathParts[3]
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		http.Error(w, "无效的测试ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 解析请求体
	var request struct {
		UserID  int               `json:"user_id"`
		Answers map[string]string `json:"answers"` // 题目ID -> 答案
	}

	err = utils.ParseJSONBody(r, &request)
	if err != nil {
		http.Error(w, "解析请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 验证测试存在
	_, err = h.store.GetQuiz(quizID)
	if err != nil {
		http.Error(w, "获取测试失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查并评分答案
	var submissionIDs []int
	correctCount := 0
	totalCount := 0

	for qIDStr, answer := range request.Answers {
		qID, err := strconv.Atoi(qIDStr)
		if err != nil {
			continue
		}

		// 获取题目
		question, err := h.store.GetOutlineQuestion(qID)
		if err != nil {
			continue
		}

		// 创建提交
		isCorrect := false
		// 选择题直接比较答案，填空题和简答题可能需要更复杂的评分逻辑
		if question.Type == models.QuestionTypeMultipleChoice || question.Type == models.QuestionTypeFillBlank {
			isCorrect = answer == question.Answer
		} else if question.Type == models.QuestionTypeShortAnswer {
			// 简答题评分逻辑可以后续完善
			// 这里简单实现一个包含关键词的检查
			keywords := utils.ExtractKeywords(question.Answer)
			keywordCount := 0
			for _, keyword := range keywords {
				if strings.Contains(answer, keyword) {
					keywordCount++
				}
			}
			// 如果包含了一半以上的关键词，认为答案正确
			isCorrect = float64(keywordCount) >= float64(len(keywords))*0.5
		}

		// 保存提交
		submission := models.QuizSubmission{
			UserID:     request.UserID,
			QuestionID: qID,
			Answer:     answer,
			IsCorrect:  isCorrect,
			CreatedAt:  time.Now(),
		}

		id, err := h.store.AddQuizSubmission(submission)
		if err != nil {
			http.Error(w, "保存提交失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		submissionIDs = append(submissionIDs, id)
		if isCorrect {
			correctCount++
		}
		totalCount++
	}

	// 计算得分
	var score float64 = 0
	if totalCount > 0 {
		score = float64(correctCount) / float64(totalCount) * 100
	}

	// 保存测试结果
	result := models.QuizResult{
		UserID:        request.UserID,
		QuizID:        quizID,
		SubmissionIDs: submissionIDs,
		Score:         score,
		CompletedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	resultID, err := h.store.AddQuizResult(result)
	if err != nil {
		http.Error(w, "保存测试结果失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result.ID = resultID
	utils.WriteJSON(w, result)
}

// GetQuizResults 获取测试结果
func (h *Handler) GetQuizResults(w http.ResponseWriter, r *http.Request) {
	// 解析URL中的测试ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "无效的URL", http.StatusBadRequest)
		return
	}

	quizIDStr := pathParts[3]
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		http.Error(w, "无效的测试ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取测试结果
	results, err := h.store.GetQuizResultsByQuiz(quizID)
	if err != nil {
		http.Error(w, "获取测试结果失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, results)
}

// GetUserQuizResults 获取用户的测试结果
func (h *Handler) GetUserQuizResults(w http.ResponseWriter, r *http.Request) {
	// 解析URL中的用户ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "无效的URL", http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[3]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "无效的用户ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取测试结果
	results, err := h.store.GetQuizResultsByUser(userID)
	if err != nil {
		http.Error(w, "获取测试结果失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取测试详情和成绩
	var responseResults []struct {
		Result *models.QuizResult `json:"result"`
		Quiz   *models.Quiz       `json:"quiz"`
	}

	for _, result := range results {
		quiz, err := h.store.GetQuiz(result.QuizID)
		if err != nil {
			continue
		}

		responseResults = append(responseResults, struct {
			Result *models.QuizResult `json:"result"`
			Quiz   *models.Quiz       `json:"quiz"`
		}{
			Result: result,
			Quiz:   quiz,
		})
	}

	utils.WriteJSON(w, responseResults)
}
