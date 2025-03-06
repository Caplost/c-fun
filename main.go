package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/cppjudge/internal/api"
	"github.com/user/cppjudge/internal/db"
	"github.com/user/cppjudge/internal/judge"
	"github.com/user/cppjudge/internal/models"
	"github.com/user/cppjudge/internal/sandbox"
)

func main() {
	// 启用详细日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("C++在线评测系统启动中...")

	// 创建内存存储
	store := db.NewMemoryStore()

	// 初始化沙箱
	sandbox, err := sandbox.NewCppSandbox()
	if err != nil {
		log.Fatalf("Failed to initialize sandbox: %v", err)
	}
	defer sandbox.Cleanup()

	// 初始化判题器
	judgeService := judge.NewJudge(store, sandbox)

	// 创建API处理器
	handler := api.NewHandler(store, judgeService)

	// 设置路由
	mux := api.SetupRoutes(handler)

	// 添加示例数据
	addSampleData(store)

	// 设置优雅关闭
	stopChan := setupGracefulShutdown()

	// 启动服务器
	port := "8089" // 更改端口为8089
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		// 添加超时设置
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 300 * time.Second, // 增加写超时时间到5分钟
		IdleTimeout:  120 * time.Second, // 闲置连接超时时间
		// 允许更多连接和更长的请求时间，适合AI生成任务
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Printf("服务器启动成功: http://localhost:%s", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器错误: %v", err)
		}
	}()

	// 等待关闭信号
	<-stopChan

	log.Println("正在关闭服务器...")

	// 创建带超时的上下文用于关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭异常: %v", err)
	}

	log.Println("服务器已安全关闭")
}

// setupGracefulShutdown 设置信号处理来处理优雅关闭
func setupGracefulShutdown() <-chan struct{} {
	stopChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		close(stopChan)
	}()

	return stopChan
}

// addSampleData adds sample problems and test cases for demo purposes
func addSampleData(store *db.MemoryStore) {
	// Add a test user
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password", // In a real app, this would be hashed
	}
	_, err := store.AddUser(user)
	if err != nil {
		log.Printf("Error adding test user: %v", err)
	}

	// 我们不再在这里添加样例问题，因为现在由InMemoryProblemStore自己管理
	// 当存储为空时会自动添加样例问题

	log.Println("示例数据处理完成")
}
