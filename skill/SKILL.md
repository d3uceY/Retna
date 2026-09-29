---
name: retna-color-expert
description: 'Build and audit accessible UI color systems with measured numbers. Use when choosing, generating, scaling or reviewing a palette; building 50–950 shade ramps; defining theme or design tokens; designing dark mode; picking a text color for a surface; checking WCAG or APCA contrast; converting between hex, rgb, hsl, lab, lch, oklab, oklch and display-p3; or diagnosing muddy gradients, clashing shades, out-of-gamut colors, and "my UI colors look off". Triggers on "build me a color scale", "is this readable", "does it pass AA", "fix this contrast", "what text color goes on this background", "color palette for my app", "dark mode palette". Runs the retna CLI so contrast, lightness and color-space facts are measured, never estimated. Pairs with refactoring-ui (hierarchy, spacing, type) and impeccable (frontend polish).'
license: MIT
metadata:
  author: d3uceY
  version: "1.0.0"
---

# Retna Color Expert

Two assets, used together:

- **`retna`** — a CLI that parses almost any color notation, converts between 15 color
  spaces, and measures WCAG and APCA contrast. It is the oracle. It writes nothing to disk.
- **`references/`** — why the numbers are what they are: perceptual color science, gamut
  behavior, accessibility thresholds, and a 182-file library ripped from the
  [color-expert](https://github.com/meodai/skill.color-expert) skill.

## The one rule

**Never state a contrast ratio, an Lc value, a luminance, or a converted hex from memory.
Run `retna`.**

Agents are bad at this specific thing — vision models name high-chroma hues well and degrade
on near-neutrals and fine lightness steps, and a "looks readable" judgement is not a
measurement. Every number below this line is one command away. Estimate freely; *publish*
only what the tool printed.

```bash
retna contrast "#777777" "#FFFFFF"
retna convert "#1a73e8" --to oklch
retna inspect "#3498db" --json
```

If `retna` is not on PATH, build it next to the project (`go build -o retna .` in the Retna
repo) or use the prebuilt binaries. Check with `retna version`.

## Command map

| What you need | Command |
| --- | --- |
| One foreground on one background | `retna contrast "$FG" "$BG"` |
| One foreground vs. many backgrounds | `retna contrast "$FG" --against "#fff,#000,#f5f5f5"` |
| Fail a build when a pair is unreadable | `retna contrast "$FG" "$BG" --min 4.5` → exit 1 |
| APCA instead of WCAG | `retna contrast "$FG" "$BG" --algorithm apca` |
| Both algorithms, side by side | `retna contrast "$FG" "$BG" --all` |
| A palette from a file or stdin | `retna palette contrast --foreground "#fff" --file palette.txt` |
| A color in every space | `retna convert "#1a73e8"` (or `--to oklch,lab`) |
| Relative luminance | `retna inspect "#3498db"` |
| Which text color belongs on a surface | `retna readable "#3498db"` (`--best` for one) |
| The nearest color that passes | `retna fix "#777777" --background "#fff" --level AA` |
| Anything machine-readable | add `--json` to any command |

`--against` repeats, and each value may be a comma-separated list. `--min` needs the WCAG
algorithm to be selected (`--all` includes it; `--algorithm apca --min 4.5` is an error rather
than a check that silently measures nothing). Full flag list and JSON shapes:
[references/retna-cli.md](references/retna-cli.md).

**Parse with `--json`.** The text output carries ANSI swatches and box-drawing tables meant
for humans. `--json` gives stable keys: `contrast`, `wcag.aa.normal`, `apca`, `oklch`,
`luminance`, `suggestions[]`.

## The build loop

1. **Name the roles before the colors.** surface, on-surface, muted-text, border, accent,
   success, warning, danger, focus. Colors without roles become a swatch collection, not a
   system.
2. **Design hierarchy in grayscale, add color last.** Size, weight and spacing should carry
   the hierarchy so color is decoration and meaning, not a crutch. (This is the
   `refactoring-ui` core principle; the two skills compose.)
3. **Build ramps in OKLCH, not HSL.** Hold hue roughly constant, walk lightness, keep chroma
   inside the gamut. `retna convert <hex> --to oklch` tells you what you actually made.
4. **Verify every text-on-surface pair with `retna contrast`.** Both polarities, both themes.
   Not one representative pair — every pair that ships in the token file.
5. **Repair with `retna fix`.** It walks OKLab lightness toward the target and leaves chroma
   and hue alone, so the fixed color still looks like the color you started with. The first
   suggestion is the smallest change that passes.
6. **Emit the token graph:** reference tokens (palette) → semantic tokens (roles) → component
   usage. Components consume semantic names, never raw hex.

## Building a ramp

A ramp is 9–11 stops, near-white to near-black. What makes it feel designed:

- **Lightness steps should be even, not equal.** OKLCH L is already perceptual, so aim for
  roughly even L deltas between neighbours; the light end needs smaller deltas than the dark
  end to feel even.
- **Chroma peaks in the middle**, at full saturation around the 500–600 stops, and tapers at
  both ends. A ramp with flat chroma looks like a fluorescent tube.
- **A template to start from** (then measure, don't trust): L ≈ 0.97 / 0.94 / 0.88 / 0.80 /
  0.71 / 0.62 / 0.55 / 0.48 / 0.40 / 0.31 / 0.24 for 50→950, with chroma rising to about
  0.19 around the 500–600 stops and tapering to 0.02 at 50 and 0.07 at 950 (see
  [references/palette-recipes.md](references/palette-recipes.md) for the same ramp worked out
  in full).
- **Don't let the darkest stop be pure black** (`#000000`) — halation, and no headroom for
  dark-mode elevation. The darkest is usually around L 0.20–0.28.
- **Verify the endpoints, not just the middle.** The stops people actually ship are text-600
  on surface-50 and text-300 on surface-900.

```bash
# check a stop you just invented
retna inspect "oklch(0.55 0.16 255)" --json
# does the pale stop still hold body text against the dark stop?
retna contrast "#1f2937" "#f9fafb" --min 4.5
```

**Out-of-gamut requests are silently clipped.** Ask for more chroma than sRGB holds and Retna
returns the nearest displayable color — with a *different* chroma and hue than you asked for:

```bash
retna convert "oklch(0.7 0.35 150)" --to hex,oklch
# → #00D100, oklch(0.745, 0.2535, 142.4953)
```

The L moved, the chroma dropped by a third, and the hue rotated 7.5°. Always convert back and
compare against what you requested. Reduce chroma rather than clipping lightness or hue, since
clipping RGB is what shifts the hue in the first place. Details:
[references/color-science.md](references/color-science.md).

## Contrast minimums

| Rule | Ratio | `retna` name |
| --- | --- | --- |
| WCAG AA, normal text | 4.5:1 | `AA Normal` |
| WCAG AA, large text (≥24px, or ≥18.66px bold) | 3:1 | `AA Large` |
| WCAG AAA, normal text | 7:1 | `AAA Normal` |
| WCAG AAA, large text | 4.5:1 | `AAA Large` |
| Non-text UI (icons, borders, focus rings) | 3:1 | measure with `contrast` |

APCA (`--algorithm apca`) returns Lc instead, and it is **polarity aware**: dark text on light
is positive, light text on dark is negative. Compare absolute values. Bands:
90 body-preferred, 75 body-minimum, 60 content text, 45 large/heavy text, 30 text floor,
15 non-text floor.

- **WCAG for compliance, APCA for reading comfort.** WCAG 2.x is the standard; APCA is a
  draft for WCAG 3 and is markedly stricter (about 12% of hex pairs clear AA, ~1.6% clear
  APCA 75).
- When a design passes APCA but not WCAG, it fails the standard — ship the WCAG-passing value
  and treat APCA as the quality bar.
- Translucent inputs are composited over white before measuring, and Retna prints a note so
  the substitution is never silent.
- CI gate: `retna contrast "$FG" "$BG" --min 4.5` prints `FAIL  4.48:1 / Required: 4.50:1`
  and exits 1; with `--against` it reports how many pairs failed.

Numbers, the AA banner, non-text contrast, and why "just check black and white" is not enough:
[references/accessibility.md](references/accessibility.md).

## Dark mode

Dark mode is not inverted light mode. Values that pass in light mode are not guaranteed to
pass when flipped, so **re-measure the dark pair set separately** — this is the single most
common source of shipped contrast bugs.

- Background is a dark gray (`#09090b`–`#18181b`), not `#000000`.
- Body text is off-white (`~#fafafa`), not `#ffffff`; max contrast is harsher in the dark.
- Saturated brand colors get lighter and *less* saturated, or they vibrate.
- Elevation is *lighter* surfaces, not bigger shadows.
- Verify the muted/secondary text you just brightened, and verify it again on every surface it
  can land on (base, card, hover, input).

Full recipe and token scales: [references/dark-mode.md](references/dark-mode.md).

## Color science shortcuts

- **OKLCH for scales, OKLAB for gradients, `color-mix(in oklab, …)` in CSS.** Both avoid the
  gray dip in the middle of an RGB/HSL gradient.
- **HSL is a geometric rearrangement of RGB, not a model of perception.** `hsl(60 100% 50%)`
  and `hsl(240 100% 50%)` share L=50% and look nothing alike in brightness; 20° of hue shift
  is dramatic near red and invisible near green. Fine for quick picking, wrong for systems.
- **Chroma ≠ saturation.** Chroma is distance from the neutral axis; saturation is relative to
  the color's own brightness. They are different dimensions and different spaces.
- **Warm/cool is not a hue rotation.** It's a systematic shift of hue *and* saturation, and
  green and purple don't map cleanly to either pole.
- **Hue-first harmony is a weak heuristic.** Complementary/triadic intervals don't predict
  mood or legibility. Character (pale / muted / deep / vivid / dark) predicts better, and
  chroma + lightness drive "calm" vs "intense" more than hue does.
- **Sorting a palette has no single correct order.** It's a path through 3D space; a
  single-channel `.sort()` produces jagged jumps.
- **CVD:** don't encode meaning in hue alone. Orange↔blue is the most accessible pair (it
  survives both red-green and yellow-blue deficiency).

Deeper: [references/color-science.md](references/color-science.md) and the rip of the
color-expert library in [references/color-expert/INDEX.md](references/color-expert/INDEX.md).

## Charts and data viz

Finding three chart colors that each hold 3:1 against *each other* is already very hard, and
beyond three it is essentially impossible. Don't try. Put a **border** on chart elements and
require 3:1 between each fill and the border color — one constraint per color instead of N².

Sequential ramps must have a **flat perceptual derivative**: the step size between consecutive
samples should be a horizontal line, in color and in grayscale. Bumps exaggerate change that
isn't in the data. Check endpoints and the middle with `retna contrast` and `retna inspect`
before shipping a colormap.

## Verifying your work

Before calling a palette done, every one of these should be a command you actually ran:

| Check | Command |
| --- | --- |
| Every body-text pair clears 4.5:1 | `retna contrast "$TEXT" "$SURFACE" --min 4.5` |
| Every large-text pair clears 3:1 | `--min 3` on the display/heading pairs |
| Borders and focus rings clear 3:1 | `retna contrast "$BORDER" "$SURFACE"` |
| Dark mode re-checked, not assumed | run the whole pair list again |
| Ramp chroma stayed in gamut | `retna convert "$HEX" --to oklch` vs. what you asked for |
| No pure black or pure white in the ramp | `retna inspect` the endpoints |
| The palette survives grayscale | squint test — hue must not be doing hierarchy's job |

## References

- [references/retna-cli.md](references/retna-cli.md) — every command, flag, JSON shape, exit code, input format
- [references/palette-recipes.md](references/palette-recipes.md) — worked recipes: brand ramp, semantic tokens, dark theme, chart palette
- [references/accessibility.md](references/accessibility.md) — WCAG vs APCA, the real numbers, non-text contrast, CI
- [references/dark-mode.md](references/dark-mode.md) — light/dark token scales and the mistakes that cause dark-mode contrast bugs
- [references/color-science.md](references/color-science.md) — spaces, gamut behavior, HSL's failures, harmony, naming
- [references/color-expert/INDEX.md](references/color-expert/INDEX.md) — 182 files of color science, theory and technique (ripped from `meodai/skill.color-expert`, CC BY 4.0)
- [ATTRIBUTION.md](ATTRIBUTION.md) — what came from where

Related skills: `refactoring-ui` (grayscale-first hierarchy, spacing, type — the non-color half
of a good UI), `impeccable` (frontend polish, motion, states), `top-design` (dramatic
composition).
