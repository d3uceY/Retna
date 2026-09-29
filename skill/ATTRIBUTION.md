# Attribution

This skill mixes original text, a ripped reference library, and verified tool output. Here is
which is which.

## Original to this skill (MIT)

Written for this skill, licensed MIT:

- `SKILL.md`
- `references/retna-cli.md`
- `references/palette-recipes.md`
- `references/accessibility.md`
- `references/dark-mode.md`
- `references/color-science.md`
- `ATTRIBUTION.md`

Every contrast ratio, Lc value, hex, OKLCH triple and command output quoted in those files was
produced by running **Retna v0.1.0**, built from
[`d3uceY/Retna`](https://github.com/d3uceY/Retna) (MIT). The CLI surface was cross-checked
against the command source in `cmd/` rather than only the README, and the numbers are
reproducible with the commands shown next to them.

## `references/color-expert/` — ripped from `meodai/skill.color-expert`

The 182-file reference library is copied verbatim from
[**meodai/skill.color-expert**](https://github.com/meodai/skill.color-expert).

- **Original project materials** in that repo (its skill framing, `references/INDEX.md`, and the
  notes written specifically for it) are **CC BY 4.0**. The license text is copied here as
  [`LICENSE-COLOR-EXPERT.md`](LICENSE-COLOR-EXPERT.md).
- **Third-party material** under `references/` — video transcripts, talk notes, article and
  paper summaries, scraped sites, tool documentation — remains subject to the original authors'
  rights and licenses. Inclusion does not relicense it. The upstream notice is copied here as
  [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

Each reference file carries its own source attribution; check the original before reusing
anything from it. The upstream `references/INDEX.md` is a full lookup table with a one-line
summary and source link per file, and it is the place to start.

What was **not** copied: the upstream `SKILL.md`, `README.md`, `MAINTENANCE.md`, `ROADMAP.md`,
`CLAUDE.md`, `SECURITY.md` and `evals/`. The frontmatter of this skill is new, so that the
description triggers on UI palette and accessibility work rather than on general color theory.

## Design inspiration

The structure, the "core principle first, then product applications, then CSS patterns" shape,
the scoring-style verification table, and the practice of shipping deep dives as separate
`references/*.md` files are modelled on the **refactoring-ui** skill (MIT, by wondelai) — in
particular its color and theming sections. This skill's take on that material is the measurement
loop: where `refactoring-ui` says "use gray-700 on white, not lighter grays", this one says run
`retna contrast` and ship the value that passes.

## Trademarks and names

Product names (Tailwind, Radix, Leonardo, Culori, Spectral.js, Mixbox, Coolors, APCA, WCAG and
the rest) belong to their owners and are used descriptively. No affiliation is implied.
