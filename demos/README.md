# Demo 场景

本目录提供一组可复制的 demo，用来快速验证 `code-agent` 当前的核心能力。

## 前置条件

- 已设置 `OPENAI_API_KEY`
- 已在仓库根目录编译出二进制，或准备直接使用 `go run`

示例：

```bash
cd /codes/code_agent
go build -o ./dist/code-agent ./cmd/code-agent
```

如果你更喜欢直接运行源码版本，也可以用：

```bash
go run ./cmd/code-agent -- "你的问题"
```

## Demo 1: 读仓库并解释架构

工作目录：仓库根目录 `/codes/code_agent`

建议 prompt 见 [`architecture-tour.md`](./architecture-tour.md)。

推荐命令：

```bash
cd /codes/code_agent
./dist/code-agent -- "请阅读当前仓库并解释整体架构。重点说明 cmd、internal/app、internal/agent、internal/tool、internal/config、internal/model 的职责边界，以及一次用户请求从 CLI 入口到模型调用再到 bash 工具执行的链路。请引用你实际查看过的文件路径。"
```

观察点：

- agent 会主动读取多个文件，而不是空泛概括
- 最终回答能区分 CLI、配置、agent、model、tool 的职责
- 回答会基于实际文件，而不是编造不存在的模块

## Demo 2: 修一个小 bug 并跑测试

工作目录：[`bugfix-fixture`](./bugfix-fixture)

这个 fixture 是一个独立 Go module，不进入主仓库默认 `make test`。它故意保留了一个小 bug，初始 `go test ./...` 会失败，这是预期行为。

建议 prompt 见 [`bugfix-and-test.md`](./bugfix-and-test.md)。

推荐命令：

```bash
cd /codes/code_agent/demos/bugfix-fixture
/codes/code_agent/dist/code-agent -- "先运行 go test ./...，定位失败原因，修复 bug，并再次运行 go test ./...。不要做与失败用例无关的重构。最后简要说明改了什么。"
```

观察点：

- agent 会先运行测试，再编辑代码
- 修改范围聚焦于失败用例，不会无关大改
- 修复后会重新运行 `go test ./...`

## Demo 3: 触发写入审批

工作目录：仓库根目录，或任意 disposable workspace

建议 prompt 见 [`write-approval.md`](./write-approval.md)。

推荐使用低风险命令验证审批链路，因为当前实现对 `touch`、`mkdir`、重定向、`rm`、`git commit` 等 mutating 命令走的是同一条审批路径。

推荐命令：

```bash
cd /codes/code_agent
./dist/code-agent -i
```

然后输入：

```text
请在当前 workspace 根目录执行 touch approval-demo.txt。如果需要写入批准，请先请求批准，不要绕过确认。
```

观察点：

- agent 在执行前会触发审批提示，而不是直接写文件
- 输入 `n` 时，当前任务被取消，但 REPL 不退出
- 输入 `y` 时，文件才会被创建

如果你想验证更危险的命令，例如 `rm`，建议只在 disposable workspace 中做，并且默认选择拒绝。
