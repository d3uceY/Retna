# Contributing to Retna

## Prerequisites

Go 1.25 or newer. Retna has no other build requirements.

## Build

```
go build -o retna .
```

On Windows the output is `retna.exe`. To put it on your PATH, move that binary
somewhere already on PATH, or add its folder to PATH.

To stamp a version into the binary, override the `cmd.version` variable:

```
go build -ldflags "-X github.com/d3uceY/Retna/cmd.version=v1.2.3" -o retna .
```

## Test

```
go test ./...
go vet ./...
```

Every package has tests. The conversions are checked against published
reference values, and the formatters are checked by converting each space and
parsing the result back.

Tests sit next to the code they cover, so a package carries its tests with it
when it moves. The CLI tests live in `cmd/` beside the commands they exercise,
and there are no test files at the module root.

## Releasing

The released version is the git tag. Nothing else is bumped by hand.

```
git tag v0.2.0
git push origin v0.2.0
```

Pushing a `v`-prefixed tag starts `.github/workflows/release.yml`, which vets
and tests the tree and then runs GoReleaser. GoReleaser builds amd64 and arm64
binaries for Linux, macOS and Windows, packs each one with the license and
readme, writes a checksum file and opens the GitHub release. The changelog comes
from the commits since the previous tag, with `docs:`, `test:` and `chore:`
commits filtered out. Once the release is up, the same job publishes the npm
package described under "The npm package" below.

The version the binary reports is injected at build time, so `retna version`
always matches the tag. The placeholder in `cmd/root.go` is only what an
untagged local build shows.

To exercise the pipeline without tagging:

```
goreleaser release --snapshot --clean
```

Artifacts land in `dist/`, which is ignored by git. A snapshot also writes the
cask it would have published to `dist/homebrew/Casks/retna.rb`, which is the
quickest way to see what the tap would receive.

### The Homebrew tap

The release publishes a cask to `d3uceY/homebrew-retna`, which is a separate
repository. Two things have to exist before the first tagged release that
includes it:

1. The tap repository itself, public, named `homebrew-retna`.
2. A secret named `PUBLISHER_TOKEN` on this repository, holding a token that can
   write to the tap. A fine-grained token needs `Contents: read and write` on
   that one repository. A classic token needs the `repo` scope.

The workflow passes it to GoReleaser as `HOMEBREW_TAP_GITHUB_TOKEN`. The
built-in `GITHUB_TOKEN` is scoped to the repository the workflow runs in, so it
cannot push to the tap, and a release without the secret fails at the tap step.

Tap commits are authored by GoReleaser's default bot. To use your own name and
email, add a `commit_author` block under `homebrew_casks` in
`.goreleaser.yaml`.

The cask installs an unsigned binary, so macOS Gatekeeper may block the first
run. If that happens, either reinstall with `brew install --cask
--no-quarantine d3uceY/retna/retna`, or clear the attribute with
`xattr -d com.apple.quarantine $(which retna)`.

Homebrew casks are a macOS feature. GoReleaser emits Linux stanzas as well, but
treat them as unused: Linux users should use the tarball.

### The npm package

`npm/` holds the one package published to npm, `retna`. npm cannot publish a
native executable, so the package ships a JS shim, `bin/retna.js`, which spawns
a binary that `install.js` downloads from the GitHub release for the package
version. It is the same arrangement as `@deuce/piper`: one package covers every
platform, instead of seven packages that all have to be published and approved
in lockstep.

The name is unscoped because it is free and the shorter install is worth it. If
npm ever refuses it as too similar to an existing package — the error is
`403 ... name too similar to existing package ...` — rename it to
`@deuce/retna`, the scope `@deuce/piper` already uses. The `name` field in
`npm/package.json` is the only place the name is written; the token's package
scope, the README and this file are the only places it is documented.

The version is the tag, as everywhere else. The workflow stamps it into
`npm/package.json` before publishing, so tag `v0.2.0` becomes npm `0.2.0`, and
`install.js` asks for `releases/download/v0.2.0/...`. The two have to agree: a
package whose version has no release installs and then fails to download, which
is why GoReleaser runs first in the same job.

`files` is an allowlist of `bin/`, `lib/` and `install.js`. The download lands
in `vendor/`, which is ignored by git and by npm. `lib/platform.js` maps
`process.platform` and `process.arch` onto the archive names GoReleaser writes,
so it has to stay in step with `name_template` in `.goreleaser.yaml`.

The package's README is generated rather than committed. npm renders a package
page out of the published tarball, where the relative paths the root README uses
for the logo and the skill folder lead nowhere, so `npm/scripts/readme.js`
rewrites every relative target to an absolute GitHub URL and writes
`npm/README.md`. That file is in `.gitignore`: the root README stays the only
copy of the documentation, and the package page is rebuilt from it on every
release.

### The postinstall is optional

npm 11.4+ blocks dependency install scripts by default, so a fresh install can
print:

```
npm warn install-scripts   retna@0.1.1 (postinstall: node install.js)
```

That is not a broken install. `install.js` treats every failure as a warning,
and `bin/retna.js` retries the download on first run, so a blocked or failed
script costs a slower first run rather than a missing command. Users who want
the download at install time allow it in one of two ways:

```
npm install -g --allow-scripts=retna                 # this install only
npm config set allow-scripts=retna --location=user   # all future installs
```

Keep that property. If the shim ever requires `vendor/` to exist, every npm
11.4+ user gets a broken CLI. `RETNA_SKIP_DOWNLOAD=1` skips the download
entirely, for offline and air-gapped installs.

### The npm token

`NPM_TOKEN` is a repository secret holding a **granular access token** created
at <https://www.npmjs.com/settings/deuce/tokens>:

| Field | Value |
| --- | --- |
| Packages and scopes | `retna` |
| Permissions | **Read and write** |
| Bypass two-factor authentication | **enabled** |

The 2FA bypass is not optional. Without it CI aborts with `EOTP: This operation
requires a one-time password`, because a runner cannot answer a 2FA prompt. An
**Automation** classic token also works (it is always 2FA-exempt) but grants
publish rights on every package, so the granular token is the narrower choice.

The workflow also asks for `permissions: id-token: write`, which
`npm publish --provenance` uses to sign an attestation through GitHub's OIDC
provider.

Never commit the token, echo it, or paste it into a chat. Read it only through
`${{ secrets.NPM_TOKEN }}` in the workflow. If it is ever exposed, revoke it on
npmjs.com immediately and add a new secret.

### npm error → cause → fix

| Error | Cause | Fix |
| --- | --- | --- |
| `404 Not Found - PUT .../retna` | Publishing to a package the account cannot write, or wrong account | Log in as the package owner |
| `403 ... name too similar to existing package` | npm considers `retna` too close to an existing name | Rename the package to `@deuce/retna`; see "The npm package" |
| `ENEEDAUTH` / `need auth` | `NPM_TOKEN` missing, empty, or expired | Create or refresh the secret |
| `EOTP` / requires one-time password | Token cannot bypass 2FA | Recreate as granular **with** 2FA bypass |
| `EUSAGE: Provenance generation ... requires "write" access to the "id-token" permission` | Missing permission | Add `id-token: write` to the workflow `permissions` |
| `EPUBLISHCONFLICT` / `403 ... cannot publish over the previously published versions` | Version already exists | Bump the tag; never reuse a version |
| Install fails to download, `HTTP 404` on the release URL | No GitHub release for the package version | Let GoReleaser run for that tag first |
| Install prints `skipping binary download` | `RETNA_SKIP_DOWNLOAD=1` set | Unset it |

### Publishing by hand

CI is the normal path. To publish from a laptop, with the release already on
GitHub:

```
cd npm
npm login                                          # answer the 2FA prompt yourself
node scripts/readme.js
npm version X.Y.Z --no-git-tag-version
npm publish --access public
```

`--provenance` will not produce an attestation outside GitHub Actions; that is
expected and harmless. Prefer `npm login` to exporting a token into a shell
history.

You cannot unpublish a version after npm's restriction window, and even inside
it, unpublishing breaks anyone who already installed. The supported move is to
publish a new patch with the fix, and `npm dist-tag add retna@X.Y.Z latest`
if the bad release is currently the default.

## Project layout

```
main.go            entry point, a one line call into cmd
cmd/
  root.go          root command, version, shared flag helpers
  contrast.go      contrast, including --against and --min
  convert.go       convert, including --to
  inspect.go       inspect
  palette.go       palette contrast, including file and stdin input
  readable.go      readable, plus the level lookup
  fix.go           fix, including the OKLab lightness search
color/
  color.go         Color type, luminance, sRGB companding
  parse.go         parsing every input notation
  convert.go       conversions between spaces and the display formatter
  gamut.go         the CSS color() spaces: matrices and transfer functions
  named.go         the CSS named colors
contrast/
  contrast.go      the Algorithm interface and the registry
  wcag.go          WCAG 2.x
  apca.go          APCA
output/
  output.go        palette, swatches, ANSI aware column and table rendering
  json.go          the JSON view models
npm/
  package.json     the npm package: bin, files, postinstall
  bin/retna.js     the retna command; spawns the downloaded binary
  install.js       postinstall: downloads the release archive into vendor/
  lib/platform.js  maps process.platform/arch onto the release asset names
  scripts/
    readme.js      builds the npm README from the root README
```

## Design notes

Color is the only representation the rest of the program sees. Every input
notation is parsed down to gamma encoded sRGB, and every space is a pair of
functions that go to and from it. Adding a space such as HSL or OKLab means
writing those two functions and listing the name in `color.Spaces`.

The CSS `color()` spaces work differently. Each one is a row in the `gamuts`
table in `color/gamut.go`: two matrices, a flag for the D50 white point, and a
decode and encode pair. One generic path then serves all of them, so a new
space is a table entry plus a name in `wideGamutNames`. Output into those spaces
is deliberately left unclipped, because CSS allows components outside 0..1 and
clipping one would name a different color. The sRGB transfer functions in
`color.go` preserve the sign for the same reason.

Contrast algorithms live behind `contrast.Algorithm`, which takes two colors
and returns a value plus a set of named checks. Nothing in `color` knows about
WCAG or APCA, and the tests in `contrast` cover both.

The renderer measures cells with `lipgloss.Width` rather than counting bytes.
Styled cells carry ANSI escapes, so a byte count would push later columns
around. Keep that in mind if you add a table.

## Adding a contrast algorithm

Copy the shape of `contrast/wcag.go`: a type with a `Name` method and a
`Calculate` method, plus an `init` that registers it. The registry is what
`--algorithm` and `--all` read from, and the help text is generated from it, so
a new algorithm shows up without touching the command layer.

## Submitting changes

Participation is covered by the [Code of Conduct](CODE_OF_CONDUCT.md).

1. Fork the repository and branch from `main`.
2. Keep the change focused on one thing.
3. Run the build, the vet pass and the tests above.
4. Open a pull request describing what changed and why.

Open an issue first for anything large, so the direction can be agreed before
the code is written.
