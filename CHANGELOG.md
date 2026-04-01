## 2026-04-01

- 增加分层测试策略：默认通过 `make test` 运行 mock/local tests，通过 `make test-integration` 运行真实大模型 smoke test。
- 为 `internal/app`、`internal/tool`、`internal/config` 和 `internal/agent` 增加自动化测试，覆盖 CLI 帮助输出、交互退出、bash 超时与越界路径、非零退出码、输出截断、配置默认值与非法环境变量回退等边界情况。
- 新增带 `integration` build tag 的 OpenAI smoke test，在未提供 `OPENAI_API_KEY` 时跳过，避免真实 API 调用进入默认回归。
- 新增真实 `code agent` 链路 integration test，使用临时工作区和哨兵文件验证 `agent + bash tool + runner` 能完成实际工具调用，并在最终回答中返回文件名和文件内容。
- 增加提交纪律约束：功能验证成功后必须先获得用户确认，再使用 GitHub 社区常见的 Conventional Commits 风格创建 commit。
