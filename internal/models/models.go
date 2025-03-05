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
