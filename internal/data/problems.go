package data

import (
	"errors"
	"log"
	"sync"
	"time"
)

// 错误定义
var ErrNotFound = errors.New("not found")

// 问题难度级别
type Difficulty string

const (
	Easy   Difficulty = "Easy"
	Medium Difficulty = "Medium"
	Hard   Difficulty = "Hard"
)

// 问题结构体定义
type Problem struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Difficulty        Difficulty `json:"difficulty"`
	TimeLimit         int        `json:"time_limit"`         // 以毫秒为单位
	MemoryLimit       int        `json:"memory_limit"`       // 以KB为单位
	KnowledgeTag      []string   `json:"knowledge_tag"`      // 知识点标签
	ReferenceSolution string     `json:"reference_solution"` // 参考解答
	ThinkingAnalysis  string     `json:"thinking_analysis"`  // 思维分析
	CreatedAt         time.Time  `json:"created_at"`
}

// 测试用例结构体定义
type TestCase struct {
	ID        int    `json:"id"`
	ProblemID int    `json:"problem_id"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	IsExample bool   `json:"is_example"`
}

// 问题存储接口
type ProblemStore interface {
	GetProblem(id int) (*Problem, error)
	GetProblems() ([]*Problem, error)
	CreateProblem(problem *Problem) (*Problem, error)
	GetTestCases(problemID int) ([]*TestCase, error)
	GetExamples(problemID int) ([]*TestCase, error)
	AddTestCase(testCase *TestCase) (*TestCase, error)
}

// 内存中的问题存储实现
type InMemoryProblemStore struct {
	problems    map[int]*Problem
	testCases   map[int][]*TestCase
	problemID   int
	testCaseID  int
	mu          sync.RWMutex
	persistence *PersistenceManager
	enableSave  bool
}

func NewInMemoryProblemStore() *InMemoryProblemStore {
	// 创建持久化管理器
	persistenceManager, err := NewPersistenceManager()
	if err != nil {
		log.Printf("警告: 创建持久化管理器失败: %v, 将使用内存存储", err)
		persistenceManager = nil
	}

	store := &InMemoryProblemStore{
		problems:    make(map[int]*Problem),
		testCases:   make(map[int][]*TestCase),
		problemID:   0,
		testCaseID:  0,
		persistence: persistenceManager,
		enableSave:  false, // 初始禁用保存，等加载完成再启用
	}

	// 尝试从文件加载数据
	if persistenceManager != nil {
		store.loadData()
	}

	// 如果没有问题，添加样例
	if len(store.problems) == 0 {
		log.Println("数据库中没有问题，添加示例问题")
		store.addSampleProblems()
	}

	// 启用保存
	store.enableSave = true

	return store
}

// loadData 从文件加载数据
func (s *InMemoryProblemStore) loadData() {
	if s.persistence == nil {
		return
	}

	// 加载问题
	problems, err := s.persistence.LoadProblems()
	if err != nil {
		log.Printf("加载问题数据失败: %v", err)
	} else {
		s.problems = problems
		// 找出最大ID
		for id := range problems {
			if id > s.problemID {
				s.problemID = id
			}
		}
		log.Printf("从存储中加载 %d 个问题", len(problems))
	}

	// 加载测试用例
	testCases, err := s.persistence.LoadTestCases()
	if err != nil {
		log.Printf("加载测试用例数据失败: %v", err)
	} else {
		s.testCases = testCases
		// 找出最大ID
		for _, cases := range testCases {
			for _, tc := range cases {
				if tc.ID > s.testCaseID {
					s.testCaseID = tc.ID
				}
			}
		}
		log.Printf("从存储中加载测试用例数据")
	}
}

// saveData 保存数据到文件
func (s *InMemoryProblemStore) saveData() {
	if s.persistence == nil || !s.enableSave {
		log.Printf("跳过保存: persistence=%v, enableSave=%v", s.persistence != nil, s.enableSave)
		return
	}

	log.Printf("开始保存数据到文件... (问题: %d, 测试用例组: %d)", len(s.problems), len(s.testCases))

	// 保存问题
	if err := s.persistence.SaveProblems(s.problems); err != nil {
		log.Printf("保存问题数据失败: %v", err)
	}

	// 保存测试用例
	if err := s.persistence.SaveTestCases(s.testCases); err != nil {
		log.Printf("保存测试用例数据失败: %v", err)
	}
}

func (s *InMemoryProblemStore) GetProblem(id int) (*Problem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	problem, exists := s.problems[id]
	if !exists {
		return nil, ErrNotFound
	}

	return problem, nil
}

func (s *InMemoryProblemStore) GetProblems() ([]*Problem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	problems := make([]*Problem, 0, len(s.problems))
	for _, problem := range s.problems {
		problems = append(problems, problem)
	}

	return problems, nil
}

func (s *InMemoryProblemStore) CreateProblem(problem *Problem) (*Problem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.problemID++
	problem.ID = s.problemID
	problem.CreatedAt = time.Now()

	s.problems[problem.ID] = problem

	log.Printf("创建新问题: ID=%d, 标题=%s", problem.ID, problem.Title)

	// 保存数据
	s.saveData()

	return problem, nil
}

func (s *InMemoryProblemStore) GetTestCases(problemID int) ([]*TestCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	testCases, exists := s.testCases[problemID]
	if !exists {
		return []*TestCase{}, nil
	}

	return testCases, nil
}

func (s *InMemoryProblemStore) GetExamples(problemID int) ([]*TestCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	testCases, exists := s.testCases[problemID]
	if !exists {
		return []*TestCase{}, nil
	}

	examples := make([]*TestCase, 0)
	for _, tc := range testCases {
		if tc.IsExample {
			examples = append(examples, tc)
		}
	}

	return examples, nil
}

func (s *InMemoryProblemStore) AddTestCase(testCase *TestCase) (*TestCase, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查问题是否存在
	_, exists := s.problems[testCase.ProblemID]
	if !exists {
		return nil, ErrNotFound
	}

	s.testCaseID++
	testCase.ID = s.testCaseID

	if _, exists := s.testCases[testCase.ProblemID]; !exists {
		s.testCases[testCase.ProblemID] = make([]*TestCase, 0)
	}

	s.testCases[testCase.ProblemID] = append(s.testCases[testCase.ProblemID], testCase)

	log.Printf("添加测试用例: ID=%d, 问题ID=%d, 示例=%v", testCase.ID, testCase.ProblemID, testCase.IsExample)

	// 保存数据
	s.saveData()

	return testCase, nil
}

// ProblemImport 表示要导入的问题及其测试用例
type ProblemImport struct {
	Problem   *Problem    `json:"problem"`
	TestCases []*TestCase `json:"test_cases"`
}

// ImportProblems 批量导入多个问题及其测试用例
func (s *InMemoryProblemStore) ImportProblems(imports []*ProblemImport) ([]int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("开始导入 %d 个问题", len(imports))

	// 存储导入的问题ID
	importedIDs := make([]int, 0, len(imports))

	for _, item := range imports {
		// 创建问题
		s.problemID++
		item.Problem.ID = s.problemID
		item.Problem.CreatedAt = time.Now()
		s.problems[item.Problem.ID] = item.Problem
		importedIDs = append(importedIDs, item.Problem.ID)

		log.Printf("导入问题: ID=%d, 标题=%s", item.Problem.ID, item.Problem.Title)

		// 添加问题的测试用例
		for _, tc := range item.TestCases {
			s.testCaseID++
			tc.ID = s.testCaseID
			tc.ProblemID = item.Problem.ID

			if _, exists := s.testCases[item.Problem.ID]; !exists {
				s.testCases[item.Problem.ID] = make([]*TestCase, 0)
			}
			s.testCases[item.Problem.ID] = append(s.testCases[item.Problem.ID], tc)

			log.Printf("导入测试用例: ID=%d, 问题ID=%d, 示例=%v", tc.ID, tc.ProblemID, tc.IsExample)
		}
	}

	log.Printf("完成导入 %d 个问题, 正在保存到文件...", len(imports))

	// 保存数据
	s.saveData()

	return importedIDs, nil
}

// addSampleProblems 添加示例问题和测试用例
func (s *InMemoryProblemStore) addSampleProblems() {
	log.Println("添加示例问题到数据库")

	// 保存当前enableSave状态并临时禁用，避免每个问题都单独保存
	originalEnableSave := s.enableSave
	s.enableSave = false
	defer func() {
		s.enableSave = originalEnableSave
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 题目1：两数之和
	problem1 := &Problem{
		Title:        "两数之和",
		Description:  "给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值的那两个整数，并返回它们的数组下标。\n\n你可以假设每种输入只会对应一个答案。但是，数组中同一个元素不能使用两次。\n\n你可以按任意顺序返回答案。",
		Difficulty:   Easy,
		TimeLimit:    1000,
		MemoryLimit:  65536,
		KnowledgeTag: []string{"数组", "哈希表"},
	}
	s.problemID++
	problem1.ID = s.problemID
	problem1.CreatedAt = time.Now()
	s.problems[problem1.ID] = problem1
	log.Printf("添加示例问题1: ID=%d, 标题=%s", problem1.ID, problem1.Title)

	// 题目1的测试用例
	testCases1 := []*TestCase{
		{
			Input:     "nums = [2,7,11,15], target = 9",
			Output:    "[0,1]",
			IsExample: true,
		},
		{
			Input:     "nums = [3,2,4], target = 6",
			Output:    "[1,2]",
			IsExample: true,
		},
		{
			Input:     "nums = [3,3], target = 6",
			Output:    "[0,1]",
			IsExample: true,
		},
	}

	for _, tc := range testCases1 {
		s.testCaseID++
		tc.ID = s.testCaseID
		tc.ProblemID = problem1.ID
	}
	s.testCases[problem1.ID] = testCases1

	// 题目2：回文数
	problem2 := &Problem{
		Title:        "回文数",
		Description:  "给你一个整数 x ，如果 x 是一个回文整数，返回 true ；否则，返回 false 。\n\n回文数是指正序（从左向右）和倒序（从右向左）读都是一样的整数。\n\n例如，121 是回文，而 123 不是。",
		Difficulty:   Easy,
		TimeLimit:    1000,
		MemoryLimit:  65536,
		KnowledgeTag: []string{"数学", "字符串"},
	}
	s.problemID++
	problem2.ID = s.problemID
	problem2.CreatedAt = time.Now()
	s.problems[problem2.ID] = problem2
	log.Printf("添加示例问题2: ID=%d, 标题=%s", problem2.ID, problem2.Title)

	// 题目2的测试用例
	testCases2 := []*TestCase{
		{
			Input:     "x = 121",
			Output:    "true",
			IsExample: true,
		},
		{
			Input:     "x = -121",
			Output:    "false",
			IsExample: true,
		},
		{
			Input:     "x = 10",
			Output:    "false",
			IsExample: true,
		},
	}

	for _, tc := range testCases2 {
		s.testCaseID++
		tc.ID = s.testCaseID
		tc.ProblemID = problem2.ID
	}
	s.testCases[problem2.ID] = testCases2

	// 题目3：合并两个有序链表
	problem3 := &Problem{
		Title:        "合并两个有序链表",
		Description:  "将两个升序链表合并为一个新的升序链表并返回。新链表是通过拼接给定的两个链表的所有节点组成的。",
		Difficulty:   Easy,
		TimeLimit:    1000,
		MemoryLimit:  65536,
		KnowledgeTag: []string{"链表", "递归"},
	}
	s.problemID++
	problem3.ID = s.problemID
	problem3.CreatedAt = time.Now()
	s.problems[problem3.ID] = problem3
	log.Printf("添加示例问题3: ID=%d, 标题=%s", problem3.ID, problem3.Title)

	// 题目3的测试用例
	testCase3 := &TestCase{
		Input:     "l1 = [1,2,4], l2 = [1,3,4]",
		Output:    "[1,1,2,3,4,4]",
		IsExample: true,
	}
	s.testCaseID++
	testCase3.ID = s.testCaseID
	testCase3.ProblemID = problem3.ID
	s.testCases[problem3.ID] = []*TestCase{testCase3}

	log.Println("示例问题添加完成，准备保存到文件")

	// 启用保存并保存所有数据
	if originalEnableSave {
		s.enableSave = true
		s.saveData()
	}
}
