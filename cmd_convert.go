package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/output"
)

type convertOptions struct {
	input string
	to    []string
	json  bool
}

func newConvertCmd() *cobra.Command {
	opts := convertOptions{}
	cmd := &cobra.Command{
		Use:   "convert <color>",
		Short: "Convert a color into other color spaces",
		Long: `Print a color in one or more color spaces.

--to takes a space name, a comma separated list or "all". With no --to the
color is printed in every space. Recognized spaces: ` + strings.Join(color.Spaces, ", ") + `.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.input = args[0]
			return runConvert(cmd.OutOrStdout(), opts)
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&opts.to, "to", "t", []string{"all"}, "color space to print, a comma separated list, or all")
	f.BoolVar(&opts.json, "json", false, "print JSON instead of a table")
	return cmd
}

func runConvert(w io.Writer, opts convertOptions) error {
	c, err := color.Parse(opts.input)
	if err != nil {
		return err
	}
	spaces, err := resolveSpaces(opts.to)
	if err != nil {
		return err
	}

	if opts.json {
		values := make(map[string]string, len(spaces))
		for _, space := range spaces {
			value, err := color.Format(c, space)
			if err != nil {
				return err
			}
			values[space] = value
		}
		return output.WriteJSON(w, output.ConvertJSON{Input: c.Hex(), Values: values})
	}

	fmt.Fprintf(w, "%s  %s\n\n", output.Swatch(c), output.Heading(c.Hex()))
	rows := make([][]string, 0, len(spaces))
	for _, space := range spaces {
		value, err := color.Format(c, space)
		if err != nil {
			return err
		}
		rows = append(rows, []string{strings.ToUpper(space), value})
	}
	fmt.Fprint(w, output.Columns(rows))
	return nil
}

// resolveSpaces expands "all" and rejects names that are not in Spaces.
func resolveSpaces(to []string) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	add := func(space string) {
		if !seen[space] {
			seen[space] = true
			out = append(out, space)
		}
	}
	for _, value := range to {
		for _, part := range strings.Split(value, ",") {
			part = strings.ToLower(strings.TrimSpace(part))
			switch {
			case part == "":
			case part == "all":
				for _, space := range color.Spaces {
					add(space)
				}
			case slices.Contains(color.Spaces, part):
				add(part)
			default:
				return nil, fmt.Errorf("unknown color space %q (want one of %s)", part, strings.Join(color.Spaces, ", "))
			}
		}
	}
	if len(out) == 0 {
		out = append(out, color.Spaces...)
	}
	return out, nil
}
