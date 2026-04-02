# 0003-interactive-repl-and-history

## Background

- 当前项目在无 prompt 或显式 `-i` 时进入 REPL，但 REPL 的 banner、退出方式和历史拼接规则尚未文档化。
- 这些行为已经有测试覆盖，适合先反向沉淀为 `as-is spec`。

## Goals

- 固定交互模式的基础 I/O 行为
- 固定多轮历史如何参与下一轮提问

## Non-Goals

- 不定义 bash 写入审批规则
- 不定义模型回答内容质量

## Behavior

- 进入交互模式后，CLI 先输出当前 `workspace` 路径，再输出 `interactive mode` 提示。
- REPL 的用户提示符为 `> `。
- 若 stdin 和 stdout 都是终端，则优先使用 `readline`；否则回退到基于 `bufio.Reader` 的逐行读取。
- 用户输入为空字符串时，本轮被忽略，不触发 agent 调用。
- 用户输入 `exit` 或 `quit` 时，REPL 立即正常退出。
- 每轮成功完成后，将原始用户输入与最终 assistant 回复追加到历史中。
- 下一轮调用 agent 前，若历史非空，则将历史按 `User` / `Assistant` 对话格式拼接到 prompt 前面，并追加 `Current user request` 段落。
- 历史最多保留 `MaxHistoryTurns` 轮，超出后从最早的一轮开始裁剪。

## Edge Cases

- `readline` 下收到中断时，CLI 打印一个空行并重新提示，不退出 REPL。
- 读取到 EOF 时，CLI 打印一个空行后正常退出。
- 如果本轮任务因写入未获批准而失败，REPL 仅取消当前任务并继续下一轮，而不是退出整个程序。
- 如果 line editor 无法创建且当前输出不是终端，CLI 允许无 editor 运行，不把它视为 fatal。

## Acceptance Criteria

- [x] REPL 启动时输出 workspace 和 interactive banner
- [x] 输入 `exit` 时，REPL 正常退出
- [x] 空输入不会触发查询
- [x] 多轮历史会以固定格式拼接到后续 prompt 中
- [x] 历史长度超过上限后会被裁剪到最近的若干轮

## Test Plan

- 默认层自动化测试：`go test ./internal/app -run 'TestRun|TestWithHistory|TestHandleInteractiveError'`
- 手工验证：`go run ./cmd/code-agent -- -i`

## Notes

- 当前历史拼接是纯文本 prompt 拼接，不是结构化消息回放。
