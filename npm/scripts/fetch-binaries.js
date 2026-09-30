#!/usr/bin/env node
// Downloads a release's archives and moves each binary into the matching
// platform package, so the packages can be staged or published from a checkout.
//
// The release workflow does the same thing on the runner, out of the archives
// GoReleaser has just uploaded. This exists for the two cases CI cannot cover:
// bootstrapping the seven names, which the registry will not let a token create,
// and the by-hand publish path described in CONTRIBUTING.md.
//
// The tag comes from the argument, or from RETNA_VERSION in .env.

'use strict';

const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const root = path.join(__dirname, '..');

// Mirrors the matrix in .github/workflows/release.yml. A platform added there
// belongs here too.
const PLATFORMS = [
  { package: 'retna-linux-x64', asset: 'retna_Linux_x86_64.tar.gz', format: 'tar.gz', binary: 'retna' },
  { package: 'retna-linux-arm64', asset: 'retna_Linux_arm64.tar.gz', format: 'tar.gz', binary: 'retna' },
  { package: 'retna-darwin-x64', asset: 'retna_Darwin_x86_64.tar.gz', format: 'tar.gz', binary: 'retna' },
  { package: 'retna-darwin-arm64', asset: 'retna_Darwin_arm64.tar.gz', format: 'tar.gz', binary: 'retna' },
  { package: 'retna-win32-x64', asset: 'retna_Windows_x86_64.zip', format: 'zip', binary: 'retna.exe' },
  { package: 'retna-win32-arm64', asset: 'retna_Windows_arm64.zip', format: 'zip', binary: 'retna.exe' },
];

const TAG = /^v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;

const given = (process.argv[2] ?? process.env.RETNA_VERSION ?? '').trim();
const tag = given.startsWith('v') ? given : `v${given}`;

if (!TAG.test(tag)) {
  console.error(`Not a release tag: ${given === '' ? '(nothing given)' : given}`);
  console.error('Usage: node npm/scripts/fetch-binaries.js [v1.2.3]');
  console.error('The tag can also come from RETNA_VERSION, set in .env.');
  process.exit(1);
}

const wrapper = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));

// "git+https://github.com/d3uceY/Retna.git" -> "d3uceY/Retna"
const slug = String(wrapper.repository?.url ?? '')
  .replace(/^git\+/, '')
  .replace(/\.git$/, '')
  .replace(/^https?:\/\/github\.com\//, '');

if (!/^[^/]+\/[^/]+$/.test(slug)) {
  console.error(
    `Could not read the GitHub repository from npm/package.json, got ${JSON.stringify(slug)}.`
  );
  process.exit(1);
}

function extract(archive, into, format) {
  // bsdtar is what Windows and macOS ship, and it reads both formats. GNU tar on
  // Linux only reads the tarballs, so a zip falls through to unzip.
  const attempts =
    format === 'zip'
      ? [
          ['tar', ['-xf', archive, '-C', into]],
          ['unzip', ['-qo', archive, '-d', into]],
        ]
      : [['tar', ['-xzf', archive, '-C', into]]];

  let last;
  for (const [command, args] of attempts) {
    try {
      execFileSync(command, args, { stdio: 'pipe' });
      return;
    } catch (err) {
      last = err;
    }
  }

  throw last;
}

async function main() {
  const work = fs.mkdtempSync(path.join(os.tmpdir(), 'retna-binaries-'));

  try {
    for (const entry of PLATFORMS) {
      const url = `https://github.com/${slug}/releases/download/${tag}/${entry.asset}`;

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`${entry.asset}: ${response.status} ${response.statusText} (${url})`);
      }

      const archive = path.join(work, entry.asset);
      fs.writeFileSync(archive, Buffer.from(await response.arrayBuffer()));

      const into = fs.mkdtempSync(path.join(work, 'out-'));
      extract(archive, into, entry.format);

      const source = path.join(into, entry.binary);

      if (!fs.existsSync(source)) {
        throw new Error(`${entry.asset} does not contain ${entry.binary}`);
      }

      // The archives also hold LICENSE and README.md. Take only the binary, the
      // same way the workflow does, so the package is one file and nothing else.
      const target = path.join(root, 'platforms', entry.package, 'bin', entry.binary);
      fs.mkdirSync(path.dirname(target), { recursive: true });
      fs.copyFileSync(source, target);

      if (process.platform !== 'win32') {
        fs.chmodSync(target, 0o755);
      }

      console.log(`${entry.package}/bin/${entry.binary}  ${fs.statSync(target).size} bytes`);
    }
  } finally {
    fs.rmSync(work, { recursive: true, force: true });
  }

  console.log(`\nBinaries from ${tag} are in place. Then:`);
  console.log('  node --env-file=.env npm/scripts/version.js');
  console.log('  node --env-file=.env npm/scripts/readme.js');
  console.log('See CONTRIBUTING.md, "The npm packages".');
}

main().catch((err) => {
  console.error(err.message);
  // Setting the code rather than calling process.exit: an abrupt exit while
  // undici is still tearing its sockets down trips a libuv assertion on Windows.
  process.exitCode = 1;
});
