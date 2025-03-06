package db

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/user/cppjudge/internal/data"
	"github.com/user/cppjudge/internal/models"
)

// MemoryStore is a simple in-memory database implementation
type MemoryStore struct {
	mu                       sync.RWMutex
	users                    map[int]models.User
	submissions              map[int]models.Submission
	testResults              map[int]models.TestResult
	problemStore             *data.InMemoryProblemStore           // 使用带持久化功能的问题存储
	userProblemStore         *data.InMemoryUserProblemStatusStore // 用户题目状态存储
	nextUserID               int
	nextSubmitID             int
	nextResultID             int
	problems                 map[int]*models.Problem
	testCases                map[int]*models.TestCase
	userProblemStatus        map[int]*models.UserProblemStatus
	outlineQuestions         map[int]*models.OutlineQuestion
	quizzes                  map[int]*models.Quiz
	quizSubmissions          map[int]*models.QuizSubmission
	quizResults              map[int]*models.QuizResult
	userIDCounter            int
	problemIDCounter         int
	testCaseIDCounter        int
	submissionIDCounter      int
	testResultIDCounter      int
	userProblemStatusCounter int
	outlineQuestionCounter   int
	quizCounter              int
	quizSubmissionCounter    int
	quizResultCounter        int
}

// NewMemoryStore creates a new in-memory database
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:                    make(map[int]models.User),
		submissions:              make(map[int]models.Submission),
		testResults:              make(map[int]models.TestResult),
		problemStore:             data.NewInMemoryProblemStore(),           // 初始化持久化问题存储
		userProblemStore:         data.NewInMemoryUserProblemStatusStore(), // 初始化用户题目状态存储
		nextUserID:               1,
		nextSubmitID:             1,
		nextResultID:             1,
		problems:                 make(map[int]*models.Problem),
		testCases:                make(map[int]*models.TestCase),
		userProblemStatus:        make(map[int]*models.UserProblemStatus),
		outlineQuestions:         make(map[int]*models.OutlineQuestion),
		quizzes:                  make(map[int]*models.Quiz),
		quizSubmissions:          make(map[int]*models.QuizSubmission),
		quizResults:              make(map[int]*models.QuizResult),
		userIDCounter:            1,
		problemIDCounter:         1,
		testCaseIDCounter:        1,
		submissionIDCounter:      1,
		testResultIDCounter:      1,
		userProblemStatusCounter: 1,
		outlineQuestionCounter:   1,
		quizCounter:              1,
		quizSubmissionCounter:    1,
		quizResultCounter:        1,
	}
}

// AddUser adds a new user to the store
func (s *MemoryStore) AddUser(user models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username already exists
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return models.User{}, errors.New("username already exists")
		}
		if existingUser.Email == user.Email {
			return models.User{}, errors.New("email already exists")
		}
	}

	user.ID = s.nextUserID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	s.nextUserID++
	s.users[user.ID] = user

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *MemoryStore) GetUserByID(id int) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *MemoryStore) GetUserByUsername(username string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}

	return models.User{}, errors.New("user not found")
}

// AddProblem adds a new problem
func (s *MemoryStore) AddProblem(problem models.Problem) (models.Problem, error) {
	// 转换为data.Problem
	dataProblem := &data.Problem{
		ID:                problem.ID,
		Title:             problem.Title,
		Description:       problem.Description,
		Difficulty:        data.Difficulty(problem.Difficulty),
		TimeLimit:         problem.TimeLimit,
		MemoryLimit:       problem.MemoryLimit,
		KnowledgeTag:      problem.KnowledgeTag,
		ReferenceSolution: problem.ReferenceSolution,
		ThinkingAnalysis:  problem.ThinkingAnalysis,
		CreatedAt:         problem.CreatedAt,
	}

	// 使用problemStore创建问题
	result, err := s.problemStore.CreateProblem(dataProblem)
	if err != nil {
		return models.Problem{}, err
	}

	// 转换回models.Problem
	return models.Problem{
		ID:                result.ID,
		Title:             result.Title,
		Description:       result.Description,
		Difficulty:        string(result.Difficulty),
		TimeLimit:         result.TimeLimit,
		MemoryLimit:       result.MemoryLimit,
		KnowledgeTag:      result.KnowledgeTag,
		ReferenceSolution: result.ReferenceSolution,
		ThinkingAnalysis:  result.ThinkingAnalysis,
		CreatedAt:         result.CreatedAt,
		UpdatedAt:         time.Now(),
	}, nil
}

// GetProblemByID retrieves a problem by ID
func (s *MemoryStore) GetProblemByID(id int) (models.Problem, error) {
	// 使用problemStore获取问题
	dataProblem, err := s.problemStore.GetProblem(id)
	if err != nil {
		return models.Problem{}, err
	}

	// 转换为models.Problem
	return models.Problem{
		ID:                dataProblem.ID,
		Title:             dataProblem.Title,
		Description:       dataProblem.Description,
		Difficulty:        string(dataProblem.Difficulty),
		TimeLimit:         dataProblem.TimeLimit,
		MemoryLimit:       dataProblem.MemoryLimit,
		KnowledgeTag:      dataProblem.KnowledgeTag,
		ReferenceSolution: dataProblem.ReferenceSolution,
		ThinkingAnalysis:  dataProblem.ThinkingAnalysis,
		CreatedAt:         dataProblem.CreatedAt,
		UpdatedAt:         time.Now(),
	}, nil
}

// ListProblems returns all problems
func (s *MemoryStore) ListProblems() []models.Problem {
	// 使用problemStore获取所有问题
	dataProblems, err := s.problemStore.GetProblems()
	if err != nil {
		return []models.Problem{}
	}

	// 转换为models.Problem切片
	problems := make([]models.Problem, 0, len(dataProblems))
	for _, p := range dataProblems {
		problems = append(problems, models.Problem{
			ID:                p.ID,
			Title:             p.Title,
			Description:       p.Description,
			Difficulty:        string(p.Difficulty),
			TimeLimit:         p.TimeLimit,
			MemoryLimit:       p.MemoryLimit,
			KnowledgeTag:      p.KnowledgeTag,
			ReferenceSolution: p.ReferenceSolution,
			ThinkingAnalysis:  p.ThinkingAnalysis,
			CreatedAt:         p.CreatedAt,
			UpdatedAt:         time.Now(),
		})
	}

	// 按照ID倒序排列
	sort.Slice(problems, func(i, j int) bool {
		return problems[i].ID > problems[j].ID
	})

	return problems
}

// AddTestCase adds a new test case
func (s *MemoryStore) AddTestCase(testCase models.TestCase) (models.TestCase, error) {
	// 转换为data.TestCase
	dataTestCase := &data.TestCase{
		ID:        testCase.ID,
		ProblemID: testCase.ProblemID,
		Input:     testCase.Input,
		Output:    testCase.Output,
		IsExample: testCase.IsExample,
	}

	// 使用problemStore添加测试用例
	result, err := s.problemStore.AddTestCase(dataTestCase)
	if err != nil {
		return models.TestCase{}, err
	}

	// 转换回models.TestCase
	return models.TestCase{
		ID:        result.ID,
		ProblemID: result.ProblemID,
		Input:     result.Input,
		Output:    result.Output,
		IsExample: result.IsExample,
	}, nil
}

// GetTestCasesByProblemID retrieves all test cases for a problem
func (s *MemoryStore) GetTestCasesByProblemID(problemID int) ([]models.TestCase, error) {
	// 使用problemStore获取测试用例
	dataTestCases, err := s.problemStore.GetTestCases(problemID)
	if err != nil {
		return nil, err
	}

	// 转换为models.TestCase切片
	testCases := make([]models.TestCase, 0, len(dataTestCases))
	for _, tc := range dataTestCases {
		testCases = append(testCases, models.TestCase{
			ID:        tc.ID,
			ProblemID: tc.ProblemID,
			Input:     tc.Input,
			Output:    tc.Output,
			IsExample: tc.IsExample,
		})
	}

	return testCases, nil
}

// 批量导入问题和测试用例
func (s *MemoryStore) ImportProblems(problems []models.Problem, testCases map[int][]models.TestCase) ([]int, error) {
	// 转换为data.ProblemImport切片
	imports := make([]*data.ProblemImport, 0, len(problems))

	for i, p := range problems {
		// 创建data.Problem
		dataProblem := &data.Problem{
			Title:        p.Title,
			Description:  p.Description,
			Difficulty:   data.Difficulty(p.Difficulty),
			TimeLimit:    p.TimeLimit,
			MemoryLimit:  p.MemoryLimit,
			KnowledgeTag: p.KnowledgeTag,
		}

		// 获取对应的测试用例
		cases := testCases[i] // 使用索引作为临时ID
		dataTestCases := make([]*data.TestCase, 0, len(cases))

		for _, tc := range cases {
			dataTestCases = append(dataTestCases, &data.TestCase{
				Input:     tc.Input,
				Output:    tc.Output,
				IsExample: tc.IsExample,
			})
		}

		imports = append(imports, &data.ProblemImport{
			Problem:   dataProblem,
			TestCases: dataTestCases,
		})
	}

	// 调用批量导入
	return s.problemStore.ImportProblems(imports)
}

// AddSubmission adds a new submission
func (s *MemoryStore) AddSubmission(submission models.Submission) (models.Submission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user and problem exist
	if _, exists := s.users[submission.UserID]; !exists {
		return models.Submission{}, errors.New("user not found")
	}

	// 检查problem是否存在
	_, err := s.problemStore.GetProblem(submission.ProblemID)
	if err != nil {
		return models.Submission{}, errors.New("problem not found")
	}

	submission.ID = s.nextSubmitID
	submission.CreatedAt = time.Now()
	submission.SubmittedAt = time.Now()
	s.nextSubmitID++
	s.submissions[submission.ID] = submission

	return submission, nil
}

// GetSubmissionByID retrieves a submission by ID
func (s *MemoryStore) GetSubmissionByID(id int) (models.Submission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submission, exists := s.submissions[id]
	if !exists {
		return models.Submission{}, errors.New("submission not found")
	}

	return submission, nil
}

// UpdateSubmission updates a submission
func (s *MemoryStore) UpdateSubmission(submission models.Submission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.submissions[submission.ID]; !exists {
		return errors.New("submission not found")
	}

	s.submissions[submission.ID] = submission
	return nil
}

// AddTestResult adds a new test result
func (s *MemoryStore) AddTestResult(result models.TestResult) (models.TestResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if submission exists
	if _, exists := s.submissions[result.SubmissionID]; !exists {
		return models.TestResult{}, errors.New("submission not found")
	}

	// 获取测试用例以验证其存在
	problemID := s.submissions[result.SubmissionID].ProblemID
	dataTestCases, err := s.problemStore.GetTestCases(problemID)
	if err != nil {
		return models.TestResult{}, errors.New("problem not found")
	}

	// 验证测试用例ID是否有效
	found := false
	for _, tc := range dataTestCases {
		if tc.ID == result.TestCaseID {
			found = true
			break
		}
	}

	if !found {
		return models.TestResult{}, errors.New("test case not found")
	}

	result.ID = s.nextResultID
	s.nextResultID++
	s.testResults[result.ID] = result

	return result, nil
}

// GetTestResultsBySubmissionID retrieves all test results for a submission
func (s *MemoryStore) GetTestResultsBySubmissionID(submissionID int) ([]models.TestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if submission exists
	if _, exists := s.submissions[submissionID]; !exists {
		return nil, errors.New("submission not found")
	}

	results := make([]models.TestResult, 0)
	for _, result := range s.testResults {
		if result.SubmissionID == submissionID {
			results = append(results, result)
		}
	}

	return results, nil
}

// GetUserProblemStatus 获取用户对特定问题的状态
func (s *MemoryStore) GetUserProblemStatus(userID, problemID int) (models.UserProblemStatus, error) {
	// 检查用户是否存在
	if _, exists := s.users[userID]; !exists {
		return models.UserProblemStatus{}, errors.New("user not found")
	}

	// 检查问题是否存在
	_, err := s.problemStore.GetProblem(problemID)
	if err != nil {
		return models.UserProblemStatus{}, errors.New("problem not found")
	}

	// 获取用户题目状态
	dataStatus, err := s.userProblemStore.GetUserProblemStatus(userID, problemID)
	if err != nil {
		if err == data.ErrNotFound {
			// 如果状态不存在，创建一个新的空状态
			return models.UserProblemStatus{
				UserID:         userID,
				ProblemID:      problemID,
				Attempted:      false,
				Solved:         false,
				FailedAttempts: 0,
			}, nil
		}
		return models.UserProblemStatus{}, err
	}

	// 转换为models.UserProblemStatus
	return models.UserProblemStatus{
		ID:             dataStatus.ID,
		UserID:         dataStatus.UserID,
		ProblemID:      dataStatus.ProblemID,
		Attempted:      dataStatus.Attempted,
		Solved:         dataStatus.Solved,
		FailedAttempts: dataStatus.FailedAttempts,
		LastAttemptAt:  dataStatus.LastAttemptAt,
		FirstSolvedAt:  dataStatus.FirstSolvedAt,
		CreatedAt:      dataStatus.CreatedAt,
		UpdatedAt:      dataStatus.UpdatedAt,
	}, nil
}

// GetUserProblemStatuses 获取用户所有题目的状态
func (s *MemoryStore) GetUserProblemStatuses(userID int) ([]models.UserProblemStatus, error) {
	// 检查用户是否存在
	if _, exists := s.users[userID]; !exists {
		return nil, errors.New("user not found")
	}

	// 获取用户所有题目状态
	dataStatuses, err := s.userProblemStore.GetUserProblemStatuses(userID)
	if err != nil {
		return nil, err
	}

	// 转换为models.UserProblemStatus切片
	statuses := make([]models.UserProblemStatus, 0, len(dataStatuses))
	for _, ds := range dataStatuses {
		statuses = append(statuses, models.UserProblemStatus{
			ID:             ds.ID,
			UserID:         ds.UserID,
			ProblemID:      ds.ProblemID,
			Attempted:      ds.Attempted,
			Solved:         ds.Solved,
			FailedAttempts: ds.FailedAttempts,
			LastAttemptAt:  ds.LastAttemptAt,
			FirstSolvedAt:  ds.FirstSolvedAt,
			CreatedAt:      ds.CreatedAt,
			UpdatedAt:      ds.UpdatedAt,
		})
	}

	return statuses, nil
}

// UpdateUserProblemStatus 更新用户题目状态
func (s *MemoryStore) UpdateUserProblemStatus(status models.UserProblemStatus) (models.UserProblemStatus, error) {
	// 检查用户是否存在
	if _, exists := s.users[status.UserID]; !exists {
		return models.UserProblemStatus{}, errors.New("user not found")
	}

	// 检查问题是否存在
	_, err := s.problemStore.GetProblem(status.ProblemID)
	if err != nil {
		return models.UserProblemStatus{}, errors.New("problem not found")
	}

	// 转换为data.UserProblemStatus
	dataStatus := &data.UserProblemStatus{
		ID:             status.ID,
		UserID:         status.UserID,
		ProblemID:      status.ProblemID,
		Attempted:      status.Attempted,
		Solved:         status.Solved,
		FailedAttempts: status.FailedAttempts,
		LastAttemptAt:  status.LastAttemptAt,
		FirstSolvedAt:  status.FirstSolvedAt,
		CreatedAt:      status.CreatedAt,
		UpdatedAt:      status.UpdatedAt,
	}

	// 更新状态
	result, err := s.userProblemStore.UpdateUserProblemStatus(dataStatus)
	if err != nil {
		return models.UserProblemStatus{}, err
	}

	// 转换回models.UserProblemStatus
	return models.UserProblemStatus{
		ID:             result.ID,
		UserID:         result.UserID,
		ProblemID:      result.ProblemID,
		Attempted:      result.Attempted,
		Solved:         result.Solved,
		FailedAttempts: result.FailedAttempts,
		LastAttemptAt:  result.LastAttemptAt,
		FirstSolvedAt:  result.FirstSolvedAt,
		CreatedAt:      result.CreatedAt,
		UpdatedAt:      result.UpdatedAt,
	}, nil
}

// AddOutlineQuestion 添加一个新的大纲题目
func (s *MemoryStore) AddOutlineQuestion(question models.OutlineQuestion) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	question.ID = s.outlineQuestionCounter
	question.CreatedAt = time.Now()
	s.outlineQuestions[question.ID] = &question
	s.outlineQuestionCounter++

	return question.ID, nil
}

// UpdateOutlineQuestion 更新现有的大纲题目
func (s *MemoryStore) UpdateOutlineQuestion(question models.OutlineQuestion) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.outlineQuestions[question.ID]; !exists {
		return fmt.Errorf("题目不存在: ID=%d", question.ID)
	}

	s.outlineQuestions[question.ID] = &question
	return nil
}

// GetOutlineQuestion 获取指定ID的大纲题目
func (s *MemoryStore) GetOutlineQuestion(id int) (*models.OutlineQuestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	question, exists := s.outlineQuestions[id]
	if !exists {
		return nil, fmt.Errorf("题目不存在: ID=%d", id)
	}

	return question, nil
}

// GetAllOutlineQuestions 获取所有大纲题目
func (s *MemoryStore) GetAllOutlineQuestions() ([]*models.OutlineQuestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	questions := make([]*models.OutlineQuestion, 0, len(s.outlineQuestions))
	for _, question := range s.outlineQuestions {
		questions = append(questions, question)
	}

	return questions, nil
}

// GetOutlineQuestionsByTags 根据标签获取大纲题目
func (s *MemoryStore) GetOutlineQuestionsByTags(tags []string) ([]*models.OutlineQuestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(tags) == 0 {
		return s.GetAllOutlineQuestions()
	}

	questions := make([]*models.OutlineQuestion, 0)
	for _, question := range s.outlineQuestions {
		// 检查题目是否包含任何一个标签
		for _, tag := range tags {
			found := false
			for _, qtag := range question.KnowledgeTag {
				if qtag == tag {
					found = true
					break
				}
			}
			if found {
				questions = append(questions, question)
				break
			}
		}
	}

	return questions, nil
}

// GetOutlineQuestionsBySection 根据大纲章节获取题目
func (s *MemoryStore) GetOutlineQuestionsBySection(section string) ([]*models.OutlineQuestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	questions := make([]*models.OutlineQuestion, 0)
	for _, question := range s.outlineQuestions {
		if strings.HasPrefix(question.OutlineRef, section) {
			questions = append(questions, question)
		}
	}

	return questions, nil
}

// DeleteOutlineQuestion 删除指定ID的大纲题目
func (s *MemoryStore) DeleteOutlineQuestion(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.outlineQuestions[id]; !exists {
		return fmt.Errorf("题目不存在: ID=%d", id)
	}

	delete(s.outlineQuestions, id)
	return nil
}

// AddQuiz 添加一个新的测试
func (s *MemoryStore) AddQuiz(quiz models.Quiz) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quiz.ID = s.quizCounter
	quiz.CreatedAt = time.Now()
	s.quizzes[quiz.ID] = &quiz
	s.quizCounter++

	return quiz.ID, nil
}

// UpdateQuiz 更新现有的测试
func (s *MemoryStore) UpdateQuiz(quiz models.Quiz) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.quizzes[quiz.ID]; !exists {
		return fmt.Errorf("测试不存在: ID=%d", quiz.ID)
	}

	s.quizzes[quiz.ID] = &quiz
	return nil
}

// GetQuiz 获取指定ID的测试
func (s *MemoryStore) GetQuiz(id int) (*models.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quiz, exists := s.quizzes[id]
	if !exists {
		return nil, fmt.Errorf("测试不存在: ID=%d", id)
	}

	return quiz, nil
}

// GetAllQuizzes 获取所有测试
func (s *MemoryStore) GetAllQuizzes() ([]*models.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quizzes := make([]*models.Quiz, 0, len(s.quizzes))
	for _, quiz := range s.quizzes {
		quizzes = append(quizzes, quiz)
	}

	return quizzes, nil
}

// GetQuizzesByTags 根据标签获取测试
func (s *MemoryStore) GetQuizzesByTags(tags []string) ([]*models.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(tags) == 0 {
		return s.GetAllQuizzes()
	}

	quizzes := make([]*models.Quiz, 0)
	for _, quiz := range s.quizzes {
		// 检查测试是否包含任何一个标签
		for _, tag := range tags {
			found := false
			for _, qtag := range quiz.KnowledgeTag {
				if qtag == tag {
					found = true
					break
				}
			}
			if found {
				quizzes = append(quizzes, quiz)
				break
			}
		}
	}

	return quizzes, nil
}

// DeleteQuiz 删除指定ID的测试
func (s *MemoryStore) DeleteQuiz(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.quizzes[id]; !exists {
		return fmt.Errorf("测试不存在: ID=%d", id)
	}

	delete(s.quizzes, id)
	return nil
}

// AddQuizSubmission 添加一个测试提交
func (s *MemoryStore) AddQuizSubmission(submission models.QuizSubmission) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission.ID = s.quizSubmissionCounter
	submission.CreatedAt = time.Now()
	s.quizSubmissions[submission.ID] = &submission
	s.quizSubmissionCounter++

	return submission.ID, nil
}

// GetQuizSubmission 获取指定ID的测试提交
func (s *MemoryStore) GetQuizSubmission(id int) (*models.QuizSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submission, exists := s.quizSubmissions[id]
	if !exists {
		return nil, fmt.Errorf("测试提交不存在: ID=%d", id)
	}

	return submission, nil
}

// GetQuizSubmissionsByUser 获取指定用户的所有测试提交
func (s *MemoryStore) GetQuizSubmissionsByUser(userID int) ([]*models.QuizSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submissions := make([]*models.QuizSubmission, 0)
	for _, submission := range s.quizSubmissions {
		if submission.UserID == userID {
			submissions = append(submissions, submission)
		}
	}

	return submissions, nil
}

// GetQuizSubmissionsByQuestion 获取指定题目的所有测试提交
func (s *MemoryStore) GetQuizSubmissionsByQuestion(questionID int) ([]*models.QuizSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submissions := make([]*models.QuizSubmission, 0)
	for _, submission := range s.quizSubmissions {
		if submission.QuestionID == questionID {
			submissions = append(submissions, submission)
		}
	}

	return submissions, nil
}

// AddQuizResult 添加一个测试结果
func (s *MemoryStore) AddQuizResult(result models.QuizResult) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result.ID = s.quizResultCounter
	result.CreatedAt = time.Now()
	s.quizResults[result.ID] = &result
	s.quizResultCounter++

	return result.ID, nil
}

// GetQuizResult 获取指定ID的测试结果
func (s *MemoryStore) GetQuizResult(id int) (*models.QuizResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.quizResults[id]
	if !exists {
		return nil, fmt.Errorf("测试结果不存在: ID=%d", id)
	}

	return result, nil
}

// GetQuizResultsByUser 获取指定用户的所有测试结果
func (s *MemoryStore) GetQuizResultsByUser(userID int) ([]*models.QuizResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*models.QuizResult, 0)
	for _, result := range s.quizResults {
		if result.UserID == userID {
			results = append(results, result)
		}
	}

	return results, nil
}

// GetQuizResultsByQuiz 获取指定测试的所有结果
func (s *MemoryStore) GetQuizResultsByQuiz(quizID int) ([]*models.QuizResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*models.QuizResult, 0)
	for _, result := range s.quizResults {
		if result.QuizID == quizID {
			results = append(results, result)
		}
	}

	return results, nil
}
