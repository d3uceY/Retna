package contrast

import (
	"math"

	"github.com/d3uceY/Retna/color"
)

// Constants from the APCA 0.1.9 reference implementation.
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

// Calculate returns the Lc value and the usual APCA guidance bands.
func (APCA) Calculate(foreground, background color.Color) Result {
	lc := apcaLc(foreground, background)
	abs := math.Abs(lc)
	return Result{
		Algorithm: "apca",
		Value:     lc,
		Checks: []Check{
			{Name: "Lc 45 Fluent Text", Min: 45, Pass: abs >= 45},
			{Name: "Lc 60 Body Text", Min: 60, Pass: abs >= 60},
			{Name: "Lc 75 Body Preferred", Min: 75, Pass: abs >= 75},
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
