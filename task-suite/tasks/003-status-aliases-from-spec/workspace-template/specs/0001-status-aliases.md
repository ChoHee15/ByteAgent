# 0001-status-aliases

## Background

- 这个模块用于把用户输入归一化为内部状态名。
- 现有实现只支持 `active` 和 `inactive`，但上游系统已经开始发送别名。

## Goals

- 接受 `enabled` 作为 `active` 的别名
- 接受 `disabled` 作为 `inactive` 的别名
- 对未知状态继续返回 `unknown`

## Non-Goals

- 不引入新的状态类型
- 不修改函数签名

## Behavior

- 输入大小写不敏感
- 输入前后空白需要被忽略
- `active` 和 `enabled` 都返回 `active`
- `inactive` 和 `disabled` 都返回 `inactive`
- 其他值返回 `unknown`

## Acceptance Criteria

- [x] `NormalizeStatus("enabled")` 返回 `active`
- [x] `NormalizeStatus("disabled")` 返回 `inactive`
- [x] `NormalizeStatus(" ACTIVE ")` 返回 `active`
- [x] `NormalizeStatus("mystery")` 返回 `unknown`
