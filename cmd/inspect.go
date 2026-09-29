package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/color"
	"github.com/d3uceY/Retna/output"
)

func newInspectCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "inspect <color>",
		Short: "Show every form of a color and its relative luminance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(cmd.OutOrStdout(), args[0], jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "print JSON instead of a table")
	return cmd
}

func runInspect(w io.Writer, input string, jsonOut bool) error {
	c, err := color.Parse(input)
	if err != nil {
		return err
	}

	if jsonOut {
		return output.WriteJSON(w, output.NewColorJSON(input, c))
	}

	fmt.Fprintf(w, "%s  %s\n\n", output.Swatch(c), output.Heading("Color"))
	rows := make([][]string, 0, len(color.Spaces))
	for _, space := range color.Spaces {
		value, err := color.Format(c, space)
		if err != nil {
			return err
		}
		rows = append(rows, []string{strings.ToUpper(space), value})
	}
	fmt.Fprint(w, output.Columns(rows))

	fmt.Fprintln(w)
	fmt.Fprintln(w, output.Heading("Relative luminance"))
	fmt.Fprint(w, output.Columns([][]string{
		{"Luminance", fmt.Sprintf("%.4f", c.Luminance())},
	}))
	return nil
}
