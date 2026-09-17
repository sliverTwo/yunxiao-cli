---
name: yunxiao-appstack
version: 1.0.1
description: "云效 AppStack：应用、变更单、编排、标签、变量组。用户问应用交付/变更单/编排/标签/变量时使用。"
metadata:
  requires:
    bins: ["yunxiao"]
  cliHelp: "yunxiao appstack --help"
---

# appstack

开始前先读 [`../yunxiao-shared/SKILL.md`](../yunxiao-shared/SKILL.md)。

> List 响应 `meta` 可能含 `has_more` / `total` / `page`；`has_more==true` 时勿把首页当全集（见 yunxiao-shared）。

```bash
yunxiao appstack apps list
yunxiao appstack apps get --name my-app
yunxiao appstack change-orders versions --app my-app
yunxiao appstack change-orders get --app my-app --sn <sn>
yunxiao appstack change-orders job-logs --app my-app --sn <sn> --job-sn <jsn>
yunxiao appstack orchestrations list --app my-app
yunxiao appstack tags search --search demo
yunxiao appstack variable-groups list --app my-app
yunxiao appstack variable-groups revision --app my-app
```

```bash
yunxiao appstack change-orders create --app my-app --data '{"changeOrderName":"d1","type":"Deploy","envs":{"prod":{}}}' --dry-run
yunxiao appstack tags create --name t --color "#4676e5" --dry-run
yunxiao appstack tags bind --app my-app --tag-names t --dry-run
yunxiao appstack variable-groups create --app my-app --name g --from-revision-sha <sha> --vars '[{"key":"K","value":"V"}]' --dry-run
```

`create` / `execute-job` / tags*mutations / variable-groups create|update|delete 为 **high-risk-write**。
来源：`operations/appstack/{appTags,variableGroups,changeOrders,appOrchestrations}.ts`。

## Change requests / apps / global-vars

```bash
yunxiao appstack apps create --name demo --dry-run
yunxiao appstack apps sources --name my-app
yunxiao appstack change-requests list --app my-app
yunxiao appstack global-vars list
```

## Release workflows / deploy hosts (v0.9)

```bash
yunxiao appstack release-workflows list --app my-app
yunxiao appstack release-workflows stage get --app a --workflow-sn w --stage-sn s
yunxiao appstack release-workflows stage execute --app a --workflow-sn w --stage-sn s --dry-run
yunxiao appstack deploy machine-log --tunnel-id 1 --machine-sn sn
yunxiao appstack deploy add-hosts --instance n --host-sns a,b --dry-run
```

Stage execute/cancel/retry/skip/pass/refuse 与 host list mutations 为 **high-risk-write**。
来源：`operations/appstack/releaseWorkflows.ts`、`deploymentResources.ts`。
