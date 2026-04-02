# Demo Prompt: Bugfix And Test

在 [`bugfix-fixture`](./bugfix-fixture) 目录下运行。

```text
先运行 go test ./...，定位失败原因，修复 bug，并再次运行 go test ./...。不要做与失败用例无关的重构。最后简要说明改了什么、为什么原实现会失败。
```
