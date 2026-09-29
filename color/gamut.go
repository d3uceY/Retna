package color

import (
	"fmt"
	"math"
	"strings"
)

// wideGamutNames lists the spaces CSS accepts inside color(), in display order.
var wideGamutNames = []string{"srgb", "display-p3", "a98-rgb", "prophoto-rgb", "rec2020"}

// gamut describes a linear RGB space. Each one is a set of primaries (a matrix
// to and from XYZ), a white point, and a pair of transfer functions. Keeping
// them in one table means the conversion path is written once.
type gamut struct {
	toXYZ   [9]float64
	fromXYZ [9]float64
	d50     bool
	decode  func(float64) float64
	encode  func(float64) float64
}

// toSRGB converts encoded components in this space to sRGB, clipping whatever
// falls outside the sRGB gamut.
func (space gamut) toSRGB(r, g, b, a float64) Color {
	x, y, z := mat3(space.toXYZ, space.decode(r), space.decode(g), space.decode(b))
	if space.d50 {
		x, y, z = d50ToD65(x, y, z)
	}
	lr, lg, lb := xyzToSrgbLinear(x, y, z)
	return fromLinear(lr, lg, lb, a)
}

// fromSRGB converts an sRGB color into this space. The result is deliberately
// not clipped: CSS allows components outside 0..1, and clipping would turn a
// faithful answer into a different color.
func (space gamut) fromSRGB(c Color) (r, g, b float64) {
	lr, lg, lb := c.Linear()
	x, y, z := srgbToXYZLinear(lr, lg, lb)
	if space.d50 {
		x, y, z = d65ToD50(x, y, z)
	}
	r, g, b = mat3(space.fromXYZ, x, y, z)
	return space.encode(r), space.encode(g), space.encode(b)
}

// mat3 applies a row major 3x3 matrix to a vector.
func mat3(m [9]float64, r, g, b float64) (x, y, z float64) {
	return m[0]*r + m[1]*g + m[2]*b,
		m[3]*r + m[4]*g + m[5]*b,
		m[6]*r + m[7]*g + m[8]*b
}

var gamuts = map[string]gamut{
	"srgb": {
		toXYZ: [9]float64{
			0.4123907992659595, 0.3575843393838780, 0.1804807884018343,
			0.2126390058715104, 0.7151686787677560, 0.0721923153607337,
			0.0193308187155918, 0.1191947797946260, 0.9505321522496607,
		},
		fromXYZ: [9]float64{
			3.2409699419045226, -1.5373831775700940, -0.4986107602930034,
			-0.9692436362808796, 1.8759675015077204, 0.0415550574071756,
			0.0556300796969936, -0.2039769588889765, 1.0569715142428786,
		},
		decode: srgbToLinear,
		encode: linearToSrgb,
	},
	"display-p3": {
		toXYZ: [9]float64{
			0.4865709486482162, 0.2656676931690931, 0.1982172852343625,
			0.2289745640697488, 0.6917385218365064, 0.0792869140937450,
			0.0000000000000000, 0.0451133818589026, 1.0439443689009760,
		},
		fromXYZ: [9]float64{
			2.4934969119414250, -0.9313836179191239, -0.4027107844507168,
			-0.8294889695615747, 1.7626640603183463, 0.0236246858419436,
			0.0358458302437845, -0.0761723892680418, 0.9568845240076872,
		},
		decode: srgbToLinear,
		encode: linearToSrgb,
	},
	"a98-rgb": {
		toXYZ: [9]float64{
			0.5766690429101305, 0.1855582379065463, 0.1882286462349947,
			0.2973449752505360, 0.6273635662554661, 0.0752914584939979,
			0.0270313613864123, 0.0706888525358272, 0.9913375368376388,
		},
		fromXYZ: [9]float64{
			2.0415879038107465, -0.5650069742788596, -0.3447313507783296,
			-0.9692436362808795, 1.8759675015077202, 0.0415550574071756,
			0.0134442806320311, -0.1183623922310184, 1.0151749943912054,
		},
		decode: a98ToLinear,
		encode: a98FromLinear,
	},
	"prophoto-rgb": {
		d50: true,
		toXYZ: [9]float64{
			0.7977604896723027, 0.1351858371757403, 0.0313493495815248,
			0.2880711282292934, 0.7118432178101014, 0.0000856539606053,
			0.0000000000000000, 0.0000000000000000, 0.8251046025104601,
		},
		fromXYZ: [9]float64{
			1.3457989731028281, -0.2555801000799753, -0.0511062850675340,
			-0.5446224939028347, 1.5082327413132781, 0.0205360323914797,
			0.0000000000000000, 0.0000000000000000, 1.2119675456389454,
		},
		decode: prophotoToLinear,
		encode: prophotoFromLinear,
	},
	"rec2020": {
		toXYZ: [9]float64{
			0.6369580483012914, 0.1446169035862083, 0.1688809751641721,
			0.2627002120112671, 0.6779980715188708, 0.0593017164698620,
			0.0000000000000000, 0.0280726930490874, 1.0609850577107910,
		},
		fromXYZ: [9]float64{
			1.7166511879712680, -0.3556707837763920, -0.2533662813736600,
			-0.6666843518324890, 1.6164812366349390, 0.0157685458139111,
			0.0176398574453110, -0.0427706132578090, 0.9421031212354740,
		},
		decode: rec2020ToLinear,
		encode: rec2020FromLinear,
	},
}

// signPow raises an absolute value to a power and puts the sign back, which is
// how CSS defines these transfer functions for out of gamut components.
func signPow(v, exp float64) float64 {
	return math.Copysign(math.Pow(math.Abs(v), exp), v)
}

const (
	a98Gamma = 563.0 / 256.0

	rec2020Alpha = 1.09929682680944
	rec2020Beta  = 0.018053968510807

	prophotoEt  = 1.0 / 512.0
	prophotoEt2 = 16.0 / 512.0
)

func a98ToLinear(v float64) float64   { return signPow(v, a98Gamma) }
func a98FromLinear(v float64) float64 { return signPow(v, 1/a98Gamma) }

func rec2020ToLinear(v float64) float64 {
	abs := math.Abs(v)
	if abs < rec2020Beta*4.5 {
		return v / 4.5
	}
	return math.Copysign(math.Pow((abs+rec2020Alpha-1)/rec2020Alpha, 1/0.45), v)
}

func rec2020FromLinear(v float64) float64 {
	if math.Abs(v) > rec2020Beta {
		return math.Copysign(rec2020Alpha*math.Pow(math.Abs(v), 0.45)-(rec2020Alpha-1), v)
	}
	return 4.5 * v
}

func prophotoToLinear(v float64) float64 {
	if math.Abs(v) <= prophotoEt2 {
		return v / 16
	}
	return signPow(v, 1.8)
}

func prophotoFromLinear(v float64) float64 {
	if math.Abs(v) >= prophotoEt {
		return signPow(v, 1/1.8)
	}
	return 16 * v
}

func wideComponents(c Color, name string) (r, g, b float64, err error) {
	space, ok := gamuts[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return 0, 0, 0, fmt.Errorf("unknown wide gamut space %q (want one of %s)", name, strings.Join(wideGamutNames, ", "))
	}
	r, g, b = space.fromSRGB(c)
	return r, g, b, nil
}

// WideGamut encodes c in a wide gamut space and returns the raw components.
// They are not clipped, because a clipped component no longer names the same
// color.
func WideGamut(c Color, name string) ([3]float64, error) {
	r, g, b, err := wideComponents(c, name)
	if err != nil {
		return [3]float64{}, err
	}
	return [3]float64{r, g, b}, nil
}

// FormatWideGamut renders c as a CSS color() function.
func FormatWideGamut(c Color, name string) (string, error) {
	r, g, b, err := wideComponents(c, name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("color(%s %s %s %s)", strings.ToLower(strings.TrimSpace(name)), num(r, 4), num(g, 4), num(b, 4)), nil
}
