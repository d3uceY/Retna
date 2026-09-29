package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Spaces lists every space Format understands, in display order. The last
// group are the CSS color() spaces.
var Spaces = []string{
	"hex", "rgb", "rgba", "hsl", "hsv", "hwb",
	"lab", "lch", "oklab", "oklch",
	"srgb", "display-p3", "a98-rgb", "prophoto-rgb", "rec2020",
}

// HSL builds a color from hue in degrees and saturation/lightness in 0..1.
func HSL(h, s, l, a float64) Color {
	h = wrapHue(h)
	if s <= 0 {
		return Color{l, l, l, a}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	h /= 360
	return Color{
		hueToChannel(p, q, h+1.0/3),
		hueToChannel(p, q, h),
		hueToChannel(p, q, h-1.0/3),
		a,
	}
}

func hueToChannel(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}

// ToHSL returns hue in degrees and saturation/lightness in 0..1.
func ToHSL(c Color) (h, s, l float64) {
	max := math.Max(c.R, math.Max(c.G, c.B))
	min := math.Min(c.R, math.Min(c.G, c.B))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case c.R:
		h = (c.G - c.B) / d
		if c.G < c.B {
			h += 6
		}
	case c.G:
		h = (c.B-c.R)/d + 2
	default:
		h = (c.R-c.G)/d + 4
	}
	return h * 60, s, l
}

// HSV builds a color from hue in degrees and saturation/value in 0..1.
func HSV(h, s, v, a float64) Color {
	h = wrapHue(h) / 60
	if s <= 0 {
		return Color{v, v, v, a}
	}
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)
	switch int(i) % 6 {
	case 0:
		return Color{v, t, p, a}
	case 1:
		return Color{q, v, p, a}
	case 2:
		return Color{p, v, t, a}
	case 3:
		return Color{p, q, v, a}
	case 4:
		return Color{t, p, v, a}
	default:
		return Color{v, p, q, a}
	}
}

// ToHSV returns hue in degrees and saturation/value in 0..1.
func ToHSV(c Color) (h, s, v float64) {
	max := math.Max(c.R, math.Max(c.G, c.B))
	min := math.Min(c.R, math.Min(c.G, c.B))
	v = max
	d := max - min
	if max > 0 {
		s = d / max
	}
	if d == 0 {
		return 0, s, v
	}
	switch max {
	case c.R:
		h = (c.G - c.B) / d
	case c.G:
		h = (c.B-c.R)/d + 2
	default:
		h = (c.R-c.G)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, v
}

// HWB builds a color from hue in degrees and whiteness/blackness in 0..1.
func HWB(h, w, bl, a float64) Color {
	if w+bl >= 1 {
		g := w / (w + bl)
		return Color{g, g, g, a}
	}
	base := HSV(h, 1, 1, 1)
	f := 1 - w - bl
	return Color{base.R*f + w, base.G*f + w, base.B*f + w, a}
}

// ToHWB returns hue in degrees and whiteness/blackness in 0..1.
func ToHWB(c Color) (h, w, bl float64) {
	h, _, v := ToHSV(c)
	w = math.Min(c.R, math.Min(c.G, c.B))
	bl = 1 - v
	return h, w, bl
}

// Lab builds a CIE Lab color. Lightness is 0..100, a and b are unbounded.
// Lab is defined against the D50 white point, matching CSS.
func Lab(l, a, b, alpha float64) Color {
	fy := (l + 16) / 116
	fx := fy + a/500
	fz := fy - b/200
	x, y, z := d50ToD65(d50Xn*labFInv(fx), d50Yn*labFInv(fy), d50Zn*labFInv(fz))
	return xyzToSrgb(x, y, z, alpha)
}

// ToLab converts to CIE Lab.
func ToLab(c Color) (l, a, b float64) {
	x, y, z := srgbToXYZ(c)
	x, y, z = d65ToD50(x, y, z)
	fx := labF(x / d50Xn)
	fy := labF(y / d50Yn)
	fz := labF(z / d50Zn)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// LCH builds a color from Lab lightness, chroma and hue in degrees.
func LCH(l, chroma, h, alpha float64) Color {
	rad := h * math.Pi / 180
	return Lab(l, chroma*math.Cos(rad), chroma*math.Sin(rad), alpha)
}

// Achromatic thresholds. Hue carries no information below these, and the
// conversion leaves a small residue on a neutral color: about 8e-6 in Lab and
// 4e-8 in OKLab. Reporting that residue as an angle makes a gray print a
// different hue in each space, so it is reported as 0 instead. Both thresholds
// sit far below any chroma the eye can see.
//
// These are CSS Color 4's own per space epsilons for a powerless hue (0.0015
// for LCH in §9.3, 0.000004 for OkLCh in §9.4), not invented numbers. CSS
// treats such a hue as missing (none); printing 0 is Retna's own choice, and
// it is why a gray reports hue 0 rather than an arbitrary angle.
const (
	achromaticLab   = 0.0015
	achromaticOKLab = 0.000004
)

// ToLCH converts to Lab lightness, chroma and hue in degrees.
func ToLCH(c Color) (l, chroma, h float64) {
	l, a, b := ToLab(c)
	chroma = math.Hypot(a, b)
	if chroma < achromaticLab {
		return l, chroma, 0
	}
	h = math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return l, chroma, h
}

// OKLab builds a color from OKLab lightness 0..1, a and b.
func OKLab(l, a, b, alpha float64) Color {
	lp := l + 0.3963377774*a + 0.2158037573*b
	mp := l - 0.1055613458*a - 0.0638541728*b
	sp := l - 0.0894841775*a - 1.2914855480*b
	lc, mc, sc := lp*lp*lp, mp*mp*mp, sp*sp*sp
	r := 4.0767416621*lc - 3.3077115913*mc + 0.2309699292*sc
	g := -1.2684380046*lc + 2.6097574011*mc - 0.3413193965*sc
	bb := -0.0041960863*lc - 0.7034186147*mc + 1.7076147010*sc
	return fromLinear(r, g, bb, alpha)
}

// ToOKLab converts to OKLab.
func ToOKLab(c Color) (l, a, b float64) {
	r, g, bb := c.Linear()
	lc := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*bb)
	mc := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*bb)
	sc := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*bb)
	return 0.2104542553*lc + 0.7936177850*mc - 0.0040720468*sc,
		1.9779984951*lc - 2.4285922050*mc + 0.4505937099*sc,
		0.0259040371*lc + 0.7827717662*mc - 0.8086757660*sc
}

// OKLCH builds a color from OKLab lightness 0..1, chroma and hue in degrees.
func OKLCH(l, chroma, h, alpha float64) Color {
	rad := h * math.Pi / 180
	return OKLab(l, chroma*math.Cos(rad), chroma*math.Sin(rad), alpha)
}

// ToOKLCH converts to OKLab lightness, chroma and hue in degrees.
func ToOKLCH(c Color) (l, chroma, h float64) {
	l, a, b := ToOKLab(c)
	chroma = math.Hypot(a, b)
	if chroma < achromaticOKLab {
		return l, chroma, 0
	}
	h = math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return l, chroma, h
}

// XYZ conversions use the D65 white point, which is what sRGB is defined
// against. Lab and ProPhoto need D50, so the two Bradford matrices below
// bridge them.
func srgbToXYZ(c Color) (x, y, z float64) {
	r, g, b := c.Linear()
	return srgbToXYZLinear(r, g, b)
}

func srgbToXYZLinear(r, g, b float64) (x, y, z float64) {
	x = 0.4123907992659595*r + 0.3575843393838780*g + 0.1804807884018343*b
	y = 0.2126390058715104*r + 0.7151686787677560*g + 0.0721923153607337*b
	z = 0.0193308187155918*r + 0.1191947797946260*g + 0.9505321522496607*b
	return x, y, z
}

func xyzToSrgb(x, y, z, a float64) Color {
	r, g, b := xyzToSrgbLinear(x, y, z)
	return fromLinear(r, g, b, a)
}

func xyzToSrgbLinear(x, y, z float64) (r, g, b float64) {
	r = 3.2409699419045226*x - 1.5373831775700940*y - 0.4986107602930034*z
	g = -0.9692436362808796*x + 1.8759675015077204*y + 0.0415550574071756*z
	b = 0.0556300796969936*x - 0.2039769588889765*y + 1.0569715142428786*z
	return r, g, b
}

// These are the linear Bradford matrices from the CSS Color 4 sample code
// (§19). The updated values were computed from the inverse of the cone
// response matrix rather than from rounded intermediate values, so this pair
// is an exact inverse (to about 1e-16) and it carries the D65 white point onto
// the chromaticity derived D50 below to the same precision. The older pair
// (1.0479298208405488 ...) was off by about 1e-7 and drifted on round trips.
func d65ToD50(x, y, z float64) (float64, float64, float64) {
	return 1.0479297925449969*x + 0.022946870601609652*y - 0.05019226628920524*z,
		0.02962780877005599*x + 0.9904344267538799*y - 0.017073799063418826*z,
		-0.009243040646204504*x + 0.015055191490298152*y + 0.7518742814281371*z
}

func d50ToD65(x, y, z float64) (float64, float64, float64) {
	return 0.955473421488075*x - 0.02309845494876471*y + 0.06325924320057072*z,
		-0.0283697093338637*x + 1.0099953980813041*y + 0.021041441191917323*z,
		0.012314014864481998*x - 0.020507649298898964*y + 1.330365926242124*z
}

const (
	labEps   = 216.0 / 24389.0
	labKappa = 24389.0 / 27.0

	// The D50 white point is derived from the chromaticity (0.3457, 0.3585)
	// that CSS Color 4 defines in its white point table (§2), which is what the
	// spec's sample code uses too. The rounded ICC style pair (0.9642, 0.8251)
	// is a different, earlier draft convention, and mixing it with the Bradford
	// matrices above would tilt Lab a and b everywhere. Do not reintroduce it.
	d50Xn = 0.3457 / 0.3585
	d50Yn = 1.0
	d50Zn = (1.0 - 0.3457 - 0.3585) / 0.3585
)

func labF(t float64) float64 {
	if t > labEps {
		return math.Cbrt(t)
	}
	return (labKappa*t + 16) / 116
}

func labFInv(t float64) float64 {
	t3 := t * t * t
	if t3 > labEps {
		return t3
	}
	return (116*t - 16) / labKappa
}

func wrapHue(h float64) float64 {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}

// Format renders a color in the named space. Recognized spaces are listed in
// Spaces.
func Format(c Color, space string) (string, error) {
	space = strings.ToLower(strings.TrimSpace(space))
	switch space {
	case "hex":
		return c.Hex(), nil
	case "rgb":
		r, g, b := c.Channels()
		return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), nil
	case "rgba":
		r, g, b := c.Channels()
		return fmt.Sprintf("rgba(%d, %d, %d, %s)", r, g, b, num(c.A, 3)), nil
	case "hsl":
		h, s, l := ToHSL(c)
		return fmt.Sprintf("hsl(%s, %s%%, %s%%)", num(h, 2), num(s*100, 2), num(l*100, 2)), nil
	case "hsv":
		h, s, v := ToHSV(c)
		return fmt.Sprintf("hsv(%s, %s%%, %s%%)", num(h, 2), num(s*100, 2), num(v*100, 2)), nil
	case "hwb":
		h, w, bl := ToHWB(c)
		return fmt.Sprintf("hwb(%s %s%% %s%%)", num(h, 2), num(w*100, 2), num(bl*100, 2)), nil
	case "lab":
		l, a, b := ToLab(c)
		return fmt.Sprintf("lab(%s, %s, %s)", num(l, 2), num(a, 2), num(b, 2)), nil
	case "lch":
		l, chroma, h := ToLCH(c)
		return fmt.Sprintf("lch(%s, %s, %s)", num(l, 2), num(chroma, 2), num(h, 2)), nil
	case "oklab":
		l, a, b := ToOKLab(c)
		return fmt.Sprintf("oklab(%s, %s, %s)", num(l, 4), num(a, 4), num(b, 4)), nil
	case "oklch":
		l, chroma, h := ToOKLCH(c)
		return fmt.Sprintf("oklch(%s, %s, %s)", num(l, 4), num(chroma, 4), num(h, 2)), nil
	case "srgb", "display-p3", "a98-rgb", "prophoto-rgb", "rec2020":
		return FormatWideGamut(c, space)
	}
	return "", fmt.Errorf("unknown color space %q (want one of %s)", space, strings.Join(Spaces, ", "))
}

// num formats a float with a fixed precision and drops trailing zeros.
func num(v float64, prec int) string {
	s := strconv.FormatFloat(v, 'f', prec, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "-0" {
		s = "0"
	}
	return s
}
