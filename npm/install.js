"use strict";

// Downloads the retna binary for this machine from the matching GitHub
// release, unpacks it into vendor/, and leaves the executable at
// vendor/<binary> for bin/retna.js to spawn.
//
// Runs as the package's postinstall. It must never fail the install for an
// optional reason (offline, rate limit, no prebuilt for the platform): the
// wrapper still ships bin/retna.js, which retries on first run.

const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const zlib = require("node:zlib");
const { execFileSync } = require("node:child_process");
const { resolve } = require("./lib/platform");

const pkg = require("./package.json");
const VENDOR = path.join(__dirname, "vendor");
const REPO = "d3uceY/Retna";

function tag() {
  // package.json 0.1.1 -> v0.1.1. Set by the release workflow.
  return `v${pkg.version}`;
}

function url(asset) {
  return `https://github.com/${REPO}/releases/download/${tag()}/${asset}`;
}

function get(href, redirects = 5) {
  return new Promise((resolveP, rejectP) => {
    https
      .get(href, { headers: { "user-agent": "retna-npm-install" } }, (res) => {
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location &&
          redirects > 0
        ) {
          res.resume();
          resolveP(get(res.headers.location, redirects - 1));
          return;
        }
        if (res.statusCode !== 200) {
          res.resume();
          rejectP(new Error(`${href} returned HTTP ${res.statusCode}`));
          return;
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolveP(Buffer.concat(chunks)));
        res.on("error", rejectP);
      })
      .on("error", rejectP);
  });
}

// unpack writes the executable from a tar.gz or zip buffer. Only shells out to
// `tar`, which ships with Windows 10+ and every Linux/macOS, rather than
// pulling in a dependency for two archive formats.
function unpack(buf, archive, dest) {
  fs.rmSync(dest, { recursive: true, force: true });
  fs.mkdirSync(dest, { recursive: true });

  const tmp = path.join(VENDOR, `.download.${archive}`);
  fs.mkdirSync(VENDOR, { recursive: true });
  fs.writeFileSync(tmp, archive === "zip" ? buf : zlib.gunzipSync(buf));

  try {
    execFileSync("tar", ["-xf", tmp, "-C", dest], { stdio: "ignore" });
  } finally {
    fs.rmSync(tmp, { force: true });
  }
}

async function main() {
  // Opt-out for offline or air-gapped installs.
  if (process.env.RETNA_SKIP_DOWNLOAD === "1") {
    console.log("retna: skipping binary download (RETNA_SKIP_DOWNLOAD=1)");
    return;
  }

  const target = resolve();
  const dest = path.join(VENDOR, `${process.platform}-${process.arch}`);
  const binary = path.join(dest, target.binary);

  if (fs.existsSync(binary)) return;

  const href = url(target.asset);
  console.log(`retna: downloading ${target.asset}`);

  const buf = await get(href);
  unpack(buf, target.archive, dest);

  if (!fs.existsSync(binary)) {
    throw new Error(`archive ${target.asset} did not contain ${target.binary}`);
  }
  if (process.platform !== "win32") fs.chmodSync(binary, 0o755);
}

// A failed download must not break `npm install`: the CLI retries on first
// run with a clearer error. Warnings only.
if (require.main === module) {
  main().catch((err) => {
    console.warn(`retna: binary download failed: ${err.message}`);
    console.warn(`retna: run "retna" once installed to retry the download.`);
  });
}

module.exports = { main };
