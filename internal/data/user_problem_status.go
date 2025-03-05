package data

import (
	"fmt"
	"log"
	"sync"
	"time"
)

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

// UserProblemStatusStore 接口定义用户题目状态的存储方法
type UserProblemStatusStore interface {
	GetUserProblemStatus(userID, problemID int) (*UserProblemStatus, error)
	GetUserProblemStatuses(userID int) ([]*UserProblemStatus, error)
	GetProblemUserStatuses(problemID int) ([]*UserProblemStatus, error)
	UpdateUserProblemStatus(status *UserProblemStatus) (*UserProblemStatus, error)
}

// InMemoryUserProblemStatusStore 实现了用户题目状态的内存存储
type InMemoryUserProblemStatusStore struct {
	statuses    map[string]*UserProblemStatus // key: "userID-problemID"
	statusID    int
	mu          sync.RWMutex
	persistence *PersistenceManager
	enableSave  bool
}

// NewInMemoryUserProblemStatusStore 创建新的用户题目状态存储
func NewInMemoryUserProblemStatusStore() *InMemoryUserProblemStatusStore {
	// 创建持久化管理器
	persistenceManager, err := NewPersistenceManager()
	if err != nil {
		log.Printf("警告: 创建持久化管理器失败: %v, 将使用内存存储", err)
		persistenceManager = nil
	}

	store := &InMemoryUserProblemStatusStore{
		statuses:    make(map[string]*UserProblemStatus),
		statusID:    0,
		persistence: persistenceManager,
		enableSave:  false, // 初始禁用保存，等加载完成再启用
	}

	// 尝试从文件加载数据
	if persistenceManager != nil {
		store.loadData()
	}

	// 启用保存
	store.enableSave = true

	return store
}

// 生成状态key
func generateStatusKey(userID, problemID int) string {
	return fmt.Sprintf("%d-%d", userID, problemID)
}

// loadData 从文件加载数据
func (s *InMemoryUserProblemStatusStore) loadData() {
	if s.persistence == nil {
		return
	}

	// 加载用户题目状态
	statuses, err := s.persistence.LoadUserProblemStatuses()
	if err != nil {
		log.Printf("加载用户题目状态数据失败: %v", err)
	} else {
		s.statuses = statuses
		// 找出最大ID
		for _, status := range statuses {
			if status.ID > s.statusID {
				s.statusID = status.ID
			}
		}
		log.Printf("从存储中加载 %d 个用户题目状态", len(statuses))
	}
}

// saveData 保存数据到文件
func (s *InMemoryUserProblemStatusStore) saveData() {
	if s.persistence == nil || !s.enableSave {
		log.Printf("跳过保存用户题目状态: persistence=%v, enableSave=%v", s.persistence != nil, s.enableSave)
		return
	}

	log.Printf("开始保存用户题目状态数据到文件... (状态数: %d)", len(s.statuses))

	// 保存用户题目状态
	if err := s.persistence.SaveUserProblemStatuses(s.statuses); err != nil {
		log.Printf("保存用户题目状态数据失败: %v", err)
	}
}

// GetUserProblemStatus 获取特定用户对特定题目的状态
func (s *InMemoryUserProblemStatusStore) GetUserProblemStatus(userID, problemID int) (*UserProblemStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := generateStatusKey(userID, problemID)
	status, exists := s.statuses[key]
	if !exists {
		return nil, ErrNotFound
	}

	return status, nil
}

// GetUserProblemStatuses 获取特定用户的所有题目状态
func (s *InMemoryUserProblemStatusStore) GetUserProblemStatuses(userID int) ([]*UserProblemStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*UserProblemStatus
	for _, status := range s.statuses {
		if status.UserID == userID {
			result = append(result, status)
		}
	}

	return result, nil
}

// GetProblemUserStatuses 获取特定题目的所有用户状态
func (s *InMemoryUserProblemStatusStore) GetProblemUserStatuses(problemID int) ([]*UserProblemStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*UserProblemStatus
	for _, status := range s.statuses {
		if status.ProblemID == problemID {
			result = append(result, status)
		}
	}

	return result, nil
}

// UpdateUserProblemStatus 更新用户题目状态
func (s *InMemoryUserProblemStatusStore) UpdateUserProblemStatus(status *UserProblemStatus) (*UserProblemStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := generateStatusKey(status.UserID, status.ProblemID)
	existingStatus, exists := s.statuses[key]

	if exists {
		// 更新现有状态
		status.ID = existingStatus.ID
		status.CreatedAt = existingStatus.CreatedAt
		status.UpdatedAt = time.Now()
	} else {
		// 创建新状态
		s.statusID++
		status.ID = s.statusID
		status.CreatedAt = time.Now()
		status.UpdatedAt = time.Now()
	}

	s.statuses[key] = status
	log.Printf("更新用户题目状态: 用户ID=%d, 题目ID=%d, 已尝试=%v, 已解决=%v, 失败次数=%d",
		status.UserID, status.ProblemID, status.Attempted, status.Solved, status.FailedAttempts)

	// 保存数据
	s.saveData()

	return status, nil
}
