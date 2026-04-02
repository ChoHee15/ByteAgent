# Task Suite

`task-suite/` 用于存放可重复执行、可脚本验证的 agent coding tasks。

它和 [`demos/`](../demos/README.md) 的区别是：

- `demos/` 更偏人工演示和产品说明
- `task-suite/` 更偏评测，要求任务有明确输入、隔离工作区和可执行验收

## 目录约定

每个任务目录至少包含：

- `task.md`：任务描述，直接给 agent 的说明
- `workspace-template/`：任务的初始工作区模板
- `verify.sh`：对 agent 运行后的工作区执行验收

顶层 runner：

- `create-workspace.sh`：仅复制任务模板
- `run-task.sh`：自动执行单个任务
- `run-all.sh`：自动执行全部任务

当前首批任务：

- `001-repo-architecture-summary`
- `002-bugfix-greeting`
- `003-status-aliases-from-spec`

## 自动运行

推荐优先使用自动 runner。

单任务执行：

```bash
./task-suite/run-task.sh 002-bugfix-greeting /codes/code_agent/dist/code-agent
```

批量执行：

```bash
./task-suite/run-all.sh /codes/code_agent/dist/code-agent
```

默认情况下，runner 会把一次运行的产物写到 `/tmp` 下的临时目录，并打印最终 `run dir` 或 `run root`。每个任务运行目录会包含：

- `workspace/`
- `prompt.txt`
- `logs/agent.stdout.log`
- `logs/agent.stderr.log`
- `logs/verify.stdout.log`
- `logs/verify.stderr.log`
- `result.env`

自动 runner 会为 agent 进程显式设置 `CODE_AGENT_UNSAFE_AUTO_APPROVE_BASH_WRITES=1`，以便在非交互评测时允许代码修改。这个开关默认不建议在日常 CLI 使用中开启。

如果你希望把多次结果统一保存到指定目录，可以显式传第三个参数：

```bash
./task-suite/run-task.sh 003-status-aliases-from-spec /codes/code_agent/dist/code-agent /tmp/task-suite-runs
./task-suite/run-all.sh /codes/code_agent/dist/code-agent /tmp/task-suite-batch
```

## 手动运行

如果你需要手工调试某个任务，也可以继续使用模板复制脚本。

先从模板复制一份一次性工作区：

```bash
./task-suite/create-workspace.sh 001-repo-architecture-summary /tmp/code-agent-task-001
```

然后在复制出的工作区中运行你的 agent，并把对应任务的 `task.md` 作为 prompt 参考。

示例：

```bash
cd /tmp/code-agent-task-001
/codes/code_agent/dist/code-agent -- "$(cat /codes/code_agent/task-suite/tasks/001-repo-architecture-summary/task.md)"
```

任务完成后，运行验收脚本：

```bash
/codes/code_agent/task-suite/tasks/001-repo-architecture-summary/verify.sh /tmp/code-agent-task-001
```

## 首批任务说明

### 001-repo-architecture-summary

- 能力目标：读仓库、抽取模块职责、生成结构化架构说明
- 验收方式：检查固定输出文件、标题和关键路径

### 002-bugfix-greeting

- 能力目标：先跑测试、定位失败、修 bug、重新跑测试
- 验收方式：`go test ./...` 必须通过

### 003-status-aliases-from-spec

- 能力目标：先读 spec，再按 spec 改代码并补测试
- 验收方式：检查测试文件存在，`go test ./...` 通过，并校验 CLI 输出

## 设计原则

- 每个任务都应小而明确，适合 10 到 30 分钟内完成
- 默认不依赖真实网络服务
- 尽量使用独立 Go module，避免污染主仓库回归
- runner 默认把运行产物放到一次性目录，避免污染仓库工作区
