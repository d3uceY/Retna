# Retna CLI Reference

Everything below was run against Retna v0.1.0. `retna` writes nothing to disk and reads no
config file: arguments in, stdout out.

```
retna [command]

  contrast    Measure the contrast between a foreground and a background
  convert     Convert a color into other color spaces
  fix         Find the nearest color that meets a contrast target
  inspect     Show every form of a color and its relative luminance
  palette     Work with a list of colors
  readable    Suggest text colors for a background   (alias: text-color)
  version     Print the Retna version
```

Add `--json` to `contrast`, `convert`, `inspect`, `palette contrast`, `readable` and `fix`.

## Exit codes

| Situation | Code |
| --- | --- |
| Worked, or `--min` not supplied | 0 |
| `contrast --min` / `palette contrast --min` found a pair below the minimum | 1 |
| Bad color, bad flag, unknown space, apca + `--min` | 1, with `error: …` on stderr |

A failed check prints its verdict and exits 1 without dumping the usage block, which is what
makes it usable as a CI step.

## contrast

```
retna contrast <foreground> [background]
```

One background as the second argument, or many via `--against`:

```
retna contrast "#777" "#fff"
retna contrast "#777" --against "#fff,#000,#f5f5f5,#101010"
retna contrast "#777" --against "#fff" --against "#000,#101010"
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

With more than one background the output becomes a table with a `Ratio`, `AA` and `AAA`
column; when `--all` or `--algorithm apca` is selected the table gains an `Lc` column.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-a, --algorithm` | `wcag` | `wcag` or `apca` |
| `--all` | off | run every algorithm and print each in turn |
| `--against` | — | backgrounds, repeatable, each may be a comma-separated list |
| `--min` | — | minimum WCAG ratio; exits 1 when any pair falls below |
| `--json` | off | JSON instead of a table |

```
$ retna contrast "#777" --against "#fff,#000" --min 4.5
Contrast

Foreground      #777777

Color       Ratio   AA    AAA
──────────  ──────  ────  ────
   #FFFFFF  4.48:1  FAIL  FAIL
   #000000  4.69:1  PASS  FAIL

FAIL  1 of 2 below 4.50:1        # exit code 1
```

A passing run prints `PASS  4.69:1` / `Required: 4.50:1` for a single pair, or
`PASS  all 2 pairs meet 4.50:1` for a list.

**`--min` is a WCAG ratio, so it requires the WCAG algorithm.** `--algorithm apca --min 4.5`
is rejected rather than quietly comparing nothing. `--min 0` is rejected too (it would pass
anything), and so is a non-finite value.

### APCA

```
$ retna contrast "#777" "#fff" --algorithm apca

APCA
Lc                    71.11
Lc 90 Body Preferred  FAIL
Lc 75 Body Minimum    FAIL
Lc 60 Content Text    PASS
Lc 45 Large Text      PASS
Lc 30 Text Floor      PASS
Lc 15 Non-text        PASS
```

Lc is signed: positive for dark-on-light, negative for light-on-dark. Compare absolute
values. The band names come from the APCA 0.1.9 reference implementation.

### JSON

```json
{
  "foreground": "#777777",
  "background": "#FFFFFF",
  "contrast": 4.478089453577214,
  "wcag": { "aa": { "normal": false, "large": true },
            "aaa": { "normal": false, "large": false } }
}
```

- `--against` produces `{ "foreground": …, "results": [ … ] }`.
- `--algorithm apca` adds `"apca": 71.11` and drops nothing else.
- `--all` produces both `wcag` and `apca` blocks.

## palette contrast

```
retna palette contrast --foreground <color> (--colors <list> | --file <path|->)
```

```
retna palette contrast --foreground white --colors "#000,#111,#666"
retna palette contrast --foreground white --file palette.txt
retna palette contrast --foreground white --file - < palette.txt
```

Output is identical to `contrast --against`. The file holds one color per line and each line
may itself be a comma-separated list. Blank lines and lines starting with `//` are skipped,
and a UTF-8 BOM at the start of the file is stripped, so a file saved by a Windows editor
works as-is. `--algorithm`, `--all`, `--min` and `--json` behave as in `contrast`.

This is the command to reach for when a palette already exists in a stylesheet, a token file,
or the output of a generator — pipe it in instead of retyping hexes.

## convert

```
retna convert <color> [--to <space>[,<space>…] | --to all]
```

```
retna convert "#1a73e8" --to hsl,oklch

    #1A73E8

HSL    hsl(214.08, 81.75%, 50.59%)
OKLCH  oklch(0.5737, 0.1946, 257.86)
```

With no `--to`, every space is printed. Unknown spaces fail loudly and list the valid ones:

```
error: unknown color space "cmyk" (want one of hex, rgb, rgba, hsl, hsv, hwb,
lab, lch, oklab, oklch, srgb, display-p3, a98-rgb, prophoto-rgb, rec2020)
```

JSON: `{ "input": "#1A73E8", "values": { "oklch": "oklch(0.5737, 0.1946, 257.86)" } }`

## inspect

```
retna inspect "#3498db"
```

Prints every space in `color.Spaces` order — `HEX, RGB, RGBA, HSL, HSV, HWB, LAB, LCH, OKLAB,
OKLCH, SRGB, DISPLAY-P3, A98-RGB, PROPHOTO-RGB, REC2020` — then the relative luminance.

`--json` returns the same data as numbers:

```json
{
  "input": "#3498db", "hex": "#3498DB", "alpha": 1, "rgb": [52, 152, 219],
  "hsl": [204.0719, 0.6987, 0.5314], "hsv": [204.0719, 0.7626, 0.8588],
  "hwb": [204.0719, 0.2039, 0.1412], "lab": [59.4961, -12.1034, -43.1352],
  "lch": [59.4961, 44.8011, 254.3263], "oklab": [0.6531, -0.0618, -0.1197],
  "oklch": [0.6531, 0.1347, 242.6867], "srgb": [0.2039, 0.5961, 0.8588],
  "display-p3": [0.3208, 0.588, 0.8369], "a98-rgb": [0.3725, 0.5905, 0.8459],
  "prophoto-rgb": [0.4327, 0.51, 0.7876], "rec2020": [0.3766, 0.5394, 0.8141],
  "luminance": 0.283
}
```

Units: hues are degrees, saturation-style channels are 0–1, Lab lightness is 0–100, OKLab
lightness is 0–1. Luminance is the WCAG relative luminance.

Use this to answer "what did I actually just create?" — especially the OKLCH triple after
authoring in another space, or the luminance when you need to reason about why two colors
measure the way they do.

## readable

```
retna readable <background> [--candidates <c1,c2,…>] [--level AA|AAA] [--best]
```

Candidates default to black and white.

```
$ retna readable "#3498db"

    Background #3498DB

Text color  Ratio   AA
──────────  ──────  ────
   #000000  6.66:1  PASS
   #FFFFFF  3.15:1  FAIL
```

`--best` prints only the winner in a small key/value block; `--json` returns
`{ background, level, minimum, candidates: [{ color, contrast, pass }] }`.

Use `--candidates` when the design already has an ink ramp and you want to know *which* ink
belongs on a given surface, rather than introducing pure black or white.

## fix

```
retna fix <color> --background <color> [--level AA|AAA | --target <ratio>] [--suggest N]
```

Walks OKLab lightness in the direction that increases contrast, holds chroma and hue, and
snaps each candidate to 8-bit channels before measuring, so the hex it prints is the hex that
actually meets the ratio.

```
$ retna fix "#777777" --background "#ffffff" --target 7 --suggest 2

Fix

Original       #777777  4.48:1  FAIL
Background     #FFFFFF
Target                  7.00:1

Suggested colors
Color       Ratio   WCAG  Note
──────────  ──────  ────  ───────────────
   #595959  7.00:1  PASS  smallest change
   #585858  7.11:1  PASS
```

- A color that already passes is reported and left alone.
- If nothing at that chroma and hue can reach the target, it says so and suggests lowering
  saturation or changing hue — the honest answer, not a wrong hex.
- Constraints: target between 1:1 and 21:1 (black on white is the ceiling, so >21 is
  rejected as unreachable rather than unfixable), `--suggest` ≥ 1, `--background` required.
- JSON: `{ input, background, target, contrast, pass, suggestions: [{ hex, contrast, pass }] }`.

**Limits:** `fix` repairs one color against one background. It does not understand ramps,
themes or token graphs — run it once per failing pair, and prefer fixing the ramp stop over
fixing the individual component usage.

## Input formats

```
#fff   #ffff   #ffffff   #ffffff80
rgb(26, 115, 232)   rgb(26 115 232)   rgb(26 115 232 / 0.5)
hsl(214.08 81.75% 50.59%)   hsla(214.08, 81.75%, 50.59%, 0.5)
hsv(214.08 88.79% 90.98%)   hwb(214.08 10.2% 9.02%)
lab(48.77 9.69 -67.47)   lch(48.77 68.16 278.18)
oklab(0.5737 -0.0409 -0.1902)   oklch(0.5737 0.1946 257.86)
color(srgb 0.2039 0.5961 0.8588)      color(display-p3 0.3208 0.588 0.8369)
color(a98-rgb 0.3725 0.5905 0.8459)   color(prophoto-rgb 0.4327 0.51 0.7876)
color(rec2020 0.3766 0.5394 0.8141)   color(display-p3 1 0 0 / 0.5)
```

- Commas, spaces, or a slash before alpha all work in functional notation.
- Angles take `deg`, `rad`, `grad` or `turn`; percentages work anywhere CSS allows them.
- Hex works with or without `#`, in 3, 4, 6 or 8 digit form.
- The 147 CSS color names match case-insensitively (`rebeccapurple`, `RebeccaPurple` and
  `#663399` are the same color). `transparent` is accepted.
- Components outside 0..1 are accepted, not treated as mistakes — that's how CSS names colors
  sRGB cannot hold.

## Two behaviors to know before you quote a number

**1. Out-of-gamut input is clipped to sRGB.** `color(display-p3 1 0 0)` becomes `#FF0000`:
Retna always works with a displayable sRGB color, so the OKLCH you read back is the OKLCH of
the clipped color, not of what you asked for.

```
$ retna convert "oklch(0.7 0.35 150)" --to hex,oklch
    #00D100
HEX         #00D100
OKLCH       oklch(0.745, 0.2535, 142.4953)
```

Asked for L 0.700 / C 0.350 / H 150°, got L 0.745 / C 0.254 / H 142.5°. The chroma dropped a
third and the hue rotated. **Always convert back and compare** when authoring in a wide space.

**2. Alpha is composited over white, then the foreground over that.** `#7777` (α ≈ 0.47) on
white measures as `#C0C0C0` at 1.83:1, and the run prints

```
note: translucent colors were composited over white before measuring
```

WCAG and APCA both describe opaque colors; this is the browser's paint order standing in for
a "real" color, and the note exists so the substitution is never silent.

## Repo facts

- Built with Go 1.25+, cobra for the command tree. Math authority is CSS Color 4, including
  the §19 sample code for matrices and white points.
- `go test ./... -count=1` from the Retna repo root runs the regression suite; `go vet ./...`
  is clean.
- Lab, LCH and ProPhoto use D50; OKLab, OKLCH, display-p3, a98-rgb and rec2020 use D65.
- Tests, golden files and `--help` output are the source of truth. When this document and the
  binary disagree, the binary wins — re-run the command.
