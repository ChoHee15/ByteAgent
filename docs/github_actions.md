# GitHub Actions 与发布说明

该项目使用 GitHub Actions 完成自动测试、真实集成验证和发布产物构建。

## 工作流

### CI

文件：[`../.github/workflows/ci.yml`](../.github/workflows/ci.yml)

- 触发条件：向 `main` 发起 Pull Request
- 执行内容：`make test`

### Release

文件：[`../.github/workflows/release.yml`](../.github/workflows/release.yml)

- 触发条件：`main` 分支收到新的 push / merge
- 执行内容：
  - `make test`
  - `make test-integration`
- 测试与集成验证在 Linux runner 上执行
- 验证通过后构建并发布二进制归档

当前发布矩阵：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

产物会同时：

- 上传为 workflow artifacts
- 发布为 GitHub Release assets

## Actions Secrets

`release.yml` 需要以下 GitHub Actions secrets：

```bash
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-4o-mini
OPENAI_BASE_URL=
```

说明：

- `OPENAI_API_KEY`：必填，否则 release 工作流会失败
- `OPENAI_MODEL`：可选，未设置时默认使用 `gpt-4o-mini`
- `OPENAI_BASE_URL`：可选，使用兼容 OpenAI 接口时配置

## 分支要求

当前工作流监听的是 `main` 分支：

- PR 到 `main` 会触发 `CI`
- Push 到 `main` 会触发 `Release`

如果仓库默认分支仍是 `master`，需要先完成分支切换，否则自动流程不会按预期触发。
