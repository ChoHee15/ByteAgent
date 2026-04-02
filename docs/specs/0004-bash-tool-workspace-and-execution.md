# 0004-bash-tool-workspace-and-execution

## Background

- `bash` 工具是当前项目最核心的执行能力之一。
- 其工作目录约束、超时语义和输出返回结构已经由单测覆盖，适合先沉淀为独立 spec。

## Goals

- 固定 bash 工具的 workspace 边界
- 固定命令执行、超时和返回结果语义

## Non-Goals

- 不定义哪些命令属于“需要写入审批”
- 不定义 agent 何时应该调用 bash 工具

## Behavior

- `bash` 工具输入必须包含非空 `command`；空命令直接报错。
- 若未指定 `working_directory`，则默认在 workspace 根目录执行。
- 若指定相对路径，则相对于 workspace 解析。
- 若指定绝对路径，则仅在其仍位于 workspace 内时才允许执行。
- 一旦工作目录解析到 workspace 之外，工具直接报错，不执行命令。
- 工具通过 `bash -lc <command>` 执行命令。
- 默认使用构造工具时传入的超时时间；若单次调用提供正数 `timeout_seconds`，则覆盖默认值。
- 命令成功时返回 `command`、`working_directory`、`exit_code`、`stdout`、`stderr` 和 `timed_out`。
- 命令以非零状态退出时，不把它作为工具调用错误抛出，而是返回结果对象，并在 `exit_code` 中保留原始退出码。
- 命令超时时，不返回 Go error，而是返回结果对象，设置 `timed_out=true` 且 `exit_code=-1`。
- `stdout` 和 `stderr` 分别按 `maxOutputBytes` 独立截断，超出后静默丢弃额外输出。

## Edge Cases

- `timeout_seconds <= 0` 时，不采用该覆盖值，而是回退到默认超时；若默认超时本身也非法，则最终回退到 30 秒。
- 即使 `stdout` 或 `stderr` 被截断，返回结果中也不会额外标记“已截断”。
- 工作目录校验基于解析后的绝对路径，防止通过 `..` 跳出 workspace。

## Acceptance Criteria

- [x] 空命令会被拒绝
- [x] 相对和绝对的 workspace 内路径都可正常执行
- [x] workspace 外路径会被拒绝
- [x] 非零退出码通过返回对象暴露，而不是作为调用错误
- [x] 超时命令返回 `timed_out=true` 和 `exit_code=-1`
- [x] 输出可按配置大小截断

## Test Plan

- 默认层自动化测试：`go test ./internal/tool -run 'TestResolveWorkdir|TestNewBashTool'`
- 手工验证：在临时目录中运行 `pwd`、`sleep`、`printf` 等命令

## Notes

- 当前返回结构偏“原始执行结果”，未对 stdout/stderr 做更高层解释。
