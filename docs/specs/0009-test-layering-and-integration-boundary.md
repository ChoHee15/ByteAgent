# 0009-test-layering-and-integration-boundary

## Background

- 项目已经将测试拆分为默认 mock/local 层和显式 integration 层。
- 这是开发流程的重要契约，需要一份 spec 固定哪些测试默认执行，哪些测试只在显式命令下运行。

## Goals

- 固定 `make test`、`make test-integration` 和 `make test-all` 的职责边界
- 固定真实模型相关测试不进入默认回归层的原则

## Non-Goals

- 不定义 CI 触发条件细节
- 不枚举每一个测试函数的内部实现

## Behavior

- `make test-unit` 运行默认 mock/local tests，命令为 `go test $(TEST_PKGS) -v -race -cover`。
- `make test` 当前等价于 `make test-unit`。
- `TEST_PKGS` 基于 `go list ./...`，并默认排除 `/playground` 下的包。
- `make test-integration` 运行带 `integration` build tag 的真实 API smoke tests，命令包含 `-tags=integration -run Integration -count=1`。
- `make test-all` 顺序执行默认层和 integration 层。
- 默认开发回归只要求跑 `make test`；真实模型验证必须通过显式命令触发。
- 当前 integration 层至少覆盖两类验证：
  - OpenAI/Eino 模型接入 smoke test
  - 真实 `code agent + bash tool + runner` 链路测试

## Edge Cases

- 若未设置 `OPENAI_API_KEY`，integration 测试会在测试函数内部 `Skip`，而不是失败。
- integration 测试依赖真实外部服务，因此不应被纳入默认快速回归。
- 默认层包含 race 和 cover 选项，因此即便是本地 mock 测试也具有一定回归强度。

## Acceptance Criteria

- [x] `make test` 不触发真实模型 API 调用
- [x] `make test-integration` 显式运行带 `integration` tag 的测试
- [x] integration 层缺少 API key 时采用 skip 而非 fail
- [x] `make test-all` 覆盖默认层和 integration 层

## Test Plan

- 默认层自动化测试：`make test`
- 可选 integration / 手工验证：`make test-integration`、`make test-all`

## Notes

- 该 spec 是流程性 spec，目的在于保护测试边界，而不是描述某个单一产品功能。
