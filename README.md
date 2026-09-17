# yunxiao-cli

Yunxiao (阿里云云效) CLI redesigned like Feishu/Lark CLI: progressive discovery, `+shortcuts`, typed API commands, raw `api` escape hatch, risk gates, and agent skills.

CLI binary name: **`yunxiao`**.

## 同事试用（5 分钟）

### 1. 安装

```bash
# 配置公司阿里云 npm 私仓（若本机还没有）
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/

npm install -g sanzhi-yunxiao-cli@0.15.7
npx sanzhi-yunxiao-cli@latest install   # 拉二进制 + 可选安装 skills

yunxiao --version   # 应显示 0.15.7

# 备选：直接下 Release，解压后把 yunxiao.exe 所在目录加入 PATH
# https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7
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
   npm install -g sanzhi-yunxiao-cli@0.15.7
   npx sanzhi-yunxiao-cli@latest install
   yunxiao --version   # 应显示 0.15.7
   备选：GitHub Release v0.15.7
   https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7

2) 认证：先打开控制台新建 PAT（一键），再登录（禁止把完整 PAT 打到回复/聊天里）
   控制台（首选）：https://account-devops.aliyun.com/settings/personalAccessToken
   帮助文档（次要）：https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token
   推荐权限（控制台按「模块」勾选读/写，不是飞书 OAuth scope）：
     - 组织/成员：读（whoami、organization）
     - 项目管理(Projex)：读+写（工作项；只读试用可只开读）
     - 代码管理(Codeup)：读+写（仓库/分支/文件/MR；只读试用可只开读）
     - 流水线(Flow)：读+写（list/run；只读试用可只开读）
     - 按需：制品 Packages、测试管理 Testhub、应用交付 AppStack
   令牌名称建议 yunxiao-cli；设合理到期；令牌只显示一次。
   yunxiao auth login --token "<PAT>"
   yunxiao whoami && yunxiao doctor

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

### Trial quickstart (5 minutes)

```bash
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
npm install -g sanzhi-yunxiao-cli@0.15.7
npx sanzhi-yunxiao-cli@latest install
yunxiao --version

yunxiao auth login --token "<PAT>"
yunxiao whoami && yunxiao doctor

yunxiao organization +whoami
yunxiao pipeline list
yunxiao codeup repos list
```

Use the [GitHub Release binaries](https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7) as an alternative. Start writes with `--dry-run`, confirm high-risk writes before adding `--yes`, and use `--data-file ./body.json` for long JSON. The skills wizard supports multiple selections; individual skills can also be installed with `yunxiao skills install --skill ...`.


### Paste for Agent

Copy-paste for an agent (install → auth → skills → list projects read-only → user picks → **local-only** profile init; do not commit 智衣/sandbox tenant data):

```text
Install and init yunxiao CLI with a LOCAL profile (never write tenant profiles into this repo):

1) Install (Aliyun npm registry):
   npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
   npm install -g sanzhi-yunxiao-cli@0.15.7
   npx sanzhi-yunxiao-cli@latest install
   yunxiao --version   # expect 0.15.7
   Fallback: GitHub Release v0.15.7
   https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7

2) Auth — open the PAT console first (never print/paste the raw PAT into chat):
   Console (primary): https://account-devops.aliyun.com/settings/personalAccessToken
   Help (secondary): https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token
   Recommended module checkboxes (not OAuth scopes):
     - Organization/members: read (whoami, organization)
     - Projex: read+write (work items; read-only trial: read only)
     - Codeup: read+write (repos/branches/files/MRs; trial: read only)
     - Flow: read+write (list/run; trial: read only)
     - As needed: Packages, Testhub, AppStack
   Token name tip: yunxiao-cli; set a sensible expiry; shown once only.
   yunxiao auth login --token "<PAT>"
   yunxiao whoami && yunxiao doctor

3) yunxiao skills install

4) Read-only: yunxiao project list — ask the user to pick a project/space_id

5) Init a LOCAL profile only (~/.config/yunxiao/profiles/), not the git tree:
   Prefer: yunxiao +onboard
   Or: yunxiao +onboard --space-id <id> --profile <name>
   Do NOT default to `profile install-example zhiyi|play` for colleagues.
   Zhiyi/sandbox specifics stay on the user's machine / private tooling.
   Repo profiles/*.example.json (if present) are examples only; real tenant profiles must stay local/untracked.

6) export YUNXIAO_PROFILE=<name> ; yunxiao profile show ; yunxiao profile doctor ; yunxiao doctor

7) Risk: --dry-run before writes; high-risk needs user confirm then --yes; long JSON via --data-file.
```

---

## English

### Install

**Recommended (company registry + npm installer):**

```bash
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
npm install -g sanzhi-yunxiao-cli@0.15.7
npx sanzhi-yunxiao-cli@latest install
yunxiao --version          # yunxiao 0.15.7
```

The `sanzhi-yunxiao-cli` package runs `postinstall` to unpack the platform archive, install companion skills, and print auth next steps. It downloads binaries from [GitHub Releases](https://github.com/sliverTwo/yunxiao-cli/releases); `YUNXIAO_CLI_GITHUB_REPO=sliverTwo/yunxiao-cli` is the default (override it when needed). You can also [download the v0.15.7 Release directly](https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7), extract it, and add the `yunxiao` binary to `PATH`.

**From source (secondary):**

```bash
make build                 # produces ./yunxiao (injects Version via -ldflags)
# or (without ldflags, Version falls back to package default 0.15.7)
go build -o yunxiao .
# pin version explicitly:
# go build -ldflags "-X github.com/yunxiao-cli/yunxiao/internal/version.Version=0.15.7" -o yunxiao .
make install               # installs to ~/.local/bin/yunxiao
# or
go install github.com/yunxiao-cli/yunxiao@latest   # when published
```

Requires Go 1.24.4+. `make build` / `make ci` set `-ldflags -X …version.Version=$(VERSION)` (`VERSION` defaults to `git describe` or `0.15.7`).

**Known limitation:** `go install` / a lone binary does **not** ship the repo `skills/` tree, so `yunxiao skills list|read|install` will not find skills unless you use the npm installer (which extracts `skills/`), run from a source checkout, or copy/`npx skills add` the tree. Prefer `npx sanzhi-yunxiao-cli@latest install` or `make build` from a checkout for skills-aware workflows.

### Auth

1. Create a Personal Access Token in the Yunxiao console (primary):  
   https://account-devops.aliyun.com/settings/personalAccessToken  
   Help: https://help.aliyun.com/zh/yunxiao/user-guide/personal-access-token

   Recommended module checkboxes for this CLI: Organization/members **read**; Projex/Codeup/Flow **read+write** (or read-only for trial); Packages/Testhub/AppStack as needed. Token name tip: `yunxiao-cli`.
2. Prefer env (CI / shells):

```bash
export YUNXIAO_ACCESS_TOKEN="<PAT>"
# optional
export YUNXIAO_ORGANIZATION_ID="<orgId>"
export YUNXIAO_API_BASE_URL="https://openapi-rdc.aliyuncs.com"   # default
export YUNXIAO_EDITION="central"   # or region
```

Or store in config (`~/.config/yunxiao/config.json`, mode 0600):

```bash
yunxiao auth login --token "<PAT>"
yunxiao auth status
yunxiao whoami
yunxiao doctor
```

Token precedence (highest first): `YUNXIAO_ACCESS_TOKEN` env → active profile `access_token` (`--profile` / `YUNXIAO_PROFILE`) → `~/.config/yunxiao/config.json`. Optional: put `"access_token"` in a profile JSON (mode 0600); do not commit real PATs. `yunxiao auth status` reports `token_source` as `env` | `profile` | `config` | `none` without printing the raw token.
**Future:** Feishu-style one-click OAuth/browser login could be added later if we register a Yunxiao OAuth app (`CreateOAuthToken` is still 内测中). Today the supported path is the PAT console above — this CLI does not implement a fake OAuth page.


### Agent quickstart

```text
Browse:     yunxiao <domain> --help
Inspect:    yunxiao schema <id>          # e.g. codeup.mrs.create
Prefer:     +shortcuts over typed over raw api
Risk:       read | write | high-risk-write
            high-risk-write needs --yes after user confirms
Preview:    --dry-run   Filter: --jq '...'
```

### Agent Skills

Companion skills live under `skills/yunxiao-*` (each has `SKILL.md`):

| Skill | Use for |
|-------|---------|
| `yunxiao-shared` | Auth, config, doctor, JSON contract, `--dry-run` / `--yes` |
| `yunxiao-organization` | Orgs, members, departments, roles |
| `yunxiao-project` | Projex projects & work items |
| `yunxiao-codeup` | Repos, branches, files, MRs |
| `yunxiao-pipeline` | Flow pipelines, runs, jobs, YAML |
| `yunxiao-packages` | Artifact repositories & artifacts |
| `yunxiao-testhub` | Test plans, results, plan comments |
| `yunxiao-appstack` | Apps, change-orders, orchestrations, tags, variable groups |
| `yunxiao-zhiyi-ops` | Zhiyi/ZYPT sprint/bug-create/transition/MR + tenant profile (optional) |

**Install** (so AI tools can discover them; default dir `~/.agents/skills`):

```bash
# 1) Recommended — local CLI install (copy into ~/.agents/skills)
yunxiao skills install
yunxiao skills install --skill yunxiao-shared --skill yunxiao-codeup
yunxiao skills install --dir /custom/skills --dry-run
yunxiao skills install --symlink --force

# 2) Via skills CLI from a local checkout
npx skills add /path/to/yunxiao-cli -y -g

# 3) After Codeup push (URL must end in .git; needs Codeup git credentials)
npx skills add https://codeup.aliyun.com/sanzhi/cli/yunxiao_cli.git -y -g
```

Then restart / reload your AI tool so skills are picked up.

Inspect without installing:

```bash
yunxiao skills list
yunxiao skills path
yunxiao skills read yunxiao-shared
```

Contributors and AI agents editing this repo: see **[AGENTS.md](AGENTS.md)**.

### Examples by domain

Long JSON bodies: prefer `--data-file path.json` or `--data @path.json` (avoids shell quoting limits).


```bash
# organization
yunxiao organization +whoami
yunxiao organization list
yunxiao organization members search --query alice

# project / work items
yunxiao project list --name demo
yunxiao project +my-open-items
yunxiao project +created-by-me --status-stage 1,2
yunxiao workitem search --assigned-to self --category Req --priority <id>
yunxiao workitem get --id <id>
yunxiao workitem comments list --id <id>
yunxiao workitem comment --id <id> --content "note" --dry-run
yunxiao workitem create --space-id <sid> --type-id <tid> --subject "title" --assigned-to self --dry-run
yunxiao workitem update --id <id> --assigned-to self --dry-run
yunxiao workitem +transition --id <id|serial> --to <alias|statusId> --dry-run

# codeup
yunxiao codeup repos list
yunxiao codeup branches list --repo <repoId>
yunxiao codeup tags list --repo <repoId>
yunxiao codeup tags create --repo <repoId> --tag-name v1.0 --ref master --dry-run
yunxiao codeup protected-branches list --repo <repoId>
yunxiao codeup protected-branches create --repo <repoId> --branch master --allow-push-roles 40,30 --dry-run
yunxiao codeup files tree --repo <repoId> --ref master
yunxiao codeup commits list --repo <repoId> --ref master
yunxiao codeup files create --repo <id> --path a.txt --branch master --message "add" --content "hi" --dry-run
yunxiao codeup files delete --repo <id> --path a.txt --branch master --message "rm" --dry-run
yunxiao codeup mrs merge --repo <id> --local-id 1 --merge-type no-fast-forward --dry-run
yunxiao codeup mrs close --repo <id> --local-id 1 --dry-run
yunxiao codeup mrs review --repo <id> --local-id 1 --opinion PASS --dry-run

yunxiao codeup mrs get --repo <id> --local-id 1
yunxiao codeup mrs diffs --repo <id> --local-id 1
yunxiao codeup mrs comments list --repo <id> --local-id 1
yunxiao codeup mrs comments create --repo <id> --local-id 1 --content "LGTM" --patchset-biz-id <biz> --dry-run
yunxiao codeup mrs labels list --repo <id> --local-id 1
yunxiao codeup mrs labels attach --repo <id> --local-id 1 --label-ids 1,2 --dry-run
yunxiao codeup mrs reopen --repo <id> --local-id 1 --dry-run
yunxiao codeup compare --repo <id> --from master --to feature
yunxiao pipeline job retry --pipeline-id <id> --run-id <r> --job-id <j> --dry-run
yunxiao pipeline job pass --pipeline-id <id> --run-id <r> --job-id <j> --dry-run
yunxiao pipeline job refuse --pipeline-id <id> --run-id <r> --job-id <j> --dry-run
yunxiao packages artifacts delete --repo-id <id> --repo-type GENERIC --id <aid> --dry-run
yunxiao workitem types list --space-id <sid> --category Req
yunxiao workitem create --space-id <sid> --type-id <tid> --subject "t" --assigned-to self --custom-fields '{"fid":"v"}' --dry-run
yunxiao workitem relations list --id <id> --relation-type ASSOCIATED
yunxiao workitem relations create --id <id> --related-id <rid> --relation-type ASSOCIATED --dry-run
yunxiao workitem delete --id <id> --dry-run
yunxiao testhub results update --plan-id <p> --testcase-id <t> --status PASSED --dry-run
yunxiao testhub plan-comments list --plan-id <p> --testcase-id <t>
yunxiao appstack change-orders job-logs --app my-app --sn <sn> --job-sn <jsn>
yunxiao appstack orchestrations list --app my-app
yunxiao appstack change-orders create --app my-app --data '{...}' --dry-run
yunxiao appstack change-orders create --app my-app --data-file order.json --dry-run
yunxiao codeup +open-mrs
yunxiao codeup mrs create --repo <id> --source feat --target master --title "x" --dry-run
yunxiao codeup mrs create --repo <id> --source feat --target master --title "x" --yes   # after user OK

# pipeline
yunxiao pipeline list
yunxiao pipeline +status --pipeline-id <id>
yunxiao pipeline run list --pipeline-id <id>
yunxiao pipeline run latest --pipeline-id <id>
yunxiao pipeline +failed --pipeline-id <id>
yunxiao pipeline job log --pipeline-id <id> --run-id <rid> --job-id <jid>
yunxiao pipeline run trigger --pipeline-id <id> --branch master --dry-run
yunxiao pipeline run cancel --pipeline-id <id> --run-id <rid> --dry-run

# packages (upload skipped — see Known gaps)
yunxiao packages repos list
yunxiao packages artifacts list --repo-id <id> --repo-type GENERIC

# testhub / appstack
yunxiao testhub plans list --project-id <id>
yunxiao testhub plans progress --plan-id <id>
yunxiao appstack apps list
yunxiao appstack change-orders versions --app my-app
yunxiao appstack change-orders job-logs --app my-app --sn <sn> --job-sn <jsn>
yunxiao appstack orchestrations list --app my-app


# v0.7
yunxiao pipeline get --id <id>
yunxiao pipeline create --name ci --file ./pipeline.yaml --dry-run
yunxiao pipeline update --id <id> --name ci --file ./pipeline.yaml --dry-run
yunxiao workitem attachments list --id <id>
yunxiao workitem attachments create --id <id> --file ./shot.png --dry-run
yunxiao appstack tags search --search demo
yunxiao appstack tags create --name t --color "#4676e5" --dry-run
yunxiao appstack tags bind --app my-app --tag-names t --dry-run
yunxiao appstack variable-groups list --app my-app
yunxiao appstack variable-groups revision --app my-app

# v0.8
yunxiao organization departments list
yunxiao organization roles list
yunxiao project get --id <id>
yunxiao sprint list --space-id <id>
yunxiao versions list --space-id <id>
yunxiao workitem fields --space-id <s> --type-id <t>
yunxiao pipeline service-connections list --type codeup
yunxiao pipeline host-groups list
yunxiao pipeline flow-variable-groups list
yunxiao codeup repos get --repo <id>
yunxiao codeup branches create --repo <id> --branch feat --ref master --dry-run
yunxiao appstack apps create --name demo --dry-run
yunxiao appstack change-requests list --app my-app
yunxiao appstack global-vars list
yunxiao testhub cases search --repo-id <id>
yunxiao testhub directories create --repo-id <id> --name folder --dry-run

# v0.9
yunxiao appstack release-workflows list --app my-app
yunxiao appstack release-workflows stage execute --app a --workflow-sn w --stage-sn s --dry-run
yunxiao appstack deploy machine-log --tunnel-id 1 --machine-sn sn
yunxiao appstack deploy add-hosts --instance n --host-sns a,b --dry-run
yunxiao pipeline vm-deploy get --pipeline-id p --deploy-id d
yunxiao pipeline vm-deploy stop --pipeline-id p --deploy-id d --dry-run
yunxiao pipeline resource-members create --resource-type pipeline --resource-id id --role-name viewer --user-id u --dry-run
yunxiao workitem efforts list --id <id>
yunxiao workitem efforts mine --start-date 2026-01-01 --end-date 2026-01-31
yunxiao workitem estimated-efforts create --id <id> --owner self --spent-time 4 --dry-run
yunxiao programs search --name demo
yunxiao codeup repos create --name my-repo --path my-repo --dry-run
# escape hatch
yunxiao api GET /oapi/v1/platform/user
yunxiao schema
```


### Profiles: play vs zhiyi (optional)
> Note: `profiles/*.example.json` in this repo (if present) are **examples only**. Real tenant profiles (智衣/沙箱/etc.) must live under `~/.config/yunxiao/profiles/` and stay local/untracked. Prefer `yunxiao +onboard` to create a generic local profile from a chosen `space_id`.


Tenant-specific Projex constants live in a **profile JSON**, not hardcoded CLI defaults.
Profiles are **project-scoped** (`space_id`); discovered workitem graphs live under `workflows` keyed by **`type_id`**.
`workitem_defaults` (keyed by **`type_id`**) stores OpenAPI field defaults + create-required ids for create payloads; `workitem create` and `+bug-create` apply those field defaults (priority/trackers/测试负责人/验收负责人, …) unless overridden by flags / `--custom-fields` or `--no-defaults`. `yunxiao profile doctor` reports which types have them and checks those field ids against live fields.

| Profile | Purpose |
|---------|---------|
| **zhiyi** | Full Zhiyi/ZYPT field set (`module` / `environment` / `ExpCompletionTime` + rich `bug_transition_required`) |
| **play** | Sandbox/YXCLI regression — minimal `bug_create_fields` (priority + seriousLevel only); `bug_transition_required` = `{"100010":["80"]}` only; sandbox bug statuses |

```bash
yunxiao profile install-example zhiyi   # or: play
export YUNXIAO_PROFILE=zhiyi            # or play
yunxiao profile show
yunxiao profile doctor                 # diff profile vs live fields/workflow (read)
yunxiao workitem get ZYPT-5768         # zhiyi serials; play uses YXCLI-…
yunxiao sprint +current --dry-run
# Zhiyi full create:
yunxiao workitem +bug-create --title "标题" --description "描述" \
  --expected-completion 2026-09-20 --sprint <id> --dry-run
# Sandbox / non-Zhiyi (omit module/env/ExpCompletionTime):
yunxiao workitem +bug-create --profile play --title "标题" --description "描述" \
  --sprint <id> --dry-run
yunxiao workitem +bug-create --minimal --title "…" --description "…" --sprint <id> --dry-run
yunxiao workitem +bug-transition --id ZYPT-5768 --to processing \
  --plan-due-date 2026-09-20 --developer <uid> --dry-run
yunxiao workitem +explore-workflow --type-id <bug_type_id> --cleanup --dry-run
yunxiao workitem relations create --id <id> --related-id <rid> --relation-type ASSOCIATED --dry-run
# Codeup --content-file accepts cwd-relative or absolute paths
yunxiao codeup files update --repo sandbox --path README.md --branch x \
  --message "…" --content-file /tmp/note.md --dry-run
yunxiao codeup mrs +create --repo iipmes_gy --source feat/x \
  --title "fix" --work-item ZYPT-5768 --wip --dry-run
```

See skill `yunxiao-zhiyi-ops`, `profiles/zhiyi.example.json`, and `profiles/play.example.json`.

### Risk / dry-run / --yes

| Level | Rule |
|-------|------|
| read | Safe to run |
| write | Confirm intent; use `--dry-run` when available |
| high-risk-write | Exit **10** + `confirmation_required` without `--yes`. Ask the user; only then append `--yes`. Never auto-confirm. |

### Build & test

```bash
make test
make build
./yunxiao --help
```


### Known gaps

Surfaces intentionally **not** wrapped (use `yunxiao api` when you have a confirmed OpenAPI path):

| Gap | Reason |
|-----|--------|
| Packages **upload** / repo create-delete | Not clear in MCP `operations/packages` / no OpenAPI for upload |
| Codeup **blame**, **cherry-pick** | No solid OpenAPI confirmed — do not invent |
| Projex **Topic / Risk** type enable on a project | Org may define types; project must enable them in **project settings UI**. Create returns `工作项类型未启用！`; no OpenAPI to enable — CLI cannot enable Topic/Risk |
| Topic / Risk **迭代** binding | Some types return `未启用此字段【迭代】` — omit `--sprint` (CLI surfaces a hint) |
| Relation types | Working: `ASSOCIATED`, `DEPEND_ON`. `RELATED` / `PARENT_SUB` often fail type constraints; Task parent via `--parent-id` on create |
| MR label **detach** | No OpenAPI in MCP |
| AppStack full CR lifecycle beyond list/create surfaces already shipped | Expand only when MCP is unambiguous |
| Flow structured pipeline YAML generator (`createPipelineWithOptions`) | MCP helper only; CLI takes raw YAML `--file` |

v0.9 landed deferred clears: AppStack release-workflows + deploy host mutations, Flow VM deploy orders, Projex efforts/programs, Flow resource-member writes, Codeup `repos create` (high-risk).

### Development

```bash
make test && make build
make ci                 # go build -ldflags … ./... && go test ./... && go vet ./...
./scripts/ci.sh         # same, POSIX; use from Codeup Flow
```

**Codeup Flow (optional):** If this repo is mirrored to Aliyun Codeup, add a Flow job whose build script is:

```bash
make ci
# or: ./scripts/ci.sh
```

GitHub Actions are active for the GitHub repository: `.github/workflows/ci.yml` runs CI on pushes to `main` and pull requests, while `.github/workflows/release.yml` builds platform archives and publishes a GitHub Release when a `v*` tag is pushed.

See [AGENTS.md](AGENTS.md) for contributor / AI-agent conventions.

### Changelog

- **0.15.7** — `yunxiao +onboard` writes a **generic local** profile under `~/.config/yunxiao/profiles/` (TTY project pick or `--space-id`); README Agent paste prompts (ZH+EN); missing-token hints include PAT console URL + module permission checklist; real 智衣/沙箱 tenant profiles stay local/untracked (do not expand example profiles for onboard)
- **0.15.6** — default newest-first for comment/activity/history-style lists (`--sort asc|desc`; invalid values rejected); comments sort by **create** time; activity/MR/runs/efforts prefer update/modified; client-side `--sort` is **page-local** when the list is paginated (`--all` sorts across collected pages)
- **0.15.5** — `--data-file` and `--data @file.json` for long JSON payloads (`api`, appstack, testhub, …)
- **0.15.2** — companion skills refresh for CLI 0.15.x (`has_more` / `meta.url` / `refresh_ok`); `client.ListAll` + `--all` on `pipeline list` & `codeup mrs list`; `scripts/flow-ci.sh` (Alibaba golang mirror + `GOPROXY=goproxy.cn`)
- **0.15.1** — B5 wave2: more cmds on `runRead`/`runJSONMutating` (workitem update/relations list; codeup writes; pipeline mutations + remaining reads; org/project/sprint/versions/packages/testhub/appstack/effort/programs reads + simple writes). Still custom: multipart attachments, cancel-reason soft-warn dry-run envelope, pipeline create/update YAML redaction preview, multi-step shortcuts (+transition/+bug*/+explore-workflow, MR create, testhub results fallback, sprint +bugs aggregate)
- **0.15.0** — structural: B5 `runRead`/`runJSONMutating` cmd helpers (partial migration); C1 split `workitem.go`; C2 precompiled date regex; C3 `Do` returns headers (lists use `Do`+`MetaWithPagination`); C4 ldflags Version injection
- **0.14.11** — pipeline/run responses include Flow console `url` (`meta.url`; list items) via `https://flow.aliyun.com/pipelines/{id}` and `.../builds/{runId}`
- **0.14.10** — A1: document accept of git-history residual (≤v0.14.5 example IDs); D1: Retry-After sleep capped at 30s; C5: README duplicate EN examples cleaned; note `go install` vs skills-tree discovery
- **0.14.9** — B1: HTTP client `context.Context` + GET/HEAD retry (429/5xx/network, Retry-After); P2: `has_more` via total/page/per_page; more lists use MetaWithPagination
- **0.14.8** — cobra Execute→exit 10 E2E; `refreshAfterTransition` + warning/`refresh_ok` unit tests; list `meta.has_more`/`total`/`page` via MetaWithPagination (MR list wired)
- **0.14.7** — help/skills sanitize real IDs to placeholders; B3 Write/gate contract tests + PostMultipart httptest; transition `refresh_ok` in success JSON
- **0.14.6** — sanitize example profiles (placeholders only); `go mod tidy`; `make ci` / `scripts/ci.sh`; transition refresh-fail stderr warning
- **0.14.4** — workitem/MR responses include clickable `url` (`meta.url`; list items get `url`); builders in `internal/zhiyi`
- **0.14.3** — optional per-profile `access_token`; token precedence env > profile > config; `auth status` / doctor report `token_source`
- **0.14.2** — `workitem create` / `+bug-create` apply profile `workitem_defaults` (priority/trackers/测试负责人/验收负责人) unless overridden or `--no-defaults`
- **0.14.1** — `workitem_defaults` in profiles (per-`type_id` field defaults + create_required); `profile doctor` reports/verifies them
- **0.14.0** — Sandbox-accurate `play` profile; `+bug-create --minimal` / omit disabled fields; `profile doctor`; relation-type docs (`ASSOCIATED`/`DEPEND_ON`); sprint/field-not-enabled hints; `--content-file` absolute paths
- **0.13.1** — Codeup `--repo` alias resolution for branches/files/commits/compare/mrs/repos (reuse `resolveCodeupRepo`)
- **0.13.0** — `workitem +transition` (any type via `workflows`); Codeup `tags` + `protected-branches`; Topic/Risk enable is UI-only
- **0.12.1** — profiles store per-`type_id` `workflows`; `--write-profile` fills that map (legacy `bug_*` kept for Bug)
- **0.12.0** — `workitem +explore-workflow` auto-discovers status transition graphs; profile `--write-profile` merge
- **0.11.0** — Zhiyi `sprint +current`, `workitem +bug-create`, `codeup mrs +create`; profile repos/create-fields
- **0.10.0** — Zhiyi tenant profile; ZYPT workitem get; `workitem +bug-transition`; skill `yunxiao-zhiyi-ops`
- **0.9.1** — `yunxiao skills install`; AGENTS.md; README skills install docs
- **0.9.0** — AppStack RW + deploy; Flow vm-deploy + resource-members write; efforts/programs; repos create
- **0.8.0** — org dept/roles; sprint/versions; Flow SC/HG/VG/RM list; codeup repos/branches; AppStack apps/CR/global-vars; testhub cases
- **0.7.0** — pipeline YAML get/create/update; AppStack tags + variable-groups; workitem attachments
- **0.6.0** — MR comments/labels; pipeline pass/refuse; testhub results; workitem relations; AppStack job-logs

### License

MIT — see [LICENSE](LICENSE).

---

## 中文

### 安装

**推荐（公司阿里云 npm 私仓 + npm 安装器）：**

```bash
npm config set registry https://packages.aliyun.com/67762490f72b227b2bf8327b/npm/npm-registry/
npm install -g sanzhi-yunxiao-cli@0.15.7
npx sanzhi-yunxiao-cli@latest install
yunxiao --version          # yunxiao 0.15.7
```

`sanzhi-yunxiao-cli` 包会在 `postinstall` 时解压平台归档、安装 companion skills，并打印认证后续步骤。二进制来自 [GitHub Releases](https://github.com/sliverTwo/yunxiao-cli/releases)；默认值为 `YUNXIAO_CLI_GITHUB_REPO=sliverTwo/yunxiao-cli`（需要时可覆盖）。也可以[直接下载 v0.15.7 Release](https://github.com/sliverTwo/yunxiao-cli/releases/tag/v0.15.7)，解压后把 `yunxiao` 加入 `PATH`。

**从源码安装（次要）：**

```bash
make build          # 生成 ./yunxiao（-ldflags 注入 Version）
make install        # 安装到 ~/.local/bin/yunxiao
go build -o yunxiao .   # 无 ldflags 时回退包内默认 0.15.7
# 显式注入：
# go build -ldflags "-X github.com/yunxiao-cli/yunxiao/internal/version.Version=0.15.7" -o yunxiao .
```

需要 Go 1.24.4+。`make build` / `make ci` 通过 `-ldflags -X …version.Version=$(VERSION)` 注入版本（`VERSION` 默认 `git describe` 或 `0.15.7`）。

**已知限制：** `go install` / 单独二进制**不包含**仓库 `skills/` 目录；请用 npm 安装器（会解压 `skills/`）、在源码检出目录运行，或另行复制 / `npx skills add`。需要技能时优先 `npx sanzhi-yunxiao-cli@latest install` 或检出目录 `make build`。

### 认证

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

令牌优先级（高→低）：`YUNXIAO_ACCESS_TOKEN` → 当前 profile 的 `access_token`（`--profile` / `YUNXIAO_PROFILE`）→ `~/.config/yunxiao/config.json`。可在 profile JSON 中加可选 `"access_token"`（建议文件权限 0600，勿提交真实 PAT）。`yunxiao auth status` 的 `token_source` 为 `env` | `profile` | `config` | `none`，不打印明文。默认 API：`https://openapi-rdc.aliyuncs.com`，请求头 `x-yunxiao-token`。

### Agent 快速上手

```text
浏览：  yunxiao <domain> --help
查看：  yunxiao schema <id>
优先：  +快捷命令 → 类型化命令 → yunxiao api
风险：  read | write | high-risk-write（高风险需用户确认后再加 --yes）
预览：  --dry-run    过滤：--jq '...'
```

### Agent 技能

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

### 分域示例

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


### Profile：play vs zhiyi（可选）

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

### 风险门禁

- `write`：先确认意图，尽量 `--dry-run`
- `high-risk-write`：缺少 `--yes` 时退出码 **10**，stderr 含 `confirmation_required`；**必须**向用户确认后再重试，禁止静默加 `--yes`

### 开发

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

### 已知缺口

Packages **上传**、Codeup **blame/cherry-pick**、MR label detach 等仍无明确 OpenAPI；**Topic/Risk** 工作项类型需在项目设置 UI 启用（CLI 无法启用）。部分类型未启用**迭代**时请省略 `--sprint`。关联类型可用 `ASSOCIATED`/`DEPEND_ON`（`RELATED`/`PARENT_SUB` 常失败）。`profile doctor` 可对照线上字段/工作流。Codeup tags / protected-branches 已支持；`--content-file` 支持绝对路径。详见英文 Known gaps。

### 变更摘要

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

### 许可证

MIT — 见 [LICENSE](LICENSE)。
