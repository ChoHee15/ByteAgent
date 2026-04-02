# 0011-task-suite

## Background

- 当前仓库已经有 `demos/` 可用于人工演示，但还缺少一组更适合重复执行、可脚本验证的 coding tasks。
- 仅依赖主仓库 `make test`，只能验证代码没有回归，不能验证 agent 是否真的具备读仓库、修 bug、按 spec 改代码和补测试的能力。

## Goals

- 提供一个顶层 `task-suite/` 目录，存放可验证的 agent 任务
- 每个任务都包含独立工作区模板和明确的验收脚本
- 首批至少覆盖仓库阅读、bugfix+测试、按 spec 实现小功能三类能力
- 提供自动 runner，减少手工复制工作区、执行 agent 和收集结果的重复操作

## Non-Goals

- 不替代主仓库的默认 `make test`
- 不将 task suite 本身接入默认自动化回归
- 不要求首批任务覆盖审批交互或长程多小时任务

## Behavior

- 仓库提供顶层 `task-suite/README.md` 作为 task suite 入口。
- 每个任务目录至少包含：
  - `task.md`
  - `workspace-template/`
  - `verify.sh`
- `workspace-template/` 用于复制出一次性评测工作区，避免污染任务模板。
- `verify.sh` 接收目标工作区路径，并根据任务类型执行自动验收。
- 仓库提供一个辅助脚本，帮助从 `workspace-template/` 复制出新的临时工作区。
- 仓库提供 `run-task.sh`，自动完成单个任务的工作区复制、agent 执行、日志落盘和验收。
- 仓库提供 `run-all.sh`，顺序执行全部任务，并输出汇总结果。
- 自动 runner 会为每次任务运行保存 prompt、workspace、agent stdout/stderr、verify stdout/stderr 和机器可读的结果摘要。
- 自动 runner 会显式为 agent 进程注入 `CODE_AGENT_UNSAFE_AUTO_APPROVE_BASH_WRITES=1`，以便在非交互评测中允许文件修改。
- README 应提供 task suite 的入口，并说明它与 `demos/` 的区别。

## Edge Cases

- 任务模板中的独立 Go module 不得破坏主仓库 `make test`。
- 仓库阅读类任务若无法完全自动评分，应通过固定输出文件名、固定标题和关键路径要求降低主观性。
- 任务应优先使用小型、可重复、低污染的 fixture，避免依赖外部服务。
- 自动 runner 默认应将运行产物落到一次性目录，而不是污染仓库工作区。

## Acceptance Criteria

- [x] 存在 `task-suite/README.md` 作为入口
- [x] 存在用于复制工作区模板的辅助脚本
- [x] 至少存在 3 个首批任务，分别覆盖仓库阅读、bugfix+测试、按 spec 改代码
- [x] 每个首批任务都带有 `verify.sh`
- [x] README 提供 task suite 入口
- [x] 存在 `run-task.sh` 自动执行单个任务
- [x] 存在 `run-all.sh` 自动执行全部任务
- [x] 自动 runner 会保存日志和机器可读结果摘要

## Test Plan

- 默认层自动化测试：运行 `make test`
- 手工验证：
  - 使用 `run-task.sh` 对单个任务做一次 smoke run
  - 使用 `run-all.sh` 对全部任务做一次批量运行

## Notes

- task suite 面向“可重复评测”，而 `demos/` 更偏“人工演示和试用”。
