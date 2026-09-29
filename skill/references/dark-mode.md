# Dark Mode

Dark mode is a second theme, not a photographic negative. Every value that passed in light
mode is a *candidate* in dark mode, and every pair has to be measured again.

The reason is mechanical: WCAG's ratio is `(hi+0.05)/(lo+0.05)` over relative luminance, and
the mapping from "a color" to "its luminance" is one-directional. Inverting a hex does not
produce the color you intended, and the luminance of the result is not the inverse of the
luminance you had. `#71717A` on `#FAFAFA` is 4.63:1; the naive swap, `#FAFAFA` on `#71717A`,
is a different design entirely.

## The token scales

Two neutral ramps and the values that were verified with `retna` (WCAG ratio, then APCA Lc).

**Light theme**

| Role | Value | On | Ratio | Lc |
| --- | --- | --- | --- | --- |
| `--surface` | `#FAFAFA` | | | |
| `--surface-raised` | `#FFFFFF` | | | |
| `--surface-sunken` | `#F4F4F5` | | | |
| `--text` | `#18181B` | `#FAFAFA` | 16.97:1 | +101.5 |
| `--text-muted` | `#71717A` | `#FAFAFA` | 4.63:1 | +70.6 |
| `--text-subtle` | `#A1A1AA` | `#FAFAFA` | 2.46:1 ✗ | +47.2 |
| `--border` (decorative) | `#E4E4E7` | `#FAFAFA` | 1.22:1 | +10.4 |
| `--border-strong` (functional) | `#94949C` | `#FFFFFF` | 3.01:1 | +19.6 |
| `--focus-ring` | `#146CDD` | `#FFFFFF` | 4.98:1 | — |

**Dark theme**

| Role | Value | On | Ratio | Lc |
| --- | --- | --- | --- | --- |
| `--surface` | `#09090B` | | | |
| `--surface-raised` | `#18181B` | | | |
| `--surface-sunken` | `#27272A` | | | |
| `--text` | `#FAFAFA` | `#09090B` | 19.06:1 | −104.5 |
| | | `#18181B` | 16.97:1 | −103.4 |
| | | `#27272A` | 14.27:1 | −101.3 |
| `--text-muted` | `#A1A1AA` | `#09090B` | 7.76:1 | −51.6 |
| | | `#18181B` | 6.91:1 | −50.6 |
| | | `#27272A` | 5.81:1 | −48.4 |
| `--text-subtle` | `#71717A` | `#18181B` | 3.67:1 (large only) | −27.0 |
| `--border` (decorative) | `#3F3F46` | `#18181B` | 1.70:1 | 0.0 |
| `--border-strong` (functional) | `#6E6E78` | `#18181B` | 3.51:1 | −14.1 |
| `--focus-ring` | `#5BA1FF` | `#18181B` | 6.74:1 | — |

Four things to read off this table:

1. **The same nominal role needs a different value.** `--text-muted` is `#71717A` in light mode
   and `#A1A1AA` in dark mode — *lighter*, not inverted.
2. **Polarity flips the Lc sign.** `#FAFAFA`-ish text is Lc +101 on light and −104 on dark.
   Always compare absolute values.
3. **Dark-mode contrast can be *higher* than light mode for the same pair**, which is exactly
   why you should bring it down: 19.06:1 is harsher than 16.97:1 to read.
4. **`--text-subtle` at 3.67:1 is a trap.** WCAG passes it for large text, and APCA says don't
   use it as text at all (Lc 27, below the Lc 30 floor). If it is decoration or a large label,
   fine; if it is a caption, it needs to move up the ramp.

## Rules, and the reason for each

**Use a dark gray, not black.** `#09090B`–`#18181B`, never `#000000`. Pure black on an OLED
panel next to off-white text produces halation — the text appears to glow and bleed — and it
destroys the headroom you need to show elevation.

**Reduce maximum contrast.** Bring body text down to roughly 16:1 rather than 21:1 by using
`#FAFAFA` instead of `#FFFFFF`. Maximum contrast is fatiguing in low light, which is the
environment dark mode exists for.

**Convey elevation with lighter surfaces, not bigger shadows.** Shadows are nearly invisible
against a dark background. `#09090B` → `#18181B` → `#27272A` for base, card, and input reads as
"above" far more reliably than `box-shadow` does.

**Desaturate saturated colors.** A brand color at full chroma vibrates on a dark surface. Move
up the ramp *and* pull chroma down:

```bash
# light mode accent — white label on it measures 4.98:1
retna convert "#146CDD" --to oklch
# OKLCH  oklch(0.5498, 0.1897, 257.81)

# dark mode accent — same hue, lighter, less chroma
retna convert "#5BA1FF" --to oklch
# OKLCH  oklch(0.7058, 0.1556, 256.39)
retna contrast "#5BA1FF" "#18181B"     # 6.74:1, Lc −49.8
```

L 0.55 → 0.71 and C 0.19 → 0.16 for the same hue. That is "the same brand color, in dark
mode".

Note the Lc on the last line: **−49.8** clears WCAG 4.5:1 comfortably but does not reach
APCA's 60 content-text band. As a focus ring or an icon it is fine (non-text needs Lc 15); if
the design uses that accent for *text* on the dark surface, it should step lighter again.

**Functional borders need to be much lighter than in light mode.** This is the single most
common dark-mode accessibility bug:

| Border | On `#18181B` | Verdict |
| --- | --- | --- |
| `#3F3F46` | 1.70:1 | invisible |
| `#52525B` | 2.29:1 | fails 3:1 |
| `#6E6E78` | 3.51:1 | passes |

A divider you can barely see in light mode is *legally fine*; the same instinct in dark mode
will make an input outline disappear.

**Re-check every surface, not one representative surface.** Muted text that passes on the base
surface can fail on a raised one:

| | `#09090B` | `#18181B` | `#27272A` |
| --- | --- | --- | --- |
| `#FAFAFA` | 19.06 | 16.97 | 14.27 | 
| `#A1A1AA` | 7.76 | 6.91 | 5.81 |

Both pass here. Now try the same table with a subtle text color of `#71717A`: 3.67:1 on the
card, and worse on the input. That is the failure mode — a token verified on one surface and
shipped on three.

## Verifying a dark theme

Run the whole matrix, not a sample:

```bash
# text roles against every surface they can land on
retna contrast "#FAFAFA" --against "#09090B,#18181B,#27272A" --min 4.5
retna contrast "#A1A1AA" --against "#09090B,#18181B,#27272A" --min 4.5
retna contrast "#71717A" --against "#09090B,#18181B,#27272A" --min 4.5

# functional borders and the focus ring
retna contrast "#6E6E78" "#18181B" --min 3
retna contrast "#5BA1FF" "#18181B" --min 3

# accents used as text or as fills with labels
retna contrast "#5BA1FF" "#18181B" --min 4.5
retna contrast "#18181B" "#5BA1FF" --min 4.5
```

Then confirm `prefers-color-scheme` and a manual toggle both land on the same tokens, and
that the toggle itself is visible in both themes.

## Implementation notes

- `light-dark(white, black)` pairs a light and dark value in one declaration (`color-scheme:
  light dark` required). Handy for one-offs, but a token layer is what keeps a large system
  auditable.
- `color-scheme: dark` on `:root` fixes form controls, scrollbars and the default canvas
  background — otherwise a native `<select>` renders light-on-light.
- Don't compute the dark value by inverting the light one at runtime. Inversion is not
  perceptually meaningful; measure both.
- Native `color-mix()` and relative color syntax can derive hover/active states from a base
  token in a chosen space: `oklch(from var(--accent) calc(l * 0.9) c h)`. Derive; don't hand-pick
  a second unrelated hex.

## Common mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| `#000000` background | halation, no elevation headroom | `#09090B`–`#18181B` |
| `#FFFFFF` body text | maximum contrast is harsh in the dark | `#FAFAFA` |
| Same accent hex in both themes | too dark to read on a dark surface | move up the ramp, reduce chroma |
| Elevation via shadow | shadows are invisible | lighter surfaces |
| Muted text kept identical | fails on raised surfaces | re-measure per surface |
| Borders carried over | 1.7:1 is invisible | lighten to clear 3:1 |
| Only one pair verified | the unverified surface fails | check the whole matrix |
| Pure inversion | luminance doesn't invert | re-derive, re-measure |

## Related reading in the bundled library

| Topic | File |
| --- | --- |
| OKLab, gamut clipping, blending in linear light | `references/color-expert/contemporary/bjorn-ottosson-oklab-articles.md` |
| Lightness vs. brightness (why "lighter" is contextual) | `references/color-expert/contemporary/lightness-vs-brightness.md` |
| CSS Color 4/5 syntax, `light-dark()`, relative colors | `references/color-expert/techniques/w3c-css-color-4-and-5.md` |
| Design-token graphs and reactive theming | `references/color-expert/techniques/designbook-reactive-design-token-spec.md` |
| Munsell's value axis, the original systematic lightness scale | `references/color-expert/historical/munsell-hue-value-chroma.md` |
