#!/usr/bin/env node
// Launches the retna binary for this machine.
//
// The binary is not in this package. It ships in retna-<platform>-<arch>
// packages listed as optionalDependencies, and npm installs only the one whose
// os and cpu match, skipping the other five. Nothing is downloaded at install
// time, no lifecycle script runs, and nothing is written at run time, so an
// install with --ignore-scripts ends up with a working command.

'use strict';

const { spawn } = require('node:child_process');

// npm's own vocabulary for what a platform package is allowed to declare in its
// os and cpu fields, which is process.platform and process.arch.
const SUPPORTED = [
  'darwin-arm64',
  'darwin-x64',
  'linux-arm64',
  'linux-x64',
  'win32-arm64',
  'win32-x64',
];

const target = `${process.platform}-${process.arch}`;
const binary = process.platform === 'win32' ? 'retna.exe' : 'retna';

if (!SUPPORTED.includes(target)) {
  console.error(`retna has no binary for ${target}.`);
  console.error(`It ships one for: ${SUPPORTED.join(', ')}.`);
  console.error(
    'The releases page has archives for other platforms: https://github.com/d3uceY/Retna/releases'
  );
  process.exit(1);
}

let binaryPath;
try {
  // Resolution goes through node rather than a join so that hoisted and nested
  // node_modules layouts both work.
  binaryPath = require.resolve(`retna-${target}/bin/${binary}`);
} catch {
  console.error(`retna could not find its retna-${target} binary.`);
  console.error('');
  console.error('That package is an optional dependency, so it is missing if you');
  console.error('installed with --no-optional, if your registry mirror skipped it,');
  console.error('or if Yarn PnP did not unpack it.');
  console.error('');
  console.error('Reinstalling usually fixes it:  npm install retna');
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
  console.error(`retna could not start ${binaryPath}: ${err.message}`);
  process.exit(1);
});

child.on('close', (code, signal) => {
  // A signal means the child was killed from outside. Re-raise it rather than
  // reporting an exit code, so retna looks the same as a bare binary would.
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }

  process.exit(code === null ? 1 : code);
});
