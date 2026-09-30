#!/usr/bin/env node
// Sets the version across every npm package file at once, and prints the version
// it settled on.
//
// The version lives in eight places: seven package files plus the wrapper's
// optionalDependencies pins. They have to agree, or `npm install retna` asks for
// platform packages that were never published. The release workflow calls this
// with the git tag and reuses the printed version for its "already published?"
// checks, which makes this the one place that knows how a tag becomes a version.

'use strict';

const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..');

// Every package published out of npm/, wrapper first.
const PACKAGES = [
  '.',
  'platforms/retna-darwin-arm64',
  'platforms/retna-darwin-x64',
  'platforms/retna-linux-arm64',
  'platforms/retna-linux-x64',
  'platforms/retna-win32-arm64',
  'platforms/retna-win32-x64',
];

const SEMVER = /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;

const given = (process.argv[2] ?? process.env.RETNA_VERSION ?? '').trim();
const version = given.startsWith('v') ? given.slice(1) : given;

if (!SEMVER.test(version)) {
  console.error(`Not a version: ${given === '' ? '(nothing given)' : given}`);
  console.error('Pass a tag (v1.2.3), a plain version (1.2.3), or set RETNA_VERSION in .env');
  console.error('and run: node --env-file=.env npm/scripts/version.js');
  process.exit(1);
}

function load(dir) {
  const file = path.join(root, dir, 'package.json');

  if (!fs.existsSync(file)) {
    console.error(`Missing ${path.relative(root, file)}.`);
    process.exit(1);
  }

  try {
    return { dir, file, json: JSON.parse(fs.readFileSync(file, 'utf8')) };
  } catch (err) {
    console.error(`Could not read ${path.relative(root, file)}: ${err.message}`);
    process.exit(1);
  }
}

const packages = PACKAGES.map(load);
const [wrapper, ...platforms] = packages;

// The pins and the packages have to be the same set. If they drift, a platform
// goes unpublished and only the machines that need it break.
const pinned = Object.keys(wrapper.json.optionalDependencies ?? {});
const named = platforms.map((platform) => platform.json.name);
const unpinned = named.filter((name) => !pinned.includes(name));
const unbuilt = pinned.filter((name) => !named.includes(name));

if (unpinned.length > 0 || unbuilt.length > 0) {
  for (const name of unpinned) {
    console.error(`${name} is not in the wrapper's optionalDependencies.`);
  }
  for (const name of unbuilt) {
    console.error(`The wrapper pins ${name}, which no package under npm/platforms provides.`);
  }
  process.exit(1);
}

for (const pkg of packages) {
  pkg.json.version = version;
}

for (const name of pinned) {
  wrapper.json.optionalDependencies[name] = version;
}

for (const pkg of packages) {
  fs.writeFileSync(pkg.file, `${JSON.stringify(pkg.json, null, 2)}\n`);
}

// The workflow reads this off stdout, so it is the only thing printed here.
console.log(version);
