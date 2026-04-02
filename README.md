# code agent

`code-agent` 是一个基于 Go 和 [CloudWeGo Eino](https://github.com/cloudwego/eino) 的命令行 code agent。它支持单次提问和交互式 REPL，并可通过大模型调用本地 `bash` 工具完成查询或代码修改。

项目现在采用轻量 spec-driven 工作流。规格文档和模板见 [`docs/specs/README.md`](docs/specs/README.md) 与 [`docs/specs/TEMPLATE.md`](docs/specs/TEMPLATE.md)。

## ???

面试后尝试构建该项目，完成了基础功能。使用deepseek-chat进行了测试，可以完成代码编写、阅读查错。大规模代码未经过测试。
该cli工具对其所处当前目录有效。可以使用release下的二进制包。
进度记录于 [TODO.md](TODO.md) 和 [CHANGELOG.md](CHANGELOG.md) 下。
原型开发趋于稳定后，引入轻量spec流程。


## 项目情况

- 使用 Go 构建的 CLI 工具
- 基于 Eino ADK 组装 `ChatModelAgent`
- 支持本地 `bash` 工具调用
- 默认在当前 workspace 内解析工作目录
- 检测到潜在写入/修改命令时，会要求用户确认后才执行

更多仓库运维和发布流水线说明见 [`docs/github_actions.md`](docs/github_actions.md)。

## API 设置

运行前至少需要设置：

```bash
export OPENAI_API_KEY=your_api_key
```

可选配置：

```bash
export OPENAI_MODEL=gpt-4o-mini
export OPENAI_BASE_URL=
```

说明：

- `OPENAI_API_KEY`：必填
- `OPENAI_MODEL`：可选，默认 `gpt-4o-mini`
- `OPENAI_BASE_URL`：可选，使用兼容 OpenAI 接口的代理或服务时填写

## 参数与行为

命令行参数：

- `-i`：启动交互模式
- `-h`：显示帮助

可选环境变量：

```bash
export CODE_AGENT_MAX_HISTORY_TURNS=8
export CODE_AGENT_MAX_ITERATIONS=26
export CODE_AGENT_COMMAND_TIMEOUT_SEC=120
export CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES=32768
```

说明：

- `CODE_AGENT_MAX_HISTORY_TURNS`：交互模式下保留的历史轮数
- `CODE_AGENT_MAX_ITERATIONS`：单次 agent 任务允许的最大内部推理/工具调用轮数
- `CODE_AGENT_COMMAND_TIMEOUT_SEC`：单条 `bash` 命令超时时间
- `CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES`：单条 `bash` 输出的最大保留字节数

交互说明：

- 交互式 REPL 现在基于 `readline`，对中文输入、退格和行编辑更友好
- 等待模型响应时，终端会显示加载提示
- 当 agent 尝试执行可能写入或修改文件的 `bash` 命令时，会要求用户确认
- 在交互模式中拒绝写入，只会取消当前任务，不会退出整个 REPL

## 源码启动

单次提问：

```bash
go run ./cmd/code-agent -- "帮我查看当前目录下有哪些文件"
```

从标准输入读取提示词：

```bash
echo "列出当前目录文件并总结用途" | go run ./cmd/code-agent
```

交互模式：

```bash
go run ./cmd/code-agent -- -i
```

也可以先编译后再运行：

```bash
go build -o ./dist/code-agent ./cmd/code-agent
./dist/code-agent -i
```

## Demo 场景

仓库提供了一组可复现 demo，用来验证 agent 的核心能力：

- 读 repo 并解释架构
- 修一个小 bug 并跑测试
- 请求执行 mutating 命令时触发审批

入口文档见 [`demos/README.md`](demos/README.md)。

## Release 二进制使用

仓库发布产物为压缩包，当前包含：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

下载并解压后即可直接运行，例如 Linux amd64：

```bash
tar -xzf code-agent-linux-amd64.tar.gz
./code-agent-linux-amd64 -- "检查当前目录"
```

macOS arm64 示例：

```bash
tar -xzf code-agent-darwin-arm64.tar.gz
./code-agent-darwin-arm64 -i
```

二进制运行前同样需要先设置 `OPENAI_API_KEY` 等环境变量。

## 测试

默认测试层只运行 mock/local tests，不依赖真实大模型 API：

```bash
make test
```

如果只想运行默认分层中的单元测试与本地集成测试：

```bash
make test-unit
```

真实模型相关验证放在独立 integration 层，仅在显式提供 API 配置时运行：

```bash
export OPENAI_API_KEY=your_api_key
make test-integration
```

`make test-integration` 当前包含：

- OpenAI/Eino 模型接入 smoke test
- 真实 `code agent + bash tool + runner` 链路测试

完整执行两层测试：

```bash
make test-all
```
