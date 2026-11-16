<img src="assets/retna.png" alt="Retna" width="260">

Convert colors and measure contrast. Retna takes a color in almost any notation,
prints it in another space, or measures it against another. WCAG 2.x by
default, APCA behind `--algorithm`. It writes nothing to disk and keeps no
config. An [agent skill](skill/SKILL.md) teaches a coding agent to drive it.

## Install

macOS, from the [Homebrew tap](https://github.com/d3uceY/homebrew-retna):

```
brew install --cask d3uceY/homebrew-retna/retna
```

npm, which downloads the binary for your platform:

```
npm install -g @deuce/retna
```

npm 11.4 and newer block install scripts, so that download can be skipped. If
`retna` fails or npm warns about install scripts, allow the package and install
again:

```
npm install -g @deuce/retna --allow-scripts=@deuce/retna
```

The flag is for global installs and `npx`; npm rejects it in a project install.
Either way, `retna` fetches the binary on first run.

Without installing:

```
npx @deuce/retna contrast "#777" "#fff"
```

Prebuilt binaries for Linux, macOS and Windows are on the
[releases page](https://github.com/d3uceY/Retna/releases); unpack one onto your
PATH. With Go 1.25 or newer:

```
go install github.com/d3uceY/Retna@latest
```

From a clone:

```
git clone https://github.com/d3uceY/Retna
cd Retna
go build -o retna .
```

## contrast

```
retna contrast "#777" "#fff"
```

```
Contrast

Foreground      #777777
Background      #FFFFFF

WCAG
Ratio       4.48:1
AA Normal   FAIL
AA Large    PASS
AAA Normal  FAIL
AAA Large   FAIL
```

`--against` measures one foreground against a list:

```
retna contrast "#777" --against "#fff,#000,#f5f5f5,#101010"
```

```
Contrast

Foreground      #777777

Color       Ratio   AA    AAA
──────────  ──────  ────  ────
   #FFFFFF  4.48:1  FAIL  FAIL
   #000000  4.69:1  PASS  FAIL
   #F5F5F5  4.11:1  FAIL  FAIL
   #101010  4.25:1  FAIL  FAIL
```

`--against` repeats, and each value may be a comma list:

```
retna contrast "#777" --against "#fff" --against "#000,#101010"
```

`--all` runs every algorithm:

```
retna contrast red white --all
```

### Using it in CI

`--min` takes a ratio and exits 1 when a pair falls short.

```
retna contrast "$FG" "$BG" --min 4.5
```

```
FAIL  4.48:1
Required: 4.50:1
```

```yaml
# .github/workflows/a11y.yml
- name: Check color contrast
  run: retna contrast "${{ vars.FG }}" "${{ vars.BG }}" --min 4.5
```

With `--against`, the summary counts the pairs that failed. `--min` compares
WCAG, so it needs the wcag algorithm: `--all` includes it, and `--min` with
`--algorithm apca` is an error. The value must be a number above zero.

## convert

`--to` takes one space, a comma separated list, or `all`.

```
retna convert "#1a73e8" --to hsl,oklch
```

```
    #1A73E8

HSL    hsl(214.08, 81.75%, 50.59%)
OKLCH  oklch(0.5737, 0.1946, 257.86)
```

With no `--to`, every space is printed.

## inspect

`inspect` prints every space plus the relative luminance.

```
retna inspect "#3498db"
```

```
    Color

HEX           #3498DB
RGB           rgb(52, 152, 219)
RGBA          rgba(52, 152, 219, 1)
HSL           hsl(204.07, 69.87%, 53.14%)
HSV           hsv(204.07, 76.26%, 85.88%)
HWB           hwb(204.07 20.39% 14.12%)
LAB           lab(59.5, -12.1, -43.14)
LCH           lch(59.5, 44.8, 254.33)
OKLAB         oklab(0.6531, -0.0618, -0.1197)
OKLCH         oklch(0.6531, 0.1347, 242.69)
SRGB          color(srgb 0.2039 0.5961 0.8588)
DISPLAY-P3    color(display-p3 0.3208 0.588 0.8369)
A98-RGB       color(a98-rgb 0.3725 0.5905 0.8459)
PROPHOTO-RGB  color(prophoto-rgb 0.4327 0.51 0.7876)
REC2020       color(rec2020 0.3766 0.5394 0.8141)

Relative luminance
Luminance  0.2830
```

## palette contrast

Same measurements as `contrast --against`, with colors from a file or stdin:

```
retna palette contrast --foreground white --file palette.txt
retna palette contrast --foreground white --file - < palette.txt
retna palette contrast --foreground white --colors "#000,#111,#666"
```

One color per line, comma lists allowed. Blank lines and `//` are skipped; a
UTF-8 BOM is ignored.

```
Contrast

Foreground      #FFFFFF

Color       Ratio    AA    AAA
──────────  ───────  ────  ────
   #000000  21.00:1  PASS  PASS
   #111111  18.88:1  PASS  PASS
   #666666  5.74:1   PASS  FAIL
```

## readable

Which text color reads best on a background.

```
retna readable "#3498db"
```

```
    Background #3498DB

Text color  Ratio   AA
──────────  ──────  ────
   #000000  6.66:1  PASS
   #FFFFFF  3.15:1  FAIL
```

Candidates default to black and white: `--candidates` for your own,
`--level AAA` to raise the bar, `--best` for the winner alone. Alias:
`text-color`.

## fix

`fix` walks OKLab lightness toward more contrast, leaving chroma and hue alone.

```
retna fix "#777777" --background "#ffffff" --level AA
```

```
Fix

Original       #777777  4.48:1  FAIL
Background     #FFFFFF
Target                  4.50:1

Suggested colors
Color       Ratio   WCAG  Note
──────────  ──────  ────  ───────────────
   #767676  4.52:1  PASS  smallest change
   #757575  4.59:1  PASS
   #747474  4.65:1  PASS
```

`--target` sets the ratio, `--suggest` the number of alternatives. A color that
already passes is reported and left alone. The target must sit in 1:1–21:1,
since black on white is the highest there is, and `--suggest` needs at least 1.
Suggestions are measured after 8-bit rounding, so the hex shown is the one that
meets the ratio.

## Input formats

Hex, hash optional, in 3, 4, 6 or 8 digits:

```
#fff   #ffff   #ffffff   #ffffff80
```

Functional notation takes commas, spaces, or a slash before alpha:

```
rgb(26, 115, 232)      rgb(26 115 232)      rgb(26 115 232 / 0.5)
hsl(214.08 81.75% 50.59%)                   hsla(214.08, 81.75%, 50.59%, 0.5)
hsv(214.08 88.79% 90.98%)                   hwb(214.08 10.2% 9.02%)
lab(48.77 9.69 -67.47)                      lch(48.77 68.16 278.18)
oklab(0.5737 -0.0409 -0.1902)               oklch(0.5737 0.1946 257.86)
```

Angles take `deg`, `rad`, `grad` or `turn`; percentages work wherever CSS allows
them. Names are case insensitive, so `rebeccapurple`, `RebeccaPurple` and
`#663399` are one color. `transparent` works. Wide gamut spaces come through
`color()`:

```
color(srgb 0.2039 0.5961 0.8588)     color(display-p3 0.3208 0.588 0.8369)
color(a98-rgb 0.3725 0.5905 0.8459)  color(prophoto-rgb 0.4327 0.51 0.7876)
color(rec2020 0.3766 0.5394 0.8141)  color(display-p3 1 0 0 / 0.5)
```

Components take numbers or percentages and may fall outside 0..1, which is how
CSS names colors sRGB cannot hold. Retna keeps them, and clips only when the
color is drawn as sRGB.

## Color spaces

```
hex, rgb, rgba, hsl, hsv, hwb, lab, lch, oklab, oklch,
srgb, display-p3, a98-rgb, prophoto-rgb, rec2020
```

Lab, LCH and ProPhoto use D50, as CSS specifies; OKLab, OKLCH, display-p3,
a98-rgb and rec2020 use D65. sRGB to a wider space is never clipped; a `color()`
value outside the sRGB gamut is clipped to the nearest displayable color.

## JSON output

`--json` works on every command:

```
retna contrast "#777" "#fff" --json
```

```json
{
  "foreground": "#777777",
  "background": "#FFFFFF",
  "contrast": 4.478089453577214,
  "wcag": {
    "aa": { "normal": false, "large": true },
    "aaa": { "normal": false, "large": false }
  }
}
```

With `--against` the pairs go under `results`; `--algorithm apca` adds `apca`
holding Lc. `inspect --json` gives each space as three numbers plus `luminance`:

```json
{
  "input": "#3498db",
  "hex": "#3498DB",
  "alpha": 1,
  "rgb": [52, 152, 219],
  "hsl": [204.0719, 0.6987, 0.5314],
  "hsv": [204.0719, 0.7626, 0.8588],
  "hwb": [204.0719, 0.2039, 0.1412],
  "lab": [59.4961, -12.1034, -43.1352],
  "lch": [59.4961, 44.8011, 254.3263],
  "oklab": [0.6531, -0.0618, -0.1197],
  "oklch": [0.6531, 0.1347, 242.6867],
  "srgb": [0.2039, 0.5961, 0.8588],
  "display-p3": [0.3208, 0.588, 0.8369],
  "a98-rgb": [0.3725, 0.5905, 0.8459],
  "prophoto-rgb": [0.4327, 0.51, 0.7876],
  "rec2020": [0.3766, 0.5394, 0.8141],
  "luminance": 0.283
}
```

One line per group here; the command indents each number. Hues are degrees,
saturation channels 0..1, Lab lightness 0..100, OKLab 0..1. Wide gamut
components are raw and may sit outside 0..1.

## Transparency

Both algorithms measure opaque colors. With alpha, Retna composites the
background over white, then the foreground over that, as a browser would paint,
and says so.

## APCA

`--algorithm apca` returns Lc: positive for dark text on light, negative for
light on dark, so compare absolute values. Checks use the 45, 60 and 75 bands.

## Using it from an AI agent

[`skill/SKILL.md`](skill/SKILL.md) builds and audits UI color systems with
Retna as the measuring instrument: pick a ramp, check every text and surface
pair. References cover the CLI, palette recipes, dark mode, thresholds and color
science. Copy it into your agent's skills directory:

```
cp -r skill ~/.copilot/skills/retna-color-expert     # VS Code / Copilot
cp -r skill ~/.claude/skills/retna-color-expert      # Claude Code
```

Plain markdown, nothing fetched at runtime. Command reference:
[`skill/references/retna-cli.md`](skill/references/retna-cli.md). Homebrew, npm,
`go install` and release archives ship the binary only, so clone for `skill/`.

## License

MIT. See [LICENSE](LICENSE).

<img src="assets/retna-rainbow.gif" alt="The Retna wordmark in the Game Boy Color rainbow palette" width="100%">
