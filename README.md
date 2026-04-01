# code agent

一个轻量的 Go CLI code agent，基于 [CloudWeGo Eino](https://github.com/cloudwego/eino) 构建，支持通过大模型自主调用本地 `bash` 工具。

## 功能

- 命令行单次提问
- 交互式 REPL
- 基于 Eino ADK 的 `ChatModelAgent`
- 本地 `bash` 工具执行，默认限制在当前工作区内

## 环境变量

```bash
export OPENAI_API_KEY=your_api_key
export OPENAI_MODEL=gpt-4o-mini
# 可选
export OPENAI_BASE_URL=
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

## 测试

默认测试层只运行 mock/local tests，不依赖真实大模型 API：

```bash
make test
```

如果只想执行默认分层中的单元测试与本地集成测试，也可以直接运行：

```bash
make test-unit
```

真实模型连通性验证放在独立 smoke test 层，只有在显式提供 API 配置时才应运行：

```bash
export OPENAI_API_KEY=your_api_key
make test-integration
```

`make test-integration` 当前包含两类真实链路验证：

- 模型接入 smoke test：验证 OpenAI/Eino 模型初始化与最小 `Generate()` 调用成功
- code agent 链路测试：验证完整 `agent + bash tool + runner` 能在临时工作区内读取文件并生成包含文件名与内容的回答

需要完整验证两层测试时再运行：

```bash
make test-all
```
