# 0005-write-approval-and-rejection-semantics

## Background

- 项目已经支持在 bash 执行前拦截潜在写入命令，但这部分规则横跨 `internal/tool` 和 `internal/app`。
- 为避免未来修改写入确认流程时破坏现有安全边界，需要先把当前语义固定下来。

## Goals

- 明确哪些 bash 命令会触发确认
- 明确批准、拒绝和无交互终端时的行为差异

## Non-Goals

- 不追求精确识别所有可能写入磁盘的 shell 语法
- 不定义未来更细粒度的沙箱或权限系统

## Behavior

- bash 工具会基于关键字启发式识别潜在写入命令。
- 当前会触发确认的典型命令包括重定向、`tee`、`touch`、`mkdir`、`rm`、`mv`、`cp`、`chmod`、`git apply`、`git commit`、`git checkout`、`git reset`、`patch` 等。
- 只读命令如 `pwd`、`cat README.md` 默认不触发确认。
- 若命令被识别为潜在写入，而工具未配置审批回调，则直接返回错误，拒绝执行。
- 默认 CLI bootstrap 会为 bash 工具注入写入审批回调。
- 审批前会暂停正在显示的 progress indicator，避免提示被 spinner 干扰。
- CLI 审批提示会显示工作目录、原始命令，并提示用户输入 `Proceed? [y/N]: `。
- 仅当用户输入 `y` 或 `yes`（大小写不敏感）时，命令才被批准执行。
- 任何其他输入都视为拒绝。
- 读取到 EOF 时，审批结果视为拒绝，不额外报错。

## Edge Cases

- 若当前 stdin/stdout 不是交互终端，则默认无法进行写入审批，命令会被拒绝并返回交互终端相关错误。
- 在交互模式下，如果写入未获批准，当前 agent 任务被取消，并向用户输出“当前任务已取消”的提示；REPL 本身继续运行。
- 在单次命令模式下，如果写入未获批准，则将错误返回给上层，CLI 以失败结束。
- 启发式识别并不完整；当前 spec 固定的是“已有 marker 列表”，不是“所有真实写入都能被识别”。

## Acceptance Criteria

- [x] `touch`、重定向和 `mkdir` 等命令会触发审批
- [x] 只读命令不会触发审批
- [x] 审批通过后，命令实际执行
- [x] 审批拒绝后，命令不会执行
- [x] 无交互终端时，潜在写入命令会被拒绝
- [x] REPL 中拒绝写入只取消当前任务，不退出程序

## Test Plan

- 默认层自动化测试：`go test ./internal/tool ./internal/app -run 'TestCommandNeedsApproval|TestNewBashToolWithApproval|TestConfirmMutatingCommand|TestHandleInteractiveError'`
- 手工验证：在 REPL 中请求创建文件并分别输入 `yes` 与 `n`

## Notes

- 当前写入识别策略偏保守但不完备，未来若更换识别器，应先更新本 spec。
