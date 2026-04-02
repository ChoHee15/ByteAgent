# Bugfix Fixture

这个目录是一个独立的 Go module，用于演示 agent 的“读代码 -> 跑测试 -> 修 bug -> 再跑测试”能力。

当前状态：

- `go test ./...` 会失败，这是故意保留的演示 bug
- 修复后，测试应恢复通过
- 该 fixture 不进入主仓库的默认 `make test`
