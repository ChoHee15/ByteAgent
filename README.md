# code agent

一个轻量的 Go CLI code agent，基于 [CloudWeGo Eino](https://github.com/cloudwego/eino) 构建，支持通过大模型自主调用本地 `bash` 工具。

## 功能

- 命令行单次提问
- 交互式 REPL
- 基于 Eino ADK 的 `ChatModelAgent`
- 本地 `bash` 工具执行，默认限制在当前工作区内

## 环境变量

```bash
# export OPENAI_API_KEY=your_api_key
sk-9dcc28e2feca43878a83b93591ad436d
export OPENAI_MODEL=deepseek-chat
# 可选
export OPENAI_BASE_URL=https://api.deepseek.com
```

## 运行

```bash
go run ./cmd/code-agent -- "帮我查看当前目录下有哪些文件"
```

或者：

```bash
code-agent -i
```

## 可调参数

```bash
export CODE_AGENT_MAX_HISTORY_TURNS=8
export CODE_AGENT_COMMAND_TIMEOUT_SEC=120
export CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES=32768
```
