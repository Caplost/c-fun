package data

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// 默认设置为相对路径，但初始化时会转为绝对路径
var (
	DataDir             = "./data"
	ProblemsFile        = "problems.json"
	TestCasesFile       = "testcases.json"
	UserProblemStatFile = "user_problem_statuses.json"
)

// PersistenceManager 管理数据持久化
type PersistenceManager struct {
	dataDir string
	mu      sync.Mutex
}

// NewPersistenceManager 创建新的持久化管理器
func NewPersistenceManager() (*PersistenceManager, error) {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 使用绝对路径
	absDataDir := filepath.Join(workDir, "data")
	log.Printf("数据存储路径: %s", absDataDir)

	// 确保数据目录存在
	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return &PersistenceManager{
		dataDir: absDataDir,
	}, nil
}

// SaveProblems 保存问题数据到文件
func (pm *PersistenceManager) SaveProblems(problems map[int]*Problem) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := json.MarshalIndent(problems, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化问题数据失败: %w", err)
	}

	filePath := filepath.Join(pm.dataDir, ProblemsFile)
	log.Printf("正在保存问题数据到: %s (%d个问题)", filePath, len(problems))

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存问题数据失败: %w", err)
	}

	log.Printf("成功保存了 %d 个问题到文件", len(problems))
	return nil
}

// LoadProblems 从文件加载问题数据
func (pm *PersistenceManager) LoadProblems() (map[int]*Problem, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	filePath := filepath.Join(pm.dataDir, ProblemsFile)
	log.Printf("尝试从 %s 加载问题数据", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("问题数据文件不存在，返回空map")
		// 文件不存在，返回空map
		return make(map[int]*Problem), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取问题数据失败: %w", err)
	}

	log.Printf("已读取文件数据，大小: %d 字节", len(data))

	var problems map[int]*Problem
	if err := json.Unmarshal(data, &problems); err != nil {
		log.Printf("解析JSON数据失败: %v", err)
		return nil, fmt.Errorf("解析问题数据失败: %w", err)
	}

	log.Printf("成功加载了 %d 个问题", len(problems))
	return problems, nil
}

// SaveTestCases 保存测试用例数据到文件
func (pm *PersistenceManager) SaveTestCases(testCases map[int][]*TestCase) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := json.MarshalIndent(testCases, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试用例数据失败: %w", err)
	}

	filePath := filepath.Join(pm.dataDir, TestCasesFile)
	log.Printf("正在保存测试用例数据到: %s", filePath)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存测试用例数据失败: %w", err)
	}

	// 计算测试用例总数
	total := 0
	for _, cases := range testCases {
		total += len(cases)
	}
	log.Printf("成功保存了 %d 个问题的 %d 个测试用例到文件", len(testCases), total)
	return nil
}

// LoadTestCases 从文件加载测试用例数据
func (pm *PersistenceManager) LoadTestCases() (map[int][]*TestCase, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	filePath := filepath.Join(pm.dataDir, TestCasesFile)
	log.Printf("尝试从 %s 加载测试用例数据", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("测试用例数据文件不存在，返回空map")
		// 文件不存在，返回空map
		return make(map[int][]*TestCase), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取测试用例数据失败: %w", err)
	}

	log.Printf("已读取测试用例文件数据，大小: %d 字节", len(data))

	var testCases map[int][]*TestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		log.Printf("解析测试用例JSON数据失败: %v", err)
		return nil, fmt.Errorf("解析测试用例数据失败: %w", err)
	}

	// 计算测试用例总数
	total := 0
	for _, cases := range testCases {
		total += len(cases)
	}
	log.Printf("成功加载了 %d 个问题的 %d 个测试用例", len(testCases), total)
	return testCases, nil
}

// SaveUserProblemStatuses 保存用户题目状态数据到文件
func (pm *PersistenceManager) SaveUserProblemStatuses(statuses map[string]*UserProblemStatus) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := json.MarshalIndent(statuses, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化用户题目状态数据失败: %w", err)
	}

	filePath := filepath.Join(pm.dataDir, UserProblemStatFile)
	log.Printf("正在保存用户题目状态数据到: %s (%d个状态)", filePath, len(statuses))

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存用户题目状态数据失败: %w", err)
	}

	log.Printf("成功保存了 %d 个用户题目状态到文件", len(statuses))
	return nil
}

// LoadUserProblemStatuses 从文件加载用户题目状态数据
func (pm *PersistenceManager) LoadUserProblemStatuses() (map[string]*UserProblemStatus, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	filePath := filepath.Join(pm.dataDir, UserProblemStatFile)
	log.Printf("尝试从 %s 加载用户题目状态数据", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("用户题目状态数据文件不存在，返回空map")
		// 文件不存在，返回空map
		return make(map[string]*UserProblemStatus), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取用户题目状态数据失败: %w", err)
	}

	log.Printf("已读取用户题目状态文件数据，大小: %d 字节", len(data))

	var statuses map[string]*UserProblemStatus
	if err := json.Unmarshal(data, &statuses); err != nil {
		log.Printf("解析用户题目状态JSON数据失败: %v", err)
		return nil, fmt.Errorf("解析用户题目状态数据失败: %w", err)
	}

	log.Printf("成功加载了 %d 个用户题目状态", len(statuses))
	return statuses, nil
}
