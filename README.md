# Retna

Convert colors and measure how readable they are.

Retna reads a color in almost any notation, normalizes it to sRGB, and then
either prints it in another color space or measures it against another color.
Contrast comes from WCAG 2.x by default, and APCA is available behind
`--algorithm`.

It writes nothing to disk and keeps no configuration file. Every command reads
its arguments and prints to stdout.

## Install

Go 1.25 or newer is required.

```
go install github.com/d3uceY/Retna@latest
```

Or build from a clone:

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

Give it a foreground and a background. Add `--against` to check one foreground
against a list instead:

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

`--against` can be repeated, and each value may hold a comma separated list:

```
retna contrast "#777" --against "#fff" --against "#000,#101010"
```

Add `--all` to run every algorithm and print each one in turn:

```
retna contrast red white --all
```

### Using it in CI

`--min` takes a ratio. If any pair falls short, Retna prints the verdict and
exits 1.

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

With `--against`, the summary line reports how many pairs failed. The exit code
is 1 when at least one did.

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

This does the same measurements as `contrast --against`, but the colors come
from a file or from stdin. It saves typing when the list already exists
somewhere else.

```
retna palette contrast --foreground white --file palette.txt
retna palette contrast --foreground white --file - < palette.txt
retna palette contrast --foreground white --colors "#000,#111,#666"
```

The file holds one color per line, and each line may also be a comma separated
list. Blank lines and lines starting with `//` are skipped.

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

`readable` answers the other direction. Given a background, which text color
reads best on it?

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

Candidates default to black and white. Use `--candidates` to try your own,
`--level AAA` to raise the bar, or `--best` to print only the winner.
`text-color` works as an alias for the command.

## fix

`fix` nudges a color until it passes. It walks OKLab lightness in the direction
that increases contrast and leaves chroma and hue alone, so the result still
looks like the color you started with.

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

Use `--target 7` for an explicit ratio, or `--suggest` to control how many
alternatives are listed. A color that already passes is reported and left
alone.

## Input formats

Hex works with or without the hash, in three, four, six or eight digit form:

```
#fff   #ffff   #ffffff   #ffffff80
```

Functional notation accepts commas, spaces, or a slash before the alpha value:

```
rgb(26, 115, 232)      rgb(26 115 232)      rgb(26 115 232 / 0.5)
hsl(214.08 81.75% 50.59%)                   hsla(214.08, 81.75%, 50.59%, 0.5)
hsv(214.08 88.79% 90.98%)                   hwb(214.08 10.2% 9.02%)
lab(48.77 9.69 -67.47)                      lch(48.77 68.16 278.18)
oklab(0.5737 -0.0409 -0.1902)               oklch(0.5737 0.1946 257.86)
```

Angles accept `deg`, `rad`, `grad` or `turn`, and percentages work anywhere CSS
allows them. The CSS color names are matched case insensitively, so
`rebeccapurple`, `RebeccaPurple` and `#663399` all name the same color.
`transparent` is accepted too.

The wide gamut spaces arrive through CSS `color()`, which names its space
first:

```
color(srgb 0.2039 0.5961 0.8588)     color(display-p3 0.3208 0.588 0.8369)
color(a98-rgb 0.3725 0.5905 0.8459)  color(prophoto-rgb 0.4327 0.51 0.7876)
color(rec2020 0.3766 0.5394 0.8141)  color(display-p3 1 0 0 / 0.5)
```

Components take numbers or percentages and are allowed to fall outside 0..1,
which is how CSS names colors the sRGB gamut cannot hold. Retna keeps those
values rather than treating them as mistakes, because a negative component
still names a real color. Clipping only happens when the color has to be drawn
as sRGB.

## Color spaces

```
hex, rgb, rgba, hsl, hsv, hwb, lab, lch, oklab, oklch,
srgb, display-p3, a98-rgb, prophoto-rgb, rec2020
```

Lab, LCH and ProPhoto use the D50 white point, which is what CSS specifies.
OKLab, OKLCH, display-p3, a98-rgb and rec2020 use D65.

Turning an sRGB color into one of the wider spaces cannot land outside their
gamuts, so those components are reported without clipping. A clipped component
would name a different color. Going the other way, a `color()` value outside
the sRGB gamut is clipped to the nearest displayable color.

## JSON output

`--json` works on every command. For contrast:

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

With `--against`, the measurements become a list under `results`. Selecting
`--algorithm apca` adds an `apca` field holding Lc.

`inspect --json` returns each space as a three number group plus `luminance`:

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

Each group is one line here for readability; the command itself indents every
number onto its own line.

Hues are degrees, saturation style channels run from 0 to 1, Lab lightness runs
from 0 to 100, and OKLab lightness runs from 0 to 1. The wide gamut components
are raw, so they can sit outside 0..1.

## Transparency

WCAG and APCA both measure opaque colors. When either input has alpha, Retna
composites the background over white and then the foreground over that
background, which is what a browser would paint, and prints a note so the
substitution is not silent.

## APCA

`--algorithm apca` returns Lc. The sign carries meaning: dark text on a light
background is positive, and light text on a dark background is negative, so
compare absolute values. The built in checks use the 45, 60 and 75 bands.

## License

MIT. See [LICENSE](LICENSE).
