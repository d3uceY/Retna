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
commits filtered out. Once the release is up, two more jobs publish the npm
packages described under "The npm packages" below.

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

### The npm packages

`npm/` holds what gets published to npm: the `retna` wrapper and six
`retna-<platform>-<arch>` packages. Each platform package carries the bare
binary for one os and cpu pair, so npm installs exactly one of them, picked by
the machine, and `retna` is a launcher that execs it. There is no `postinstall`
anywhere, which is what makes an install with `--ignore-scripts` work.

Nothing is downloaded on the user's machine and the release page is not used as
a package registry. `npm-platform` in the workflow does the fetching, on the
runner: it downloads the archive GoReleaser just uploaded, takes the binary out
of it, drops it into the matching package and publishes that. The wrapper goes
last, from `npm-wrapper`, because its `optionalDependencies` pin the six
versions and it can only be published once they exist.

The wrapper's README is generated rather than committed. npm renders a package
page out of the published tarball, where the relative paths the root README uses
for the logo and the skill folder lead nowhere, so `npm/scripts/readme.js`
rewrites every relative target to an absolute GitHub URL and writes
`npm/README.md`. That file is in `.gitignore`: the root README stays the only
copy of the documentation, and the package page is rebuilt from it on every
release.

The version is the tag, as everywhere else. `npm/scripts/version.js` takes the
tag, strips the leading `v`, checks the rest is semver npm will accept, and
writes the result into all seven `package.json` files and the wrapper's six
pins. Those files sit at `0.0.0` in the repo and nothing bumps them by hand:

```
node npm/scripts/version.js v0.2.0
```

Publishing by hand is not the supported path, but if you have to, run
`node npm/scripts/version.js` and `node npm/scripts/readme.js` first, then
`npm publish` from each package directory, platforms before the wrapper.
Publishing `0.0.0` straight out of a clean checkout, or a wrapper with no
README, is the one way to get this wrong.

The packages need a secret named `NPM_AUTH_TOKEN`. It has to be a token npm will
accept for publishing, and the failure when it will not is misleading: the token
authenticates, provenance gets signed and logged, and then the publish stops
with `E403 ... Two-factor authentication or granular access token with bypass
2fa enabled is required`. That is a 403 rather than a 401, so it reads like a
permissions problem when it is a token type problem. Either of these works:

- A granular access token with `Read and write` permission, `Bypass 2FA`
  enabled, and `All packages` selected. Not a list of packages: on the first
  release none of the seven exists yet, so there is nothing to select, and a
  token scoped to packages that do exist cannot create new ones.
- A classic `Automation` token, which bypasses 2FA by definition. A classic
  `Publish` token does not, because it wants a one time password that CI cannot
  supply.

Provenance is on, so both jobs ask for `id-token: write`. That is a workflow
permission and nothing to do with the token; each published tarball carries an
attestation naming the commit it was built from.

Publishing a version twice is an error, so both jobs check the registry first
and skip whatever is already there. That is what makes re-running a release
workflow after a failure safe.

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
  package.json     the retna wrapper: bin, optionalDependencies, files
  cli.js           execs the binary from the matching platform package
  scripts/
    version.js     writes the tag-derived version into every package file
    readme.js      builds the npm README from the root README
  platforms/       one package per os and arch pair, each holding bin/retna
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
