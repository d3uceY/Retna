package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

type readableOptions struct {
	level      string
	candidates []string
	best       bool
	json       bool
}

func newReadableCmd() *cobra.Command {
	opts := readableOptions{}
	cmd := &cobra.Command{
		Use:     "readable <background>",
		Aliases: []string{"text-color"},
		Short:   "Suggest text colors for a background",
		Long: `Rank candidate text colors by how well they read on a background.

Candidates default to black and white, which is what you want when picking a
text color for a surface. Pass --candidates to try your own.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReadable(cmd.OutOrStdout(), args[0], opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.level, "level", "AA", "target WCAG level: AA or AAA")
	f.StringSliceVar(&opts.candidates, "candidates", nil, "candidate text colors, defaults to black and white")
	f.BoolVar(&opts.best, "best", false, "print only the highest contrast candidate")
	f.BoolVar(&opts.json, "json", false, "print JSON instead of a table")
	return cmd
}

func runReadable(w io.Writer, background string, opts readableOptions) error {
	bg, err := color.Parse(background)
	if err != nil {
		return fmt.Errorf("background: %w", err)
	}
	minimum, err := levelMin(opts.level)
	if err != nil {
		return err
	}

	// Composite the background over white up front, the same way the
	// measurement does, so the color shown is the one the ratios belong to.
	composited := bg.A < 1
	if composited {
		bg = bg.Over(color.RGB(1, 1, 1))
	}

	candidates := opts.candidates
	if len(candidates) == 0 {
		candidates = []string{"#000000", "#FFFFFF"}
	}

	type scored struct {
		color color.Color
		hex   string
		ratio float64
	}
	scores := make([]scored, 0, len(candidates))
	for _, candidate := range candidates {
		fg, err := color.Parse(candidate)
		if err != nil {
			return fmt.Errorf("candidate %q: %w", candidate, err)
		}
		if fg.A < 1 {
			composited = true
		}
		flatFG, flatBG := flatten(fg, bg)
		scores = append(scores, scored{
			color: flatFG,
			hex:   flatFG.Hex(),
			ratio: contrast.WCAG{}.Calculate(flatFG, flatBG).Value,
		})
	}
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].ratio > scores[j].ratio })
	if opts.best && len(scores) > 1 {
		scores = scores[:1]
	}

	if opts.json {
		view := output.ReadableJSON{
			Background: bg.Hex(),
			Level:      strings.ToUpper(opts.level),
			Minimum:    minimum,
			Candidates: make([]output.CandidateJSON, 0, len(scores)),
		}
		for _, s := range scores {
			view.Candidates = append(view.Candidates, output.CandidateJSON{
				Color:    s.hex,
				Contrast: s.ratio,
				Pass:     s.ratio >= minimum,
			})
		}
		return output.WriteJSON(w, view)
	}

	if opts.best {
		best := scores[0]
		fmt.Fprintf(w, "%s  %s\n", output.Heading("Best contrast"), best.hex)
		fmt.Fprint(w, output.Columns([][]string{
			{"Contrast", formatRatio(best.ratio)},
			{"Level", strings.ToUpper(opts.level)},
			{"Verdict", output.Verdict(best.ratio >= minimum)},
		}))
		noteCompositing(w, composited)
		return nil
	}

	fmt.Fprintf(w, "%s  %s\n\n", output.Swatch(bg), output.Heading("Background "+bg.Hex()))
	rows := make([][]string, 0, len(scores))
	for _, s := range scores {
		rows = append(rows, []string{
			output.Swatch(s.color) + " " + s.hex,
			formatRatio(s.ratio),
			output.Verdict(s.ratio >= minimum),
		})
	}
	fmt.Fprint(w, output.Table([]string{"Text color", "Ratio", strings.ToUpper(opts.level)}, rows))
	noteCompositing(w, composited)
	return nil
}

// levelMin maps a WCAG level name to the ratio it needs for normal text.
func levelMin(level string) (float64, error) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "AA":
		return 4.5, nil
	case "AAA":
		return 7.0, nil
	}
	return 0, fmt.Errorf("unknown level %q (want AA or AAA)", level)
}
