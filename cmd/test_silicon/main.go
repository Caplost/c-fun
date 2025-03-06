package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/user/cppjudge/internal/config"
	"github.com/user/cppjudge/internal/llm"
)

func main() {
	// 设置详细日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("测试硅基流动API调用")

	// 加载Silicon Flow配置
	configPath, err := filepath.Abs("config/deepseek_silicon_config.json")
	if err != nil {
		log.Fatalf("获取配置路径失败: %v", err)
	}

	llmConfig, err := config.LoadLLMConfig(configPath)
	if err != nil {
		log.Fatalf("加载LLM配置失败: %v", err)
	}

	// 创建DeepSeek客户端
	client, err := llm.Factory(llmConfig)
	if err != nil {
		log.Fatalf("创建LLM客户端失败: %v", err)
	}

	// 准备测试消息
	messages := []llm.Message{
		{
			Role:    "user",
			Content: "What opportunities and challenges will the Chinese large model industry face in 2025?",
		},
	}

	// 调用API
	log.Println("正在调用硅基流动API...")
	response, err := client.GenerateCompletion(messages)
	if err != nil {
		log.Fatalf("API调用失败: %v", err)
	}

	// 打印响应
	fmt.Println("\n===API响应===")
	fmt.Println(response)
	fmt.Println("===响应结束===")

	log.Println("测试完成")
}
