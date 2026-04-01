## 2026-04-01

- 增加分层测试策略：默认通过 `make test` 运行 mock/local tests，通过 `make test-integration` 运行真实大模型 smoke test。
- 为 `internal/app`、`internal/tool`、`internal/config` 和 `internal/agent` 增加自动化测试，覆盖 CLI 帮助输出、交互退出、bash 超时与越界路径、非零退出码、输出截断、配置默认值与非法环境变量回退等边界情况。
- 新增带 `integration` build tag 的 OpenAI smoke test，在未提供 `OPENAI_API_KEY` 时跳过，避免真实 API 调用进入默认回归。
- 新增真实 `code agent` 链路 integration test，使用临时工作区和哨兵文件验证 `agent + bash tool + runner` 能完成实际工具调用，并在最终回答中返回文件名和文件内容。
- 增加提交纪律约束：功能验证成功后必须先获得用户确认，再使用 GitHub 社区常见的 Conventional Commits 风格创建 commit。
- 为 CLI 添加运行中加载提示；同时将默认测试入口排除 `playground/` 下的 scratch 程序，避免示例 `main` 文件破坏常规回归。
- 为 `bash` 工具增加写入前确认：检测潜在写入/修改命令时，要求用户在交互终端中显式批准；覆盖批准执行、拒绝执行、无交互终端时拒绝，以及只读命令不触发确认等边界情况。
- 调整交互模式下的写入拒绝语义：用户拒绝写入时，不再退出整个 CLI，而是取消当前 agent 任务并返回提示，允许继续下一轮输入；单次命令模式仍保留错误返回。
- 增加 GitHub CI/CD 配置：Pull Request 到 `main` 自动执行 `make test`，`main` 分支 push 自动执行测试与集成验证，并发布 Linux `amd64` / `arm64` 二进制产物和 GitHub Release 资产。
