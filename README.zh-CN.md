语言：中文 | [English](README.md)

# yunxiao-cli

云效 CLI，对标飞书 / Lark CLI：渐进发现、`+shortcuts`、类型化 API 命令、原始 `api` 逃生舱、风险门禁与 Agent skills。

CLI 二进制名：**`yunxiao`**。

## 同事试用（5 分钟）

### 1. 安装

```bash
# 配置公司阿里云 npm 私仓（若本机还没有）
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/

npm install -g sanzhi-yunxiao-cli@0.16.1
npx sanzhi-yunxiao-cli@latest install   # 拉二进制 + 可选安装 skills

yunxiao --version   # 应显示 0.16.1

# 备选：直接下 Release，解压后把 yunxiao.exe 所在目录加入 PATH
# https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.16.1
```

### 2. 认证

打开 [个人访问令牌控制台](https://account-devops.aliyun.com/settings/personalAccessToken) 新建 PAT（名称建议 `yunxiao-cli`；勾选组织读 + 项目/代码/流水线读写，按需制品/测试/应用；令牌只显示一次），然后：

```bash
yunxiao auth login --token "<PAT>"
yunxiao whoami && yunxiao doctor
```

### 3. 试用几条只读命令

```bash
yunxiao organization +whoami
yunxiao pipeline list
yunxiao codeup repos list
```

### 4. 注意

- 写操作先 `--dry-run`；高风险要确认后再加 `--yes`
- 长 JSON 用 `--data-file ./body.json`
- Skills 向导可多选；也可 `yunxiao skills install --skill ...`


### 给 Agent 粘贴

把下面整段发给 Agent（会装 CLI、认证、装 skills、只读列项目、让你选项目、**只在本机**写 profile，不会往仓库塞智衣/沙箱租户配置）：

```text
请帮我安装并初始化 yunxiao CLI（本地 profile，勿写入本仓库）：

1) 安装（公司阿里云 npm 私仓）：
   npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
   npm install -g sanzhi-yunxiao-cli@0.16.1
   npx sanzhi-yunxiao-cli@latest install
   yunxiao --version   # 应显示 0.16.1
   备选：GitHub Release v0.16.1
   https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.16.1

2) 认证（优先浏览器 OAuth；无图形界面再用 PAT。禁止把完整 token 打到回复/聊天里）
   推荐：yunxiao auth login --browser
   注意：OAuth 同意 = 账号 API 全能力（平台不按模块限权，宽于细粒度 PAT）。
   登录后探测：yunxiao auth probe-oauth
   PAT 回落（CI/无浏览器）：
     控制台：https://account-devops.aliyun.com/settings/personalAccessToken
     帮助：https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token
     yunxiao auth login --token "<PAT>"
   yunxiao whoami && yunxiao doctor && yunxiao auth status

3) 安装 companion skills：
   yunxiao skills install

4) 只读列项目，请用户挑选一个 space_id / 项目：
   yunxiao project list

5) 仅为所选项目初始化 **本地** profile（写入 ~/.config/yunxiao/profiles/，不要写入 git 工作区）：
   推荐交互：yunxiao +onboard
   或非交互：yunxiao +onboard --space-id <id> --profile <name>
   不要默认走 `profile install-example zhiyi|play`；智衣/沙箱字段留在用户本机/私有工具，不要提交到本公开仓库。
   若仓库里已有 profiles/*.example.json，仅作示例；真实租户 profile 必须本地、勿跟踪。

6) 校验：
   export YUNXIAO_PROFILE=<name>
   yunxiao profile show
   yunxiao profile doctor
   yunxiao doctor

7) 风险规则：写操作先 --dry-run；高风险需用户确认后再加 --yes；长 JSON 用 --data-file。
```

## 安装

**推荐（公司阿里云 npm 私仓 + npm 安装器）：**

```bash
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
npm install -g sanzhi-yunxiao-cli@0.16.1
npx sanzhi-yunxiao-cli@latest install
yunxiao --version          # yunxiao 0.16.1
```

`sanzhi-yunxiao-cli` 包会在 `postinstall` 时解压平台归档、安装 companion skills，并打印认证后续步骤。二进制来自 [GitHub Releases](https://github.com/sliverTwo/yunxiao-cli/releases)；默认值为 `YUNXIAO_CLI_GITHUB_REPO=sliverTwo/yunxiao-cli`（需要时可覆盖）。也可以[直接下载 v0.16.1 Release](https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.16.1)，解压后把 `yunxiao` 加入 `PATH`。

**从源码安装（次要）：**

```bash
make build          # 生成 ./yunxiao（-ldflags 注入 Version）
make install        # 安装到 ~/.local/bin/yunxiao
go build -o yunxiao .   # 无 ldflags 时回退包内默认 0.16.1
# 显式注入：
# go build -ldflags "-X github.com/yunxiao-cli/yunxiao/internal/version.Version=0.16.1" -o yunxiao .
```

需要 Go 1.24.4+。`make build` / `make ci` 通过 `-ldflags -X …version.Version=$(VERSION)` 注入版本（`VERSION` 默认 `git describe` 或 `0.16.1`）。

**已知限制：** `go install` / 单独二进制**不包含**仓库 `skills/` 目录；请用 npm 安装器（会解压 `skills/`）、在源码检出目录运行，或另行复制 / `npx skills add`。需要技能时优先 `npx sanzhi-yunxiao-cli@latest install` 或检出目录 `make build`。

## 认证

1. 在云效控制台创建个人访问令牌（PAT，首选一键链接）：  
   https://account-devops.aliyun.com/settings/personalAccessToken  
   帮助文档：https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token

   推荐勾选：组织/成员**读**；项目管理/代码管理/流水线**读+写**（只读试用可只开读）；制品/测试/应用交付按需。令牌名建议 `yunxiao-cli`。
2. 推荐环境变量：

```bash
export YUNXIAO_ACCESS_TOKEN="<PAT>"
export YUNXIAO_ORGANIZATION_ID="<企业ID>"   # 可选
```

或写入配置文件：

```bash
yunxiao auth login --token "<PAT>"
yunxiao auth status
yunxiao doctor
```

令牌优先级（高→低）：`YUNXIAO_ACCESS_TOKEN` → `~/.config/yunxiao/credentials.json`（最近一次成功的 `auth login`，browser 或 token）→ 当前 profile 的 `access_token` → 旧版 `config.json`。OAuth 凭证只写 `credentials.json`（0600），不写 profile JSON。`yunxiao auth status` 含 `token_source` / `token_kind`（`pat`|`oauth`），不打印明文。推荐：`yunxiao auth login --browser`（授权=账号 API 全能力）；CI 仍用 `--token`/env。默认 API：`https://openapi-rdc.aliyuncs.com`；OAuth 探测通过后按记录的头发送（优先 `x-yunxiao-token`，否则 `Authorization: Bearer`）。

## Agent 快速上手

```text
浏览：  yunxiao <domain> --help
查看：  yunxiao schema <id>
优先：  +快捷命令 → 类型化命令 → yunxiao api
风险：  read | write | high-risk-write（高风险需用户确认后再加 --yes）
预览：  --dry-run    过滤：--jq '...'
```

## Agent 技能

仓库 `skills/yunxiao-*` 下技能（均含 `SKILL.md`）：

| 技能 | 用途 |
|------|------|
| `yunxiao-shared` | 认证、配置、doctor、JSON 约定、`--dry-run` / `--yes` |
| `yunxiao-organization` | 组织、成员、部门、角色 |
| `yunxiao-project` | Projex 项目与工作项 |
| `yunxiao-codeup` | 代码库、分支、文件、MR |
| `yunxiao-pipeline` | Flow 流水线、运行、任务、YAML |
| `yunxiao-packages` | 制品仓库与制品 |
| `yunxiao-testhub` | 测试计划、结果、计划用例评论 |
| `yunxiao-appstack` | 应用、变更单、编排、标签、变量组 |
| `yunxiao-zhiyi-ops` | 智衣/ZYPT 迭代建议、开缺陷、流转、建 MR 与租户 profile（可选） |

**安装**（默认目录 `~/.agents/skills`，供 AI 工具发现）：

```bash
# 1) 推荐 — 本地 CLI 安装
yunxiao skills install
yunxiao skills install --skill yunxiao-shared --skill yunxiao-codeup
yunxiao skills install --dir /custom/skills --dry-run
yunxiao skills install --symlink --force

# 2) 从本地仓库路径
npx skills add /path/to/yunxiao-cli -y -g

# 3) Codeup 推送后（URL 必须以 .git 结尾；需 Codeup git 凭证）
npx skills add https://codeup.aliyun.com/sanzhi/cli/yunxiao_cli.git -y -g
```

安装后请重启 / 重载 AI 工具。查看：`yunxiao skills list|path|read <name>`。

贡献者与 AI Agent 请先读 **[AGENTS.md](AGENTS.md)**。

## 分域示例

长 JSON 请求体请用 `--data-file path.json` 或 `--data @path.json`（避免 shell 引号长度限制）。


```bash
yunxiao organization +whoami
yunxiao project +my-open-items
yunxiao project +created-by-me
yunxiao codeup +open-mrs
yunxiao codeup files tree --repo <id> --ref master
yunxiao codeup mrs create ... --dry-run    # 高风险：确认后再 --yes
yunxiao pipeline +status --pipeline-id <id>
yunxiao pipeline +failed --pipeline-id <id>
yunxiao packages repos list
yunxiao testhub plans list
yunxiao appstack apps list
yunxiao pipeline run trigger ... --dry-run
yunxiao pipeline run cancel ... --dry-run
yunxiao workitem create ... --dry-run
yunxiao codeup mrs merge ... --dry-run
yunxiao workitem delete ... --dry-run
yunxiao packages artifacts delete ... --dry-run
```


## Profile：play vs zhiyi（可选）

租户级 Projex 常量放在 **profile JSON**，不写进 CLI 全局默认。按项目（`space_id`）隔离；`workflows` 按 **`type_id`** 存放已探索状态图。
`workitem_defaults` 同样按 **`type_id`** 存放 OpenAPI 字段默认值与创建必填；`workitem create` / `+bug-create` 会自动填入（可用 `--no-defaults` 跳过）；`profile doctor` 会列出并校验这些字段 id。

| Profile | 用途 |
|---------|------|
| **zhiyi** | 智衣/ZYPT 全字段（module/environment/ExpCompletionTime + 完整流转必填） |
| **play** | 沙箱/YXCLI 回归 — 精简 `bug_create_fields`（仅 priority + seriousLevel）；`bug_transition_required` 仅 `{"100010":["80"]}` |

```bash
yunxiao profile install-example zhiyi   # 或 play
export YUNXIAO_PROFILE=zhiyi
yunxiao profile doctor
yunxiao workitem +bug-create --profile play --title "标题" --description "描述" --sprint <id> --dry-run
yunxiao workitem +bug-create --minimal --title "…" --description "…" --sprint <id> --dry-run
yunxiao workitem relations create --id <id> --related-id <rid> --relation-type ASSOCIATED --dry-run
```

详见 skill `yunxiao-zhiyi-ops`、`profiles/zhiyi.example.json`、`profiles/play.example.json`。

## 风险门禁

- `write`：先确认意图，尽量 `--dry-run`
- `high-risk-write`：缺少 `--yes` 时退出码 **10**，stderr 含 `confirmation_required`；**必须**向用户确认后再重试，禁止静默加 `--yes`

## 开发

```bash
make test && make build
make ci                 # go build -ldflags … ./... && go test ./... && go vet ./...
./scripts/ci.sh         # 同上，供 Codeup Flow 调用
```

**云效 Flow（可选）：** 如果把本仓库镜像到阿里云 Codeup，可在 Flow 中新增构建任务，脚本写：

```bash
make ci
# 或: ./scripts/ci.sh
```

GitHub 仓库已启用真实的 GitHub Actions：`.github/workflows/ci.yml` 在推送到 `main` 或提交 Pull Request 时运行 CI；`.github/workflows/release.yml` 在推送 `v*` 标签时构建各平台归档并发布 GitHub Release。

## 已知缺口

Packages **上传**、Codeup **blame/cherry-pick**、MR label detach 等仍无明确 OpenAPI；**Topic/Risk** 工作项类型需在项目设置 UI 启用（CLI 无法启用）。部分类型未启用**迭代**时请省略 `--sprint`。关联类型可用 `ASSOCIATED`/`DEPEND_ON`（`RELATED`/`PARENT_SUB` 常失败）。`profile doctor` 可对照线上字段/工作流。Codeup tags / protected-branches 已支持；`--content-file` 支持绝对路径。详见 [README.md](README.md) 的 Known gaps。

## 变更摘要

- **0.16.1** — 冒烟修复：`appstack apps list` 补齐必填 `pagination=keyset`；`workitem search` / `project +my-open-items` 回退 profile `space_id` 或给出清晰 CLI 错误；`programs search` 非高级版组织返回更友好提示
- **0.16.0** — 浏览器 OAuth（`auth login --browser` / `--dry-run`）、`credentials.json`（0600）、`auth probe-oauth`、oauth 自动 refresh；Agent 粘贴优先 browser；CI 保留 `--token`
- **0.15.7** — `yunxiao +onboard`：按所选项目/`space_id` 仅写入本机 `~/.config/yunxiao/profiles/` 的通用 profile（TTY 选择或 `--space-id`）；README「给 Agent 粘贴」；缺 token 时提示 PAT 控制台链接与模块权限清单；智衣/沙箱租户配置留在本机、勿提交本仓库（不借 onboard 扩展示例 profile）
- **0.15.6** — 评论/活动/历史类列表默认最新在前（`--sort asc|desc`，非法值报错）；评论按**创建时间**排序；活动/MR/流水线运行/工时等仍偏好更新时间；分页列表的客户端 `--sort` 仅作用于**当前页**（`--all` 时对已拉取页整体排序）
- **0.15.5** — 长 JSON 支持 `--data-file` / `--data @file.json`（`api`、appstack、testhub 等）
- **0.15.2** — companion skills 对齐 CLI 0.15.x（`has_more` / `meta.url` / `refresh_ok`）；`client.ListAll` + `pipeline list --all` / `codeup mrs list --all`；`scripts/flow-ci.sh`（阿里云 golang 镜像 + `GOPROXY=goproxy.cn`）
- **0.15.1** — B5 wave2：更多命令迁到 `runRead`/`runJSONMutating`（workitem update/relations list；codeup 写；pipeline 变更+剩余读；org/project/sprint/versions/packages/testhub/appstack/effort/programs 读与简单写）。仍自定义：multipart 附件、cancel-reason soft-warn dry-run、pipeline create/update YAML 预览脱敏、多步快捷命令
- **0.15.0** — 结构重构：B5 `runRead`/`runJSONMutating` 命令模板（部分迁移）；C1 拆分 `workitem.go`；C2 预编译日期正则；C3 `Do` 返回 headers；C4 ldflags 注入 Version
- **0.14.11** — pipeline/run 输出附带 Flow 控制台 `url`（`meta.url`；列表项注入）`https://flow.aliyun.com/pipelines/{id}` 与 `.../builds/{runId}`
- **0.14.10** — A1：文档记录接受 git 历史残留（≤v0.14.5 示例 ID）；D1：Retry-After 睡眠上限 30s；C5：清理 README 英文重复示例；标注 `go install` 与 skills 探测限制
- **0.14.9** — B1：HTTP 客户端 `context.Context` + GET/HEAD 重试（429/5xx/网络错误，尊重 Retry-After）；P2：`has_more` 结合 total/page/per_page；更多 list 接入 MetaWithPagination
- **0.14.8** — cobra Execute→exit 10 E2E；`refreshAfterTransition` + warning/`refresh_ok` 单测；列表 `meta.has_more`/`total`/`page`（MR list 接入 MetaWithPagination）
- **0.14.7** — help/skills 真实 ID 占位化；B3 Write/门禁契约测 + PostMultipart httptest；流转成功 JSON 增加 `refresh_ok`
- **0.14.6** — 示例 profile 脱敏为占位符；`go mod tidy`；`make ci` / `scripts/ci.sh`；流转后刷新失败打 stderr warning
- **0.14.4** — workitem/MR 输出附带可点击 `url`（`meta.url`；列表项注入 `url`）；URL 构建集中在 `internal/zhiyi`
- **0.14.3** — profile 可选 `access_token`；优先级 env > profile > config；`auth status` / doctor 报告 `token_source`
- **0.14.2** — `workitem create` / `+bug-create` 自动应用 `workitem_defaults`（priority/trackers/测试负责人/验收负责人），可用 `--no-defaults` 跳过
- **0.14.1** — profile `workitem_defaults`（按 `type_id` 存字段默认值与创建必填）；`profile doctor` 报告/校验
- **0.14.0** — 沙箱准确 `play` profile；`+bug-create --minimal`；`profile doctor`；关联类型文档；`--content-file` 绝对路径
- **0.13.1** — Codeup `--repo` 别名解析覆盖 branches/files/commits/compare/mrs/repos
- **0.13.0** — `workitem +transition`；Codeup `tags` / `protected-branches`；Topic/Risk 需项目 UI 启用
- **0.12.1** — profile 按 `type_id` 存 `workflows`；`--write-profile` 写入该映射（Bug 仍保留 `bug_*`）
- **0.12.0** — `workitem +explore-workflow` 探测状态流转图；`--write-profile` 写回 profile
- **0.11.0** — 智衣 `sprint +current`、`workitem +bug-create`、`codeup mrs +create`；profile 仓库/创建字段
- **0.10.0** — 智衣 profile；ZYPT workitem get；`workitem +bug-transition`；skill `yunxiao-zhiyi-ops`
- **0.9.1** — `yunxiao skills install`；AGENTS.md；README 技能安装说明
- **0.9.0** — AppStack 发布流/部署主机；Flow VM 部署单与资源成员写；工时/项目集；Codeup 建库

## 许可证

MIT — 见 [LICENSE](LICENSE)。
