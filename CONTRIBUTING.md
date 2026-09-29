# Contributing to Retna

## Prerequisites

Go 1.25 or newer. Retna has no other build requirements.

## Build

```
go build -o retna .
```

On Windows the output is `retna.exe`. To put it on your PATH, move that binary
somewhere already on PATH, or add its folder to PATH.

## Test

```
go test ./...
go vet ./...
```

Every package has tests. The conversions are checked against published
reference values, and the formatters are checked by converting each space and
parsing the result back.

## Project layout

```
main.go            root command, version, shared flag helpers
cmd_contrast.go    contrast, including --against and --min
cmd_convert.go     convert, including --to
cmd_inspect.go     inspect
cmd_palette.go     palette contrast, including file and stdin input
cmd_readable.go    readable, plus the level lookup
cmd_fix.go         fix, including the OKLab lightness search
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

1. Fork the repository and branch from `main`.
2. Keep the change focused on one thing.
3. Run the build, the vet pass and the tests above.
4. Open a pull request describing what changed and why.

Open an issue first for anything large, so the direction can be agreed before
the code is written.
