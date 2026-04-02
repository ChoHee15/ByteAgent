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
- 扩展发布流程的产物矩阵：在保留 Linux 测试/集成验证的前提下，额外发布 macOS `amd64` / `arm64` 二进制归档。
- 重构 `README.md`：仅保留项目情况、API 设置、参数、源码启动、release 二进制使用和测试说明；将 GitHub Actions、Secrets 和发布流水线细节迁移到 `docs/github_actions.md`。
- 优化 GitHub Actions 触发条件：纯文档修改默认不再触发 `CI` 或 `Release`；仅在源码、依赖、`Makefile` 或发布 workflow 相关文件变化时自动运行。
- 增加 `CODE_AGENT_MAX_ITERATIONS` 环境变量，将 agent 的最大内部迭代轮数从硬编码配置改为可调；同时将 `exceeds max iterations` 底层错误翻译为更清楚的 CLI 提示，指导用户缩小请求范围或调大迭代上限。
- 将默认 `CODE_AGENT_MAX_ITERATIONS` 从 12 提高到 26，并同步更新 CLI 帮助和 README 示例，降低较大代码库分析任务在默认配置下过早触发迭代上限的概率。
- 将交互式 REPL 的输入层切换到 `readline`，改善中文输入、退格和行编辑体验；同时为 REPL 和写入确认流程引入可注入的 line editor 抽象，以保持自动化测试可控。
- 将项目升级为轻量 spec-driven 工作流：新增 `docs/specs/` 目录、spec 模板和示例 spec，并在 `AGENTS.MD` 中要求功能实现前先补充轻量规格，再按 spec 编写代码与测试。
