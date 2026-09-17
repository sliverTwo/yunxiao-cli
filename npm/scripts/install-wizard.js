#!/usr/bin/env node
"use strict";

/**
 * Feishu-like one-click setup wizard for yunxiao-cli.
 * Steps: ensure binary → selectable AI skills → print auth next-steps.
 * Non-TTY: install all skills (or honor YUNXIAO_CLI_SKILLS=none|all|comma-list).
 */

const fs = require("fs");
const path = require("path");
const readline = require("readline");
const { execFileSync } = require("child_process");

const isWindows = process.platform === "win32";
const pkgRoot = path.join(__dirname, "..");
const binPath = path.join(pkgRoot, "bin", "yunxiao" + (isWindows ? ".exe" : ""));
const skillsRoot = path.join(pkgRoot, "skills");
const installJs = path.join(__dirname, "install.js");
const pkg = require("../package.json");
const VERSION = String(pkg.version).replace(/-.*$/, "");

function detectLocale() {
  const lang = (
    process.env.LC_ALL ||
    process.env.LANG ||
    process.env.LC_MESSAGES ||
    ""
  ).toLowerCase();
  return lang.startsWith("zh") ? "zh" : "en";
}

const messages = {
  zh: {
    setup: "正在设置云效 CLI (yunxiao-cli)...",
    step1: "安装原生二进制",
    step1Done: "二进制已就绪 (%s)",
    step1Fail: "二进制安装失败。请手动运行: node scripts/install.js",
    step2Header: "可选 AI Skills（安装到 ~/.agents/skills）",
    step2None: "未发现可安装的 Skills（归档中无 skills/yunxiao-*）",
    step2Prompt:
      "选择 Skills：all/a 或回车=全部；n=跳过；或输入序号/名称（逗号分隔）",
    step2Confirm: "将安装: %s  确认？(Y/n)",
    step2Skip: "已跳过 Skills 安装",
    step2Spinner: "正在安装 Skills...",
    step2Done: "Skills 已安装",
    step2Fail: "Skills 安装失败。可稍后运行: yunxiao skills install --force --skill <name>",
    step2Empty: "未选择任何 Skill",
    step3: "下一步（认证）",
    step3Body:
      "  yunxiao auth login --token <PAT>\n" +
      "  # 或: export YUNXIAO_ACCESS_TOKEN=<PAT>\n" +
      "  # 可选: yunxiao profile install-example zhiyi\n" +
      "  yunxiao auth status && yunxiao doctor",
    done: "安装完成！可对 AI 工具说：用 yunxiao 帮我看流水线 / 建缺陷 / 提 MR。",
    nonTty: "非交互环境：已完成二进制与 Skills；请手动完成认证步骤。",
  },
  en: {
    setup: "Setting up Yunxiao CLI (yunxiao-cli)...",
    step1: "Install native binary",
    step1Done: "Binary ready (%s)",
    step1Fail: "Binary install failed. Run manually: node scripts/install.js",
    step2Header: "Selectable AI skills (install into ~/.agents/skills)",
    step2None: "No installable skills found (no skills/yunxiao-* in archive)",
    step2Prompt:
      "Select skills: all/a or Enter = all; n = skip; or comma-separated numbers/names",
    step2Confirm: "Will install: %s  Confirm? (Y/n)",
    step2Skip: "Skipped skills install",
    step2Spinner: "Installing skills...",
    step2Done: "Skills installed",
    step2Fail: "Skills install failed. Later: yunxiao skills install --force --skill <name>",
    step2Empty: "No skills selected",
    step3: "Next steps (auth)",
    step3Body:
      "  yunxiao auth login --token <PAT>\n" +
      "  # or: export YUNXIAO_ACCESS_TOKEN=<PAT>\n" +
      "  # optional: yunxiao profile install-example zhiyi\n" +
      "  yunxiao auth status && yunxiao doctor",
    done: "You are all set! Ask your AI tool: use yunxiao to list pipelines / create a bug / open an MR.",
    nonTty: "Non-TTY: binary + skills done; complete auth steps manually.",
  },
};

function fmt(template, ...values) {
  let i = 0;
  return template.replace(/%s/g, () => values[i++] ?? "");
}

function ensureBinary(msg) {
  process.stderr.write(`→ ${msg.step1}\n`);
  try {
    execFileSync(process.execPath, [installJs], {
      stdio: "inherit",
      env: { ...process.env, YUNXIAO_CLI_RUN: "1" },
    });
  } catch (_) {
    process.stderr.write(`✗ ${msg.step1Fail}\n`);
    process.exit(1);
  }
  if (!fs.existsSync(binPath)) {
    process.stderr.write(`✗ ${msg.step1Fail}\n`);
    process.exit(1);
  }
  let ver = VERSION;
  try {
    const out = execFileSync(binPath, ["--version"], {
      stdio: ["ignore", "pipe", "ignore"],
      encoding: "utf8",
      timeout: 15000,
    }).trim();
    const m = out.match(/(\d+\.\d+\.\d+[\w.-]*)/);
    if (m) ver = m[1];
  } catch (_) {}
  process.stderr.write(`✓ ${fmt(msg.step1Done, ver)}\n`);
}

function runYunxiao(args, opts = {}) {
  return execFileSync(binPath, args, {
    stdio: opts.silent ? ["ignore", "pipe", "pipe"] : "inherit",
    encoding: "utf8",
    env: process.env,
    ...opts,
  });
}

function parseFrontmatter(skillMdPath) {
  const result = { name: null, description: "" };
  try {
    const text = fs.readFileSync(skillMdPath, "utf8");
    if (!text.startsWith("---")) return result;
    const end = text.indexOf("\n---", 3);
    if (end === -1) return result;
    const fm = text.slice(3, end);
    for (const line of fm.split("\n")) {
      const m = line.match(/^(\w+):\s*(.*)$/);
      if (!m) continue;
      const key = m[1];
      let val = m[2].trim();
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      if (key === "name") result.name = val;
      if (key === "description") result.description = val;
    }
  } catch (_) {}
  return result;
}

function discoverSkills() {
  const skills = [];
  if (!fs.existsSync(skillsRoot)) return skills;
  let entries;
  try {
    entries = fs.readdirSync(skillsRoot, { withFileTypes: true });
  } catch (_) {
    return skills;
  }
  for (const ent of entries) {
    if (!ent.isDirectory()) continue;
    if (!ent.name.startsWith("yunxiao-")) continue;
    const skillMd = path.join(skillsRoot, ent.name, "SKILL.md");
    if (!fs.existsSync(skillMd)) continue;
    const fm = parseFrontmatter(skillMd);
    const name = fm.name || ent.name;
    let desc = fm.description || "";
    if (desc.length > 72) desc = desc.slice(0, 69) + "...";
    skills.push({ id: ent.name, name, description: desc });
  }
  skills.sort((a, b) => a.id.localeCompare(b.id));
  return skills;
}

function askLine(question) {
  return new Promise((resolve) => {
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
    rl.question(`${question} `, (answer) => {
      rl.close();
      resolve(String(answer || "").trim());
    });
  });
}

async function askYesNo(question, defaultYes) {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    return defaultYes;
  }
  const a = (await askLine(question)).toLowerCase();
  if (!a) return defaultYes;
  return !(a === "n" || a === "no" || a === "否");
}

function resolveEnvSkills(available) {
  const raw = (process.env.YUNXIAO_CLI_SKILLS || "").trim();
  if (!raw) return { mode: "all" };
  const lower = raw.toLowerCase();
  if (lower === "none" || lower === "skip" || lower === "n") {
    return { mode: "skip" };
  }
  if (lower === "all" || lower === "a") {
    return { mode: "all" };
  }
  const tokens = raw.split(/[,;\s]+/).filter(Boolean);
  const selected = [];
  const byId = new Map(available.map((s) => [s.id.toLowerCase(), s]));
  const byName = new Map(available.map((s) => [s.name.toLowerCase(), s]));
  for (const t of tokens) {
    const hit = byId.get(t.toLowerCase()) || byName.get(t.toLowerCase());
    if (hit && !selected.find((x) => x.id === hit.id)) selected.push(hit);
  }
  return { mode: "list", selected };
}

function parseSelection(answer, available) {
  const a = String(answer || "").trim().toLowerCase();
  if (!a || a === "all" || a === "a" || a === "y" || a === "yes") {
    return { mode: "all" };
  }
  if (a === "n" || a === "no" || a === "skip" || a === "否") {
    return { mode: "skip" };
  }
  const tokens = answer.split(/[,;\s]+/).filter(Boolean);
  const selected = [];
  const byId = new Map(available.map((s) => [s.id.toLowerCase(), s]));
  const byName = new Map(available.map((s) => [s.name.toLowerCase(), s]));
  for (const t of tokens) {
    if (/^\d+$/.test(t)) {
      const idx = parseInt(t, 10) - 1;
      if (idx >= 0 && idx < available.length) {
        const hit = available[idx];
        if (!selected.find((x) => x.id === hit.id)) selected.push(hit);
      }
      continue;
    }
    const hit = byId.get(t.toLowerCase()) || byName.get(t.toLowerCase());
    if (hit && !selected.find((x) => x.id === hit.id)) selected.push(hit);
  }
  return { mode: "list", selected };
}

function printSkillList(available, msg) {
  process.stderr.write(`\n${msg.step2Header}\n`);
  available.forEach((s, i) => {
    const desc = s.description ? ` — ${s.description}` : "";
    process.stderr.write(`  ${i + 1}. ${s.id}${desc}\n`);
  });
  process.stderr.write("\n");
}

async function pickSkillsInteractive(available, msg) {
  printSkillList(available, msg);
  const answer = await askLine(msg.step2Prompt);
  const parsed = parseSelection(answer, available);
  if (parsed.mode === "skip") return null;
  const selected = parsed.mode === "all" ? available.slice() : parsed.selected;
  if (!selected.length) {
    process.stderr.write(`- ${msg.step2Empty}\n`);
    return null;
  }
  const names = selected.map((s) => s.id).join(", ");
  const ok = await askYesNo(fmt(msg.step2Confirm, names), true);
  if (!ok) return null;
  return selected;
}

function installSelectedSkills(selected, msg) {
  process.stderr.write(`→ ${msg.step2Spinner}\n`);
  try {
    const args = ["skills", "install", "--force"];
    for (const s of selected) {
      args.push("--skill", s.id);
    }
    runYunxiao(args);
    process.stderr.write(`✓ ${msg.step2Done}\n`);
  } catch (_) {
    process.stderr.write(`✗ ${msg.step2Fail}\n`);
  }
}

async function installSkills(msg) {
  const available = discoverSkills();
  if (!available.length) {
    process.stderr.write(`- ${msg.step2None}\n`);
    return;
  }

  const interactive = process.stdin.isTTY && process.stdout.isTTY;
  let selected = null;

  if (!interactive) {
    const envSel = resolveEnvSkills(available);
    if (envSel.mode === "skip") {
      process.stderr.write(`- ${msg.step2Skip}\n`);
      return;
    }
    selected = envSel.mode === "list" ? envSel.selected : available.slice();
    if (!selected.length) {
      process.stderr.write(`- ${msg.step2Empty}\n`);
      return;
    }
  } else {
    selected = await pickSkillsInteractive(available, msg);
    if (!selected) {
      process.stderr.write(`- ${msg.step2Skip}\n`);
      return;
    }
  }

  installSelectedSkills(selected, msg);
}

function printAuthNextSteps(msg) {
  process.stderr.write(`\n${msg.step3}\n${msg.step3Body}\n`);
}

async function main() {
  const msg = messages[detectLocale()];
  const interactive = process.stdin.isTTY && process.stdout.isTTY;
  process.stderr.write(`${msg.setup}\n\n`);

  ensureBinary(msg);
  await installSkills(msg);
  printAuthNextSteps(msg);

  if (!interactive) {
    process.stderr.write(`\n${msg.nonTty}\n`);
  }
  process.stderr.write(`\n${msg.done}\n`);
}

main().catch((err) => {
  console.error(err && err.message ? err.message : err);
  process.exit(1);
});
