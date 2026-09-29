# Accessibility: What the Numbers Actually Are

Contrast is not a matter of taste, and it is not something to eyeball. It is a ratio you can
compute, and `retna` computes it. This file is the reference for *which* ratio applies to
*which* pair, and why two different algorithms disagree.

## Two algorithms, two jobs

**WCAG 2.x** is the Recommendation. It is what an audit, a procurement review, or a lawsuit
measures. If you are asked "does this pass", this is the answer.

**APCA** (Accessible Perceptual Contrast Algorithm) is a draft for WCAG 3. It models perceived
contrast rather than relative luminance, so it accounts for polarity, text weight, and font
size. It is generally stricter, and it is the better predictor of whether text is actually
comfortable to read.

| | WCAG 2.x | APCA |
| --- | --- | --- |
| Output | ratio, `1:1`–`21:1` | `Lc`, signed, roughly ±108 |
| Polarity | symmetric — swap the colors, same ratio | **asymmetric** — dark-on-light ≠ light-on-dark |
| Relative luminance | WCAG's own definition | a different, softer-clamped luminance |
| Status | the standard | a draft, and a better quality bar |
| `retna` flag | default | `--algorithm apca` |

Both are available at once with `--all`, which is what you want while designing:

```bash
retna contrast "#777" "#fff" --all
```

```
APCA
Lc                    71.11
Lc 90 Body Preferred  FAIL
Lc 75 Body Minimum    FAIL
Lc 60 Content Text    PASS

WCAG
Ratio       4.48:1
AA Normal   FAIL
```

Note the split verdict: `#777` on white fails WCAG AA for body text (4.48 vs 4.5 required) but
clears APCA's *content text* band. This is why "it looks fine" and "it fails the audit" can
both be true.

**Prefer WCAG as the gate and APCA as the floor you'd rather hit.** When a pair passes APCA but
not WCAG, it still fails the standard.

## The thresholds

| Use | Minimum | `retna` check name |
| --- | --- | --- |
| Body text — anything below the large-text size | **4.5:1** | `AA Normal` |
| Large text: ≥24px (18pt), or ≥18.66px (14pt) bold | **3:1** | `AA Large` |
| Body text, AAA | **7:1** | `AAA Normal` |
| Large text, AAA | **4.5:1** | `AAA Large` |
| Icons, borders, focus rings, chart marks, control outlines | **3:1** | measured with `contrast`, no named check |
| Disabled controls, decorative elements, logotypes | exempt | — |

APCA bands, from the reference implementation:

| Lc | Meaning |
| --- | --- |
| 90 | preferred for body text |
| 75 | minimum for body text |
| 60 | content text that is not body text |
| 45 | large or heavy text |
| 30 | the floor for any text at all |
| 15 | the floor for non-text elements |

Compare **absolute** Lc: dark-on-light is positive, light-on-dark negative.

## How rare a passing pair is

A brute-force run over all ~281 trillion hex pairs (mrmrs):

| Threshold | Pairs that pass |
| --- | --- |
| WCAG 3:1 | 26.5% |
| WCAG 4.5:1 | 12.0% |
| WCAG 7:1 | 3.6% |
| APCA 60 | 7.3% |
| APCA 75 | 1.6% |
| APCA 90 | **0.08%** |

Roughly one in eight color pairs clears AA, and one in 1,250 clears APCA 90. Picking colors by
eye is a coin flip against those odds; that is the entire case for running the tool.
See `references/color-expert/contemporary/accessible-color-combinations-count.md` and
`references/color-expert/techniques/apca-myndex-contrast.md`.

## Which pairs to check

Do not spot-check. Build the two lists and cross them:

| Background ↓ / Foreground → | body | muted | subtle | link | border | focus |
| --- | --- | --- | --- | --- | --- | --- |
| page surface | 4.5 | 4.5 | 3 (large only) | 4.5 | 3 if functional | 3 |
| raised surface (card) | 4.5 | 4.5 | 3 | 4.5 | 3 if functional | 3 |
| subtle fill / hover | 4.5 | 4.5 | 3 | 4.5 | — | 3 |
| dark base | 4.5 | 4.5 | 3 | 4.5 | 3 if functional | 3 |
| accent / button fill | 4.5 | — | — | — | — | 3 |

Practical version, one command per surface:

```bash
retna palette contrast --foreground "#18181B" \
  --colors "#FFFFFF,#FAFAFA,#F4F4F5,#E4E4E7" --min 4.5
```

Every color that can be a background is a row; every color that can be text is a `--foreground`
run. The muted text and the small-text-on-subtle-fill pairs are the ones that fail in practice.

## Non-text contrast: the rule people forget

WCAG 1.4.11 requires **3:1 for the visual information needed to identify a control** and its
state: input borders, checkbox and radio outlines, toggle tracks, focus indicators, selected
states, icon buttons, chart marks.

Common measured failures, using ordinary-looking values:

| Pair | Ratio | Verdict |
| --- | --- | --- |
| `#3F3F46` border on `#18181B` dark card | **1.70:1** | invisible |
| `#52525B` border on `#18181B` dark card | 2.29:1 | fails |
| `#E4E4E7` border on `#FAFAFA` light page | 1.22:1 | invisible |
| `#D4D4D8` border on `#FAFAFA` light page | 1.42:1 | fails |
| `#A1A1AA` border on `#FFFFFF` | 2.56:1 | fails |
| `#94949C` border on `#FFFFFF` | **3.01:1** | passes |
| `#6E6E78` border on `#18181B` | **3.51:1** | passes |

Two things follow:

1. **Accessible borders are darker/lighter than instinct.** On light surfaces you need roughly
   the muted-text stop; on dark surfaces you need the "subtle text" stop. Anything that reads
   as a delicate hairline is decorative, and a decorative hairline cannot be the only thing
   distinguishing a control.
2. **Decorative separators are exempt.** A card divider is not "needed to identify a control",
   so 1.2:1 is legal. It *is* a problem if a user must perceive it to know what to click.

Focus rings are the highest-risk case because they must clear 3:1 against *both* the component
and the page behind it:

```bash
retna contrast "#146CDD" "#FFFFFF"    # 4.98:1  ring on light
retna contrast "#146CDD" "#FAFAFA"    # 4.77:1  ring on off-white
retna contrast "#5BA1FF" "#18181B"    # 6.74:1  ring on dark
```

A 2px ring with a 2px offset is the usual shape; the offset is what keeps it off the control's
own fill, where 3:1 is often impossible.

## Color-vision deficiency

CVD removes one of the two opponent axes. Red-green deficiency collapses the red↔green axis;
yellow-blue deficiency collapses the yellow↔blue axis.

- **Orange↔blue survives both.** It is the most accessible pair of hues for categorical data.
- **Never encode meaning in hue alone.** Add luminance separation, a shape, a pattern, or a
  label. A "red/green pass/fail" pair becomes the same color; `✓ / ✕` does not.
- **Luminance is the axis that survives.** Compare the `luminance` field from
  `retna inspect --json`. Two categories at nearly the same luminance merge no matter how
  different their hues are.
- Retna does **not** simulate CVD. It gives you the luminance and ratios that let you avoid the
  collision; use a simulator (Sim Daltonism, the Chrome DevTools rendering emulation) to
  confirm.

See `references/color-expert/contemporary/opponent-process-color-blindness.md`.

## Charts

Requiring 3:1 between every *pair* of fills is n² constraints and becomes impossible past three
categories. Use the border trick instead: give every mark a border and require 3:1 between each
fill and the border — n constraints, and every mark is guaranteed distinguishable from the
background and from its own outline.

Sequential ramps have their own rule: a **flat perceptual derivative**. The perceptual step
between consecutive samples should be constant — a straight line in color and in grayscale.
Check by sampling every stop and comparing the OKLCH lightness deltas:

```bash
retna inspect "#440154" --json | Select-String '"oklch"'
```

A bump in the deltas means the colormap exaggerates change that isn't in the data. This is the
real defect in `jet`, not that it's ugly.

## Transparency

WCAG and APCA both describe opaque colors. When an input has alpha, `retna` composites the
background over white and then the foreground over that — browser paint order — and prints:

```
note: translucent colors were composited over white before measuring
```

The ratio is for the *composited* color, so the hex in the table is the one the numbers belong
to. Text over an image or a gradient is the same problem with no single answer: flatten the
composite you care about — the lightest and darkest regions separately — and measure each.

## Automating the check

```bash
retna contrast "$FG" "$BG" --min 4.5     # exit 1 on failure
retna palette contrast --foreground "$FG" --file pairs.txt --min 4.5
```

Keep the pair list in the repo next to the tokens. A gate that drifts from the stylesheet is
worse than no gate, because it certifies something that isn't what ships.

The exit-code and CI patterns are in
[../references/palette-recipes.md](palette-recipes.md#recipe-7--ci-gate); the CLI details are in
[retna-cli.md](retna-cli.md).

## Related reading in the bundled library

| Topic | File |
| --- | --- |
| APCA, in depth | `references/color-expert/techniques/apca-myndex-contrast.md` |
| How many pairs pass | `references/color-expert/contemporary/accessible-color-combinations-count.md` |
| Accessibility extraction from colorandcontrast.com | `references/color-expert/contemporary/colorandcontrast/accessibility.md` |
| Color-vision deficiency | `references/color-expert/contemporary/opponent-process-color-blindness.md` |
| Lint rules beyond contrast (38 of them) | `references/color-expert/techniques/color-buddy-palette-lint.md` |
| UI design with color | `references/color-expert/contemporary/colorandcontrast/ui-design.md` |
