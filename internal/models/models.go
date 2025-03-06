package models

import (
	"time"
)

// User represents a registered user
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Hashed password, not returned in JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Problem represents a programming problem
type Problem struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Difficulty   string    `json:"difficulty"`    // Easy, Medium, Hard
	TimeLimit    int       `json:"time_limit"`    // In milliseconds
	MemoryLimit  int       `json:"memory_limit"`  // In kilobytes
	KnowledgeTag []string  `json:"knowledge_tag"` // 知识点标签，例如：["数组", "二分搜索", "动态规划"]
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TestCase represents input/output test data for a problem
type TestCase struct {
	ID        int    `json:"id"`
	ProblemID int    `json:"problem_id"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	IsExample bool   `json:"is_example"` // Whether this test case is shown to users
}

// Submission represents a user's code submission
type Submission struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ProblemID   int       `json:"problem_id"`
	Language    string    `json:"language"` // Currently only C++
	Code        string    `json:"code"`
	Status      string    `json:"status"`   // Pending, Accepted, Wrong Answer, Time Limit Exceeded, etc.
	RunTime     int       `json:"run_time"` // In milliseconds
	Memory      int       `json:"memory"`   // In kilobytes
	CreatedAt   time.Time `json:"created_at"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// TestResult represents the result of a submission on a specific test case
type TestResult struct {
	ID           int    `json:"id"`
	SubmissionID int    `json:"submission_id"`
	TestCaseID   int    `json:"test_case_id"`
	Status       string `json:"status"`   // Accepted, Wrong Answer, Time Limit Exceeded, etc.
	Output       string `json:"output"`   // The actual output produced by the submission
	RunTime      int    `json:"run_time"` // In milliseconds
	Memory       int    `json:"memory"`   // In kilobytes
}

// UserProblemStatus 表示用户对特定问题的解题状态
type UserProblemStatus struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	ProblemID      int       `json:"problem_id"`
	Attempted      bool      `json:"attempted"`       // 用户是否尝试过该题目
	Solved         bool      `json:"solved"`          // 用户是否解决了该题目
	FailedAttempts int       `json:"failed_attempts"` // 失败尝试次数
	LastAttemptAt  time.Time `json:"last_attempt_at"` // 最后一次尝试的时间
	FirstSolvedAt  time.Time `json:"first_solved_at"` // 首次解决的时间
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// 题目类型常量
const (
	QuestionTypeMultipleChoice = "multiple_choice" // 选择题
	QuestionTypeFillBlank      = "fill_blank"      // 填空题
	QuestionTypeShortAnswer    = "short_answer"    // 简答题
)

// OutlineQuestion 表示根据大纲生成的题目
type OutlineQuestion struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`          // 题目类型: multiple_choice, fill_blank, short_answer
	Difficulty   int       `json:"difficulty"`    // 难度等级 1-10
	Content      string    `json:"content"`       // 题目内容
	Options      []string  `json:"options"`       // 选择题选项
	Answer       string    `json:"answer"`        // 正确答案
	Explanation  string    `json:"explanation"`   // 答案解释
	KnowledgeTag []string  `json:"knowledge_tag"` // 关联的知识点
	OutlineRef   string    `json:"outline_ref"`   // 大纲参考，例如："2.1.2"
	CreatedAt    time.Time `json:"created_at"`
}

// QuizSubmission 表示用户提交的答题结果
type QuizSubmission struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	QuestionID int       `json:"question_id"`
	Answer     string    `json:"answer"`     // 用户的答案
	IsCorrect  bool      `json:"is_correct"` // 是否正确
	CreatedAt  time.Time `json:"created_at"`
}

// Quiz 表示一次测试，包含多个题目
type Quiz struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`         // 测试标题
	Description  string    `json:"description"`   // 测试描述
	QuestionIDs  []int     `json:"question_ids"`  // 包含的题目ID
	KnowledgeTag []string  `json:"knowledge_tag"` // 关联的知识点
	CreatedAt    time.Time `json:"created_at"`
}

// QuizResult 表示用户完成的测试结果
type QuizResult struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	QuizID        int       `json:"quiz_id"`
	SubmissionIDs []int     `json:"submission_ids"` // 相关的提交ID
	Score         float64   `json:"score"`          // 得分
	CompletedAt   time.Time `json:"completed_at"`
	CreatedAt     time.Time `json:"created_at"`
}
