package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

type contrastOptions struct {
	foreground string
	background string
	against    []string
	algorithm  string
	all        bool
	min        float64
	minSet     bool
	json       bool
}

// contrastRow is one measured background for a single foreground.
// fg is the foreground after alpha has been resolved, which is the color the
// numbers belong to.
type contrastRow struct {
	fg    color.Color
	fgHex string
	bg    color.Color
	bgHex string
	res   []contrast.Result
}

func newContrastCmd() *cobra.Command {
	opts := contrastOptions{}
	cmd := &cobra.Command{
		Use:   "contrast <foreground> [background]",
		Short: "Measure the contrast between a foreground and a background",
		Long: `Measure how readable a foreground color is.

Give one background as the second argument, or list several with --against,
which can be repeated and takes comma separated values. Use --min to turn the
result into a check that exits 1 when a pair falls short, which is what CI
wants. Translucent colors are composited over white before measuring.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.foreground = args[0]
			if len(args) == 2 {
				opts.background = args[1]
			}
			opts.minSet = cmd.Flags().Changed("min")
			code, err := runContrast(cmd.OutOrStdout(), opts)
			if err != nil {
				return err
			}
			exitCode = code
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.algorithm, "algorithm", "a", "wcag", algorithmFlagHelp())
	f.BoolVar(&opts.all, "all", false, "run every algorithm")
	f.StringArrayVar(&opts.against, "against", nil, "backgrounds to compare against, repeatable and comma separated")
	f.Float64Var(&opts.min, "min", 0, "minimum WCAG ratio, exits 1 when any pair falls below it")
	f.BoolVar(&opts.json, "json", false, "print JSON instead of a table")
	return cmd
}

func runContrast(w io.Writer, opts contrastOptions) (int, error) {
	fg, err := color.Parse(opts.foreground)
	if err != nil {
		return 0, fmt.Errorf("foreground: %w", err)
	}

	targets := splitList(opts.against)
	if len(targets) == 0 {
		if opts.background == "" {
			return 0, errors.New("give a background color or list some with --against")
		}
		targets = []string{opts.background}
	}

	algorithms, err := selectAlgorithms(opts.algorithm, opts.all)
	if err != nil {
		return 0, err
	}
	rows := make([]contrastRow, 0, len(targets))
	composited := false
	for _, target := range targets {
		bg, err := color.Parse(target)
		if err != nil {
			return 0, fmt.Errorf("background %q: %w", target, err)
		}
		if fg.A < 1 || bg.A < 1 {
			composited = true
		}
		flatFG, flatBG := flatten(fg, bg)
		results := make([]contrast.Result, 0, len(algorithms))
		for _, algorithm := range algorithms {
			results = append(results, algorithm.Calculate(flatFG, flatBG))
		}
		rows = append(rows, contrastRow{
			fg:    flatFG,
			fgHex: flatFG.Hex(),
			bg:    flatBG,
			bgHex: flatBG.Hex(),
			res:   results,
		})
	}

	below := 0
	if opts.minSet {
		for _, row := range rows {
			if value, ok := wcagValue(row.res); ok && value < opts.min {
				below++
			}
		}
	}

	if opts.json {
		if err := writeContrastJSON(w, rows); err != nil {
			return 0, err
		}
	} else {
		writeContrastText(w, rows, opts, below, composited)
	}

	if below > 0 {
		return 1, nil
	}
	return 0, nil
}

// splitList flattens repeated flags that may also hold comma separated values.
func splitList(values []string) []string {
	var out []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func selectAlgorithms(name string, all bool) ([]contrast.Algorithm, error) {
	if !all {
		if name == "" {
			name = "wcag"
		}
		algorithm, err := contrast.Get(name)
		if err != nil {
			return nil, err
		}
		return []contrast.Algorithm{algorithm}, nil
	}
	names := contrast.Names()
	algorithms := make([]contrast.Algorithm, 0, len(names))
	for _, n := range names {
		algorithm, err := contrast.Get(n)
		if err != nil {
			return nil, err
		}
		algorithms = append(algorithms, algorithm)
	}
	return algorithms, nil
}

// flatten resolves alpha the way a browser paints it: the background sits on a
// white page, then the foreground sits on that background.
func flatten(fg, bg color.Color) (color.Color, color.Color) {
	if bg.A < 1 {
		bg = bg.Over(color.RGB(1, 1, 1))
	}
	if fg.A < 1 {
		fg = fg.Over(bg)
	}
	return fg, bg
}

func wcagValue(results []contrast.Result) (float64, bool) {
	for _, r := range results {
		if r.Algorithm == "wcag" {
			return r.Value, true
		}
	}
	return 0, false
}

func checkPass(r contrast.Result, name string) bool {
	for _, c := range r.Checks {
		if c.Name == name {
			return c.Pass
		}
	}
	return false
}

func formatRatio(v float64) string { return fmt.Sprintf("%.2f:1", v) }

func writeContrastText(w io.Writer, rows []contrastRow, opts contrastOptions, below int, composited bool) {
	if len(rows) == 1 {
		writeSingleContrast(w, rows[0])
	} else {
		writeContrastTable(w, rows)
	}
	if opts.minSet {
		writeMinVerdict(w, rows, opts.min, below)
	}
	if composited {
		fmt.Fprintln(w)
		fmt.Fprintln(w, output.Warn.Render("note: translucent colors were composited over white before measuring"))
	}
}

func writeSingleContrast(w io.Writer, row contrastRow) {
	fmt.Fprintln(w, output.Heading("Contrast"))
	fmt.Fprintln(w)
	fmt.Fprint(w, output.Columns([][]string{
		{"Foreground", output.Swatch(row.fg), row.fgHex},
		{"Background", output.Swatch(row.bg), row.bgHex},
	}))
	for _, result := range row.res {
		fmt.Fprintln(w)
		fmt.Fprintln(w, output.Heading(strings.ToUpper(result.Algorithm)))
		label, value := "Ratio", formatRatio(result.Value)
		if result.Algorithm != "wcag" {
			label, value = "Lc", fmt.Sprintf("%.2f", result.Value)
		}
		lines := [][]string{{label, value}}
		for _, check := range result.Checks {
			lines = append(lines, []string{check.Name, output.Verdict(check.Pass)})
		}
		fmt.Fprint(w, output.Columns(lines))
	}
	fmt.Fprintln(w)
}

func writeContrastTable(w io.Writer, rows []contrastRow) {
	headers := []string{"Color"}
	for _, name := range algorithmOrder(rows) {
		switch name {
		case "wcag":
			headers = append(headers, "Ratio", "AA", "AAA")
		case "apca":
			headers = append(headers, "Lc")
		default:
			headers = append(headers, name)
		}
	}

	body := make([][]string, 0, len(rows))
	for _, row := range rows {
		cells := []string{output.Swatch(row.bg) + " " + row.bgHex}
		for _, result := range row.res {
			switch result.Algorithm {
			case "wcag":
				cells = append(cells,
					formatRatio(result.Value),
					output.Verdict(checkPass(result, "AA Normal")),
					output.Verdict(checkPass(result, "AAA Normal")),
				)
			default:
				cells = append(cells, fmt.Sprintf("%.2f", result.Value))
			}
		}
		body = append(body, cells)
	}

	fmt.Fprintln(w, output.Heading("Contrast"))
	fmt.Fprintln(w)
	fmt.Fprint(w, output.Columns([][]string{{"Foreground", output.Swatch(rows[0].fg), rows[0].fgHex}}))
	fmt.Fprintln(w)
	fmt.Fprint(w, output.Table(headers, body))
}

func algorithmOrder(rows []contrastRow) []string {
	if len(rows) == 0 {
		return nil
	}
	names := make([]string, 0, len(rows[0].res))
	for _, result := range rows[0].res {
		names = append(names, result.Algorithm)
	}
	return names
}

func writeMinVerdict(w io.Writer, rows []contrastRow, min float64, below int) {
	fmt.Fprintln(w)
	if len(rows) == 1 {
		value, _ := wcagValue(rows[0].res)
		label := output.Pass.Render("PASS")
		if below > 0 {
			label = output.Fail.Render("FAIL")
		}
		fmt.Fprintf(w, "%s  %s\n", label, formatRatio(value))
		fmt.Fprintln(w, output.Label(fmt.Sprintf("Required: %.2f:1", min)))
		return
	}
	if below > 0 {
		fmt.Fprintln(w, output.Fail.Render(fmt.Sprintf("FAIL  %d of %d below %.2f:1", below, len(rows), min)))
		return
	}
	fmt.Fprintln(w, output.Pass.Render(fmt.Sprintf("PASS  all %d pairs meet %.2f:1", len(rows), min)))
}

func writeContrastJSON(w io.Writer, rows []contrastRow) error {
	if len(rows) == 1 {
		return output.WriteJSON(w, contrastEntry(rows[0]))
	}
	list := output.ContrastListJSON{
		Foreground: rows[0].fgHex,
		Results:    make([]output.ContrastJSON, 0, len(rows)),
	}
	for _, row := range rows {
		list.Results = append(list.Results, contrastEntry(row))
	}
	return output.WriteJSON(w, list)
}

func contrastEntry(row contrastRow) output.ContrastJSON {
	entry := output.ContrastJSON{Foreground: row.fgHex, Background: row.bgHex}
	for _, result := range row.res {
		switch result.Algorithm {
		case "wcag":
			entry.Contrast = result.Value
			entry.WCAG = &output.WCAGJSON{
				AA: output.LevelJSON{
					Normal: checkPass(result, "AA Normal"),
					Large:  checkPass(result, "AA Large"),
				},
				AAA: output.LevelJSON{
					Normal: checkPass(result, "AAA Normal"),
					Large:  checkPass(result, "AAA Large"),
				},
			}
		case "apca":
			value := result.Value
			entry.APCA = &value
		}
	}
	return entry
}
