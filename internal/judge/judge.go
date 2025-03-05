package judge

import (
	"fmt"
	"log"
	"time"

	"github.com/user/cppjudge/internal/db"
	"github.com/user/cppjudge/internal/models"
	"github.com/user/cppjudge/internal/sandbox"
)

// Status constants
const (
	StatusPending           = "Pending"
	StatusAccepted          = "Accepted"
	StatusWrongAnswer       = "Wrong Answer"
	StatusCompileError      = "Compilation Error"
	StatusRuntimeError      = "Runtime Error"
	StatusTimeLimitExceeded = "Time Limit Exceeded"
	StatusInternalError     = "Internal Error"
)

// Judge handles evaluating code submissions
type Judge struct {
	store   *db.MemoryStore
	sandbox *sandbox.CppSandbox
}

// NewJudge creates a new judge
func NewJudge(store *db.MemoryStore, sandbox *sandbox.CppSandbox) *Judge {
	return &Judge{
		store:   store,
		sandbox: sandbox,
	}
}

// EvaluateSubmission 评估一个提交
func (j *Judge) EvaluateSubmission(submissionID int) error {
	// 获取提交信息
	submission, err := j.store.GetSubmissionByID(submissionID)
	if err != nil {
		return fmt.Errorf("获取提交信息失败: %w", err)
	}

	// 获取问题信息
	problem, err := j.store.GetProblemByID(submission.ProblemID)
	if err != nil {
		return fmt.Errorf("获取问题信息失败: %w", err)
	}

	// 获取测试用例
	testCases, err := j.store.GetTestCasesByProblemID(problem.ID)
	if err != nil {
		return fmt.Errorf("获取测试用例失败: %w", err)
	}

	if len(testCases) == 0 {
		return fmt.Errorf("问题没有测试用例")
	}

	// 将状态更新为"测试中"
	submission.Status = "Testing"
	if err := j.store.UpdateSubmission(submission); err != nil {
		return fmt.Errorf("更新提交状态失败: %w", err)
	}

	// 获取当前用户的解题状态，初始化用户提交计数
	userStatus, err := j.store.GetUserProblemStatus(submission.UserID, submission.ProblemID)
	if err != nil && err.Error() != "user not found" && err.Error() != "problem not found" {
		return fmt.Errorf("获取用户解题状态失败: %w", err)
	}

	// 标记为用户已尝试
	userStatus.UserID = submission.UserID
	userStatus.ProblemID = submission.ProblemID
	userStatus.Attempted = true
	userStatus.LastAttemptAt = time.Now()

	// 执行测试
	allPassed := true
	maxTime := 0
	maxMemory := 0

	for _, tc := range testCases {
		// 运行测试用例
		result := j.evaluateTestCase(submission, tc, problem.TimeLimit)

		// 保存测试结果
		savedResult, err := j.store.AddTestResult(result)
		if err != nil {
			log.Printf("保存测试结果失败: %v", err)
			allPassed = false
			continue
		}

		// 更新最大运行时间和内存使用
		if savedResult.RunTime > maxTime {
			maxTime = savedResult.RunTime
		}
		if savedResult.Memory > maxMemory {
			maxMemory = savedResult.Memory
		}

		// 如果有测试用例失败，整体结果就是失败
		if savedResult.Status != "Accepted" {
			allPassed = false
		}
	}

	// 更新提交状态
	submission.RunTime = maxTime
	submission.Memory = maxMemory
	if allPassed {
		submission.Status = "Accepted"

		// 更新用户解题状态为已解决
		userStatus.Solved = true
		if userStatus.FirstSolvedAt.IsZero() {
			userStatus.FirstSolvedAt = time.Now()
		}
	} else {
		submission.Status = "Failed"

		// 增加失败次数
		userStatus.FailedAttempts++
	}

	// 保存用户解题状态
	_, err = j.store.UpdateUserProblemStatus(userStatus)
	if err != nil {
		log.Printf("更新用户解题状态失败: %v", err)
	}

	// 保存最终提交结果
	if err := j.store.UpdateSubmission(submission); err != nil {
		return fmt.Errorf("更新提交状态失败: %w", err)
	}

	return nil
}

// evaluateTestCase evaluates a submission against a single test case
func (j *Judge) evaluateTestCase(submission models.Submission, testCase models.TestCase, timeLimit int) models.TestResult {
	result := models.TestResult{
		SubmissionID: submission.ID,
		TestCaseID:   testCase.ID,
	}

	// Set the sandbox time limit to match the problem time limit (with some buffer)
	j.sandbox.TimeLimit = timeLimit

	// Execute the code
	execResult, err := j.sandbox.Execute(submission.Code, testCase.Input)
	if err != nil {
		result.Status = StatusInternalError
		return result
	}

	// Store the execution result
	result.Output = execResult.Output
	result.RunTime = execResult.RunTime

	// Determine the status based on execution result
	switch execResult.Status {
	case "Compilation Error":
		result.Status = StatusCompileError
	case "Runtime Error":
		result.Status = StatusRuntimeError
	case "Time Limit Exceeded":
		result.Status = StatusTimeLimitExceeded
	case "Success":
		// Compare output with expected output
		if sandbox.CompareOutput(testCase.Output, execResult.Output) {
			result.Status = StatusAccepted
		} else {
			result.Status = StatusWrongAnswer
		}
	default:
		result.Status = StatusInternalError
	}

	return result
}
