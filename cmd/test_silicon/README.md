# 硅基流动 API 测试工具

这个简单的工具用于测试 DeepSeek 硅基流动 API 的调用。它使用 `/config/deepseek_silicon_config.json` 中的配置信息进行调用，并输出 API 返回的结果。

## 使用方法

1. 确保已经正确设置了 `config/deepseek_silicon_config.json` 文件中的 API 密钥和 Model ID。

2. 在项目根目录运行：

```bash
go run cmd/test_silicon/main.go
```

3. 程序会发送一个测试请求，然后在控制台显示 API 的响应。

## 配置信息

当前配置：
- API Endpoint: https://api.siliconflow.cn/v1/chat/completions
- Model: deepseek-ai/DeepSeek-R1

如需修改请求内容，请编辑 `main.go` 中的 `messages` 变量。 