// Command retna converts colors between color spaces and measures how
// readable one color is on another.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/spf13/cobra"

	"github.com/d3uceY/Retna/contrast"
	"github.com/d3uceY/Retna/output"
)

// version can be overridden at build time:
//
//	go build -ldflags "-X main.version=v1.2.3"
var version = "v0.1.0"

// exitCode lets a command report a failed check without returning an error,
// which would make cobra print the whole usage block.
var exitCode int

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, output.Fail.Render("error: "+err.Error()))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "retna",
		Short: "Convert colors and measure their contrast",
		Long: `Retna converts colors between sRGB, HSL, HWB, Lab, OKLab and the rest,
and measures how readable one color is on another.

Colors can be written as hex, rgb(), hsl(), hsv(), hwb(), lab(), lch(),
oklab(), oklch() or a CSS color name. Contrast can be measured with WCAG 2.x
or APCA.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newContrastCmd(),
		newConvertCmd(),
		newInspectCmd(),
		newPaletteCmd(),
		newReadableCmd(),
		newFixCmd(),
		newVersionCmd(),
	)
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Retna version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Retna %s\n", version)
			return nil
		},
	}
}

// algorithmFlagHelp describes the algorithms Retna can run, using the registry
// so the help text cannot drift from the code.
func algorithmFlagHelp() string {
	return "algorithm to run: " + joinNames(contrast.Names())
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

// positiveFinite reports whether v is a real number above zero, which is what
// the ratio flags need before anything is compared against them. The > 0 test
// is false for NaN, and the infinity check catches what it would let through.
func positiveFinite(v float64) bool {
	return v > 0 && !math.IsInf(v, 0)
}
