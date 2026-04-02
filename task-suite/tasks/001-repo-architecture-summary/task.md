请阅读当前工作区，并在仓库根目录新增 `architecture-summary.md`。

要求：

1. 使用以下固定标题：
   - `## Entry Point`
   - `## Config`
   - `## Runner`
   - `## Tooling`
   - `## Request Flow`
2. 在文档中明确提到这些路径：
   - `cmd/mini-agent/main.go`
   - `internal/config/config.go`
   - `internal/runner/runner.go`
   - `internal/tool/echo.go`
3. `Request Flow` 段落需要描述一次请求从 CLI 入口到 runner 再到 tool 的链路。
4. 不要修改其他业务代码。
