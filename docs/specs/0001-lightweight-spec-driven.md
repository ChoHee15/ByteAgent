# 0001-lightweight-spec-driven

## Background

- 当前项目已经具备测试、CHANGELOG、TODO 和提交纪律，但在实现新功能时仍以对话驱动为主。
- 对于交互语义、错误语义和边界行为，缺少一个统一的规格落点，导致需求澄清常常发生在实现之后。

## Goals

- 为项目建立一套轻量 spec-driven 工作流
- 让功能实现前先固定核心语义和验收标准
- 提供可直接复用的 spec 模板，降低写 spec 的成本

## Non-Goals

- 不引入重型设计审批流程
- 不要求每次小改动都写长文档
- 不要求引入额外的 spec 生成工具链

## Behavior

- 新功能或重要行为调整开始前，开发者先在 `docs/specs/` 下创建或更新一个 spec
- spec 应包含：背景、目标、非目标、行为定义、边界情况、验收标准
- 开发实现、测试和文档更新必须与该 spec 保持一致
- 当实现中途改变语义时，必须先更新 spec，再视为完成

## Edge Cases

- 对于原型探索中的需求，spec 可以简化，但仍需明确要验证的假设和成功标准
- 对于纯文档修改、拼写修复等非行为变更，可不新增 spec，但如影响现有行为描述仍应更新相关 spec
- 对于单纯内部重构，如果不改变外部行为，通常不需要新 spec

## Acceptance Criteria

- [x] 仓库中存在 `docs/specs/README.md` 说明流程
- [x] 仓库中存在 `docs/specs/TEMPLATE.md` 作为模板
- [x] `AGENTS.MD` 明确要求在实现前补充轻量 spec
- [x] README 提供 spec 文档入口

## Test Plan

- 默认层自动化测试：运行 `make test`
- 手工验证：检查 README、AGENTS 和 `docs/specs/` 中的说明是否一致

## Notes

- 该流程刻意偏轻，目标是减少返工，而不是增加流程负担
