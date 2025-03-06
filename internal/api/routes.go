package api

import (
	"net/http"
	"strings"
)

// SetupRoutes configures the API routes using standard Go 1.21 patterns
func SetupRoutes(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// 添加一个API路由来查看系统中的所有题目详细信息（包括测试用例）
	mux.HandleFunc("/api/problems/debug", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAllProblemsDebug(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 批量导入题目路由
	mux.HandleFunc("/api/problems/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.ImportProblems(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// AI生成题目路由
	mux.HandleFunc("/api/problems/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.GenerateAIProblem(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 保存生成的题目路由
	mux.HandleFunc("/api/problems/save-generated", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.SaveGeneratedProblem(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 大纲知识点获取路由
	mux.HandleFunc("/api/outline/knowledge-points", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetOutlineKnowledgePoints(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 带用户状态的所有题目路由
	mux.HandleFunc("/api/problems/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAllProblemsWithStatus(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 大纲解析路由
	mux.HandleFunc("/api/outline/parse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ParseOutline(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 生成大纲题目路由
	mux.HandleFunc("/api/outline/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.GenerateOutlineQuestions(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 测试相关路由
	mux.HandleFunc("/api/quizzes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAllQuizzes(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateQuiz(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 测试详情路由
	mux.HandleFunc("/api/quizzes/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/quizzes/")

		// 处理提交答案
		if strings.HasSuffix(path, "/submit") {
			if r.Method == http.MethodPost {
				handler.SubmitQuizAnswer(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// 处理获取测试结果
		if strings.HasSuffix(path, "/results") {
			if r.Method == http.MethodGet {
				handler.GetQuizResults(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// 获取测试详情
		if r.Method == http.MethodGet {
			handler.GetQuiz(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 用户测试结果路由
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")

		// 处理获取用户测试结果
		if strings.HasSuffix(path, "/quizzes") {
			if r.Method == http.MethodGet {
				handler.GetUserQuizResults(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// 原有用户相关路由
		if r.Method == http.MethodGet {
			handler.GetUser(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Problem routes
	mux.HandleFunc("/api/problems", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetProblems(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateProblem(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Problem detail and testcase routes
	mux.HandleFunc("/api/problems/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/problems/")

		// 处理带状态的题目详情
		if strings.HasSuffix(path, "/status") {
			problemPath := strings.TrimSuffix(path, "/status")
			r.URL.Path = "/api/problems/" + problemPath
			if r.Method == http.MethodGet {
				handler.GetProblemWithStatus(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// Handle testcase routes
		if strings.HasSuffix(path, "/testcases") {
			if r.Method == http.MethodPost {
				handler.AddTestCase(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// Handle submission routes
		if strings.HasSuffix(path, "/submissions") {
			if r.Method == http.MethodPost {
				handler.SubmitSolution(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// Handle problem detail
		if r.Method == http.MethodGet {
			handler.GetProblem(w, r)
		} else if r.Method == http.MethodPut {
			handler.UpdateProblem(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Submission routes
	mux.HandleFunc("/api/submissions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetSubmission(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// User routes
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.RegisterUser(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Static file serving for web frontend
	fileServer := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve the main HTML page for all other routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "./web/templates/index.html")
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}
