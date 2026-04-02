# 0008-openai-model-integration

## Background

- 当前模型接入层非常薄，只负责把运行时配置映射到 Eino 的 OpenAI ChatModel。
- 该层虽简单，但仍然是外部依赖边界，适合有一份最小 spec 说明其契约。

## Goals

- 固定 OpenAI 兼容模型初始化的当前参数映射
- 固定真实 smoke test 的最小验证目标

## Non-Goals

- 不定义具体模型选择策略
- 不评估回答质量、成本或延迟

## Behavior

- `NewOpenAI` 使用 `Config` 中的 `APIKey`、`Model` 和 `BaseURL` 创建 Eino OpenAI ChatModel。
- 当 `BaseURL` 为空时，底层使用默认端点。
- 初始化失败时，函数返回带上下文的错误，不进行重试。
- 真实 smoke test 位于 `integration` build tag 下，不进入默认回归层。
- smoke test 在缺少 `OPENAI_API_KEY` 时直接跳过，不视为失败。
- smoke test 默认模型为 `gpt-4o-mini`，并验证模型能够返回非空 assistant 文本。

## Edge Cases

- 该层不主动校验 `APIKey` 是否为空；通常由更上层配置加载先做必填校验。
- 该层不增加额外的超时、重试或 fallback 逻辑；其行为主要受调用方 `context` 控制。

## Acceptance Criteria

- [x] `Config` 中的 `APIKey`、`Model`、`BaseURL` 会被传入底层 chat model
- [x] integration smoke test 缺少 API key 时自动跳过
- [x] integration smoke test 成功路径要求拿到非空回复

## Test Plan

- 默认层自动化测试：无单独 unit test，依赖集成层验证
- 可选 integration / 手工验证：`OPENAI_API_KEY=... make test-integration`

## Notes

- 该 spec 当前刻意保持很薄，只记录外部模型接入契约。
