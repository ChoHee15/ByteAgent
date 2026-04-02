# Demo Prompt: Write Approval

推荐在交互模式下运行，用低风险命令验证审批链路。

```text
请在当前 workspace 根目录执行 touch approval-demo.txt。
```

如果你想验证删除类命令，可以把 `touch approval-demo.txt` 换成 disposable workspace 中的 `rm -f some-temp-file`，但默认应选择拒绝。
