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
of it, drops it into the matching package and stages that. The wrapper goes last,
from `npm-wrapper`, because its `optionalDependencies` pin the six versions and
it can only go out once they exist.

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
`npm stage publish` from each package directory, platforms before the wrapper,
and approve each one. Publishing `0.0.0` straight out of a clean checkout, or a
wrapper with no README, is the one way to get this wrong.

### Releases are staged, not published

A release puts the seven packages in npm's staging area, and someone approves
them with 2FA before they become installable. So `NPM_AUTH_TOKEN` is a stage-only
granular token with no 2FA bypass, and it does not need to be anything else.
Staged publishing exists so that bypass tokens stop being necessary; npm's own
guidance is to delete them and use either a trust relationship or a stage-only
token, which is what this repo does.

Four things follow from that:

- Nothing is live until it is approved. All seven packages, every version.
  `npm stage list` shows what is waiting.
- `npm stage publish` needs npm 11.15.0 or newer. Node 24.12 ships 11.6.2, which
  is the only reason both jobs install npm before doing anything.
- Staged and published versions share one index, so a version cannot be staged
  twice. Re-running after a successful release is safe, because the `npm view`
  guard skips what is already published. Re-running while something is staged
  and unapproved fails instead, and `npm stage reject <stage-id>` clears it.
- Staging a package that does not exist creates it, as a public `0.0.0-stage`
  placeholder. That is what lets the seven new names exist before anything else
  can be configured for them.

A token that cannot stage stops the run with `E403 ... bypass 2fa enabled is
required`, and one that can publish but wants a one-time password stops it with
`EOTP: This operation requires a one-time password from your authenticator`.
Neither reads like a token type problem, and both are one.

### Moving to trusted publishing

Once the seven packages exist, a trust relationship lets CI authenticate with
OIDC instead of a token, and provenance is generated automatically rather than
from the `--provenance` flag. The packages have to exist first, which is why this
comes after a staged release and not before.

With npm 11.15.0 or newer locally, logged in as yourself:

```
for p in retna retna-linux-x64 retna-linux-arm64 retna-darwin-x64 \
         retna-darwin-arm64 retna-win32-x64 retna-win32-arm64; do
  npm trust github "$p" --file release.yml --repo d3uceY/Retna \
    --allow-stage-publish --yes
  sleep 2
done
```

`--file` is the workflow filename on its own, not its path. `--allow-stage-publish`
and not `--allow-publish`: staging is the whole point, and npm recommends giving
a trust relationship the narrower of the two. The first call prompts for 2FA and
offers to skip it for the next five minutes, which covers the other six.
`npm trust` refuses a token with 2FA bypass enabled, so use a session login.

The same thing can be done by hand on each package's Settings page under
Trusted publishing. Prefer the loop: npm does not validate those fields when you
save them and every one is case sensitive, so a typo shows up only as `ENEEDAUTH`
at the next release. `npm trust` errors if a relationship already exists there;
`npm trust list <package>` prints the id and `npm trust revoke --id <id>`
removes it so you can create the replacement.

After a release has gone through OIDC, revoke the token and delete the secret.
The CLI prefers OIDC when it is available and falls back to the token only when
it is not, so nothing breaks in between.

Provenance is on throughout, so both jobs ask for `id-token: write`. Under a
trust relationship that permission is how npm authenticates at all; with the
token it only feeds the attestation.

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
