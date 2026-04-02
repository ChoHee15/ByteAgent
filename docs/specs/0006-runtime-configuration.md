# 0006-runtime-configuration

## Background

- 项目当前通过环境变量加载模型、工作区和执行上限相关配置。
- 这些默认值和非法值回退规则已经有测试，但尚未形成独立规范。

## Goals

- 固定运行时环境变量的当前契约
- 固定默认值和非法值回退语义

## Non-Goals

- 不定义 secrets 管理方案
- 不定义未来新增配置项

## Behavior

- 程序启动时通过环境变量加载配置。
- `OPENAI_API_KEY` 为必填项；缺失时直接返回错误。
- `OPENAI_MODEL` 可选，默认值为 `gpt-4o-mini`。
- `OPENAI_BASE_URL` 可选，默认值为空。
- `WorkspaceDir` 取当前工作目录，并在加载时转为绝对路径。
- `CODE_AGENT_MAX_HISTORY_TURNS` 默认值为 `8`。
- `CODE_AGENT_MAX_ITERATIONS` 默认值为 `26`。
- `CODE_AGENT_COMMAND_TIMEOUT_SEC` 默认值为 `120`。
- `CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES` 默认值为 `32768`。
- 对于整数环境变量，若值为空或无法解析为整数，则直接使用默认值。
- 对于已经解析成功但不合法的整数值，系统执行二次兜底：
  - `MaxHistoryTurns < 1` 时回退到 `8`
  - `MaxIterations < 1` 时回退到 `26`
  - `MaxCommandBytes < 1024` 时回退到 `32768`
  - `CommandTimeoutSec < 1` 时回退到 `120`

## Edge Cases

- `OPENAI_BASE_URL` 为空并不报错，表示直接使用默认 OpenAI 端点。
- 即使设置了 `CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES=0`，最终也会回退到默认值，而不是允许无限输出。
- `WorkspaceDir` 基于进程启动时的当前目录；当前没有额外配置项允许用户覆盖它。

## Acceptance Criteria

- [x] 缺少 `OPENAI_API_KEY` 时返回明确错误
- [x] 未设置可选环境变量时使用默认值
- [x] 非法整数配置回退到默认值
- [x] `WorkspaceDir` 总是被解析为绝对路径

## Test Plan

- 默认层自动化测试：`go test ./internal/config`
- 手工验证：设置不同环境变量后执行 `go run ./cmd/code-agent -- -h` 或正常启动

## Notes

- 该 spec 固定的是“配置加载结果”，不描述模型服务端是否接受对应参数。
