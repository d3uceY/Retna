package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

type fixOptions struct {
	background string
	level      string
	target     float64
	targetSet  bool
	suggest    int
	json       bool
}

func newFixCmd() *cobra.Command {
	opts := fixOptions{}
	cmd := &cobra.Command{
		Use:   "fix <color>",
		Short: "Find the nearest color that meets a contrast target",
		Long: `Nudge a color until it is readable on a background.

Retna walks OKLab lightness in the direction that increases contrast and keeps
chroma and hue alone, so the fixed color still looks like the original. The
first suggestions are the smallest changes that meet the target.

Set the target with --level (AA or AAA) or an explicit --target ratio.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.targetSet = cmd.Flags().Changed("target")
			return runFix(cmd.OutOrStdout(), args[0], opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.background, "background", "b", "", "background color to measure against (required)")
	f.StringVar(&opts.level, "level", "AA", "target WCAG level: AA or AAA")
	f.Float64Var(&opts.target, "target", 0, "explicit contrast ratio to reach, overrides --level")
	f.IntVar(&opts.suggest, "suggest", 3, "how many passing colors to print")
	f.BoolVar(&opts.json, "json", false, "print JSON instead of a table")
	_ = cmd.MarkFlagRequired("background")
	return cmd
}

func runFix(w io.Writer, input string, opts fixOptions) error {
	fg, err := color.Parse(input)
	if err != nil {
		return err
	}
	bg, err := color.Parse(opts.background)
	if err != nil {
		return fmt.Errorf("background: %w", err)
	}

	target := opts.target
	if !opts.targetSet {
		target, err = levelMin(opts.level)
		if err != nil {
			return err
		}
	}
	if !positiveFinite(target) || target < 1 {
		return fmt.Errorf("target must be a ratio of at least 1:1, got %v", target)
	}
	// Black on white is the highest ratio there is, so nothing can reach more.
	if target > 21 {
		return fmt.Errorf("target cannot be higher than 21:1, got %v", target)
	}
	if opts.suggest < 1 {
		return fmt.Errorf("--suggest must be at least 1, got %d", opts.suggest)
	}

	flatFG, flatBG := flatten(fg, bg)
	original := contrast.WCAG{}.Calculate(flatFG, flatBG).Value
	suggestions := passingColors(flatFG, flatBG, target, opts.suggest)

	if opts.json {
		view := output.FixJSON{
			Input:       flatFG.Hex(),
			Background:  flatBG.Hex(),
			Target:      target,
			Contrast:    original,
			Pass:        original >= target,
			Suggestions: make([]output.FixSuggestionJSON, 0, len(suggestions)),
		}
		for _, s := range suggestions {
			view.Suggestions = append(view.Suggestions, output.FixSuggestionJSON{
				Hex:      s.hex,
				Contrast: s.ratio,
				Pass:     true,
			})
		}
		return output.WriteJSON(w, view)
	}

	fmt.Fprintln(w, output.Heading("Fix"))
	fmt.Fprintln(w)
	fmt.Fprint(w, output.Columns([][]string{
		{"Original", output.Swatch(flatFG) + " " + flatFG.Hex(), formatRatio(original), output.Verdict(original >= target)},
		{"Background", output.Swatch(flatBG) + " " + flatBG.Hex(), "", ""},
		{"Target", "", formatRatio(target), ""},
	}))

	if original >= target {
		fmt.Fprintln(w)
		fmt.Fprintln(w, output.Pass.Render("Already passes, nothing to change."))
		return nil
	}
	if len(suggestions) == 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, output.Fail.Render("No color at this chroma and hue reaches the target, try lowering the color's saturation or changing its hue."))
		return nil
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, output.Heading("Suggested colors"))
	rows := make([][]string, 0, len(suggestions))
	for i, s := range suggestions {
		note := ""
		if i == 0 {
			note = output.Label("smallest change")
		}
		rows = append(rows, []string{
			output.Swatch(s.color) + " " + s.hex,
			formatRatio(s.ratio),
			output.Verdict(true),
			note,
		})
	}
	fmt.Fprint(w, output.Table([]string{"Color", "Ratio", "WCAG", "Note"}, rows))
	return nil
}

type fixedColor struct {
	color color.Color
	hex   string
	ratio float64
}

// passingColors walks OKLab lightness away from the background until the
// target ratio is met, then keeps collecting distinct colors that also pass.
// Chroma and hue are left alone so the result still resembles the input.
func passingColors(fg, bg color.Color, target float64, count int) []fixedColor {
	if count <= 0 {
		return nil
	}
	l, chroma, hue := color.ToOKLCH(fg)
	lighten := ratioOf(color.OKLCH(1, chroma, hue, 1), bg) > ratioOf(color.OKLCH(0, chroma, hue, 1), bg)

	const step = 0.001
	var out []fixedColor
	seen := map[string]bool{fg.Hex(): true}
	for i := 0; i < 1000 && len(out) < count; i++ {
		if lighten {
			l += step
		} else {
			l -= step
		}
		if l < 0 || l > 1 {
			break
		}
		candidate := color.OKLCH(l, chroma, hue, 1)
		// Snap to a color a screen can show before measuring it. The unquantized
		// value can sit on the far side of the target from the byte triple it
		// rounds to, which would report a pass for a hex that actually fails.
		r8, g8, b8 := candidate.Channels()
		quantized := color.RGB8(r8, g8, b8)
		ratio := ratioOf(quantized, bg)
		if ratio < target {
			continue
		}
		hex := quantized.Hex()
		if seen[hex] {
			continue
		}
		seen[hex] = true
		out = append(out, fixedColor{color: quantized, hex: hex, ratio: ratio})
	}
	return out
}

func ratioOf(fg, bg color.Color) float64 {
	return contrast.WCAG{}.Calculate(fg, bg).Value
}
