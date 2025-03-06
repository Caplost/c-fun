package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/user/cppjudge/internal/config"
)

// Message 表示对话消息
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // 消息内容
}

// Client 表示LLM客户端接口
type Client interface {
	// GenerateCompletion 生成文本补全
	GenerateCompletion(messages []Message) (string, error)
	// GenerateCompletionWithContext 使用指定上下文生成文本补全
	GenerateCompletionWithContext(ctx context.Context, messages []Message) (string, error)
}

// Factory 创建LLM客户端的工厂函数
func Factory(config config.LLMConfig) (Client, error) {
	switch config.Provider {
	case "openai":
		return NewOpenAIClient(config), nil
	case "deepseek":
		return NewDeepSeekClient(config), nil
	default:
		return nil, fmt.Errorf("不支持的LLM提供商: %s", config.Provider)
	}
}

// BaseClient 是所有LLM客户端的基类
type BaseClient struct {
	Config config.LLMConfig
	Client *http.Client
}

// OpenAIClient 是OpenAI API的客户端
type OpenAIClient struct {
	BaseClient
}

// NewOpenAIClient 创建一个新的OpenAI客户端
func NewOpenAIClient(config config.LLMConfig) *OpenAIClient {
	return &OpenAIClient{
		BaseClient: BaseClient{
			Config: config,
			Client: &http.Client{
				Timeout: 120 * time.Second,
			},
		},
	}
}

// OpenAIRequest 表示OpenAI API的请求
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// OpenAIResponse 表示OpenAI API的响应
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateCompletion 实现Client接口，调用OpenAI API生成文本补全
func (c *OpenAIClient) GenerateCompletion(messages []Message) (string, error) {
	// 创建一个默认的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return c.GenerateCompletionWithContext(ctx, messages)
}

// GenerateCompletionWithContext 实现Client接口，调用OpenAI API生成文本补全，支持上下文控制
func (c *OpenAIClient) GenerateCompletionWithContext(ctx context.Context, messages []Message) (string, error) {
	// 构建请求
	reqBody := OpenAIRequest{
		Model:       c.Config.Model,
		Messages:    messages,
		MaxTokens:   c.Config.MaxTokens,
		Temperature: c.Config.Temperature,
		TopP:        c.Config.TopP,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.Config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		// 检查是否是上下文取消导致的错误
		if ctx.Err() != nil {
			return "", fmt.Errorf("请求已取消或超时: %w", ctx.Err())
		}
		return "", fmt.Errorf("请求OpenAI API失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API错误: %s", openAIResp.Error.Message)
	}

	// 检查是否有结果
	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API没有返回结果")
	}

	// 返回结果
	return openAIResp.Choices[0].Message.Content, nil
}

// DeepSeekClient 是硅基流动DeepSeek API的客户端
type DeepSeekClient struct {
	BaseClient
}

// NewDeepSeekClient 创建一个新的DeepSeek客户端
func NewDeepSeekClient(config config.LLMConfig) *DeepSeekClient {
	return &DeepSeekClient{
		BaseClient: BaseClient{
			Config: config,
			Client: &http.Client{
				Timeout: 180 * time.Second,
			},
		},
	}
}

// DeepSeekRequest 表示DeepSeek API的请求
type DeepSeekRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Stop             *[]string `json:"stop,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	TopK             int       `json:"top_k,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	N                int       `json:"n,omitempty"`
	ResponseFormat   struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

// DeepSeekResponse 表示DeepSeek API的响应
type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateCompletion 实现Client接口，调用DeepSeek API生成文本补全
func (c *DeepSeekClient) GenerateCompletion(messages []Message) (string, error) {
	// 创建一个默认的上下文，增加超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	return c.GenerateCompletionWithContext(ctx, messages)
}

// GenerateCompletionWithContext 实现Client接口，调用DeepSeek API生成文本补全，支持上下文控制
func (c *DeepSeekClient) GenerateCompletionWithContext(ctx context.Context, messages []Message) (string, error) {
	// 构建请求
	reqBody := DeepSeekRequest{
		Model:            c.Config.Model,
		Messages:         messages,
		MaxTokens:        c.Config.MaxTokens,
		Stream:           false,
		Stop:             nil,
		Temperature:      c.Config.Temperature,
		TopP:             c.Config.TopP,
		TopK:             50,
		FrequencyPenalty: 0.5,
		N:                1,
	}

	// 设置响应格式
	reqBody.ResponseFormat.Type = "text"

	// 移除tools字段，因为当前模型不支持函数调用

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 打印请求体以便调试
	log.Printf("硅基流动API请求体: %s", string(jsonData))

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.Config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		// 检查是否是上下文取消导致的错误
		if ctx.Err() != nil {
			return "", fmt.Errorf("请求已取消或超时: %w", ctx.Err())
		}
		return "", fmt.Errorf("请求DeepSeek API失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印响应状态码和响应体以便调试
	log.Printf("硅基流动API响应状态码: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("响应体: %s", string(body))
		return "", fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var deepSeekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if deepSeekResp.Error != nil {
		return "", fmt.Errorf("DeepSeek API错误: %s", deepSeekResp.Error.Message)
	}

	// 检查是否有结果
	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek API没有返回结果")
	}

	// 返回结果
	return deepSeekResp.Choices[0].Message.Content, nil
}
