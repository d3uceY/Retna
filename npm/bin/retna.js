#!/usr/bin/env node
"use strict";

// The npm entry point. npm cannot publish a native executable, so the binary
// is downloaded beside this shim by install.js and spawned here. If the
// download did not happen (offline install, skipped postinstall, npm's
// install-script gate), it is retried once before giving up.

const fs = require("node:fs");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { resolve } = require("../lib/platform");

function binPath(target) {
  return path.join(
    __dirname,
    "..",
    "vendor",
    `${process.platform}-${process.arch}`,
    target.binary
  );
}

async function main() {
  let target;
  try {
    target = resolve();
  } catch (err) {
    process.stderr.write(`retna: ${err.message}\n`);
    process.exit(1);
  }

  let binary = binPath(target);

  if (!fs.existsSync(binary)) {
    process.stderr.write("retna: binary missing, downloading...\n");
    try {
      // install.js exports main, so the download can be awaited here. Awaited
      // rather than required for its side effect: otherwise the existence
      // check below runs before the first byte has arrived.
      await require("../install.js").main();
    } catch (err) {
      process.stderr.write(`retna: download failed: ${err.message}\n`);
    }
    binary = binPath(target);
  }

  if (!fs.existsSync(binary)) {
    process.stderr.write(
      "retna: could not install the binary.\n" +
        "retna: check https://github.com/d3uceY/Retna/releases for a manual download.\n"
    );
    process.exit(1);
  }

  const child = spawn(binary, process.argv.slice(2), { stdio: "inherit" });

  child.on("error", (err) => {
    process.stderr.write(`retna: ${err.message}\n`);
    process.exit(1);
  });

  child.on("exit", (code, signal) => {
    if (signal) process.kill(process.pid, signal);
    else process.exit(code ?? 1);
  });
}

main();
