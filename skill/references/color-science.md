# Color Science

The knowledge half of the skill: which space to think in, what the axes mean, where the tidy
model breaks. `retna` supplies the numbers; this file supplies the reasons.

## Which space, when

| Task | Space | Why |
| --- | --- | --- |
| Building ramps, scales, tokens | **OKLCH** | best uniformity for lightness, chroma and hue; fixes CIELAB's blue problem |
| Gradients and mixing | **OKLAB** / `color-mix(in oklab, …)` | no gray dip in the middle |
| Picking colors interactively | **OKHSL / OKHSV** | cylindrical like HSL, but perceptually grounded |
| Normalized saturation (0–100%) | **HSLuv** | CIELUV chroma normalized per hue and lightness |
| Print workflows | **CIELAB D50** | the ICC standard illuminant |
| Screen workflows | **CIELAB D65** or OKLAB | D65 is the screen white point |
| Cross-media appearance matching | CAM16 / CIECAM02 | accounts for surround, adaptation, luminance |
| HDR | Jzazbz / ICtCp | designed for extended dynamic range |
| Pigment/paint mixing | Kubelka-Munk (Spectral.js, Mixbox) | spectral reflectance, not RGB averaging |
| Precise color difference | CIEDE2000 | the gold standard perceptual distance |
| Fast color difference | Euclidean in OKLAB | good enough for most UI work |
| Dithering, alpha compositing, downsampling | **linearized sRGB** | adjacent emitters add *light*; model the device, not the eye |

What Retna gives you, and the white points it uses:

```
hex, rgb, rgba, hsl, hsv, hwb, lab, lch, oklab, oklch,
srgb, display-p3, a98-rgb, prophoto-rgb, rec2020
```

Lab, LCH and ProPhoto are **D50** (CSS specifies D50 for those); OKLab, OKLCH, display-p3,
a98-rgb and rec2020 are **D65**. `retna inspect` prints all of them at once, which makes it the
quickest way to answer "what is this color, actually".

## OKLCH

Three axes, and only three things you can decide:

- **L** — perceptual lightness, 0–1. This is the axis that controls legibility against a
  surface, and the axis `retna fix` moves.
- **C** — chroma, distance from the neutral axis. Not bounded by 1; bounded by the **gamut**,
  which depends on L and H. `oklch(0.7 0.35 150)` is not a color you can display.
- **H** — hue angle in degrees. Roughly stable across lightness in OKLab, unlike CIELAB where
  a fixed hue bends as L changes.

Because L is perceptual, equal L steps look like equal steps of lightness. That is the entire
reason to build ramps here instead of in HSL.

**Converting an sRGB color into a wider space cannot land outside it.** So `display-p3`,
`a98-rgb`, `prophoto-rgb` and `rec2020` components from `retna` are reported without clipping —
a clipped component would name a different color. The reverse direction does clip.

## Gamut, and what clipping does

Retna always hands back a displayable sRGB color, so an out-of-gamut request is silently
clipped:

```
$ retna convert "oklch(0.7 0.35 150)" --to hex,oklch
    #00D100
HEX         #00D100
OKLCH       oklch(0.745, 0.2535, 142.4953)
```

Asked for L 0.700 / C 0.350 / H 150°; got L 0.745 / C 0.254 / H 142.5°. Every axis moved. This
is what "the browser maps it for you" looks like from the other side — and it is worse in JS,
where a naive `oklch → hex` truncates channels with no mapping at all.

Rules:

- **Reduce chroma, not lightness or hue.** Clipping R/G/B is what shifts the hue; pulling
  chroma toward the gamut boundary preserves the color's identity.
- **Test against the actual target.** A color valid in Display P3 can still clip in sRGB.
- **Or avoid it by construction:** express chroma *relative to the gamut boundary* rather than
  as an absolute number, so generated colors cannot ask for chroma that isn't there.
- **Detect it, since Retna won't:** convert back and compare with what you asked for. If the
  numbers differ, you were out of gamut.

## HSL's three failures

HSL is not "bad" — it is a fast geometric rearrangement of RGB into a cylinder. It is fine for
quick picking and simple tweaks. It is wrong for systems, because its channels do not correspond
to perception:

- **Lightness.** `hsl(60 100% 50%)` (yellow) and `hsl(240 100% 50%)` (blue) both have L=50% and
  have wildly different perceived brightness. L is a mathematical average.
- **Hue.** Non-uniform spacing: 20° near red is a dramatic change, 20° near green is barely
  visible. The green region is compressed and the reds are stretched.
- **Saturation.** Does not correlate with perceived saturation. A color can be S=100% and look
  muted (a dark saturated blue).

When accuracy matters, use OKLCH for scales, OKLAB for gradients, OKHSL for picking, HSLuv for
normalized saturation.

## Four words that are not synonyms

- **Chroma** — colorfulness relative to a same-lightness neutral reference.
- **Saturation** — perceived colorfulness relative to the color's *own* brightness.
- **Lightness** — perceived reflectance relative to a similarly lit white.
- **Brightness** — perceived intensity of light coming from a stimulus.

Same chroma ≠ same saturation; they are different dimensions in different spaces. Lightness is
contextual (relative to illumination); brightness is absolute. This distinction is why "make it
lighter" and "make it brighter" are different instructions.

## Hue names, as degree ranges

For constraining or generating by hue name:

| Name | Degrees | | Name | Degrees |
| --- | --- | --- | --- | --- |
| red | 345–360, 0–15 | | blue | 195–260 |
| orange | 15–45 | | purple | 260–310 |
| yellow | 45–70 | | pink | 310–345 |
| green | 70–165 | | warm | 0–70 |
| cyan | 165–195 | | cool | 165–310 |

## Harmony, honestly

**Hue-first harmony is a weak standalone heuristic.** Complementary, triadic and tetradic
intervals are poor predictors of mood, legibility or accessibility on their own. Every hue plane
has a different shape in perceptual space, so geometric hue intervals do not guarantee perceptual
balance. Complements are also **pigment-specific**: in OKLAB, cadmium red, quinacridone red and
alizarin crimson all sit opposite cobalt teal, and chrome oxide green sits opposite a purple, not
red. Compute the opposite from the actual color (OKLCH hue + 180°), never from its name.

**Character-first harmony works better.** Organize by character — pale / muted / deep / vivid /
dark — rather than by hue. Hue is usually a weaker predictor of emotional response than chroma
and lightness: a muted palette reads as calm across many hues, and "relaxed vs. intense" is
driven more by chroma and lightness than by hue.

**Legibility comes from lightness variation.** Grayscale is a quick sanity check that lightness
separation exists, not an accessibility proof — you still verify with WCAG/APCA. Same character
with varied lightness is often readable; same lightness regardless of hue is usually illegible.

**The 60-30-10 rule.** 60% dominant, 30% secondary, 10% accent. One color dominates so three
colors aren't "three equally-sized gorillas fighting".

**Meaning is per-context, not from a symbolism table.** In narrative work, ask what a color is
*associated* with by repetition and whether it *transitions* as that subject changes. A single
saturated outlier on a balanced scheme is a focal device; a new hue entering a settled scheme
reads as disruption.

## Color temperature

Temperature is **not** hue. It is a systematic shift of hue *and* saturation that depends on the
starting hue, plus a spectral bias (which end of the spectrum a light favors). Cool daylight
fills shadows with blue scatter; paint neutral highlights and blue shadows. Green and purple do
not map cleanly to either pole — perceived temperature for them depends heavily on context.

## Naming colors

| System | Register | Example |
| --- | --- | --- |
| ISCC-NBS | scientific precision | "vivid yellowish green" |
| Munsell | systematic notation | "5GY 7/10" |
| XKCD | common perception | "ugly yellow", "hospital green" |
| RAL | industrial reproducibility | RAL 5002 |
| Ridgway (1912) | ornithological | 1,115 named colors, public domain |
| CSS named colors | web standard | 147 names, matched by Retna |
| color-description | emotional adjectives | "pale, delicate, glistening" |
| COLIBRI | graded / fuzzy | "0.6 cyan · 0.4 light blue" |

**Do not name colors from a screenshot by looking at it.** Vision models — this agent included —
name prototypical high-chroma hues reliably and degrade badly on non-prototypical shades,
near-neutrals, and fine lightness steps; CLIP-style encoders read the *word* "red" written over
blue ink and rarely label white/grey/black at all. Sample the pixels, convert with
`retna inspect --json`, and name the OKLCH triple. Treat a visual impression as a hypothesis.

The same caution applies to *measuring* rather than naming: never read a contrast ratio off an
image. Flatten the composite, then measure.

## Sorting a palette

There is **no single correct linear order** for a set of colors. Color is 3D, so any 1D ordering
is a *path* through a 3D space — closely related to the travelling salesman problem. Naive sorts
fail predictably: sorting by hue alone interleaves lights and darks; sorting by lightness alone
collapses distinct hues; `.sort()` on a packed RGB or HSL integer produces jagged, meaningless
jumps. Treat "put these in a sensible order" as a smoothest-path problem in perceptual space,
not a one-channel sort.

## Historical corrections worth knowing

- **Moses Harris (1769)** was the first to place RYB at equal 120° spacing — Newton and Boutet
  did not. The origin of most bad color theory.
- **"Indigo" was killed by von Bezold (1874).** Newton's "blue" ≈ modern cyan; Newton's "indigo"
  ≈ modern blue.
- **"Magenta" was not the subtractive primary's name until 1907.** Before that: pink, crimson,
  purpur.
- **Amy Sawyer patented a CMY wheel around 1911** — primrose, rose, turquoise — decades before it
  became mainstream.
- **Elizabeth Lewis (1931)** married trichromatic and opponent-process theories on one wheel,
  anticipating CIE Lab by 30 years.

**Easy and accurate are two different things.** The tidy model (RYB primaries, a 12-hue wheel,
"red is opposite green") is easy precisely because it summarizes beliefs rather than
measurements. Give the tidy version, and say where it breaks.

## Pigment mixing, briefly

Pigment mixing is not well described by the simple subtractive model alone. CMY mixing paths
curve outward (vivid secondaries); RGB mixing paths curve inward (dull browns). Mixing is
non-linear — proportion of paint is not proportional hue change, and blue→yellow is a much longer
road than red→yellow. Tinting strength varies with pigment (blues are strong, yellows weak), and
white does not merely lighten: it shifts hue *and* kills chroma.

## Where to go deeper

The bundled library is a 182-file rip of `meodai/skill.color-expert` (CC BY 4.0) at
[`references/color-expert/`](color-expert/INDEX.md). Start from
[`color-expert/INDEX.md`](color-expert/INDEX.md) — it is a lookup table of every file with a
one-line summary and its source.

High-value entries for UI work:

| Topic | File |
| --- | --- |
| OKLab, OKHSV/OKHSL, gamut clipping, blending in linear light | `color-expert/contemporary/bjorn-ottosson-oklab-articles.md` |
| Chroma vs. saturation | `color-expert/contemporary/chroma-vs-saturation.md` |
| Lightness vs. brightness | `color-expert/contemporary/lightness-vs-brightness.md` |
| The full conversion pipeline, HSL's failures | `color-expert/contemporary/your-colors-suck-acerola.md` |
| CIE 1931, XYZ, the standard observer | `color-expert/contemporary/cie-1931-standard-observer.md` |
| Conversion equations and matrices | `color-expert/techniques/brucelindbloom-color-math.md` |
| CSS Color 4/5 syntax and gamut mapping | `color-expert/techniques/w3c-css-color-4-and-5.md` |
| HSLuv, contrast-correct saturation | `color-expert/techniques/hsluv-better-than-hsl.md` |
| Palette generators compared, with algorithms | `color-expert/techniques/rampensau-palette-generation.md`, `color-expert/techniques/cusphanger-gamut-triangle-palettes.md`, `color-expert/techniques/nutelch-gamut-relative-chroma.md` |
| Perceptual palette sorting | `color-expert/techniques/colorsort-js.md` |
| Lint rules for a whole palette | `color-expert/techniques/color-buddy-palette-lint.md` |
| Munsell: hue, value, chroma | `color-expert/historical/munsell-hue-value-chroma.md` |
| Why the RYB wheel is a bad tool | `color-expert/contemporary/color-theory-dogma-problem.md` |
| Color science as the agent's own blind spot | `color-expert/contemporary/colour-in-computer-vision-vlm.md` |
