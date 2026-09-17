# sanzhi-yunxiao-cli

Yunxiao (云效) CLI — Feishu-style one-click install. Native binaries are fetched from **GitHub Releases** (not bundled in the npm tarball).

## Quick start

```bash
npm install -g sanzhi-yunxiao-cli
npx sanzhi-yunxiao-cli@latest install
```

```bash
yunxiao --version
# interactive skill picker (TTY) or install all (non-TTY)
```

### Aliyun npm registry (optional)

```bash
npm install -g sanzhi-yunxiao-cli --registry=https://registry.npmmirror.com
```

Until published:

```bash
npm install -g ./npm
# or from a packed tarball (slim — no releases/)
npm pack
npm install -g ./sanzhi-yunxiao-cli-0.15.4.tgz
npx --yes ./sanzhi-yunxiao-cli-0.15.4.tgz install
```

## Auth (after install)

```bash
yunxiao auth login --token <PAT>
# or
export YUNXIAO_ACCESS_TOKEN=<PAT>

# optional tenant profile
yunxiao profile install-example zhiyi

yunxiao auth status
yunxiao doctor
```

## Environment

| Variable | Meaning |
|----------|---------|
| `YUNXIAO_CLI_GITHUB_REPO` | GitHub `owner/repo` for Releases (default: `sliverTwo/yunxiao-cli`) |
| `YUNXIAO_CLI_DOWNLOAD_BASE` | `https://` base URL hosting archives (tried before GitHub) |
| `YUNXIAO_CLI_SKILLS` | Non-TTY skills: `all` (default), `none`, or comma-separated skill names |
| `YUNXIAO_ACCESS_TOKEN` | PAT for API calls |
| `YUNXIAO_PROFILE` | Active tenant profile name |

### Download order (`scripts/install.js`)

1. Bundled `releases/<archive>` (local/dev offline)
2. `YUNXIAO_CLI_DOWNLOAD_BASE/<archive>`
3. `https://github.com/${REPO}/releases/download/v${VERSION}/<archive>`
4. China proxy mirrors (`ghproxy.net` / `mirror.ghproxy.com` prefix)

Asset names: `yunxiao-cli-${VERSION}-${platform}-${arch}.tar.gz` (Windows: `.zip`).

### Skills picker (`npx … install`)

On a TTY, after the binary is ready:

```
可选 AI Skills（安装到 ~/.agents/skills）
  1. yunxiao-appstack — …
  2. yunxiao-codeup — …
  …
选择 Skills：all/a 或回车=全部；n=跳过；或输入序号/名称（逗号分隔）
```

Example: `1,3,yunxiao-pipeline` then confirm → `yunxiao skills install --force --skill …`.

## Layout

```
npm/
  package.json               # bin → scripts/run.js ; postinstall → scripts/install.js
  scripts/run.js             # yunxiao …  ("install" → wizard)
  scripts/install.js         # download/extract native binary + skills/
  scripts/install-wizard.js  # one-click setup + skill multi-select
  checksums.txt              # SHA-256 of release archives
  README.md
  # releases/ is gitignored — used only for local/dev offline installs
```

Requires Node.js ≥ 18. OS: darwin / linux / win32; CPU: x64 / arm64 (Windows: amd64).
