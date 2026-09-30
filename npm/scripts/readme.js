#!/usr/bin/env node
// Builds npm/README.md from the root README, and prints what it wrote.
//
// npm renders a package's README out of the published tarball, and the tarball
// holds cli.js and the license and nothing else. Every relative path the root
// README uses, for the logo and for the skill folder, would point at nothing
// there. Rather than keep a second copy of the documentation that drifts, the
// npm page is built from the one the repo already maintains, with each relative
// target rewritten to an absolute GitHub URL.
//
// The output is generated, so it is not committed. See .gitignore.

'use strict';

const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..');
const source = path.join(root, '..', 'README.md');
const target = path.join(root, 'README.md');

// Pinned to the branch rather than the tag, so a documentation fix shows up on
// the package page without waiting for a release.
const BRANCH = 'main';

const IMAGE = /\.(png|jpe?g|gif|svg|webp|avif)$/i;
const ABSOLUTE = /^[a-z][a-z0-9+.-]*:/i;

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

const blob = `https://github.com/${slug}/blob/${BRANCH}`;
const raw = `https://raw.githubusercontent.com/${slug}/${BRANCH}`;

let rewrites = 0;

function absolute(href) {
  // Already absolute, an in-page anchor, or a mailto. Leave it alone.
  if (ABSOLUTE.test(href) || href.startsWith('#')) {
    return href;
  }

  const hash = href.indexOf('#');
  const file = (hash === -1 ? href : href.slice(0, hash)).replace(/^\.?\//, '');
  const anchor = hash === -1 ? '' : href.slice(hash);

  // Images have to come off raw, everything else reads better as a file page.
  const base = IMAGE.test(file) ? raw : blob;

  rewrites += 1;
  return `${base}/${file}${anchor}`;
}

let text = fs.readFileSync(source, 'utf8');

// [label](target)
text = text.replace(/(\]\()([^)\s]+)(\))/g, (match, open, href, close) => {
  return open + absolute(href) + close;
});

// <img src="target">
text = text.replace(/(<img\b[^>]*?\bsrc=")([^"]+)(")/g, (match, open, href, close) => {
  return open + absolute(href) + close;
});

fs.writeFileSync(target, text);

console.log(`npm/README.md written from the root README, ${rewrites} relative targets made absolute.`);
