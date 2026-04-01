# 任务目标
为当前基于 `eino` 框架的 Code Agent 项目编写完善的自动化测试。

# 测试范围与要求
1. **默认测试层 (必写、默认执行)：**
   - 默认通过 `make test` 运行，只包含 mock/unit tests 和本地可控的集成测试。
   - 该层禁止真实外部网络请求，必须稳定、可重复、适合日常开发回归。

2. **工具层测试 (Bash Executor)：**
   - 使用表驱动测试 (Table-driven tests) 覆盖正常命令执行、命令执行失败、以及命令执行超时的边界情况。
   - 注意：不要在测试中执行具有破坏性的真实 bash 命令，可以使用 `echo` 或 `sleep` 来模拟。

3. **Agent / App 逻辑层测试 (Mocking eino)：**
   - 针对调用大模型的部分，**严禁在测试中发起真实的外部网络请求**。
   - 优先在项目边界注入 fake runner / fake model，而不是把真实 OpenAI 调用带进单元测试。
   - 需要模拟普通文本响应、错误返回，以及空响应等边界情况。

4. **真实 API 测试层 (可选)：**
   - 真实大模型 API 只能出现在带 `integration` build tag 的 smoke test 中。
   - 该层默认不进入 `make test`，仅通过 `make test-integration` 或 `make test-all` 运行。
   - 只验证最小链路可用和响应非空，不校验具体措辞。

5. **代码规范：**
   - 测试文件必须与被测试文件在同一包下，命名规范为 `xxx_test.go`。
   - 在不会污染全局状态时使用 `t.Parallel()` 提升测试执行速度。
