package contrast

import "github.com/d3uceY/Retna/color"

// The WCAG 2.2 contrast minimums.
const (
	aaNormal  = 4.5
	aaLarge   = 3.0
	aaaNormal = 7.0
	aaaLarge  = 4.5
)

// WCAG is the WCAG 2.x contrast ratio.
type WCAG struct{}

// Name returns "wcag".
func (WCAG) Name() string { return "wcag" }

// Calculate returns the contrast ratio and the four WCAG verdicts.
func (WCAG) Calculate(foreground, background color.Color) Result {
	hi, lo := foreground.Luminance(), background.Luminance()
	if hi < lo {
		hi, lo = lo, hi
	}
	ratio := (hi + 0.05) / (lo + 0.05)
	return Result{
		Algorithm: "wcag",
		Value:     ratio,
		Checks: []Check{
			{Name: "AA Normal", Min: aaNormal, Pass: ratio >= aaNormal},
			{Name: "AA Large", Min: aaLarge, Pass: ratio >= aaLarge},
			{Name: "AAA Normal", Min: aaaNormal, Pass: ratio >= aaaNormal},
			{Name: "AAA Large", Min: aaaLarge, Pass: ratio >= aaaLarge},
		},
	}
}

func init() { register(WCAG{}) }
