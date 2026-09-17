#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "..", "bin", "yunxiao" + ext);

// Intercept "install" → setup wizard (binary may not exist yet under npx).
const args = process.argv.slice(2);
if (args[0] === "install") {
  require("./install-wizard.js");
} else {
  if (!fs.existsSync(bin)) {
    try {
      execFileSync(process.execPath, [path.join(__dirname, "install.js")], {
        stdio: "inherit",
        env: { ...process.env, YUNXIAO_CLI_RUN: "1" },
      });
    } catch (_) {
      console.error(
        `\nFailed to auto-install yunxiao binary.\n` +
          `Fix by running:\n` +
          `  node "${path.join(__dirname, "install.js")}"\n`
      );
      process.exit(1);
    }
  }

  try {
    execFileSync(bin, args, { stdio: "inherit" });
  } catch (e) {
    process.exit(e.status || 1);
  }
}
