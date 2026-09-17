# Output contract & high-risk approval

## Success vs failure

- Success: stdout JSON with `ok: true`, exit 0. There is **no** top-level `code`/`msg` on success.
- Failure: stderr JSON with `ok: false` and `error`, non-zero exit.
- Agents must branch on `ok` or exit code — never on `code == 0`.

## List pagination (`meta.has_more` / `total` / `page` / `pagination`)

- When list responses include Yunxiao `x-*` pagination headers, CLI merges them into `meta`: top-level `has_more`, optional `total`/`page`, plus nested `pagination`.
- **`has_more == true`** → first page is truncated; keep paging or use `--all` where wired (`pipeline list`, `codeup mrs list`).
- **Absent `has_more`/pagination fields ≠ complete set** — only means this response had no pagination headers.
- Console links often appear as `meta.url` (single-object) or per-item `url` (lists: workitem relations, MRs, pipelines/runs).

## `refresh_ok`

Transition shortcuts may return `refresh_ok: false` with a stderr warning when post-PUT refresh GET fails. Treat the transition as succeeded (`ok: true`); do not retry the status change solely because refresh failed.

## Exit 10

`confirmation_required` is **not** a network/auth error. Stop, ask the user, then retry with `--yes` only after explicit consent.

## Dry-run

```bash
yunxiao codeup mrs create --repo 123 --source feat --target master --title "x" --dry-run
```

Returns `{ "ok": true, "dry_run": true, "risk": "high-risk-write", "request": { ... } }` with token redacted.
