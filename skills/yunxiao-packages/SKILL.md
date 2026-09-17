---
name: yunxiao-packages
version: 1.0.1
description: "云效制品仓库（Packages）：列仓库/制品、删除制品。用户问 Maven/NPM/通用制品库列表时使用。上传未封装（MCP 无清晰 GENERIC 上传路径）。"
metadata:
  requires:
    bins: ["yunxiao"]
  cliHelp: "yunxiao packages --help"
---

# packages

开始前先读 [`../yunxiao-shared/SKILL.md`](../yunxiao-shared/SKILL.md)。

> List 响应 `meta` 可能含 `has_more` / `total` / `page`；`has_more==true` 时勿把首页当全集（见 yunxiao-shared）。

## Typed commands

```bash
yunxiao packages repos list
yunxiao packages repos list --repo-types MAVEN --page 1
yunxiao packages artifacts list --repo-id <id> --repo-type GENERIC
yunxiao packages artifacts get --repo-id <id> --repo-type GENERIC --id <aid>
```

```bash
yunxiao packages artifacts delete --repo-id <id> --repo-type GENERIC --id <aid> --dry-run
yunxiao packages artifacts delete --repo-id <id> --repo-type MAVEN --id <aid> --version-id <vid> --yes
```

`artifacts delete` 为 **high-risk-write**。Artifact upload 尚未封装（API 形态因仓库类型而异）→ 用 `yunxiao api` 或后续版本。

> **Upload**：v0.6 跳过——MCP/operations 中无清晰 GENERIC 上传路径；已知端点可用 `yunxiao api`。

## Upload（跳过）

已核对 `operations/packages/{artifacts,repositories,types}.ts` 与 `tool-registry/packages.ts`：仅有 list repositories / list+get artifacts，**无 upload OpenAPI**。公开帮助文档为流水线「构建物上传」步骤，非 OpenAPI。故 CLI 不封装 upload。

## Known gaps

Upload / repo create-delete 仍跳过（MCP 无清晰路径）。
