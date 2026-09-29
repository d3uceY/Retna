package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newPaletteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "palette",
		Short: "Work with a list of colors",
	}
	cmd.AddCommand(newPaletteContrastCmd())
	return cmd
}

func newPaletteContrastCmd() *cobra.Command {
	var (
		foreground string
		colors     []string
		file       string
		algorithm  string
		all        bool
		min        float64
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "contrast",
		Short: "Check a foreground against colors from a flag, a file or stdin",
		Long: `Measure one foreground color against a whole palette.

Colors come from --colors, from --file, or from stdin when the file is "-".
One color per line is expected, and blank lines and // comments are skipped.
The output matches "retna contrast --against".`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			list, err := readColorList(cmd, colors, file)
			if err != nil {
				return err
			}
			code, err := runContrast(cmd.OutOrStdout(), contrastOptions{
				foreground: foreground,
				against:    list,
				algorithm:  algorithm,
				all:        all,
				min:        min,
				minSet:     cmd.Flags().Changed("min"),
				json:       jsonOut,
			})
			if err != nil {
				return err
			}
			exitCode = code
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&foreground, "foreground", "f", "", "foreground color to measure (required)")
	f.StringSliceVarP(&colors, "colors", "c", nil, "colors to compare against, repeatable and comma separated")
	f.StringVar(&file, "file", "", "read colors from a file, or - for stdin")
	f.StringVarP(&algorithm, "algorithm", "a", "wcag", algorithmFlagHelp())
	f.BoolVar(&all, "all", false, "run every algorithm")
	f.Float64Var(&min, "min", 0, "minimum WCAG ratio, exits 1 when any pair falls below it")
	f.BoolVar(&jsonOut, "json", false, "print JSON instead of a table")
	_ = cmd.MarkFlagRequired("foreground")
	return cmd
}

// readColorList collects colors from --colors and from a file or stdin.
func readColorList(cmd *cobra.Command, colors []string, file string) ([]string, error) {
	list := splitList(colors)
	if file != "" {
		fromFile, err := readColorFile(cmd, file)
		if err != nil {
			return nil, err
		}
		list = append(list, fromFile...)
	}
	if len(list) == 0 {
		return nil, errors.New("no colors given, use --colors or --file")
	}
	return list, nil
}

func readColorFile(cmd *cobra.Command, path string) ([]string, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(cmd.InOrStdin())
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("reading colors from %s: %w", path, err)
	}

	// Windows editors write a UTF-8 BOM without asking, and it would otherwise
	// end up glued to the first color on the first line.
	text := strings.TrimPrefix(string(data), "\ufeff")

	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		out = append(out, splitList([]string{line})...)
	}
	return out, nil
}
