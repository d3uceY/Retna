// Package color holds the internal sRGB color model plus the conversions
// between the color spaces Retna understands.
//
// Every input format normalizes to Color, so the contrast algorithms only ever
// deal with one representation.
package color

import (
	"fmt"
	"math"
)

// Color is an sRGB color. R, G and B are gamma-encoded sRGB channels in the
// 0..1 range. A is alpha, where 0 is fully transparent and 1 is opaque.
type Color struct {
	R, G, B, A float64
}

// RGB builds an opaque color from 0..1 gamma-encoded sRGB channels.
func RGB(r, g, b float64) Color { return Color{r, g, b, 1} }

// RGB8 builds an opaque color from 0..255 sRGB channels.
func RGB8(r, g, b uint8) Color {
	return Color{float64(r) / 255, float64(g) / 255, float64(b) / 255, 1}
}

// RGBA8 builds a color from 0..255 sRGB channels and 0..1 alpha.
func RGBA8(r, g, b uint8, a float64) Color {
	c := RGB8(r, g, b)
	c.A = a
	return c
}

// Channels returns the color as rounded 0..255 byte values.
func (c Color) Channels() (r, g, b uint8) {
	return byteChannel(c.R), byteChannel(c.G), byteChannel(c.B)
}

func byteChannel(v float64) uint8 {
	return uint8(math.Round(clamp01(v) * 255))
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

// Clamp limits every channel to the displayable sRGB range.
func (c Color) Clamp() Color {
	return Color{clamp01(c.R), clamp01(c.G), clamp01(c.B), clamp01(c.A)}
}

// Linear returns the linear-light sRGB channels.
func (c Color) Linear() (r, g, b float64) {
	return srgbToLinear(c.R), srgbToLinear(c.G), srgbToLinear(c.B)
}

func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearToSrgb(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// fromLinear builds a color from linear-light sRGB channels. Channels outside
// the gamut are clipped, which is what a screen would do anyway.
func fromLinear(r, g, b, a float64) Color {
	return Color{
		linearToSrgb(clamp01(r)),
		linearToSrgb(clamp01(g)),
		linearToSrgb(clamp01(b)),
		a,
	}
}

// Luminance returns the WCAG 2.x relative luminance.
func (c Color) Luminance() float64 {
	r, g, b := c.Linear()
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// Over composites c on top of an opaque background.
func (c Color) Over(bg Color) Color {
	if c.A >= 1 {
		return c
	}
	a := c.A
	return Color{
		R: c.R*a + bg.R*(1-a),
		G: c.G*a + bg.G*(1-a),
		B: c.B*a + bg.B*(1-a),
		A: 1,
	}
}

// Hex returns the color as uppercase #RRGGBB, or #RRGGBBAA when translucent.
func (c Color) Hex() string {
	c = c.Clamp()
	r, g, b := c.Channels()
	if c.A >= 1 {
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, byteChannel(c.A))
}

// String reports the color as hex.
func (c Color) String() string { return c.Hex() }
