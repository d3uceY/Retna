package contrast

import (
	"math"

	"github.com/d3uceY/Retna/color"
)

// Constants from the APCA 0.1.9 reference implementation (apca-w3).
//
// APCA is a draft for WCAG 3, not a standard. WCAG 2.2 remains the
// Recommendation, which is why the WCAG algorithm is the default elsewhere.
const (
	apcaSR = 0.2126729
	apcaSG = 0.7151522
	apcaSB = 0.0721750

	apcaNormBG  = 0.56
	apcaNormTXT = 0.57
	apcaRevBG   = 0.65
	apcaRevTXT  = 0.62

	apcaBlkThrs   = 0.022
	apcaBlkClmp   = 1.414
	apcaScale     = 1.14
	apcaLoOffset  = 0.027
	apcaLoClip    = 0.1
	apcaDeltaYMin = 0.0005
)

// APCA is the Accessible Perceptual Contrast Algorithm.
//
// APCA is polarity aware: dark text on a light background is a positive Lc,
// light text on a dark background is negative. Compare on the absolute value.
type APCA struct{}

// Name returns "apca".
func (APCA) Name() string { return "apca" }

// Calculate returns the Lc value and APCA's published use case bands.
func (APCA) Calculate(foreground, background color.Color) Result {
	lc := apcaLc(foreground, background)
	abs := math.Abs(lc)
	return Result{
		Algorithm: "apca",
		Value:     lc,
		// The names follow the use case table in the reference: 90 is the
		// preferred level for body text, 75 its minimum, 60 is content text
		// that is not body text, 45 is large or heavy text, 30 is the floor for
		// any text at all, and 15 is the floor for non-text elements.
		Checks: []Check{
			{Name: "Lc 90 Body Preferred", Min: 90, Pass: abs >= 90},
			{Name: "Lc 75 Body Minimum", Min: 75, Pass: abs >= 75},
			{Name: "Lc 60 Content Text", Min: 60, Pass: abs >= 60},
			{Name: "Lc 45 Large Text", Min: 45, Pass: abs >= 45},
			{Name: "Lc 30 Text Floor", Min: 30, Pass: abs >= 30},
			{Name: "Lc 15 Non-text", Min: 15, Pass: abs >= 15},
		},
	}
}

// apcaLc is the ported APCA contrast calculation.
func apcaLc(foreground, background color.Color) float64 {
	txtY := apcaLuminance(foreground)
	bgY := apcaLuminance(background)
	if math.Abs(bgY-txtY) < apcaDeltaYMin {
		return 0
	}
	if bgY > txtY {
		sapc := (math.Pow(bgY, apcaNormBG) - math.Pow(txtY, apcaNormTXT)) * apcaScale
		if sapc < apcaLoClip {
			return 0
		}
		return (sapc - apcaLoOffset) * 100
	}
	sapc := (math.Pow(bgY, apcaRevBG) - math.Pow(txtY, apcaRevTXT)) * apcaScale
	if sapc > -apcaLoClip {
		return 0
	}
	return (sapc + apcaLoOffset) * 100
}

// apcaLuminance is APCA's own luminance, which uses a plain 2.4 exponent and
// soft clamps near black.
func apcaLuminance(c color.Color) float64 {
	y := apcaSR*pow24(c.R) + apcaSG*pow24(c.G) + apcaSB*pow24(c.B)
	if y > apcaBlkThrs {
		return y
	}
	return y + math.Pow(apcaBlkThrs-y, apcaBlkClmp)
}

func pow24(v float64) float64 {
	return math.Pow(clamp01(v), 2.4)
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

func init() { register(APCA{}) }
