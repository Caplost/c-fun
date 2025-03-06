package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LLMConfig 存储大语言模型配置
type LLMConfig struct {
	Provider     string  `json:"provider"`      // "openai" 或 "deepseek"
	APIKey       string  `json:"api_key"`       // API密钥
	APIEndpoint  string  `json:"api_endpoint"`  // API端点URL
	APIVersion   string  `json:"api_version"`   // API版本
	Model        string  `json:"model"`         // 模型名称
	MaxTokens    int     `json:"max_tokens"`    // 最大生成令牌数
	Temperature  float64 `json:"temperature"`   // 温度参数
	TopP         float64 `json:"top_p"`         // Top-p参数
	SystemPrompt string  `json:"system_prompt"` // 系统提示词
}

// DefaultLLMConfig 返回默认LLM配置
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "openai",
		APIEndpoint: "https://api.openai.com/v1/chat/completions",
		Model:       "gpt-4o",
		MaxTokens:   4000,
		Temperature: 0.7,
		TopP:        0.95,
		SystemPrompt: `你是一个专业的编程问题生成器，擅长创建有挑战性、有教育意义的编程题目。
根据用户的要求，你将生成完整的编程问题，包括详细描述、示例输入输出和测试用例。
请确保生成的问题在难度上符合要求，并且与指定的知识点相关。`,
	}
}

// LoadLLMConfig 从文件加载LLM配置
func LoadLLMConfig(configPath string) (LLMConfig, error) {
	// 默认配置
	config := DefaultLLMConfig()

	// 如果配置文件不存在，创建一个默认配置文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 确保目录存在
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return config, fmt.Errorf("无法创建配置目录: %w", err)
		}

		// 保存默认配置
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, fmt.Errorf("序列化配置失败: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return config, fmt.Errorf("保存默认配置失败: %w", err)
		}

		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// SaveLLMConfig 保存LLM配置到文件
func SaveLLMConfig(config LLMConfig, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	return nil
}
