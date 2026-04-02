# 0007-agent-run-limits-and-user-facing-errors

## Background

- 当前 agent 执行链已经具备迭代上限、无结果保护和终端进度提示，但这些用户可见语义散落在 `internal/app` 和 `internal/agent` 中。
- 需要把这些行为沉淀为单独 spec，避免后续改动时破坏用户提示。

## Goals

- 固定单次 agent 任务的结果提取和错误翻译行为
- 固定 progress indicator 的显示条件

## Non-Goals

- 不定义模型推理质量
- 不定义工具调用策略或 prompt 优化策略

## Behavior

- 创建 agent 时，若传入的 `maxIterations < 1`，则回退到 `26`。
- `ask()` 会遍历 agent 事件流，并记录最后一个非空 assistant 文本作为最终回答。
- 若事件流中出现 `event.Err`，CLI 会优先返回错误，而不是继续消费后续消息。
- 若底层错误文本包含 `exceeds max iterations`，CLI 会将其翻译为更面向用户的错误：
  `task exceeded max iterations (<当前配置值>); try narrowing the request or increasing CODE_AGENT_MAX_ITERATIONS`
- 若事件流结束后没有拿到非空 assistant 文本，则返回 `agent returned no assistant message`。
- 当 stdout 是终端时，CLI 在执行 `ask()` 期间显示 progress indicator；任务结束后停止显示。
- 当 stdout 不是终端时，默认不显示 progress indicator。

## Edge Cases

- progress indicator 的显示依据只看输出是否为终端；测试中可通过 `forceProgress` 强制开启。
- 若审批写入前需要用户输入，CLI 会先暂停 progress indicator，审批结束后再恢复。
- 当前最终回答采用“最后一个非空 assistant 文本”，并不尝试拼接多个 assistant chunk。

## Acceptance Criteria

- [x] 非法 `maxIterations` 会回退到默认值
- [x] `ask()` 返回最后一个非空 assistant 文本
- [x] 超过最大迭代的错误会被翻译为更清楚的用户提示
- [x] 没有 assistant 文本时返回明确错误
- [x] 非终端输出默认不显示 progress indicator

## Test Plan

- 默认层自动化测试：`go test ./internal/app ./internal/agent -run 'TestAsk|TestAskShowsProgressForTerminalOutput|TestNew'`
- 手工验证：提交一个明显需要很多轮工具调用的请求，观察超限报错

## Notes

- 当前错误翻译依赖字符串匹配，不是结构化错误类型。
