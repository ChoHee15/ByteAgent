# 0010-demo-scenarios

## Background

- 当前项目已经具备 CLI、REPL、bash 工具、写入审批和测试分层等核心能力，但缺少一组可以直接复现的 demo。
- 新用户很难快速验证 agent 是否真的能读代码、修小 bug、跑测试，以及在写入前触发审批。

## Goals

- 提供一组可直接复制执行的 demo 场景
- 至少覆盖读 repo、修小 bug 并跑测试、写入审批三类能力
- 为“修 bug”场景提供一个与主仓库隔离的可控 fixture

## Non-Goals

- 不新增新的 CLI 功能或新的审批机制
- 不要求 demo 覆盖所有 agent 能力
- 不把 demo fixture 纳入主仓库默认回归测试

## Behavior

- 仓库提供 `demos/README.md` 作为 demo 总入口。
- demo 文档应说明每个场景的目标、推荐工作目录、建议 prompt 和期望观察点。
- “读 repo 并解释架构” demo 以当前仓库为 workspace。
- “修小 bug 并跑测试” demo 使用一个独立的 bugfix fixture workspace，并包含一个当前会失败的测试。
- “写入审批” demo 说明如何触发审批，以及推荐使用安全的 mutating 命令来演示相同审批路径。
- README 应提供 demo 文档入口，方便首次使用者发现。

## Edge Cases

- bugfix fixture 必须与主仓库的默认 `make test` 隔离，避免故意失败的测试污染正常回归。
- 写入审批 demo 默认应使用低风险命令进行演示；如果用户想验证更危险的命令，应在 disposable workspace 中进行。
- demo 文档应显式说明哪些失败是“预期用于演示”的，而不是仓库 regression。

## Acceptance Criteria

- [x] 存在 `demos/README.md` 作为 demo 入口
- [x] README 提供 demo 文档入口
- [x] 仓库包含一个独立的 bugfix fixture，且其初始测试失败是可复现的
- [x] demo 文档覆盖读 repo、修 bug 跑测试、写入审批三类场景

## Test Plan

- 默认层自动化测试：运行 `make test`
- 手工验证：
  - 阅读 `demos/README.md`
  - 在 demo fixture 目录下执行 `go test ./...`，确认初始状态存在预期失败

## Notes

- 该 spec 约束 demo 资产和文档，不改变主程序运行时语义。
