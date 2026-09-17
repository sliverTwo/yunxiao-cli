---
name: yunxiao-organization
version: 1.0.1
description: "云效组织：当前用户、组织列表、成员、部门、角色。"
metadata:
  requires:
    bins: ["yunxiao"]
  cliHelp: "yunxiao organization --help"
---

# organization

开始前先读 [`../yunxiao-shared/SKILL.md`](../yunxiao-shared/SKILL.md)。

> List 响应 `meta` 可能含 `has_more` / `total` / `page`；`has_more==true` 时勿把首页当全集（见 yunxiao-shared）。

```bash
yunxiao organization +whoami
yunxiao organization list
yunxiao organization members list
yunxiao organization members search --query alice
yunxiao organization departments list
yunxiao organization departments get --id <id>
yunxiao organization roles list
yunxiao organization roles get --id <id>
```

Risk: **read**。
