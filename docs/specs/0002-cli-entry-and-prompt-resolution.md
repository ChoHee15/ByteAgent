# 0002-cli-entry-and-prompt-resolution

## Background

- 当前 CLI 同时支持位置参数、stdin 和交互模式，但这些入口优先级还没有独立 spec。
- 需要先把“程序在什么输入条件下进入哪种模式”固定为当前行为基线。

## Goals

- 明确 CLI 的帮助、单次提问和交互模式入口语义
- 明确 prompt 的解析优先级和空输入行为

## Non-Goals

- 不描述 REPL 内部历史拼接细节
- 不描述 bash 工具、模型接入或错误翻译细节

## Behavior

- `code-agent -h` 打印 usage 并直接返回，不执行 bootstrap。
- CLI 先解析 flag，再解析 prompt。
- 若存在位置参数，则将所有位置参数按空格拼接并 `TrimSpace`，作为最终 prompt。
- 若不存在位置参数且 stdin 不是 TTY，则读取全部 stdin 内容并 `TrimSpace`，作为最终 prompt。
- 若显式传入 `-i`，无论是否能解析出 prompt，都进入交互模式。
- 若未传入 `-i` 且最终 prompt 为空，则进入交互模式。
- 若未传入 `-i` 且最终 prompt 非空，则执行单次问答，并只输出最终 assistant 文本。

## Edge Cases

- 位置参数优先级高于 stdin；当二者同时存在时，stdin 内容被忽略。
- stdin 为空时，解析结果为空字符串，不报错；随后按“空 prompt 进入交互模式”处理。
- flag 解析失败时，先打印 usage，再返回解析错误。
- stdin `Stat` 或读取失败时，返回带上下文的错误，不降级为交互模式。

## Acceptance Criteria

- [x] `-h` 可在未初始化配置和模型的情况下打印 usage
- [x] 有位置参数时，CLI 使用参数而不是 stdin
- [x] 无位置参数且 stdin 有内容时，CLI 使用 stdin 作为 prompt
- [x] 无位置参数且 prompt 为空时，CLI 进入交互模式

## Test Plan

- 默认层自动化测试：`go test ./internal/app -run 'TestRun|TestResolvePrompt'`
- 手工验证：执行 `go run ./cmd/code-agent -- -h`、`echo "hi" | go run ./cmd/code-agent`

## Notes

- 该 spec 描述的是当前“入口选择”行为，不约束未来是否引入更多 flag。
